package discovery

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nareshgcv/raaya/pkg/blastradius"
	"github.com/nareshgcv/raaya/pkg/graph"
)

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestStripJSONC(t *testing.T) {
	in := `{
	  // line comment
	  "servers": {
	    "x": { "url": "https://example.com/a//b", /* block */ "type": "http", },
	  },
	}`
	var v struct {
		Servers map[string]struct {
			URL string `json:"url"`
		} `json:"servers"`
	}
	if err := json.Unmarshal(stripJSONC([]byte(in)), &v); err != nil {
		t.Fatal(err)
	}
	if got := v.Servers["x"].URL; got != "https://example.com/a//b" {
		t.Fatalf("// inside a string must survive, got %q", got)
	}
}

func TestParseConfigAcceptsBothLayouts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".vscode/mcp.json", `{"servers": {"gh": {"type": "http", "url": "https://api.example.com/mcp"}}}`)
	writeFile(t, root, ".mcp.json", `{"mcpServers": {"fs": {"command": "npx", "args": ["x"]}}}`)
	vs, err := ParseConfig(filepath.Join(root, ".vscode/mcp.json"))
	if err != nil || len(vs) != 1 || vs[0].Name != "gh" || vs[0].Transport() != "http" {
		t.Fatalf("vscode layout: %+v %v", vs, err)
	}
	cc, err := ParseConfig(filepath.Join(root, ".mcp.json"))
	if err != nil || len(cc) != 1 || cc[0].Transport() != "stdio" {
		t.Fatalf("mcpServers layout: %+v %v", cc, err)
	}
}

func TestInferToolPermission(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		name string
		ann  *ToolAnnotations
		want graph.PermissionLevel
		how  string
	}{
		{"read_file", &ToolAnnotations{ReadOnlyHint: &yes}, graph.PermRead, permFromReadOnlyHint},
		{"read_shell_history", &ToolAnnotations{ReadOnlyHint: &yes}, graph.PermRead, permFromReadOnlyHint},
		{"run_shell", nil, graph.PermExecute, permFromName},
		{"execute_command", &ToolAnnotations{DestructiveHint: &no}, graph.PermExecute, permFromName},
		{"create_issue", nil, graph.PermWrite, graph.PermissionAssumed},
		{"create_issue", &ToolAnnotations{DestructiveHint: &no}, graph.PermWrite, permFromDestructiveHint},
	}
	for _, c := range cases {
		got, how := InferToolPermission(c.name, c.ann)
		if got != c.want || how != c.how {
			t.Errorf("%s: got %s/%s, want %s/%s", c.name, got, how, c.want, c.how)
		}
	}
}

func TestUnpinnedPackage(t *testing.T) {
	cases := []struct {
		cmd      string
		args     []string
		pkg      string
		unpinned bool
	}{
		{"npx", []string{"-y", "@modelcontextprotocol/server-github"}, "@modelcontextprotocol/server-github", true},
		{"npx", []string{"-y", "@modelcontextprotocol/server-github@1.2.3"}, "@modelcontextprotocol/server-github@1.2.3", false},
		{"npx", []string{"some-server@latest"}, "some-server@latest", true},
		{"uvx", []string{"mcp-server-git"}, "mcp-server-git", true},
		{"uvx", []string{"--python", "3.12", "mcp-server-git==0.6.2"}, "mcp-server-git==0.6.2", false},
		{"node", []string{"server.js"}, "", false},
	}
	for _, c := range cases {
		pkg, unpinned := UnpinnedPackage(c.cmd, c.args)
		if pkg != c.pkg || unpinned != c.unpinned {
			t.Errorf("%s %v: got %q/%v, want %q/%v", c.cmd, c.args, pkg, unpinned, c.pkg, c.unpinned)
		}
	}
}

func TestLooksLikeHardcodedSecret(t *testing.T) {
	cases := []struct {
		name, value string
		want        bool
	}{
		{"GITHUB_PERSONAL_ACCESS_TOKEN", "ghp_abcdefghijklmnopqrstuvwxyz0123456789", true},
		{"GITHUB_PERSONAL_ACCESS_TOKEN", "${GITHUB_TOKEN}", false},
		{"API_KEY", "${env:API_KEY}", false},
		{"API_KEY", "${input:apiKey}", false},
		{"API_KEY", "<your-api-key>", false},
		{"Authorization", "Bearer abcdefghijklmnopqrstuvwxyz", true},
		{"LOG_LEVEL", "debug", false},
	}
	for _, c := range cases {
		if got := LooksLikeHardcodedSecret(c.name, c.value); got != c.want {
			t.Errorf("%s=%q: got %v, want %v", c.name, c.value, got, c.want)
		}
	}
}

const sampleConfig = `{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {"GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_abcdefghijklmnopqrstuvwxyz0123456789"}
    },
    "files": {"command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem@2025.1.0", "/"]},
    "local": {"command": "python", "args": ["server.py"]}
  }
}`

const sampleServer = `from mcp.server.fastmcp import FastMCP
mcp = FastMCP("demo")

# @raaya:capability EXECUTE
@mcp.tool()
def deploy(env: str) -> str:
    return env

@mcp.tool()
def status() -> str:
    return "ok"
`

const sampleSubagent = `---
name: reviewer
description: Reviews code
tools: Read, Grep, mcp__local__status
---
You review code.
`

func sampleRepo(t *testing.T) string {
	root := t.TempDir()
	writeFile(t, root, ".mcp.json", sampleConfig)
	writeFile(t, root, "server.py", sampleServer)
	writeFile(t, root, ".claude/agents/reviewer.md", sampleSubagent)
	writeFile(t, root, ".claude/agents/helper.md", "---\nname: helper\n---\nInherits everything.\n")
	return root
}

func TestDiscoverBuildsGraphFromConfigAndSource(t *testing.T) {
	root := sampleRepo(t)
	g, err := Discover(context.Background(), Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}

	reach := blastradius.From(g, "agent:claude-code")
	want := map[string]graph.PermissionLevel{
		"tool:local:deploy": graph.PermExecute, // @raaya:capability
		"tool:local:status": graph.PermWrite,   // assumed
		"resource:github":   graph.PermWrite,   // via GITHUB_* credential
		"resource:path:/":   graph.PermWrite,   // filesystem argument
	}
	for id, p := range want {
		if reach[id] != p {
			t.Errorf("%s: got %q, want %q", id, reach[id], p)
		}
	}

	gh := g.Nodes["server:github"].Metadata
	if gh[graph.MetaUnpinnedPackage] != "@modelcontextprotocol/server-github" {
		t.Errorf("unpinned package not recorded: %v", gh)
	}
	if !strings.Contains(gh[graph.MetaHardcodedSecrets], "GITHUB_PERSONAL_ACCESS_TOKEN") {
		t.Errorf("hardcoded secret not recorded: %v", gh)
	}
	if v := g.Nodes["server:files"].Metadata[graph.MetaUnpinnedPackage]; v != "" {
		t.Errorf("pinned package flagged as unpinned: %q", v)
	}
	if g.Nodes["resource:path:/"].Metadata[graph.MetaBroadPath] != "true" {
		t.Error("filesystem root should be marked broad")
	}

	raw, _ := json.Marshal(g)
	if strings.Contains(string(raw), "ghp_abcdef") {
		t.Fatal("secret value leaked into the graph")
	}
	if strings.Contains(string(raw), filepath.ToSlash(root)) {
		t.Fatal("absolute checkout path leaked into the graph; base/head diffs would break")
	}
}

func TestRestrictedSubagentReachesOnlyGrantedTools(t *testing.T) {
	g, err := Discover(context.Background(), Options{Root: sampleRepo(t)})
	if err != nil {
		t.Fatal(err)
	}
	reach := blastradius.From(g, "agent:claude-code/reviewer")
	if reach["tool:local:status"] != graph.PermWrite {
		t.Errorf("granted tool: got %q, want WRITE", reach["tool:local:status"])
	}
	for _, id := range []string{"tool:local:deploy", "resource:github", "server:github"} {
		if _, ok := reach[id]; ok {
			t.Errorf("reviewer should not reach %s", id)
		}
	}
	if _, ok := g.Nodes["agent:claude-code/helper"]; ok {
		t.Error("a subagent that inherits all tools adds nothing and should be skipped")
	}
}

func TestDiscoverIsDeterministic(t *testing.T) {
	root := sampleRepo(t)
	a, err := Discover(context.Background(), Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Discover(context.Background(), Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.SortedEdges(), b.SortedEdges()) {
		t.Fatal("two scans of the same tree produced different edges")
	}
}

func TestUnboundSourceToolsAttachToRepositoryAgent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "tools/index.ts", `server.registerTool("delete_repo", {}, async () => {})`)
	g, err := Discover(context.Background(), Options{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := blastradius.From(g, repoAgentID)["tool:source:delete_repo"]; !ok {
		t.Fatalf("unbound tool should be reachable from %s", repoAgentID)
	}
}

func TestParseAgentFrontmatterBlockList(t *testing.T) {
	def, ok := parseAgentFrontmatter("---\nname: ops\ntools:\n  - mcp__github\n  - Bash\n---\n")
	if !ok || def.Name != "ops" || !reflect.DeepEqual(def.Tools, []string{"mcp__github", "Bash"}) {
		t.Fatalf("got %+v %v", def, ok)
	}
}
