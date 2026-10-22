# PLAN — Narrated protocol videos

**Initiative:** `INIT-PIDL-005`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

`INIT-PIDL-003` narration, briefings and pronunciations delivered.

## Build Order

1. **Viewer video mode and `window.pidl.goTo`.**
2. **Per-scenario videoascode configurations** with `evaluate` steps, plus the combined configuration.
3. **Briefing decks** with network diagrams.
4. **Compatibility test, guide and release**, with an end-to-end AAuth render.
5. **Narration audio**: narration export, then viewer playback against a fixture manifest, then packaging once videoascode writes manifests, then multi-language and per-representation video configurations.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| videoascode configuration changes | Compatibility test against its types |
| Frames captured before render | `goTo` resolves after render; `evaluate` awaits it |
| Narration or briefing gaps | Fall back to descriptions; `INIT-PIDL-004` adds content |

## Follow-On

Chapter markers from segment timings; captions highlighted from word timings; outro slides with links.
