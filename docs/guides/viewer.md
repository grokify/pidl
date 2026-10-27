# Walkthrough Viewer

The walkthrough viewer turns a PIDL file into a page you can step through, one message at a time. It is built for learning a protocol and for presenting it: pick a scenario, press → and watch the active message, its two roles, and its request and response.

```bash
pidl generate -f viewer -o api.html oauth2_api_access
open api.html
```

The output is one self-contained HTML file. It needs no server and no network, so it can be emailed, attached to an issue or opened straight from disk.

To show walkthroughs inside other pages, such as a docs site, load the viewer from a CDN instead: see [Embedding the Viewer](embedding.md).

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

The URL records the scenario and step, for example `#federated-authorization/12`, so a link opens at that step. A step with an [ID](scenarios.md#sections-variants-and-step-ids) can be linked by name instead, such as `#ps-authorization/consent.approve`, which keeps working when steps are added before it.

## Requests, Responses and Tokens

When a flow has an [example](../SPECIFICATION.md#flow-examples), the cards show it:

- **HTTP**: the request line, `Host`, headers in written order and body, or the status line, headers and body. Headers covered by the HTTP message signature (from `Signature-Input`) are marked *signed*. Long values such as JWTs shorten; click to expand.
- **cURL** and **code**: the request as a generated `curl` command, and the step's [code samples](samples.md), one tab per language, titled by label and role, with a copy button. The tab you choose stays selected from step to step; a step without that form shows HTTP with the tab disabled.
- **Tokens**: each decoded token's header and claims. Claims that the receiving role checks, according to the token's `claims` and `checked_by`, are highlighted.
- **Provenance**: a `draft` badge for examples taken from the cited source with values aligned, or `derived` for examples written from its normative text, with the source section and the example's note.

A request's response appears only when you reach it, so the walkthrough reveals each answer in order. A pending response pairs with the poll it answers.

## Observed Runs

A walkthrough generated with an observation (see [Observed Runs](observed-runs.md)) has an entry for each run of a scenario observed in a real system, right after the scenario, named "Observed: …". On such an entry a **Concept / Observed** switch appears in the top bar:

- **Observed** (the default) shows each step's real exchange in the cards, with an *observed* badge and the time it happened. The step caption adds the gap from the previous exchange, a *Weak match* mark when only the participants and method agreed, and why a step has no exchange: not seen, skipped, not matchable, or not observable because a person acts in a browser.
- **Concept** shows each step's example, as on any scenario.

The header summarizes the run: its instance and source, steps matched (and how many weakly), steps not seen, and exchanges that matched no step. Those extra exchanges, such as servers fetching each other's keys, appear dotted between the steps they fell between, and open in the cards like any step. Tokens are decoded but not verified, and raw token, signature and credential values are masked unless the walkthrough was generated with `--raw`.

## Presenting

**Present** (or P) places the page on a fixed 1600×900 stage scaled to the window, letterboxed, so it looks the same on any projector or shared screen. **Theme** cycles auto, light and dark.

## Data

The viewer reads a [walkthrough](../reference/walkthrough-format.md) (`pidl-walkthrough/v1`), the format `pidl generate -f walkthrough-json` produces: every scenario with fragments expanded, each step's example, and analyzer findings attached to the steps they concern (SEC016, flow provenance, is left out). Use that format to build other viewers.

## Developing the Viewer

The viewer lives in `viewer/` (Vite, React and TypeScript):

```bash
cd viewer
npm ci
npm run data       # write public/walkthrough.json from the oauth2_api_access example
npm run dev        # http://localhost:5173
npm test
npm run build:embed  # rebuild render/viewer/viewer.html for the Go binary
npm run build:lib    # build the embeddable viewer into dist-lib/
```

`go generate ./render` runs the same build. CI fails when the committed `render/viewer/viewer.html` does not match a fresh build.
