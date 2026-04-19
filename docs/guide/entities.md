# Entities

Entities represent participants in the protocol: systems, services, actors, or trust domains.

## Definition

```json
{
  "entities": [
    {
      "id": "client",
      "name": "Client Application",
      "type": "client",
      "description": "Application requesting access to protected resources"
    },
    {
      "id": "auth_server",
      "name": "Authorization Server",
      "type": "authorization_server"
    }
  ]
}
```

## Entity Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier (lowercase, alphanumeric, underscores) |
| `name` | string | Yes | Human-readable display name |
| `type` | enum | Yes | Entity type classification |
| `description` | string | No | Role description |

## Entity Types

### OAuth/OIDC Types

| Type | Description |
|------|-------------|
| `client` | Application or service initiating requests |
| `authorization_server` | Issues tokens and handles authentication |
| `resource_server` | Hosts protected resources |
| `user` | Human actor / resource owner |
| `browser` | User agent / web browser |

### Agent Protocol Types

| Type | Description |
|------|-------------|
| `agent` | AI/LLM agent |
| `tool_server` | Exposes tools via protocol (MCP) |
| `tool` | Individual tool |
| `delegated_agent` | Agent receiving delegated tasks (A2A) |

### Identity Types

| Type | Description |
|------|-------------|
| `identity_provider` | Authenticates users (SAML IdP) |
| `service_provider` | Relies on identity provider (SAML SP) |

### Generic Types

| Type | Description |
|------|-------------|
| `server` | Generic server |
| `other` | Custom entity type |

## ID Patterns

Entity IDs must match: `^[a-z][a-z0-9_]*$`

Valid: `client`, `auth_server`, `tool_1`

Invalid: `Client`, `auth-server`, `1tool`

## Diagram Rendering

Entity types affect how they're rendered in diagrams:

- **D2**: Different shapes (person, hexagon, cylinder, etc.)
- **DOT**: Different node shapes
- **PlantUML/Mermaid**: Participant declarations
