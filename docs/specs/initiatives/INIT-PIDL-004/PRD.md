# PRD — Complete AAuth coverage

**Initiative:** `INIT-PIDL-004`
**Repository:** `github.com/grokify/pidl`

## Problem

The AAuth example covers the five resource access modes and three shared steps. AAuth is larger: how agents sign, how they bootstrap on each platform, missions and their governance endpoints, call and interaction chaining, sub-agents, clarification, deferred delivery, step-up, revocation, and the experimental Rich Resource Requests. An existing AAuth explorer shows these areas, but its examples predate draft-11 (non-matching digests, forbidden algorithms) and its code is AAuth-specific. PIDL is to be the source of truth for AAuth walkthroughs, so it needs the whole protocol.

## Goals

- Every area of AAuth draft-11 and its companion drafts, as PIDL scenarios in ordered sections, with variants where a scenario has distinct forms.
- Each scenario has an introduction and spoken narration; each step has a narrative, narration and, where it is an HTTP exchange, an example.
- All examples come from the drafts (`draft`) or their normative text (`derived`), share one set of keys, identifiers and timeline, and pass the generic checks and the `aauth-draft-11` profile.
- Agent, resource and server roles show Go code from `aauth-go`'s tested examples.

## Non-Goals

- Copying content from other AAuth tools. The explorer is used only as a checklist of topics.
- Languages other than Go until their SDKs are chosen.
- Video production (`INIT-PIDL-005`).

## Scope

| Section | Scenarios | Source |
|---------|-----------|--------|
| Signing | Pseudonymous signing (`hwk`); agent tokens (`jwt`); hardware-backed delegation (`jkt-jwt`); server signing (`jwks_uri`); comparison | draft-hardt-httpbis-signature-key-09; draft-11 §11.3 |
| Agent bootstrap | Web app; iOS; Android; self-hosted; many agents under one operator; web and mobile renewal; sub-agent issuance | draft-hardt-aauth-bootstrap-02 |
| Resource access | Agent identity; resource-managed; person identity; PS authorization; federated authorization; comparison | draft-11 §4.2, §6, §7, §9 |
| Interaction and delivery | Deferred delivery (202) and polling; step-up; clarification chat; interaction chaining | draft-11 §6.5.1, §7.5, §10.1.2, §11.8 |
| Missions | Proposal and approval; mission context at a resource; update and completion; permission endpoint; audit endpoint; out-of-bounds access; end-to-end; with and without missions | draft-11 §7.7, §7.8, §8 |
| Delegation | Call chaining; sub-agents (parent-mediated authorization) | draft-11 §10 |
| Lifecycle | Revocation | draft-11 §11.12 |
| Rich Resource Requests (experimental) | Vocabulary-based access; per-call operations; content addressing and consent | draft-hardt-aauth-r3 |

Variants: autonomous and with-user-approval for the PS and federated modes; platform variants under bootstrap.

## Requirements

1. Scenarios are grouped as in the scope table, in that order.
2. Shared steps (agent token, person token, consent, deferred polling) are fragments reused across scenarios.
3. Every HTTP step has an example; the example checks and `aauth-draft-11` profile report nothing.
4. Every scenario and step has narration, and the protocol declares pronunciations for its identifiers.
5. Agent, resource, PS and AS roles show Go samples from `aauth-go` for the scenarios it covers, referenced by region.
6. Findings the analyzer reports against the drafts (such as SEC019 in federated access) remain visible as callouts.
7. A coverage table in the documentation lists each scenario, its source sections and its status.

## Success Criteria

- The viewer navigates every section in the scope table.
- `pidl analyze --profile aauth-draft-11` reports no example findings.
- Each Go sample resolves from `aauth-go` and its region hash is recorded.

## Decisions

- **Content is written from the drafts.** Other tools inform the topic list only.
- **One example world.** `agent.example`, `ps.example`, `resource.example`, `as.resource.example`, the RFC 8037 agent key and one timeline across every scenario, so values link up across the whole protocol.
- **Experimental drafts are labeled** in the section name and scenario introductions.
