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

Raaya calculates transitive capability propagation, evaluates custom security policies, measures blast radius, and compares security-graph states across Git references to detect newly introduced security-surface expansions before they land in production.

✨ Key Features
🔍 Hybrid Discovery Engine
Raaya avoids relying on a single source of truth. It combines static AST parsing, configuration file scanning, and live endpoint probing into a single discovery pipeline:
```text
┌─────────────────────────┐
                       │   Hybrid Discovery      │
                       └────────────┬────────────┘
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         ▼                          ▼                          ▼
  ┌─────────────┐            ┌─────────────┐            ┌──────────────┐
  │ Source AST  │            │ MCP Configs │            │ Live Ports   │
  └──────┬──────┘            └──────┬──────┘            └──────┬───────┘
         │                          │                          │
         └──────────────────────────┼──────────────────────────┘
                                    ▼
                         ┌─────────────────────┐
                         │ Discovered Assets   │
                         └─────────────────────┘

```
## AST Parsing:## 
Python & TypeScript tool definitions, agent/prompt references, and @raaya:capability inline annotations.

## Configuration Scanning:## .cursor/mcp.json, mcp.json, claude_desktop_config.json, and environment credential files (.env).

## Runtime Inspection:##
Local TCP listeners and live SSE MCP endpoints.

🕸️ Security Graph
Discovered assets are mapped into an in-memory directed Security Graph:

