# ROADMAP — Protocol walkthrough viewer and AAuth examples

**Initiative:** `INIT-PIDL-002`
**Repository:** `github.com/grokify/pidl`

A generic PIDL viewer that walks through each scenario step by step, with example requests, responses and decoded tokens, and a fully worked AAuth draft-11 example. Commits implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — Examples and Walkthrough Model

**Theme:** Message content in PIDL and a ready-to-render walkthrough

- [x] `RMI-PIDL-022` Flow `example` type: source, kind, request, response and decoded tokens, with validation
- [x] `RMI-PIDL-023` JSON Schema for flow examples
  - Depends on: `RMI-PIDL-022`
- [x] `RMI-PIDL-024` Walkthrough model: expanded steps with fragment trail, written locations and attached findings
  - Depends on: `RMI-PIDL-022`
- [x] `RMI-PIDL-025` `walkthrough-json` format with golden tests
  - Depends on: `RMI-PIDL-024`

## Phase 2 — AAuth Examples

**Theme:** Every AAuth step with a request, response and tokens from draft-11

- [x] `RMI-PIDL-026` AAuth fragments: agent token, person token and consent examples, with the shared values used across scenarios
  - Depends on: `RMI-PIDL-022`
- [x] `RMI-PIDL-027` AAuth agent identity and resource-managed examples
  - Depends on: `RMI-PIDL-026`
- [x] `RMI-PIDL-028` AAuth person identity and PS authorization examples
  - Depends on: `RMI-PIDL-026`
- [x] `RMI-PIDL-029` AAuth federated authorization examples
  - Depends on: `RMI-PIDL-028`

## Phase 3 — Viewer

**Theme:** Step through a scenario in a 16:9 layout

- [ ] `RMI-PIDL-030` Viewer scaffold: Vite, React, TypeScript, Vitest, data loading
  - Depends on: `RMI-PIDL-025`
- [ ] `RMI-PIDL-031` Sequence diagram with fragment bands and active-step highlighting
  - Depends on: `RMI-PIDL-030`
- [ ] `RMI-PIDL-032` Stepping: buttons, keyboard, URL hash, scenario selector
  - Depends on: `RMI-PIDL-031`
- [ ] `RMI-PIDL-033` Request, response and token panels with provenance badges, notes and findings
  - Depends on: `RMI-PIDL-032`
- [ ] `RMI-PIDL-034` 16:9 stage scaling, light and dark themes
  - Depends on: `RMI-PIDL-033`

## Phase 4 — Single-File Output and Release

**Theme:** One HTML file per protocol

- [ ] `RMI-PIDL-035` Single-file build embedded in the Go binary and `generate -f viewer`, with a CI freshness check
  - Depends on: `RMI-PIDL-034`
- [ ] `RMI-PIDL-036` Documentation: viewer guide, example field in the specification, README and CLI reference
  - Depends on: `RMI-PIDL-035`
- [ ] `RMI-PIDL-037` v0.8.0 changelog and release notes
  - Depends on: `RMI-PIDL-036`
