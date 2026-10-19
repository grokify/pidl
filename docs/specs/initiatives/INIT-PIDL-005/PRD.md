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
- **The web viewer speaks too.** The narration audio videoascode synthesizes for videos is played in the viewer, in the reader's language and for the representation they chose (HTTP, cURL, Go, TypeScript, Python, C# and others), so the website and the videos sound the same and each clip is paid for once.

## Non-Goals

- Text-to-speech, recording, encoding or subtitles in PIDL. videoascode owns all media work; PIDL only plays the clips it is given.
- Uploading to YouTube.
- Writing narration and briefings. They are content (`INIT-PIDL-004` for AAuth).

## Requirements

1. **Viewer video mode** (`?mode=video`): presentation stage on, pointer hidden, instant scrolling, hover states off, menu bar reduced to the protocol and scenario name.
2. **Navigation hook**: in video mode the viewer exposes `window.pidl.goTo(scenarioId, step)`, which navigates and returns a promise that resolves after the step has rendered and the diagram has scrolled into place. Targets are absolute, so a retried step cannot drift.
3. **`generate -f vac`** writes one videoascode configuration per selected scenario (`<NN>-<scenario-id>.vac.yaml`), each with:
   - a **slide segment** for that scenario's briefing deck;
   - a **browser segment** that opens the viewer at the scenario's overview, holds on it with the scenario narration, then runs one `evaluate` step per walkthrough step calling `window.pidl.goTo` with that step's narration as its voiceover, and holds briefly at the end.
4. **Tours in video.** When a step has a tour, each beat follows the step as its own `evaluate` step calling `window.pidl.focus` with the beat's narration as voiceover, so the panel scrolls to and highlights what the narrator is describing, held for exactly that narration. A long panel is covered by several beats.
5. **Combined configuration** (`--combined`): one configuration listing the selected scenarios' segments in navigation order, for a single longer video.
6. **Briefing deck** (`generate -f marp-briefing`): a Marp deck per scenario with speaker notes as voiceover:
   - title slide: group and scenario name;
   - why slide: the briefing's `why` prose;
   - key points slide: the briefing's `points`;
   - network slide: the scenario's roles and trust boundaries as an image rendered by PIDL (`svg-network` for the scenario projection).
   Slides without briefing content are omitted; narration falls back to the scenario description.
7. **Selection**: `--scenario`, `--group`, or all main scenarios; fragments only when asked; navigation order throughout.
8. **Viewer location**: a file URL to a generated viewer HTML (default) or an HTTP URL supplied with `--viewer-url`.
9. **Pronunciations and language**: PIDL's pronunciations and `protocol.language` are written into each configuration unchanged.
10. **Voice settings** (`--voice-provider`, `--voice-id`, `--model`) become the configuration's default voice; credentials stay in the environment, where videoascode reads them.
11. **Output layout**: one directory per run with the viewer HTML, per-scenario decks, diagram images and configurations, ready for `vac browser video --config`.

12. **Narration export** (`generate -f narration`): a videoascode narration set with one unit per spoken item: scenario narration (`<scenario>`), briefing narration (`<scenario>/briefing`), step narration (`<scenario>/<flow-id>`) and tour beats (`<scenario>/<flow-id>/beat-<n>`), each with its texts by language. Representation-specific narration becomes extra units with that representation as the `variant`. `--lang` and `--representation` limit the set.
13. **Viewer audio.** When the viewer has an audio manifest, the controls show a speaker button (speaker icon when on, muted speaker when off; off by default and remembered per browser):
    - turning it on plays the current step's clip;
    - forward, back, a click on a step or a deep link stop the current clip and play the new step's clip, then its tour beats' clips in order, focusing each region as its clip starts;
    - play advances when the step's clips finish instead of on the estimated duration, and the speed control sets the playback rate;
    - the language selector and the representation tabs switch the on-screen narration and the audio together;
    - a missing clip, or one whose text hash no longer matches the narration, is skipped silently with timed advance, and the step shows that audio is unavailable.
14. **Audio packaging** (`generate -f viewer --audio <manifest>`): the manifest is embedded in the viewer HTML and the clips copied beside it (`audio/clips/…`), referenced by relative path, so the page works from a static host or opened as a local file.
15. **Videos per language and representation**: `generate -f vac --lang en-US,fr-FR,zh-Hans` writes each configuration with segment `transcripts` per language (videoascode's multi-language form, producing one MP4 per language), and `--representation curl` selects that representation's narration, tours and tab, naming files `<NN>-<scenario-id>.<representation>.vac.yaml`. Voiceovers carry the same unit IDs as the narration set, so videos and the viewer share clips.

## Success Criteria

- `vac browser video` renders an AAuth scenario from its generated configuration into an MP4 that opens with the briefing slides and then shows each step in sync with its narration, with no frame showing the wrong step.
- Re-running generation after editing one scenario's narration changes only that scenario's files.
- A combined configuration renders several scenarios into one video.
- With audio on, pressing play on an AAuth scenario in the viewer narrates every step and beat in French when French is selected, and the cURL narration when the cURL tab is selected, using clips videoascode generated once for both the viewer and the videos.

## Decisions

- **videoascode is the media engine**; PIDL generates its inputs.
- **One video per scenario** is the unit; longer videos are compositions.
- **`evaluate` with an absolute navigation hook**, not key presses: videoascode screenshots right after an action, so the action must not return until the step has rendered.
- **Audio round trip.** PIDL supplies the text (narration sets and video configurations); videoascode returns clips and a manifest; PIDL's viewer plays them. Neither tool needs the other's internals: the contract is the narration set and the audio manifest (`INIT-VAC-001`).
- **No browser speech synthesis fallback** by default: voices differ across browsers and would not match the videos.
- **Briefings are content in PIDL** (`INIT-PIDL-003`), so slides, viewer and video share one source.
