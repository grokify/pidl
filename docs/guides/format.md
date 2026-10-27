# PIDL Format

A PIDL document is a JSON file with four sections:

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
| `entities` | Yes | Participants in the protocol (min 2) |
| `phases` | No | Logical grouping of flows |
| `flows` | Yes | Interactions between entities (min 1) |

## Protocol Metadata

```json
{
  "protocol": {
    "id": "oauth2-authorization-code",
    "name": "OAuth 2.0 Authorization Code Flow",
    "version": "1.0",
    "description": "OAuth 2.0 Authorization Code Grant",
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

## File Extension

PIDL files use the `.pidl.json` extension by convention.

## Validation

Use the CLI to validate files:

```bash
pidl validate protocol.json
```

Or programmatically:

```go
p, _ := pidl.ParseFile("protocol.json")
if errs := p.Validate(); errs.HasErrors() {
    fmt.Println(errs)
}
```
