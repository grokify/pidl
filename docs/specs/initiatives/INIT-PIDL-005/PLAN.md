# PLAN — Narrated protocol videos

**Initiative:** `INIT-PIDL-005`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

`INIT-PIDL-003` narration and pronunciations delivered.

## Build Order

1. **Viewer video mode.**
2. **`generate -f vac`** with scenario and group selection, voice settings and the viewer file.
3. **Compatibility test and guide**, with an end-to-end AAuth render.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| videoascode configuration changes | Compatibility test against its types |
| Narration gaps | Fall back to descriptions; `INIT-PIDL-004` adds narration |

## Follow-On

Intro and outro Marp slide segments; chapter markers from segment timings; translated narration.
