package subcmds

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/nareshgcv/raaya/pkg/blastradius"
	"github.com/nareshgcv/raaya/pkg/config"
	"github.com/nareshgcv/raaya/pkg/discovery"
	"github.com/nareshgcv/raaya/pkg/reporter"
)

// Graph exports the discovered topology. Render SVG from DOT with
// `raaya graph --format dot | dot -Tsvg > graph.svg`.
func Graph(args []string) (int, error) {
	fs := flag.NewFlagSet("graph", flag.ContinueOnError)
	var dc config.Discovery
	dc.Register(fs)
	format := fs.String("format", "mermaid", "output format: mermaid, dot, json")
	output := fs.String("output", "", "write to this file instead of stdout")
	if _, err := config.ParseArgs(fs, args); err != nil {
		return 2, err
	}
	g, err := discovery.Discover(context.Background(), dc.Options())
	if err != nil {
		return 2, err
	}
	w, closeOut, err := config.OpenOutput(*output)
	if err != nil {
		return 2, err
	}
	switch strings.ToLower(*format) {
	case "mermaid":
		_, err = fmt.Fprint(w, g.Mermaid())
	case "dot":
		_, err = fmt.Fprint(w, g.DOT())
	case "json":
		err = g.WriteJSON(w)
	default:
		err = fmt.Errorf("unknown --format %q", *format)
	}
	if cerr := closeOut(); err == nil {
		err = cerr
	}
	if err != nil {
		return 2, err
	}
	return 0, nil
}

// Blast shows what a compromised node could reach.
func Blast(args []string) (int, error) {
	fs := flag.NewFlagSet("blast", flag.ContinueOnError)
	var dc config.Discovery
	dc.Register(fs)
	format := fs.String("format", "terminal", "output format: terminal, json")
	pos, err := config.ParseArgs(fs, args)
	if err != nil {
		return 2, err
	}
	if len(pos) != 1 {
		return 2, errors.New("usage: raaya blast <node-id> [flags]   (list IDs with: raaya graph --format json)")
	}
	g, err := discovery.Discover(context.Background(), dc.Options())
	if err != nil {
		return 2, err
	}
	id := pos[0]
	if _, ok := g.Nodes[id]; !ok {
		var similar []string
		for _, n := range g.SortedNodes() {
			if strings.Contains(strings.ToLower(n.ID), strings.ToLower(id)) {
				similar = append(similar, n.ID)
			}
		}
		if len(similar) > 0 {
			return 2, fmt.Errorf("no node %q; did you mean: %s", id, strings.Join(similar, ", "))
		}
		return 2, fmt.Errorf("no node %q", id)
	}
	r := blastradius.Calculate(g, id)
	if strings.EqualFold(*format, "json") {
		return 0, reporter.JSON(os.Stdout, r)
	}
	reporter.Blast(os.Stdout, g, r, reporter.IsTerminal(os.Stdout))
	return 0, nil
}
