// Package config holds the command-line configuration shared by raaya's
// subcommands.
package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/nareshgcv/raaya/pkg/discovery"
)

// Discovery are the flags that control how the graph is built.
type Discovery struct {
	Root    string
	User    bool
	Live    bool
	Timeout time.Duration
	Configs StringList
}

// Register adds the discovery flags to fs.
func (d *Discovery) Register(fs *flag.FlagSet) {
	fs.StringVar(&d.Root, "path", ".", "repository root to scan")
	fs.BoolVar(&d.User, "user", false, "also scan user-level client configs (Claude Desktop, Cursor, Claude Code)")
	fs.BoolVar(&d.Live, "live", false, "launch stdio MCP servers and call tools/list; this EXECUTES configured commands, never use on untrusted PRs")
	fs.DurationVar(&d.Timeout, "live-timeout", 20*time.Second, "per-server timeout for --live")
	fs.Var(&d.Configs, "config", "additional MCP config file (repeatable)")
}

// Options converts the flags into discovery options. Warnings go to stderr.
func (d *Discovery) Options() discovery.Options {
	return discovery.Options{
		Root:         d.Root,
		IncludeUser:  d.User,
		ExtraConfigs: d.Configs,
		Live:         d.Live,
		LiveTimeout:  d.Timeout,
		Logf: func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, "raaya: "+format+"\n", args...)
		},
	}
}

// StringList is a repeatable string flag.
type StringList []string

func (s *StringList) String() string     { return strings.Join(*s, ",") }
func (s *StringList) Set(v string) error { *s = append(*s, v); return nil }

// ParseArgs parses flags that may appear before or after positional
// arguments (the standard flag package stops at the first positional one).
func ParseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			return positional, nil
		}
		positional = append(positional, rest[0])
		args = rest[1:]
	}
}

// OpenOutput returns stdout, or a created file when path is set, plus a
// close function that is always safe to call.
func OpenOutput(path string) (io.Writer, func() error, error) {
	if path == "" || path == "-" {
		return os.Stdout, func() error { return nil }, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return f, f.Close, nil
}
