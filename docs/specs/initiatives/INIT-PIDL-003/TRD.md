# TRD — PIDL for developer content

**Initiative:** `INIT-PIDL-003`
**Repository:** `github.com/grokify/pidl`

## Design

### 1. Types

```go
type Group struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
}
// Protocol gains: Groups []Group; Pronunciations map[string]map[string]string
// (term -> language -> spoken form), matching videoascode's pronunciations,
// e.g. {"agent_jkt": {"en-US": "agent J K T"}}. A map is right here: entries
// are looked up by term, never listed in order.
// ProtocolMeta gains: Language string (BCP 47, default "en-US")
// Scenario gains: Group, VariantOf, VariantLabel, Narration string; Briefing *Briefing
type Briefing struct {
    Why       string   `json:"why,omitempty"`       // why the scenario exists, as prose
    Points    []string `json:"points,omitempty"`    // key points, in display order
    Narration string   `json:"narration,omitempty"` // spoken over the briefing slides
}
// Flow gains: ID, Narration string
// FlowExample and Scenario gain: Samples []Sample
type Sample struct {
    Language string        `json:"language"`         // curl is generated, not stored
    Label    string        `json:"label,omitempty"`
    Role     string        `json:"role,omitempty"`   // entity ID whose developer writes this
    Code     string        `json:"code,omitempty"`
    Source   *SampleSource `json:"source,omitempty"`
}
type SampleSource struct {
    Repo   string `json:"repo"`             // e.g. github.com/aauth-dev/aauth-go
    Ref    string `json:"ref,omitempty"`    // tag or commit
    Path   string `json:"path"`
    Region string `json:"region"`
    SHA256 string `json:"sha256,omitempty"` // recorded by resolve
}
```

Language identifiers follow common highlighter names: `go`, `typescript`, `javascript`, `python`, `csharp`, `java`, `curl`.

### 2. Validation

- Group IDs unique; a scenario's `group` exists.
- `variant_of` names a scenario that is not itself a variant, in the same group; `variant_label` required with it.
- Flow IDs unique within a scenario; include steps may have IDs.
- Samples: exactly one of `code` or `source`; `role` names an entity.

### 3. Sample resolution (`samples.go`, `cmd/pidl samples`)

- `pidl samples resolve --root <dir>` maps `repo` to a local checkout under `<dir>` (default `$GOPATH/src`), extracts the region between `pidl:begin <name>` and `pidl:end <name>` (comment syntax per language), dedents it, and records `sha256`.
- `pidl samples check` fails when a region is missing or its hash changed without re-resolving.
- Generation (`walkthrough-json`, `viewer`) embeds resolved code; it never reaches the network.

### 4. cURL generation (`render/curl.go`)

From `ExampleRequest`: method, URL, each header as `-H`, and the body as `--data-raw` with the exact bytes the digest covers. A comment states that the signature headers are replayed from the example.

### 5. Example checks (`analyze/examples.go`)

Generic rules, scope document:

| Rule | Check |
|------|-------|
| EX001 | `Content-Digest` sha-256 equals the digest of the compact JSON body (and of the pretty-printed form, to allow either) |
| EX002 | A full compact JWT in a header or body decodes to a decoded token on the same example |
| EX003 | Every component in `Signature-Input` is present as a header or derived component |
| EX004 | `created` is within each presented token's `iat`..`exp` |
| EX005 | Decoded tokens have `exp > iat` |

Profiles (`analyze/profiles`): `aauth-draft-11` adds AA101 (`alg` must be fully specified; `EdDSA` rejected), AA102 (no `alg` parameter in `Signature-Input`), AA103 (`cnf.jwk` carries `alg`), AA104 (token lifetimes within the draft's limits), AA105 (resource token `agent_jkt` equals the thumbprint of the presenting key).

`protocol.profiles` selects profiles; `pidl analyze --profile` overrides.

### 6. Walkthrough model

`Walkthrough` gains `Groups`, `Pronunciations`; scenarios gain `Group`, `VariantOf`, `VariantLabel`, `Narration`, `Briefing`; steps gain `ID`, `Narration`, `Samples` (resolved), `Curl`, and `TokenEvents` (issued, forwarded, presented, checked) derived from `issues`, `verifies` and example tokens.

### 7. Viewer

- **Left navigation**: groups as sections; variants collapse under their base with a toggle in the scenario header.
- **Autoplay**: play, pause, replay; speed 0.5×–3×; a step's duration is `max(2.5 s, words / 2.7 s⁻¹)` from narration, or audio length when video timing is supplied.
- **Token lifecycle panel**: per token, the steps where it is issued, forwarded, presented and checked; the active step is highlighted.
- **Representation tabs**: HTTP, cURL and one tab per sample language present; the chosen language is remembered per browser.
- **Decoded/raw toggle** on tokens.

### 8. Explorer parity

```go
type GlossaryEntry struct {
    Term        string `json:"term"`                // e.g. "Signature-Key", "agent_jkt"
    Kind        string `json:"kind"`                // header | claim | term
    Category    string `json:"category,omitempty"`  // e.g. "protocol" to badge protocol-defined headers
    Description string `json:"description"`
    Reference   string `json:"reference,omitempty"` // spec section or URL
}
type Artifact struct {
    ID, Name  string
    MediaType string          // application/json | text/markdown
    Content   json.RawMessage // JSON value, or a JSON string for Markdown
    Digest    *ArtifactDigest // {alg: "sha-256", encoding: "base64url", of: "raw"|"canonical-json"}
}
type Comparison struct {
    ID, Name, Group string
    Scenarios []string        // column order
    Rows      []ComparisonRow // {question, answers: map[scenarioID]string}
}
type Page struct {
    ID, Name, Group string
    Markdown        string
}
// Scenario gains References []Reference; Protocol gains Glossary, Comparisons, Pages; scenarios and flows gain Artifacts.
```

- **Signature base** (`httpsig.go`): build the RFC 9421 base from `Signature-Input` components: derived components from the URL and method, header fields by name, and `@signature-params` from the input's parameters. Unknown components are reported as an example check finding.
- **Deferred timeline**: derived in the walkthrough from steps whose response is 202 with `Location`, pairing later polls of that URL until a non-202 response.
- **Value tracing**: the walkthrough indexes string values of at least 16 characters across headers, bodies, token claims and artifacts; the viewer highlights matches on selection.
- **Pages** render Markdown with a small, dependency-free renderer in the viewer (headings, lists, code, tables, links).

### 9. Guided tours

```go
type TourBeat struct {
    Focus     string `json:"focus"`
    Narration string `json:"narration,omitempty"`
    Hold      int    `json:"hold,omitempty"` // minimum milliseconds on screen
}
// Flow gains Tour []TourBeat
```

- **Regions** resolve against the walkthrough step: header names case-insensitively, JSON pointers per RFC 6901 into bodies, token claims by token definition ID and claim name, artifacts by ID.
- **Viewer** elements carry `data-pidl-region` attributes. `window.pidl.focus` selects the HTTP or Tokens tab as needed, expands collapsed values, scrolls the panel with `behavior: "instant"` in video mode, applies a highlight ring, and resolves after two animation frames.
- **`pidl tour draft`** emits beats for the step's notable regions (signature headers, challenge headers, body fields, claims the receiver checks) with narration drafted from glossary descriptions; values are not read aloud.

## Sequencing

Types and validation precede sample resolution and checks; walkthrough changes precede viewer work; the AAuth example gains groups, IDs and narration in `INIT-PIDL-004`, with a small fixture here for tests.

## Risks

- **Region markers in SDK repos** require SDK maintainers' cooperation. Mitigation: markers are comments, and inline code remains possible.
- **cURL cannot sign.** Mitigation: label replayed signatures clearly.
- **Profile rules encode a draft that changes.** Mitigation: profiles are versioned by draft (`aauth-draft-11`).

## Testing Strategy

Table tests for validation, region extraction (per comment syntax), cURL generation and each check rule; golden walkthrough JSON; viewer unit tests for navigation grouping, autoplay timing and lifecycle derivation.
