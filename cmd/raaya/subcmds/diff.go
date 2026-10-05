package subcmds

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nareshgcv/raaya/pkg/analysis"
	"github.com/nareshgcv/raaya/pkg/config"
	"github.com/nareshgcv/raaya/pkg/discovery"
	"github.com/nareshgcv/raaya/pkg/graph"
	"github.com/nareshgcv/raaya/pkg/reporter"
)

// Diff compares what agents can reach at HEAD against a git ref, or against
// a graph saved with `raaya graph --format json` (--base-graph).
func Diff(args []string) (int, error) {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	var dc config.Discovery
	dc.Register(fs)
	format := fs.String("format", "terminal", "output format: terminal, json, markdown")
	threshold := fs.String("threshold", "WRITE", "newly reachable nodes at or above this permission are regressions")
	output := fs.String("output", "", "write the report to this file instead of stdout")
	baseGraph := fs.String("base-graph", "", "compare against a saved graph JSON file instead of a git ref")
	noFail := fs.Bool("no-fail", false, "exit 0 even when there are regressions")
	pos, err := config.ParseArgs(fs, args)
	if err != nil {
		return 2, err
	}
	if (len(pos) == 1) == (*baseGraph != "") {
		return 2, errors.New("usage: raaya diff <base-ref> [flags]  or  raaya diff --base-graph graph.json [flags]")
	}
	thr, ok := graph.ParsePermission(*threshold)
	if !ok {
		return 2, fmt.Errorf("invalid --threshold %q", *threshold)
	}

	ctx := context.Background()
	opts := dc.Options()
	head, err := discovery.Discover(ctx, opts)
	if err != nil {
		return 2, fmt.Errorf("head: %w", err)
	}

	var base *graph.Graph
	if *baseGraph != "" {
		f, err := os.Open(*baseGraph)
		if err != nil {
			return 2, err
		}
		base, err = graph.ReadJSON(f)
		f.Close()
		if err != nil {
			return 2, err
		}
	} else {
		if dc.Live {
			fmt.Fprintln(os.Stderr, "raaya: --live also launches the servers configured at the base ref")
		}
		if base, err = graphAtRef(ctx, opts, pos[0]); err != nil {
			return 2, fmt.Errorf("base %s: %w", pos[0], err)
		}
	}
	d := analysis.ComputeDiff(base, head, thr)

	w, closeOut, err := config.OpenOutput(*output)
	if err != nil {
		return 2, err
	}
	switch strings.ToLower(*format) {
	case "terminal", "text":
		reporter.Diff(w, d, *output == "" && reporter.IsTerminal(os.Stdout))
	case "json":
		err = reporter.JSON(w, d)
	case "markdown", "md":
		reporter.DiffMarkdown(w, d)
	default:
		err = fmt.Errorf("unknown --format %q", *format)
	}
	if cerr := closeOut(); err == nil {
		err = cerr
	}
	if err != nil {
		return 2, err
	}
	if d.Regressions > 0 && !*noFail {
		return 1, nil
	}
	return 0, nil
}

// graphAtRef checks ref out into a temporary worktree and scans it, so the
// base graph is built by exactly the same code as the head graph.
func graphAtRef(ctx context.Context, opts discovery.Options, ref string) (*graph.Graph, error) {
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, err
	}
	top, err := git(root, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	rel, err := relativeTo(top, root)
	if err != nil {
		return nil, err
	}

	tmp, err := os.MkdirTemp("", "raaya-base-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	wt := filepath.Join(tmp, "wt")
	if _, err := git(top, "worktree", "add", "--detach", "--quiet", wt, ref); err != nil {
		return nil, err
	}
	defer func() { _, _ = git(top, "worktree", "remove", "--force", wt) }()

	opts.Root = filepath.Join(wt, rel)
	return discovery.Discover(ctx, opts)
}

// relativeTo returns root relative to top, resolving symlinks on both sides
// (macOS temp dirs and some CI checkouts are symlinked).
func relativeTo(top, root string) (string, error) {
	if t, err := filepath.EvalSymlinks(top); err == nil {
		top = t
	}
	if r, err := filepath.EvalSymlinks(root); err == nil {
		root = r
	}
	rel, err := filepath.Rel(top, root)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%s is not inside the git repository at %s", root, top)
	}
	return rel, nil
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}
