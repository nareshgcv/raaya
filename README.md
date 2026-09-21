# 🦍 raaya

> ** Raaya Engine** — Local-first security & dependency graph analysis for AI agents, MCP configurations, and tool annotations.

`raaya` parses your repository’s AI assets—Model Context Protocol (MCP) servers, agent prompt definitions, tool declarations, and runtime capabilities—and builds a light-in-memory **Security Graph**. It evaluates policy checks (via OPA/Rego and native static rules) completely offline with zero database dependencies.

---

## 🎯 Key Features

- ⚡ **Zero Infrastructure Required**: Runs 100% locally in CLI without external backends or databases.
- 🕸️ **In-Memory Graph Representation**: Builds a DAG (Directed Acyclic Graph) of agent capabilities, MCP servers, and tool permissions.
- 🛡️ **Embedded Policy Enforcement**: Runs local Rego policies against the computed security graph.
- 📊 **SARIF & CI/CD Native**: Emits standard SARIF reports for GitHub Code Scanning/GitLab Security Dashboard, or formatted CLI visualizations.

---

## 🚀 Quick Start

### Installation

```bash
# Install via Go
go install [github.com/raaya/ape@latest](https://github.com/raaya/ape@latest)

# Or build locally
git clone [https://github.com/raaya/ape.git](https://github.com/raaya/ape.git)
cd ape
go build -o raaya ./cmd/raaya # raaya
