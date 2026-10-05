package fixer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nareshgcv/raaya/pkg/discovery"
)

const token = "ghp_abcdefghijklmnopqrstuvwxyz0123456789"

func TestExtractSecretsRewritesInPlace(t *testing.T) {
	root := t.TempDir()
	config := `{
  "mcpServers": {
    "github": {
      "command": "npx",
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "` + token + `",
        "LOG_LEVEL": "debug"
      }
    }
  }
}
`
	path := filepath.Join(root, ".mcp.json")
	if err := os.WriteFile(path, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	dry, err := ExtractSecrets(root, false)
	if err != nil || len(dry.Changes) != 1 {
		t.Fatalf("dry run: %+v %v", dry, err)
	}
	if data, _ := os.ReadFile(path); !strings.Contains(string(data), token) {
		t.Fatal("dry run must not modify the file")
	}

	res, err := ExtractSecrets(root, true)
	if err != nil || len(res.Changes) != 1 || res.Changes[0].Reference != "${GITHUB_PERSONAL_ACCESS_TOKEN}" {
		t.Fatalf("got %+v %v", res, err)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), token) {
		t.Fatal("secret still present")
	}
	if !strings.Contains(string(data), `"LOG_LEVEL": "debug"`) {
		t.Fatal("unrelated values and formatting must be preserved")
	}
	specs, err := discovery.ParseConfig(path)
	if err != nil || specs[0].Env["GITHUB_PERSONAL_ACCESS_TOKEN"] != "${GITHUB_PERSONAL_ACCESS_TOKEN}" {
		t.Fatalf("rewritten config is not valid: %+v %v", specs, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".env")); !os.IsNotExist(err) {
		t.Fatal("the fixer must never write secrets to a .env file")
	}
}

func TestExtractSecretsUsesClientSyntax(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".vscode"), 0o755); err != nil {
		t.Fatal(err)
	}
	config := `{"servers": {"gh": {"command": "npx", "env": {"GITHUB_TOKEN": "` + token + `"}}}}`
	if err := os.WriteFile(filepath.Join(root, ".vscode", "mcp.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := ExtractSecrets(root, true)
	if err != nil || len(res.Changes) != 1 || res.Changes[0].Reference != "${env:GITHUB_TOKEN}" {
		t.Fatalf("got %+v %v", res, err)
	}
}
