# TRD — Live runs in the viewer

**Initiative:** `INIT-PIDL-011`
**Repository:** `github.com/grokify/pidl`

## Overview

Three files meet in the existing generator.

```text
 document (pidl)      observation (pidl-observation/v1)      binding (pidl-binding/v1)
 steps, examples,     what a real run exchanged,             hosts, labels, paths,
 token carriage,      from a trace, HAR or Newman            rewrites, step overrides
 match rules
        \                      |                      /
   pidl generate -f walkthrough-json | viewer  --observation … --binding …
                              |
        walkthrough (or self-contained page) with observed exchanges
                              |
                    @grokify/pidl-viewer
                  (Concept / Observed switch)
```

The document says what a protocol is, the observation says what happened, and the binding says how the two line up on one deployment. Nothing in the tool names a protocol.

**Steps.** In the current source format a scenario lists `flows`, and an entry with `include` brings in a fragment scenario's flows. In `pidl/v1` (`INIT-PIDL-008`) they are `steps`. This document says *step* for either. The matcher works on a scenario after fragments are expanded, the same expansion the walkthrough builder already does.

## Package Layout

| Path | Contents |
|------|----------|
| `observe/` | the observation and binding types, reader and writer, validation, joining and redaction |
| `observe/match/` | rules, scoring, alignment, repeated steps, scenario instances and statuses |
| `observe/adapter/sandbox/` | reads a role-labeled trace |
| `observe/adapter/har/` | reads HAR 1.2 |
| `observe/adapter/newman/` | reads a Newman JSON report |
| `cmd/pidl/` | `--observation` and `--binding` on `generate`; `pidl observe import` and `pidl observe check` |
| `schema/observation.schema.json`, `schema/binding.schema.json` | generated from the Go types, with the usual drift test |
| `viewer/` | the Concept / Observed switch, instance label, badges and extra exchanges in place |

## Observation Format

```json
{
  "format": "pidl-observation/v1",
  "source": {"tool": "aauth-sandbox", "requestFidelity": "wire",
             "from": "2026-01-01T00:00:00Z", "to": "2026-01-01T00:00:03Z"},
  "exchanges": [
    {
      "time": "2026-01-01T00:00:00.120Z",
      "from": "agent",
      "to": "ap",
      "requestId": "4f1c…",
      "request":  {"method": "POST", "url": "https://ap.example.test/agent-token",
                   "headers": [["Content-Type", "application/json"]], "body": {"…": "…"}},
      "response": {"status": 200, "headers": [["Content-Type", "application/json"]], "body": {"…": "…"}},
      "label": "1. POST agent token",
      "annotations": [{"kind": "assertion", "text": "status 200", "ok": true}]
    }
  ]
}
```

- `from` and `to` are the capturing tool's labels, kept as written. The binding maps them to participants.
- `requestFidelity` is `wire` when the requests are the messages servers received, and `client` when they are what a client sent before a proxy changed them.
- `requestId` is the value of a request ID header when the capture has one; it joins client events and sources.
- Headers keep their order and their case, as pairs. Comparisons of header names ignore case.
- A JSON body is a JSON value; any other body is a string; a binary body is `{"base64": "…", "mediaType": "…"}`. A body larger than a limit (64 KiB by default) is cut and the exchange records `bodyTruncated: true`.
- `label` and `annotations` carry evidence a tool produced about an exchange, such as a Newman request name and its assertions, or a client's own events. They never change a match.
- Times are RFC 3339. An adapter that has no clock keeps the order and omits `time`.

## Adapters

Each adapter reads a file and writes an observation. Each declares the few fields it reads, in its own small types, so pidl imports no other module.

### Role-labeled trace

The aauth-go sandbox's trace is a list of exchanges with `time`, `from`, `to`, and a request and response each holding method or status, a URL, headers as ordered pairs and a body. It maps one to one, with `requestFidelity: "wire"`.

Its labels are the sandbox's names, not participants. `agent`, `ap` and `ps` name its roles; a resource it fronts appears by host name; and `client` is any caller that is not one of its roles, such as a resource server fetching the agent provider's keys. The binding resolves each one.

### HAR

HAR 1.2: `log.entries[]` with `startedDateTime`, `time`, `request` (method, URL, headers as name and value, `postData` with `mimeType` and `text`) and `response` (status, headers, `content` with `text`, `mimeType` and `encoding`). `content.encoding: "base64"` is decoded when the type is text and kept as base64 otherwise. The custom fields `_from` and `_to` on an entry become the labels. Other tools' custom fields are ignored. A HAR from a proxy in front of the servers is `wire`; one from a browser is `client`.

### Newman

The JSON reporter writes the run summary: `collection`, `environment`, `globals` and `run`. The adapter reads:

- `run.executions[]`, in order. Each has `item.name` (kept as `label`), `request` (`method`, `url` as an object of `protocol`, `host[]`, `port`, `path[]` and `query[]`, `header[]` as `key` and `value`, and `body.raw`), `response` (`code`, `header[]`, `responseTime`, and `stream`) and `assertions[]` (`assertion`, `skipped` and an optional `error`).
- `stream` is a Node `Buffer` written as `{"type": "Buffer", "data": [bytes…]}`. The adapter decodes the bytes as text when they are text.
- `run.timings.started` and the response times give times for the exchanges.

Each assertion becomes an `annotations` entry with its result; a failed assertion records its message. The source is `requestFidelity: "client"`: when a signing proxy sits between Postman and the server, Newman's request carries the proxy's URL prefix and lacks the signature headers the proxy adds.

### Joining sources

Given a `client` observation and a `wire` observation of the same run, the joiner pairs each client exchange with the first unpaired wire exchange that has the same method and the same URL after the binding's rewrites, within a time window (two seconds by default, configurable), and copies the client exchange's `label` and `annotations` onto it. Unpaired client exchanges are reported. Position alone is never used: a server's trace holds exchanges the client never made, such as key fetches between servers.

## Binding File

```json
{
  "format": "pidl-binding/v1",
  "participants": {
    "agent":    {"labels": ["agent"]},
    "ap":       {"labels": ["ap"], "hosts": ["ap.example.test"]},
    "ps":       {"labels": ["ps"], "hosts": ["ps.example.test"]},
    "resource": {"labels": ["petstore.example.test"], "hosts": ["petstore.example.test"]}
  },
  "labelsByPath": [
    {"label": "client", "path": "^/(\\.well-known/.*|jwks\\.json)$", "participant": "resource"}
  ],
  "paths": {
    "ap":       {"agent-token.request": "/agent-token"},
    "resource": {"api-request": "/v2/store/order"}
  },
  "rewrite": [
    {"match": "^http://sandbox:8099/(https://.*)$", "to": "$1"}
  ],
  "steps": {
    "consent": {"match": {"method": "GET", "path": "^/pending/", "repeat": true}}
  }
}
```

- `participants` maps a PIDL participant ID to the labels a capturing tool uses for it and the hosts it answers at. An exchange's `to` resolves by label, then by the URL's host. Its `from` resolves by label.
- `labelsByPath` resolves a generic label, such as a trace's `client`, by the request path. Where it cannot be resolved, the exchange keeps an unknown sender, which still matches a step whose sender is not required to agree.
- `paths` gives a participant's path for a step where the deployment's path differs from the document's example. It replaces the example's path for that step's rule.
- `rewrite` applies regular-expression rewrites to every exchange's URL before matching, in order.
- `steps` overrides a step's `match` for this deployment only.

## Matching

### Rules

Each step has a rule. It comes from, in order of precedence: the binding's `steps` override, the step's own `match`, and the defaults drawn from its example. The binding's `paths` replace the rule's path where they name the step.

**Defaults.** Where the example has a request and a response, the defaults are its method, its path read as a template (segments in braces are wildcards), and its status. Where the example has no request, the step has no default rule. Unless a `match` or the binding supplies one, it is `unmatchable`, and the tool reports it once per run, so an author knows which steps need rules.

**Explicit rule.**

```json
"match": {
  "method": "POST",
  "path": "/v2/store/order",
  "status": 401,
  "headers": {"AAuth-Requirement": "requirement=auth-token*"},
  "body": {"/grant_type": "authorization_code"},
  "repeat": false,
  "optional": false
}
```

- `path` is a template, or a regular expression when it starts with `^`.
- `headers` compare names in any case. A value is an exact string, a prefix when it ends with `*`, or a regular expression when it starts with `^`.
- `body` maps a JSON pointer to a value that must be equal, or to `true` meaning the member must exist.
- `repeat: true` lets the step take every consecutive exchange that matches it.
- `optional: true` lets a run omit the step without failing a check.

### Scoring

A pair of an exchange and a step is a candidate only when the participants agree (an unknown sender agrees with any) and the method agrees. Its score is the sum of the tests that pass, with initial weights: path 3, status 2, each named header 2, each body predicate 2, sender known and agreeing 1. A test the rule does not name is not scored. The weights are constants in one place, and the report prints each pair's score so a poor default shows.

### Alignment within an instance

Exchanges are assigned to the steps of one scenario instance by dynamic programming over the two sequences. The assignment keeps order, uses each exchange once, assigns each step at most one exchange (or a consecutive run of them when it repeats), and maximizes the total score. Exchanges left over are `extra`; steps left over are `notSeen`, `skipped` when optional or on an alternative path, or `unmatchable` when they have no rule. Ties prefer the earlier exchange.

### Scenario instances

A capture usually holds more than one journey. The matcher reads it from left to right within the time window:

1. At the first unconsumed exchange, try each candidate scenario: the next one in the user's `--scenarios` list, or every scenario when no list is given.
2. An instance of a scenario starts at the first exchange that matches its first matchable step, and runs until the exchange before one that starts a new instance of any candidate after the current instance has matched at least one later step, or to the end of the window.
3. Align the instance. Score it by the share of its required steps matched, then by total score.
4. Take the best-scoring candidate, record it as an instance, mark exchanges before its start as `extra`, and continue after its last exchange.

The report lists the instances found and, for each, the runner-up scenario and its score. Interleaved concurrent journeys are not separated (a non-goal); they show as one instance with many extra exchanges.

### Time window

`--from` and `--to` limit matching to exchanges in a time range. Exchanges outside it are left out of the walkthrough and the counts, and the summary says how many.

## Tokens

A token definition gains `carried_in` (`carriedIn` in `pidl/v1`):

```json
{"id": "agent_token", "carried_in": [
  {"in": "header", "name": "Signature-Key", "param": "jwt"},
  {"in": "body", "path": "/agent_token"}
]}
```

The current examples name a token's place only in a free-text label, so this field is new. For a matched step, the builder finds each defined token in the observed request and response, splits a JWT into header, payload and signature, and writes the first two decoded into the observed exchange's tokens, with the claims the definition lists marked. It does not verify a signature, and the output labels each token "decoded, not verified". A value that is not a JWT is shown as it is.

## Walkthrough Output

The walkthrough changes only by addition. The new names follow the version of the walkthrough format being written:

| Meaning | `pidl-walkthrough/v1` | `pidl-walkthrough/v2` |
|---------|-----------------------|-----------------------|
| The step's observed exchange | `observed` | `observed` |
| The step's status | `observed_status` | `observedStatus` |
| The run's summary on a scenario | `observed_run` | `observedRun` |
| Exchanges between steps | `observed_extra` | `observedExtra` |
| The step an extra exchange followed | `after_step` | `afterStep` |
| An observed exchange's request ID | `request_id` | `requestId` |

In v1:

```json
{
  "scenarios": [{
    "id": "ps-authorization",
    "observed_run": {"source": "aauth-sandbox", "instance": 2, "instances": 2,
                     "matched": 9, "not_seen": 0, "skipped": 0, "unmatchable": 0, "extra": 6},
    "steps": [{
      "id": "…",
      "example": {"…": "unchanged"},
      "observed": {"time": "…", "request": {"…": "…"}, "response": {"…": "…"},
                   "tokens": [], "annotations": [], "repeats": 1},
      "observed_status": "matched"
    }],
    "observed_extra": [{"after_step": "person-token", "time": "…", "request": {"…": "…"}, "response": {"…": "…"}}]
  }]
}
```

A scenario instance found in the capture is written as its own scenario entry, with the instance number in its summary. `observed` and the status are absent on a step with no observation, and a viewer that does not know them ignores them. Both schemas gain the fields, and the generated TypeScript types follow.

## Redaction

Every output masks raw token strings (wherever a token definition's carriage finds them, and any value shaped like a JWT), `Signature` values, the token inside `Signature-Key`, and the credentials in `Authorization` and `Cookie` headers. A masked value becomes `[redacted:jwt]` or `[redacted]`. Decoded token headers and claims are kept, because they are what a reader learns from. `--raw` writes full values, for a developer's own sandbox; the command then warns that the file is as sensitive as the traffic it came from.

## Client Events

A client may write its own events as an observation with no exchanges and one annotation per event (`{"kind": "client-event", "text": "…", "time": "…", "request": {"method": "…", "url": "…"}, "requestId": "…"}`). The joiner attaches each event to the exchange with the same `requestId` when both have one, and otherwise to the first exchange with the same method and URL at or after the event's time, within the time window. Events that attach to nothing are listed in the report. omniagent is the first client expected to write them.

## Commands

```bash
pidl observe import --from sandbox|har|newman|events <file> [-o observation.json]
pidl generate -f walkthrough-json|viewer <document> --observation <file>... [--binding b.json]
              [--scenarios id,id] [--from t] [--to t] [--raw] [-o out]
pidl observe check <document> --observation <file>... [--binding b.json]
              [--scenarios id,id] [--from t] [--to t] [--json report.json]
```

- `generate` accepts several observations: a `wire` source, and `client` sources or client events to join to it. With `-f viewer` it writes a self-contained page that opens from disk, because the embeddable viewer fetches only same-origin and `https:` URLs.
- `check` exits 0 when every required step of every instance is matched, 1 when a required step is `notSeen` or `unmatchable`, and 2 for an input error. It prints the instances, a table of step statuses and the extra exchanges.

## Viewer

- A scenario with an observed run shows a **Concept / Observed** switch. The switch changes which of a step's example or observed exchange the request, response and token panels show; the sequence diagram does not change.
- The header names the instance, such as "Observed: instance 2 of 2, aauth-sandbox".
- A badge on each step shows its status, and a repeated step shows how many exchanges it took. A `notSeen` or `unmatchable` step has no observed panel and says which.
- Extra exchanges appear between the steps they fell between, drawn dashed between their participants, and open in the same panels.
- An observed step shows its time and the gap from the previous observed exchange.
- The viewer keeps reading walkthroughs with none of these fields.

## Dependencies

No new Go dependencies. HAR and Newman are read with small types of pidl's own, because pidl has one dependency today and each of the alternatives adds about twenty-seven modules:

- `github.com/grokify/mogo/net/http/har` has HAR types, and a package for them is worth fixing on its own account, but importing it brings in all of mogo's module graph.
- `github.com/grokify/postman-go` is a client for Postman's cloud API, generated from its OpenAPI description. A Newman report is the output of a different tool.
- A new module for capture formats is more structure than two readers justify.

If a second tool needs the Newman reader, it moves to a hand-written `newman` package in postman-go, and pidl can import it then.

## Testing

- **Generated tokens.** Tests that decode tokens build them at test time, signed with a throwaway key, so they are real JWTs with known claims.
- **Captured fixtures.** A sandbox trace of the AAuth Petstore run (two scenario instances, a health check before them, key fetches between steps), the same run as a Newman report, and a small HAR. Token and signature values in committed fixtures are truncated with an ellipsis; tests that need a decodable token use the generated ones.
- **Golden tests.** Each adapter's output, the join, the built walkthrough in v1 and v2, the self-contained page, and the check report.
- **Matching properties.** Exchanges in order all match; swapping two makes one extra or a step not seen; a duplicated exchange is extra; a polled step absorbs its repeats; removing a required step gives `notSeen` and exit 1; a step with no example request and no rule is `unmatchable`; header names match in any case.
- **Instances.** A capture with two journeys gives two instances; noise before the first is extra or left out by the time window.
- **A second protocol.** An OAuth authorization-code example from the built-in documents, with a small hand-written run, exercises the matcher with no change to the code.
- **Redaction.** No output written without `--raw` contains a fixture's raw token or signature value.
- **Schemas.** The generated schemas pass the drift test and the camelCase lint.
- **Viewer.** Rendering tests for the switch, the instance label, the badges and the extra exchanges in place.

## Security and Privacy

- An observation of a real system holds real tokens. The tool reads and writes local files; the viewer loads a file or page it is given and calls no service.
- Outputs are redacted unless `--raw` is given, and the guide says that an observation is as sensitive as the traffic it came from and is not for committing or sharing.
- Committed fixtures carry no values that grant access.

## Open Questions

- Whether a step's `match` belongs in the source format or only in a binding. It starts in the document, because most deployments want the same rules, and the binding overrides it.
- Whether alternatives should be matched as separate candidates in the first release or only reported as `skipped`.
- Whether `pidl observe check` should accept a threshold for extra exchanges.
- Whether key and metadata fetches, which every AAuth run contains, should become optional steps in the AAuth document rather than extra exchanges.
