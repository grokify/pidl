# Phases

Phases provide logical grouping of flows for readability and diagram organization.

## Basic Phases

```json
{
  "phases": [
    {
      "id": "authorization",
      "name": "Authorization",
      "description": "User authentication and consent"
    },
    {
      "id": "token_exchange",
      "name": "Token Exchange",
      "description": "Exchange code for tokens"
    }
  ]
}
```

## Phase Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `name` | string | Yes | Human-readable name |
| `description` | string | No | Phase description |
| `parent` | string | No | Parent phase ID for nesting |

## Nested Phases

Phases support hierarchical nesting via the `parent` field:

```json
{
  "phases": [
    {"id": "auth", "name": "Authentication"},
    {"id": "login", "name": "Login", "parent": "auth"},
    {"id": "mfa", "name": "Multi-Factor Auth", "parent": "auth"},
    {"id": "token", "name": "Token Exchange"}
  ]
}
```

### Hierarchy Rules

- A phase with no `parent` is a root phase
- A phase with a `parent` must reference a valid phase ID
- Circular references are not allowed
- A phase cannot be its own parent

## Assigning Flows to Phases

Reference the phase ID in flows:

```json
{
  "flows": [
    {"from": "client", "to": "server", "action": "login", "phase": "login"},
    {"from": "server", "to": "client", "action": "mfa_challenge", "phase": "mfa"},
    {"from": "client", "to": "server", "action": "token_request", "phase": "token"}
  ]
}
```

## Diagram Rendering

| Format | Rendering |
|--------|-----------|
| PlantUML | Colored `box` containers |
| Mermaid | Colored `rect` blocks |
| D2 | Nested groups |
| DOT | Subgraphs (if enabled) |

## Helper Methods (Go)

```go
// Get root phases (no parent)
roots := protocol.RootPhases()

// Get children of a phase
children := protocol.ChildPhases("auth")

// Get nesting depth (0 for root)
depth := protocol.PhaseDepth("mfa")  // returns 1 if parent is "auth"
```
