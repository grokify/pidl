# PIDL Roadmap

**Protocol Interaction Description Language** - A JSON-based DSL for describing protocol choreography that compiles to diagrams.

**Current Version**: v0.5.0
**Status**: Active Development

---

## Release Summary

| Version | Milestone | Status |
|---------|-----------|--------|
| v0.1.0 | Core DSL & Diagram Generation | ✅ Done |
| v0.2.0 | Enhanced Flow Semantics | ✅ Done |
| v0.3.0 | SVG Rendering & Network Boundaries | ✅ Done |
| v0.4.0 | State, Security, Roles, Composition, Execution, Analysis | ✅ Done |
| v0.5.0 | Process Spec Foundation | ✅ Done |
| v0.6.0 | Process Spec Rendering | ✅ Done |
| v0.6.1 | Process Spec Examples | ✅ Done |
| v0.7.0 | Process Spec Analysis | ✅ Done |
| v0.8.0 | Infographic Renderer & Documentation | ✅ Done |
| v0.5.0+ | Data Lineage, Parallel, Cost, Workflow Exports, Integrations | ✅ Done |
| v1.0.0 | Production Ready Release | 🔲 Next |

---

## Completed: Core PIDL (v0.1.0 - v0.4.0)

### v0.1.0: Core DSL & Diagram Generation ✅

- [x] Core PIDL JSON schema with entities, flows, phases
- [x] Entity types for OAuth/OIDC, MCP, A2A protocols
- [x] Flow modes: request, response, redirect, callback, interactive, event
- [x] JSON Schema for validation
- [x] PlantUML, Mermaid, Graphviz DOT renderers
- [x] CLI tool: validate, generate, examples, init
- [x] Go library with parser, model types, validation

### v0.2.0: Enhanced Flow Semantics ✅

- [x] Conditional flows (`condition` field)
- [x] Alternative paths (`alternatives` field)
- [x] Flow annotations with typed notes
- [x] Nested phases with parent hierarchy
- [x] D2 diagram renderer (sequence, flow, arch styles)

### v0.3.0: SVG Rendering & Network Boundaries ✅

- [x] Static and animated SVG sequence diagrams
- [x] Per-message animation control with semantic presets
- [x] Light/dark theme support with template system
- [x] Network boundary renderer with trust zone inference
- [x] Boundary styles: trusted, dmz, external, cloud

### v0.4.0: Advanced Features ✅

#### State Model

- [x] Entity state definitions with `states` field
- [x] State mutations on flows with `sets` clause
- [x] Mermaid `stateDiagram-v2` output format

#### Security & Trust

- [x] Trust levels: trusted, semi_trusted, untrusted, authoritative
- [x] Security requirements: token, signature, encryption, mtls, mac
- [x] Token definitions with issuer, audience, binding
- [x] Security badges in SVG and notes in PlantUML/Mermaid

#### Protocol Roles & Components

- [x] `ProtocolRole` type with protocol/role/variant
- [x] `DeploymentComponent` type with entities and implements
- [x] `TrustRelationship` type with credentials
- [x] Component and trust diagram renderers (Mermaid, SVG)
- [x] Markdown role matrix generator

#### Protocol Composition

- [x] Import mechanism for protocol modules (`imports` field)
- [x] Protocol inheritance/extension (`extends` field)
- [x] Standard library: `stdlib/oauth_entities.json`, `stdlib/mcp_entities.json`
- [x] CLI `resolve` command

#### Execution Engine

- [x] Protocol Executor with state tracking
- [x] Execution trace recording with timestamps
- [x] Conditional flow evaluation
- [x] CLI `simulate` command with verbose and JSON output

#### Analysis & Tooling

- [x] Protocol comparison (`pidl diff`)
- [x] Execution trace visualization
- [x] Interactive debugger (`pidl debug`)
- [x] Security analysis (`pidl analyze`)

---

## Completed: Process Spec Profile (v0.5.0 - v0.6.0)

### v0.5.0: Process Spec Foundation ✅

**Goal**: Extend PIDL to describe process/workflow specifications with step types and data ports.

#### New Types

| Type | Description |
|------|-------------|
| `ProtocolKind` | `protocol` vs `process` enum |
| `StepType` | `deterministic`, `llm`, `human`, `external`, `tool` |
| `DataPort` | Input/output with kind, schema, required, sensitive |
| `DataPortKind` | `file`, `object`, `api`, `database`, `queue`, `stream` |
| `ProcessingConfig` | Engine, determinism, timeout, caching |
| `FailureMode` | Failure scenarios with severity and recovery |
| `RetryStrategy` | Retry configuration with backoff |

#### Entity Extensions

- [x] `step_type` field for process step classification
- [x] `inputs` and `outputs` ([]DataPort) for data flow
- [x] `processing` configuration for step execution
- [x] `failure_modes` and `retry_strategy` for resilience

#### Validation

- [x] Step type validation for process specs
- [x] Data port kind validation
- [x] Duplicate port name detection
- [x] Processing config validation

#### Helper Methods

- [x] `Protocol.IsProcessSpec()`, `Protocol.Kind()`
- [x] `Protocol.ProcessSteps()`, `LLMSteps()`, `DeterministicSteps()`
- [x] `Entity.IsProcessStep()`, `IsLLMStep()`, `IsHumanStep()`, etc.
- [x] `Entity.RequiredInputs()`, `SensitiveInputs()`, `SensitiveOutputs()`

---

### v0.6.0: Process Spec Rendering ✅

**Goal**: Visual differentiation for process steps and data ports in diagrams.

#### Step Type Styling

| Step Type | Color | Icon | Use |
|-----------|-------|------|-----|
| `deterministic` | Blue (#1976D2) | ⚙️ | Predictable processing |
| `llm` | Purple (#7B1FA2) | 🧠 | AI/ML processing |
| `human` | Green (#388E3C) | 👤 | Human involvement |
| `external` | Orange (#F57C00) | ☁️ | External services |
| `tool` | Gray (#607D8B) | 🔧 | Tool invocations |

#### Renderer Updates

- [x] PlantUML: stereotypes (`<<llm>>`) with skinparam colors
- [x] Mermaid: emoji badges in participant names
- [x] D2: fill/stroke colors for step types
- [x] SVG: inline styles and emoji badges

#### Data Port Visualization

| Port Kind | Shape | Icon | Color |
|-----------|-------|------|-------|
| `file` | page | 📄 | Amber (#FFF8E1) |
| `object` | rectangle | 📦 | Blue (#E3F2FD) |
| `api` | cloud | 🌐 | Green (#E8F5E9) |
| `database` | cylinder | 🗄️ | Pink (#FCE4EC) |
| `queue` | queue | 📬 | Purple (#F3E5F5) |
| `stream` | parallelogram | 🌊 | Cyan (#E0F7FA) |

- [x] D2Flow: data ports as connected nodes with shapes
- [x] D2Arch: data ports grouped by kind (Files, Databases, etc.)
- [x] `ShowDataPorts` option in `SequenceRenderOptions`

---

## Upcoming: Process Spec Continuation

### v0.6.1: Process Spec Examples ✅

**Goal**: Provide reference process specifications.

| Example | Description | Status |
|---------|-------------|--------|
| `visionspec_execution.json` | MRD → PRD → TRD pipeline | ✅ Done |
| `etl_pipeline.json` | Extract-Transform-Load | ✅ Done |
| `llm_document_review.json` | LLM with human approval | ✅ Done |
| `ci_cd_pipeline.json` | Build-test-deploy | ✅ Done |

**Acceptance Criteria**:

- [x] All examples validate successfully
- [x] All examples render to all formats
- [x] Examples discoverable via `pidl examples`
- [x] Examples embedded in binary

---

### v0.7.0: Process Spec Analysis ✅

**Goal**: Process-aware security analysis and simulation enhancements.

#### Security Rules ✅

| Rule | Severity | Description | Status |
|------|----------|-------------|--------|
| SEC011 | Medium | LLM step without downstream validation | ✅ Done |
| SEC012 | High | Sensitive data flows to LLM step | ✅ Done |
| SEC013 | Medium | Non-deterministic step in critical path | ✅ Done |
| SEC014 | Low | External step without failure modes | ✅ Done |
| SEC015 | Medium | Human step without timeout | ✅ Done |

- [x] Added `CategoryProcessSecurity` risk category
- [x] Process spec rules skip non-process protocols
- [x] Full test coverage for all rules

#### Simulation Enhancements ✅

- [x] Input availability tracking (`AvailableInputs` map in `ProcessExecutionContext`)
- [x] Output production tracking (`ProducedOutputs` map with data flow propagation)
- [x] Dependency analysis (`GetBlockedSteps`, `AnalyzeDependencies`)
- [x] Topological execution ordering (`TopologicalSort` using Kahn's algorithm)

New types and methods:

- `ProcessExecutor` - Process-aware executor with data flow tracking
- `ProcessExecutionContext` - Extended context with input/output tracking
- `DependencyAnalysis` - Detailed analysis of step dependencies
- `ExecutionReadiness` - Summary of execution readiness
- `StepExecutionStatus` - Step status enum (pending, ready, blocked, etc.)

---

### v0.8.0: Infographic Renderer & Documentation ✅

**Goal**: LinkedIn-friendly infographic renderer and complete documentation.

#### Infographic Renderer

Visual infographics for social media and datasheets with animated data flow.

| Feature | Description | Status |
|---------|-------------|--------|
| Compact SVG renderer | Simplified layout for small images | ✅ Done |
| Size presets | linkedin-square, linkedin-portrait, datasheet-tile, etc. | ✅ Done |
| Theme support | bold, minimal, dark, tech themes | ✅ Done |
| Animated dots | SVG animateMotion along edges | ✅ Done |
| Step type icons | ⚙️🧠👤☁️🔧 per step type | ✅ Done |
| CLI integration | `pidl generate -f infographic` | ✅ Done |
| Regular protocol support | Handle non-process protocols gracefully | ✅ Done |
| Entity type icons | 🤖💻👤🔐🗄️🖥️ per entity type | ✅ Done |
| Entity type colors | Color coding per entity type | ✅ Done |
| Custom node shapes | Per-entity-type shapes | ✅ Done |

**Preset Sizes**:

| Preset | Dimensions | Use Case |
|--------|------------|----------|
| `linkedin-square` | 1200×1200 | LinkedIn feed posts |
| `linkedin-portrait` | 1080×1350 | LinkedIn portrait posts |
| `linkedin-landscape` | 1200×627 | LinkedIn link previews |
| `datasheet-tile` | 400×400 | Datasheet grid layouts |
| `datasheet-wide` | 600×300 | Wide datasheet tiles |

**CLI Options (planned)**:

```bash
pidl generate -f infographic --size=linkedin-square --theme=tech --title="OAuth Flow" example.json
pidl generate -f infographic --size=datasheet-tile --no-animate --theme=minimal example.json
```

#### Documentation ✅

| Document | Updates | Status |
|----------|---------|--------|
| `SPECIFICATION.md` | Process profile section | ✅ Done |
| `README.md` | Process spec quick start | ✅ Done |
| `docs/guide/process.md` | When to use process vs protocol | ✅ Done |
| `mkdocs.yml` | Navigation updated | ✅ Done |

---

## Completed: Advanced Features (v0.5.0)

### Integrations ✅

- [x] VS Code extension (syntax highlighting, preview, validation, export)
- [x] MkDocs plugin for embedding PIDL diagrams
- [x] GitHub Action for CI validation
- [ ] Web playground for interactive editing

### Advanced Data Flow ✅

- [x] Data lineage tracking through flows
- [ ] JSON Schema references for port validation
- [ ] Data quality gates between steps
- [x] Parallel execution modeling (fork/join, race, scatter/gather)

### Workflow Engine Exports ✅

- [x] Temporal workflow export (Go)
- [x] Prefect flow export (Python)
- [x] BPMN 2.0 export (XML)
- [x] AWS Step Functions export (JSON)
- [ ] Custom runner plugin system

### Cost Tracking ✅

- [x] Cost model types (fixed, token_based, time_based, api_call, hybrid)
- [x] LLM cost presets (GPT-4, Claude-3)
- [x] Process cost analysis
- [x] Execution cost calculation

---

## Future Milestones

### v1.0.0: Production Ready

- [ ] Web playground for interactive editing
- [ ] JSON Schema references for port validation
- [ ] Data quality gates between steps
- [ ] Custom runner plugin system

### v1.1.0: Observability

- [ ] Metrics definitions (SLIs for steps)
- [ ] OpenTelemetry trace integration
- [ ] Latency budget definitions

---

## Target Protocols

| Protocol | Category | Status |
|----------|----------|--------|
| OAuth 2.0 | Auth | ✅ Example |
| OAuth 2.0 + PKCE | Auth | ✅ Example |
| OpenID Connect | Auth | ✅ Example |
| MCP (Model Context Protocol) | Agent | ✅ Example |
| A2A (Agent-to-Agent) | Agent | ✅ Example |
| OAIAF Architecture | Multi-Protocol | ✅ Example |
| SAML 2.0 | Auth | ✅ Example |
| WebAuthn/FIDO2 | Auth | ✅ Example |
| SCIM | Provisioning | ✅ Example |

---

## Non-Goals

The following are explicitly out of scope for PIDL:

- **HTTP/transport-level details** - Use OpenAPI for that
- **Message schema definitions** - Use JSON Schema or protobuf
- **Code generation for implementation** - PIDL describes, not implements
- **Full formal verification** - Use TLA+ for rigorous proofs

---

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines on contributing to PIDL development.
