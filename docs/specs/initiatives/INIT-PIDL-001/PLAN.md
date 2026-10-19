# PLAN — Ordered roles and multiple scenarios per protocol

**Initiative:** `INIT-PIDL-001`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

- The PRD and TRD are approved, including the naming decision
  ("scenario") and the ordering rule (array position only).
- The current golden-file tests pass on `main`. They are the baseline
  for byte-for-byte compatibility.

## Build Order

### Phase 1 — Model and ordering

Add the `Scenario` type, include steps, ordering helpers, validation and
single-scenario projection. Write the "Ordering" section of the
specification first, so that every later item follows the same rules.
At the end of this phase, a file with scenarios parses, validates and
projects. Legacy files behave as before.

### Phase 2 — Rendering and CLI

Add `--scenario`, `--all-scenarios` with ordered file names, collapsed
include rendering in every sequence format, and the `scenarios`
command. At the end of this phase, every main scenario of a
multi-scenario file can be rendered.

### Phase 3 — Overview outputs and verification semantics

Add the scenario matrix, token table and outline outputs. Add the
verification fields on tokens and flow security, and the analyzer rules
that use them. At the end of this phase, a protocol's roles, tokens and
scenarios can be documented in one consistent order.

### Phase 4 — Composition, diff, example, release

Make `extends`, `imports` and `diff` scenario-aware. Add the AAuth
example as the end-to-end test. Update the schema, specification, guide
and changelog, and release v0.9.0.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Legacy output changes | Golden files for existing examples gate every PR |
| Include boxes render differently by format | One golden file per format, plus visual review of the AAuth diagrams |
| Index-based references break under expansion | Re-base indices during projection; nested-include table tests |
| The AAuth draft changes during the work | Pin the example to draft -11 in its references; `pidl diff` covers later drafts |

## Follow-On

- Deck and guide generators that read `outline-json`, such as a Marp
  deck template.
- A `pidl migrate` command that wraps a legacy `flows` file into a
  single scenario.
- Splitting the legacy `metadata.networks` map into an ordered array in
  the next schema major version.
