# Flows

Flows are the core semantic unit: directed interactions between entities.

## Basic Flow

```json
{
  "flows": [
    {
      "from": "client",
      "to": "server",
      "action": "request",
      "label": "GET /resource",
      "mode": "request"
    },
    {
      "from": "server",
      "to": "client",
      "action": "response",
      "label": "200 OK",
      "mode": "response"
    }
  ]
}
```

## Flow Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `from` | string | Yes | Source entity ID |
| `to` | string | Yes | Target entity ID |
| `action` | string | Yes | Action identifier |
| `label` | string | No | Display label (defaults to action) |
| `mode` | enum | No | Interaction mode (defaults to `request`) |
| `phase` | string | No | Phase ID this flow belongs to |
| `description` | string | No | Detailed description |
| `sequence` | integer | No | Deprecated. Display-only step number; never affects order |
| `condition` | string | No | Conditional execution clause |
| `note` | string | No | Visible note on diagram |
| `annotations` | array | No | Typed annotations |
| `alternatives` | array | No | Alternative flow paths |
| `verification` | object | No | Evidentiary provenance (see [Verification](#verification)) |

## Flow Modes

| Mode | Description | Arrow Style |
|------|-------------|-------------|
| `request` | Synchronous request | Solid `->` |
| `response` | Synchronous response | Dashed `-->` |
| `redirect` | HTTP redirect | Solid with annotation |
| `callback` | Callback/webhook | Solid with annotation |
| `interactive` | Human interaction | Solid |
| `event` | Asynchronous event | Dashed |
| `tool_call` | Tool invocation (MCP) | Solid with annotation |
| `tool_result` | Tool result (MCP) | Dashed with annotation |

## Ordering

Flows run in array order. To change the order, move the flow in the array.

The `sequence` field is deprecated. It is a display-only step number and does not reorder flows. See [Ordering](../SPECIFICATION.md#ordering) in the specification for the rules that apply to entities, phases and scenarios.

## Conditional Flows

Use `condition` to specify when a flow executes:

```json
{
  "from": "client",
  "to": "server",
  "action": "refresh_token",
  "condition": "token_expired"
}
```

Renders as `opt` block in sequence diagrams.

## Notes

Add visible notes to flows:

```json
{
  "from": "client",
  "to": "server",
  "action": "authenticate",
  "note": "Requires TLS 1.3"
}
```

## Verification

For incident attack-flows, a flow can record its **evidentiary provenance** — how
well-substantiated the step is and where it is sourced from. This lets a published
attack chain distinguish a step observed in a primary source from one that is
inferred.

```json
{
  "from": "agent-collective",
  "to": "hugging-face",
  "action": "exploit_worker",
  "verification": {
    "tier": "corroborated",
    "source": "METR",
    "citation": "https://metr.org/..."
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tier` | enum | Yes | `reported` \| `corroborated` \| `reproduced` \| `partially_reproduced` |
| `source` | string | No | Originating source (e.g. "OpenAI technical report", "METR") |
| `citation` | string | No | URL or reference locating the source |

The tier vocabulary matches the aisecurity-incidents `deepdive.json`. Sequence
renderers emit a compact provenance badge (e.g. `🔎 corroborated — METR`), toggled
by the renderer's `ShowVerification` option, and the security analyzer flags flows
below the `corroborated` tier.

See also:

- [Annotations](annotations.md) for typed metadata
- [Alternatives](alternatives.md) for branching paths
