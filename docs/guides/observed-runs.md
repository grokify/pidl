# Observed Runs

A walkthrough explains a protocol with examples written to teach it. An observed run puts a real system's messages beside those examples: the request each step actually sent, the response it got, the tokens it carried, and the messages the protocol does not describe. A developer sees the concept and their own run side by side.

Three files meet in `pidl generate`:

| File | Format | Says |
|------|--------|------|
| The document | PIDL | what the protocol is: steps, examples, where tokens travel, match rules |
| An observation | `pidl-observation/v1` | what a real run exchanged, from a capture |
| A binding | `pidl-binding/v1` | how one deployment names its participants, hosts and paths |

Nothing in pidl knows any one protocol; the document and the binding carry everything it matches by.

## Quick Start

```bash
# Turn a capture into an observation
pidl observe import --from sandbox -o run.json trace.json

# Show it beside the document's steps, in a page that opens from disk
pidl generate -f viewer --observation run.json --binding deployment.json -o live.html protocol.json
```

`generate` reports the scenario instances it found and any warnings for the document's author:

```text
observed: 2 scenario instance(s) [agent-identity, ps-authorization-challenge]; 1 exchange(s) outside any instance, 0 outside the time window
```

`-f walkthrough-json` writes the same data as a walkthrough for your own viewer page. For an end-to-end example, the [aauth-postman](https://github.com/aistandardsio/aauth-postman) repository's `make demo` runs the AAuth Petstore flow against the aauth-go sandbox and writes the page.

## Captures

`pidl observe import --from sandbox` reads the aauth-go sandbox's trace (`aauth-sandbox trace`), which records every message between the roles it serves, including those a client never sees. An observation keeps each exchange's time, the capturing tool's labels for sender and receiver, the request and response with headers in order, and bodies as JSON, text or base64.

## Bindings

A binding describes one deployment, so the same document can be matched to any of them:

```json
{
  "format": "pidl-binding/v1",
  "participants": {
    "agent": {"labels": ["agent"]},
    "ap": {"labels": ["ap"], "hosts": ["ap.example.test"]},
    "resource": {"labels": ["petstore.example.test"], "hosts": ["petstore.example.test"]}
  },
  "labelsByPath": [
    {"label": "client", "path": "^/(\\.well-known/.*|jwks\\.json)$", "participant": "resource"}
  ],
  "paths": {"resource": {"agent-identity/request": "/v2/pet/findByStatus"}},
  "rewrite": [{"match": "^http://proxy:8099/(https://.*)$", "to": "$1"}],
  "unobservable": ["person"],
  "steps": {
    "*/api-request": {"match": {"method": "POST", "path": "/v2/store/order", "status": 200}}
  }
}
```

- `participants` maps each document participant to the labels a capture uses for it and the hosts it answers at.
- `labelsByPath` resolves a generic sender label, such as a caller the capture cannot name, by the request path.
- `paths` gives a deployment's path for a step where it differs from the example.
- `rewrite` fixes URLs before matching, such as the prefix a signing proxy adds.
- `unobservable` names participants a capture cannot see, in addition to people (role type `user`).
- `steps` replaces a step's rule for this deployment.

Keys in `steps` and `paths` name a step, most specific first: `<scenario>/<step-id>` as the walkthrough qualifies it (`ps-authorization/consent.poll`), `<defining-scenario>/<flow-id>` for a flow wherever it appears (`consent/poll`), or `*/<flow-id>` for any flow with that ID.

## Matching

Each step has a rule, built in layers: the method, path and status of its example, then the document's `match`, then the binding's, each replacing the fields it sets.

- **Required.** An exchange can be a step only when its participants and method agree.
- **Scored.** The path (a template such as `/pending/{id}`, or a regular expression starting with `^`), the status, named headers and body predicates raise the score. A match is **strong** when one of them passed and **weak** when only the required tests did.
- **Pairs.** A document describes one HTTP exchange as two steps, the request and its response. The response step takes its request's exchange.
- **Repeats.** A step with `"repeat": true`, such as a poll, takes every matching exchange between its neighbours.
- **Instances.** A capture usually holds several journeys. They are split where a scenario's first step recurs; at each point pidl takes the scenario, other than a fragment, that explains the most exchanges while leaving the fewest required steps unseen, or the ones you name with `--scenarios`. `--from` and `--to` leave out activity before or after the run.

Each step ends **matched**, **not seen**, **skipped** (an optional step), **not matchable** (no rule), or **unobservable**; each exchange that matched no step is **extra**, shown after the step it followed.

A match means an exchange looked like the step. It does not mean the message was valid: signatures, schemas and policy are not checked.

## Rules in the Document

A document can help every deployment match it:

```json
{"id": "poll", "from": "agent", "to": "ps", "action": "poll", "label": "Poll",
 "match": {"path": "/pending/{id}", "repeat": true}}
```

pidl warns about a step whose example path holds a literal identifier, such as `/pending/f7a3b9c`, that no real run repeats, and suggests a template.

Tokens are found where their definitions say they travel:

```json
{"id": "agent_token", "carried_in": [
  {"in": "header", "name": "Signature-Key", "param": "jwt", "typ": "aa-agent+jwt"},
  {"in": "body", "path": "/agent_token", "typ": "aa-agent+jwt"}
]}
```

`typ` tells tokens that travel in the same place apart. Found tokens are decoded, labeled "decoded, not verified", and shown in the viewer's Tokens tab.

## Real Tokens

An observation of a real system holds real tokens. Outputs mask raw token strings, signatures, and credentials in `Authorization` and `Cookie` headers, and keep the decoded claims, which are what a reader learns from. `--raw` keeps the raw values for a developer's own sandbox, with a warning. Treat observations and raw outputs as being as sensitive as the traffic they came from: do not commit or share them.

## Related

- [Walkthrough Viewer](viewer.md): the Concept / Observed switch
- [Walkthrough format](../reference/walkthrough-format.md): the observed fields
- Schemas: `schema/observation.schema.json` and `schema/binding.schema.json`
