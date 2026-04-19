# PIDL Tasks

Protocol Interaction Description Language - a DSL for describing protocol choreography that compiles to diagrams.

## Phase 1: Core DSL & Diagram Generation (Current)

Primary goal: JSON-based DSL that generates sequence diagrams and data flow diagrams for MCP, A2A, OAuth, and OIDC protocols.

### 1.1 Specification

- [x] Define core PIDL JSON schema
  - [x] Entities (nodes): id, name, type, description
  - [x] Flows (edges): from, to, action, mode, label
  - [x] Phases (grouping): name, description, flows
  - [x] Protocol metadata: name, version, description
- [x] Define entity types for target protocols
  - [x] OAuth/OIDC: client, authorization_server, resource_server, user, browser
  - [x] MCP: agent, tool_server, tool
  - [x] A2A: agent, delegated_agent
- [x] Define flow modes: request, response, redirect, callback, interactive, event, tool_call, tool_result
- [x] Create JSON Schema for validation (`schema/pidl.schema.json`)
- [x] Write specification document (`docs/SPECIFICATION.md`)

### 1.2 Example Protocols

- [x] OAuth 2.0 Authorization Code flow (`examples/oauth2_authorization_code.json`)
- [x] OAuth 2.0 + PKCE flow (`examples/oauth2_pkce.json`)
- [x] OIDC Authentication flow (`examples/oidc_authentication.json`)
- [x] MCP tool invocation flow (`examples/mcp_tool_invocation.json`)
- [x] A2A agent delegation flow (`examples/a2a_agent_delegation.json`)

### 1.3 Diagram Generators

- [x] PlantUML sequence diagram output (`render/plantuml.go`)
- [x] Mermaid sequence diagram output (`render/mermaid.go`)
- [x] Graphviz DOT data flow diagram output (`render/dot.go`)
- [ ] D2 diagram output (optional)

### 1.4 CLI Tool (`cmd/pidl/`)

- [x] `pidl validate <file>` - validate PIDL JSON against schema
- [x] `pidl generate <file> -f plantuml|mermaid|dot` - generate diagram
- [x] `pidl examples` - list built-in protocol examples
- [x] `pidl examples <name>` - show example details
- [x] `pidl examples <name> -json` - show example JSON
- [x] `pidl init <file>` - scaffold a new protocol file
- [x] `pidl init -from <example> <file>` - initialize from example

### 1.5 Go Library

- [x] Parser: JSON to internal model (`parse.go`)
- [x] Model types: Protocol, Entity, Flow, Phase (`pidl.go`)
- [x] Validation logic (`validate.go`)
- [x] Renderer interface with implementations for each output format (`render/`)
- [x] Embed JSON Schema for validation (`schema/embed.go`)

## Phase 2: Enhanced Flow Semantics

- [ ] Conditional flows (`when:` clauses)
- [ ] Alternative paths (`alt:` blocks)
- [ ] Loop constructs (`loop:` blocks)
- [ ] Optional flows (`opt:` blocks)
- [ ] Flow annotations (notes, comments)

## Phase 3: State Model

- [ ] Entity state definitions
- [ ] State mutations on flows (`sets:` clause)
- [ ] State-based conditions
- [ ] State diagram generation (Mermaid stateDiagram)

## Phase 4: Security & Trust Annotations

- [ ] Trust levels: trusted, semi_trusted, untrusted, authoritative
- [ ] Security requirements on flows: token, signature, encryption
- [ ] Token definitions: type, issuer, audience, binding
- [ ] Trust boundary visualization in diagrams

## Phase 5: Protocol Composition

- [ ] Import mechanism for protocol modules
- [ ] Protocol inheritance/extension
- [ ] Reusable entity definitions
- [ ] Standard library of common protocols
  - [ ] oauth2_core
  - [ ] oauth2_pkce
  - [ ] oidc_core
  - [ ] mcp_core
  - [ ] a2a_core

## Phase 6: Runtime Execution Engine

- [ ] Protocol Execution Engine (PEX)
- [ ] Event queue and execution loop
- [ ] State store per entity
- [ ] Execution trace recording
- [ ] Step-by-step protocol simulation

## Phase 7: Analysis & Tooling

- [ ] Protocol comparison (diff two protocols)
- [ ] Execution trace visualization
- [ ] Interactive protocol debugger
- [ ] Attack surface analysis
- [ ] Replay attack detection
- [ ] Token theft simulation

## Phase 8: Integrations

- [ ] VS Code extension (syntax highlighting, preview)
- [ ] MkDocs plugin for embedding diagrams
- [ ] GitHub Action for CI validation
- [ ] Web playground for interactive editing

## Non-Goals (Out of Scope)

- HTTP/transport-level details (use OpenAPI for that)
- Message schema definitions (use JSON Schema/protobuf)
- Code generation for protocol implementation
- Full formal verification (use TLA+ for that)

## Target Protocols

| Protocol | Category | Priority |
|----------|----------|----------|
| OAuth 2.0 | Auth | P0 |
| OAuth 2.1 + PKCE | Auth | P0 |
| OpenID Connect | Auth | P0 |
| MCP (Model Context Protocol) | Agent | P0 |
| A2A (Agent-to-Agent) | Agent | P0 |
| SAML 2.0 | Auth | P1 |
| WebAuthn/FIDO2 | Auth | P1 |
| SCIM | Provisioning | P2 |
