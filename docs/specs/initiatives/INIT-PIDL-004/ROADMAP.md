# ROADMAP — Complete AAuth coverage

**Initiative:** `INIT-PIDL-004`
**Repository:** `github.com/grokify/pidl`

Every area of AAuth draft-11 and its companion drafts in PIDL, with draft-sourced examples, narration and Go samples. Commits implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — Foundation

**Theme:** Reproducible examples and a structure for the whole protocol

- [ ] `RMI-PIDL-054` Move the AAuth example generator into the repository (`tools/aauthgen`) with a regeneration test
- [ ] `RMI-PIDL-055` Split the AAuth example by section and declare its groups
  - Depends on: `RMI-PIDL-054`
- [ ] `RMI-PIDL-056` Narration, pronunciations and flow IDs for the existing scenarios
  - Depends on: `RMI-PIDL-055`

## Phase 2 — Signing and Bootstrap

**Theme:** How agents sign and get their identity

- [ ] `RMI-PIDL-057` Signing scenarios: hwk, jwt, jkt-jwt, jwks_uri, and comparison
  - Depends on: `RMI-PIDL-055`
- [ ] `RMI-PIDL-058` Bootstrap variants: web, iOS, Android, self-hosted, one operator with many agents
  - Depends on: `RMI-PIDL-055`
- [ ] `RMI-PIDL-059` Bootstrap renewal (web and mobile) and sub-agent issuance
  - Depends on: `RMI-PIDL-058`

## Phase 3 — Interaction, Delegation and Lifecycle

**Theme:** What happens between the happy-path steps

- [ ] `RMI-PIDL-060` Deferred delivery and polling, step-up, and with-approval variants of the PS and federated modes
  - Depends on: `RMI-PIDL-056`
- [ ] `RMI-PIDL-061` Clarification chat and interaction chaining
  - Depends on: `RMI-PIDL-060`
- [ ] `RMI-PIDL-062` Call chaining and sub-agents (parent-mediated authorization)
  - Depends on: `RMI-PIDL-056`
- [ ] `RMI-PIDL-063` Revocation
  - Depends on: `RMI-PIDL-056`

## Phase 4 — Missions

**Theme:** Governance for long-running agents

- [ ] `RMI-PIDL-064` Mission proposal and approval, update and completion, and end-to-end
  - Depends on: `RMI-PIDL-056`
- [ ] `RMI-PIDL-065` Mission context at a resource, permission and audit endpoints, out-of-bounds access, and with and without missions
  - Depends on: `RMI-PIDL-064`

## Phase 5 — Rich Resource Requests, Samples and Release

**Theme:** The experimental edge and code developers can copy

- [ ] `RMI-PIDL-066` Rich Resource Requests: vocabulary-based access, per-call operations, content addressing and consent
  - Depends on: `RMI-PIDL-056`
- [ ] `RMI-PIDL-067` Go samples from aauth-go for agent, resource, PS and AS roles
  - Depends on: `RMI-PIDL-056`
- [ ] `RMI-PIDL-068` Coverage table, documentation and release notes
  - Depends on: `RMI-PIDL-057`, `RMI-PIDL-059`, `RMI-PIDL-061`, `RMI-PIDL-062`, `RMI-PIDL-063`, `RMI-PIDL-065`, `RMI-PIDL-066`, `RMI-PIDL-067`
