# PIDL Specification

**Protocol Interaction Description Language v1.0**

## Overview

PIDL (Protocol Interaction Description Language) is a JSON-based DSL for describing protocol choreography. It models protocols as directed interaction graphs between entities, enabling generation of sequence diagrams, data flow diagrams, and other visualizations.

PIDL is designed for protocols where the primary concern is "who talks to whom, in what order" rather than message schemas or transport details. Target protocols include OAuth 2.0, OpenID Connect, MCP (Model Context Protocol), and A2A (Agent-to-Agent).

## Design Principles

1. **Choreography-focused**: Models interactions between systems, not API endpoints
2. **Transport-agnostic**: No HTTP methods, URLs, or wire formats
3. **Diagram-first**: Optimized for generating visual representations
4. **JSON-native**: Simple parsing, strict validation, universal tooling

## Document Structure

A PIDL document is a JSON object with three required sections and one optional section:

```json
{
  "protocol": { ... },
  "entities": [ ... ],
  "phases": [ ... ],
  "flows": [ ... ]
}
```

| Section | Required | Description |
|---------|----------|-------------|
| `protocol` | Yes | Metadata about the protocol |
| `entities` | Yes | Participants in the protocol |
| `phases` | No | Logical grouping of flows |
| `flows` | Yes | Interactions between entities |

## Protocol Metadata

The `protocol` object contains metadata about the protocol being described.

```json
{
  "protocol": {
    "id": "oauth2-authorization-code",
    "name": "OAuth 2.0 Authorization Code Flow",
    "version": "1.0",
    "description": "OAuth 2.0 Authorization Code Grant as defined in RFC 6749",
    "category": "auth",
    "references": [
      {
        "name": "RFC 6749",
        "url": "https://datatracker.ietf.org/doc/html/rfc6749"
      }
    ]
  }
}
```

### Protocol Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier (lowercase, alphanumeric, hyphens) |
| `name` | string | Yes | Human-readable name |
| `version` | string | No | Version of this protocol description |
| `description` | string | No | Brief description |
| `category` | enum | No | One of: `auth`, `agent`, `messaging`, `provisioning`, `other` |
| `references` | array | No | Links to specifications |

### Reference Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Reference name (e.g., "RFC 6749") |
| `url` | string | Yes | URL to the reference |

## Entities

Entities represent participants in the protocol: systems, services, actors, or trust domains.

```json
{
  "entities": [
    {
      "id": "client",
      "name": "Client Application",
      "type": "client",
      "description": "Application requesting access to protected resources"
    }
  ]
}
```

### Entity Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier (lowercase, alphanumeric, underscores) |
| `name` | string | Yes | Human-readable display name |
| `type` | enum | Yes | Entity type classification |
| `description` | string | No | Role description |

### Entity Types

| Type | Description | Typical Use |
|------|-------------|-------------|
| `client` | Application or service initiating requests | OAuth clients, API consumers |
| `authorization_server` | Issues tokens and handles authentication | OAuth/OIDC providers |
| `resource_server` | Hosts protected resources | APIs, data services |
| `user` | Human actor | End users, resource owners |
| `browser` | User agent | Web browsers |
| `agent` | AI/LLM agent | MCP agents, A2A agents |
| `tool_server` | Exposes tools via protocol | MCP tool servers |
| `tool` | Individual tool | MCP tools |
| `delegated_agent` | Agent receiving delegated tasks | A2A secondary agents |
| `identity_provider` | Authenticates users | SAML IdPs |
| `service_provider` | Relies on identity provider | SAML SPs |
| `server` | Generic server | General purpose |
| `other` | Custom entity type | Extension point |

## Phases

Phases provide optional logical grouping of flows for readability and diagram organization. Phases support hierarchical nesting via the `parent` field.

```json
{
  "phases": [
    {
      "id": "authorization",
      "name": "Authorization",
      "description": "User authentication and consent"
    },
    {
      "id": "mfa",
      "name": "Multi-Factor Authentication",
      "parent": "authorization",
      "description": "Optional MFA challenge"
    },
    {
      "id": "token_exchange",
      "name": "Token Exchange",
      "description": "Exchange code for tokens"
    }
  ]
}
```

### Phase Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `name` | string | Yes | Human-readable name |
| `description` | string | No | Phase description |
| `parent` | string | No | Parent phase ID for nesting |

### Phase Hierarchy

Phases can be nested to create hierarchical groupings:

- A phase with no `parent` is a root phase
- A phase with a `parent` must reference a valid phase ID
- Circular references are not allowed
- A phase cannot be its own parent

Phases are rendered as grouping constructs in diagrams:

- **PlantUML**: Colored `box` containers
- **Mermaid**: Colored `rect` blocks
- **D2**: Nested groups

## Flows

Flows are the core semantic unit: directed interactions between entities.

```json
{
  "flows": [
    {
      "from": "client",
      "to": "auth_server",
      "action": "token_request",
      "label": "POST /token",
      "mode": "request",
      "phase": "token_exchange",
      "description": "Exchange authorization code for access token",
      "condition": "code_valid",
      "note": "Code must be exchanged within 10 minutes",
      "annotations": [
        {"type": "security", "text": "Validate code verifier (PKCE)"}
      ],
      "alternatives": [
        {
          "condition": "code_invalid",
          "flows": [
            {"from": "auth_server", "to": "client", "action": "error", "mode": "response"}
          ]
        }
      ]
    }
  ]
}
```

### Flow Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `from` | string | Yes | Source entity ID |
| `to` | string | Yes | Target entity ID |
| `action` | string | Yes | Action identifier |
| `label` | string | No | Display label (defaults to action) |
| `mode` | enum | No | Interaction mode (defaults to `request`) |
| `phase` | string | No | Phase ID this flow belongs to |
| `description` | string | No | Detailed description |
| `sequence` | integer | No | Explicit ordering (default: array order) |
| `condition` | string | No | Conditional execution clause |
| `note` | string | No | Visible note displayed on diagram |
| `annotations` | array | No | Typed annotations for tooling |
| `alternatives` | array | No | Alternative flow paths |

### Flow Modes

| Mode | Description | Diagram Representation |
|------|-------------|------------------------|
| `request` | Synchronous request | Solid arrow `->` |
| `response` | Synchronous response | Dashed arrow `-->` |
| `redirect` | HTTP redirect | Solid arrow with redirect annotation |
| `callback` | Callback/webhook | Solid arrow with callback annotation |
| `interactive` | Human interaction | Solid arrow (user involved) |
| `event` | Asynchronous event | Dashed arrow |
| `tool_call` | Tool invocation (MCP) | Solid arrow with tool annotation |
| `tool_result` | Tool result (MCP) | Dashed arrow with result annotation |

### Conditional Flows

The `condition` field specifies when a flow is executed. Conditions are rendered as `opt` blocks in sequence diagrams.

```json
{
  "from": "client",
  "to": "server",
  "action": "refresh_token",
  "condition": "token_expired"
}
```

### Annotations

Annotations provide typed metadata for flows. They are rendered as notes in diagrams.

```json
{
  "annotations": [
    {
      "type": "security",
      "text": "Validate PKCE code_verifier",
      "details": "Compare SHA256(code_verifier) with stored code_challenge"
    }
  ]
}
```

#### Annotation Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | enum | Yes | Annotation category |
| `text` | string | Yes | Annotation message |
| `details` | string | No | Additional context |

#### Annotation Types

| Type | Description |
|------|-------------|
| `security` | Security considerations |
| `performance` | Performance implications |
| `deprecated` | Deprecated functionality |
| `info` | General information |
| `warning` | Warning conditions |
| `error` | Error conditions |

### Alternative Paths

The `alternatives` field defines branching paths from a flow. They are rendered as `alt/else` blocks in sequence diagrams.

```json
{
  "from": "client",
  "to": "server",
  "action": "authenticate",
  "alternatives": [
    {
      "condition": "invalid_credentials",
      "description": "Authentication failed",
      "flows": [
        {"from": "server", "to": "client", "action": "auth_error", "mode": "response"}
      ]
    },
    {
      "condition": "mfa_required",
      "description": "Multi-factor authentication needed",
      "flows": [
        {"from": "server", "to": "client", "action": "mfa_challenge", "mode": "response"},
        {"from": "client", "to": "server", "action": "mfa_response", "mode": "request"}
      ]
    }
  ]
}
```

#### Alternative Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `condition` | string | Yes | Condition that triggers this path |
| `flows` | array | Yes | Flows in this alternative path |
| `description` | string | No | Description of this alternative |

## Ordering

Flows are processed in array order by default. The optional `sequence` field provides explicit ordering when needed.

## Validation Rules

1. All entity IDs referenced in flows must be defined in `entities`
2. All phase IDs referenced in flows must be defined in `phases`
3. A protocol must have at least 2 entities
4. A protocol must have at least 1 flow
5. Entity and phase IDs must be unique within their respective arrays
6. IDs must match pattern: `^[a-z][a-z0-9_]*$` (entities/phases) or `^[a-z][a-z0-9_-]*$` (protocol)

## Schema

The canonical JSON Schema is available at:

- Repository: `schema/pidl.schema.json`
- URL: `https://github.com/grokify/pidl/schema/pidl.schema.json`

## Example

Minimal OAuth 2.0 token exchange:

```json
{
  "protocol": {
    "id": "oauth2-token-exchange",
    "name": "OAuth 2.0 Token Exchange"
  },
  "entities": [
    {"id": "client", "name": "Client", "type": "client"},
    {"id": "auth", "name": "Auth Server", "type": "authorization_server"}
  ],
  "flows": [
    {"from": "client", "to": "auth", "action": "token_request", "mode": "request"},
    {"from": "auth", "to": "client", "action": "token_response", "mode": "response"}
  ]
}
```

## File Extension

PIDL files use the `.pidl.json` extension by convention.

## Future Extensions

The following features are planned for future versions:

- State model and mutations
- Protocol composition and imports
- Loop constructs (`loop` blocks)
- Break/continue semantics
- External tool integration (PlantUML server, Kroki)

See TASKS.md for the complete roadmap.
