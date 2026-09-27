package static

import (
	"fmt"
	"regexp"
	"strings"

	"raaya/pkg/discovery"
)

type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARN"
)

type Finding struct {
	RuleID      string
	RuleName    string
	Severity    Severity
	FilePath    string
	Line        int
	Col         int
	Message     string
	Snippet     string
	FixHint     string
	IsFixable   bool
}

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Evaluate(configs []discovery.MCPConfig, prompts []discovery.PromptAsset) []Finding {
	var findings []Finding

	findings = append(findings, checkPlaintextSecrets(configs)...)
	findings = append(findings, checkUnauthenticatedEndpoints(configs)...)
	findings = append(findings, checkOverpermissionedScopes(configs)...)
	findings = append(findings, checkPromptToolMismatch(configs, prompts)...)
	findings = append(findings, checkOrphanedCapabilities(configs, prompts)...)

	return findings
}

// RAA004: Plaintext Secrets
func checkPlaintextSecrets(configs []discovery.MCPConfig) []Finding {
	var findings []Finding
	secretPattern := regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password)\s*:\s*["'](sk-[a-zA-Z0-9]{20,}|[a-zA-Z0-9_-]{32,})["']`)

	for _, cfg := range configs {
		for lineNum, line := range cfg.Lines {
			if matches := secretPattern.FindStringSubmatch(line); len(matches) > 0 {
				findings = append(findings, Finding{
					RuleID:    "RAA004",
					RuleName:  "plaintext-secret",
					Severity:  SeverityError,
					FilePath:  cfg.Path,
					Line:      lineNum + 1,
					Col:       strings.Index(line, matches[0]) + 1,
					Message:   fmt.Sprintf("Hardcoded secret detected in server configuration"),
					Snippet:   strings.TrimSpace(line),
					FixHint:   "Move secret to .env and reference via ${ENV_VAR}",
					IsFixable: true,
				})
			}
		}
	}
	return findings
}

// RAA002: Unauthenticated Endpoints
func checkUnauthenticatedEndpoints(configs []discovery.MCPConfig) []Finding {
	var findings []Finding
	for _, cfg := range configs {
		for sName, server := range cfg.MCPServers {
			if strings.Contains(server.URL, "0.0.0.0") || (strings.HasPrefix(server.URL, "http://") && len(server.Headers) == 0) {
				findings = append(findings, Finding{
					RuleID:    "RAA002",
					RuleName:  "unauthenticated-endpoint",
					Severity:  SeverityError,
					FilePath:  cfg.Path,
					Line:      server.LineNumber,
					Col:       1,
					Message:   fmt.Sprintf("MCP Server '%s' exposes an HTTP endpoint without auth headers or binds to 0.0.0.0", sName),
					Snippet:   server.URL,
					FixHint:   "Bind endpoint to 127.0.0.1 or provide authorization headers",
					IsFixable: false,
				})
			}
		}
	}
	return findings
}

// RAA003: Over-Permissioned Scopes
func checkOverpermissionedScopes(configs []discovery.MCPConfig) []Finding {
	var findings []Finding
	for _, cfg := range configs {
		for sName, server := range cfg.MCPServers {
			for _, arg := range server.Args {
				if arg == "*" || arg == "/" || arg == "root" {
					findings = append(findings, Finding{
						RuleID:    "RAA003",
						RuleName:  "overpermissioned-scope",
						Severity:  SeverityWarning,
						FilePath:  cfg.Path,
						Line:      server.LineNumber,
						Col:       1,
						Message:   fmt.Sprintf("Server '%s' requests unrestricted root or wildcard file scope ('%s')", sName, arg),
						Snippet:   arg,
						FixHint:   "Restrict execution path to specific subdirectories (e.g., ./workspace)",
						IsFixable: false,
					})
				}
			}
		}
	}
	return findings
}

// RAA001: Prompt/Tool Mismatch
func checkPromptToolMismatch(configs []discovery.MCPConfig, prompts []discovery.PromptAsset) []Finding {
	var findings []Finding
	declaredTools := make(map[string]bool)
	for _, cfg := range configs {
		for _, server := range cfg.MCPServers {
			for _, tool := range server.DeclaredTools {
				declaredTools[tool] = true
			}
		}
	}

	for _, prompt := range prompts {
		for _, ref := range prompt.ToolReferences {
			if !declaredTools[ref.Name] {
				findings = append(findings, Finding{
					RuleID:    "RAA001",
					RuleName:  "prompt-tool-mismatch",
					Severity:  SeverityError,
					FilePath:  prompt.Path,
					Line:      ref.Line,
					Col:       ref.Col,
					Message:   fmt.Sprintf("Prompt references tool '%s' which is not defined in any MCP server config", ref.Name),
					Snippet:   ref.ContextSnippet,
					FixHint:   fmt.Sprintf("Add tool '%s' to mcp_config.json or update system prompt", ref.Name),
					IsFixable: false,
				})
			}
		}
	}
	return findings
}

// RAA005: Orphaned Capabilities
func checkOrphanedCapabilities(configs []discovery.MCPConfig, prompts []discovery.PromptAsset) []Finding {
	var findings []Finding
	usedTools := make(map[string]bool)
	for _, p := range prompts {
		for _, ref := range p.ToolReferences {
			usedTools[ref.Name] = true
		}
	}

	for _, cfg := range configs {
		for _, server := range cfg.MCPServers {
			for _, tool := range server.DeclaredTools {
				if !usedTools[tool] {
					findings = append(findings, Finding{
						RuleID:    "RAA005",
						RuleName:  "orphaned-capability",
						Severity:  SeverityWarning,
						FilePath:  cfg.Path,
						Line:      server.LineNumber,
						Col:       1,
						Message:   fmt.Sprintf("Tool capability '%s' is declared but never referenced in any prompt", tool),
						Snippet:   tool,
						FixHint:   "Remove unreferenced tool definition to reduce attack surface",
						IsFixable: true,
					})
				}
			}
		}
	}
	return findings
}
