// Command raaya maps what AI agents can reach through MCP servers, tools and
// resources, and reports risky capability paths.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/nareshgcv/raaya/cmd/raaya/subcmds"
	"github.com/nareshgcv/raaya/pkg/discovery"
)

// version is set at build time: go build -ldflags "-X main.version=v0.2.0"
var version = "dev"

const usage = `raaya maps what AI agents can reach through MCP servers, tools and resources.

Usage:
  raaya check  [flags]             scan and report findings (--fix rewrites literal secrets)
  raaya diff   <base-ref> [flags]  compare agent reach against a git ref
  raaya graph  [flags]             export the graph (mermaid, dot, json)
  raaya blast  <node-id> [flags]   what a compromised node could reach
  raaya doctor [flags]             check the environment and configs
  raaya hook   install|uninstall   manage the git pre-commit hook
  raaya version

Run "raaya <command> -h" for flags.
`

// Exit codes: 0 ok, 1 findings or regressions, 2 usage or runtime error.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	subcmds.Version = version
	discovery.ClientVersion = version

	commands := map[string]func([]string) (int, error){
		"check":  subcmds.Check,
		"diff":   subcmds.Diff,
		"graph":  subcmds.Graph,
		"blast":  subcmds.Blast,
		"doctor": subcmds.Doctor,
		"hook":   subcmds.Hook,
	}
	name, args := os.Args[1], os.Args[2:]
	switch name {
	case "version", "--version":
		fmt.Println(version)
		return
	case "help", "-h", "--help":
		fmt.Print(usage)
		return
	}
	run, ok := commands[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", name, usage)
		os.Exit(2)
	}

	code, err := run(args)
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "raaya:", err)
		if code == 0 {
			code = 2
		}
	}
	os.Exit(code)
}
