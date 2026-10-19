# ROADMAP — PIDL for developer content

**Initiative:** `INIT-PIDL-003`
**Repository:** `github.com/grokify/pidl`

Protocol-neutral extensions that make PIDL a source for developer education. Commits implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — Structure

**Theme:** Sections, variants and stable step identity

- [ ] `RMI-PIDL-038` Scenario groups: ordered `groups`, scenario `group`, validation and ordering
- [ ] `RMI-PIDL-039` Scenario variants: `variant_of` and `variant_label`, with validation
  - Depends on: `RMI-PIDL-038`
- [ ] `RMI-PIDL-040` Flow IDs with validation and `#<scenario>/<flow-id>` deep links
- [ ] `RMI-PIDL-041` Schema and specification for groups, variants and flow IDs
  - Depends on: `RMI-PIDL-039`, `RMI-PIDL-040`

## Phase 2 — Narration and Representations

**Theme:** Spoken scripts and the same message in many forms

- [ ] `RMI-PIDL-042` Narration on scenarios and flows, scenario briefings (why, key points, narration), and protocol pronunciations
- [ ] `RMI-PIDL-043` Samples: inline code or source references to tested files, with region extraction and `pidl samples resolve` and `check`
  - Depends on: `RMI-PIDL-040`
- [ ] `RMI-PIDL-044` Generated cURL for example requests
- [ ] `RMI-PIDL-090` Narration keyed by language, and narration and tours per representation
  - Depends on: `RMI-PIDL-042`
- [ ] `RMI-PIDL-045` Walkthrough data carries groups, variants, IDs, narration, samples, cURL and token lifecycle events
  - Depends on: `RMI-PIDL-042`, `RMI-PIDL-043`, `RMI-PIDL-044`, `RMI-PIDL-090`

## Phase 3 — Example Checks

**Theme:** Examples that can be trusted

- [ ] `RMI-PIDL-046` Generic example checks: digests, encoded and decoded tokens, covered components, timestamps
- [ ] `RMI-PIDL-047` Check profiles and the `aauth-draft-11` profile
  - Depends on: `RMI-PIDL-046`
- [ ] `RMI-PIDL-048` Profile selection in documents and `pidl analyze --profile`
  - Depends on: `RMI-PIDL-047`

## Phase 4 — Viewer

**Theme:** Navigate, play and read in your language

- [ ] `RMI-PIDL-049` Grouped left navigation and variant toggle
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-050` Autoplay: play, pause, replay and speed, timed from narration
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-051` Token lifecycle panel and decoded/raw toggle
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-052` Representation tabs: HTTP, cURL and sample languages, remembered per browser
  - Depends on: `RMI-PIDL-045`

## Phase 5 — Explorer Parity

**Theme:** Everything a dedicated protocol explorer shows, defined generically

- [ ] `RMI-PIDL-074` Glossary of headers, claims and terms, with help and category badges in the viewer
- [ ] `RMI-PIDL-075` Computed RFC 9421 signature base and a signature panel
- [ ] `RMI-PIDL-076` Deferred-response timeline derived from 202 steps
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-077` Artifacts: named JSON and Markdown documents with computed digests, and an artifact panel
- [ ] `RMI-PIDL-078` Value tracing across steps, headers, bodies, claims and artifacts
  - Depends on: `RMI-PIDL-077`
- [ ] `RMI-PIDL-079` Comparison tables as viewer pages
  - Depends on: `RMI-PIDL-049`
- [ ] `RMI-PIDL-080` Reference pages, a generated overview page, and scenario spec links
  - Depends on: `RMI-PIDL-049`
- [ ] `RMI-PIDL-081` Copy buttons and a collapsible, resizable details panel

## Phase 6 — Guided Tours

**Theme:** Walk the reader through each part of a step

- [ ] `RMI-PIDL-082` Tour beats on flows: focus regions, narration and holds, with region validation
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-083` Viewer focus hook (`window.pidl.focus`) with tab switching, scrolling and highlighting; autoplay plays beats
  - Depends on: `RMI-PIDL-082`
- [ ] `RMI-PIDL-084` `pidl tour draft`: propose beats and draft narration from the example and glossary
  - Depends on: `RMI-PIDL-074`, `RMI-PIDL-082`

## Phase 7 — Release

**Theme:** Document and ship v0.9.0

- [ ] `RMI-PIDL-053` Documentation and v0.9.0 changelog and release notes
  - Depends on: `RMI-PIDL-049`, `RMI-PIDL-050`, `RMI-PIDL-051`, `RMI-PIDL-052`, `RMI-PIDL-081`, `RMI-PIDL-084`
