# PRD — Protocol walkthrough viewer and AAuth examples

**Initiative:** `INIT-PIDL-002`
**Repository:** `github.com/grokify/pidl`

## Problem

PIDL v0.7.0 can describe a protocol's roles and scenarios and render each scenario as a static sequence diagram. Static diagrams are a poor way to learn a protocol like AAuth:

1. **No step-by-step view.** A five-party diagram with a dozen messages shows everything at once. A presenter cannot walk an audience through it one message at a time, and a reader cannot tell which roles take part in a given step.
2. **No message content.** A diagram shows that the agent sends a resource token to the PS, but not what that request looks like: the signature headers, the body, or the claims inside the token. Readers have to cross-reference the specification.
3. **No connection between steps.** A protocol's security depends on values that link steps: a resource token's `agent_jkt` must match the key that signs the request; `presented_jti` must match the token the agent presented. Static diagrams cannot show those links.

## Goals

- A generic viewer for any PIDL file: pick a scenario, step forward and backward through it, and see the active step highlighted on the sequence diagram (its two roles, its arrow and its label).
- Example requests, responses and decoded tokens for each step, shown beside the diagram.
- A 16:9 presentation layout that works full screen in a talk and in a browser tab.
- One self-contained HTML file per protocol, with no server, so a walkthrough can be shared by sending a file.
- A fully worked AAuth draft-11 example: every step of the five access modes and their fragments with requests, responses and tokens taken from or derived from the draft, using consistent values across steps.

## Non-Goals

- Message schemas or protocol conformance checking. Examples are illustrative.
- Editing PIDL in the viewer.
- Live protocol execution against real servers.
- Replacing the static diagram formats; the viewer is an additional output.

## Users

- **Presenters** explaining a protocol in a talk or meeting.
- **Readers** learning a protocol from its PIDL description.
- **Spec reviewers** checking that the values each step relies on line up end to end.

## Requirements

### Functional

1. **Flow examples.** A flow MAY carry an `example` with a `source` (citation), a `kind` (`draft`: taken from the source with values aligned; `derived`: written from the source's normative text), an HTTP `request` and/or `response` (method or status, URL, ordered headers, body), and decoded `tokens` (header and payload) keyed to token definitions.
2. **Validation.** `kind` is `draft` or `derived`; `source` is required; header names are non-empty; token references resolve to token definitions.
3. **Walkthrough model.** The library produces a walkthrough document: protocol, roles in entities order, and for each scenario in output order its fully expanded steps. Each step carries its position, from and to roles, label, response flag, the fragments it sits inside, note, example, and analyzer findings located at that flow.
4. **Walkthrough output.** `generate -f walkthrough-json` writes the walkthrough model.
5. **Viewer layout.** A top bar with the protocol name and a scenario selector; a sequence diagram on the left; the request on the upper right; the response on the lower right; step controls.
6. **Stepping.** Next and previous buttons, arrow keys, and Home/End. The active step's arrow, label and two role headers are highlighted; earlier steps are dimmed; later steps are faint. The URL hash records the scenario and step.
7. **Fragments.** Steps inside an included fragment show the fragment's name; the diagram groups them visually.
8. **Tokens.** Decoded token header and payload are shown in readable form; claims a step's receiver checks are highlighted.
9. **Findings and notes.** A step's notes and analyzer findings appear as callouts.
10. **Example provenance.** Each example shows its source and a `draft` or `derived` badge.
11. **Single-file output.** `generate -f viewer` writes one HTML file containing the viewer and the walkthrough data.
12. **AAuth examples.** The `aauth_protocol` example gains examples for every HTTP step of all five scenarios and three fragments, with one consistent set of identifiers, keys and token values.

### Non-Functional

- The viewer renders the AAuth walkthrough in under a second on a laptop and works offline.
- The viewer's built HTML is reproducible from source with one documented command.
- Go code remains library-first; the viewer contains presentation logic only.
- Existing output formats are unchanged.

## Success Criteria

- `pidl generate -f viewer aauth_protocol -o aauth.html` produces a file that opens in a browser and steps through all five AAuth scenarios.
- In the federated scenario, stepping to the PS-to-AS request shows the request signed with the PS's key and the resource token whose `agent_jkt` names the agent's key, with the SEC019 finding as a callout.
- Every AAuth example cites a draft-11 section.

## Decisions

- **Examples live in the PIDL file** as an optional per-flow field, so they travel with the protocol.
- **Provenance is explicit.** `draft` examples come from the draft with values aligned to the scenario; `derived` examples are written from normative text. The viewer shows which.
- **Go computes, the viewer renders.** Include expansion, ordering and findings come from the Go library through the walkthrough model, so the viewer does not reimplement PIDL semantics.
- **Self-contained HTML.** The viewer is built into a single file and embedded in the Go binary.
- **Ships as v0.8.0.**
