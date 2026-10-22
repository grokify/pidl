# PRD — PIDL v1 format and the system view

**Initiative:** `INIT-PIDL-008`
**Repository:** `github.com/grokify/pidl`

## Problem

PIDL began as a way to draw one protocol flow. It now feeds a step-by-step viewer, narrated video, Postman collections and code samples, and it is being asked to describe more than protocols: AI workflows such as retrieval-augmented generation (RAG), agent orchestration and system architectures. The source format has not kept up:

- **Two ways to write steps.** A document has either a top-level `flows` list or `scenarios`. The viewer, walkthroughs and Postman need scenarios, so most existing documents cannot use them, and every renderer carries both code paths.
- **Process documents model steps as entities.** In a process document, `ingest` and `llm_analysis` are entities and flows are the hand-offs between them. That draws a data-flow graph, but it cannot show a system: fixed components (an application, an embedder, a vector store, a model) with steps passing between them. A RAG explorer needs exactly that.
- **Presentation mixed into meaning.** Animation settings and network colors sit on flows and networks, beside what the protocol does.
- **Optional step IDs.** Sample sets, viewer links and any view that names a step depend on step IDs, but a document may omit them.
- **No version marker.** The walkthrough and sample-set formats carry `format`, the source format does not, so a tool cannot tell an old document from a new one.
- **snake_case names.** Other formats in this ecosystem moved to camelCase for JavaScript and JSON Schema; PIDL and its walkthrough are the exception.
- **A hand-written schema.** `schema/pidl.schema.json` is maintained by hand, so it drifts from the Go types.

Separately, readers asked for a second view of a protocol: the system diagram a specification draws (who talks to whom), kept in place while the steps advance, so they keep one picture in mind instead of reading a long sequence diagram.

PIDL is pre-1.0 and its documents are maintained by the same people who maintain PIDL, so this is the time to make one deliberate breaking change, with a tool that upgrades every existing document.

## Goals

- **One source format, `pidl/v1`,** used for protocols and processes alike: participants, scenarios of steps, and views.
- **Participants and steps kept apart** for every kind: participants are the things that exist; steps are what happens between them, including the process details (step type, inputs and outputs, retries, failure modes, cost).
- **Meaning separate from presentation.** The document's semantic part has no coordinates, colors or animation; those live in `views`, and colors live with the host page.
- **A system view,** authored as a small grid layout, with links derived from the steps, shown beside the sequence diagram and kept in step with it.
- **camelCase** property names and enumeration values, in the source format and the walkthrough format.
- **A migration tool** that upgrades every existing document, mechanically for protocol documents and with review notes where a process document needs a decision.
- **A generated source schema,** from the Go types, as the walkthrough schema already is.
- **A broader description:** PIDL, the **Process & Interaction Description Language**.

## Non-Goals

- Renaming the repository, the Go module or the npm package.
- Free-form coordinates, named per-device layouts or automatic layout engines. The grid comes first; these can follow if a document needs them.
- A style language inside PIDL. Colors, fonts and themes belong to the page hosting the viewer, through documented CSS custom properties. A design system can target those properties; PIDL depends on none.
- Profiles for model architectures or training pipelines (layers, tensors, gradients). They need their own design and are a later initiative.
- An inspector for steps that are not HTTP messages, and a RAG example. They build on this format and are the next initiative.
- YAML input. It can be added later without changing the format.

## Requirements

### Format

1. **Format identifier.** A document starts with `"format": "pidl/v1"`. A reader refuses another major version with a message naming the version it found.
2. **Flat header.** `id`, `name`, `kind`, `version`, `description`, `category`, `references` and `language` sit at the top level, as in the walkthrough, instead of under a `protocol` object.
3. **Participants.** `participants` replaces `entities`. A participant may carry `status: "proposed"` for a role that a draft describes but has not adopted, and `network` for its trust boundary.
4. **Scenarios only.** Steps live in `scenarios[].steps`. The top-level `flows` list is gone; a document with one flow has one scenario. Fragments stay as scenarios with `fragment: true`.
5. **Required step IDs.** Every step has an `id`, unique within its scenario.
6. **Process details on steps.** `stepType`, `inputs`, `outputs`, `processing`, `failureModes`, `retryStrategy` and `parallel` move from the participant to the step. A step with only `from` is work done by that participant.
7. **Order is array order.** `sequence` is removed.
8. **Top-level definitions.** `tokens`, `networks`, `components` and `trustRelations` move from `metadata` to the top level. Network colors move to views.
9. **camelCase.** Property names and enumeration values are camelCase: `variantOf`, `trustLevel`, `semiTrusted`, `toolCall`, `authorizationServer`. Identifiers authors choose (participant and step IDs) are free-form.
10. **Generated schema.** `schema/pidl.schema.json` is generated from the Go types, embedded, and checked for drift, with `schemakit lint --property-case camelCase` passing.

### Views

11. **Views block.** `views` holds presentation only. Removing it changes no analysis, validation result or walkthrough content apart from the views themselves.
12. **System layout.** `views.system.layout` is a grid: a list of rows, each a list of cells holding a participant ID or `null`. A participant named in adjacent cells of one row spans them.
13. **Derived links.** The system view has one link for each pair of participants that exchange a step, carrying the IDs of those steps in order. The author writes no link for them.
14. **Extra links and labels.** `views.system.links` adds links that are not steps, such as a setup relationship drawn dashed, and sets a summary label on a derived link, such as "signed request" for every message between an agent and a resource.
15. **Sequence presentation.** Per-step animation moves to `views.sequence.steps.<id>.animation`.
16. **Outputs.** `pidl generate -f svg-system` and `-f system-text` (an ASCII drawing) render the system view for a scenario or for the whole document.

### Walkthrough and Viewer

17. **Walkthrough v2.** `pidl-walkthrough/v2` uses camelCase, names `participants` as the source does, carries the system view (layout, links and their step IDs) and gives each step its link ID.
18. **Reading v1.** The viewer reads v1 and v2 walkthroughs, so pages pinned to an older walkthrough keep working with a new viewer.
19. **System view in the viewer.** A Sequence / System switch. The system view keeps every participant in place and offers three modes: the current step only, every step so far, or every link. The current link and its two participants are highlighted; a proposed participant and an extra link are drawn dashed.
20. **Styling contract.** The viewer's colors, fonts and line styles come from documented `--pidl-*` CSS custom properties, which a host page sets on the viewer's container (custom properties pass into the viewer's shadow root). Diagram elements carry stable classes for participant type, trust level, `proposed`, `current` and `done`.

### Migration

21. **`pidl migrate`.** Upgrades a document written before v1 to `pidl/v1`: the header, names and values, `flows` into one scenario, missing step IDs from the action, process details from participants to steps, metadata to the top level, and animation and colors into views. It prints a note for every change that needs review, writes in place with `-w`, and with `--check` exits non-zero when a document needs migrating.
22. **Reading old documents.** For the v0.10 series, reading a document without `format` migrates it in memory and reports that it should be migrated. v1.0.0 removes this.
23. **Every built-in example** is migrated, and its walkthrough and diagrams regenerated.

### Documentation

24. The specification describes `pidl/v1`; a migration guide lists every rename and move with before-and-after examples; a guide covers the system view and the styling contract; the README, docs site and roadmap use the new description.

## Success Criteria

- `pidl migrate --check` passes on every built-in example and on every document in the projects that consume PIDL.
- The AAuth document gains a system view matching the draft's architecture diagram, with the Supervision Server as a proposed participant, and the explorer switches between the sequence and system views on the same step.
- A host page restyles the viewer by setting `--pidl-*` properties only.
- The generated source schema validates every built-in example.
