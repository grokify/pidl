# Walkthrough Viewer

The walkthrough viewer turns a PIDL file into a page you can step through, one message at a time. It is built for learning a protocol and for presenting it: pick a scenario, press → and watch the active message, its two roles, and its request and response.

```bash
pidl generate -f viewer -o aauth.html aauth_protocol
open aauth.html
```

The output is one self-contained HTML file. It needs no server and no network, so it can be emailed, attached to an issue or opened straight from disk.

## The Page

| Area | Shows |
|------|-------|
| Menu bar | The protocol name, a scenario selector (main scenarios, then shared steps), Overview, ← and →, and the step count |
| Sequence card (left) | The scenario's roles and steps. Shared steps (fragments) appear as labeled bands. The active step's arrow and label are highlighted, with its sender outlined in blue and its receiver in green; earlier steps are dimmed and later ones faint. |
| Caption (under the diagram) | The active step, the fragments it sits inside, its description and note, and any analyzer findings |
| Request card (upper right) | The active request, or the request a response answers |
| Response card (lower right) | The response, once you step to it |

## Keys

| Key | Action |
|-----|--------|
| → ↓ Space Page Down | Next step |
| ← ↑ Page Up | Previous step |
| Home | Overview (all steps, none active) |
| End | Last step |
| P | Present on a 16:9 stage |

The URL records the scenario and step, for example `#federated-authorization/12`, so a link opens at that step.

## Requests, Responses and Tokens

When a flow has an [example](../SPECIFICATION.md#flow-examples), the cards show it:

- **HTTP**: the request line, `Host`, headers in written order and body, or the status line, headers and body. Headers covered by the HTTP message signature (from `Signature-Input`) are marked *signed*. Long values such as JWTs shorten; click to expand.
- **Tokens**: each decoded token's header and claims. Claims that the receiving role checks, according to the token's `claims` and `checked_by`, are highlighted.
- **Provenance**: a `draft` badge for examples taken from the cited source with values aligned, or `derived` for examples written from its normative text, with the source section and the example's note.

A request's response appears only when you reach it, so the walkthrough reveals each answer in order. A pending response pairs with the poll it answers.

## Presenting

**Present** (or P) places the page on a fixed 1600×900 stage scaled to the window, letterboxed, so it looks the same on any projector or shared screen. **Theme** cycles auto, light and dark.

## Data

The viewer reads the walkthrough produced by `pidl generate -f walkthrough-json`: every scenario with fragments expanded, each step's example, and analyzer findings attached to the steps they concern (SEC016, flow provenance, is left out). Use that format to build other viewers.

## Developing the Viewer

The viewer lives in `viewer/` (Vite, React and TypeScript):

```bash
cd viewer
npm ci
npm run data       # write public/walkthrough.json from the AAuth example
npm run dev        # http://localhost:5173
npm test
npm run build:embed  # rebuild render/viewer/viewer.html for the Go binary
```

`go generate ./render` runs the same build. CI fails when the committed `render/viewer/viewer.html` does not match a fresh build.
