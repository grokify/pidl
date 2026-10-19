# TRD — Narrated protocol videos

**Initiative:** `INIT-PIDL-005`
**Repository:** `github.com/grokify/pidl`

## How videoascode Runs a Browser Segment

From its recorder (`pkg/browser/recorder.go`): it loads the segment URL, then for each step runs the action, takes one screenshot, and holds that frame for the step's `MinDuration`, which it sets from the voiceover's audio length plus 500 ms (`pkg/segment/browser_segment.go`). The `evaluate` action calls `page.Eval`, which awaits a returned promise. Slide segments render Marp decks with speaker notes as voiceover. Mixed configurations combine slide and browser segments.

## Design

### 1. Viewer video mode and navigation hook

```ts
// Present only with ?mode=video
window.pidl = {
  goTo(scenarioId: string, step: number): Promise<void>,
};
```

`goTo` sets the position, waits for React to commit, scrolls the active row into view with `behavior: "instant"`, then waits two animation frames before resolving. Video mode also sets `cursor: none`, disables transitions and hover styles, and forces the presentation stage.

### 2. Configuration per scenario (`render/vac.go`)

```yaml
version: "1.0"
title: "AAuth: Federated authorization (four-party)"
defaultLanguage: en-US
defaultVoice: { provider: elevenlabs, voiceId: "<--voice-id>" }
resolution: { width: 1920, height: 1080 }
frameRate: 30
pronunciations:
  AAuth: { en-US: "ay auth" }
segments:
  - type: slide
    name: federated-authorization-briefing
    source: decks/05-federated-authorization.md
  - type: browser
    name: federated-authorization
    url: "file:///…/viewer.html?mode=video#federated-authorization/0"
    steps:
      - action: evaluate
        script: "() => window.pidl.goTo('federated-authorization', 0)"
        voiceover: "<scenario narration>"
      - action: evaluate
        script: "() => window.pidl.goTo('federated-authorization', 1)"
        voiceover: "<step 1 narration>"
      # a step with a tour adds one evaluate per beat:
      - action: evaluate
        script: "() => window.pidl.focus('federated-authorization', 12, 'token.resource_token.claims.agent_jkt')"
        voiceover: "<beat narration>"
      # … one step per walkthrough step, then its beats
      - action: wait
        duration: 1500
```

Beats come from `INIT-PIDL-003` tours; region strings are validated against the step and escaped for the script string. Scenario IDs are validated to `^[a-z][a-z0-9_-]*$`, so they are safe to embed in the script string. Narration falls back to the step description; a step with neither gets `minDuration` 2500.

### 3. Briefing deck (`render/marp.go`)

A Marp deck per scenario with `marp: true`, a 16:9 size, and one slide per briefing part; each slide's spoken text is an HTML comment, the voiceover format videoascode reads. The network slide embeds `images/<NN>-<scenario-id>-network.svg`, rendered with the SVG network renderer from the scenario's projection.

### 4. CLI and output

`pidl generate -f vac [--scenario id | --group id | --all] [--combined] [--viewer-url url] [--voice-provider p --voice-id v --model m] -o <dir> <file>` writes:

```
<dir>/
  viewer.html
  decks/<NN>-<scenario-id>.md
  images/<NN>-<scenario-id>-network.svg
  <NN>-<scenario-id>.vac.yaml
  all.vac.yaml            # with --combined
```

Files are written only when their content changes, so unchanged scenarios keep their videoascode audio caches warm.

## Sequencing

Requires `INIT-PIDL-003` narration, briefings and pronunciations. Video mode and the navigation hook come first, then configurations, then decks.

## Risks

- **Configuration drift** between PIDL's output and videoascode. Mitigation: a test that parses generated YAML with videoascode's configuration types, pinned by version.
- **Marp rendering differences** for SVG images. Mitigation: render network diagrams to PNG as well and prefer PNG in decks if SVG fails.
- **File URLs** may be restricted in some setups. Mitigation: `--viewer-url`.

## Testing Strategy

Golden configuration and deck for the AAuth fixture; viewer unit tests for the navigation hook (resolves after render) and video-mode flags; a test that unchanged scenarios are not rewritten; a documented end-to-end render of one AAuth scenario.
