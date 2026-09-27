

# Raaya Policy Engine

> **Local-first security & dependency graph analysis for AI agents, MCP configurations, and tool annotations.**

Raaya parses your repository’s AI assets—Model Context Protocol (MCP) servers, agent prompt definitions, tool declarations, and runtime capabilities—and builds an in-memory **Security Graph**. It evaluates policy checks completely offline with zero database or external cloud dependencies.

```text
 ┌─────────────────┐       ┌──────────────────┐       ┌─────────────────┐
 │ Agent Prompts   │ ────> │  Security Graph  │ ────> │ Policy Engine   │
 │ MCP Configs     │       │  (In-Memory DAG) │       │ (SARIF/Terminal)│
 │ Code Decorators │ ────> └──────────────────┘ ────> └─────────────────┘
 └─────────────────┘
🎯 Key Features
- ⚡ Zero-Config & Offline: Runs 100% locally in your terminal or CI runner with zero database or server dependencies.

- 🕸️ In-Memory Security Graph: Constructs a Directed Acyclic Graph (DAG) mapping AGENT → MODEL → SERVER → TOOL relationships and capabilities.

- 🛡️ Static Policy Enforcement: Built-in checks for hardcoded secrets, unauthenticated HTTP/SSE endpoints, prompt/tool capability mismatches, over-permissioned scopes, and orphaned tools.

- 🛠️ Automated Remediation: Auto-extract plaintext credentials directly into .env files via raaya check --fix.

- 📊 Terminal & CI/CD Native: Renders formatted terminal output with code snippets, ASCII/Mermaid dependency graphs, or standard SARIF output for GitHub Code Scanning.

---

## 🚀 Quick Start

### Installation

```bash
# Install via Go
go install [github.com/raaya/cmd/raaya@latest](https://github.com/raaya/cmd/raaya@latest)

# Or build locally
git clone [https://github.com/raaya/raaya.git](https://github.com/raaya/raaya.git)
cd raaya
go build -o raaya ./cmd/raaya

🚀 Quick Start
1. Run Workspace Diagnostics
Verify your local environment, discovered MCP configuration files, and prompt assets in under 5 seconds:

```bash
raaya doctor


2. Perform Static Policy Analysis
Scan the workspace for credential exposure, misconfigured MCP endpoints, and orphaned capability declarations:

```bash
raaya check

Automatically extract hardcoded API secrets into .env files:
```bash
raaya check --fix

Output standard SARIF for GitHub Actions or GitLab CI integration:

```bash
raaya check --format sarif > results.sarif

3. Visualize Agent & Tool Dependency Graphs
Render an ASCII visual tree of all agents, MCP servers, and referenced tools directly in your terminal:

```bash
raaya graph

Export a Mermaid.js diagram for markdown documentation or GitHub PR comments:

```bash
raaya graph --format mermaid

4. Install Git Pre-Commit Hook
Prevent unencrypted credentials or policy violations from reaching source control:

```bash
raaya hook install

🧠 How It Works

📁 # Raaya Policy Engine

> **Local-first security & dependency graph analysis for AI agents, MCP configurations, and tool annotations.**

Raaya parses your repository’s AI assets—Model Context Protocol (MCP) servers, agent prompt definitions, tool declarations, and runtime capabilities—and builds an in-memory **Security Graph**. It evaluates policy checks completely offline with zero database or external cloud dependencies.

```text
 ┌─────────────────┐       ┌──────────────────┐       ┌─────────────────┐
 │ Agent Prompts   │ ────> │  Security Graph  │ ────> │ Policy Engine   │
 │ MCP Configs     │       │  (In-Memory DAG) │       │ (SARIF/Terminal)│
 │ Code Decorators │ ────> └──────────────────┘ ────> └─────────────────┘
 └─────────────────┘


**Built-in Security checks**

| Rule ID | Rule Name | Severity | Default Action | Target Asset | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **RAA001** | `PromptToolMismatch` | **Error** | Block Commit | `.prompt`, `system_prompt.txt` | Prompt references a tool name or function call that is not declared in any MCP config or source decorator. |
| **RAA002** | `UnauthenticatedEndpoint` | **Error** | Block Commit | `mcp_config.json` | Remote MCP HTTP or SSE transport server defined without authorization headers or access tokens. |
| **RAA003** | `OverPermissionedScope` | **Warning** | Flag / Warn | `mcp_config.json` | MCP server granted unrestricted filesystem write access, root privileges, or wildcard execution scopes. |
| **RAA004** | `PlaintextSecret` | **Error** | Auto-Fix (`--fix`) | `mcp_config.json`, `.env` | Hardcoded API keys, bearer tokens, or database credentials detected in environment variable declarations. |
| **RAA005** | `OrphanedCapability` | **Warning** | Flag / Warn | Python/JS/TS Decorators | Tool or MCP server declared in source code or config but never invoked or referenced by any registered agent. |

