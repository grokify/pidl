# PIDL Roadmap

Protocol Interaction Description Language - a JSON-based DSL for describing protocol choreography that compiles to diagrams.

## Release History

| Version | Date | Highlights |
|---------|------|------------|
| v0.3.0 | 2025-05-18 | SVG rendering, animations, templates, network boundaries |
| v0.2.0 | 2025-04-18 | Annotations, alternatives, conditions, nested phases, D2 |
| v0.1.0 | 2025-04-18 | Core DSL, PlantUML, Mermaid, DOT renderers |

---

## Completed Phases

### Phase 1: Core DSL & Diagram Generation ✅ (v0.1.0)

- [x] Core PIDL JSON schema with entities, flows, phases
- [x] Entity types for OAuth/OIDC, MCP, A2A protocols
- [x] Flow modes: request, response, redirect, callback, interactive, event
- [x] JSON Schema for validation
- [x] PlantUML, Mermaid, Graphviz DOT renderers
- [x] CLI tool: validate, generate, examples, init
- [x] Go library with parser, model types, validation

### Phase 2: Enhanced Flow Semantics ✅ (v0.2.0)

- [x] Conditional flows (`condition` field)
- [x] Alternative paths (`alternatives` field)
- [x] Flow annotations with typed notes
- [x] Nested phases with parent hierarchy
- [x] D2 diagram renderer (sequence, flow, arch styles)

### Phase 3: SVG Rendering ✅ (v0.3.0)

- [x] Static SVG sequence diagrams
- [x] Animated SVG with CSS offset-path flow dots
- [x] Per-message animation control
- [x] Semantic animation presets (request, success, error, warning, highlight)
- [x] Light/dark theme support
- [x] Template system with 5 built-in styles

### Phase 4: Network Boundary Diagrams ✅ (v0.3.0)

- [x] SVG network boundary renderer
- [x] Boundary inference from entity metadata
- [x] Explicit boundary configuration
- [x] Boundary styles: trusted, dmz, external, cloud
- [x] CLI `--boundary` override flags

### Phase 5: Advanced SVG Features ✅ (v0.4.0)

- [x] Phase boxes with color-coded depth
- [x] Alternative paths with ALT badges
- [x] High-contrast WCAG AA template
- [x] Interactive hover states
- [x] Animation pulse/glow effects
- [x] Note rendering on flows

### Phase 6: State Model ✅ (v0.5.0)

- [x] Entity state definitions with `states` field
- [x] State mutations on flows with `sets` clause
- [x] Initial and final state markers
- [x] State validation (unique IDs, single initial, entity/state references)
- [x] Mermaid `stateDiagram-v2` output format
- [x] Entity filtering for state diagrams (`--entity` flag)
- [x] Example: `oauth2_with_states.json`

### Phase 7: Security & Trust Annotations ✅ (v0.6.0)

- [x] Trust levels: trusted, semi_trusted, untrusted, authoritative
- [x] Security requirements on flows: token, signature, encryption, mtls, mac
- [x] Token definitions: id, name, type, issuer, audience, binding
- [x] Trust boundary inference from entity trust levels
- [x] Security badges on SVG sequence diagrams
- [x] Security notes in PlantUML and Mermaid output
- [x] CLI flags: `--show-security`, `--show-trust`
- [x] Example: `oauth2_with_security.json`

---

## Future Phases

### Phase 8: Protocol Roles & Deployment Architecture (v0.7.0)

**Target:** Enable PIDL to describe protocol roles, logical deployment components, and trust relationships for architecture documentation.

**Use Case:** OAIAF needs to document how protocols map to real-world deployments (IdP, IGA, Gateway, MCP Client, etc.) with explicit trust relationships.

#### 8.1 Protocol Roles

Add protocol-specific role annotations to entities:

```go
// New type
type ProtocolRole struct {
    Protocol string `json:"protocol"`          // "oauth", "scim", "aauth", "authzen", "mcp", "a2a", "spiffe"
    Role     string `json:"role"`              // "authorization_server", "client", "pep", "pdp", etc.
    Variant  string `json:"variant,omitempty"` // Sub-role: "person_server" vs "access_server"
}

// Add to Entity
type Entity struct {
    // ... existing fields
    ProtocolRoles []ProtocolRole `json:"protocol_roles,omitempty"`
}
```

| Task | Status |
|------|--------|
| Add `ProtocolRole` type | Planned |
| Add `protocol_roles` field to Entity | Planned |
| Define standard role vocabulary per protocol | Planned |
| Validate protocol/role combinations | Planned |
| Update JSON Schema | Planned |

**Standard Role Vocabulary:**

| Protocol | Valid Roles |
|----------|-------------|
| `oauth` | `resource_owner`, `client`, `authorization_server`, `resource_server` |
| `scim` | `client`, `service_provider` |
| `spiffe` | `workload`, `agent`, `server`, `trust_domain` |
| `aauth` | `agent`, `person_server`, `access_server`, `agent_provider` |
| `idjag` | `assertion_issuer`, `token_endpoint`, `relying_party` |
| `authzen` | `pep`, `pdp`, `pap`, `pip` |
| `mcp` | `host`, `client`, `server` |
| `a2a` | `agent`, `registry` |

#### 8.2 Deployment Components

Add logical deployment component grouping:

```go
// New type
type DeploymentComponent struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    Type        string         `json:"type"`        // "idp", "iga", "gateway", "mcp_client", "resource_api", etc.
    Description string         `json:"description,omitempty"`
    Entities    []string       `json:"entities"`    // Entity IDs contained in this component
    Implements  []ProtocolRole `json:"implements"`  // Aggregate roles this component implements
    Examples    []string       `json:"examples,omitempty"` // Real-world products: "Okta", "Entra ID"
}

// Add to ProtocolMetadata
type ProtocolMetadata struct {
    // ... existing fields
    Components []DeploymentComponent `json:"components,omitempty"`
}
```

| Task | Status |
|------|--------|
| Add `DeploymentComponent` type | Planned |
| Add `components` field to ProtocolMetadata | Planned |
| Define standard component type vocabulary | Planned |
| Validate entity references in components | Planned |
| Update JSON Schema | Planned |

**Standard Component Types:**

| Type | Description | Typical Roles |
|------|-------------|---------------|
| `idp` | Identity Provider | OAuth AS, OIDC, SCIM SP |
| `iga` | Identity Governance | SCIM Client, Audit |
| `agent_provider` | Agent registration/tokens | AAuth Agent Provider |
| `person_server` | Human consent | AAuth Person Server |
| `access_server` | Token issuance | AAuth Access Server, OAuth AS |
| `pdp` | Policy Decision Point | AuthZEN PDP, PAP |
| `gateway` | Access Gateway | AuthZEN PEP, OAuth RS |
| `mcp_client` | MCP Host/Client | MCP Client, AAuth Agent |
| `mcp_server` | MCP Tool Server | MCP Server |
| `resource_api` | Protected Resource | OAuth RS, SPIFFE Workload |
| `spire` | SPIFFE Infrastructure | SPIRE Server, Agent |

#### 8.3 Trust Relationships

Add explicit trust relationship modeling:

```go
// New type
type TrustRelationship struct {
    ID          string   `json:"id,omitempty"`
    From        string   `json:"from"`              // Entity or component ID
    To          string   `json:"to"`                // Entity or component ID
    Type        string   `json:"type"`              // Relationship type
    Credentials []string `json:"credentials,omitempty"` // What is exchanged
    Mutual      bool     `json:"mutual,omitempty"`  // Bidirectional (e.g., mTLS)
    Description string   `json:"description,omitempty"`
}

// Add to ProtocolMetadata
type ProtocolMetadata struct {
    // ... existing fields
    TrustRelations []TrustRelationship `json:"trust_relations,omitempty"`
}
```

| Task | Status |
|------|--------|
| Add `TrustRelationship` type | Planned |
| Add `trust_relations` field to ProtocolMetadata | Planned |
| Define standard relationship type vocabulary | Planned |
| Validate entity/component references | Planned |
| Update JSON Schema | Planned |

**Standard Relationship Types:**

| Type | Description | Example |
|------|-------------|---------|
| `authenticates` | Verifies identity | IdP authenticates User |
| `validates` | Verifies claims/tokens | PDP validates Agent token |
| `delegates` | Grants delegated authority | User delegates to Agent |
| `authorizes` | Grants access permission | PDP authorizes action |
| `issues` | Creates token/credential | Agent Provider issues JWT |
| `trusts` | Accepts tokens from | RS trusts IdP via JWKS |
| `provisions` | Creates/manages lifecycle | IGA provisions Agent in SCIM |
| `attests` | Cryptographically vouches | SPIRE attests Workload |

**Standard Credentials:**

| Credential | Description |
|------------|-------------|
| `x509_svid` | SPIFFE X.509 SVID |
| `jwt_svid` | SPIFFE JWT SVID |
| `jwt_assertion` | ID-JAG assertion |
| `access_token` | OAuth access token |
| `id_token` | OIDC ID token |
| `aa_agent_jwt` | AAuth agent token |
| `aa_auth_jwt` | AAuth auth token |
| `mtls` | Mutual TLS |
| `api_key` | API key |

#### 8.4 New Diagram Types

Add renderers for new diagram types:

| Diagram Type | Output | Description |
|--------------|--------|-------------|
| Component Diagram | SVG, Mermaid | Shows deployment components and their roles |
| Trust Diagram | SVG, Mermaid | Shows trust relationships between components |
| Role Matrix | Markdown, HTML | Protocol × Component role matrix |

| Task | Status |
|------|--------|
| SVG component diagram renderer | Planned |
| SVG trust relationship diagram renderer | Planned |
| Mermaid component diagram (flowchart) | Planned |
| Mermaid trust diagram (flowchart) | Planned |
| Markdown role matrix generator | Planned |
| CLI `--format=component`, `--format=trust`, `--format=matrix` | Planned |

#### 8.5 CLI Enhancements

| Task | Status |
|------|--------|
| `pidl generate --format=component` | Planned |
| `pidl generate --format=trust` | Planned |
| `pidl generate --format=matrix` | Planned |
| `pidl roles list` - list all roles in a file | Planned |
| `pidl components list` - list all components | Planned |
| `pidl trust list` - list trust relationships | Planned |

#### 8.6 Examples

| Task | Status |
|------|--------|
| `examples/oaiaf-roles.json` - OAIAF protocol roles | Planned |
| `examples/oaiaf-components.json` - Deployment topology | Planned |
| `examples/oaiaf-trust.json` - Trust relationships | Planned |
| Documentation for new features | Planned |

---

### Phase 9: Code Quality & Refactoring (v0.7.x)

**Target:** Improve code organization, eliminate duplication, and standardize patterns for maintainability.

#### 9.1 Split Large Files

| Task | File | Lines | Status |
|------|------|-------|--------|
| Split `pidl.go` into modules | pidl.go | 1,119 | Planned |
| Create `types.go` | - | - | Planned |
| Create `protocol_methods.go` | - | - | Planned |
| Create `entity_methods.go` | - | - | Planned |
| Create `validators.go` | - | - | Planned |
| Split `validate.go` into modules | validate.go | 917 | Planned |
| Create `validate_protocol.go` | - | - | Planned |
| Create `validate_entities.go` | - | - | Planned |
| Create `validate_flows.go` | - | - | Planned |
| Create `validate_metadata.go` | - | - | Planned |

#### 9.2 Consolidate Duplicate Validators

Eliminate duplicate validation functions between `validate.go` (private) and `pidl.go` (public):

| Task | Status |
|------|--------|
| Move all `IsValid*` functions to `validators.go` | Planned |
| Remove duplicate private validators from `validate.go` | Planned |
| Add missing public validators: `IsValidCategory`, `IsValidEntityType`, `IsValidFlowMode` | Planned |

#### 9.3 Standardize Return Types

| Task | Status |
|------|--------|
| Return empty slices instead of nil in `ComponentsByType()` | Planned |
| Return empty slices instead of nil in `EntitiesInComponent()` | Planned |
| Return empty slices instead of nil in `TrustRelationsFrom()` | Planned |
| Return empty slices instead of nil in `TrustRelationsTo()` | Planned |
| Audit all slice-returning methods for consistency | Planned |

#### 9.4 Extract Common Renderer Config

| Task | Status |
|------|--------|
| Create `RenderOptions` struct with shared Show* fields | Planned |
| Refactor `PlantUMLRenderer` to use `RenderOptions` | Planned |
| Refactor `MermaidRenderer` to use `RenderOptions` | Planned |
| Refactor `D2Renderer` to use `RenderOptions` | Planned |
| Refactor `SVGRenderer` to use `RenderOptions` | Planned |

#### 9.5 Consolidate Write Methods

| Task | Status |
|------|--------|
| Remove `WriteProtocolFile()` from operations.go | Planned |
| Keep only `Protocol.WriteFile()` in parse.go | Planned |
| Update CLI to use `Protocol.WriteFile()` | Planned |

#### 9.6 Remove/Refactor Unused Code

| Task | Status |
|------|--------|
| Remove or document `NewProtocol()` in operations.go | Planned |
| Move `SanitizeID()` to cmd/pidl or unexport | Planned |
| Move `TitleCase()` to cmd/pidl or unexport | Planned |

#### 9.7 Add Missing Tests

| Task | Status |
|------|--------|
| Add tests for `SanitizeID()` | Planned |
| Add tests for `TitleCase()` | Planned |
| Add tests for `SupportedFormats()` | Planned |
| Add negative tests for parse error cases | Planned |
| Add circular phase reference validation test | Planned |

---

### Phase 10: Protocol Composition (v0.8.0)

- [ ] Import mechanism for protocol modules
- [ ] Protocol inheritance/extension
- [ ] Reusable entity definitions
- [ ] Standard library of common protocols

### Phase 11: Runtime Execution Engine

- [ ] Protocol Execution Engine (PEX)
- [ ] Event queue and execution loop
- [ ] State store per entity
- [ ] Execution trace recording
- [ ] Step-by-step protocol simulation

### Phase 12: Analysis & Tooling

- [ ] Protocol comparison (diff two protocols)
- [ ] Execution trace visualization
- [ ] Interactive protocol debugger
- [ ] Attack surface analysis

### Phase 13: Integrations

- [ ] VS Code extension (syntax highlighting, preview)
- [ ] MkDocs plugin for embedding diagrams
- [ ] GitHub Action for CI validation
- [ ] Web playground for interactive editing

---

## Target Protocols

| Protocol | Category | Status |
|----------|----------|--------|
| OAuth 2.0 | Auth | ✅ Example |
| OAuth 2.0 + PKCE | Auth | ✅ Example |
| OpenID Connect | Auth | ✅ Example |
| MCP (Model Context Protocol) | Agent | ✅ Example |
| A2A (Agent-to-Agent) | Agent | ✅ Example |
| SAML 2.0 | Auth | Planned |
| WebAuthn/FIDO2 | Auth | Planned |
| SCIM | Provisioning | Planned |

---

## Non-Goals

- HTTP/transport-level details (use OpenAPI)
- Message schema definitions (use JSON Schema/protobuf)
- Code generation for protocol implementation
- Full formal verification (use TLA+)

---

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.
