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
- [ ] `RMI-PIDL-073` Video guide with an end-to-end AAuth render, and release notes
  - Depends on: `RMI-PIDL-072`
