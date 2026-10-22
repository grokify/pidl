# TRD — PIDL v1 format and the system view

**Initiative:** `INIT-PIDL-008`
**Repository:** `github.com/grokify/pidl`

## Document Shape

```json
{
  "format": "pidl/v1",
  "id": "aauth-protocol",
  "name": "AAuth",
  "kind": "protocol",
  "version": "draft-hardt-oauth-aauth-protocol-11",
  "references": [{"name": "AAuth Protocol", "url": "https://datatracker.ietf.org/doc/draft-hardt-oauth-aauth-protocol/"}],
  "language": "en",

  "participants": [
    {"id": "person", "name": "Person", "type": "user"},
    {"id": "supervision-server", "name": "Supervision Server", "type": "authorizationServer", "status": "proposed"},
    {"id": "person-server", "name": "Person Server", "type": "authorizationServer"},
    {"id": "access-server", "name": "Access Server", "type": "authorizationServer"},
    {"id": "agent-provider", "name": "Agent Provider", "type": "identityProvider"},
    {"id": "agent", "name": "Agent", "type": "agent"},
    {"id": "resource", "name": "Resource", "type": "resourceServer"}
  ],
  "groups": [ ... ],
  "phases": [ ... ],
  "scenarios": [
    {
      "id": "agent-identity",
      "name": "Agent identity",
      "steps": [
        {"id": "agent-token", "from": "agent-provider", "to": "agent", "action": "issue_agent_token", "label": "agent token"},
        {"id": "signed-request", "from": "agent", "to": "resource", "action": "signed_request", "label": "signed request"},
        {"id": "resource-response", "from": "resource", "to": "agent", "action": "resource_response", "mode": "response"}
      ]
    }
  ],
  "tokens": [ ... ],
  "networks": { ... },
  "components": [ ... ],
  "trustRelations": [ ... ],
  "pronunciations": { ... },

  "views": {
    "system": {
      "layout": [
        [null,                 "person",        "person",  null],
        ["supervision-server", "person-server", null,      "access-server"],
        ["agent-provider",     "agent",         null,      "resource"]
      ],
      "labels": [
        {"from": "agent", "to": "resource", "label": "signed request"},
        {"from": "resource", "to": "agent", "label": "resource response or resource token"}
      ],
      "links": [
        {"id": "confirm-ps", "from": "person", "to": "supervision-server", "label": "confirm PS", "both": true, "line": "dashed"}
      ]
    },
    "sequence": {
      "steps": {"agent-identity/signed-request": {"animation": {"preset": "request"}}}
    }
  }
}
```

The JSON is illustrative; the generated schema is normative.

## Go Model

| v0 | v1 | Notes |
|----|----|-------|
| `Protocol` | `Document` | `type Protocol = Document` kept, deprecated, through the v0.10 series |
| `ProtocolMeta` (`protocol` object) | fields on `Document` | |
| `Entity` (`entities`) | `Participant` (`participants`) | `type Entity = Participant`, deprecated |
| `Flow` (`flows`) | `Step` (`steps`) | `type Flow = Step`, deprecated |
| `Protocol.Flows` | removed | a single-flow document is one scenario |
| `Scenario.Flows` | `Scenario.Steps` | |
| `Flow.Sequence` | removed | order is array order |
| `Flow.Animation` | `Views.Sequence.Steps[addr].Animation` | |
| `Entity.StepType`, `Inputs`, `Outputs`, `Processing`, `FailureModes`, `RetryStrategy`, `Parallel` | same fields on `Step` | |
| `Entity.Metadata.Network` | `Participant.Network` | |
| — | `Participant.Status` | `""` or `proposed` |
| `ProtocolMetadata.Tokens`, `Networks`, `Components`, `TrustRelations` | fields on `Document` | `metadata` removed |
| `ProtocolMetadata.NetworkLayout`, `NetworkConfig.Color` | `Views.Network` | presentation |
| `ProtocolExtends.ExcludeFlows []int` | `ExcludeSteps []string` | step addresses |
| `ProtocolImport.Entities`, `IncludeFlows` | `Participants`, `IncludeScenarios` | |

`Document.Format` is required and set to `pidl.Format` (`"pidl/v1"`) by constructors.

### Step semantics

A step moves something from `from` to `to`, or, with no `to`, is work `from` does alone. Its process details (`stepType`, `inputs`, `outputs`, `processing`, `failureModes`, `retryStrategy`, `parallel`) describe the work of the participant that receives it: `to`, or `from` when there is no `to`. Data mappings connect an earlier step's output to this step's input:

```json
{"id": "analyze", "from": "app", "to": "model", "stepType": "llm",
 "inputs": [{"name": "context"}], "outputs": [{"name": "answer"}],
 "dataMappings": [{"from": "retrieve.chunks", "input": "context"}]}
```

Lineage, cost, latency, parallel execution and the process executor read steps instead of entities. A parallel block lists step IDs.

### Addresses

A step address is `<scenario>/<step-id>`, as sample sets already use. Inside a fragment, the included steps are `<include>.<step-id>`. `views.sequence.steps`, `excludeSteps` and the system view's step lists use addresses.

### Naming

Every JSON property name is camelCase. Enumeration values are camelCase too: `semiTrusted`, `toolCall`, `toolResult`, `authorizationServer`, `agentProvider`, `jwtSvid`. Go constant names do not change. The migrator's rename table is generated from a list of `{old, new}` pairs kept beside the constants, and a test fails when a constant's value is not camelCase or has no entry.

## Views

```go
type Views struct {
	System   *SystemView   `json:"system,omitempty"`
	Sequence *SequenceView `json:"sequence,omitempty"`
	Network  *NetworkView  `json:"network,omitempty"`
}

type SystemView struct {
	Layout [][]*string   `json:"layout"`
	Labels []LinkLabel   `json:"labels,omitempty"`
	Links  []ExtraLink   `json:"links,omitempty"`
}

type LinkLabel struct{ From, To, Label string }

type ExtraLink struct {
	ID, From, To, Label string
	Both bool            // drawn with arrows at both ends
	Line string          // "solid" (default) or "dashed"
}
```

### Validation

- Every participant appears in the layout at most once per row and in one row; cells naming one participant are adjacent. A participant missing from the layout is a warning, and the renderers place it in a row of its own below the grid.
- Every ID in `labels` and `links` names a participant; a label's pair must exchange at least one step; an extra link's ID is unique and differs from every derived link ID.
- `views.sequence.steps` keys are valid step addresses.
- Removing `views` changes no other validation result (a test checks this for every example).

### Derived links

For each pair of participants that exchange at least one step, in any scenario after fragments are expanded, there is one link:

- **ID:** `<a>~<b>`, where `a` comes before `b` in `participants`.
- **Directions:** the steps from `a` to `b`, and the steps from `b` to `a`, each as an ordered list of addresses.
- **Label per direction:** the `labels` entry for that direction, or else the distinct step labels (or actions) in order, joined with ", ".

Steps with no `to` belong to no link; the view highlights their participant instead.

### Rendering

`render/system.go` turns a document and an optional scenario into a `SystemDiagram`: participants with grid positions and spans, and links with directions and labels. The two renderers share it.

- **`svg-system`:** boxes on the grid; each link drawn between the two boxes' centers and clipped at their edges; the two directions of one link drawn as parallel arrows a fixed offset apart, each labelled on its own side; extra links and proposed participants dashed. With `--scenario`, only that scenario's links are drawn. Links are drawn straight; authors avoid crossings through the grid.
- **`system-text`:** the same diagram on a character grid. Links between participants in one row or column are straight; others take one elbow. Labels sit beside the line.

## Walkthrough v2

`pidl-walkthrough/v2` changes:

- camelCase throughout (`variantOf`, `variantLabel`, `signedBy`, `verifiedBy`, `checkedBy`).
- `roles` becomes `participants`, and each participant carries `status`.
- A top-level `system` object, present when the document has a system view: `layout`, and `links`, each with `id`, `from`, `to`, `labels` per direction, `steps` per direction (addresses as the walkthrough numbers them), and `extra`, `both`, `line` for extra links.
- Each step carries `link`, the ID of its derived link.

The viewer reads v1 by renaming its five snake_case properties and `roles` on load. Its TypeScript types are generated from the v2 schema. `samples.schema.json` has no multi-word properties and stays `pidl-samples/v1`.

## Viewer

- **View switch.** Sequence / System, shown when the walkthrough has `system`; kept in the URL hash like the step.
- **Modes.** Current step (default), steps so far, all links. In the first two, links and participants not yet involved are drawn faint, not hidden, so the picture never moves.
- **Highlight.** The current step's link direction is drawn in the accent color, and its two participants are outlined. A step without `to` outlines its participant.
- **One state.** The selected step drives the sequence diagram, the system diagram and the request and response panels.
- **Narrow screens.** The system diagram scales to the panel's width; it does not reflow.

## Styling Contract

The viewer already styles itself with custom properties (`--accent`, `--muted`, `--border`, …). v2 renames them to a documented `--pidl-*` set and reads only those:

| Property | Use |
|----------|-----|
| `--pidl-font`, `--pidl-font-mono` | text and code |
| `--pidl-bg`, `--pidl-panel`, `--pidl-text`, `--pidl-muted`, `--pidl-border` | surfaces and text |
| `--pidl-accent`, `--pidl-accent-soft` | current step, links, focus |
| `--pidl-line` | inactive links and lifelines |
| `--pidl-warn`, `--pidl-warn-soft` | findings |
| `--pidl-participant-fill`, `--pidl-participant-stroke` | participant boxes |
| `--pidl-dash` | dash pattern for proposed participants and extra links |

Each has a light and a dark default. Custom properties inherit into the shadow root, so a host page sets them on the viewer's container. Diagram elements carry stable classes and attributes for finer styling through `::part`: `part="participant"` with `data-type`, `data-trust` and `data-status`, and `part="link"` with `data-state` (`current`, `done`, `idle`). A design system can generate these properties from its tokens; PIDL itself depends on none.

## Migration

`pidl.Migrate(data []byte) (*Document, []MigrationNote, error)` reads a document without `format` and returns it as v1, with a note for each change that needs review. `pidl migrate [-w] [--check] file...` wraps it and writes the result with the formatter's indentation. Steps, in order:

1. Wrap `protocol` fields into the top level and add `format`.
2. Rename every property and enumeration value with the rename table.
3. `entities` → `participants`; `metadata.network` → `network`.
4. Top-level `flows` → one scenario, ID and name from the document; a document with both is an error.
5. Give each step without an ID one from its action, de-duplicated within the scenario with `-2`, `-3`, … (note when generated).
6. Process documents: move each participant's process details to the first step it receives, or to a new step with only `from` at the start of the first scenario when it receives none. A participant receiving several steps gets a note naming them. Rewrite data mappings to `<step>.<port>` and parallel blocks to step IDs.
7. Drop `sequence`, after checking it agrees with array order (note when it does not).
8. Move `animation` to `views.sequence.steps`, and network colors and `network_layout` to `views.network`.
9. Move `metadata` fields to the top level.
10. `extends.exclude_flows` indexes → step addresses; imports' `entities`, `include_flows` → `participants`, `includeScenarios`.

`Parse` calls `Migrate` for a document without `format` and returns the notes as warnings, which `pidl validate` and `pidl generate` print with a pointer to `pidl migrate`. v1.0.0 removes this path.

## Schemas

`schema/gen` generates both `pidl.schema.json` (from `pidl.Document`, replacing the hand-written file) and `walkthrough.schema.json` (v2). Each has a drift test and passes `schemakit lint --property-case camelCase`. A root test validates every built-in example against the embedded source schema's top-level fields, as the walkthrough test does.

## Testing

- Golden migration tests: each v0 built-in example, as it stood before migration, kept under `testdata/v0/`, migrates to the committed v1 example with the expected notes.
- Round trip: parse, write and parse again yields the same document for every example.
- Removing `views` leaves validation results and walkthrough content (apart from `system`) unchanged.
- Derived links: fragments, both directions, self-steps, label override, extra links.
- `svg-system` and `system-text` golden files for AAuth-shaped and process examples.
- Viewer: v1 and v2 loading, view switch, the three modes, highlight for self-steps, the `--pidl-*` properties overriding defaults from the host page.
