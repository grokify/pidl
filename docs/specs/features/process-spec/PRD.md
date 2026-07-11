# Process Spec Profile - Product Requirements Document

**Feature**: Process Spec Profile for PIDL
**Version**: 1.0
**Status**: Draft
**Author**: Generated from IDEATION_PROCESSSPEC.md analysis

## Executive Summary

Extend PIDL to support **Process Spec** - a profile for describing workflow and process semantics. This enables PIDL to model not just protocol choreography but also data pipelines, AI/LLM workflows, and multi-step processes with rich node semantics.

## Problem Statement

Current PIDL excels at describing "who talks to whom, in what order" for protocols like OAuth and MCP. However, modern systems require describing:

- **Data pipelines** - ETL/ELT with deterministic transformations
- **AI/LLM workflows** - Non-deterministic processing with model policies
- **Multi-step processes** - VisionSpec-style document generation flows
- **Hybrid processes** - Mixing deterministic validation with LLM generation

Users currently must use separate tools (BPMN, custom YAML) to describe these workflows, losing the benefits of PIDL's diagram generation, validation, and analysis capabilities.

## Goals

1. **G1**: Enable PIDL to describe process/workflow semantics without breaking existing protocol support
2. **G2**: Support differentiation between deterministic and non-deterministic (LLM) processing steps
3. **G3**: Generate the same diagram outputs (PlantUML, Mermaid, D2, SVG) from process specs
4. **G4**: Maintain backward compatibility with existing PIDL documents
5. **G5**: Enable security analysis and simulation for process flows

## Non-Goals

- Full BPMN compatibility (use dedicated BPMN tools for that)
- Code generation for workflow execution engines
- Runtime execution of processes (simulation only)
- Process mining or discovery from logs

## User Stories

### US1: AI Workflow Designer

> As an AI systems architect, I want to describe an LLM-powered document generation pipeline so that I can visualize the flow, identify non-deterministic steps, and generate documentation.

**Acceptance Criteria**:

- Can define nodes with `stepType: llm` vs `stepType: deterministic`
- Can specify inputs (files, objects, APIs) and outputs
- Generated diagrams show processing type visually
- Can run simulation to trace execution

### US2: Data Pipeline Engineer

> As a data engineer, I want to describe ETL pipelines in PIDL so that I can use the same tooling for both API protocols and data flows.

**Acceptance Criteria**:

- Can define source, transform, and sink nodes
- Can specify data lineage through flows
- Can annotate PII/sensitive data handling
- DOT/D2 output shows data flow direction

### US3: VisionSpec User

> As a VisionSpec user, I want to describe the MRD → PRD → TRD → Execution Spec pipeline so that I can visualize and validate the document generation process.

**Acceptance Criteria**:

- Can model the full VisionSpec document flow
- Can distinguish between deterministic parsing and LLM generation
- Can specify quality gates between steps
- Can generate both sequence and data flow diagrams

### US4: Security Analyst

> As a security analyst, I want to identify which process steps involve non-deterministic LLM processing so that I can focus review on those areas.

**Acceptance Criteria**:

- Security analysis flags LLM steps for review
- Can annotate guardrails and validation requirements
- Trust boundary analysis works with process nodes

## Requirements

### Functional Requirements

#### FR1: Profile Identification

- PIDL documents MUST support a `kind` field to identify the profile
- Valid values: `protocol` (default), `process`
- Existing documents without `kind` are treated as `protocol`

#### FR2: Process Node Semantics

Process entities MUST support additional fields:

| Field | Type | Description |
|-------|------|-------------|
| `step_type` | enum | `deterministic`, `llm`, `human`, `external`, `tool` |
| `inputs` | array | Input specifications |
| `outputs` | array | Output specifications |
| `processing` | object | Processing configuration |
| `failure_modes` | array | Possible failure scenarios |
| `retry_strategy` | object | Retry configuration |

#### FR3: Input/Output Specifications

Inputs and outputs MUST specify:

| Field | Type | Description |
|-------|------|-------------|
| `kind` | enum | `file`, `object`, `api`, `database`, `queue`, `stream` |
| `name` | string | Identifier for the input/output |
| `schema` | string | Optional reference to JSON Schema |
| `required` | bool | Whether this input is required |

#### FR4: Processing Configuration

The `processing` object MUST support:

| Field | Type | Description |
|-------|------|-------------|
| `engine` | string | Processing engine identifier |
| `determinism` | enum | `deterministic`, `non_deterministic` |
| `model_policy` | string | For LLM steps, the model selection policy |
| `timeout` | duration | Maximum processing time |
| `idempotent` | bool | Whether the step is idempotent |

#### FR5: Diagram Generation

- Sequence diagrams MUST render process flows like protocol flows
- Data flow diagrams MUST show inputs/outputs on edges
- Process nodes SHOULD be visually distinguished by `step_type`
- SVG output SHOULD use different colors/icons for deterministic vs LLM

#### FR6: Validation

- Process specs MUST validate that all inputs are produced by prior steps or defined as external
- Circular dependencies MUST be detected and rejected
- Required inputs MUST have producing flows

#### FR7: Simulation

- Simulator MUST track input/output availability
- Simulator MUST flag steps that cannot execute due to missing inputs
- Trace output MUST include step type and processing metadata

### Non-Functional Requirements

#### NFR1: Backward Compatibility

- Existing PIDL documents MUST parse and render without modification
- Default `kind` is `protocol` for documents without explicit kind

#### NFR2: Schema Evolution

- New fields MUST be optional with sensible defaults
- Schema version MUST be incremented

#### NFR3: Performance

- Validation and rendering performance MUST not regress for protocol specs
- Process spec validation MAY have additional O(n) passes for dependency analysis

## Success Metrics

| Metric | Target |
|--------|--------|
| Existing tests pass | 100% |
| New process spec examples | 3+ (VisionSpec, ETL, LLM pipeline) |
| Diagram formats supporting process | All existing formats |
| Documentation coverage | Full spec + examples |

## Dependencies

- PIDL v0.4.0 (current)
- No external dependencies required

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Schema complexity explosion | Medium | High | Start minimal, add fields incrementally |
| Breaking existing users | Low | High | Strict backward compatibility, default `kind` |
| Diagram rendering confusion | Medium | Medium | Clear visual distinction for step types |

## Timeline

See ROADMAP.md for phased delivery plan.

## References

- [IDEATION_PROCESSSPEC.md](../../../../IDEATION_PROCESSSPEC.md) - Original ideation conversation
- [PIDL Specification](../../../SPECIFICATION.md) - Current PIDL spec
- [VisionSpec](https://github.com/grokify/visionspec) - Target use case
