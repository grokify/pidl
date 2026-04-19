# PIDL

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/grokify/pidl/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/grokify/pidl/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/grokify/pidl/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/grokify/pidl/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/grokify/pidl/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/grokify/pidl/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/pidl
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/pidl
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/pidl
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/pidl
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=grokify%2Fpidl
 [loc-svg]: https://tokei.rs/b1/github/grokify/pidl
 [repo-url]: https://github.com/grokify/pidl
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/pidl/blob/master/LICENSE

**Protocol Interaction Description Language** - A JSON-based DSL for describing protocol choreography that compiles to diagrams.

PIDL models protocols as directed interaction graphs between entities, enabling generation of sequence diagrams, data flow diagrams, and other visualizations. It's designed for protocols where the primary concern is "who talks to whom, in what order" rather than message schemas or transport details.

## Features

- **JSON-based DSL** for describing protocol flows
- **Multiple output formats**: PlantUML, Mermaid, Graphviz DOT
- **Built-in examples**: OAuth 2.0, PKCE, OIDC, MCP, A2A
- **CLI tool** for validation and diagram generation
- **Go library** for programmatic use

## Installation

```bash
go install github.com/grokify/pidl/cmd/pidl@latest
```

Or clone and build:

```bash
git clone https://github.com/grokify/pidl.git
cd pidl
go build -o pidl ./cmd/pidl
```

## Quick Start

### List available examples

```bash
pidl examples
```

### Generate a diagram from an example

```bash
# PlantUML sequence diagram
pidl generate oauth2_authorization_code

# Mermaid sequence diagram
pidl generate -f mermaid oauth2_pkce

# Graphviz DOT data flow diagram
pidl generate -f dot mcp_tool_invocation
```

### Create a new protocol file

```bash
# Create from scratch
pidl init my-protocol.json

# Copy from an example
pidl init -from oauth2_authorization_code my-oauth.json
```

### Validate protocol files

```bash
pidl validate my-protocol.json
pidl validate *.json
```

## PIDL Format

A PIDL document is a JSON file with three main sections:

```json
{
  "protocol": {
    "id": "my-protocol",
    "name": "My Protocol",
    "version": "1.0",
    "description": "A sample protocol",
    "category": "auth"
  },
  "entities": [
    {"id": "client", "name": "Client", "type": "client"},
    {"id": "server", "name": "Server", "type": "server"}
  ],
  "phases": [
    {"id": "main", "name": "Main Flow"}
  ],
  "flows": [
    {"from": "client", "to": "server", "action": "request", "label": "Request", "mode": "request", "phase": "main"},
    {"from": "server", "to": "client", "action": "response", "label": "Response", "mode": "response", "phase": "main"}
  ]
}
```

### Entity Types

| Type | Description |
|------|-------------|
| `client` | Application or service initiating requests |
| `authorization_server` | Issues tokens and handles authentication |
| `resource_server` | Hosts protected resources |
| `user` | Human actor |
| `browser` | User agent / web browser |
| `agent` | AI/LLM agent |
| `tool_server` | Exposes tools via protocol (MCP) |
| `delegated_agent` | Agent receiving delegated tasks (A2A) |

### Flow Modes

| Mode | Description | Arrow Style |
|------|-------------|-------------|
| `request` | Synchronous request | Solid `->` |
| `response` | Synchronous response | Dashed `-->` |
| `redirect` | HTTP redirect | Solid with annotation |
| `callback` | Callback/webhook | Solid with annotation |
| `interactive` | Human interaction | Solid |
| `event` | Asynchronous event | Dashed |
| `tool_call` | Tool invocation (MCP) | Solid with annotation |
| `tool_result` | Tool result (MCP) | Dashed with annotation |

## CLI Reference

```
pidl <command> [options] [arguments]

Commands:
  validate   Validate PIDL JSON files
  generate   Generate diagrams from PIDL files
  examples   List or show built-in examples
  init       Create a new PIDL file from template
  version    Print version information
  help       Show help message
```

### validate

```bash
pidl validate [options] <file> [file...]

Options:
  -q    Quiet mode (only show errors)
```

### generate

```bash
pidl generate [options] <file>

Options:
  -f string   Output format: plantuml, mermaid, dot (default "plantuml")
  -o string   Output file (default: stdout)
```

### examples

```bash
pidl examples [options] [name]

Options:
  -json   Show example JSON content
```

### init

```bash
pidl init [options] <filename>

Options:
  -name string   Protocol name
  -from string   Initialize from example
```

## Go Library

```go
import (
    "github.com/grokify/pidl"
    "github.com/grokify/pidl/render"
    "github.com/grokify/pidl/examples"
)

// Parse a PIDL file
p, err := pidl.ParseFile("protocol.json")

// Validate
if errs := p.Validate(); errs.HasErrors() {
    log.Fatal(errs)
}

// Generate PlantUML
diagram, err := render.RenderString(render.FormatPlantUML, p)

// Use built-in examples
names := examples.List()
oauth, err := examples.GetProtocol("oauth2_authorization_code")

// Create a new protocol
p := pidl.NewMinimalProtocol("my-protocol", "My Protocol")
pidl.WriteProtocolFile("output.json", p)
```

## Built-in Examples

| Example | Protocol |
|---------|----------|
| `oauth2_authorization_code` | OAuth 2.0 Authorization Code Flow |
| `oauth2_pkce` | OAuth 2.0 with PKCE |
| `oidc_authentication` | OpenID Connect Authentication |
| `mcp_tool_invocation` | MCP Tool Invocation |
| `a2a_agent_delegation` | A2A Agent Delegation |

## Target Protocols

PIDL is designed for describing:

- **Authentication/Authorization**: OAuth 2.0, OpenID Connect, SAML
- **Agent Protocols**: MCP (Model Context Protocol), A2A (Agent-to-Agent)
- **API Flows**: Multi-party API choreography

## Documentation

- [Specification](docs/SPECIFICATION.md) - Full language specification
- [JSON Schema](schema/pidl.schema.json) - Schema for validation

## License

MIT License - see [LICENSE](LICENSE) for details.
