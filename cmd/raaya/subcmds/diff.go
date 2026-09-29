package subcmds

import (
	"fmt"
	"os"

	"raaya/pkg/analysis"
	"raaya/pkg/discovery"
	"raaya/pkg/reporter"

	"github.com/spf13/cobra"
)

var (
	baseRef  string
	jsonOut  bool
	prComment bool
)

// DiffCmd executes capability diff checks between base and current head states.
var DiffCmd = &cobra.Command{
	Use:   "diff [path]",
	Short: "Compute security-surface capability diffs between codebases/branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		analyzer := discovery.NewHybridAnalyzer()
		
		// Discover current HEAD graph
		headGraph, err := analyzer.DiscoverGraph(cmd.Context(), path)
		if err != nil {
			return fmt.Errorf("failed to analyze head: %w", err)
		}

		// Mock base graph for demonstration (In production, loaded from git base ref)
		baseGraph := analyzer.LoadBaseGraph(baseRef)

		// Compute Diff
		diff := analysis.ComputeDiff(baseGraph, headGraph)

		if prComment {
			comment, err := reporter.GeneratePRComment(diff)
			if err != nil {
				return err
			}
			fmt.Println(comment)
			return nil
		}

		if jsonOut {
			return reporter.RenderDiffJSON(os.Stdout, diff)
		}

		reporter.RenderDiffTerminal(os.Stdout, diff)
		return nil
	},
}

func init() {
	DiffCmd.Flags().StringVarP(&baseRef, "base", "b", "main", "Base branch/ref for diff")
	DiffCmd.Flags().BoolVar(&jsonOut, "json", false, "Output diff report in JSON format")
	DiffCmd.Flags().BoolVar(&prComment, "github-comment", false, "Render formatted PR markdown comment")
}
