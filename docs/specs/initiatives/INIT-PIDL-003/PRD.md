# PRD — PIDL for developer content

**Initiative:** `INIT-PIDL-003`
**Repository:** `github.com/grokify/pidl`

## Problem

PIDL v0.8.0 describes a protocol's scenarios and shows one illustrative HTTP exchange per step in a walkthrough viewer. Developer relations needs more from the same source:

1. **Navigation for many scenarios.** A complete protocol has dozens of scenarios in sections (AAuth: signing, bootstrap, access modes, missions, advanced patterns, R3). A dropdown does not scale, and PIDL has no notion of sections.
2. **Variants.** The same scenario often has an autonomous and a with-approval form, or a web and a mobile form. Today each is an unrelated scenario.
3. **Stable step identity.** Samples, deep links and narration need to target a step. Flow `action` values repeat (AAuth federated access has two `auth_token` steps), and array positions shift when steps are added.
4. **Spoken narration.** On-screen narratives contain identifiers such as `agent_jkt` and `aa-auth+jwt` that read badly aloud. Video needs a script written for the ear, plus pronunciations.
5. **More than HTTP.** Developers want the same message as cURL and as SDK code (Go, TypeScript, Python, C# and others). SDK code pasted into documents goes stale as SDKs change.
6. **Trustworthy examples.** Examples are hand-assembled. Nothing checks that a `Content-Digest` matches its body, that an encoded token matches its decoded form, or that examples follow a protocol's rules (AAuth draft-11 forbids `EdDSA`, for instance). An existing AAuth explorer shows how easily these drift: its digests do not match its bodies and its tokens use algorithms draft-11 forbids.

## Goals

- PIDL stays the single, protocol-neutral source of truth for protocol walkthroughs; nothing in this initiative is AAuth-specific except an optional check profile.
- Scenarios can be grouped into ordered sections and related as variants.
- Steps have stable identifiers.
- Each scenario and step can carry a spoken narration script, and a protocol can declare pronunciations.
- A message can be shown as HTTP, as generated cURL, and as SDK code drawn from tested source files.
- Examples are checked: generic HTTP and JWT consistency for every protocol, and protocol-specific rules through profiles.
- The viewer gains grouped left navigation, variant toggles, autoplay, a token lifecycle panel, and representation tabs.
- The viewer reaches feature parity with dedicated protocol explorers: header and claim help, signature bases, deferred-response timelines, protocol documents, value tracing, comparison tables, and reference pages, all defined generically in PIDL.
- Steps can carry guided tours: ordered beats that focus a region of the page with narration, for live autoplay and for narrated video.

## Non-Goals

- Text-to-speech and video production (`INIT-PIDL-005`).
- AAuth content beyond what tests need (`INIT-PIDL-004`).
- Generating SDK code from PIDL. Samples come from real SDKs.
- Executing samples or examples against live servers.

## Users

- **Developer relations** writing tutorials, presentations and videos from one source.
- **Application developers** learning to call a protocol from their language.
- **Spec authors and reviewers** who need examples they can trust.
- **SDK maintainers** whose tested examples appear in walkthroughs.

## Requirements

### Functional

1. **Groups.** A document MAY declare an ordered `groups` array (`id`, `name`, `description`). A scenario MAY name its `group`. Group order, then scenario order within a group, sets navigation order.
2. **Variants.** A scenario MAY declare `variant_of` (another scenario's ID) and `variant_label` (e.g. "With user approval"). The base scenario and its variants share a navigation entry with a toggle.
3. **Flow IDs.** A flow MAY have an `id`, unique within its scenario and matching `^[a-z][a-z0-9_-]*$`. The viewer accepts `#<scenario>/<flow-id>` as well as `#<scenario>/<n>`.
4. **Narration.** Scenarios and flows MAY carry `narration`: plain text written to be spoken, in the document's `protocol.language` (default `en-US`). A document MAY declare `pronunciations`, a map from a term to its spoken form per language (`{"AAuth": {"en-US": "ay auth"}}`), the same shape videoascode reads, so video tooling can use it unchanged. When narration is absent, tools fall back to `description`.
5. **Briefing.** A scenario MAY carry a `briefing` for the slides that introduce it: `why` (prose on why the scenario exists and what problem it solves), `points` (a short ordered list of key points), and `narration` (what is spoken over those slides). The scenario `description` stays a one-paragraph summary.
6. **Samples.** A flow example and a scenario MAY carry `samples`: each has a `language`, an optional `label` and `role`, and either inline `code` or a `source` reference (`repo`, `path`, `region`). Region markers in source files (`pidl:begin <name>` and `pidl:end <name>` in a comment) delimit the code.
7. **Sample resolution.** `pidl samples resolve` fetches referenced regions from local checkouts and records each sample's content hash; generation embeds resolved code so viewers stay self-contained; a check fails when a referenced region is missing.
8. **Generated cURL.** For any example request, tools generate an equivalent cURL command. The viewer notes that cURL replays the example's signature headers and cannot create HTTP message signatures.
9. **Example checks** (generic): `Content-Digest` matches the body it covers; encoded JWTs in headers and bodies decode to the decoded tokens shown; every component named in `Signature-Input` is present; `created` and token `iat`/`exp` values are consistent; decoded token claims respect `exp > iat`.
10. **Check profiles.** Protocol-specific rules live in named profiles (first: `aauth-draft-11`) selected by the document (`protocol.profiles`) or on the command line.
11. **Walkthrough data** carries groups, variants, flow IDs, narration, briefings, pronunciations, resolved samples, generated cURL and token lifecycle events.
12. **Viewer:** grouped left navigation; variant toggle; play, pause and replay with speeds (0.5× to 3×) and per-step durations from narration length; a token lifecycle panel (issued, forwarded, presented, checked); a decoded/raw token toggle; representation tabs (HTTP, cURL, and each sample language) with the chosen language remembered.

13. **Glossary.** A document MAY declare a `glossary` of headers, claims and terms (`term`, `kind`, `category`, `description`, optional `reference`). The viewer shows descriptions as help on headers and claims and badges terms by category (for example, protocol-defined headers).
14. **Signature base.** For each example request with `Signature-Input`, tools compute the RFC 9421 signature base from the request; the viewer shows it with each covered component distinguished.
15. **Deferred timeline.** When steps return `202 Accepted` with a `Location`, the walkthrough derives the pending request's timeline (each poll and its status) and the viewer shows it.
16. **Artifacts.** Scenarios and steps MAY carry `artifacts`: named JSON or Markdown documents (such as a mission document), optionally with a computed digest (`sha-256`, base64url) the viewer shows.
17. **Value tracing.** The viewer lets the reader select a value (a digest, identifier or thumbprint) and highlights every step, header, body field, token claim and artifact where it appears.
18. **Comparisons.** A document MAY declare `comparisons`: tables whose columns are scenarios and whose rows are questions with a short answer per scenario. The viewer shows them as pages in the navigation.
19. **Pages.** A document MAY declare `pages`: Markdown reference documents placed in a group. The viewer also generates an overview page from the protocol description, roles and groups. Scenarios MAY carry `references` (name and URL) shown as spec links.
20. **Viewer conveniences:** copy buttons on headers, bodies and tokens; a collapsible and resizable details panel, remembered per browser.
21. **Guided tours.** A flow MAY carry a `tour`: ordered beats, each with a `focus` region and `narration` (and optional `hold` milliseconds). Regions name parts of the page generically (`diagram`, `narrative`, `request.line`, `request.headers.<name>`, `request.body[/<json-pointer>]`, the same for `response`, `token.<id>.header`, `token.<id>.claims.<claim>`, `artifact.<id>[/<json-pointer>]`, `lifecycle`, `page.<id>#<heading>`). Validation checks each region resolves against the step.
22. **Focus hook.** The viewer exposes `window.pidl.focus(scenarioId, step, region)`, which switches to the tab holding the region, scrolls its panel so the region is visible, highlights it, and resolves after rendering. Autoplay plays a step's beats in order.
23. **Tour drafting.** `pidl tour draft` proposes beats for a step from its example (headers, body fields, token claims) with narration drafted from the glossary, for an author to edit.

### Non-Functional

- All additions are optional; v0.8.0 documents validate and render unchanged.
- Check profiles are data plus small Go functions in the `analyze` package; adding one does not change the core.
- The viewer remains a single self-contained HTML file.

## Success Criteria

- The AAuth example declares groups and at least one variant, and every step has an ID and narration.
- The viewer navigates AAuth by section, autoplays a scenario, and shows HTTP, cURL and Go for steps with samples.
- The example checks report no findings on the AAuth example, and report the digest and algorithm problems when run on a deliberately broken copy.

## Decisions

- **Variants are scenarios** linked by `variant_of`, so ordering, validation and projection rules apply unchanged.
- **Samples come from tested source files by reference**, with inline code allowed for snippets that have no home. Resolved code is embedded at generation time.
- **cURL is generated, never stored.**
- **Protocol rules live in profiles**, keeping PIDL's core neutral.
- **Ships as v0.9.0.**
