# TRD — Complete AAuth coverage

**Initiative:** `INIT-PIDL-004`
**Repository:** `github.com/grokify/pidl`

## Design

### 1. File layout

`examples/aauth_protocol.json` grows large. It is split into a base file (protocol, entities, tokens, pronunciations, groups, shared fragments) and one file per section that `imports` the base's scenarios as needed, composed by `extends`/`imports` into the built-in example. Composition already preserves scenario order and fragments.

### 2. Example generation

The generator used for the current examples moves into the repository as `tools/aauthgen` (Go), so examples are reproducible:

- One table of shared values (hosts, keys, identifiers, timeline).
- Builders for each message type (signed request, 401/202 challenges, token responses, mission documents) that compute real `Content-Digest` values and encode tokens from their decoded form.
- Signatures remain placeholders; tokens appear shortened (`header.eyJ…`) wherever a full compact form would trip secret scanners.
- `go generate ./examples` regenerates examples; a test fails when the committed file differs.

### 3. Sources

Each scenario cites draft sections in its introduction; each example cites its source in `source`. Mission documents follow draft-11 §8.2 blob members; R3 documents follow the R3 draft's vocabulary format.

### 4. Go samples

Samples reference `aauth-go` regions (`pidl:begin`/`pidl:end` comments added to its example tests in a PR to that repository): agent transport, resource handler, PS and AS configuration, mission proposal, permission requests. `samples resolve` records hashes; CI runs `samples check` against a pinned `aauth-go` version.

## Sequencing

Requires `INIT-PIDL-003` phases 1–3 (groups, variants, IDs, narration, samples, checks). The generator move comes first so every new scenario is produced the same way.

## Risks

- **Draft churn.** A new draft revision changes wire details. Mitigation: values live in the generator; the profile is versioned by draft.
- **R3 maturity.** The R3 draft is experimental. Mitigation: label it and keep it in its own section and file.
- **SDK coordination.** Region markers need a change in `aauth-go`. Mitigation: inline samples as an interim.

## Testing Strategy

Examples regenerate identically; the analyzer and `aauth-draft-11` profile report no example findings; every sample resolves; the viewer renders each section (smoke test over the walkthrough JSON).
