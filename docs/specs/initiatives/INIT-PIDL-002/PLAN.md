# PLAN — Protocol walkthrough viewer and AAuth examples

**Initiative:** `INIT-PIDL-002`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

- v0.7.0 is released (scenarios, fragments, outline).
- PRD and TRD reviewed.

## Build Order

### Phase 1 — Examples and Walkthrough Model

Add the `example` field with validation and schema, the walkthrough model with findings attached by location, and the `walkthrough-json` format. At the end of this phase any PIDL file produces walkthrough JSON.

### Phase 2 — AAuth Examples

Write the AAuth examples, fragments first, then each access mode. One consistent set of identifiers, keys and token values runs through all scenarios. Each example cites draft-11.

### Phase 3 — Viewer

Scaffold the viewer, then the sequence diagram with highlighting, stepping and URL hash, the request, response and token panels, and the 16:9 stage with scenario selection.

### Phase 4 — Single-File Output and Release

Build the viewer into one HTML file embedded in the Go binary, add `generate -f viewer`, document it, and release v0.8.0.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Examples drift from the draft | Section citations, `draft`/`derived` badges, and a review pass against draft-11 |
| Built viewer out of date | CI rebuilds and compares the committed HTML |
| Node toolchain churn | Versions pinned in `package-lock.json` |

## Follow-On

- Examples for other built-in protocols (OAuth, MCP).
- Presenter notes per step and a speaker view.
- Export a walkthrough as a slide deck.
