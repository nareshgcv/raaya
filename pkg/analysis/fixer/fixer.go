// Package fixer applies safe, mechanical remediations to MCP configs.
//
// It never moves a secret anywhere. A literal credential in a config's env
// block is replaced with a reference to an environment variable of the same
// name; you set that variable from your secret manager. The old value is
// still in git history, so it must be rotated regardless.
package fixer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/nareshgcv/raaya/pkg/discovery"
)

// Change is one rewritten secret. It deliberately carries no value.
type Change struct {
	File      string `json:"file"`
	Server    string `json:"server"`
	EnvVar    string `json:"env_var"`
	Reference string `json:"reference"`
}

// Skipped is a secret the fixer found but would not rewrite.
type Skipped struct {
	File   string `json:"file"`
	Server string `json:"server"`
	EnvVar string `json:"env_var"`
	Reason string `json:"reason"`
}

// Result of ExtractSecrets.
type Result struct {
	Changes []Change  `json:"changes"`
	Skipped []Skipped `json:"skipped"`
}

// reference returns the env-var syntax each client expands in its config.
// Claude Desktop does not expand variables, so its configs are skipped.
func reference(agent, name string) (string, bool) {
	switch agent {
	case "claude-code", "generic", "custom":
		return "${" + name + "}", true
	case "cursor", "vscode":
		return "${env:" + name + "}", true
	default:
		return "", false
	}
}

// ExtractSecrets replaces literal credentials in the env blocks of the
// project's MCP configs with environment-variable references. With
// write=false it only reports what it would change.
func ExtractSecrets(root string, write bool) (Result, error) {
	res := Result{Changes: []Change{}, Skipped: []Skipped{}}
	for _, cf := range discovery.ProjectConfigs(root) {
		specs, err := discovery.ParseConfig(cf.Path)
		if err != nil {
			return res, err
		}
		data, err := os.ReadFile(cf.Path)
		if err != nil {
			return res, err
		}
		rel, err := filepath.Rel(root, cf.Path)
		if err != nil {
			rel = cf.Path
		}
		rel = filepath.ToSlash(rel)

		updated := data
		for _, s := range specs {
			names := make([]string, 0, len(s.Env))
			for name := range s.Env {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				if !discovery.LooksLikeHardcodedSecret(name, s.Env[name]) {
					continue
				}
				ref, ok := reference(cf.Agent, name)
				if !ok {
					res.Skipped = append(res.Skipped, Skipped{rel, s.Name, name, "this client does not expand environment variables in its config"})
					continue
				}
				next, n := replaceEnvValue(updated, name, s.Env[name], ref)
				if n == 0 {
					res.Skipped = append(res.Skipped, Skipped{rel, s.Name, name, "value is encoded unusually; edit it by hand"})
					continue
				}
				updated = next
				res.Changes = append(res.Changes, Change{File: rel, Server: s.Name, EnvVar: name, Reference: ref})
			}
		}

		if write && !bytes.Equal(updated, data) {
			info, err := os.Stat(cf.Path)
			if err != nil {
				return res, err
			}
			if err := os.WriteFile(cf.Path, updated, info.Mode().Perm()); err != nil {
				return res, fmt.Errorf("write %s: %w", rel, err)
			}
		}
	}
	return res, nil
}

// replaceEnvValue rewrites `"NAME": "<value>"` to `"NAME": "<ref>"` in place,
// leaving the rest of the file's formatting untouched.
func replaceEnvValue(data []byte, name, value, ref string) ([]byte, int) {
	encodedValue, err := jsonString(value)
	if err != nil {
		return data, 0
	}
	encodedRef, err := jsonString(ref)
	if err != nil {
		return data, 0
	}
	pattern := regexp.MustCompile(`("` + regexp.QuoteMeta(name) + `"\s*:\s*)` + regexp.QuoteMeta(encodedValue))
	count := 0
	out := pattern.ReplaceAllFunc(data, func(match []byte) []byte {
		count++
		prefix := pattern.FindSubmatch(match)[1]
		return append(append([]byte{}, prefix...), encodedRef...)
	})
	return out, count
}

func jsonString(s string) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return "", err
	}
	return string(bytes.TrimRight(buf.Bytes(), "\n")), nil
}
