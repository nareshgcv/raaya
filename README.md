

🛡️ Raaya

> **Static & Dynamic Security Reachability Scanner for AI Agents and MCP Infrastructure**
Raaya is a lightweight, single-binary security scanner built in Go for analyzing the security surface of AI agent ecosystems and Model Context Protocol (MCP) integrations.

It combines source-code analysis, MCP configuration discovery, runtime inspection, and security-graph analysis to answer a critical question:

**What can this AI agent reach, what capabilities can flow through its MCP integrations, and what happens if one of those components is compromised?**

Raaya constructs a Security Graph mapping:

Agent → MCP Server → Tool → Capability → Resource

It then computes transitive capability propagation, evaluates security policies, measures blast radius, and can compare security-graph states across Git branches or commit references to detect newly introduced security-surface expansions.

Raaya is designed to run locally, quickly, and without external infrastructure.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

✨ Key Features

🔍 Hybrid Discovery Engine

Raaya combines multiple discovery techniques instead of relying on a single source of truth.

It can inspect:

Source-code ASTs

MCP configuration files

.cursor/mcp.json

mcp.json

claude_desktop_config.json

Python tool definitions

TypeScript tool definitions

Agent and prompt references

// @raaya:capability annotations

Environment and credential configuration

Local TCP/SSE listeners

Live MCP runtime endpoints

Conceptually:

                 ┌──────────────────────┐
                 │   Hybrid Discovery   │
                 └──────────┬───────────┘
                            │
       ┌────────────────────┼────────────────────┐
       ▼                    ▼                    ▼
   Source AST          MCP Configs          Runtime Ports
       │                    │                    │
       └────────────────────┼────────────────────┘
                            ▼
                    Discovered Assets

This allows Raaya to identify security relationships that may exist only in source code, configuration, or the running environment.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🕸️ Security Graph

Raaya converts discovered AI infrastructure into an in-memory directed Security Graph.

The core topology is:

Agent
  │
  ▼
MCP Server
  │
  ▼
Tool
  │
  ▼
Capability
  │
  ▼
Resource

Resources may include:

Filesystem
Network
Database
Secrets
Cloud APIs
Repositories
Execution environments

Graph nodes and edges carry security-relevant metadata, including permissions such as:

READ
WRITE
EXECUTE
ADMIN

The graph provides the foundation for:

Reachability analysis

Capability propagation

Blast-radius analysis

Policy evaluation

Capability diffing

Security-surface change detection

Graph visualization

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🔗 Transitive Capability Propagation

Raaya does not only inspect direct permissions.

It follows relationships through the Security Graph to determine how capabilities propagate across an agent ecosystem.

For example:

Agent
  │
  ▼
github-mcp
  │
  ▼
github_search
  │
  ▼
Repository

Or:

Agent
  │
  ▼
filesystem-mcp
  │
  ▼
read_file
  │
  ▼
Filesystem
  │
  ▼
Sensitive Resource

The propagation engine calculates downstream capabilities using graph reachability so that indirect access is visible.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

💥 Blast Radius Analysis

Raaya can analyze the downstream security impact of a compromised component.

It answers questions such as:

**If this Agent, MCP Server, Tool, or Resource is compromised, what can be reached next?**

Example:

research-agent
      │
      ▼
   github-mcp
      │
      ▼
 github_search
      │
      ▼
   repository
      │
      ▼
   source code

The blast-radius engine evaluates transitive graph relationships and produces security-relevant exposure information for compromised nodes.

Implementation foundation:

pkg/blastradius/
├── calculator.go
└── propagation.go

The propagation solver is based on graph traversal and capability reachability.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🔀 PR Capability Diffing

Traditional security scanners often compare only findings.

Raaya can compare the security graph itself between Git references.

Conceptually:

                 Base Commit
                     │
                     ▼
               Security Graph
                     │
                     │
                 capability
                   diff
                     │
                     ▼
                New Commit
                     │
                     ▼
               Security Graph
                     │
                     ▼
             Security Surface
                  Changes

This allows Raaya to detect changes such as:

Agent
  │
  └── READ → repository

becoming:

Agent
  │
  ├── READ    → repository
  ├── WRITE   → repository
  └── EXECUTE → tool

The capability-diff engine can identify:

New capability paths

Permission escalations

Newly reachable resources

New Agent → MCP Server relationships

New Tool → Resource relationships

Security-surface expansions

Implementation:

pkg/analysis/delta.go

This is intended for CI/CD and pull-request security checks.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🛡️ Security Policy Engine

Raaya combines deterministic native Go rules with graph-aware OPA/Rego policy evaluation.

                    Security Graph
                         │
              ┌──────────┴──────────┐
              ▼                     ▼
        Native Go Rules         OPA / Rego
              │                     │
              └──────────┬──────────┘
                         ▼
                    Findings

Native Go Rules

High-confidence built-in security checks live under:

pkg/analysis/rules/

These rules provide fast, deterministic checks for common AI-agent and MCP security issues.

OPA / Rego Policies

Graph-based policies live under:

pkg/analysis/policies/

They can evaluate relationships across:

Agents

MCP Servers

Tools

Capabilities

Resources

Permissions

Reachability paths

The Rego engine is implemented in:

pkg/analysis/rego_engine.go

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🔐 Detection Rules

The current Phase 1 security checks include native security rules and Rego-based policy enforcement.

| Rule ID | Name | Severity | Description |

|---|---|---:|---|

| RAAYA-001-SECRETS | Hardcoded Credentials | CRITICAL | Detects plaintext API keys, tokens, passwords, and other credential-like values in supported configuration files. |

| RAAYA-002-UNAUTH-ENDPOINT | Unauthenticated Server | HIGH | Detects HTTP/SSE MCP servers running without identifiable authentication headers, credentials, or secrets. |

| RAAYA-REGO-001 | Rego Policy Violation | Policy-defined | Reports violations produced by embedded OPA/Rego policies evaluated against the Security Graph. |

RAAYA-001-SECRETS

Example:

{
  "apiKey": "sk-live-xxxxxxxx"
}

Run automatic remediation where supported:

raaya check --fix

Raaya can extract supported credentials into an environment file, replace the original configuration value, and verify .gitignore protection.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

RAAYA-002-UNAUTH-ENDPOINT

Example:

MCP Server
    │
    ├── Transport: SSE
    ├── Address: 0.0.0.0:3000
    └── Authentication: None

Live listener discovery:

raaya doctor --live

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

RAAYA-REGO-001

Rego policies can reason about graph relationships rather than isolated files.

For example:

Agent
  │
  ▼
MCP Server
  │
  ▼
Tool
  │
  ▼
Capability
  │
  ▼
Resource

A policy can inspect the complete relationship and reject a path that violates an organization's security requirements.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🌐 Active Runtime Discovery

Raaya can inspect locally running services and supported MCP listeners.

raaya doctor --live

It probes supported local TCP/SSE endpoints and analyzes discovered MCP listeners for potentially unauthenticated exposure.

This complements static discovery by identifying security issues that may not be visible in repository configuration alone.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

📊 Graph Export

Security topology can be exported into visualization-friendly formats.

Mermaid

raaya graph --format mermaid

Graphviz DOT

raaya graph --format dot

Example topology:

Agent
 │
 ├── READ ───────► MCP Server
 │                   │
 │                   ├── EXECUTE ─► Tool
 │                   │                 │
 │                   │                 └── READ ─► Resource
 │                   │
 │                   └── WRITE ───► Resource

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

📈 Differential Security Analysis

Raaya supports both traditional finding baselines and Security Graph capability diffs.

Base
 │
 ▼
Security Graph
 │
 ├── Existing findings ─────► Existing
 │
 └── New capabilities ─────► Security Surface Change
                                  │
                                  ▼
                              CI decision

This enables teams to distinguish historical findings from newly introduced security exposure.

It also allows CI to focus on changes that expand an agent's reachable capabilities rather than requiring an existing repository to become completely clean before adoption.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

📝 CI/CD & GitHub PR Integration

Raaya supports security-oriented CI/CD workflows with machine-readable and developer-friendly outputs.

Supported output formats include:

Terminal
JSON
SARIF
Markdown

Generate JSON:

raaya check --format json

Generate SARIF:

raaya check --format sarif > results.sarif

The reporting layer can also generate Markdown summaries suitable for GitHub pull-request comments.

Conceptually:

Git Commit / Pull Request
          │
          ▼
       Raaya
          │
          ├── Discovery
          ├── Security Graph
          ├── Capability Diff
          ├── Rule Evaluation
          └── Blast Radius
          │
          ▼
   ┌──────┼─────────┐
   ▼      ▼         ▼
  JSON   SARIF   PR Markdown

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🪝 Git Pre-Commit Hooks

Raaya can install a Git pre-commit hook:

raaya hook install

This allows security checks to run before changes are committed.

The goal is to catch AI-agent security-surface changes as early as possible in the development lifecycle.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🚀 Quick Start

Installation

Clone the repository:

git clone https://github.com/your-org/raaya.git
cd raaya

Build the binary:

go build -o raaya ./cmd/raaya

Run discovery:

./raaya doctor

Run the security scan:

./raaya check

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

⚡ 30-Second Workflow

        ┌─────────────┐
        │  Repository │
        └──────┬──────┘
               │
               ▼
       ┌────────────────┐
       │ raaya doctor   │
       │   Discovery    │
       └───────┬────────┘
               │
               ▼
       ┌────────────────┐
       │  Security      │
       │     Graph      │
       └───────┬────────┘
               │
        ┌──────┼─────────┐
        ▼      ▼         ▼
      Rules  Blast     Graph
             Radius    Analysis
        │      │         │
        └──────┼─────────┘
               ▼
        ┌───────────────┐
        │ raaya check   │
        │   Findings    │
        └───────┬───────┘
                │
          ┌─────┼──────┐
          ▼     ▼      ▼
        Fix    Diff    CI/CD

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🩺 raaya doctor

Use doctor to inspect what Raaya discovers in the current workspace.

raaya doctor

For live listener discovery:

raaya doctor --live

Example:

Raaya Doctor

Workspace: ./my-agent

Discovery
────────────────────────────────────

✓ MCP configuration
  .cursor/mcp.json

✓ MCP configuration
  mcp.json

✓ Python tools
  src/tools/search.py

✓ TypeScript tools
  src/tools/files.ts

✓ Agent / prompt references
  src/agent.py

✓ Environment configuration
  .env

✓ Local runtime listeners
  127.0.0.1:3000

Discovery complete.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🔎 raaya check

Run the security analysis:

raaya check

Example:

Raaya Security Scan

✗ RAAYA-001-SECRETS CRITICAL
  .cursor/mcp.json:14

  Hardcoded credential detected.

  Fix:
  Move the credential to an environment variable.

────────────────────────────────────────────

✗ RAAYA-002-UNAUTH-ENDPOINT HIGH
  mcp.json:12

  MCP server is running without
  identifiable authentication.

  Fix:
  Configure authentication or restrict
  endpoint exposure.

────────────────────────────────────────────

✗ RAAYA-REGO-001
  Security Graph

  Policy violation detected by
  embedded Rego policy.

Scan complete.

The intended diagnostic structure is:

Rule
  ↓
Location
  ↓
Problem
  ↓
Security context
  ↓
Suggested remediation

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🏗️ Architecture

Raaya is organized around discovery, graph construction, analysis, blast-radius calculation, capability diffing, remediation, and reporting.

raaya/
├── cmd/
│   └── raaya/
│       ├── main.go
│       └── subcmds/
│           ├── check.go          # Security analysis & rule evaluation
│           ├── diff.go           # Capability diff & security-surface checks
│           ├── doctor.go         # Environment & discovery health check
│           ├── graph.go          # Mermaid / DOT graph export
│           └── hook.go           # Git pre-commit hook manager
│
├── pkg/
│   ├── analysis/
│   │   ├── delta.go              # Capability diffing engine
│   │   ├── engine.go             # Analysis orchestrator
│   │   ├── rego_engine.go        # Embedded OPA/Rego evaluator
│   │   ├── fixer/                # Automated remediation
│   │   ├── policies/             # OPA/Rego policies
│   │   └── rules/                # Native security rules
│   │
│   ├── blastradius/
│   │   ├── calculator.go         # Risk / blast-radius evaluation
│   │   └── propagation.go        # Transitive capability propagation
│   │
│   ├── config/
│   │   └── config.go             # Local runtime configuration
│   │
│   ├── discovery/
│   │   ├── annotation.go         # @raaya:capability parser
│   │   ├── hybrid.go             # Hybrid discovery orchestrator
│   │   ├── live.go               # Runtime / listener scanner
│   │   ├── mcp.go                # MCP schema & config scanner
│   │   └── prompt.go             # Prompt / AST scanner
│   │
│   ├── graph/
│   │   ├── builder.go            # Security Graph construction
│   │   ├── export.go             # Graph serialization
│   │   ├── types.go              # Nodes, edges & permissions
│   │   └── visualizer.go         # Mermaid / Graphviz exporters
│   │
│   └── reporter/
│       ├── comment.go             # GitHub PR Markdown renderer
│       ├── diff_reporter.go       # Terminal / JSON diff output
│       ├── json.go                # Structured JSON output
│       ├── sarif.go               # SARIF output
│       └── terminal.go            # CLI output

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🔄 Analysis Flow

                         Repository / Runtime
                                  │
                                  ▼
                    ┌──────────────────────────┐
                    │   Hybrid Discovery       │
                    │                          │
                    │ AST / Config / MCP       │
                    │ Annotations / Prompts    │
                    │ Runtime Ports            │
                    └────────────┬─────────────┘
                                 │
                                 ▼
                    ┌──────────────────────────┐
                    │     Security Graph       │
                    │                          │
                    │ Agent                    │
                    │ MCP Server               │
                    │ Tool                     │
                    │ Capability               │
                    │ Resource                 │
                    └────────────┬─────────────┘
                                 │
              ┌──────────────────┼──────────────────┐
              ▼                  ▼                  ▼
        Native Rules        OPA / Rego        Blast Radius
              │                  │                  │
              └──────────────────┼──────────────────┘
                                 ▼
                    ┌──────────────────────────┐
                    │ Capability / Graph Diff  │
                    └────────────┬─────────────┘
                                 │
                                 ▼
                       Security Findings
                                 │
              ┌──────────────────┼──────────────────┐
              ▼                  ▼                  ▼
             CLI               SARIF              JSON
                                 │
                                 ▼
                           GitHub / CI/CD

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

📦 Package Responsibilities

| Package | Responsibility |

|---|---|

| cmd/raaya | CLI commands and user interaction |

| pkg/discovery | Hybrid discovery of AI assets, configurations, AST definitions, prompts, annotations, and runtime listeners |

| pkg/graph | Build and export the in-memory Security Graph |

| pkg/blastradius | Calculate transitive capability propagation and blast radius |

| pkg/analysis | Orchestrate security analysis and capability diffing |

| pkg/analysis/rules | Native Go security rules |

| pkg/analysis/policies | Embedded OPA/Rego policies |

| pkg/analysis/fixer | Safe automated remediation |

| pkg/reporter | Terminal, JSON, SARIF, and PR-oriented reporting |

| pkg/config | Local Raaya configuration |

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🎯 Design Principles

Local First

Raaya's core analysis runs locally.

There is no requirement for a hosted database or external security service.

Zero Configuration

The first scan should work without requiring developers to write a policy file.

raaya check

Hybrid Visibility

Security information can exist in source code, configuration, annotations, or the live environment.

Raaya combines these sources into one security model.

Graph-First Security

Rather than treating every finding as an isolated file-level issue, Raaya models the relationships between agents, servers, tools, capabilities, and resources.

Fast Feedback

Security analysis should fit naturally into local development, Git hooks, pull requests, and CI workflows.

Deterministic Analysis

Raaya focuses on observable configuration, source code, graph relationships, permissions, and runtime exposure rather than attempting to predict arbitrary LLM behavior.

Actionable Findings

A security finding should answer:

What was found?
Where was it found?
What capability or exposure is involved?
Why does it matter?
How can I fix it?

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🗺️ Roadmap

Raaya is being developed around three progressively deeper capabilities:

DISCOVER
   ↓
UNDERSTAND
   ↓
CONTROL

Phase 1 — Asset & Security-Surface Discovery

Current foundation:

Zero-config discovery

MCP configuration discovery

Python / TypeScript tool discovery

Prompt and agent discovery

Local listener discovery

Credential detection

Native security rules

OPA/Rego policy evaluation

Terminal diagnostics

JSON output

SARIF output

Differential finding analysis

Safe remediation

Git hooks

Goal:

Understand the AI application's security surface within seconds.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

Phase 2: Hybrid Parsing & PR Capability Diff (The Retention Engine)

Hybrid Parsing (Schema + AST): Acknowledging that AST alone isn't enough saves you months of refactoring. Combining explicit JSON manifests with code ASTs gives you maximum coverage with fewer false negatives.

PR Capability Diff: This is your killer feature. Commenting on a GitHub PR with “⚠️ This PR expands Agent-Beta's reachability to MCP-Database-Write” is the exact moment an engineering manager says, “We need to install this across all our repos.”

Goal:

Make security-surface changes visible directly in the pull request where the change is being reviewed.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

Phase 3: Developer-First YAML Policy (The Scale Factor)

YAML Over Rego: Making Semgrep-style YAML the default policy format lowers the barrier to entry by 90%. Rego as an optional enterprise adapter keeps the deeper policy engine available for organizations that need it.

Goal:

Define what AI agents are allowed to reach with developer-friendly policies and enforce those rules automatically.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🧪 Development

Run tests:

go test ./...

Run static analysis:

go vet ./...

Build:

go build ./...

Run locally:

go run ./cmd/raaya doctor
go run ./cmd/raaya check

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🤝 Contributing

Contributions are welcome.

Useful areas include:

MCP discovery

Agent/tool parsers

Security rules

Capability modeling

Security Graph construction

Graph exporters

Blast-radius analysis

Capability diffing

Rego policies

SARIF integration

GitHub PR reporting

Remediation

CI integrations

Test fixtures

Documentation

Before opening a pull request:

go test ./...
go vet ./...
go build ./...

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

📄 License

License information will be added here.

――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

🛡️ Raaya

Discover the AI security surface. Understand the blast radius. Detect capability expansion. Control what your agents can reach.

raaya doctor
raaya check
raaya graph
raaya diff

Raaya — Security Reachability Scanner

