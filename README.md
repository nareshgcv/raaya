# 🛡️ Raaya

**Static Security & Blast Radius Analysis for AI Agents and MCP Infrastructure**

Raaya is a lightweight, single-binary security scanner built in Go for analyzing the security surface of AI agents and Model Context Protocol (MCP) infrastructure.

It discovers AI-agent configurations, MCP servers, tool definitions, prompts, local listeners, and credentials, then builds an in-memory **Security Graph** to identify security findings and calculate transitive capability exposure.

Raaya is designed to run **locally, quickly, and without external infrastructure**.

> **Discover what your AI agent can access before it reaches production.**

---

## ✨ Features

### 🔍 Multi-Source Discovery

Raaya automatically discovers AI security assets from multiple sources:

* `.cursor/mcp.json`
* `mcp.json`
* MCP server configurations
* Python tool definitions
* TypeScript tool definitions
* Agent/prompt references
* Environment and credential configuration
* Local TCP/SSE listeners

Supported Python patterns include:

```python
@mcp.tool
def search(...):
    ...
```

and:

```python
@tool
def execute(...):
    ...
```

---

### 🌐 Active Listener Discovery

Raaya can inspect locally running services:

```bash
raaya doctor --live
```

It probes supported local TCP/SSE endpoints and analyzes discovered MCP listeners for potentially unauthenticated exposure.

This allows Raaya to detect security issues that may not be obvious from static configuration alone.

---

### 🕸️ Security Graph

Raaya converts discovered AI infrastructure into an in-memory directed graph.

Conceptually:

```text
Agent
  │
  ▼
MCP Server
  │
  ▼
Tool
  │
  ├────────► Filesystem
  ├────────► Network
  ├────────► Database
  └────────► Secrets
```

The graph provides the foundation for security analysis and transitive capability analysis.

Graph output can be exported for visualization:

```bash
raaya graph
```

Mermaid:

```bash
raaya graph --format mermaid
```

DOT:

```bash
raaya graph --format dot
```

---

### 💥 Transitive Blast Radius Analysis

Raaya analyzes graph reachability to answer questions such as:

> **If an agent, prompt, or MCP server is compromised, what capabilities and resources can it reach?**

Example:

```text
research-agent
      │
      ▼
github-mcp
      │
      ▼
github_search
      │
      ▼
network
```

The blast-radius engine provides the foundation for identifying transitive relationships between agents, tools, capabilities, resources, and sensitive assets.

---

### 🛡️ Dual Security Rule Engine

Raaya combines two analysis mechanisms:

```text
                 Security Graph
                       │
              ┌────────┴────────┐
              ▼                 ▼
        Native Go Rules      OPA / Rego
              │                 │
              └────────┬────────┘
                       ▼
                   Findings
```

#### Native Go Rules

High-confidence built-in security checks are implemented under:

```text
pkg/analysis/rules/
```

These rules provide fast, deterministic checks for common AI-agent and MCP security issues.

#### OPA / Rego Policies

Raaya also supports graph-based policy evaluation through embedded OPA/Rego policies.

Policies are stored under:

```text
pkg/analysis/policies/
```

This allows more complex relationships and organization-specific security requirements to be evaluated against the Security Graph.

---

# 🔐 Detection Rules

The current Phase 1 implementation includes native security checks and Rego-based policy enforcement.

| **Rule ID**                 | **Name**               |       **Severity** | **Description**                                                                                                 |
| --------------------------- | ---------------------- | -----------------: | --------------------------------------------------------------------------------------------------------------- |
| `RAAYA-001-SECRETS`         | Hardcoded Credentials  |       **CRITICAL** | Flags plaintext API keys, tokens, passwords, and other credential-like values in supported configuration files. |
| `RAAYA-002-UNAUTH-ENDPOINT` | Unauthenticated Server |           **HIGH** | Detects HTTP/SSE MCP servers running without identifiable authentication headers, credentials, or secrets.      |
| `RAAYA-REGO-001`            | Rego Policy Violation  | **Policy-defined** | Reports violations produced by the embedded OPA/Rego security policies evaluated against the Security Graph.    |

### `RAAYA-001-SECRETS`

Detects credential-like values embedded directly in supported AI/MCP configuration files.

Example:

```json
{
  "apiKey": "sk-live-xxxxxxxx"
}
```

Raaya reports the location and provides a remediation path where supported.

Run automatic remediation with:

```bash
raaya check --fix
```

For supported findings, Raaya can extract credentials into an environment file and verify that the file is protected by `.gitignore`.

---

### `RAAYA-002-UNAUTH-ENDPOINT`

Detects potentially unauthenticated MCP HTTP/SSE endpoints.

For example:

```text
MCP Server
    │
    ├── Transport: SSE
    ├── Address: 0.0.0.0:3000
    └── Authentication: None
```

Raaya reports the endpoint as a high-severity security finding.

Live listener discovery can be enabled with:

```bash
raaya doctor --live
```

---

### `RAAYA-REGO-001`

Rego policy violations are evaluated against the Security Graph.

This enables policies to reason about relationships rather than isolated files.

Conceptually:

```text
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
```

A policy can evaluate properties and relationships across those graph nodes.

The Rego engine is implemented in:

```text
pkg/analysis/rego_engine.go
```

Embedded policies are located in:

```text
pkg/analysis/policies/
```

---

# 📈 Differential Analysis

Raaya includes a differential baseline comparator:

```text
pkg/analysis/delta.go
```

This allows CI workflows to distinguish existing findings from newly introduced findings.

Conceptually:

```text
Baseline
   │
   ├── Existing findings ──────► Existing
   │
   └── New findings ───────────► New
                                      │
                                      ▼
                                 CI decision
```

This makes it possible to introduce Raaya into an existing repository without requiring every historical finding to be fixed immediately.

---

# 🛠️ Automated Remediation

Raaya includes a remediation engine:

```text
pkg/fixer/
```

For supported credential findings, Raaya can:

1. Detect hardcoded credentials
2. Extract the value into an environment file
3. Replace the original configuration value
4. Verify `.gitignore` protection
5. Report the resulting changes

Run:

```bash
raaya check --fix
```

Raaya only applies deterministic remediation where it can do so safely.

---

# 📊 CI/CD & SARIF

Raaya is designed to integrate into existing security pipelines.

Generate JSON:

```bash
raaya check --format json
```

Generate SARIF:

```bash
raaya check --format sarif > results.sarif
```

Raaya also supports differential analysis for CI/CD workflows.

This allows teams to focus on newly introduced security findings rather than being blocked immediately by historical findings.

---

# 🪝 Git Pre-Commit Hooks

Raaya can install a Git pre-commit hook:

```bash
raaya hook install
```

This allows security checks to run before changes are committed.

The objective is to catch AI-agent security issues as early as possible in the development lifecycle.

---

# 🚀 Quick Start

## Installation

Clone the repository:

```bash
git clone https://github.com/your-org/raaya.git
cd raaya
```

Build the binary:

```bash
go build -o raaya ./cmd/raaya
```

Run discovery:

```bash
./raaya doctor
```

Run the security scan:

```bash
./raaya check
```

---

# ⚡ 30-Second Workflow

```text
        ┌─────────────┐
        │  Repository │
        └──────┬──────┘
               │
               ▼
       ┌───────────────┐
       │ raaya doctor  │
       │   Discovery   │
       └───────┬───────┘
               │
               ▼
       ┌───────────────┐
       │  raaya check  │
       │ Security Scan │
       └───────┬───────┘
               │
        ┌──────┴──────┐
        ▼             ▼
     Findings       Graph
        │             │
        ▼             ▼
       Fix       Blast Radius
        │
        ▼
       CI/CD
```

---

# 🩺 `raaya doctor`

Use `doctor` to inspect what Raaya discovers in the current workspace.

```bash
raaya doctor
```

For live listener discovery:

```bash
raaya doctor --live
```

Example:

```text
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

✓ Environment configuration
  .env

Discovery complete.
```

---

# 🔎 `raaya check`

Run the security analysis:

```bash
raaya check
```

Example:

```text
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
  the endpoint exposure.

────────────────────────────────────────────

✗ RAAYA-REGO-001
  Security Graph

  Policy violation detected by
  embedded Rego policy.

Scan complete.

3 findings
```

The intended diagnostic structure is:

```text
Rule
  ↓
Location
  ↓
Problem
  ↓
Security context
  ↓
Suggested remediation
```

---

# 🏗️ Architecture

Raaya is organized into focused analysis components:

```text
raaya/
├── cmd/raaya/
│   └── CLI entry point & subcommands
│
├── pkg/
│   │
│   ├── analysis/
│   │   ├── engine.go
│   │   ├── rego_engine.go
│   │   ├── delta.go
│   │   │
│   │   ├── policies/
│   │   │   ├── default.rego
│   │   │   └── embed.go
│   │   │
│   │   └── rules/
│   │       ├── secrets.go
│   │       ├── unauth.go
│   │       └── rules.go
│   │
│   ├── blastradius/
│   │   └── Transitive graph reachability
│   │
│   ├── discovery/
│   │   └── AST, configuration,
│   │       prompt & port discovery
│   │
│   ├── fixer/
│   │   └── Credential remediation
│   │
│   ├── graph/
│   │   └── Security Graph & exporters
│   │
│   └── reporter/
│       └── SARIF reporting
```

### Analysis Flow

```text
Repository
    │
    ▼
┌──────────────────┐
│    Discovery     │
│                  │
│ AST              │
│ JSON Config      │
│ Prompts          │
│ Ports            │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Security Graph  │
│                  │
│ Agent            │
│ MCP Server       │
│ Tool             │
│ Capability       │
│ Resource         │
└────────┬─────────┘
         │
    ┌────┴─────┐
    ▼          ▼
Native Go    OPA/Rego
Rules        Policies
    │          │
    └────┬─────┘
         ▼
   Security Findings
         │
    ┌────┼────────┐
    ▼    ▼        ▼
   CLI  SARIF    JSON
         │
         ▼
      CI / Git
```

---

# 📦 Package Responsibilities

| Package                 | Responsibility                                                          |
| ----------------------- | ----------------------------------------------------------------------- |
| `cmd/raaya`             | CLI commands and user interaction                                       |
| `pkg/discovery`         | Discover AI assets, configurations, AST definitions, prompts, and ports |
| `pkg/graph`             | Build and export the in-memory Security Graph                           |
| `pkg/blastradius`       | Calculate transitive graph reachability                                 |
| `pkg/analysis`          | Orchestrate security analysis                                           |
| `pkg/analysis/rules`    | Native Go security rules                                                |
| `pkg/analysis/policies` | Embedded Rego policies                                                  |
| `pkg/fixer`             | Safe automated remediation                                              |
| `pkg/reporter`          | SARIF report generation                                                 |

---

# 🎯 Design Principles

### Local First

Raaya's core analysis runs locally.

There is no requirement for a hosted database or external security service.

### Zero Configuration

The first scan should work without requiring developers to write a policy file.

```bash
raaya check
```

### Fast Feedback

Security analysis should fit naturally into local development and CI workflows.

### Deterministic Analysis

Phase 1 focuses on observable configuration, code, graph relationships, and local runtime exposure rather than attempting to predict arbitrary LLM behavior.

### Actionable Findings

A security finding should answer:

```text
What was found?
Where was it found?
Why does it matter?
How can I fix it?
```

---

# 🗺️ Roadmap

Raaya is being developed around three progressively deeper capabilities:

```text
DISCOVER
   ↓
UNDERSTAND
   ↓
CONTROL
```

## Phase 1 — Instant Asset & Security-Surface Discovery

**Current implementation**

* Zero-config discovery
* MCP configuration discovery
* Python/TypeScript tool discovery
* Prompt/tool discovery
* Local listener discovery
* Credential detection
* Native security rules
* OPA/Rego policy evaluation
* Terminal diagnostics
* JSON output
* SARIF output
* Differential analysis
* Safe remediation
* Git hooks

**Goal:**

> Understand the AI application's security surface within seconds.

---

## Phase 2 — Security Graph & Capability Diff

Planned expansion of the existing graph and blast-radius foundations:

* MCP schema + configuration + AST hybrid analysis
* Agent → MCP Server → Tool → Resource reachability
* Transitive capability propagation
* Expanded blast-radius analysis
* Graph visualization
* PR capability diff
* GitHub PR comments
* Security-surface change detection

**Goal:**

> Understand how a code or configuration change changes the capabilities reachable by an AI agent.

---

## Phase 3 — Developer-First Policy

Planned policy experience:

* Declarative YAML rules
* Semgrep-style policy authoring
* Pre-built security policy bundles
* Organization-specific policies
* CI/CD enforcement
* Compliance reporting
* Optional OPA/Rego integration

The objective is to allow developers and security engineers to express policies in terms of **agents, tools, capabilities, resources, and reachability** without requiring deep knowledge of Rego.

OPA/Rego remains available as an integration path for organizations that already use OPA-based policy infrastructure.

**Goal:**

> Define what AI agents are allowed to reach — and enforce it automatically.

---

# 🧪 Development

Run tests:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

Build:

```bash
go build ./...
```

Run locally:

```bash
go run ./cmd/raaya doctor
go run ./cmd/raaya check
```

---

# 🤝 Contributing

Contributions are welcome.

Useful areas include:

* MCP discovery
* Agent/tool parsers
* Security rules
* Graph analysis
* Blast-radius analysis
* Rego policies
* SARIF integration
* Remediation
* CI integrations
* Test fixtures
* Documentation

Before opening a pull request:

```bash
go test ./...
go vet ./...
go build ./...
```

---

# 📄 License

License information will be added here.

---

# 🛡️ Raaya

**Discover the AI security surface. Understand the blast radius. Control what your agents can reach.**

```bash
raaya doctor
raaya check
```
