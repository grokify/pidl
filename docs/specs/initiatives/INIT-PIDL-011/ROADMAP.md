# ROADMAP — Live runs in the viewer

**Initiative:** `INIT-PIDL-011`
**Repository:** `github.com/grokify/pidl`

A walkthrough keeps its explanation and gains the exchanges that a real run produced, so a developer sees the concept and their own system's messages side by side. Observations come from a trace, HAR or a Newman report, are split into the scenario instances they contain, are matched to steps by rules the document and a binding carry, and are shown in the same viewer.

## Phase 1 — Observations and Matching

**Theme:** A neutral format and a matcher that holds no knowledge of any protocol and copes with real captures

- [x] `RMI-PIDL-131` `pidl-observation/v1` and `pidl-binding/v1`: types, reader, writer, validation and generated schemas with the drift test
- [x] `RMI-PIDL-132` Role-labeled trace adapter, with the AAuth Petstore capture as the first fixture
  - Depends on: `RMI-PIDL-131`
- [x] `RMI-PIDL-133` Rules: request, response, standalone and unobservable steps; example defaults and step `match`; binding labels, hosts, paths, rewrites and overrides; header and value matching; the literal-identifier warning
  - Depends on: `RMI-PIDL-131`
- [x] `RMI-PIDL-151` Repeated steps (`repeat`) and body predicates by JSON pointer
  - Depends on: `RMI-PIDL-133`
- [x] `RMI-PIDL-134` Scoring, strong and weak matches, ordered alignment within an instance by dynamic programming with response steps paired, step and exchange statuses, and the property tests
  - Depends on: `RMI-PIDL-151`
- [x] `RMI-PIDL-150` Scenario instances: splitting a capture where a first step recurs, a named scenario list or the best-ranked non-fragment candidate (exchanges explained minus required steps not seen), the time window, and the instance report with the runner-up
  - Depends on: `RMI-PIDL-134`

## Phase 2 — The Overlay and the Viewer

**Theme:** The same viewer shows the concept or the run

- [ ] `RMI-PIDL-135` `--observation` and `--binding` on `pidl generate -f walkthrough-json` and `-f viewer`, with `pidl observe import`, and golden tests on the trace fixture
  - Depends on: `RMI-PIDL-132`, `RMI-PIDL-150`
- [ ] `RMI-PIDL-136` Walkthrough fields for observed exchanges, statuses, run summaries and extra exchanges, named for v1 and v2, with schemas and TypeScript types
  - Depends on: `RMI-PIDL-135`
- [ ] `RMI-PIDL-137` Token carriage in token definitions and decoding of observed tokens, labeled decoded, not verified, tested with generated tokens
  - Depends on: `RMI-PIDL-136`
- [ ] `RMI-PIDL-138` Redaction of raw tokens, signatures and credentials in every output, and `--raw`
  - Depends on: `RMI-PIDL-137`
- [ ] `RMI-PIDL-139` Viewer: Concept / Observed switch, the instance label, status badges with weak matches marked, repeat counts and observed times
  - Depends on: `RMI-PIDL-136`
- [ ] `RMI-PIDL-140` Viewer: extra exchanges between the steps they fell between, and steps that were not seen or not matchable
  - Depends on: `RMI-PIDL-139`

## Phase 3 — More Sources and Checking

**Theme:** HAR and Newman in, and a command that fails when a step is missing

- [ ] `RMI-PIDL-141` HAR 1.2 adapter with `_from` and `_to` labels, base64 content and request fidelity, with its fixture
  - Depends on: `RMI-PIDL-135`
- [ ] `RMI-PIDL-142` Newman report adapter: executions, `Buffer` response bodies, assertions as annotations, and client request fidelity
  - Depends on: `RMI-PIDL-135`
- [ ] `RMI-PIDL-143` Joining a client source to a wire source by method, rewritten URL and time window, with a report of unpaired exchanges
  - Depends on: `RMI-PIDL-142`
- [ ] `RMI-PIDL-144` `pidl observe check`: exit codes, the instance and status tables, and the JSON report
  - Depends on: `RMI-PIDL-150`
- [ ] `RMI-PIDL-145` A second protocol example with a small captured run, passing with no change to the tool
  - Depends on: `RMI-PIDL-144`

## Phase 4 — The AAuth Demo and Release

**Theme:** Developers learning AAuth see their own run beside the explanation

- [ ] `RMI-PIDL-146` AAuth document: match rules for its standalone steps, templates for the literal identifiers in its example paths, token carriage, and the challenge-first resource, checked against the Petstore capture with every observable step a strong match
  - Depends on: `RMI-PIDL-137`, `RMI-PIDL-150`
- [ ] `RMI-PIDL-147` Client events: the event observation, joined by request ID or by method, URL and time, and shown on the step they explain
  - Depends on: `RMI-PIDL-143`
- [ ] `RMI-PIDL-148` Guide for the three files and the commands, what a match does and does not mean, and handling observations of real systems
  - Depends on: `RMI-PIDL-140`, `RMI-PIDL-144`
- [ ] `RMI-PIDL-149` Release notes and changelog, with the demo target documented for consuming repositories
  - Depends on: `RMI-PIDL-146`, `RMI-PIDL-148`
