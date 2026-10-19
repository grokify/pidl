# PLAN — Complete AAuth coverage

**Initiative:** `INIT-PIDL-004`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

`INIT-PIDL-003` phases 1–3 delivered (groups, variants, IDs, narration, samples, example checks).

## Build Order

1. **Foundation**: example generator in the repository, file split by section, groups, narration and pronunciations for existing scenarios.
2. **Signing and bootstrap**.
3. **Interaction, delegation and lifecycle**: deferred delivery, step-up, clarification, interaction chaining, call chaining, sub-agents, revocation.
4. **Missions**.
5. **R3 and samples**: Rich Resource Requests, Go samples, coverage table, release.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Large content volume | One section per phase; each phase ships on its own |
| Draft revisions mid-way | Generator holds wire details; regenerate on a new draft |

## Follow-On

Samples in other languages as SDKs are chosen; narrated videos per section (`INIT-PIDL-005`).
