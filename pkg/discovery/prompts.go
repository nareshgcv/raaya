package discovery

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/nareshgcv/raaya/pkg/graph"
)

// AgentDefinition is a prompt-defined subagent, such as a Claude Code
// subagent in .claude/agents/*.md. Its YAML frontmatter may restrict which
// tools it gets; MCP tools are named mcp__<server>__<tool>.
type AgentDefinition struct {
	Name  string
	File  string   // slash-separated, relative to the repo root
	Tools []string // nil means the subagent inherits every tool of its parent
}

const subagentDir = ".claude/agents"

// ScanAgentDefinitions reads subagent definitions under root.
func ScanAgentDefinitions(root string) ([]AgentDefinition, error) {
	dir := filepath.Join(root, filepath.FromSlash(subagentDir))
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []AgentDefinition
	for _, e := range entries { // ReadDir returns entries sorted by name
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		def, ok := parseAgentFrontmatter(string(data))
		if !ok {
			continue
		}
		if def.Name == "" {
			def.Name = strings.TrimSuffix(e.Name(), ".md")
		}
		def.File = subagentDir + "/" + e.Name()
		out = append(out, def)
	}
}

// parseAgentFrontmatter reads `name` and `tools` from YAML frontmatter.
// tools may be a comma list, a [flow, list] or a block list of "- item" lines.
func parseAgentFrontmatter(doc string) (AgentDefinition, bool) {
	var def AgentDefinition
	doc = strings.ReplaceAll(doc, "\r\n", "\n")
	if !strings.HasPrefix(doc, "---\n") {
		return def, false
	}
	body := doc[4:]
	end := strings.Index(body, "\n---")
	if end < 0 {
		return def, false
	}

	inToolList := false
	for _, line := range strings.Split(body[:end], "\n") {
		trimmed := strings.TrimSpace(line)
		if inToolList {
			if item, ok := strings.CutPrefix(trimmed, "- "); ok {
				def.Tools = append(def.Tools, unquote(item))
				continue
			}
			inToolList = false
		}
		if line == "" || line[0] == ' ' || line[0] == '\t' {
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "name":
			def.Name = unquote(value)
		case "tools":
			def.Tools = []string{} // present, so the subagent is restricted
			if value == "" {
				inToolList = true
				continue
			}
			for _, t := range strings.Split(strings.Trim(value, "[]"), ",") {
				if t = unquote(strings.TrimSpace(t)); t != "" {
					def.Tools = append(def.Tools, t)
				}
			}
		}
	}
	return def, true
}

func unquote(s string) string {
	return strings.Trim(strings.TrimSpace(s), `"'`)
}

// parseMCPToolName splits mcp__server__tool. tool is "" (or "*") when the
// whole server is granted.
func parseMCPToolName(s string) (server, tool string, ok bool) {
	rest, found := strings.CutPrefix(s, "mcp__")
	if !found || rest == "" {
		return "", "", false
	}
	server, tool, _ = strings.Cut(rest, "__")
	return server, tool, server != ""
}

// addAgentDefinitions models restricted subagents as agents that reach only
// the MCP servers and tools they are granted. Subagents that inherit every
// tool add nothing their parent doesn't already show, so they are skipped.
func (d *discoverer) addAgentDefinitions(defs []AgentDefinition) {
	const parentID = "agent:claude-code"
	parentServers := map[string]string{} // server name -> node ID
	for _, e := range d.g.Edges {
		if e.SourceID == parentID && e.Relation == "connects" {
			if n := d.g.Nodes[e.TargetID]; n != nil {
				parentServers[n.Name] = n.ID
			}
		}
	}

	for _, def := range defs {
		if def.Tools == nil {
			continue
		}
		id := parentID + "/" + def.Name
		d.g.AddNode(&graph.Node{ID: id, Name: "Subagent " + def.Name, Type: graph.NodeAgent, Metadata: map[string]string{
		}})
		for _, t := range def.Tools {
			server, tool, ok := parseMCPToolName(t)
			if !ok {
				continue // built-in tools such as Read or Bash are outside MCP
			}
			serverID, known := parentServers[server]
			if !known {
				continue
			}
			if tool == "" || tool == "*" {
				d.g.AddEdge(graph.Edge{SourceID: id, TargetID: serverID, Relation: "connects"})
			} else {
				d.g.AddEdge(graph.Edge{SourceID: id, TargetID: toolID(serverID, tool), Relation: "allowed-tool"})
			}
		}
	}
}

// resolveAllowedTools gives each subagent→tool grant the tool's own
// permission, once live discovery and overrides have settled it.
func (d *discoverer) resolveAllowedTools() {
	toolPerm := map[string]graph.PermissionLevel{}
	for _, e := range d.g.Edges {
		if e.Relation == "exposes" && e.Permission.Rank() > toolPerm[e.TargetID].Rank() {
			toolPerm[e.TargetID] = e.Permission
		}
	}
	for target, p := range toolPerm {
		target := target
		d.g.SetPermission(func(e graph.Edge) bool { return e.Relation == "allowed-tool" && e.TargetID == target }, p)
	}
}
