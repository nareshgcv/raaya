# Raaya Policy Engine

> **Local-first security & dependency graph analysis for AI agents, MCP configurations, and tool annotations.**

Raaya scans your repository's AI assets—**MCP servers, agent prompts, tool declarations, and runtime capabilities**—and builds an in-memory **Security Graph**.

It then evaluates security policies completely offline, helping developers detect dangerous agent capabilities, configuration mistakes, exposed credentials, and permission problems **before they reach production**.

No database. No cloud service. No external security backend.

```text
┌─────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│ Agent Prompts   │       │                  │       │                  │
│ MCP Configs     │ ────> │  Security Graph  │ ────> │  Policy Engine   │
│ Tool Definitions│       │  (In-Memory DAG) │       │                  │
│ Code Decorators │       │                  │       │                  │
└─────────────────┘       └──────────────────┘       └────────┬─────────┘
                                                              │
                                           ┌──────────────────┼─────────────────┐
                                           ▼                  ▼                 ▼
                                      Terminal             SARIF             CI/CD
```

## 🎯 Key Features

* ⚡ **Zero-Config & Offline** — Run entirely locally in your terminal or CI runner with no database, server, or cloud dependency.
* 🕸️ **In-Memory Security Graph** — Maps `AGENT → MODEL → SERVER → TOOL` relationships and capabilities.
* 🛡️ **Built-in Security Policies** — Detect hardcoded secrets, unauthenticated endpoints, capability mismatches, excessive permissions, and unused capabilities.
* 🛠️ **Automated Remediation** — Safely extract supported plaintext credentials into `.env` files with `raaya check --fix`.
* 📊 **Developer & CI/CD Native** — Human-readable terminal diagnostics plus SARIF output for GitHub Code Scanning and other CI security systems.
* 🔎 **Dependency Visualization** — Inspect agent, MCP server, and tool relationships directly from the terminal or export them as Mermaid diagrams.

---

# 🚀 Quick Start

## 1. Install Raaya
One-command installation

The fastest way to try Raaya:

npx raaya

Or install it with Homebrew:

brew install raaya

### Install with Go

```bash
go install github.com/raaya/raaya/cmd/raaya@latest
```

### Build locally

```bash
git clone https://github.com/raaya/raaya.git
cd raaya
go build -o raaya ./cmd/raaya
```

Verify the installation:

```bash
raaya --version
```

---

## 2. Run Workspace Diagnostics

Start with `doctor` to verify your environment and discover supported AI assets.

```bash
raaya doctor
```

Raaya reports discovered:

* MCP configuration files
* Agent prompt definitions
* Tool declarations
* Supported source-code annotations
* Runtime capability definitions

Example:

```text
$ raaya doctor

Raaya Workspace Diagnostics

✓ Workspace detected
✓ 3 MCP configurations found
✓ 5 agent definitions found
✓ 27 tool declarations found
✓ 2 source decorators discovered

Ready to run security checks.
```

---

## 3. Run Static Security Analysis

Run the built-in policy engine:

```bash
raaya check
```

Raaya scans the workspace for security and configuration problems.

Example:

```text
$ raaya check

Raaya Security Scan

✗ RAA001  ERROR
  agents/researcher.prompt:18

  Tool "github_search" is referenced by the agent
  but is not declared by any MCP server or tool definition.

  Fix:
  Remove the reference or declare the tool.

✗ RAA002  ERROR
  mcp/github.json:12

  Remote MCP endpoint does not define authentication.

  Fix:
  Configure authorization headers or access tokens.

⚠ RAA003  WARNING
  mcp/filesystem.json:24

  Tool has unrestricted filesystem write access.

  Fix:
  Restrict the filesystem scope to required paths.

──────────────────────────────────────────────

2 errors · 1 warning

Exit code: 1
```

### Automatically remediate supported findings

For supported secret findings, Raaya can extract credentials into `.env`:

```bash
raaya check --fix
```

> **Note:** `--fix` should only modify findings for which Raaya can perform a deterministic and safe transformation. Destructive or ambiguous changes should remain suggestions rather than automatic fixes.

---

## 4. Generate SARIF for CI/CD

Export findings as standard SARIF:

```bash
raaya check --format sarif > results.sarif
```

SARIF can be consumed by security tooling such as GitHub Code Scanning and compatible CI/CD platforms.

This allows Raaya to move from a local developer check to a **pre-merge security gate**.

---

## 5. Visualize Agent & Tool Dependencies

Inspect the discovered Security Graph:

```bash
raaya graph
```

Example:

```text
Agent: researcher
│
├── Model: gpt-*
│
├── MCP: github
│   ├── search
│   └── issues
│
└── MCP: filesystem
    ├── read
    └── write
```

Export the graph as Mermaid:

```bash
raaya graph --format mermaid
```

This can be used in Markdown documentation, architecture reviews, and pull requests.

---

## 6. Install the Git Pre-Commit Hook

Prevent supported policy violations from being committed:

```bash
raaya hook install
```

After installation, Raaya can run checks automatically before commits are created.

```text
git commit
    │
    ▼
Raaya pre-commit check
    │
    ├── ✓ No policy violations
    │       ↓
    │     Commit
    │
    └── ✗ Policy violation
            ↓
          Commit blocked
```

---

# 🛡️ Built-in Security Checks

Raaya ships with native checks that work without requiring an OPA/Rego setup.

| Rule ID    | Rule Name                 | Severity    | Default Action     | Target Asset                   | Description                                                                                                                    |
| ---------- | ------------------------- | ----------- | ------------------ | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| **RAA001** | `PromptToolMismatch`      | **Error**   | Block Commit       | `.prompt`, `system_prompt.txt` | Detects prompts that reference a tool or function that is not declared by an MCP configuration or supported source definition. |
| **RAA002** | `UnauthenticatedEndpoint` | **Error**   | Block Commit       | `mcp_config.json`              | Detects remote MCP HTTP/SSE endpoints without configured authentication.                                                       |
| **RAA003** | `OverPermissionedScope`   | **Warning** | Flag / Warn        | `mcp_config.json`              | Detects unrestricted filesystem access, root privileges, wildcard execution scopes, and other excessive capabilities.          |
| **RAA004** | `PlaintextSecret`         | **Error**   | Auto-Fix (`--fix`) | MCP/config files               | Detects hardcoded API keys, bearer tokens, and database credentials in supported configuration files.                          |
| **RAA005** | `OrphanedCapability`      | **Warning** | Flag / Warn        | Python/JS/TS decorators        | Detects tools or MCP capabilities that are declared but not referenced by any registered agent.                                |

### Example finding

```text
RAA003  WARNING

agents/researcher.yaml:31

Agent has unrestricted filesystem write access.

Capability path:

researcher
  └── filesystem
      └── write *
```

The graph context makes the finding more useful than a simple line-based configuration warning: developers can see **which agent can reach which capability and through which MCP server**.

---

# 🧠 How It Works

Raaya follows a simple local analysis pipeline:

```text
┌─────────────────────┐
│ Repository          │
│                     │
│ Agent Prompts       │
│ MCP Configs         │
│ Tool Definitions    │
│ Code Decorators     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Security Graph      │
│                     │
│ AGENT               │
│   ↓                 │
│ MODEL               │
│   ↓                 │
│ SERVER              │
│   ↓                 │
│ TOOL / CAPABILITY   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Policy Engine       │
│                     │
│ Native Rules        │
│ OPA / Rego          │
│                     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Results             │
│                     │
│ Terminal            │
│ SARIF               │
│ Graph / Mermaid     │
└─────────────────────┘
```

Everything required for the static analysis runs locally.

---

# 🔐 Why Raaya?

AI applications increasingly combine agents, models, MCP servers, tools, credentials, and runtime capabilities.

A traditional source-code scanner may identify a hardcoded credential.

A configuration scanner may identify an open endpoint.

But the security question is often **relational**:

> Which agent can reach which tool, through which server, with which permissions?

Raaya models those relationships as a Security Graph and evaluates policies against the resulting capability graph.

```text
Agent
  │
  ├── Model
  │
  ├── MCP Server
  │      │
  │      ├── Tool A
  │      ├── Tool B
  │      └── Tool C
  │
  └── Runtime Capabilities
```

This allows Raaya to detect risks that depend on the **relationship between multiple AI assets**, rather than examining each file in isolation.

---

# 🤖 Policy Engine

Raaya's built-in checks require no policy configuration.

For teams that need custom security requirements, Raaya can additionally evaluate policies using OPA/Rego.

This gives developers a simple progression:

```text
Zero Configuration
       │
       ▼
Built-in Rules
       │
       ▼
Custom Policies
       │
       ▼
CI/CD Enforcement
```

Developers can start with:

```bash
raaya check
```

and introduce custom policy enforcement later as their AI infrastructure grows.

---

# 📦 CI/CD Usage

A typical CI pipeline can run:

```bash
raaya check --format sarif > results.sarif
```

and fail the pipeline when configured severity thresholds are exceeded.

Recommended workflow:

```text
Developer
    │
    ▼
raaya check
    │
    ▼
Fix findings
    │
    ▼
Git commit
    │
    ▼
Pre-commit hook
    │
    ▼
Pull Request
    │
    ▼
CI / SARIF
    │
    ▼
Production
```

---

# 🗺️ Roadmap

### Phase 1 — Developer Experience

* [ ] Zero-config native security checks
* [ ] Human-friendly terminal diagnostics
* [ ] Stable rule IDs
* [ ] Line-level findings
* [ ] Actionable remediation
* [ ] SARIF output
* [ ] Git pre-commit integration

### Phase 2 — Security Graph

* [ ] Expanded MCP capability discovery
* [ ] Agent → model → server → tool relationship analysis
* [ ] Transitive capability analysis
* [ ] Risk propagation across dependency paths
* [ ] Improved graph visualization

### Phase 3 — Policy-as-Code

* [ ] OPA/Rego integration
* [ ] Custom organization policies
* [ ] Policy bundles
* [ ] CI policy enforcement
* [ ] Security policy testing

---

# ⚡ The 30-Second Workflow

```bash
# Install
go install github.com/raaya/raaya/cmd/raaya@latest

# Discover your AI assets
raaya doctor

# Find security and configuration issues
raaya check

# Inspect the capability graph
raaya graph

# Export findings for CI
raaya check --format sarif > results.sarif
```

**Raaya turns AI-agent configuration into a security graph you can inspect, analyze, and enforce—without sending your repository to a cloud service.**
