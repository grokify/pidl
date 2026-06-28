# PIDL Roadmap

Protocol Interaction Description Language - a JSON-based DSL for describing protocol choreography that compiles to diagrams.

## Release History

| Version | Date | Highlights |
|---------|------|------------|
| v0.10.0 | 2026-06-28 | Protocol diff, trace visualization, debugger, security analysis |
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

### Phase 8: Protocol Roles & Deployment Architecture ✅ (v0.7.0)

**Target:** Enable PIDL to describe protocol roles, logical deployment components, and trust relationships for architecture documentation.

**Use Case:** OAIAF needs to document how protocols map to real-world deployments (IdP, IGA, Gateway, MCP Client, etc.) with explicit trust relationships.

#### 8.1 Data Model - Types & Validators ✅

| Task | Status |
|------|--------|
| Add `ProtocolRole` type with Protocol, Role, Variant fields | Done |
| Add `protocol_roles` field to Entity | Done |
| Add `DeploymentComponent` type | Done |
| Add `components` field to ProtocolMetadata | Done |
| Add `TrustRelationship` type | Done |
| Add `trust_relations` field to ProtocolMetadata | Done |
| Add `IsValidProtocol()` validator | Done |
| Add `IsValidComponentType()` validator | Done |
| Add `IsValidTrustRelationType()` validator | Done |
| Add `IsValidCredential()` validator | Done |

**ProtocolRole Type:**

```go
type ProtocolRole struct {
    Protocol string `json:"protocol"`          // "oauth", "scim", "aauth", "authzen", "mcp", "a2a", "spiffe"
    Role     string `json:"role"`              // "authorization_server", "client", "pep", "pdp", etc.
    Variant  string `json:"variant,omitempty"` // Sub-role: "person_server" vs "access_server"
}
```

**DeploymentComponent Type:**

```go
type DeploymentComponent struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    Type        string         `json:"type"`
    Description string         `json:"description,omitempty"`
    Entities    []string       `json:"entities"`
    Implements  []ProtocolRole `json:"implements"`
    Examples    []string       `json:"examples,omitempty"`
}
```

**TrustRelationship Type:**

```go
type TrustRelationship struct {
    ID          string   `json:"id,omitempty"`
    From        string   `json:"from"`
    To          string   `json:"to"`
    Type        string   `json:"type"`
    Credentials []string `json:"credentials,omitempty"`
    Mutual      bool     `json:"mutual,omitempty"`
    Description string   `json:"description,omitempty"`
}
```

#### 8.2 Standard Vocabularies

**Protocols:**

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

**Component Types:**

| Type | Description |
|------|-------------|
| `idp` | Identity Provider (OAuth AS, OIDC, SCIM SP) |
| `iga` | Identity Governance (SCIM Client, Audit) |
| `agent_provider` | Agent registration/tokens |
| `person_server` | Human consent |
| `access_server` | Token issuance |
| `pdp` | Policy Decision Point |
| `gateway` | Access Gateway (PEP, OAuth RS) |
| `mcp_client` | MCP Host/Client |
| `mcp_server` | MCP Tool Server |
| `resource_api` | Protected Resource |
| `spire` | SPIFFE Infrastructure |

**Trust Relationship Types:**

| Type | Description |
|------|-------------|
| `authenticates` | Verifies identity |
| `validates` | Verifies claims/tokens |
| `delegates` | Grants delegated authority |
| `authorizes` | Grants access permission |
| `issues` | Creates token/credential |
| `trusts` | Accepts tokens from |
| `provisions` | Creates/manages lifecycle |
| `attests` | Cryptographically vouches |

**Credentials:**

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

#### 8.3 Validation ✅

| Task | Status |
|------|--------|
| Validate protocol/role combinations | Done |
| Validate entity references in components | Done |
| Validate entity/component references in trust relations | Done |
| Validate credential types in trust relations | Done |
| Update JSON Schema | Done |

#### 8.4 Protocol Query Methods ✅

| Task | Status |
|------|--------|
| `Protocol.ComponentsByType(type)` | Done |
| `Protocol.EntitiesInComponent(id)` | Done |
| `Protocol.TrustRelationsFrom(id)` | Done |
| `Protocol.TrustRelationsTo(id)` | Done |
| `Protocol.EntitiesByRole(protocol, role)` | Done |
| `Protocol.EntitiesByProtocol(protocol)` | Done |
| `Protocol.EntitiesWithProtocolRoles()` | Done |
| `Protocol.ComponentByID(id)` | Done |
| `Protocol.ComponentForEntity(entityID)` | Done |
| `Protocol.TrustRelationByID(id)` | Done |
| `Protocol.TrustRelationsByType(type)` | Done |
| `Protocol.AllProtocols()` | Done |
| `Protocol.AllComponentTypes()` | Done |
| `Protocol.AllTrustRelationTypes()` | Done |

#### 8.5 New Diagram Types ✅

| Diagram Type | Output | Status |
|--------------|--------|--------|
| Component Diagram | Mermaid | ✅ Done |
| Trust Diagram | Mermaid | ✅ Done |
| Component Diagram | SVG | ✅ Done |
| Trust Diagram | SVG | ✅ Done |
| Role Matrix | Markdown | ✅ Done |

| Task | Status |
|------|--------|
| Mermaid component diagram (flowchart) | ✅ Done |
| Mermaid trust diagram (flowchart) | ✅ Done |
| SVG component diagram renderer | ✅ Done |
| SVG trust relationship diagram renderer | ✅ Done |
| Markdown role matrix generator | ✅ Done |

#### 8.6 CLI Enhancements ✅

| Task | Status |
|------|--------|
| `pidl generate --format=component` | ✅ Done |
| `pidl generate --format=trust` | ✅ Done |
| `pidl generate --format=matrix` | ✅ Done |
| `pidl roles <file>` | ✅ Done |
| `pidl components <file>` | ✅ Done |
| `pidl trust <file>` | ✅ Done |

#### 8.7 Examples & Documentation ✅

| Task | Status |
|------|--------|
| `examples/oaiaf_architecture.json` (combined roles, components, trust) | ✅ Done |
| Update docs/index.md with new features | Planned |

---

### Phase 9: Code Quality & Refactoring ✅ (v0.7.x)

**Target:** Improve code organization, eliminate duplication, and standardize patterns for maintainability.

- [x] Split `pidl.go` (1,119 lines) into 5 modules: types.go, protocol_methods.go, entity_methods.go, flow_methods.go, validators.go
- [x] Consolidate duplicate validators into `validators.go`
- [x] Standardize slice-returning methods to return empty slices instead of nil
- [x] Extract `SequenceRenderOptions` struct for renderer config sharing
- [x] Consolidate write methods into `Protocol.WriteFile()`
- [x] Remove unused `NewProtocol()` function

**Completed (deferred items):**

- [x] Split `validate.go` into 4 modules: validate.go (core), validate_entity.go, validate_flow.go, validate_metadata.go
- [x] Review `SanitizeID()` and `TitleCase()` - kept exported as useful public API (has comprehensive tests)
- [x] Verify utility function tests - TitleCase and SanitizeID already have tests in operations_test.go

---

### Phase 10: Protocol Composition ✅ (v0.8.0)

**Target:** Enable modular protocol design with imports and inheritance.

- [x] Import mechanism for protocol modules (`imports` field)
- [x] Protocol inheritance/extension (`extends` field)
- [x] Reusable entity definitions via imports with optional aliasing
- [x] Standard library of common protocols (`examples/stdlib/`)
- [x] Circular import/extends detection
- [x] CLI `resolve` command for debugging composition
- [x] Automatic resolution in `generate` command

**New Types:**

```go
type ProtocolExtends struct {
    Path            string   `json:"path"`
    ExcludeEntities []string `json:"exclude_entities,omitempty"`
    ExcludePhases   []string `json:"exclude_phases,omitempty"`
    ExcludeFlows    []int    `json:"exclude_flows,omitempty"`
}

type ProtocolImport struct {
    Path         string   `json:"path"`
    Alias        string   `json:"alias,omitempty"`
    Entities     []string `json:"entities,omitempty"`
    Phases       []string `json:"phases,omitempty"`
    IncludeFlows bool     `json:"include_flows,omitempty"`
}
```

**Standard Library:**

| File | Description |
|------|-------------|
| `stdlib/oauth_entities.json` | OAuth 2.0 base entities and flows |
| `stdlib/mcp_entities.json` | MCP base entities and flows |

**CLI Commands:**

| Command | Description |
|---------|-------------|
| `pidl resolve <file>` | Resolve imports/extends and output merged protocol |
| `pidl generate --resolve` | Explicitly request resolution before generating |

### Phase 11: Runtime Execution Engine ✅ (v0.9.0)

**Target:** Enable protocol simulation with state tracking and execution traces.

- [x] Protocol Execution Engine (Executor)
- [x] Event queue for future event-driven execution
- [x] State store per entity with initial state initialization
- [x] Execution trace recording with timestamps
- [x] Step-by-step protocol simulation
- [x] Conditional flow support with evaluator callbacks
- [x] State mutation validation (from state checks)
- [x] CLI `simulate` command with verbose and JSON output

**New Types:**

```go
type Executor struct {
    Protocol           *Protocol
    ConditionEvaluator func(ctx *ExecutionContext, flow *Flow) bool
}

type ExecutionContext struct {
    Protocol     *Protocol
    EntityStates map[string]string
    FlowIndex    int
    Trace        *ExecutionTrace
    Completed    bool
}

type ExecutionTrace struct {
    ProtocolID    string
    Steps         []ExecutionStep
    InitialStates map[string]string
    FinalStates   map[string]string
    Duration()    time.Duration
}
```

**CLI Commands:**

| Command | Description |
|---------|-------------|
| `pidl simulate <file>` | Run full protocol simulation |
| `pidl simulate -steps=N <file>` | Run N steps only |
| `pidl simulate -v <file>` | Verbose output with each step |
| `pidl simulate -json <file>` | Output trace as JSON |

### Phase 12: Analysis & Tooling ✅ (v0.10.0)

- [x] Protocol comparison (diff two protocols)
- [x] Execution trace visualization
- [x] Interactive protocol debugger
- [x] Attack surface analysis

**New Types & Commands:**

```go
// Protocol Comparison
type ProtocolDiff struct {
    Items   []DiffItem
    Summary DiffSummary
}

// Trace Visualization
type TraceRenderer struct {
    ShowStates, ShowTimings, HighlightSkipped bool
}

// Interactive Debugger
type DebugSession struct {
    Executor    *Executor
    Breakpoints map[int]*Breakpoint
}

// Security Analysis
type SecurityAnalysis struct {
    Risks   []SecurityRisk
    Summary SecuritySummary
}
```

**CLI Commands:**

| Command | Description |
|---------|-------------|
| `pidl diff <base> <new>` | Compare two protocols |
| `pidl simulate --trace-format=svg` | Render execution trace as SVG |
| `pidl debug <file>` | Interactive step-through debugger |
| `pidl analyze <file>` | Security analysis with risk identification |

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
