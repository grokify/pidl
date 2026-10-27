# PLAN — PIDL v1 format and the system view

**Initiative:** `INIT-PIDL-008`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

The walkthrough format and embeddable viewer (`INIT-PIDL-007`) released in v0.9.0.

## Build Order

1. **Model first, on a long-lived branch.** The new types, process details on steps and camelCase land together, because each touches every renderer. Ports of the renderers and analyzers follow before anything merges to `main`, so `main` always builds and every example renders.
2. **Migration next,** written against the v0 examples as they stand, which become its golden inputs. Migrating the examples is the migrator's first real test.
3. **Views,** once the model is settled: the sequence and network presentation moved out of the data, then the system view and its two renderers.
4. **Walkthrough v2 and the viewer,** reading the system view from the walkthrough. The viewer keeps reading v1.
5. **Documentation and the release.**

Phases 1 and 2 are one release unit; phases 3 to 5 can follow in the same release or the next one.

## Release

pidl v0.12.0 and `@grokify/pidl-viewer` 0.12.0 together. Each consuming project migrates in its own pull request after the release:

1. Run `pidl migrate -w` on its documents and review the notes.
2. Regenerate its diagrams and walkthroughs.
3. Upgrade the pinned viewer version and its integrity hash.

`agent-protocols` goes first, and adds the AAuth system view with the Supervision Server as a proposed participant. Go modules that import pidl upgrade when they next take a pidl release; the deprecated aliases keep most code compiling through the v0.10 series.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Moving process details from participants to steps changes their meaning for documents with merges | The migrator notes each participant that receives several steps; the five process examples are reviewed by hand |
| Every renderer changes at once | A long-lived branch; golden tests for every renderer and example before and after |
| Published walkthroughs break with a new viewer | The viewer reads v1 and v2 |
| Pages pinned to viewer 0.9.0 | Pinning is the point; they upgrade when they migrate |
| Grid layouts with crossing links | Authors arrange the grid; free-form coordinates are a later option, not a v1 field |
| A design system becoming a dependency | The styling contract is CSS custom properties; PIDL imports no design system |

## Follow-On

- **AI workflow explorers (`INIT-PIDL-009`):** an inspector for steps that are not HTTP messages (inputs, outputs, data, cost and latency), and a RAG example, then agentic RAG.
- **Hosted app (`INIT-PIDL-010`):** upload or link a walkthrough or PIDL file, render it server-side, share a link.
- **Architecture and training profiles:** a design document first.
