# A2A Agent Delegation

Agent-to-Agent delegation protocol.

```bash
pidl generate a2a_agent_delegation
```

## Overview

A2A (Agent-to-Agent) enables agents to delegate tasks to other agents, with optional human authorization.

## Entities

| Entity | Type | Role |
|--------|------|------|
| `user` | user | Human authorizing delegation |
| `primary_agent` | agent | Agent requesting delegation |
| `delegated_agent` | delegated_agent | Agent receiving delegated task |

## Flow

1. **Delegation Request**: Primary agent requests task delegation
2. **Human Authorization**: User authorizes the delegation (optional)
3. **Task Execution**: Delegated agent executes task
4. **Result Return**: Result returned to primary agent

## Entity Types

### agent

The primary AI agent initiating interactions.

### delegated_agent

An agent that receives delegated tasks from other agents. Has limited scope based on delegation permissions.

## Example

```json
{
  "protocol": {
    "id": "a2a-agent-delegation",
    "name": "A2A Agent Delegation",
    "category": "agent"
  },
  "entities": [
    {"id": "user", "name": "User", "type": "user"},
    {"id": "primary", "name": "Primary Agent", "type": "agent"},
    {"id": "delegated", "name": "Delegated Agent", "type": "delegated_agent"}
  ],
  "phases": [
    {"id": "auth", "name": "Authorization"},
    {"id": "exec", "name": "Execution"}
  ],
  "flows": [
    {
      "from": "primary",
      "to": "user",
      "action": "request_delegation",
      "phase": "auth"
    },
    {
      "from": "user",
      "to": "primary",
      "action": "authorize",
      "mode": "interactive",
      "phase": "auth"
    },
    {
      "from": "primary",
      "to": "delegated",
      "action": "delegate_task",
      "phase": "exec"
    },
    {
      "from": "delegated",
      "to": "primary",
      "action": "task_result",
      "mode": "response",
      "phase": "exec"
    }
  ]
}
```

## With Alternatives

Handling delegation rejection:

```json
{
  "from": "primary",
  "to": "user",
  "action": "request_delegation",
  "alternatives": [
    {
      "condition": "rejected",
      "flows": [
        {"from": "user", "to": "primary", "action": "rejection", "mode": "response"}
      ]
    }
  ]
}
```

## Generate Diagrams

```bash
# Mermaid
pidl generate -f mermaid a2a_agent_delegation

# D2 Architecture
pidl generate -f d2-arch a2a_agent_delegation
```
