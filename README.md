# 🛡️ Raaya
### Static & Dynamic Security Reachability Scanner for AI Agents and MCP Infrastructure

**Raaya** is a lightweight, single-binary security scanner written in Go designed to analyze the security exposure of AI agent ecosystems and Model Context Protocol (MCP) integrations.

It unifies source-code analysis, MCP configuration discovery, runtime inspection, and security-graph topology to answer three critical questions:
1. *What can this AI agent reach?*
2. *What capabilities can flow through its MCP integrations?*
3. *What happens if one of those components is compromised?*

```text
  ┌───────┐      ┌────────────┐      ┌──────┐      ┌────────────┐      ┌──────────┐
  │ Agent │ ───► │ MCP Server │ ───► │ Tool │ ───► │ Capability │ ───► │ Resource │
  └───────┘      └────────────┘      └──────┘      └────────────┘      └──────────┘
```
___
