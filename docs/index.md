# PIDL

**Protocol Interaction Description Language** - A JSON-based DSL for describing protocol choreography that compiles to diagrams.

PIDL models protocols as directed interaction graphs between entities, enabling generation of sequence diagrams, data flow diagrams, and other visualizations. It's designed for protocols where the primary concern is "who talks to whom, in what order" rather than message schemas or transport details.

## Features

- **JSON-based DSL** for describing protocol flows
- **Multiple output formats**: PlantUML, Mermaid, Graphviz DOT, D2, SVG
- **Animated SVG** with CSS flow visualizations and moving dots
- **Network boundary diagrams** for trust zone visualization
- **SVG templates**: default, minimal, sketch, blueprint, dark, high-contrast
- **Accessibility**: WCAG AA compliant high-contrast template
- **Phase boxes**: Visual grouping of flows by phase with color-coded depth
- **Interactive SVG**: Hover effects on participants, messages, and flow dots
- **Animation effects**: Pulse and glow for error/warning/highlight presets
- **Built-in examples**: OAuth 2.0, PKCE, OIDC, MCP, A2A
- **CLI tool** for validation and diagram generation
- **Go library** for programmatic use
- **Conditional flows** with `condition` field for when clauses
- **Alternative paths** with `alternatives` for error handling and branching
- **Annotations** with typed notes (security, performance, deprecated, etc.)
- **Nested phases** with parent hierarchy support
- **State model** with entity states and state transitions
- **State diagrams** via Mermaid `stateDiagram-v2` output

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

# SVG sequence diagram
pidl generate -f svg protocol.json

# Animated SVG with flow dots
pidl generate -f svg-animated --theme=dark protocol.json

# SVG with template
pidl generate -f svg --template=high-contrast protocol.json

# Network boundary diagram
pidl generate -f svg-network protocol.json

# Mermaid state diagram
pidl generate -f mermaid-state protocol.json

# State diagram for specific entity
pidl generate -f mermaid-state --entity=client protocol.json
```

## Protocol Architecture

PIDL models protocols as directed graphs where **nodes** are entities/components and **edges** are interactions/relationships.

### Nodes

#### Entity Types

Entity types are classified as client-side (initiators) or server-side (responders):

| Type | Classification | Description |
|------|----------------|-------------|
| `client` | Client | Application or service initiating requests |
| `user` | Client | Human actor |
| `browser` | Client | User agent / web browser |
| `agent` | Client | AI/LLM agent |
| `delegated_agent` | Client | Agent receiving delegated tasks (A2A) |
| `authorization_server` | Server | Issues tokens and handles authentication |
| `resource_server` | Server | Hosts protected resources |
| `identity_provider` | Server | Authenticates users and issues identity assertions |
| `service_provider` | Server | Consumes identity assertions |
| `server` | Server | Generic server |
| `tool_server` | Server | Exposes tools via protocol (MCP) |
| `tool` | Other | Individual tool capability |
| `other` | Other | Custom entity type |

#### Protocol Roles

Entities can implement protocol-specific roles:

| Protocol | Description | Common Roles |
|----------|-------------|--------------|
| `oauth` | OAuth 2.0 / OIDC | `authorization_server`, `client`, `resource_server` |
| `scim` | SCIM Provisioning | `client`, `service_provider` |
| `spiffe` | SPIFFE Identity | `workload`, `spire_server`, `spire_agent` |
| `aauth` | Agent Authentication | `person_server`, `access_server`, `client` |
| `idjag` | ID-JAG Identity | `client`, `person_server`, `access_server` |
| `authzen` | AuthZEN Authorization | `pep`, `pdp`, `pip`, `pap` |
| `mcp` | Model Context Protocol | `client`, `server`, `tool` |
| `a2a` | Agent-to-Agent | `client_agent`, `remote_agent` |

#### Deployment Components

Components group entities into logical deployment units:

| Component Type | Description | Example Products |
|----------------|-------------|------------------|
| `idp` | Identity Provider | Okta, Entra ID, Auth0 |
| `iga` | Identity Governance | SailPoint, Saviynt |
| `agent_provider` | AI Agent Platform | Claude, ChatGPT, Gemini |
| `person_server` | AAuth Person Server | Identity verification service |
| `access_server` | AAuth Access Server | Token issuance service |
| `pdp` | Policy Decision Point | OPA, Cedar, Topaz |
| `gateway` | API Gateway | Kong, Apigee, AWS API Gateway |
| `mcp_client` | MCP Client | Claude Code, Cursor |
| `mcp_server` | MCP Server | Custom tool servers |
| `resource_api` | Resource API | Backend services |
| `spire` | SPIFFE Runtime | SPIRE Server/Agent |

### Edges

#### Protocols

| Protocol | Full Name | Specification |
|----------|-----------|---------------|
| `oauth` | OAuth 2.0 / OpenID Connect | RFC 6749, RFC 6750, OIDC Core |
| `scim` | System for Cross-domain Identity Management | RFC 7643, RFC 7644 |
| `spiffe` | Secure Production Identity Framework | SPIFFE/SPIRE |
| `aauth` | Agent Authentication | AAuth Spec |
| `idjag` | Identity for Just-in-time Access Grant | ID-JAG Spec |
| `authzen` | Authorization API | OpenID AuthZEN |
| `mcp` | Model Context Protocol | MCP Specification |
| `a2a` | Agent-to-Agent Protocol | A2A Specification |

#### Trust Relationships

| Type | Description | Direction |
|------|-------------|-----------|
| `authenticates` | Verifies identity | A authenticates B |
| `validates` | Validates credentials/tokens | A validates B's tokens |
| `delegates` | Grants delegated authority | A delegates to B |
| `authorizes` | Grants access permissions | A authorizes B |
| `issues` | Creates credentials/tokens | A issues tokens for B |
| `trusts` | General trust relationship | A trusts B |
| `provisions` | Creates/manages accounts | A provisions B |
| `attests` | Provides identity attestation | A attests B's identity |

#### Credentials

| Credential | Description | Used By |
|------------|-------------|---------|
| `x509_svid` | X.509 SVID certificate | SPIFFE/SPIRE |
| `jwt_svid` | JWT SVID token | SPIFFE/SPIRE |
| `jwt_assertion` | JWT assertion for client auth | OAuth 2.0 |
| `access_token` | OAuth access token | OAuth 2.0 |
| `id_token` | OIDC ID token | OpenID Connect |
| `aa_agent_jwt` | AAuth Agent JWT | AAuth |
| `aa_auth_jwt` | AAuth Authorization JWT | AAuth |
| `mtls` | Mutual TLS certificate | mTLS |
| `api_key` | API key | Various |

## Target Protocols

PIDL is designed for describing:

- **Authentication/Authorization**: OAuth 2.0, OpenID Connect, SAML
- **Agent Protocols**: MCP (Model Context Protocol), A2A (Agent-to-Agent)
- **API Flows**: Multi-party API choreography

## Links

- [GitHub Repository](https://github.com/grokify/pidl)
- [Go Package Documentation](https://pkg.go.dev/github.com/grokify/pidl)
- [JSON Schema](https://github.com/grokify/pidl/blob/main/schema/pidl.schema.json)
