// Package discovery builds Raaya's security graph from MCP client configs,
// source code, subagent definitions and (optionally) live MCP servers.
package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nareshgcv/raaya/pkg/graph"
)

// Options control discovery.
type Options struct {
	Root         string        // repository root; defaults to "."
	IncludeUser  bool          // also read user-level client configs
	ExtraConfigs []string      // additional config files
	Live         bool          // launch stdio servers and call tools/list
	LiveTimeout  time.Duration // per server; defaults to 20s
	Logf         func(format string, args ...any)
}

const (
	repoAgentID      = "agent:repository"
	sourceServerID   = "server:source"
	maxLiveResources = 200
)

// Discover builds the security graph for a repository.
//
// Node IDs never contain absolute paths of the scanned checkout, so graphs of
// two checkouts (for example a PR head and its base) can be compared.
func Discover(ctx context.Context, opts Options) (*graph.Graph, error) {
	if opts.Root == "" {
		opts.Root = "."
	}
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}
	d := &discoverer{
		g:          graph.NewGraph(),
		root:       root,
		logf:       opts.Logf,
		sourceDirs: map[string][]string{},
		specs:      map[string]ServerSpec{},
	}
	if d.logf == nil {
		d.logf = func(string, ...any) {}
	}

	// 1. MCP client configs: agents, servers and inferred resources.
	configs := ProjectConfigs(root)
	if opts.IncludeUser {
		configs = append(configs, UserConfigs()...)
	}
	for _, p := range opts.ExtraConfigs {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		configs = append(configs, ConfigFile{Path: abs, Agent: "custom", Scope: "custom"})
	}
	for _, cf := range configs {
		if err := d.addConfig(cf); err != nil {
			return nil, err
		}
	}

	// 2. Tool definitions and @raaya:capability annotations in source code.
	tools, err := ScanSource(root)
	if err != nil {
		return nil, fmt.Errorf("source scan: %w", err)
	}
	d.addSourceTools(tools)

	// 3. Subagents defined in prompt files.
	defs, err := ScanAgentDefinitions(root)
	if err != nil {
		return nil, fmt.Errorf("agent definitions: %w", err)
	}
	d.addAgentDefinitions(defs)

	// 4. Optional: ask running servers what they expose.
	if opts.Live {
		d.addLive(ctx, opts.LiveTimeout)
	}

	d.applyOverrides(tools)
	d.resolveAllowedTools()
	d.g.Normalize()
	return d.g, nil
}

type discoverer struct {
	g          *graph.Graph
	root       string
	logf       func(string, ...any)
	sourceDirs map[string][]string   // server ID -> repo dirs holding its code
	specs      map[string]ServerSpec // stdio servers, for live discovery
}

func (d *discoverer) addSourceTools(tools []SourceTool) {
	for _, t := range tools {
		serverID := d.serverForFile(t.File)
		if serverID == "" {
			// Tools whose server isn't in any client config are still part of
			// the repo's surface; attach them to a synthetic agent so that a PR
			// adding one still shows up in the diff.
			serverID = sourceServerID
			d.g.AddNode(&graph.Node{ID: repoAgentID, Name: "Repository code (not in a client config)", Type: graph.NodeAgent})
			d.g.AddNode(&graph.Node{ID: sourceServerID, Name: "source", Type: graph.NodeMCPServer, Metadata: map[string]string{graph.MetaTransport: "source"}})
			d.g.AddEdge(graph.Edge{SourceID: repoAgentID, TargetID: sourceServerID, Relation: "connects"})
		}
		perm, how := graph.MaxPermission(t.Permissions...), permFromOverride
		if perm == graph.PermNone {
			perm, how = InferToolPermission(t.Name, nil)
		}
		d.addTool(serverID, t.Name, perm, how, map[string]string{
			graph.MetaFile: t.File,
			graph.MetaLine: strconv.Itoa(t.Line),
		})
	}
}

// serverForFile returns the server whose code directory most specifically
// contains file, or "".
func (d *discoverer) serverForFile(file string) string {
	best, bestLen := "", -1
	for _, serverID := range sortedKeys(d.sourceDirs) {
		for _, dir := range d.sourceDirs[serverID] {
			if dir == "." || file == dir || strings.HasPrefix(file, dir+"/") {
				if l := len(dir); l > bestLen {
					best, bestLen = serverID, l
				}
			}
		}
	}
	return best
}

func (d *discoverer) addLive(ctx context.Context, timeout time.Duration) {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	for _, serverID := range sortedKeys(d.specs) {
		spec := d.specs[serverID]
		server := d.g.Nodes[serverID]
		res, err := ListServer(ctx, spec, timeout)
		if err != nil {
			server.Metadata["live_error"] = err.Error()
			d.logf("live discovery for %s failed: %v", spec.Name, err)
			continue
		}
		if res.ServerInfo != "" {
			server.Metadata["server_info"] = res.ServerInfo
		}
		for _, t := range res.Tools {
			perm, how := InferToolPermission(t.Name, t.Annotations)
			meta := map[string]string{}
			if a := t.Annotations; a != nil {
				setBoolMeta(meta, "readOnlyHint", a.ReadOnlyHint)
				setBoolMeta(meta, "destructiveHint", a.DestructiveHint)
				setBoolMeta(meta, "openWorldHint", a.OpenWorldHint)
			}
			d.addTool(serverID, t.Name, perm, how, meta)
		}
		for i, r := range res.Resources {
			if i >= maxLiveResources {
				server.Metadata["resources_truncated"] = "true"
				break
			}
			name := r.Name
			if name == "" {
				name = r.URI
			}
			d.addResource(serverID, "resource:mcp:"+r.URI, name, "exposes-resource", graph.PermRead, false, nil)
		}
	}
}

func toolID(serverID, name string) string {
	return "tool:" + strings.TrimPrefix(serverID, "server:") + ":" + name
}

// addTool adds a tool under a server. Better-informed permissions (live
// annotations, overrides) replace an earlier assumed one, never the reverse.
func (d *discoverer) addTool(serverID, name string, perm graph.PermissionLevel, how string, meta map[string]string) {
	id := toolID(serverID, name)
	if meta == nil {
		meta = map[string]string{}
	}
	meta[graph.MetaPermissionSource] = how
	exposes := func(e graph.Edge) bool { return e.TargetID == id && e.Relation == "exposes" }

	if existing, ok := d.g.Nodes[id]; ok {
		prev := existing.Metadata[graph.MetaPermissionSource]
		switch {
		case how == graph.PermissionAssumed && prev != graph.PermissionAssumed:
			d.g.AddNode(&graph.Node{ID: id, Metadata: meta}) // merge file/line only
			return
		case prev == graph.PermissionAssumed && how != graph.PermissionAssumed:
			d.g.SetPermission(exposes, perm)
			existing.Metadata[graph.MetaPermissionSource] = how
		}
	}
	d.g.AddNode(&graph.Node{ID: id, Name: name, Type: graph.NodeTool, Metadata: meta})
	d.g.AddEdge(graph.Edge{SourceID: serverID, TargetID: id, Relation: "exposes", Permission: perm})
}

func (d *discoverer) addResource(serverID, id, name, relation string, perm graph.PermissionLevel, inferred bool, meta map[string]string) {
	d.g.AddNode(&graph.Node{ID: id, Name: name, Type: graph.NodeResource, Metadata: meta})
	d.g.AddEdge(graph.Edge{SourceID: serverID, TargetID: id, Relation: relation, Permission: perm, Inferred: inferred})
}

// applyOverrides gives @raaya:capability the final say on a tool's permission.
func (d *discoverer) applyOverrides(tools []SourceTool) {
	overrides := map[string]graph.PermissionLevel{}
	for _, t := range tools {
		if p := graph.MaxPermission(t.Permissions...); p.Rank() > overrides[t.Name].Rank() {
			overrides[t.Name] = p
		}
	}
	for _, n := range d.g.NodesOfType(graph.NodeTool) {
		p, ok := overrides[n.Name]
		if !ok {
			continue
		}
		id := n.ID
		d.g.SetPermission(func(e graph.Edge) bool { return e.TargetID == id && e.Relation == "exposes" }, p)
		n.Metadata[graph.MetaPermissionSource] = permFromOverride
	}
}

// displayPath shows config paths relative to the repo, or to ~ for user configs.
func (d *discoverer) displayPath(p string) string {
	if rel, err := filepath.Rel(d.root, p); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
		if rel, err := filepath.Rel(home, p); err == nil && !strings.HasPrefix(rel, "..") {
			return "~/" + filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(p)
}

// resolveDir returns the directory a server runs in.
func resolveDir(configDir, cwd string, project bool) string {
	switch {
	case cwd != "" && filepath.IsAbs(cwd):
		return cwd
	case cwd != "":
		return filepath.Join(configDir, cwd)
	case project:
		return configDir
	default:
		return ""
	}
}

func appendMeta(n *graph.Node, key, value string) {
	parts := []string{value}
	if cur := n.Metadata[key]; cur != "" {
		parts = append(strings.Split(cur, ","), value)
	}
	seen := map[string]bool{}
	var out []string
	for _, p := range parts {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	n.Metadata[key] = strings.Join(out, ",")
}

func setBoolMeta(m map[string]string, key string, v *bool) {
	if v != nil {
		m[key] = strconv.FormatBool(*v)
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
