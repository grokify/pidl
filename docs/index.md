# PIDL

**Protocol Interaction Description Language** - A JSON-based DSL for describing protocol choreography that compiles to diagrams.

PIDL models protocols as directed interaction graphs between entities, enabling generation of sequence diagrams, data flow diagrams, and other visualizations. It's designed for protocols where the primary concern is "who talks to whom, in what order" rather than message schemas or transport details.

## Features

- **JSON-based DSL** for describing protocol flows
- **Multiple output formats**: PlantUML, Mermaid, Graphviz DOT, D2
- **Built-in examples**: OAuth 2.0, PKCE, OIDC, MCP, A2A
- **CLI tool** for validation and diagram generation
- **Go library** for programmatic use
- **Conditional flows** with `condition` field for when clauses
- **Alternative paths** with `alternatives` for error handling and branching
- **Annotations** with typed notes (security, performance, deprecated, etc.)
- **Nested phases** with parent hierarchy support

## Quick Example

```json
{
  "protocol": {
    "id": "simple-request",
    "name": "Simple Request/Response"
  },
  "entities": [
    {"id": "client", "name": "Client", "type": "client"},
    {"id": "server", "name": "Server", "type": "server"}
  ],
  "flows": [
    {"from": "client", "to": "server", "action": "request", "mode": "request"},
    {"from": "server", "to": "client", "action": "response", "mode": "response"}
  ]
}
```

## Installation

```bash
go install github.com/grokify/pidl/cmd/pidl@latest
```

## Generate Diagrams

```bash
# PlantUML
pidl generate -f plantuml protocol.json

# Mermaid
pidl generate -f mermaid protocol.json

# D2
pidl generate -f d2 protocol.json
```

## Target Protocols

PIDL is designed for describing:

- **Authentication/Authorization**: OAuth 2.0, OpenID Connect, SAML
- **Agent Protocols**: MCP (Model Context Protocol), A2A (Agent-to-Agent)
- **API Flows**: Multi-party API choreography

## Links

- [GitHub Repository](https://github.com/grokify/pidl)
- [Go Package Documentation](https://pkg.go.dev/github.com/grokify/pidl)
- [JSON Schema](https://github.com/grokify/pidl/blob/main/schema/pidl.schema.json)
