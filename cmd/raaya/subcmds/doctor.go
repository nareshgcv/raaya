package subcmds

import (
	"flag"
	"fmt"
	"os/exec"

	"github.com/nareshgcv/raaya/pkg/analysis"
	"github.com/nareshgcv/raaya/pkg/config"
	"github.com/nareshgcv/raaya/pkg/discovery"
)

// Doctor checks the environment and summarises what raaya can see.
func Doctor(args []string) (int, error) {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	root := fs.String("path", ".", "repository root")
	user := fs.Bool("user", false, "also check user-level client configs")
	if _, err := config.ParseArgs(fs, args); err != nil {
		return 2, err
	}
	ok := func(format string, a ...any) { fmt.Printf("  ✓ "+format+"\n", a...) }
	warn := func(format string, a ...any) { fmt.Printf("  ! "+format+"\n", a...) }

	fmt.Println("Environment")
	if _, err := exec.LookPath("git"); err != nil {
		warn("git not found; `raaya diff <ref>` and hooks need it")
	} else if _, err := git(*root, "rev-parse", "--is-inside-work-tree"); err != nil {
		warn("%s is not a git work tree; `raaya diff <ref>` needs one", *root)
	} else {
		ok("git repository")
	}
	if analysis.RegoEnabled {
		ok("Rego policies supported")
	} else {
		warn("built without Rego; --policy needs a build with -tags rego")
	}

	fmt.Println("\nMCP configs")
	configs := discovery.ProjectConfigs(*root)
	if *user {
		configs = append(configs, discovery.UserConfigs()...)
	}
	if len(configs) == 0 {
		warn("none found (looked for .mcp.json, mcp.json, .cursor/mcp.json, .vscode/mcp.json, claude_desktop_config.json)")
	}
	for _, cf := range configs {
		specs, err := discovery.ParseConfig(cf.Path)
		if err != nil {
			warn("%s: %v", cf.Path, err)
			continue
		}
		ok("%s (%s, %d server(s))", cf.Path, discovery.AgentName(cf.Agent), len(specs))
		for _, s := range specs {
			switch {
			case s.Command == "":
				fmt.Printf("      %-20s %s (not queried by --live)\n", s.Name, s.Transport())
			default:
				if _, err := exec.LookPath(s.Command); err != nil {
					fmt.Printf("      %-20s stdio, command %q not on PATH\n", s.Name, s.Command)
				} else {
					fmt.Printf("      %-20s stdio, %s\n", s.Name, s.Command)
				}
			}
		}
	}

	fmt.Println("\nSource and prompts")
	if tools, err := discovery.ScanSource(*root); err != nil {
		warn("source scan failed: %v", err)
	} else {
		annotated := 0
		for _, t := range tools {
			if len(t.Permissions) > 0 {
				annotated++
			}
		}
		ok("%d tool definition(s), %d with @raaya:capability", len(tools), annotated)
	}
	if defs, err := discovery.ScanAgentDefinitions(*root); err != nil {
		warn("subagent scan failed: %v", err)
	} else {
		restricted := 0
		for _, d := range defs {
			if d.Tools != nil {
				restricted++
			}
		}
		ok("%d subagent(s) in .claude/agents, %d with a tool allowlist", len(defs), restricted)
	}
	return 0, nil
}
