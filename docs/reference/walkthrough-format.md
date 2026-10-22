# Walkthrough Format

PIDL has two formats:

| | PIDL | Walkthrough |
|---|---|---|
| Written by | people | `pidl generate -f walkthrough-json` |
| Holds | roles, scenarios and flows, with `extends`, includes and sample references | every scenario with fragments expanded in place, cURL commands rendered, code samples resolved and analyzer findings attached |
| Read by | the pidl tools | the walkthrough viewer, and any tool that presents a protocol step by step |
| Schema | [`schema/pidl.schema.json`](https://github.com/grokify/pidl/blob/main/schema/pidl.schema.json) | [`schema/walkthrough.schema.json`](https://github.com/grokify/pidl/blob/main/schema/walkthrough.schema.json) |

A walkthrough is self-contained: a reader needs none of PIDL's include, ordering, resolution or analysis rules. This page describes version 1, `pidl-walkthrough/v1`.

## Producing a Walkthrough

```bash
pidl generate -f walkthrough-json -o api.walkthrough.json oauth2_api_access
```

From Go, `render.NewWalkthroughJSON().Render(w, protocol)` writes the same JSON, and `protocol.Walkthrough(findings)` returns it as a `pidl.Walkthrough`.

Name walkthrough files `<name>.walkthrough.json`, so their format is clear without opening them.

## Structure

```json
{
  "format": "pidl-walkthrough/v1",
  "id": "oauth2-api-access",
  "name": "OAuth 2.0 API Access",
  "language": "en-US",
  "roles": [{ "position": 1, "id": "person", "name": "Person" }],
  "scenarios": [
    {
      "position": 1,
      "id": "agent-identity",
      "name": "Agent identity",
      "participants": ["agent", "resource"],
      "steps": [
        { "number": 1, "from": "agent", "to": "resource", "label": "signed request", "location": "scenarios[0].flows[0] (agent-identity)" }
      ]
    }
  ]
}
```

### Walkthrough

| Field | Required | Holds |
|---|---|---|
| `format` | yes | `pidl-walkthrough/v1` |
| `id`, `name` | yes | the protocol's identifier and display name |
| `description` | | what the protocol is for |
| `language` | yes | the BCP 47 tag of the narration's source language |
| `pronunciations` | | how to speak terms, by term and language |
| `roles` | yes | the participants, in display order |
| `groups` | | navigation sections scenarios are listed under |
| `scenarios` | yes | the scenarios, in navigation order |
| `tokens` | | token definitions, with the claims each role checks |
| `findings` | | analyzer findings that concern no single step |

### Scenario

| Field | Required | Holds |
|---|---|---|
| `position`, `id`, `name` | yes | order, identifier and display name |
| `participants` | yes | the role IDs taking part, in display order |
| `steps` | yes | the messages, with included fragments expanded in place |
| `fragment` | | true for shared steps that other scenarios include |
| `group`, `variant_of`, `variant_label` | | navigation: its group, and the scenario it is a variant of |
| `description`, `narration`, `briefing` | | prose, spoken text by language, and an introduction |
| `samples` | | code samples for the whole scenario |

### Step

| Field | Required | Holds |
|---|---|---|
| `number` | yes | the 1-based position in the expanded scenario |
| `from`, `to`, `label` | yes | sender, receiver and message name |
| `location` | yes | where the flow is written in the PIDL file, such as `scenarios[1].flows[0] (consent)` |
| `id` | | the flow ID, `<include>.<id>` inside a fragment |
| `response` | | true for a response |
| `phase`, `fragments` | | the phase, and the fragments the step sits inside, outermost first |
| `description`, `note`, `narration`, `representations` | | prose, and spoken text by language, with overrides per representation (such as `curl`) |
| `signed_by`, `issues`, `verifies` | | security: the signer, and the tokens issued and verified |
| `example` | | the HTTP request or response, decoded tokens and code samples |
| `curl` | | a cURL command equivalent to the example request |
| `findings` | | analyzer findings about this step |

The schema has every field, with descriptions.

## Versions and Compatibility

`format` names the major version. Within a major version:

- New fields are optional, so every v1 walkthrough stays valid.
- Readers ignore fields they do not know. The schema allows extra properties for this reason.
- A field's type and meaning do not change.

Removing or renaming a field, changing its type or meaning, or making a field required starts a new major version, `pidl-walkthrough/v2`.

A reader accepts its own major version and refuses others with a message naming both. Walkthroughs written before the format had a version have no `format` field; read them as v1.

## Reading a Walkthrough

- **Validate** with any JSON Schema 2020-12 validator against `schema/walkthrough.schema.json`. In Go, the schema is embedded as `schema.WalkthroughSchemaJSON`.
- **TypeScript types** generated from the schema are in [`viewer/src/walkthrough.gen.ts`](https://github.com/grokify/pidl/blob/main/viewer/src/walkthrough.gen.ts).
- **Treat the content as data.** A walkthrough may come from anywhere, so render its text as text, never as HTML, and open only `http:` and `https:` links from it.

## Changing the Format

The Go types (`pidl.Walkthrough`) are the source. After changing them, regenerate the schema and the TypeScript types:

```bash
(cd schema/gen && go run .)
(cd viewer && npm run types)
```

CI fails when either file is out of date, and validates every built-in example against the schema.
