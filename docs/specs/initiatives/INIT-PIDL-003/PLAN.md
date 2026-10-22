# PLAN — PIDL for developer content

**Initiative:** `INIT-PIDL-003`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

v0.8.0 (walkthrough viewer) merged.

## Build Order

1. **Structure**: groups, variants, flow IDs, schema and specification.
2. **Narration and representations**: narration and pronunciations, samples with source references and resolution, generated cURL, walkthrough data.
3. **Example checks**: generic rules, profiles with `aauth-draft-11`, CLI.
4. **Viewer**: left navigation and variants, autoplay, token lifecycle, representation tabs and raw tokens; documentation and v0.9.0.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Scope creep into AAuth content | AAuth content lives in `INIT-PIDL-004` |
| Samples drift from SDKs | Hashes recorded at resolve; `samples check` in CI |

## Follow-On

`INIT-PIDL-004` (AAuth coverage) and `INIT-PIDL-005` (narrated video) build on this.
