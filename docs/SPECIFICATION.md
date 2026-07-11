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

---

## Process Profile

PIDL supports a **process profile** for describing data processing workflows in addition to the standard protocol profile. Process specs model step-by-step data transformations with explicit input/output ports, processing characteristics, and failure handling.

### Protocol Kind

The `kind` field in the protocol metadata distinguishes between protocol and process specifications:

```json
{
  "protocol": {
    "id": "etl-pipeline",
    "name": "ETL Pipeline",
    "kind": "process"
  }
}
```

| Kind | Description |
|------|-------------|
| `protocol` | Standard protocol (default) - models interactions between systems |
| `process` | Process specification - models data processing workflows |

### Step Types

In process specifications, entities represent processing steps. The `step_type` field classifies each step:

```json
{
  "entities": [
    {
      "id": "extract",
      "name": "Data Extraction",
      "type": "server",
      "step_type": "deterministic"
    },
    {
      "id": "transform",
      "name": "LLM Transform",
      "type": "server",
      "step_type": "llm"
    }
  ]
}
```

| Step Type | Icon | Description | Characteristics |
|-----------|------|-------------|-----------------|
| `deterministic` | ⚙️ | Predictable processing | Same input → same output |
| `llm` | 🧠 | AI/ML processing | Non-deterministic, may require validation |
| `human` | 👤 | Human involvement | Manual review, approval, or input |
| `external` | ☁️ | External services | API calls, third-party services |
| `tool` | 🔧 | Tool invocations | Function calls, utilities |

### Data Ports

Entities can define input and output data ports for explicit data flow modeling:

```json
{
  "entities": [
    {
      "id": "transform",
      "name": "Transform Step",
      "step_type": "llm",
      "inputs": [
        {
          "name": "raw_data",
          "kind": "object",
          "required": true,
          "description": "Raw input data to transform"
        }
      ],
      "outputs": [
        {
          "name": "processed_data",
          "kind": "object",
          "description": "Transformed output data"
        },
        {
          "name": "audit_log",
          "kind": "file",
          "sensitive": true,
          "description": "Processing audit trail"
        }
      ]
    }
  ]
}
```

#### Data Port Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Port identifier |
| `kind` | enum | Yes | Data type classification |
| `description` | string | No | Port description |
| `required` | boolean | No | Whether input is required (default: false) |
| `sensitive` | boolean | No | Contains sensitive data (default: false) |
| `schema` | string | No | JSON Schema reference for validation |

#### Data Port Kinds

| Kind | Icon | Description |
|------|------|-------------|
| `file` | 📄 | File-based data |
| `object` | 📦 | In-memory object/struct |
| `api` | 🌐 | API request/response |
| `database` | 🗄️ | Database record/query |
| `queue` | 📬 | Message queue item |
| `stream` | 🌊 | Streaming data |

### Processing Configuration

The `processing` field configures step execution characteristics:

```json
{
  "entities": [
    {
      "id": "llm_step",
      "name": "LLM Processing",
      "step_type": "llm",
      "processing": {
        "engine": "claude-3-opus",
        "deterministic": false,
        "timeout": "PT30S",
        "cache_key": "input_hash",
        "max_retries": 3
      }
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `engine` | string | Processing engine identifier |
| `deterministic` | boolean | Whether output is deterministic |
| `timeout` | string | ISO 8601 duration (e.g., "PT30S") |
| `cache_key` | string | Cache key expression |
| `max_retries` | integer | Maximum retry attempts |

### Failure Modes

The `failure_modes` field documents potential failure scenarios:

```json
{
  "entities": [
    {
      "id": "external_api",
      "name": "External API Call",
      "step_type": "external",
      "failure_modes": [
        {
          "id": "timeout",
          "name": "Request Timeout",
          "severity": "medium",
          "recovery": "Retry with exponential backoff"
        },
        {
          "id": "rate_limit",
          "name": "Rate Limited",
          "severity": "low",
          "recovery": "Wait and retry"
        }
      ]
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Failure mode identifier |
| `name` | string | Yes | Human-readable name |
| `description` | string | No | Detailed description |
| `severity` | enum | No | `critical`, `high`, `medium`, `low` |
| `recovery` | string | No | Recovery strategy description |

### Retry Strategy

The `retry_strategy` field configures automatic retry behavior:

```json
{
  "entities": [
    {
      "id": "api_call",
      "name": "API Call",
      "step_type": "external",
      "retry_strategy": {
        "max_attempts": 3,
        "initial_delay": "PT1S",
        "max_delay": "PT30S",
        "backoff_multiplier": 2.0,
        "retryable_errors": ["timeout", "rate_limit"]
      }
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `max_attempts` | integer | Maximum retry attempts |
| `initial_delay` | string | Initial delay (ISO 8601 duration) |
| `max_delay` | string | Maximum delay between retries |
| `backoff_multiplier` | number | Exponential backoff multiplier |
| `retryable_errors` | array | Error types that trigger retry |

### Process Spec Example

Complete ETL pipeline example:

```json
{
  "protocol": {
    "id": "etl-pipeline",
    "name": "ETL Pipeline",
    "kind": "process",
    "description": "Extract, transform, and load data pipeline"
  },
  "entities": [
    {
      "id": "extract",
      "name": "Extract",
      "type": "server",
      "step_type": "deterministic",
      "outputs": [
        {"name": "raw_data", "kind": "object"}
      ]
    },
    {
      "id": "transform",
      "name": "Transform",
      "type": "server",
      "step_type": "llm",
      "inputs": [
        {"name": "raw_data", "kind": "object", "required": true}
      ],
      "outputs": [
        {"name": "processed_data", "kind": "object"}
      ],
      "processing": {
        "engine": "claude-3-opus",
        "deterministic": false,
        "timeout": "PT60S"
      }
    },
    {
      "id": "validate",
      "name": "Human Review",
      "type": "server",
      "step_type": "human",
      "inputs": [
        {"name": "processed_data", "kind": "object", "required": true}
      ],
      "outputs": [
        {"name": "approved_data", "kind": "object"}
      ]
    },
    {
      "id": "load",
      "name": "Load",
      "type": "server",
      "step_type": "external",
      "inputs": [
        {"name": "approved_data", "kind": "object", "required": true}
      ],
      "failure_modes": [
        {"id": "connection_error", "name": "Connection Error", "severity": "high"}
      ],
      "retry_strategy": {
        "max_attempts": 3,
        "initial_delay": "PT5S",
        "backoff_multiplier": 2.0
      }
    }
  ],
  "flows": [
    {"from": "extract", "to": "transform", "action": "send_data"},
    {"from": "transform", "to": "validate", "action": "review"},
    {"from": "validate", "to": "load", "action": "approve"}
  ]
}
```

### Process Spec Security Analysis

PIDL includes security rules specific to process specifications:

| Rule | Severity | Description |
|------|----------|-------------|
| SEC011 | Medium | LLM step without downstream validation |
| SEC012 | High | Sensitive data flows to LLM step |
| SEC013 | Medium | Non-deterministic step in critical path |
| SEC014 | Low | External step without failure modes |
| SEC015 | Medium | Human step without timeout |

### Process Spec Rendering

Process specifications render with step-type-specific styling:

| Output Format | Step Type Rendering |
|---------------|---------------------|
| PlantUML | Stereotypes (`<<llm>>`) with colored participants |
| Mermaid | Emoji badges in participant names |
| D2 | Fill/stroke colors per step type |
| SVG | Inline styles and emoji badges |
| Infographic | Custom shapes per step type |

#### Infographic Output

Process specs can be rendered as compact infographics for social media and datasheets:

```bash
# LinkedIn-optimized infographic
pidl generate -f infographic --size=linkedin-square --title="ETL Pipeline" etl.json

# Datasheet tile
pidl generate -f infographic --size=datasheet-tile --theme=minimal etl.json
```

| Infographic Size | Dimensions | Use Case |
|------------------|------------|----------|
| `linkedin-square` | 1200×1200 | LinkedIn feed posts |
| `linkedin-portrait` | 1080×1350 | LinkedIn portrait posts |
| `linkedin-landscape` | 1200×627 | LinkedIn link previews |
| `datasheet-tile` | 400×400 | Datasheet grid layouts |
| `datasheet-wide` | 600×300 | Wide datasheet tiles |

| Theme | Description |
|-------|-------------|
| `bold` | High contrast, saturated (default) |
| `minimal` | Clean, subtle |
| `dark` | Dark background |
| `tech` | Tech/engineering feel |

---

## Future Extensions

The following features are planned for future versions:

- Loop constructs (`loop` blocks)
- Break/continue semantics
- External tool integration (PlantUML server, Kroki)
- Workflow engine exports (Temporal, Airflow, Step Functions)
- Data lineage tracking

See [ROADMAP.md](specs/ROADMAP.md) for the complete roadmap.
