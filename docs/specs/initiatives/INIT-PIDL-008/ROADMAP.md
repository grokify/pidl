# ROADMAP — PIDL v1 format and the system view

**Initiative:** `INIT-PIDL-008`
**Repository:** `github.com/grokify/pidl`

PIDL's source format becomes `pidl/v1`: participants and scenarios of steps for protocols and processes alike, camelCase names, and a `views` block that holds presentation, starting with a system view. A migrator upgrades existing documents. Commits implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — The v1 Model

**Theme:** One shape for every document, with meaning kept apart from presentation

- [ ] `RMI-PIDL-114` `Document`, `Participant` and `Step` types: flat header, `format`, scenarios only, required step IDs, `sequence` removed, metadata fields at the top level, deprecated aliases for the old names
- [ ] `RMI-PIDL-115` Process details on steps, with lineage, cost, latency, parallel execution and the process executor reading steps
  - Depends on: `RMI-PIDL-114`
- [ ] `RMI-PIDL-116` camelCase property names and enumeration values, with the rename table and its test
  - Depends on: `RMI-PIDL-114`
- [ ] `RMI-PIDL-117` Renderers, analyzers, Postman and the walkthrough builder ported to the v1 model
  - Depends on: `RMI-PIDL-115`, `RMI-PIDL-116`
- [ ] `RMI-PIDL-118` `schema/pidl.schema.json` generated from `pidl.Document`, with a drift test and the camelCase lint
  - Depends on: `RMI-PIDL-116`

## Phase 2 — Migration

**Theme:** Every existing document upgrades with one command

- [ ] `RMI-PIDL-119` `pidl.Migrate`: the ten migration steps, with review notes
  - Depends on: `RMI-PIDL-117`
- [ ] `RMI-PIDL-120` `pidl migrate` with `-w` and `--check`, and in-memory migration of old documents on read, with a warning
  - Depends on: `RMI-PIDL-119`
- [ ] `RMI-PIDL-121` Built-in examples migrated, with their v0 forms kept as golden inputs, and every generated output regenerated
  - Depends on: `RMI-PIDL-120`

## Phase 3 — Views

**Theme:** Presentation in one place, and a system diagram drawn from the steps

- [ ] `RMI-PIDL-122` `views`: `sequence` (animation), `network` (layout and colors), and participant `status`
  - Depends on: `RMI-PIDL-114`
- [ ] `RMI-PIDL-123` `views.system`: grid layout, derived links, labels and extra links, with validation
  - Depends on: `RMI-PIDL-122`
- [ ] `RMI-PIDL-124` `svg-system` and `system-text` renderers, for the document or one scenario
  - Depends on: `RMI-PIDL-123`

## Phase 4 — Walkthrough v2 and the Viewer

**Theme:** Sequence and system views in step with each other, styled by the host page

- [ ] `RMI-PIDL-125` `pidl-walkthrough/v2`: camelCase, `participants`, the system view and each step's link; schema and generated TypeScript types
  - Depends on: `RMI-PIDL-117`, `RMI-PIDL-123`
- [ ] `RMI-PIDL-126` The viewer reads v1 and v2
  - Depends on: `RMI-PIDL-125`
- [ ] `RMI-PIDL-127` Sequence / System switch, the three modes and the shared step highlight
  - Depends on: `RMI-PIDL-126`
- [ ] `RMI-PIDL-128` Styling contract: the `--pidl-*` properties with light and dark defaults, and the participant and link parts
  - Depends on: `RMI-PIDL-127`

## Phase 5 — Documentation and Release

**Theme:** PIDL, the Process & Interaction Description Language

- [ ] `RMI-PIDL-129` Specification for `pidl/v1`, a migration guide with every rename and move, and a system view and styling guide
  - Depends on: `RMI-PIDL-121`, `RMI-PIDL-124`, `RMI-PIDL-128`
- [ ] `RMI-PIDL-130` The new description in the README, docs site and roadmap; v0.10.0 release notes and changelog
  - Depends on: `RMI-PIDL-129`
