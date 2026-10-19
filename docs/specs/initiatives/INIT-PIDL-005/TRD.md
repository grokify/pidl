# TRD — Narrated protocol videos

**Initiative:** `INIT-PIDL-005`
**Repository:** `github.com/grokify/pidl`

## Design

### 1. Generated configuration (`render/vac.go`)

```yaml
version: "1.0"
title: "AAuth"
description: "<protocol description>"
defaultLanguage: en-US
defaultVoice:
  provider: elevenlabs
  voiceId: "<--voice-id>"
resolution: { width: 1920, height: 1080 }
frameRate: 30
pronunciations:
  AAuth: { en-US: "ay auth" }
segments:
  - type: browser
    name: federated-authorization
    url: "file:///…/aauth.html?mode=video#federated-authorization/0"
    steps:
      - action: wait
        duration: 1500
        voiceover: "<scenario narration>"
      - action: keypress
        key: ArrowRight
        voiceover: "<step 1 narration>"
      # … one keypress per step
      - action: wait
        duration: 1500
```

The renderer reads the walkthrough model, so groups, variants and fragment expansion are already resolved. Field names follow videoascode's browser configuration; a test parses the output with videoascode's configuration types to catch drift.

### 2. CLI

`pidl generate -f vac [--scenario id | --group id] [--viewer-url url] [--voice-provider p --voice-id v --model m] -o video.yaml <file>`. Without `--viewer-url`, the command also writes the viewer HTML next to the YAML and uses its file URL.

### 3. Viewer video mode

`?mode=video`: forces the stage on, sets `cursor: none`, disables smooth scrolling and hover styles, and hides the theme and Present controls. It keeps keyboard stepping so videoascode can drive it.

## Sequencing

Requires `INIT-PIDL-003` narration and pronunciations, and the viewer from v0.8.0.

## Risks

- **Configuration drift** between PIDL's output and videoascode. Mitigation: a compatibility test against videoascode's types, pinned by version.
- **One frame per step** in videoascode's browser capture. The viewer changes discretely per step, which suits it; video mode removes animations.
- **File URLs** may be restricted in some browser automation setups. Mitigation: `--viewer-url` for an HTTP-served viewer.

## Testing Strategy

Golden YAML for the AAuth example; viewer unit test for video-mode flags; a manual end-to-end render of one scenario documented in the guide.
