# PRD — Live runs in the viewer

**Initiative:** `INIT-PIDL-011`
**Repository:** `github.com/grokify/pidl`

## Problem

A PIDL walkthrough explains a protocol with examples written to teach it: invented hosts, placeholder signatures, tokens made up to show their claims. That is the right way to teach a protocol, and it leaves a gap for the people who build with it. A developer reads that the resource answers with a challenge, then runs a real system and cannot tell which message in their logs is that challenge, whether the exchange between two servers they never see happened, or why their client sent the next request.

Real systems already leave records. A reference implementation can keep a trace of every message its servers exchange; a proxy exports HAR; Postman's command-line runner, Newman, writes a run report with a pass or fail for each step. None of them can be put beside the walkthrough, so a developer switches between the documentation and a wall of JSON, and an implementer has no quick way to see which steps their system performed and which it skipped.

A real capture is also messier than an example. It holds several journeys one after another, health checks and other noise, key and metadata fetches between the steps a document describes, messages that repeat while a client polls, and the deployment's own paths and host names.

## Goals

- **A neutral observation format.** One shape for a captured run: a list of HTTP exchanges, each with its request and response, whatever tool captured it.
- **Adapters for the captures that exist.** A role-labeled trace (the aauth-go sandbox's), HAR, and a Newman run report.
- **Matching by the document and a binding.** Observed exchanges are matched to a scenario's steps using rules the PIDL document carries and a binding file that describes one deployment: its hosts, the labels its tools use, and its paths. The tool holds no knowledge of any one protocol.
- **Real captures, not tidy ones.** A capture splits into the scenario instances it contains, repeated messages fold into one step, and noise outside a time window is left out.
- **The same viewer for both.** A step keeps its explanation and example and gains the exchange that was observed, with every token decoded. A switch shows the concept or the observed run, and the exchanges the document does not describe appear where they happened.
- **Honest gaps.** An observed exchange that matches no step is shown as extra. A step that never appeared is marked not seen.
- **A check mode.** A command that fails when required steps were not observed, for continuous integration.
- **A first consumer.** The AAuth example, run against the aauth-go sandbox from Newman and from omniagent, so developers learning AAuth see their own run beside the explanation.

## Non-Goals

- **Conformance verification.** A match means an exchange looks like the step: the right participants and method, and, where the rules say so, path, status, headers and body. It does not mean the message is valid. Signature verification, schema checks and policy are out of scope, and the viewer says "decoded, not verified" beside every token.
- **Capturing traffic.** PIDL does not proxy, record or replay. It reads files that other tools produce.
- **Interleaved concurrent journeys.** Scenario instances are matched one after another. Two journeys whose messages interleave in time are not separated in this initiative.
- **Steps that are not HTTP messages.** The inspector for those belongs to `INIT-PIDL-009`.
- **Live streaming.** Observations are files read after a run.
- **Exporting to telemetry or security-event systems.** OpenTelemetry and OCSF exports are a later initiative that can read the same observation format.
- **A hosted service.** The viewer reads a local file or a self-contained page, as it does today.

## Requirements

### Observation Format

1. **Format identifier.** An observation file starts with `"format": "pidl-observation/v1"`. A reader refuses another major version with a message naming the version it found.
2. **Exchanges.** `exchanges` is an ordered list. Each has a time, optional `from` and `to` labels as the capturing tool wrote them, an optional request ID, a request (method, URL, headers in order, body) and a response (status, headers in order, body).
3. **Source.** The file records which tool produced it, the time range it covers, and whether its requests are the messages on the wire or what a client sent before a proxy changed them.
4. **Bodies.** A JSON body is kept as JSON and any other body as a string; a binary body is base64 with its media type. A body above a size limit is cut, and the cut is recorded.

### Adapters

5. **Role-labeled trace.** An adapter reads the aauth-go sandbox trace, which already has the observation's shape, and keeps its labels as they are. Labels are the tool's own names, not PIDL participant IDs; the binding maps them.
6. **HAR.** An adapter reads HAR 1.2. It takes labels from the custom fields `_from` and `_to` when present and otherwise leaves the binding to resolve the hosts.
7. **Newman.** An adapter reads the JSON reporter's output. It yields each request's name, position and assertions as evidence, and plain HTTP exchanges for APIs that need no signing. It records that its requests are what Postman sent, not what was on the wire when a signing proxy sat between them.
8. **Joining sources.** A Newman observation can be joined to a trace of the same run, so the wire messages carry Newman's names and assertions. Exchanges are joined by method, the URL after the binding's rewrites, and a time window, never by position alone.
9. **No new Go dependencies.** Adapters use small internal types for the fields they read.

### Binding

10. **A binding file.** `pidl-binding/v1` describes one deployment: each participant's hosts, the labels a capturing tool uses for it, the paths of its endpoints where they differ from the document's examples, URL rewrites a capture needs (such as the prefix a signing proxy adds), and per-step overrides.

### Matching

11. **Required tests.** An exchange can match a step only when the participants agree (by label or host, through the binding) and the method agrees.
12. **Scored tests.** The path (as a template or as the binding maps it), the response status, and anything an explicit rule names raise the score. A deployment's own paths and hosts never prevent a match on their own.
13. **Default rules.** Where a step's example has a request and response, its method, path and status are the default rules. A step whose example has none matches only through an explicit rule or the binding; until then it is reported as not matchable, never silently not seen.
14. **Explicit rules.** A step may carry a `match`: method; path as a template or regular expression; status; headers, compared with names in any case and values as exact strings, prefixes or regular expressions; body predicates by JSON pointer; `repeat`; and `optional`. A binding may override a step's rule for its deployment.
15. **Ordered alignment.** Within a scenario instance, exchanges are aligned to steps in order, maximizing the total score, so a message that recurs in a scenario matches its own step.
16. **Repeated messages.** A step with `repeat: true` takes every consecutive exchange that matches it, such as a client polling while a person decides.
17. **Scenario instances.** A capture is split into the consecutive scenario instances it contains. The user names the scenarios to look for, in order, or the tool chooses at each point the scenario that matches best and reports its choice.
18. **Time window.** Matching can be limited to a time range, to leave out activity before or after the run.
19. **Statuses.** Each step ends `matched`, `notSeen`, `skipped` (optional, or on an alternative path) or `unmatchable` (no rule). Each exchange is `matched` or `extra`, and an extra exchange records the step it followed.

### Output and Check

20. **Built by the existing generator.** `pidl generate -f walkthrough-json` and `-f viewer` take an observation and a binding, the way they already take code samples. The viewer output is a self-contained page that opens from disk.
21. **Additive fields.** The walkthrough keeps each step's `example` and adds the observed exchange and the step's status. Field names follow the walkthrough format's version: snake_case in v1, camelCase in v2. Nothing already in the walkthrough changes meaning.
22. **Run summary.** The walkthrough records the source, the time range, the scenario instances found, and the counts of matched, not-seen, skipped, unmatchable and extra items.
23. **Check.** `pidl observe check` exits non-zero when a required step is `notSeen`, prints a short report, and can write it as JSON.

### Tokens and Redaction

24. **Token carriage in the document.** A token definition says where the token travels: a header and parameter, or a body field. Decoding follows the definition, so the tool names no protocol's headers.
25. **Decoded, not verified.** Tokens are decoded to their header and claims. The output never claims a token is valid.
26. **Redaction by default.** Raw token strings, signatures, and credentials in `Authorization` and `Cookie` headers are masked in every output. Decoded claims are kept, because they are what teaches the protocol. `--raw` keeps the full values for a developer's own use.

### Client Events

27. **Why a request was sent.** A client may record its own events (such as "received a challenge for a person token, requesting one"). They join the observation by a request ID header when the client sends one, and otherwise by method, URL and time, and appear on the step they explain.

### Viewer

28. **Concept and Observed.** A switch shows a step's example or its observed exchange. A scenario without an observation shows no switch.
29. **Instances and badges.** The viewer names the scenario instance on screen (for example, instance 2 of 2), and each step shows its status.
30. **Extra exchanges in place.** An exchange the document does not describe appears between the steps it fell between, and opens in the same panels.
31. **Time.** An observed step shows when it happened and the gap from the step before it.

### Documentation and the First Consumer

32. **Guide.** A guide explains the three files (document, observation, binding), the commands, what a match does and does not mean, and how to handle observations of real systems.
33. **AAuth example.** The AAuth document gains the match rules and token carriage it needs, including rules for its steps whose examples have no request. The aauth-postman repository keeps the binding for the sandbox and a target that runs the scenarios and opens the result.

## Success Criteria

1. The sandbox trace of the Petstore run builds a walkthrough with two scenario instances (browsing, then ordering), every step of each matched, the health check before the run left out by the time window or shown as extra, and the key and metadata fetches shown as extra between the steps they fell between.
2. The ordering flow driven from Newman and from omniagent matches its steps in both runs. omniagent's additional calls, such as loading the API description, appear as extra.
3. Removing a step from the run, or sending it out of order, makes `check` fail and the viewer mark the step.
4. A second protocol example with a small captured run works with no change to the tool, only to its document and binding.
5. Every token in the AAuth run is decoded in the viewer, and an output written without `--raw` contains no raw token or signature value.
