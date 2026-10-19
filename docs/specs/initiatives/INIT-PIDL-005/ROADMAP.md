# ROADMAP — Narrated protocol videos

**Initiative:** `INIT-PIDL-005`
**Repository:** `github.com/grokify/pidl`

PIDL generates per-scenario videoascode configurations and briefing decks; videoascode produces the narrated videos. Commits implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — Video Output via videoascode

**Theme:** One narrated video per scenario, from a PIDL file

- [ ] `RMI-PIDL-069` Viewer video mode and `window.pidl.goTo`, which resolves after the step has rendered
- [ ] `RMI-PIDL-070` `generate -f vac`: one configuration per scenario with `evaluate` steps and narration, plus a combined configuration
  - Depends on: `RMI-PIDL-069`
- [ ] `RMI-PIDL-071` Briefing decks: title, why, key points and network diagram slides with spoken notes, placed before the browser segment
  - Depends on: `RMI-PIDL-070`
- [ ] `RMI-PIDL-085` Tours in video: one `evaluate` focus step per beat with its narration, so scrolling and highlighting follow the voiceover
  - Depends on: `RMI-PIDL-070`, `RMI-PIDL-083`
- [ ] `RMI-PIDL-072` Compatibility test against videoascode's configuration types
  - Depends on: `RMI-PIDL-071`, `RMI-PIDL-085`

## Phase 2 — Narration Audio

**Theme:** The viewer speaks with the same voice as the videos, in every language and representation

- [ ] `RMI-PIDL-091` `generate -f narration`: a videoascode narration set with unit IDs from scenario and flow IDs, languages and representation variants
  - Depends on: `RMI-PIDL-045`, `RMI-PIDL-082`, `RMI-PIDL-090`
- [ ] `RMI-PIDL-092` Viewer audio: speaker and mute button, playback on play and navigation, beat clips with focus, language and representation selection, text-hash checks
  - Depends on: `RMI-PIDL-050`, `RMI-PIDL-052`, `RMI-PIDL-083`
- [ ] `RMI-PIDL-093` `generate -f viewer --audio`: embed the manifest and copy clips for static hosting and local files
  - Depends on: `RMI-PIDL-092`, `RMI-VAC-003`
- [ ] `RMI-PIDL-094` Video configurations per language and representation, with shared voiceover IDs
  - Depends on: `RMI-PIDL-070`, `RMI-PIDL-091`, `RMI-VAC-002`

## Phase 3 — Release

**Theme:** Guide and release

- [ ] `RMI-PIDL-073` Video and narration audio guide with an end-to-end AAuth render, and release notes
  - Depends on: `RMI-PIDL-072`, `RMI-PIDL-093`, `RMI-PIDL-094`, `RMI-VAC-005`
