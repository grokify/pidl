# ROADMAP — Ordered roles and multiple scenarios per protocol

**Initiative:** `INIT-PIDL-001`
**Repository:** `github.com/grokify/pidl`

One PIDL file holds a protocol's ordered roles and an ordered list of
scenarios. Array position is the only source of order, so diagrams,
guides and presentations generated from the file agree. Commits
implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — Model and Ordering

**Theme:** Scenarios, includes and a single ordering rule

- [x] `RMI-PIDL-002` Specification: "Ordering" section (array position is the only order) and deprecation of `Flow.sequence` as display-only
- [x] `RMI-PIDL-003` `Scenario` type, `Protocol.Scenarios`, `Flow.Include`, and ordering helpers (`EntityIndex`, `OrderEntities`, `ScenarioIDs`)
  - Depends on: `RMI-PIDL-002`
- [x] `RMI-PIDL-004` Scenario validation: unique IDs, flows-or-scenarios, participants in entity order, include targets are fragments, no include cycles, scenario-scoped flow indices
  - Depends on: `RMI-PIDL-003`
- [x] `RMI-PIDL-005` Single-scenario projection `Protocol.Scenario(id, ProjectOptions)` with collapsed and expanded includes
  - Depends on: `RMI-PIDL-004`
- [x] `RMI-PIDL-006` JSON Schema update for scenarios and include steps, with every example validating
  - Depends on: `RMI-PIDL-003`

## Phase 2 — Rendering and CLI

**Theme:** One diagram per scenario, in file order

- [x] `RMI-PIDL-007` `--scenario` on `generate`, `simulate`, `debug` and `analyze`, running on projections
  - Depends on: `RMI-PIDL-005`
- [x] `RMI-PIDL-008` Collapsed include rendering: PlantUML `ref over`, Mermaid `rect` with note, D2 group, SVG box, with golden files
  - Depends on: `RMI-PIDL-005`
- [x] `RMI-PIDL-009` `generate --all-scenarios -o <dir>` with zero-padded ordered file names, `--include-fragments` and `--expand`
  - Depends on: `RMI-PIDL-007`
- [x] `RMI-PIDL-010` `pidl scenarios` command listing scenarios in order (table and JSON)
  - Depends on: `RMI-PIDL-005`
- [x] `RMI-PIDL-011` Deterministic ordering for map-backed metadata (`networks`) in every renderer, with a determinism test
  - Depends on: `RMI-PIDL-002`

## Phase 3 — Overview Outputs and Verification

**Theme:** Roles, tokens and scenarios documented in one order

- [x] `RMI-PIDL-012` `scenario-matrix` output: entities × scenarios in declared order
  - Depends on: `RMI-PIDL-005`
- [x] `RMI-PIDL-013` Verification fields: `TokenDefinition.VerifiedBy`, `TokenClaim`, `FlowSecurity.SignedBy`, `Issues`, `Verifies`, with validation and schema
  - Depends on: `RMI-PIDL-004`
- [x] `RMI-PIDL-014` `token-table` output: issuer, audience, binding, verified by and claims, in declared order
  - Depends on: `RMI-PIDL-013`
- [x] `RMI-PIDL-015` `outline` and `outline-json` outputs: roles, then scenarios with steps, for deck and guide tooling
  - Depends on: `RMI-PIDL-005`
- [x] `RMI-PIDL-016` Analyzer rules: unused role, token checked before issued, signer mismatch, verifier not a participant, unused fragment
  - Depends on: `RMI-PIDL-013`

## Phase 4 — Composition, Diff, Example and Release

**Theme:** Scenario-aware composition and an end-to-end AAuth example

- [x] `RMI-PIDL-017` Scenario-aware `extends` and `imports`: merge by ID with base first, `exclude_scenarios`, import selection with alias prefixing
  - Depends on: `RMI-PIDL-004`
- [x] `RMI-PIDL-018` Scenario-aware `diff`: added, removed and renamed scenarios, per-scenario flow changes, entity and scenario order changes
  - Depends on: `RMI-PIDL-005`
- [x] `RMI-PIDL-019` AAuth example (draft -11): seven roles, five main scenarios and shared fragments, with verification annotations
  - Depends on: `RMI-PIDL-008`, `RMI-PIDL-013`
- [x] `RMI-PIDL-020` Documentation: Scenarios section in the specification, user guide page, README feature list and main roadmap entry
  - Depends on: `RMI-PIDL-019`
- [x] `RMI-PIDL-021` v0.7.0 changelog and release notes, covering scenarios and the pending fixes
  - Depends on: `RMI-PIDL-020`
