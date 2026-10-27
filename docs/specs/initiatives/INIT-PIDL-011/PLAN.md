# PLAN — Live runs in the viewer

**Initiative:** `INIT-PIDL-011`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

The walkthrough format and embeddable viewer (`INIT-PIDL-007`) released in v0.9.0. The observed fields are additive and named for both walkthrough versions, so the work starts against the current format and carries over to `pidl-walkthrough/v2` when `INIT-PIDL-008` lands. If that initiative has landed by the time the viewer work starts, the viewer builds on its version.

## Build Order

1. **A spike first.** The captured AAuth Petstore run from the sandbox, with its two journeys, its health check and its key fetches, aligned by throwaway code to the AAuth scenarios and shown in the viewer. The spike answers three questions before any format is fixed: whether the default rules plus a binding are enough, how often a step needs an explicit rule, and whether instance splitting by first step works on a real capture. It also produces the fixture the later tests use.
2. **The formats and the matcher.** The observation and binding types and schemas, then rules, scoring, alignment, repeated steps, scenario instances and the time window, with their property tests, before any adapter beyond the sandbox trace.
3. **The overlay.** `--observation` and `--binding` on `pidl generate`, the walkthrough fields in both versions, the self-contained page, token carriage and decoding, and redaction.
4. **The viewer.** The switch, the instance label, the badges, and extra exchanges in place.
5. **The other sources.** The HAR and Newman adapters, and the join of a client source to a wire source by method, URL and time. Newman comes after the trace because its requests are not the wire messages when a signing proxy is in use.
6. **Check.** `pidl observe check`, its exit codes and its report.
7. **The first consumer.** The AAuth document gains match rules for its steps whose examples have no request, and token carriage; aauth-postman keeps the sandbox binding and a target that runs the scenarios and opens the self-contained page; omniagent writes client events that say why it sent each next request.
8. **Documentation and the release.**

Phases 1 and 2 can be reviewed on their own: the matcher and the overlay have no viewer dependency.

## Release

pidl v0.11.0 and `@grokify/pidl-viewer` 0.11.0 together, or the next versions after `INIT-PIDL-008` ships. Consumers upgrade in their own changes:

1. `agent-protocols` adds the AAuth match rules and token carriage to its document and regenerates the explorer.
2. `aauth-postman` adds the binding and the demo target.
3. omniagent writes client events and points its guide at the demo.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Default rules mismatch on real traffic: deployment paths, discovered endpoints, extra headers | Path and status only add to the score; the binding maps paths per deployment; the report prints each score; the spike runs on a real capture first |
| Many steps have no example request to derive a rule from | Such steps are reported as `unmatchable`, never silently not seen, and the AAuth document gains rules for them as part of the first consumer |
| A capture holds several journeys and noise | Scenario instances split by first step, a time window, and extra exchanges kept in place; interleaved concurrent journeys are a stated non-goal |
| A polled or repeated message matches the wrong step | `repeat` on the step; ordered alignment by dynamic programming with property tests for duplicates and swaps |
| A client source is mistaken for the wire messages | `requestFidelity` in the observation, a note in the viewer, and a join to the server's trace by method, URL and time |
| Labels in a capture differ from participants | Label aliases and path-based resolution in the binding; an unknown sender still matches where the rule allows |
| Real tokens end up in a shared file | Redaction of raw tokens and signatures in every output unless `--raw`; fixtures with truncated values; tests use generated tokens |
| "Matched" is read as "correct" | Every token is labeled decoded, not verified; the guide and the viewer say a match is resemblance, not conformance |
| Two formats change at once (`INIT-PIDL-008`) | The new fields are additive with names for both versions; the matcher reads the document through the same model as the other tools |
