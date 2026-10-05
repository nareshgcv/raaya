import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nareshgcv/raaya/pkg/config"
)

const hookMarker = "# installed by raaya"

// Hook installs or removes a git pre-commit hook that runs `raaya check`.
func Hook(args []string) (int, error) {
	if len(args) == 0 || (args[0] != "install" && args[0] != "uninstall") {
		return 2, errors.New("usage: raaya hook install|uninstall [--force] [--path dir]")
	}
	sub := args[0]
	fs := flag.NewFlagSet("hook "+sub, flag.ContinueOnError)
	root := fs.String("path", ".", "repository root")
	force := fs.Bool("force", false, "replace an existing pre-commit hook not installed by raaya")
	if _, err := config.ParseArgs(fs, args[1:]); err != nil {
		return 2, err
	}
	hooksDir, err := git(*root, "rev-parse", "--git-path", "hooks")
	if err != nil {
		return 2, err
	}
	if !filepath.IsAbs(hooksDir) {
		hooksDir = filepath.Join(*root, hooksDir)
	}
	hookPath := filepath.Join(hooksDir, "pre-commit")
	existing, readErr := os.ReadFile(hookPath)
	ours := readErr == nil && strings.Contains(string(existing), hookMarker)

	if sub == "uninstall" {
		if readErr != nil {
			fmt.Println("no pre-commit hook installed")
			return 0, nil
		}
		if !ours {
			return 2, fmt.Errorf("%s was not installed by raaya; leaving it alone", hookPath)
		}
		return 0, os.Remove(hookPath)
	}

	if readErr == nil && !ours && !*force {
		return 2, fmt.Errorf("%s already exists; use --force to replace it", hookPath)
	}
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return 2, err
	}
	script := "#!/bin/sh\n" + hookMarker + "\nexec raaya check --fail-on HIGH\n"
	if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
		return 2, err
	}
	if err := os.Chmod(hookPath, 0o755); err != nil {
		return 2, err
	}
	fmt.Println("installed", hookPath)
	return 0, nil
}
