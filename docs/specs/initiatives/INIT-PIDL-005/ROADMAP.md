# ROADMAP — Narrated protocol videos

**Initiative:** `INIT-PIDL-005`
**Repository:** `github.com/grokify/pidl`

PIDL generates videoascode configurations; videoascode produces the narrated video. Commits implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — Video Output via videoascode

**Theme:** From a PIDL file to a narrated MP4

- [ ] `RMI-PIDL-069` Viewer video mode: stage on, no pointer, no smooth scrolling or hover, reduced menu bar
- [ ] `RMI-PIDL-070` `generate -f vac`: videoascode browser configuration with narration, pronunciations, voice settings, and scenario or group selection
  - Depends on: `RMI-PIDL-069`
- [ ] `RMI-PIDL-071` Compatibility test against videoascode's configuration types
  - Depends on: `RMI-PIDL-070`
- [ ] `RMI-PIDL-072` Video guide with an end-to-end AAuth render, and release notes
  - Depends on: `RMI-PIDL-071`
