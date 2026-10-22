# Scenarios

Most protocols have more than one journey through the same roles. AAuth has five access modes, from agent identity to federated authorization. OAuth has several grants. Scenarios let one PIDL file describe all of them with one shared list of roles, so every diagram, table and outline generated from the file shows roles and journeys in the same order.

The built-in `aauth_protocol` example is used throughout this page:

```bash
pidl scenarios aauth_protocol
```

```text
#   ID                       NAME                           KIND      STEPS  PARTICIPANTS
1   agent-identity           Agent identity                 main          3  ap, agent, resource
2   resource-managed         Resource-managed access (tw... main          8  person, ap, agent, resource
3   person-identity          Person identity                main          4  ap, agent, resource, ps
4   ps-authorization         PS authorization (three-party) main          9  person, ap, agent, resource, ps
5   federated-authorization  Federated authorization (fo... main         11  person, ap, agent, resource, ps, as
6   agent-token              Obtain agent token             fragment      2  ap, agent
7   person-token             Obtain person token            fragment      2  agent, ps
8   consent                  User consent                   fragment      4  person, agent, ps
```

## Writing Scenarios

Replace the top-level `flows` array with a `scenarios` array. Each scenario has an `id`, a `name` and its own `flows`:

```json
{
  "entities": [
    {"id": "person", "name": "Person", "type": "user"},
    {"id": "agent", "name": "Agent", "type": "agent"},
    {"id": "resource", "name": "Resource", "type": "resource_server"},
    {"id": "ps", "name": "Person Server", "type": "authorization_server"}
  ],
  "scenarios": [
    {
      "id": "agent-identity",
      "name": "Agent identity",
      "flows": [
        {"from": "agent", "to": "resource", "action": "signed_request"},
        {"from": "resource", "to": "agent", "action": "ok", "mode": "response"}
      ]
    }
  ]
}
```

## Order Comes from the File

Order is never set with numbers. It comes from where things sit in the file:

- **Roles** follow the `entities` array. Diagram columns, tables and outlines all use it. To change the column order of every diagram, move the entity.
- **Scenarios** follow the `scenarios` array, arranged by group when the file declares groups, with variants after their base and fragments last.
- **Steps** follow each scenario's `flows` array.

A scenario's `participants` list chooses which roles appear. It never reorders them, and validation reports a list that is out of `entities` order.

## Sections, Variants and Step IDs

A complete protocol has many scenarios. Groups put them in sections, variants pair alternative forms of one journey, and step IDs give each step a name that links, samples and narration can rely on:

```json
{
  "groups": [
    {"id": "identity", "name": "Agent identity"},
    {"id": "access", "name": "Access modes"}
  ],
  "scenarios": [
    {"id": "agent-identity", "name": "Agent identity", "group": "identity", "flows": [
      {"id": "sign", "from": "agent", "to": "resource", "action": "signed_request"},
      {"id": "ok", "from": "resource", "to": "agent", "action": "response", "mode": "response"}
    ]},
    {"id": "agent-identity-approval", "name": "Agent identity with approval", "group": "identity",
     "variant_of": "agent-identity", "variant_label": "With user approval", "flows": [ ... ]},
    {"id": "ps-authorization", "name": "PS authorization", "group": "access", "flows": [
      {"id": "request", "from": "agent", "to": "resource", "action": "request"},
      {"id": "consent", "include": "consent"}
    ]}
  ]
}
```

- **Groups** set the order of sections. `pidl scenarios` lists scenarios under their group's name, and `--all-scenarios` numbers files in the same order.
- **Variants** follow their base and share its group. Viewers show a variant as a toggle on its base's entry.
- **Step IDs** are unique within a scenario. A step inside an included fragment is addressed as `<include>.<id>`, for example `consent.approve`, so `#ps-authorization/consent.approve` links to that step in the viewer even after steps are added before it.

## Shared Steps: Fragments and Includes

Steps that several scenarios share, such as obtaining a token or asking the person for consent, go in a fragment: a scenario with `"fragment": true`. Other scenarios insert it with an include step:

```json
{
  "id": "ps-authorization",
  "name": "PS authorization (three-party)",
  "flows": [
    {"include": "agent-token"},
    {"include": "person-token"},
    {"from": "agent", "to": "resource", "action": "authorize"},
    {"include": "consent"}
  ]
}
```

Diagrams draw an include as a labeled box across the fragment's participants. Use `--expand` to show the fragment's steps in place instead.

## Generating Output

```bash
# One scenario
pidl generate -f mermaid --scenario ps-authorization aauth_protocol

# One scenario with fragments shown step by step
pidl generate -f mermaid --scenario ps-authorization --expand aauth_protocol

# Every main scenario: 01-agent-identity.svg, 02-resource-managed.svg, ...
pidl generate -f svg --all-scenarios -o diagrams/ aauth_protocol

# Add the fragments after the main scenarios: ... 06-agent-token.svg
pidl generate -f svg --all-scenarios --include-fragments -o diagrams/ aauth_protocol
```

File numbers match the `#` column of `pidl scenarios`, so a sorted file listing, a slide deck and the scenario list agree.

Some outputs describe the whole file rather than one scenario:

| Format | Output |
|--------|--------|
| `scenario-matrix` | Markdown table of roles by scenario. ● takes part directly, ○ takes part only through an included fragment |
| `token-table` | Markdown table of tokens: issuer, audience, binding, who checks each token, which claims they rely on, and the scenarios that issue and check it |
| `outline` | Markdown outline of roles, scenarios with numbered steps, and tokens, as a starting point for guides and slide decks |
| `outline-json` | The same outline as JSON, for deck and guide tooling |

```bash
pidl generate -f scenario-matrix aauth_protocol
pidl generate -f outline -o aauth-outline.md aauth_protocol
```

`simulate` and `debug` take `--scenario` and run the scenario with fragments expanded.

## Recording Who Checks What

Token definitions can list who checks each token and which claims they rely on. Flow security can say which tokens a step issues and checks, and who signs the request:

```json
{
  "metadata": {
    "tokens": [
      {
        "id": "resource_token",
        "issuer": "resource",
        "verified_by": ["ps", "as"],
        "claims": [
          {"name": "agent_jkt", "checked_by": ["ps", "as"],
           "description": "Thumbprint of the agent's key, compared with the key that signed the request"}
        ]
      }
    ]
  }
}
```

```json
{"from": "resource", "to": "agent", "action": "challenge", "mode": "response",
 "security": {"issues": ["resource_token"]}}
```

```json
{"from": "ps", "to": "as", "action": "federate",
 "security": {"requires": ["signature"], "verifies": ["resource_token"]}}
```

With these fields, `pidl analyze` follows each token through a scenario:

| Rule | Reports |
|------|---------|
| SEC017 | A role that no flow uses |
| SEC018 | A step that checks a token before any earlier step issues it |
| SEC019 | A request signed by someone other than its sender, or than the entity the presented token was delivered to |
| SEC020 | A receiver that checks a token without being listed in its `verified_by` |
| SEC021 | A fragment that no scenario includes |

In the AAuth example, SEC019 flags federated access: the PS signs the request to the AS, but the resource token it presents was delivered to the agent, so a check of `agent_jkt` against the signing key would compare against the PS's key.

```bash
pidl analyze aauth_protocol                                      # whole file
pidl analyze --scenario federated-authorization aauth_protocol   # one scenario
```

## Composing Files

`extends` and `imports` understand scenarios:

- **extends** keeps the base file's scenario order. A scenario with the same `id` in the derived file replaces the base one in place; new scenarios follow. `exclude_scenarios` removes base scenarios.
- **imports** can bring in selected scenarios with `"scenarios": ["id", ...]`. The fragments they include come with them, and `alias` prefixes every ID they reference.
- When composition leaves a file with both top-level `flows` and `scenarios`, the flows become the first scenario, named after the protocol.

## Comparing Versions

`pidl diff` matches scenarios by `id`. It reports scenarios that were added, removed or renamed, changed steps within each scenario, include steps that point at a different fragment, and roles or scenarios that moved relative to each other.

```bash
pidl diff aauth-draft-10.json aauth-draft-11.json
```

## In Go

```go
p, err := pidl.ParseFile("aauth.json")
if err != nil {
    return err
}
for _, s := range p.ScenarioSummaries() {
    proj, err := p.Scenario(s.ID, pidl.ProjectOptions{})
    if err != nil {
        return err
    }
    svg, err := render.RenderString(render.FormatSVG, proj)
    if err != nil {
        return err
    }
    name := fmt.Sprintf("%02d-%s.svg", s.Position, s.ID)
    if err := os.WriteFile(name, []byte(svg), 0o600); err != nil {
        return err
    }
}
```

`Protocol.Outline()` returns the ordered outline as Go values, `Protocol.Flatten()` returns every scenario's flows once, and `Protocol.TokenUsage(id)` lists the scenarios that issue and check a token.
