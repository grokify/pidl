# PRD — Narrated protocol videos

**Initiative:** `INIT-PIDL-005`
**Repository:** `github.com/grokify/pidl`

## Problem

A PIDL walkthrough explains a protocol step by step in a browser. Developer relations also needs videos (MP4 for YouTube and course platforms) where a narrator walks through each scenario while the diagram advances. Producing them by hand means recording the screen, writing and recording narration, and re-editing whenever the protocol or narration changes.

videoascode (`vac`) already turns a configuration of browser steps with voiceover text into a narrated video: text-to-speech through OmniVoice (ElevenLabs, Deepgram, local F5-TTS), timing each step to its audio, recording, subtitles, audio caching and multiple languages. What is missing is the bridge: a way to turn a PIDL walkthrough into that configuration, and a viewer mode that records cleanly.

## Goals

- Generate a videoascode configuration from a PIDL document: one browser segment per scenario that opens the walkthrough viewer and advances step by step, with each step's narration as its voiceover.
- Pass PIDL's pronunciations through to videoascode unchanged.
- A viewer video mode that looks right on camera: the presentation stage, no pointer, no smooth scrolling, no hover effects.
- One command line from a PIDL file to MP4 and subtitles, with videoascode doing the media work.

## Non-Goals

- Text-to-speech, recording, encoding or subtitles in PIDL. videoascode owns all media work.
- Uploading to YouTube.
- Writing narration. Narration is content (`INIT-PIDL-004` for AAuth).

## Requirements

1. **`generate -f vac`** writes a videoascode browser configuration (YAML): metadata from the protocol, the protocol language as the default language, pronunciations, the 1920×1080 resolution, and one segment per selected scenario.
2. **Segments** open the viewer at `<viewer URL>?mode=video#<scenario>/0`. The first step waits on the scenario introduction with the scenario's narration; each following step presses → with that step's narration; a final step holds briefly. Steps without narration use their description; a step with neither holds for a minimum duration.
3. **Selection**: `--scenario`, `--group`, or all main scenarios; fragments only when asked. Segments follow navigation order.
4. **Viewer URL**: a file URL to a generated viewer HTML (default) or an HTTP URL supplied by the user.
5. **Video mode** in the viewer (`?mode=video`): presentation stage on, pointer hidden, instant scrolling, hover states off, the narrative and controls kept visible, and the menu bar reduced to the protocol and scenario name.
6. **Voice settings** (`--voice-provider`, `--voice-id`, `--model`) are written to the configuration's default voice; credentials stay in the environment where videoascode reads them.
7. **Documentation** shows the end-to-end commands: `pidl generate -f viewer`, `pidl generate -f vac`, `vac browser video`.

## Success Criteria

- `vac browser video` renders an AAuth scenario from the generated configuration into an MP4 whose diagram advances with the narration.
- The same configuration with `--lang` produces a second language when the PIDL narration is translated.

## Decisions

- **videoascode is the media engine**; PIDL generates its input. This keeps PIDL free of audio and video dependencies and lets both tools evolve independently.
- **Keyboard stepping, not clicks**: the viewer's → key is stable across layout changes, unlike CSS selectors.
