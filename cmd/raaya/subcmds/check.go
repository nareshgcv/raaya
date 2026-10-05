// Package subcmds implements raaya's commands. Each returns an exit code
// (0 ok, 1 findings/regressions, 2 error) and an error.
package subcmds

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/nareshgcv/raaya/pkg/analysis"
	"github.com/nareshgcv/raaya/pkg/analysis/fixer"
	"github.com/nareshgcv/raaya/pkg/config"
	"github.com/nareshgcv/raaya/pkg/discovery"
	"github.com/nareshgcv/raaya/pkg/reporter"
)

// Version is reported in SARIF output; main sets it.
var Version = "dev"

// Check scans the repository and reports findings.
func Check(args []string) (int, error) {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	var dc config.Discovery
	dc.Register(fs)
	format := fs.String("format", "terminal", "output format: terminal, json, sarif, markdown")
	failOn := fs.String("fail-on", "HIGH", "exit 1 if any finding is at or above this severity: HIGH, MEDIUM, LOW or none")
	output := fs.String("output", "", "write the report to this file instead of stdout")
	fix := fs.Bool("fix", false, "replace literal secrets in project MCP configs with environment-variable references")
	var policies config.StringList
	fs.Var(&policies, "policy", "Rego policy file (repeatable; needs a build with -tags rego)")
	if _, err := config.ParseArgs(fs, args); err != nil {
		return 2, err
	}

	var threshold analysis.Severity
	if !strings.EqualFold(*failOn, "none") {
		var ok bool
		if threshold, ok = analysis.ParseSeverity(*failOn); !ok {
			return 2, fmt.Errorf("invalid --fail-on %q", *failOn)
		}
	}

	if *fix {
		if err := applyFixes(dc.Root); err != nil {
			return 2, err
		}
	}

	ctx := context.Background()
	g, err := discovery.Discover(ctx, dc.Options())
	if err != nil {
		return 2, err
	}
	findings := analysis.Evaluate(g)
	extra, err := analysis.EvaluatePolicies(ctx, policies, analysis.NewPolicyInput(g, findings))
	if err != nil {
		return 2, err
	}
	findings = append(findings, extra...)
	analysis.SortFindings(findings)

	w, closeOut, err := config.OpenOutput(*output)
	if err != nil {
		return 2, err
	}
	switch strings.ToLower(*format) {
	case "terminal", "text":
	default:
		err = fmt.Errorf("unknown --format %q", *format)
	}
	if cerr := closeOut(); err == nil {
		err = cerr
	}
	if err != nil {
		return 2, err
	}
	if threshold != "" && analysis.AnyAtOrAbove(findings, threshold) {
		return 1, nil
	}
	return 0, nil
}

// applyFixes rewrites literal secrets and explains what the user must do
// next. It prints variable names only, never values.
func applyFixes(root string) error {
	res, err := fixer.ExtractSecrets(root, true)
	if err != nil {
		return fmt.Errorf("fix: %w", err)
	}
	for _, c := range res.Changes {
		fmt.Fprintf(os.Stderr, "fixed: %s (%s): %s now reads %s\n", c.File, c.Server, c.EnvVar, c.Reference)
	}
	for _, s := range res.Skipped {
		fmt.Fprintf(os.Stderr, "not fixed: %s (%s): %s: %s\n", s.File, s.Server, s.EnvVar, s.Reason)
	}
	if len(res.Changes) > 0 {
		fmt.Fprintln(os.Stderr, "Set those variables from your secret manager, and rotate the old values: they remain in git history.")
	}
	return nil
}
