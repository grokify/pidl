# Annotations

Annotations provide typed metadata for flows. They are rendered as notes in diagrams.

## Definition

```json
{
  "from": "client",
  "to": "server",
  "action": "authenticate",
  "annotations": [
    {
      "type": "security",
      "text": "Validate PKCE code_verifier",
      "details": "Compare SHA256(code_verifier) with stored code_challenge"
    },
    {
      "type": "performance",
      "text": "Cache token for TTL"
    }
  ]
}
```

## Annotation Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | enum | Yes | Annotation category |
| `text` | string | Yes | Annotation message |
| `details` | string | No | Additional context |

## Annotation Types

| Type | Description | Use Cases |
|------|-------------|-----------|
| `security` | Security considerations | Auth checks, encryption, validation |
| `performance` | Performance implications | Caching, timeouts, rate limits |
| `deprecated` | Deprecated functionality | Legacy endpoints, migration notes |
| `info` | General information | Documentation, explanations |
| `warning` | Warning conditions | Edge cases, prerequisites |
| `error` | Error conditions | Failure modes, error handling |

## Diagram Rendering

Annotations are rendered as notes in sequence diagrams:

### Mermaid

```
note right of server: [SECURITY] Validate PKCE code_verifier
```

### PlantUML

```
note right: <&warning> SECURITY: Validate PKCE code_verifier
```

### D2

Rendered as separate note messages or tooltips.

## Multiple Annotations

A flow can have multiple annotations:

```json
{
  "annotations": [
    {"type": "security", "text": "Requires TLS 1.3"},
    {"type": "security", "text": "Validate client certificate"},
    {"type": "performance", "text": "Response cached for 5 minutes"}
  ]
}
```

## vs Notes

| Feature | `note` | `annotations` |
|---------|--------|---------------|
| Structure | Plain string | Typed objects |
| Multiple | No | Yes |
| Metadata | No | Type + details |
| Tooling | Display only | Processable |

Use `note` for simple display text. Use `annotations` for structured metadata that tools can process.

## Helper Methods (Go)

```go
// Check if flow has annotations
if flow.HasAnnotations() {
    for _, ann := range flow.Annotations {
        fmt.Printf("[%s] %s\n", ann.Type, ann.Text)
    }
}

// Validate annotation type
if pidl.IsValidAnnotationType("security") {
    // valid
}
```
