# MCP Tool Invocation

Model Context Protocol tool invocation flow.

```bash
pidl generate mcp_tool_invocation
```

## Overview

MCP (Model Context Protocol) enables AI agents to invoke tools provided by tool servers.

## Entities

| Entity | Type | Role |
|--------|------|------|
| `agent` | agent | AI/LLM agent |
| `tool_server` | tool_server | Exposes tools via MCP |

## Flow

1. **Tool Discovery**: Agent discovers available tools
2. **Tool Invocation**: Agent calls a tool
3. **Tool Result**: Server returns result

## Flow Modes

| Mode | Use |
|------|-----|
| `tool_call` | Agent invoking a tool |
| `tool_result` | Server returning result |

## Example

```json
{
  "protocol": {
    "id": "mcp-tool-invocation",
    "name": "MCP Tool Invocation",
    "category": "agent"
  },
  "entities": [
    {"id": "agent", "name": "AI Agent", "type": "agent"},
    {"id": "tool_server", "name": "Tool Server", "type": "tool_server"}
  ],
  "flows": [
    {
      "from": "agent",
      "to": "tool_server",
      "action": "list_tools",
      "mode": "request"
    },
    {
      "from": "tool_server",
      "to": "agent",
      "action": "tools_list",
      "mode": "response"
    },
    {
      "from": "agent",
      "to": "tool_server",
      "action": "invoke_tool",
      "mode": "tool_call"
    },
    {
      "from": "tool_server",
      "to": "agent",
      "action": "tool_result",
      "mode": "tool_result"
    }
  ]
}
```

## Generate Diagrams

```bash
# Mermaid
pidl generate -f mermaid mcp_tool_invocation

# D2
pidl generate -f d2 mcp_tool_invocation
```
