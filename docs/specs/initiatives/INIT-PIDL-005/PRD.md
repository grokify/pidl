# PRD — Narrated protocol videos

**Initiative:** `INIT-PIDL-005`
**Repository:** `github.com/grokify/pidl`

## Problem

A PIDL walkthrough explains a protocol step by step in a browser. Developer relations also needs videos (MP4 for YouTube and course platforms) where a narrator introduces each scenario and then walks through it while the diagram advances. Producing them by hand means recording the screen, writing and recording narration, building intro slides, and re-editing whenever the protocol or narration changes.

videoascode (`vac`) already turns a configuration of Marp slides and browser steps with voiceover text into a narrated video: text-to-speech through OmniVoice (ElevenLabs, Deepgram, local F5-TTS), timing each step to its audio, recording, subtitles, audio caching and multiple languages. What is missing is the bridge from PIDL: generated configurations and briefing slides, and a viewer that can be driven to an exact step and captured cleanly.

## Goals

- **One video per scenario**, so a scenario re-renders on its own when its content changes, maps to one lecture or playlist entry, and composes into longer videos.
- Each scenario video opens with **briefing slides** (title, why the scenario exists, key points, network diagram) and then walks the sequence diagram step by step.
- PIDL generates everything videoascode needs: a configuration per scenario, a Marp deck per scenario, the viewer page and diagram images.
- Every captured frame shows the intended step: the viewer exposes a navigation hook that resolves only after the step has rendered.

## Non-Goals

- Text-to-speech, recording, encoding or subtitles in PIDL. videoascode owns all media work.
- Uploading to YouTube.
- Writing narration and briefings. They are content (`INIT-PIDL-004` for AAuth).

## Requirements

1. **Viewer video mode** (`?mode=video`): presentation stage on, pointer hidden, instant scrolling, hover states off, menu bar reduced to the protocol and scenario name.
2. **Navigation hook**: in video mode the viewer exposes `window.pidl.goTo(scenarioId, step)`, which navigates and returns a promise that resolves after the step has rendered and the diagram has scrolled into place. Targets are absolute, so a retried step cannot drift.
3. **`generate -f vac`** writes one videoascode configuration per selected scenario (`<NN>-<scenario-id>.vac.yaml`), each with:
   - a **slide segment** for that scenario's briefing deck;
   - a **browser segment** that opens the viewer at the scenario's overview, holds on it with the scenario narration, then runs one `evaluate` step per walkthrough step calling `window.pidl.goTo` with that step's narration as its voiceover, and holds briefly at the end.
4. **Combined configuration** (`--combined`): one configuration listing the selected scenarios' segments in navigation order, for a single longer video.
5. **Briefing deck** (`generate -f marp-briefing`): a Marp deck per scenario with speaker notes as voiceover:
   - title slide: group and scenario name;
   - why slide: the briefing's `why` prose;
   - key points slide: the briefing's `points`;
   - network slide: the scenario's roles and trust boundaries as an image rendered by PIDL (`svg-network` for the scenario projection).
   Slides without briefing content are omitted; narration falls back to the scenario description.
6. **Selection**: `--scenario`, `--group`, or all main scenarios; fragments only when asked; navigation order throughout.
7. **Viewer location**: a file URL to a generated viewer HTML (default) or an HTTP URL supplied with `--viewer-url`.
8. **Pronunciations and language**: PIDL's pronunciations and `protocol.language` are written into each configuration unchanged.
9. **Voice settings** (`--voice-provider`, `--voice-id`, `--model`) become the configuration's default voice; credentials stay in the environment, where videoascode reads them.
10. **Output layout**: one directory per run with the viewer HTML, per-scenario decks, diagram images and configurations, ready for `vac browser video --config`.

## Success Criteria

- `vac browser video` renders an AAuth scenario from its generated configuration into an MP4 that opens with the briefing slides and then shows each step in sync with its narration, with no frame showing the wrong step.
- Re-running generation after editing one scenario's narration changes only that scenario's files.
- A combined configuration renders several scenarios into one video.

## Decisions

- **videoascode is the media engine**; PIDL generates its inputs.
- **One video per scenario** is the unit; longer videos are compositions.
- **`evaluate` with an absolute navigation hook**, not key presses: videoascode screenshots right after an action, so the action must not return until the step has rendered.
- **Briefings are content in PIDL** (`INIT-PIDL-003`), so slides, viewer and video share one source.
