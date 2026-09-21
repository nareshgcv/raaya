# 🦍 raaya

>**Raaya Policy Engine** — Local-first security & dependency graph analysis for AI agents, MCP configurations, and tool annotations.

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



Basic Usage
1. Scan a Repository
Scan your local project directory for agent tools, MCP server setups, and prompt dependencies:

raaya scan .

2. Visualise the Dependency Graph
Render an ASCII summary or graph representation of how tools and agents interact in terminal:
raaya graph .

3. Export SARIF for GitHub / GitLab CI
Generate standard SARIF outputs for automated security scanning:

raaya scan . --format=sarif --output=results.sarif

🧠 How It Works

📁 Repo Files         🔍 In-Memory Engine            🛡️ Policy Rule Engine            📊 Output
+------------------+   +------------------------+    +--------------------------+    +------------------+
| mcp.json         |   |                        |    |                          |    | Terminal Summary |
| tool_defs.py     |-->| SecurityGraph Builder  |--->| OPA / Rego Evaluator     |--->| SARIF Report     |
| agent_prompts.md |   | (Nodes & Capabilities) |    | Native Rule Checking     |    | Terminal Graph   |
+------------------+   +--------------------- +  +--------------------------+    +------------------+ 

1.Graph Construction: ape parses configurations (mcp.json, agent descriptors, tool code annotations) into an internal SecurityGraph.

2.Local Evaluation: Evaluates embedded Rego policies (or custom .rego files provided locally) directly against the graph using an embedded Open Policy Agent runtime.

3.Structured Reporting: Outputs actionable violations with context and risk metrics.
