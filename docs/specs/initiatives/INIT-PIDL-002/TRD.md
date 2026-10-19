# TRD — Protocol walkthrough viewer and AAuth examples

**Initiative:** `INIT-PIDL-002`
**Repository:** `github.com/grokify/pidl`

## Current State

- `Flow` has no message content. `Protocol.Outline()` returns ordered roles and scenarios with numbered steps, but keeps include steps collapsed.
- `Protocol.Scenario(id, ProjectOptions{Expand: true})` expands includes but drops which fragment a step came from.
- Analyzer risks carry a `Location` string such as `scenarios[4].flows[6] (federated-authorization)`.
- The repo has TypeScript only in `integrations/`; there is no web app.

## Design

### 1. Flow examples (`types.go`)

```go
type FlowExample struct {
    Source   string            `json:"source"`             // e.g. "draft-11 §7.2.1"
    Kind     ExampleKind       `json:"kind"`               // "draft" | "derived"
    Request  *ExampleRequest   `json:"request,omitempty"`
    Response *ExampleResponse  `json:"response,omitempty"`
    Tokens   []ExampleToken    `json:"tokens,omitempty"`
    Note     string            `json:"note,omitempty"`
}
type ExampleRequest struct {
    Method  string          `json:"method"`
    URL     string          `json:"url"`
    Headers []ExampleHeader `json:"headers,omitempty"`
    Body    json.RawMessage `json:"body,omitempty"`
}
type ExampleResponse struct {
    Status  int             `json:"status"`
    Headers []ExampleHeader `json:"headers,omitempty"`
    Body    json.RawMessage `json:"body,omitempty"`
}
type ExampleHeader struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}
type ExampleToken struct {
    Token   string          `json:"token"`             // token definition ID
    Label   string          `json:"label,omitempty"`   // e.g. "in Signature-Key"
    Header  json.RawMessage `json:"header,omitempty"`
    Payload json.RawMessage `json:"payload,omitempty"`
}
// Flow gains: Example *FlowExample `json:"example,omitempty"`
```

Headers are an ordered array, not a map, so they render in the order written. Bodies, token headers and payloads are `json.RawMessage`, so key order survives a round trip through Go.

### 2. Walkthrough model (`walkthrough.go`)

```go
type Walkthrough struct {
    ID, Name, Description string
    Roles     []OutlineRole
    Scenarios []WalkthroughScenario
    Tokens    []OutlineToken
}
type WalkthroughScenario struct {
    Position int; ID, Name, Description string; Fragment bool
    Participants []string
    Steps []WalkthroughStep
}
type WalkthroughStep struct {
    Number   int      // 1-based within the expanded scenario
    From, To string
    Label    string
    Response bool
    Phase    string
    Fragments []string   // fragment names, outermost first
    Note     string
    Example  *FlowExample
    Location string      // written location, e.g. "scenarios[1].flows[0] (consent)"
    Findings []WalkthroughFinding
}
type WalkthroughFinding struct { ID, Severity, Title, Description string }

func (p *Protocol) Walkthrough(findings func(*Protocol) []Finding) Walkthrough
```

- Steps come from expanding includes in order, recording each enclosing fragment's name.
- `Location` uses the same strings as the analyzer's document rules, so findings attach by location. The `pidl` package cannot import `analyze`, so the caller passes findings in; the CLI and renderer run the analyzer.
- Include steps disappear in expansion; their fragment name appears in `Fragments` of the inlined steps.

### 3. Formats

- `walkthrough-json`: the walkthrough model as JSON, a whole-file format.
- `viewer`: the built viewer HTML with the walkthrough JSON embedded in a `<script type="application/json" id="pidl-walkthrough">` element, a whole-file format.

### 4. Viewer app (`viewer/`)

- Vite 8, React 19, TypeScript, `vite-plugin-singlefile`; Vitest for unit tests.
- Data: reads the embedded `#pidl-walkthrough` element; in development, loads `public/walkthrough.json`.
- Layout: a fixed 1600×900 stage scaled to fit the window (16:9). Top bar: protocol name, scenario selector, step counter and controls. Left (60%): SVG sequence diagram drawn by the app. Right (40%): request panel above, response panel below, each with tabs for HTTP and tokens.
- Diagram: role headers in walkthrough role order, limited to the scenario's participants; one row per step; fragment groups drawn as shaded bands labeled with the fragment name; the active step's arrow, label and role headers highlighted, earlier steps dimmed, later steps faint. The diagram scrolls to keep the active step visible.
- Stepping: buttons, Left/Right arrows, Home/End; URL hash `#<scenario-id>/<step>`.
- Panels: method, URL and headers; JSON body pretty-printed; decoded tokens with claims named in the receiver's `checked_by` highlighted; provenance badge (`draft` or `derived`) with source; notes and findings as callouts.
- Theme: light and dark via `prefers-color-scheme`.

### 5. Build and embedding

- `npm run build` in `viewer/` writes `render/viewer/viewer.html` (committed), which `render` embeds with `go:embed`.
- `go generate ./render/...` runs the build; CI checks that the committed file matches a fresh build.

## Risks

- **Committed build artifact drift.** Mitigation: a CI check rebuilds and diffs.
- **Example accuracy.** Examples could misstate the draft. Mitigation: each cites a section; `draft` versus `derived` is explicit; consistent identifiers are listed once in the example file's description of values.
- **Large diagrams.** Five-party scenarios with fifteen steps must stay legible at 16:9. Mitigation: the diagram scrolls and keeps the active step in view.

## Testing Strategy

- Go: validation table tests for examples; walkthrough expansion with nested fragments, locations and findings; golden files for `walkthrough-json`; `viewer` output contains the embedded data.
- Viewer: Vitest unit tests for layout computation, step navigation and hash parsing.
- Manual: open the AAuth viewer in a browser and step through every scenario; screenshots in the PR.
