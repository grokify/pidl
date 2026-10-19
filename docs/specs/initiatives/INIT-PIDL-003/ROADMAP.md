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

- [ ] `RMI-PIDL-042` Narration on scenarios and flows, and protocol pronunciations
- [ ] `RMI-PIDL-043` Samples: inline code or source references to tested files, with region extraction and `pidl samples resolve` and `check`
  - Depends on: `RMI-PIDL-040`
- [ ] `RMI-PIDL-044` Generated cURL for example requests
- [ ] `RMI-PIDL-045` Walkthrough data carries groups, variants, IDs, narration, samples, cURL and token lifecycle events
  - Depends on: `RMI-PIDL-042`, `RMI-PIDL-043`, `RMI-PIDL-044`

## Phase 3 — Example Checks

**Theme:** Examples that can be trusted

- [ ] `RMI-PIDL-046` Generic example checks: digests, encoded and decoded tokens, covered components, timestamps
- [ ] `RMI-PIDL-047` Check profiles and the `aauth-draft-11` profile
  - Depends on: `RMI-PIDL-046`
- [ ] `RMI-PIDL-048` Profile selection in documents and `pidl analyze --profile`
  - Depends on: `RMI-PIDL-047`

## Phase 4 — Viewer and Release

**Theme:** Navigate, play and read in your language

- [ ] `RMI-PIDL-049` Grouped left navigation and variant toggle
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-050` Autoplay: play, pause, replay and speed, timed from narration
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-051` Token lifecycle panel and decoded/raw toggle
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-052` Representation tabs: HTTP, cURL and sample languages, remembered per browser
  - Depends on: `RMI-PIDL-045`
- [ ] `RMI-PIDL-053` Documentation and v0.9.0 changelog and release notes
  - Depends on: `RMI-PIDL-049`, `RMI-PIDL-050`, `RMI-PIDL-051`, `RMI-PIDL-052`
