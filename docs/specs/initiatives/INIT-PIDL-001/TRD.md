# TRD — Ordered roles and multiple scenarios per protocol

**Initiative:** `INIT-PIDL-001`
**Repository:** `github.com/grokify/pidl`

## Current State

- `Protocol` (`types.go`) has `Entities []Entity`, `Phases []Phase` and
  `Flows []Flow`. A file describes exactly one flow.
- Sequence renderers (`render/mermaid.go`, `render/plantuml.go`,
  `render/d2.go`) declare participants by iterating `p.Entities`, so
  column order already follows the array. Each renderer declares every
  entity, including ones the flow never uses.
- `Flow.Sequence` is documented as "explicit ordering". The only place
  it is read is a non-negative check (`validate_flow.go`). Nothing sorts
  by it.
- `ProtocolMetadata.Networks` is a `map[string]*NetworkConfig`. Its
  order is undefined unless `network_layout.order` is set.
- `ProtocolExtends.ExcludeFlows` and `ParallelBranch.FlowRefs` are
  integer indices into `Protocol.Flows`.
- `Executor` and `Debugger` index into `Protocol.Flows` directly.
- `render/markdown_matrix.go` builds an entities × protocol-roles table
  and sorts the protocol columns alphabetically.
- `ValidationError` has no severity, so every reported problem is an
  error.

## Design

### 1. Types

```go
// Scenario is a named, ordered sequence of flows within a protocol.
type Scenario struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    Description  string   `json:"description,omitempty"`
    Participants []string `json:"participants,omitempty"` // subset of entity IDs, in entity order
    Fragment     bool     `json:"fragment,omitempty"`     // reusable via include; skipped by --all-scenarios
    Flows        []Flow   `json:"flows"`
}

type Protocol struct {
    // ...existing fields...
    Flows     []Flow     `json:"flows,omitempty"`     // was required; now optional when Scenarios is set
    Scenarios []Scenario `json:"scenarios,omitempty"`
}

type Flow struct {
    // ...existing fields...
    // Include references a scenario whose flows are inserted at this
    // point. When set, From, To and Action must be empty.
    Include string `json:"include,omitempty"`
    // Sequence is a display-only step number. It never affects order.
    // Deprecated: order is the position in the flows array.
    Sequence int `json:"sequence,omitempty"`
}
```

Verification semantics:

```go
type TokenDefinition struct {
    // ...existing fields...
    VerifiedBy []string     `json:"verified_by,omitempty"` // entity IDs, in entity order
    Claims     []TokenClaim `json:"claims,omitempty"`      // ordered
}

type TokenClaim struct {
    Name        string   `json:"name"`
    CheckedBy   []string `json:"checked_by,omitempty"`
    Description string   `json:"description,omitempty"`
}

type FlowSecurity struct {
    // ...existing fields...
    SignedBy string   `json:"signed_by,omitempty"` // entity ID; defaults to Flow.From
    Issues   []string `json:"issues,omitempty"`    // token IDs carried in the response
    Verifies []string `json:"verifies,omitempty"`  // token IDs the receiver checks
}

type ProtocolExtends struct {
    // ...existing fields...
    ExcludeScenarios []string `json:"exclude_scenarios,omitempty"`
}
```

`Issues` and `Verifies` give the analyzer the data it needs for the
"used before issued" check without a full simulation.

### 2. Ordering rules

These rules are normative in `docs/SPECIFICATION.md`:

1. Role order is the order of `entities`.
2. Scenario order is the order of `scenarios`.
3. Phase order is the order of `phases`.
4. Message order within a scenario is the order of its `flows`. An
   include expands in place.
5. Every list derived from these arrays keeps their order. Examples
   are participants, token checkers, matrix rows and columns, outline
   entries and generated file names.
6. Collections keyed by name must be arrays of objects with an `id`.
   `metadata.networks` stays a map for compatibility, but renderers
   order it by `network_layout.order` when set and otherwise by
   first appearance of its entities in `entities`. They never order it
   by map iteration.

Helpers in a new file `order.go`:

```go
func (p *Protocol) EntityIndex(id string) int          // -1 if absent
func (p *Protocol) OrderEntities(ids []string) []string // stable sort by EntityIndex
func (p *Protocol) ScenarioIDs(includeFragments bool) []string
```

### 3. Validation (`validate_scenario.go`)

Added to `Protocol.Validate()`:

- `flows` must be non-empty when `scenarios` is empty, and empty when
  `scenarios` is non-empty.
- Scenario `id`s are unique and match the existing ID pattern.
- Every `participants` entry is an entity ID. The list has no
  duplicates and is in entity order. Out-of-order lists fail with a
  message that gives the expected order.
- Each flow is either an include (only `include`, plus optional
  `phase`, `note`, `condition`) or a message (`from`, `to`, `action`).
  It cannot be both.
- `include` targets exist and have `fragment: true`.
- Includes form no cycle. The check is a depth-first search with the
  path in the error.
- When `participants` is set, every flow's `from` and `to` and every
  included fragment's participants are in it.
- `FlowSecurity.SignedBy`, `TokenDefinition.VerifiedBy` and
  `TokenClaim.CheckedBy` reference entity IDs.
  `Issues`/`Verifies` reference token IDs.
- `ParallelBranch.FlowRefs` inside a scenario are indices into that
  scenario's flows.

The existing per-flow checks in `validate_flow.go` run on each
scenario's flows. Errors carry the path, for example
`scenarios[2].flows[4].to`.

### 4. Projection (`scenario.go`)

```go
type ProjectOptions struct {
    Expand bool // inline included fragments; default renders them as include steps
}

// Scenario returns a single-flow Protocol for scenario id. Entities are
// filtered to the scenario's participants in entity order; phases are
// filtered to those referenced, in phase order; metadata is copied.
func (p *Protocol) Scenario(id string, opts ProjectOptions) (*Protocol, error)

// Legacy files: Scenario("", opts) returns p unchanged when Scenarios is empty.
```

- With `Expand`, include steps are replaced recursively by the fragment's
  flows. Index-based references are re-based onto the expanded list.
- Without `Expand`, include steps stay in `Flows`. Renderers that do not
  understand includes would break, so each sequence renderer gains an
  include case (section 5), and non-sequence renderers (DOT, network,
  infographic) always receive an expanded projection.
- The projected protocol carries the scenario's `id` and `name` in
  `ProtocolMeta` (`Name: "<protocol> — <scenario>"`) so titles render
  without renderer changes.

Executor and debugger run on a projection with `Expand: true`. Their
index-based code is unchanged.

### 5. Rendering includes (collapsed)

The box spans the fragment's participants, from the first to the last
in entity order.

| Format | Output |
|--------|--------|
| PlantUML | `ref over a, b : <fragment name>` |
| Mermaid | `rect rgb(...)` block containing `Note over a,b: <fragment name>` |
| D2 sequence | a group labeled with the fragment name, spanning the participants |
| SVG | a rounded rectangle across the participant lifelines, labeled |

Golden files in `render/testdata` cover each format.

### 6. CLI (`cmd/pidl`)

- `generate`, `simulate`, `debug` and `analyze` take `--scenario <id>`.
  On a file with scenarios, omitting it is an error that lists the
  available IDs in order, except for `analyze`, which runs on all.
- `generate --all-scenarios -o <dir>` writes
  `NN-<scenario-id>.<ext>`. `NN` is the 1-based position among
  non-fragment scenarios, zero-padded to at least two digits
  (wider when there are 100 or more scenarios).
  `--include-fragments` adds fragments after the main scenarios, with
  continued numbering.
- `--expand` on `generate` inlines fragments.
- New formats on `generate`:
  - `scenario-matrix` (Markdown) — rows are entities in order, columns
    are scenarios in order, cells mark participation.
  - `token-table` (Markdown) — rows are tokens in array order, with
    issuer, audience, binding, verified by and claims.
  - `outline` (Markdown) and `outline-json` — protocol, then roles in
    order with descriptions, then scenarios in order with each step's
    label (includes shown by fragment name). This is the input for deck
    and guide tooling.
- `pidl scenarios <file>` lists scenarios in order, with the participant
  count, flow count and fragment flag. `-f json` gives JSON output.

### 7. Analyzer (`analyze`)

New rules, each with an ID, a severity and an explanation:

| Rule | Reports |
|------|---------|
| `SEC017` Unused role | an entity that no flow uses |
| `SEC018` Token checked before issued | a flow `verifies` a token that no earlier step of the expanded scenario `issues`; tokens no flow issues are treated as external and skipped |
| `SEC019` Signer differs from sender or token holder | `signed_by` differs from `from`, or a request presents a token that an earlier step delivered to someone other than the signer (the AAuth four-party `agent_jkt` case); skipped when the flow has a security annotation |
| `SEC020` Token verifier mismatch | a flow's receiver checks a token whose `verified_by` does not list it |
| `SEC021` Unused fragment | a fragment no scenario includes |

Rules carry a scope. Flow rules (the existing SEC001 to SEC016) check
`Protocol.Flatten()` for a document with scenarios; document rules
receive the protocol as written and report scenario paths such as
`scenarios[1].flows[0] (exchange)`.

### 8. Composition (`resolve.go`)

- `mergeBase`: base scenarios come first. A derived scenario with the
  same `id` replaces the base one in its base position, so an override
  never reorders scenarios. New derived scenarios follow.
- `ExcludeScenarios` runs before the merge. `ExcludeEntities` also drops
  scenario flows that reference an excluded entity and removes it from
  `participants`. `ExcludePhases` also clears phase references in
  scenario flows.
- `mergeImport`: `ProtocolImport.Scenarios []string` selects scenarios
  to import. Fragments they include come with them, in the imported
  document's order. The alias prefix applies to scenario IDs, include
  targets, entity references, participants and phases.
- Mixing documents with and without scenarios: when composition leaves
  both top-level flows and scenarios, the flows become the first
  scenario, with the resolved protocol's ID and name (plus a `-flows`
  suffix if that ID is taken).
- The JSON Schema does not yet describe `extends` or `imports`; that gap
  predates this initiative.

### 9. Diff (`compare.go`)

- Scenarios are matched by `id`. The diff reports added, removed and
  renamed scenarios (same id, new name), as well as changed flows within
  each matched scenario, using the existing flow diff.
- Order changes are reported separately for entities and scenarios
  ("moved from position 3 to 1").

### 10. Schema and specification

- `schema/pidl.schema.json`:
  - The top-level `required` drops `flows`.
  - It gains a `oneOf` requiring either a non-empty `flows` or a
    non-empty `scenarios`.
  - It gains the `scenario` and `tokenClaim` definitions, plus the new
    fields on `flow`, `flowSecurity`, `tokenDefinition` and `extends`.
- `docs/SPECIFICATION.md` gets a "Scenarios" section and an "Ordering"
  section with the rules from section 2.

## Sequencing Constraints

- Types and validation come before projection, and projection before any
  renderer or CLI change.
- Rendering includes depends on projection without expansion.
- The analyzer rules depend on the verification fields.
- The AAuth example depends on includes, verification fields and
  rendering. It serves as the end-to-end test.

## Risks

- **Byte-for-byte compatibility.** Changing a renderer for includes
  could alter output for legacy files. Mitigation: the existing golden
  files must pass unchanged, and include rendering only runs when a
  flow has `Include` set.
- **Unused-participant declarations.** Projection drops entities a
  scenario does not use. Legacy files are not projected, so their output
  is unchanged.
- **Index re-basing on expansion.** Errors here would misreport
  validation paths. Mitigation: table tests covering nested includes
  with parallel branches.

## Testing Strategy

- **Unit tests:**
  - ordering helpers
  - each validation rule, both pass and fail cases
  - projection with and without expansion, including nested fragments
  - cycle detection
- **Golden files:** each sequence format for a collapsed include, an
  expanded include and the AAuth main scenarios.
- **CLI table tests:** `--scenario`, `--all-scenarios` file naming and
  ordering, `scenarios` listing, and errors.
- **Determinism:** render every format twice and compare the outputs.
  Also shuffle map-backed metadata by re-encoding the input and confirm
  the output is unchanged.
- **Schema:** every file in `examples/` validates against the updated
  schema.
