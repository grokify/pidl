# PRD — Ordered roles and multiple scenarios per protocol

**Initiative:** `INIT-PIDL-001`
**Repository:** `github.com/grokify/pidl`

## Problem

A PIDL file describes one protocol as a list of entities and a single
ordered list of flows. Real protocols are not one flow. AAuth, for
example, has seven roles, five access modes and about eight shared
sub-flows (agent bootstrap, discovery, interaction, deferred response,
mission lifecycle, step-up, call chaining, revocation). Documenting it
today means one file per flow, tied together with `imports` or
`extends`. That leaves four gaps:

1. **No shared order.** Each file decides its own entity order, and
   there is no order across flows. A slide deck, a guide and a set of
   diagrams generated from those files can each present roles and flows
   in a different sequence.
2. **No reuse inside a flow.** A shared sub-flow such as "interaction"
   has to be copied into every flow that uses it, or left out.
3. **No checks across flows.** Validation runs per file, so it cannot
   report a role that no flow uses or a token that is used before any
   flow issues it.
4. **No overview outputs.** Nothing renders a roles × flows table or a
   token table for the whole protocol.

`Flow.sequence` adds confusion: it is documented as "explicit ordering"
but no code sorts by it, so its value never affects the output.

## Goals

- One PIDL file can hold a protocol's roles and an ordered list of
  named **scenarios**, each with its own ordered flows.
- Array position is the only source of order, for roles, scenarios,
  phases and flows. Every output (diagram, table, outline) shows the
  same order.
- A scenario can include another scenario as a reusable fragment,
  shown as a labeled box or expanded in place.
- Validation and analysis work across all scenarios in a file.
- New outputs give an overview of the whole protocol: a roles ×
  scenarios matrix, a token table and an ordered outline that guides and
  presentations can be built from.
- Existing files keep working without changes.

## Non-Goals

- Generating slide decks or guides directly. PIDL emits ordered,
  structured outputs. Deck and guide tools consume them.
- A "journey" concept tied to human users or UX. Scenarios cover flows
  with or without a person. A display `name` can say "journey" where an
  audience prefers it.
- Message schemas or transport details. These stay out of scope as
  stated in the main roadmap.
- Numeric ordering fields. Order comes from array position only.

## Users and Stakeholders

- **Protocol reviewers** who need one diagram per scenario, plus a
  token table that shows who issues and who checks each token.
- **Developer relations and educators** who build presentations and
  guides and need roles and scenarios in the same order everywhere.
- **Spec authors** who track how a draft changes between versions.
  `pidl diff` should report scenario-level changes.
- **Tool builders** who use the Go library and need a stable API for
  one scenario at a time.

## Requirements

### Functional

1. **Scenarios.** A protocol file MAY contain a top-level `scenarios`
   array. Each scenario has an `id`, a `name`, an optional
   `description`, an optional `participants` list, an optional
   `fragment` flag and a `flows` array.
2. **Includes.** An entry in a scenario's `flows` array MAY be an
   include step (`{"include": "<scenario-id>"}`) instead of a message.
3. **Ordering rule.**
   - The order of the `entities` array sets the order of roles everywhere.
   - The order of the `scenarios` array sets the order of scenarios.
   - The order of the `phases` array sets the order of phases.
   - The order of flows within a scenario sets the order of messages.
4. **Participants follow entity order.** A scenario's `participants`
   list selects roles. It never reorders them. A list that is out of
   order is a validation error.
5. **Default participants.** When `participants` is omitted, a
   scenario's participants are the entities its flows and includes
   use, in `entities` order.
6. **Fragments.** A scenario with `"fragment": true` can be included
   by other scenarios. Commands that render every scenario skip
   fragments unless asked to include them.
7. **Backward compatibility.**
   - A file with a top-level `flows` array and no `scenarios` behaves
     exactly as it does today.
   - A file that has both a non-empty `flows` array and a `scenarios`
     array is invalid.
8. **Single-scenario projection.** The library can turn one scenario
   into a single-flow `Protocol`, so every existing renderer, the
   executor and the debugger work on it unchanged.
9. **CLI selection.** `generate`, `simulate`, `debug` and `analyze`
   accept `--scenario <id>`.
10. **Rendering every scenario.** `generate --all-scenarios -o <dir>`
    writes one file per scenario. Each file name starts with a
    zero-padded position, such as `01-agent-identity.mmd`, so sorting
    the files by name keeps scenario order.
11. **Collapsed includes.** By default, a renderer draws an include
    as a labeled box over the fragment's participants. `--expand`
    inlines the fragment's flows instead.
12. **Overview outputs:**
    - a roles × scenarios matrix
    - a token table: issuer, audience, binding, who checks it and the
      claims they rely on
    - an ordered outline (Markdown and JSON) listing roles, then
      scenarios with their steps
13. **Verification semantics.**
    - `TokenDefinition` records who checks each token and which claims
      they rely on.
    - `FlowSecurity` records who signs a request and which tokens the
      receiver checks.
14. **Checks across scenarios.** The analyzer reports:
    - roles used by no scenario
    - tokens checked in a flow before any earlier step issued them
    - flows signed by someone other than the sender, unless annotated
    - token checkers that are not participants of the scenario
15. **Composition.**
    - `extends` and `imports` merge scenarios by `id`, base scenarios
      first.
    - `exclude_scenarios` removes scenarios from the base.
16. **Diff.** `pidl diff` matches scenarios by `id` and reports
    scenarios that were added or removed, changed flows, and changes in
    the order of roles or scenarios.
17. **`Flow.sequence` deprecation.** The field is documented as display
    only, and no code may sort by it.
18. **Example.** The built-in examples include AAuth, with seven roles,
    five main scenarios and its shared fragments.

### Non-Functional

- **Deterministic output.** The same input always produces identical
  output. No output order may depend on Go map iteration.
- **No change for existing files.** For every existing example,
  `generate` output stays byte-for-byte the same.
- **Library first.** All behavior lives in the `pidl` and `render`
  packages. The CLI only parses flags and calls them.
- **The JSON Schema** in `schema/pidl.schema.json` describes the new
  fields and validates every example.

## Success Criteria

- The AAuth example validates and renders five main scenario diagrams
  in Mermaid, PlantUML and D2, with roles in the same column order.
- `--all-scenarios` output and the outline list scenarios in the same
  order as the file.
- The golden-file tests for existing examples pass unchanged.
- A four-party AAuth scenario annotated with `signed_by` makes the
  analyzer report that the PS, not the agent, signs the request to the
  AS.

## Risks

- **Flow index references.** `exclude_flows` and
  `ParallelBranch.flow_refs` use indices into the top-level `flows`
  array. Inside scenarios they need a scenario scope. Mitigation: index
  within the scenario, and reject index references that don't say which
  scenario they mean.
- **Collapsed include rendering differs by format.** Mermaid has no
  `ref` fragment. Mitigation: use `rect` plus a note in Mermaid, `ref
  over` in PlantUML and a group in D2, and test each with golden files.
- **Schema growth.** New optional fields make the format larger.
  Mitigation: all are optional, and the scenario fields are grouped
  together in the specification.

## Decisions

- **"Scenario" rather than "journey".** "Journey" usually means a human
  user's path through a product. Many protocol scenarios have no
  person in them (for example agent-to-resource identity). "Scenario"
  is the common term in specs and testing, and it fits PIDL's
  simulator ("run scenario four-party"). Display names may still say
  "journey".
- **Array position is the only order.** A second mechanism (numeric
  `order` or `sequence` fields) lets outputs drift apart.
- **Fragments are scenarios.** Shared sub-flows use the same type, so
  the ordering and validation rules need no special cases.
- **Collapsed by default.** Presentations and overview diagrams read
  better with a sub-flow drawn as one labeled box. `--expand` serves
  detailed review.
