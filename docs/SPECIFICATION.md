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

A PIDL document is a JSON object. It describes either one flow (`flows`) or an ordered list of named scenarios (`scenarios`):

```json
{
  "protocol": { ... },
  "entities": [ ... ],
  "phases": [ ... ],
  "flows": [ ... ],
  "scenarios": [ ... ]
}
```

| Section | Required | Description |
|---------|----------|-------------|
| `protocol` | Yes | Metadata about the protocol |
| `entities` | Yes | Participants in the protocol, in display order |
| `phases` | No | Logical grouping of flows |
| `flows` | One of `flows` or `scenarios` | Interactions between entities, as a single flow |
| `groups` | No | Ordered navigation sections for scenarios (see [Groups](#groups)) |
| `scenarios` | One of `flows` or `scenarios` | Named, ordered sequences of flows (see [Scenarios](#scenarios)) |
| `pronunciations` | No | Spoken forms of terms per language (see [Narration](#narration)) |
| `metadata` | No | Tokens, networks, components and trust relations |

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
| `language` | string | No | BCP 47 tag of the narration source language (default `en-US`) |

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
| `id` | string | No | Step identifier, unique within its scenario and matching `^[a-z][a-z0-9_-]*$` (see [Step IDs](#step-ids)) |
| `narration` | object | No | Spoken for this step, by language (see [Narration](#narration)) |
| `representations` | object | No | Narration overrides for `http` or `curl` |
| `from` | string | Yes¹ | Source entity ID |
| `to` | string | Yes¹ | Target entity ID |
| `action` | string | Yes¹ | Action identifier |
| `include` | string | No | ID of a fragment scenario to insert here (scenarios only; see [Include Steps](#include-steps-and-fragments)) |
| `label` | string | No | Display label (defaults to action) |
| `mode` | enum | No | Interaction mode (defaults to `request`) |
| `phase` | string | No | Phase ID this flow belongs to |
| `description` | string | No | Detailed description |
| `sequence` | integer | No | Deprecated. Display-only step number; never affects order (see [Ordering](#ordering)) |
| `condition` | string | No | Conditional execution clause |
| `note` | string | No | Visible note displayed on diagram |
| `annotations` | array | No | Typed annotations for tooling |
| `alternatives` | array | No | Alternative flow paths |
| `verification` | object | No | Evidentiary provenance (tier/source/citation) |
| `example` | object | No | Illustrative request, response and decoded tokens (see [Flow Examples](#flow-examples)) |

¹ Not set on an include step, which has `include` instead.

### Flow Verification

Records the evidentiary provenance of a flow — used for incident attack-flows
where every step should trace to a source (or be marked reproduced/inferred).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tier` | enum | Yes | `reported` \| `corroborated` \| `reproduced` \| `partially_reproduced` |
| `source` | string | No | Originating source (e.g. "OpenAI technical report", "METR") |
| `citation` | string | No | URL or reference locating the source |

```json
{
  "from": "agent", "to": "target", "action": "exfiltrate",
  "verification": {"tier": "corroborated", "source": "METR", "citation": "https://metr.org/..."}
}
```

The tier vocabulary matches the aisecurity-incidents `deepdive.json`.

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

## Scenarios

A protocol often has several journeys through the same roles: AAuth has five access modes, OAuth has several grants. Instead of one file per journey, a document can hold an ordered list of named scenarios that share the same `entities`, `phases` and `metadata`.

```json
{
  "protocol": {"id": "demo", "name": "Demo"},
  "entities": [
    {"id": "person", "name": "Person", "type": "user"},
    {"id": "agent", "name": "Agent", "type": "agent"},
    {"id": "ps", "name": "Person Server", "type": "authorization_server"}
  ],
  "scenarios": [
    {
      "id": "three-party",
      "name": "PS authorization",
      "flows": [
        {"from": "agent", "to": "ps", "action": "auth_token_request"},
        {"include": "consent"},
        {"from": "ps", "to": "agent", "action": "auth_token", "mode": "response"}
      ]
    },
    {
      "id": "consent",
      "name": "User consent",
      "fragment": true,
      "flows": [
        {"from": "ps", "to": "person", "action": "consent_prompt"},
        {"from": "person", "to": "ps", "action": "approve", "mode": "response"}
      ]
    }
  ]
}
```

A document has either `flows` or `scenarios`. A document with `flows` behaves as a single unnamed scenario, so existing files keep working.

### Scenario Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier, matching `^[a-z][a-z0-9_-]*$` |
| `name` | string | Yes | Display name |
| `description` | string | No | Description of the scenario |
| `participants` | array | No | Entity IDs to show, in `entities` order. Defaults to the entities the scenario uses, including those of included fragments |
| `fragment` | boolean | No | Marks a reusable sub-flow that other scenarios include |
| `group` | string | No | ID of the [group](#groups) the scenario is listed under. Fragments have no group |
| `variant_of` | string | No | ID of the main scenario this one is a [variant](#variants) of |
| `variant_label` | string | With `variant_of` | Label for the variant in a toggle, such as "With user approval" |
| `narration` | object | No | Spoken when the scenario opens, by language (see [Narration](#narration)) |
| `briefing` | object | No | Why the scenario exists, its key points, and narration for them |
| `samples` | array | No | Code for the scenario as a whole (see [Samples](#samples)) |
| `flows` | array | Yes | The scenario's steps, in order |

### Groups

A document MAY declare `groups`, an ordered array of navigation sections, each with an `id`, a `name` and an optional `description`. A scenario names its section with `group`:

```json
{
  "groups": [
    {"id": "identity", "name": "Agent identity"},
    {"id": "access", "name": "Access modes"}
  ],
  "scenarios": [
    {"id": "agent-identity", "name": "Agent identity", "group": "identity", "flows": [ ... ]},
    {"id": "ps-authorization", "name": "PS authorization", "group": "access", "flows": [ ... ]}
  ]
}
```

Groups set output order: scenarios in group order, then main scenarios without a group, then fragments (see [Ordering](#ordering)).

### Variants

A variant is another form of the same journey, such as an autonomous flow and the same flow with user approval. It names its base with `variant_of` and a short `variant_label`. A variant shares its base's navigation entry, follows it directly in output order, and sits in the same group. A base cannot itself be a variant or a fragment.

### Step IDs

A flow MAY have an `id`, unique within its scenario. IDs give samples, narration and links a stable target that survives steps being added or moved, unlike step numbers. Inside an expanded scenario, a step from an included fragment is addressed as `<include>.<id>`, where `<include>` is the include step's `id` or, without one, the fragment's ID. A scenario that includes the same fragment twice gives those include steps distinct IDs. Viewers accept `#<scenario>/<step-id>` as well as `#<scenario>/<n>`.

### Include Steps and Fragments

A step `{"include": "<scenario-id>"}` inserts a fragment's flows at that point. The target must be a scenario with `"fragment": true`, and includes must not form a cycle. An include step may also set `phase`, `condition`, `label`, `note`, `description` and `annotations`; inlined flows inherit the include's `phase` and `condition` when they have none of their own.

Diagram formats draw an include step as a labeled box spanning the fragment's participants (PlantUML `ref over`, a shaded Mermaid `rect`, a D2 group, an SVG box), or inline the fragment's flows when expanded.

Fragments are listed after main scenarios wherever scenarios are numbered or listed, so a deck or file listing shows the journeys first and the shared pieces after them.

### Working with One Scenario

Most outputs describe one flow. Tools select a scenario and project it into a single-flow protocol: the scenario's flows, its participants (in `entities` order), the phases it uses, and the shared metadata. In the Go library this is `Protocol.Scenario(id, ProjectOptions{Expand: ...})`; on the command line it is `--scenario <id>`.

### Token Verification

Token definitions and flow security can say who checks what, so reviews and generated tables show where each guarantee is enforced:

| Field | On | Description |
|-------|----|-------------|
| `verified_by` | token definition | Entity IDs that check the token |
| `claims` | token definition | Claims checkers rely on, each with `name`, `checked_by` and `description` |
| `signed_by` | flow `security` | Entity that signs the request, when it is not the sender |
| `issues` | flow `security` | Token IDs the flow delivers |
| `verifies` | flow `security` | Token IDs the receiver checks |

`issues` and `verifies` let tools follow a token through a scenario: who it was delivered to, who presents it, and who signs the request that carries it.

## Narration

Narration is text written to be spoken, for autoplay, audio and video. On-screen prose such as `description` often contains identifiers (`agent_jkt`, `aa-auth+jwt`) that read badly aloud, so narration is a separate script.

Every narration value is an object keyed by BCP 47 language tag:

```json
{
  "protocol": {"id": "aauth", "name": "AAuth", "language": "en-US"},
  "pronunciations": {"agent_jkt": {"en-US": "agent J K T"}},
  "scenarios": [{
    "id": "agent-identity",
    "name": "Agent identity",
    "narration": {"en-US": "The agent proves who it is.", "fr-FR": "L'agent prouve son identité."},
    "briefing": {
      "why": "Resources need to know which agent is calling before deciding anything else.",
      "points": ["Every request is signed", "No server issues a token"],
      "narration": {"en-US": "Why agent identity comes first."}
    },
    "flows": [{
      "id": "sign",
      "from": "agent", "to": "resource", "action": "signed_request",
      "narration": {"en-US": "The agent signs the request with its key."},
      "representations": {
        "curl": {"narration": {"en-US": "This curl command replays the signed request."}}
      }
    }]
  }]
}
```

- **Source language.** `protocol.language` (default `en-US`) is the language a document is written in. Every narration object includes it, and it is the fallback for any language a narration lacks.
- **Representations.** A step's `representations` override its narration when the reader views the step as `http`, `curl`, or one of the step's sample languages. The fallback order is: the representation's narration in the chosen language, then in the source language, then the step's own narration in the chosen language, then in the source language, then `description`.
- **Pronunciations** map a term to its spoken form per language, in the shape videoascode reads, so text-to-speech tools can use them unchanged.
- **Include steps** carry no narration; their fragment's steps do.

## Flow Examples

A flow can carry an illustrative message for walkthrough viewers. Examples explain a protocol; they are not message schemas.

```json
{
  "from": "agent", "to": "ps", "action": "auth_token_request",
  "example": {
    "source": "draft-11 §7.2.1",
    "kind": "draft",
    "request": {
      "method": "POST",
      "url": "https://ps.example/token",
      "headers": [
        {"name": "Content-Type", "value": "application/json"},
        {"name": "Signature-Key", "value": "sig=jwt;jwt=\"eyJ...\""}
      ],
      "body": {"resource_token": "eyJ...", "presented_token": "eyJ..."}
    },
    "tokens": [
      {"token": "resource_token", "label": "request body",
       "header": {"alg": "Ed25519", "typ": "aa-resource+jwt"},
       "payload": {"iss": "https://resource.example", "agent_jkt": "kPrK..."}}
    ],
    "note": "The PS checks agent_jkt against the key that signed this request."
  }
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `source` | Yes | Citation, such as a specification section |
| `kind` | Yes | `draft`: taken from the source, values aligned to the scenario; `derived`: written from the source's normative text |
| `request` | One of `request`, `response`, `tokens` | `method`, `url`, ordered `headers` (`name`, `value`) and a JSON `body` |
| `response` | One of `request`, `response`, `tokens` | `status`, ordered `headers` and a JSON `body` |
| `tokens` | One of `request`, `response`, `tokens` | Decoded tokens: `token` (a token definition ID), `label`, `header` and `payload` |
| `note` | No | What the reader should notice |
| `samples` | No | Code that sends or handles the message (see [Samples](#samples)) |

Headers are a list so they keep their order. Include steps cannot carry examples.

**Wire form of a body.** A body is written as JSON for readability. On the wire it is the compact JSON encoding with keys in the order written; a body written as a JSON string is sent as the string's contents, for media types that are not JSON. A `Content-Digest` in an example covers this wire form.

**cURL.** Tools generate an equivalent `curl` command for every example request: the method, the URL, each header in order and the wire body. curl cannot create HTTP message signatures, so a signed request's command replays the example's signature headers and says so in a leading comment.

## Samples

A sample is code that sends or handles a step's message, such as an SDK call, shown beside the HTTP and cURL forms. Flow examples and scenarios carry `samples`:

```json
"samples": [
  {"language": "go", "label": "Agent", "role": "agent",
   "source": {"repo": "github.com/aauth-dev/aauth-go", "ref": "v0.3.0",
              "path": "examples/walkthrough/federated_test.go", "region": "federated.ps-to-as"}},
  {"language": "python", "code": "client.request_auth_token(resource_token)"}
]
```

| Field | Required | Description |
|-------|----------|-------------|
| `language` | Yes | Highlighter name, such as `go`, `typescript`, `python` or `csharp`. `http` and `curl` are generated, not samples |
| `label` | No | Distinguishes samples in one language |
| `role` | No | Entity ID whose developer writes this code |
| `code` | One of `code`, `source` | Inline code |
| `source` | One of `code`, `source` | `repo`, optional `ref` (tag or commit), `path` within the repository, and `region` |

**Regions.** A source names the code between `pidl:begin <region>` and `pidl:end <region>` markers, written in comments in any syntax. Marker lines of other regions inside it are dropped, common indentation is removed, and blank lines at either end are trimmed. Sources should point at tested files, so samples compile and run with each release of the code they show.

**Lock file.** `pidl samples resolve <file>` reads each source from checkouts under a root directory (default `$GOPATH/src`, at `<root>/<repo>`; with `ref`, the file at that ref) and writes the code and its SHA-256 to `<file without .json>.samples.json`. Generation reads the lock, so outputs never need the source repositories or the network. `pidl samples check <file>` fails when a source is not in the lock, its region is missing, or its code changed since it was resolved; CI runs it to keep samples current.

## Ordering

Order in PIDL comes from one source: the position of an item in its JSON array. Every output generated from a document (diagrams, tables, outlines, file listings) uses the same order, so presentations, guides and diagrams built from one file agree.

1. **Roles** follow the order of `entities`. Diagram columns, role tables and any other list of entities use this order.
2. **Scenarios** follow the order of `scenarios`, arranged by `groups` when present: grouped scenarios in group order, then main scenarios without a group, with each variant directly after its base, then fragments. Within each section, the order of `scenarios` applies.
3. **Phases** follow the order of `phases`.
4. **Messages** follow the order of `flows` (top-level or within a scenario). An include step expands in place.
5. **Derived lists keep the order of the array they come from.** A subset of entities (such as a scenario's participants or a token's checkers) is listed in `entities` order, never in the order written.
6. **Collections keyed by name are arrays of objects with an `id`.** The one exception is `metadata.networks`, a map kept for compatibility, which tools never read in map iteration order:
   - Network boundaries are ordered by `metadata.network_layout.order` first. Boundaries it does not list follow by style (external, dmz, trusted, cloud, default) and then by ID.
   - An entity listed in more than one network belongs to the network whose ID sorts last.
   - Entities inside each boundary follow `entities` order, like every other list of entities. To list them alphabetically, sort the `entities` array.

Numeric ordering fields are not used. The `sequence` field on a flow is deprecated: it is a display-only step number, and tools must not sort by it.

## Validation Rules

1. All entity IDs referenced in flows must be defined in `entities`
2. All phase IDs referenced in flows must be defined in `phases`
3. A protocol must have at least 2 entities
4. A protocol must have at least 1 flow, or at least 1 scenario that is not a fragment; a document with scenarios has no top-level flows
5. Entity, phase and scenario IDs must be unique within their respective arrays
6. IDs must match pattern: `^[a-z][a-z0-9_]*$` (entities/phases) or `^[a-z][a-z0-9_-]*$` (protocol/scenarios)
7. A scenario must have a name and at least 1 flow
8. A flow is either a message (`from`, `to`, `action`) or an include step, and include steps appear only in scenarios
9. An include must name a fragment scenario, and includes must not form a cycle
10. A scenario's `participants` must name known entities, once each, in `entities` order, and must cover every entity the scenario uses
11. Token and entity references in `verified_by`, `claims`, `signed_by`, `issues` and `verifies` must resolve
12. Group IDs are unique and match `^[a-z][a-z0-9_-]*$`; a scenario's `group` names a declared group, and fragments have no group
13. `variant_of` names a main scenario that is not itself a variant, in the same group, and `variant_label` is set exactly when `variant_of` is
14. Flow IDs match `^[a-z][a-z0-9_-]*$` and are unique within their flow list, and step IDs stay unique once fragments are expanded
15. Narration keys are BCP 47 language tags with non-empty text, and every narration object includes the source language; representation keys are `http`, `curl` or a sample language on the step; include steps carry no narration
16. A sample has a language other than `http` and `curl`, exactly one of `code` and `source`, a known `role`, and a source path inside its repository

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
