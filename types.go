// Package pidl provides types and utilities for the Protocol Interaction Description Language.
// PIDL is a JSON-based DSL for describing protocol choreography that compiles to diagrams.
package pidl

// Protocol represents a complete PIDL document describing a protocol's choreography.
type Protocol struct {
	// ProtocolMeta contains metadata about the protocol.
	ProtocolMeta ProtocolMeta `json:"protocol"`

	// Extends specifies a base protocol to extend.
	Extends *ProtocolExtends `json:"extends,omitempty"`

	// Imports specifies other protocol files to import from.
	Imports []ProtocolImport `json:"imports,omitempty"`

	// Entities are the participants in the protocol (systems, actors, services).
	Entities []Entity `json:"entities"`

	// Phases provide optional logical grouping of flows.
	Phases []Phase `json:"phases,omitempty"`

	// Flows are the interactions between entities.
	Flows []Flow `json:"flows"`

	// Metadata contains additional protocol-level configuration.
	Metadata *ProtocolMetadata `json:"metadata,omitempty"`

	// resolved indicates whether imports/extends have been resolved.
	resolved bool
}

// ProtocolExtends specifies a base protocol to extend.
type ProtocolExtends struct {
	// Path is the file path to the base protocol (relative to current file).
	Path string `json:"path"`

	// ExcludeEntities lists entity IDs from the base to exclude.
	ExcludeEntities []string `json:"exclude_entities,omitempty"`

	// ExcludePhases lists phase IDs from the base to exclude.
	ExcludePhases []string `json:"exclude_phases,omitempty"`

	// ExcludeFlows lists flow indices from the base to exclude.
	ExcludeFlows []int `json:"exclude_flows,omitempty"`
}

// ProtocolImport specifies another protocol file to import from.
type ProtocolImport struct {
	// Path is the file path to import (relative to current file).
	Path string `json:"path"`

	// Alias prefixes imported IDs to avoid collisions (e.g., "oauth_").
	Alias string `json:"alias,omitempty"`

	// Entities lists specific entity IDs to import (empty = all).
	Entities []string `json:"entities,omitempty"`

	// Phases lists specific phase IDs to import (empty = none unless entities need them).
	Phases []string `json:"phases,omitempty"`

	// IncludeFlows imports flows between the specified entities.
	IncludeFlows bool `json:"include_flows,omitempty"`
}

// ProtocolMetadata contains additional protocol-level configuration.
type ProtocolMetadata struct {
	// Networks defines network boundary configurations.
	Networks map[string]*NetworkConfig `json:"networks,omitempty"`
	// NetworkLayout configures network diagram layout.
	NetworkLayout *NetworkLayoutConfig `json:"network_layout,omitempty"`
	// Tokens defines token types used in the protocol.
	Tokens []TokenDefinition `json:"tokens,omitempty"`
	// Components defines logical deployment components.
	Components []DeploymentComponent `json:"components,omitempty"`
	// TrustRelations defines trust relationships between entities or components.
	TrustRelations []TrustRelationship `json:"trust_relations,omitempty"`
}

// TokenDefinition describes a token type used in the protocol.
type TokenDefinition struct {
	// ID is the unique identifier for this token definition.
	ID string `json:"id"`
	// Name is the human-readable display name.
	Name string `json:"name,omitempty"`
	// Type is the token type: jwt, opaque, saml, api_key.
	Type string `json:"type,omitempty"`
	// Issuer is the entity ID that issues this token.
	Issuer string `json:"issuer,omitempty"`
	// Audience is the entity ID that consumes this token.
	Audience string `json:"audience,omitempty"`
	// Binding is the token binding method: bearer, mtls, dpop.
	Binding string `json:"binding,omitempty"`
	// Description provides additional context.
	Description string `json:"description,omitempty"`
}

// NetworkConfig defines a network boundary.
type NetworkConfig struct {
	// Name is the display name for the boundary.
	Name string `json:"name,omitempty"`
	// Style is the visual style: trusted, dmz, external, cloud.
	Style string `json:"style,omitempty"`
	// Entities explicitly lists entity IDs in this boundary.
	Entities []string `json:"entities,omitempty"`
	// Description provides tooltip/hover text.
	Description string `json:"description,omitempty"`
	// Color overrides the default boundary color.
	Color string `json:"color,omitempty"`
}

// NetworkLayoutConfig configures network diagram layout.
type NetworkLayoutConfig struct {
	// Direction is the layout direction: horizontal or vertical.
	Direction string `json:"direction,omitempty"`
	// Order specifies the boundary ordering.
	Order []string `json:"order,omitempty"`
}

// ProtocolMeta contains metadata about a protocol.
type ProtocolMeta struct {
	// ID is the unique identifier for the protocol.
	ID string `json:"id"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Kind identifies the PIDL profile type (protocol or process).
	Kind ProtocolKind `json:"kind,omitempty"`

	// Version of this protocol description.
	Version string `json:"version,omitempty"`

	// Description provides a brief summary.
	Description string `json:"description,omitempty"`

	// Category classifies the protocol type.
	Category Category `json:"category,omitempty"`

	// References links to relevant specifications.
	References []Reference `json:"references,omitempty"`
}

// Reference links to external documentation.
type Reference struct {
	// Name of the reference (e.g., "RFC 6749").
	Name string `json:"name"`

	// URL to the reference.
	URL string `json:"url"`
}

// Category represents the protocol category.
type Category string

const (
	CategoryAuth         Category = "auth"
	CategoryAgent        Category = "agent"
	CategoryMessaging    Category = "messaging"
	CategoryProvisioning Category = "provisioning"
	CategoryOther        Category = "other"
)

// ProtocolKind identifies the PIDL profile type.
type ProtocolKind string

const (
	// ProtocolKindProtocol is the default for protocol choreography.
	ProtocolKindProtocol ProtocolKind = "protocol"
	// ProtocolKindProcess is for workflow/process specifications.
	ProtocolKindProcess ProtocolKind = "process"
)

// TrustLevel represents the trust classification of an entity.
type TrustLevel string

const (
	// TrustLevelTrusted is for fully trusted internal systems.
	TrustLevelTrusted TrustLevel = "trusted"
	// TrustLevelSemiTrusted is for partially trusted systems (DMZ, partners).
	TrustLevelSemiTrusted TrustLevel = "semi_trusted"
	// TrustLevelUntrusted is for external/public systems.
	TrustLevelUntrusted TrustLevel = "untrusted"
	// TrustLevelAuthoritative is for sources of truth (IdPs, CAs).
	TrustLevelAuthoritative TrustLevel = "authoritative"
)

// StepType classifies the processing nature of a process step.
type StepType string

const (
	// StepTypeDeterministic is for repeatable, predictable processing.
	StepTypeDeterministic StepType = "deterministic"
	// StepTypeLLM is for LLM/AI-powered non-deterministic processing.
	StepTypeLLM StepType = "llm"
	// StepTypeHuman is for human-in-the-loop steps.
	StepTypeHuman StepType = "human"
	// StepTypeExternal is for external API/service calls.
	StepTypeExternal StepType = "external"
	// StepTypeTool is for tool invocations (MCP-style).
	StepTypeTool StepType = "tool"
	// StepTypeParallel is for parallel execution blocks.
	StepTypeParallel StepType = "parallel"
	// StepTypeConditional is for conditional branching.
	StepTypeConditional StepType = "conditional"
)

// ParallelConfig configures parallel execution for a step or flow.
type ParallelConfig struct {
	// Mode specifies the parallel execution mode.
	Mode ParallelMode `json:"mode"`
	// Branches lists the parallel branches (entity IDs or flow groups).
	Branches []ParallelBranch `json:"branches,omitempty"`
	// JoinCondition specifies when parallel execution is considered complete.
	JoinCondition JoinCondition `json:"join_condition,omitempty"`
	// MaxConcurrency limits the number of concurrent executions.
	MaxConcurrency int `json:"max_concurrency,omitempty"`
	// Timeout for the entire parallel block.
	Timeout string `json:"timeout,omitempty"`
}

// ParallelMode specifies how parallel execution is structured.
type ParallelMode string

const (
	// ParallelModeForkJoin executes branches in parallel and waits for all.
	ParallelModeForkJoin ParallelMode = "fork_join"
	// ParallelModeRace executes branches in parallel, first to complete wins.
	ParallelModeRace ParallelMode = "race"
	// ParallelModeScatter distributes work across multiple instances.
	ParallelModeScatter ParallelMode = "scatter"
	// ParallelModeGather collects results from parallel branches.
	ParallelModeGather ParallelMode = "gather"
)

// ParallelBranch represents a branch in parallel execution.
type ParallelBranch struct {
	// ID is the unique identifier for this branch.
	ID string `json:"id"`
	// Name is the human-readable name.
	Name string `json:"name,omitempty"`
	// EntityID is the target entity for this branch (optional).
	EntityID string `json:"entity_id,omitempty"`
	// FlowRefs lists flow indices that belong to this branch.
	FlowRefs []int `json:"flow_refs,omitempty"`
	// Condition specifies when this branch is executed.
	Condition string `json:"condition,omitempty"`
	// Weight for load balancing in scatter mode.
	Weight float64 `json:"weight,omitempty"`
}

// JoinCondition specifies when parallel execution is considered complete.
type JoinCondition string

const (
	// JoinConditionAll waits for all branches to complete.
	JoinConditionAll JoinCondition = "all"
	// JoinConditionAny completes when any branch completes.
	JoinConditionAny JoinCondition = "any"
	// JoinConditionMajority completes when majority of branches complete.
	JoinConditionMajority JoinCondition = "majority"
	// JoinConditionQuorum completes when a quorum of branches complete.
	JoinConditionQuorum JoinCondition = "quorum"
)

// DataPortKind classifies the type of data port.
type DataPortKind string

const (
	// DataPortKindFile represents file-based data.
	DataPortKindFile DataPortKind = "file"
	// DataPortKindObject represents in-memory object data.
	DataPortKindObject DataPortKind = "object"
	// DataPortKindAPI represents API endpoint data.
	DataPortKindAPI DataPortKind = "api"
	// DataPortKindDatabase represents database data.
	DataPortKindDatabase DataPortKind = "database"
	// DataPortKindQueue represents message queue data.
	DataPortKindQueue DataPortKind = "queue"
	// DataPortKindStream represents streaming data.
	DataPortKindStream DataPortKind = "stream"
)

// DataPort represents an input or output of a process step.
type DataPort struct {
	// Kind classifies the data type.
	Kind DataPortKind `json:"kind"`

	// Name is the identifier for this port.
	Name string `json:"name"`

	// Description provides additional context.
	Description string `json:"description,omitempty"`

	// Schema is an optional reference to a JSON Schema.
	Schema string `json:"schema,omitempty"`

	// Required indicates if this input must be provided.
	Required bool `json:"required,omitempty"`

	// Sensitive marks data as containing PII or secrets.
	Sensitive bool `json:"sensitive,omitempty"`
}

// Determinism classifies processing predictability.
type Determinism string

const (
	// DeterminismDeterministic indicates repeatable, predictable results.
	DeterminismDeterministic Determinism = "deterministic"
	// DeterminismNonDeterministic indicates variable results (e.g., LLM output).
	DeterminismNonDeterministic Determinism = "non_deterministic"
)

// ProcessingConfig describes how a step processes its inputs.
type ProcessingConfig struct {
	// Engine identifies the processing engine.
	Engine string `json:"engine,omitempty"`

	// Determinism indicates processing predictability.
	Determinism Determinism `json:"determinism,omitempty"`

	// ModelPolicy specifies model selection for LLM steps.
	ModelPolicy string `json:"model_policy,omitempty"`

	// Timeout is the maximum processing duration.
	Timeout string `json:"timeout,omitempty"`

	// Idempotent indicates if repeated execution is safe.
	Idempotent bool `json:"idempotent,omitempty"`

	// Cacheable indicates if outputs can be cached.
	Cacheable bool `json:"cacheable,omitempty"`

	// CacheTTL is the cache time-to-live if cacheable.
	CacheTTL string `json:"cache_ttl,omitempty"`

	// CostModel specifies the cost tracking model for this step.
	CostModel *CostModel `json:"cost_model,omitempty"`

	// LatencyBudget specifies timing SLIs for this step.
	LatencyBudget *LatencyBudget `json:"latency_budget,omitempty"`
}

// CostModel defines the cost characteristics of a processing step.
type CostModel struct {
	// Type classifies the cost model.
	Type CostModelType `json:"type"`

	// FixedCost is the base cost per execution (in the cost unit).
	FixedCost float64 `json:"fixed_cost,omitempty"`

	// VariableCost is the cost per unit of work.
	VariableCost float64 `json:"variable_cost,omitempty"`

	// CostUnit is the currency or unit for costs (e.g., "USD", "tokens", "credits").
	CostUnit string `json:"cost_unit,omitempty"`

	// TokenCosts for LLM steps (per 1K tokens).
	InputTokenCost  float64 `json:"input_token_cost,omitempty"`
	OutputTokenCost float64 `json:"output_token_cost,omitempty"`

	// EstimatedInputTokens for LLM cost estimation.
	EstimatedInputTokens int `json:"estimated_input_tokens,omitempty"`
	// EstimatedOutputTokens for LLM cost estimation.
	EstimatedOutputTokens int `json:"estimated_output_tokens,omitempty"`

	// ComputeCostPerSecond for compute-based steps.
	ComputeCostPerSecond float64 `json:"compute_cost_per_second,omitempty"`

	// APICallCost for external API steps.
	APICallCost float64 `json:"api_call_cost,omitempty"`
}

// CostModelType classifies how costs are calculated.
type CostModelType string

const (
	// CostModelTypeFixed has a fixed cost per execution.
	CostModelTypeFixed CostModelType = "fixed"
	// CostModelTypeTokenBased costs based on input/output tokens.
	CostModelTypeTokenBased CostModelType = "token_based"
	// CostModelTypeTimeBased costs based on execution time.
	CostModelTypeTimeBased CostModelType = "time_based"
	// CostModelTypeAPICall costs per API call.
	CostModelTypeAPICall CostModelType = "api_call"
	// CostModelTypeHybrid combines multiple cost factors.
	CostModelTypeHybrid CostModelType = "hybrid"
)

// LatencyBudget defines timing SLIs (Service Level Indicators) for a step.
type LatencyBudget struct {
	// P50 is the 50th percentile latency target.
	P50 string `json:"p50,omitempty"`
	// P95 is the 95th percentile latency target.
	P95 string `json:"p95,omitempty"`
	// P99 is the 99th percentile latency target.
	P99 string `json:"p99,omitempty"`
	// Max is the maximum acceptable latency (hard limit).
	Max string `json:"max,omitempty"`
	// ExpectedLatency is the expected typical latency.
	ExpectedLatency string `json:"expected_latency,omitempty"`
	// Critical marks this step as on the critical path.
	Critical bool `json:"critical,omitempty"`
	// VarianceClass indicates latency variability (low, medium, high).
	VarianceClass LatencyVarianceClass `json:"variance_class,omitempty"`
}

// LatencyVarianceClass classifies the variability of step latency.
type LatencyVarianceClass string

const (
	// LatencyVarianceLow indicates consistent, predictable latency.
	LatencyVarianceLow LatencyVarianceClass = "low"
	// LatencyVarianceMedium indicates moderate latency variance.
	LatencyVarianceMedium LatencyVarianceClass = "medium"
	// LatencyVarianceHigh indicates high variance (e.g., LLM, human tasks).
	LatencyVarianceHigh LatencyVarianceClass = "high"
)

// FailureMode describes a possible failure scenario.
type FailureMode struct {
	// ID is the unique identifier for this failure mode.
	ID string `json:"id"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Description explains the failure scenario.
	Description string `json:"description,omitempty"`

	// Severity indicates the impact level.
	Severity string `json:"severity,omitempty"`

	// Recovery describes how to handle this failure.
	Recovery string `json:"recovery,omitempty"`
}

// RetryStrategy configures retry behavior for a step.
type RetryStrategy struct {
	// MaxAttempts is the maximum number of retry attempts.
	MaxAttempts int `json:"max_attempts,omitempty"`

	// InitialDelay is the delay before the first retry.
	InitialDelay string `json:"initial_delay,omitempty"`

	// MaxDelay is the maximum delay between retries.
	MaxDelay string `json:"max_delay,omitempty"`

	// BackoffMultiplier increases delay between retries.
	BackoffMultiplier float64 `json:"backoff_multiplier,omitempty"`

	// RetryOn lists failure mode IDs that trigger retry.
	RetryOn []string `json:"retry_on,omitempty"`
}

// ProtocolRole defines a protocol-specific role that an entity implements.
type ProtocolRole struct {
	// Protocol is the protocol identifier (e.g., "oauth", "scim", "aauth", "authzen", "mcp", "a2a", "spiffe").
	Protocol string `json:"protocol"`
	// Role is the role within that protocol (e.g., "authorization_server", "client", "pep", "pdp").
	Role string `json:"role"`
	// Variant is an optional sub-role or variant (e.g., "person_server" vs "access_server" for AAuth).
	Variant string `json:"variant,omitempty"`
	// Description provides additional context.
	Description string `json:"description,omitempty"`
}

// Protocol identifier constants.
const (
	ProtocolOAuth    = "oauth"
	ProtocolSCIM     = "scim"
	ProtocolSPIFFE   = "spiffe"
	ProtocolAAuth    = "aauth"
	ProtocolIDJAG    = "idjag"
	ProtocolAuthZEN  = "authzen"
	ProtocolMCP      = "mcp"
	ProtocolA2A      = "a2a"
	ProtocolSAML     = "saml"
	ProtocolWebAuthn = "webauthn"
	ProtocolFIDO2    = "fido2"
	ProtocolOIDC     = "oidc"
)

// DeploymentComponent defines a logical deployment component that groups entities.
type DeploymentComponent struct {
	// ID is the unique identifier for this component.
	ID string `json:"id"`
	// Name is the human-readable display name.
	Name string `json:"name"`
	// Type classifies the component (e.g., "idp", "iga", "gateway", "mcp_client").
	Type string `json:"type"`
	// Description provides additional context.
	Description string `json:"description,omitempty"`
	// Entities lists the entity IDs contained in this component.
	Entities []string `json:"entities,omitempty"`
	// Implements lists the protocol roles this component implements.
	Implements []ProtocolRole `json:"implements,omitempty"`
	// Examples lists real-world products that represent this component type.
	Examples []string `json:"examples,omitempty"`
}

// Component type constants.
const (
	ComponentTypeIdP           = "idp"
	ComponentTypeIGA           = "iga"
	ComponentTypeAgentProvider = "agent_provider"
	ComponentTypePersonServer  = "person_server"
	ComponentTypeAccessServer  = "access_server"
	ComponentTypePDP           = "pdp"
	ComponentTypeGateway       = "gateway"
	ComponentTypeMCPClient     = "mcp_client"
	ComponentTypeMCPServer     = "mcp_server"
	ComponentTypeResourceAPI   = "resource_api"
	ComponentTypeSPIRE         = "spire"
)

// TrustRelationship defines a trust relationship between entities or components.
type TrustRelationship struct {
	// ID is an optional unique identifier for this relationship.
	ID string `json:"id,omitempty"`
	// From is the source entity or component ID.
	From string `json:"from"`
	// To is the target entity or component ID.
	To string `json:"to"`
	// Type is the relationship type (e.g., "authenticates", "validates", "delegates").
	Type string `json:"type"`
	// Credentials lists what credentials are exchanged in this relationship.
	Credentials []string `json:"credentials,omitempty"`
	// Mutual indicates if this is a bidirectional trust (e.g., mTLS).
	Mutual bool `json:"mutual,omitempty"`
	// Description provides additional context.
	Description string `json:"description,omitempty"`
}

// Trust relationship type constants.
const (
	TrustTypeAuthenticates = "authenticates"
	TrustTypeValidates     = "validates"
	TrustTypeDelegates     = "delegates"
	TrustTypeAuthorizes    = "authorizes"
	TrustTypeIssues        = "issues"
	TrustTypeTrusts        = "trusts"
	TrustTypeProvisions    = "provisions"
	TrustTypeAttests       = "attests"
)

// Credential type constants.
//
//nolint:gosec // G101 false positive - these are credential type identifiers, not actual credentials
const (
	CredentialX509SVID        = "x509_svid"
	CredentialJWTSVID         = "jwt_svid"
	CredentialJWTAssertion    = "jwt_assertion"
	CredentialAccessToken     = "access_token"
	CredentialIDToken         = "id_token"
	CredentialAAAgentJWT      = "aa_agent_jwt"
	CredentialAAAuthJWT       = "aa_auth_jwt"
	CredentialMTLS            = "mtls"
	CredentialAPIKey          = "api_key"
	CredentialX509Certificate = "x509_certificate"
	CredentialAttestationCert = "attestation_certificate"
	CredentialBearerToken     = "bearer_token"
	CredentialSAMLAssertion   = "saml_assertion"
	CredentialSessionCookie   = "session_cookie"
)

// Entity represents a participant in the protocol.
type Entity struct {
	// ID is the unique identifier used in flow references.
	ID string `json:"id"`

	// Name is the human-readable display name.
	Name string `json:"name"`

	// Type classifies the entity.
	Type EntityType `json:"type"`

	// Description of the entity's role.
	Description string `json:"description,omitempty"`

	// TrustLevel classifies the trust level of this entity.
	TrustLevel TrustLevel `json:"trust_level,omitempty"`

	// Metadata contains additional entity properties for rendering.
	Metadata *EntityMetadata `json:"metadata,omitempty"`

	// States defines the possible states for this entity.
	States []EntityState `json:"states,omitempty"`

	// ProtocolRoles defines the protocol-specific roles this entity implements.
	ProtocolRoles []ProtocolRole `json:"protocol_roles,omitempty"`

	// Process profile fields (used when protocol.kind = "process")

	// StepType classifies the processing nature (process profile).
	StepType StepType `json:"step_type,omitempty"`

	// Inputs defines input specifications (process profile).
	Inputs []DataPort `json:"inputs,omitempty"`

	// Outputs defines output specifications (process profile).
	Outputs []DataPort `json:"outputs,omitempty"`

	// Processing configures step execution (process profile).
	Processing *ProcessingConfig `json:"processing,omitempty"`

	// FailureModes lists possible failure scenarios (process profile).
	FailureModes []FailureMode `json:"failure_modes,omitempty"`

	// RetryStrategy configures retry behavior (process profile).
	RetryStrategy *RetryStrategy `json:"retry_strategy,omitempty"`

	// Parallel configures parallel execution for this step.
	Parallel *ParallelConfig `json:"parallel,omitempty"`
}

// EntityState represents a possible state for an entity.
type EntityState struct {
	// ID is the unique identifier for this state within the entity.
	ID string `json:"id"`

	// Name is the human-readable display name.
	Name string `json:"name,omitempty"`

	// Description provides additional context.
	Description string `json:"description,omitempty"`

	// Initial marks this as the initial state.
	Initial bool `json:"initial,omitempty"`

	// Final marks this as a terminal/final state.
	Final bool `json:"final,omitempty"`
}

// EntityMetadata contains additional entity properties.
type EntityMetadata struct {
	// Network is the network boundary this entity belongs to.
	Network string `json:"network,omitempty"`
	// ServiceType classifies the service (e.g., "api", "database", "gateway").
	ServiceType string `json:"type,omitempty"`
}

// EntityType represents the type of an entity.
type EntityType string

const (
	EntityTypeClient              EntityType = "client"
	EntityTypeAuthorizationServer EntityType = "authorization_server"
	EntityTypeResourceServer      EntityType = "resource_server"
	EntityTypeUser                EntityType = "user"
	EntityTypeBrowser             EntityType = "browser"
	EntityTypeAgent               EntityType = "agent"
	EntityTypeToolServer          EntityType = "tool_server"
	EntityTypeTool                EntityType = "tool"
	EntityTypeDelegatedAgent      EntityType = "delegated_agent"
	EntityTypeIdentityProvider    EntityType = "identity_provider"
	EntityTypeServiceProvider     EntityType = "service_provider"
	EntityTypeServer              EntityType = "server"
	EntityTypeOther               EntityType = "other"
)

// Phase represents a logical grouping of flows.
type Phase struct {
	// ID is the unique identifier.
	ID string `json:"id"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Description of the phase.
	Description string `json:"description,omitempty"`

	// Parent is the ID of the parent phase for nested phases.
	Parent string `json:"parent,omitempty"`
}

// AnnotationType represents the type of annotation.
type AnnotationType string

const (
	AnnotationTypeSecurity    AnnotationType = "security"
	AnnotationTypePerformance AnnotationType = "performance"
	AnnotationTypeDeprecated  AnnotationType = "deprecated"
	AnnotationTypeInfo        AnnotationType = "info"
	AnnotationTypeWarning     AnnotationType = "warning"
	AnnotationTypeError       AnnotationType = "error"
)

// Annotation represents a typed annotation on a flow.
type Annotation struct {
	// Type categorizes the annotation.
	Type AnnotationType `json:"type"`

	// Text is the annotation message.
	Text string `json:"text"`

	// Details provides additional context.
	Details string `json:"details,omitempty"`
}

// Alternative represents an alternative path in the flow.
type Alternative struct {
	// Condition describes when this alternative is taken.
	Condition string `json:"condition"`

	// Flows are the steps in this alternative path.
	Flows []Flow `json:"flows"`

	// Description provides additional context.
	Description string `json:"description,omitempty"`
}

// SecurityRequirement represents a security mechanism required for a flow.
type SecurityRequirement string

const (
	// SecurityRequirementToken requires a bearer or bound token.
	SecurityRequirementToken SecurityRequirement = "token"
	// SecurityRequirementSignature requires message signing.
	SecurityRequirementSignature SecurityRequirement = "signature"
	// SecurityRequirementEncryption requires message encryption.
	SecurityRequirementEncryption SecurityRequirement = "encryption"
	// SecurityRequirementMTLS requires mutual TLS authentication.
	SecurityRequirementMTLS SecurityRequirement = "mtls"
	// SecurityRequirementMAC requires message authentication code.
	SecurityRequirementMAC SecurityRequirement = "mac"
)

// FlowSecurity describes security requirements for a flow.
type FlowSecurity struct {
	// Requires lists security mechanisms required for this flow.
	Requires []SecurityRequirement `json:"requires,omitempty"`
	// Token is the token definition ID used for this flow.
	Token string `json:"token,omitempty"`
	// Confidential indicates if the flow carries sensitive data.
	Confidential bool `json:"confidential,omitempty"`
	// Description provides additional security context.
	Description string `json:"description,omitempty"`
}

// Flow represents an interaction between two entities.
type Flow struct {
	// From is the source entity ID.
	From string `json:"from"`

	// To is the target entity ID.
	To string `json:"to"`

	// Action identifies the action being performed.
	Action string `json:"action"`

	// Label is the display label (defaults to Action).
	Label string `json:"label,omitempty"`

	// Mode is the interaction mode.
	Mode FlowMode `json:"mode,omitempty"`

	// Phase is the phase ID this flow belongs to.
	Phase string `json:"phase,omitempty"`

	// Description provides additional details.
	Description string `json:"description,omitempty"`

	// Sequence provides explicit ordering.
	Sequence int `json:"sequence,omitempty"`

	// Condition specifies when this flow is executed (e.g., "token_valid", "error").
	Condition string `json:"condition,omitempty"`

	// Note is a visible annotation displayed on the diagram.
	Note string `json:"note,omitempty"`

	// Annotations are typed annotations for tooling and documentation.
	Annotations []Annotation `json:"annotations,omitempty"`

	// Alternatives are alternative paths from this flow point.
	Alternatives []Alternative `json:"alternatives,omitempty"`

	// Animation configures animation for this flow in animated SVG output.
	// Can be a string (preset name) or FlowAnimation object.
	Animation *FlowAnimation `json:"animation,omitempty"`

	// Sets defines state mutations that occur when this flow executes.
	Sets []StateMutation `json:"sets,omitempty"`

	// Security specifies security requirements for this flow.
	Security *FlowSecurity `json:"security,omitempty"`

	// DataMappings explicitly maps output ports to input ports for data lineage.
	DataMappings []DataPortMapping `json:"data_mappings,omitempty"`
}

// StateMutation represents a state change for an entity triggered by a flow.
type StateMutation struct {
	// Entity is the ID of the entity whose state changes.
	Entity string `json:"entity"`

	// To is the target state ID.
	To string `json:"to"`

	// From is the optional required prior state (for validation).
	From string `json:"from,omitempty"`
}

// FlowAnimation configures animation for a flow in animated SVG output.
type FlowAnimation struct {
	// Enabled controls whether this flow is animated. Defaults to true.
	Enabled *bool `json:"enabled,omitempty"`

	// Preset is a semantic preset name (request, success, error, warning, highlight, none).
	Preset AnimationPreset `json:"preset,omitempty"`

	// Duration is the animation cycle duration (e.g., "2s", "1.5s").
	Duration string `json:"duration,omitempty"`

	// Delay is the animation start delay (e.g., "0.5s").
	Delay string `json:"delay,omitempty"`

	// DotColor is the animated dot fill color.
	DotColor string `json:"dot_color,omitempty"`

	// DotSize is the animated dot radius in pixels.
	DotSize int `json:"dot_size,omitempty"`

	// Pulse adds a pulsing effect (useful for errors/warnings).
	Pulse bool `json:"pulse,omitempty"`

	// Easing is the CSS easing function (linear, ease-in-out, etc.).
	Easing string `json:"easing,omitempty"`
}

// AnimationPreset represents a semantic animation preset.
type AnimationPreset string

const (
	// AnimationPresetRequest is the default for outgoing requests.
	AnimationPresetRequest AnimationPreset = "request"
	// AnimationPresetResponse is for return values (gray, dashed).
	AnimationPresetResponse AnimationPreset = "response"
	// AnimationPresetSuccess indicates successful operations (green).
	AnimationPresetSuccess AnimationPreset = "success"
	// AnimationPresetError indicates errors/failures (red, pulsing).
	AnimationPresetError AnimationPreset = "error"
	// AnimationPresetWarning indicates warnings (orange, pulsing).
	AnimationPresetWarning AnimationPreset = "warning"
	// AnimationPresetHighlight emphasizes critical paths (yellow, larger dot).
	AnimationPresetHighlight AnimationPreset = "highlight"
	// AnimationPresetNone disables animation (static arrow).
	AnimationPresetNone AnimationPreset = "none"
)

// FlowMode represents the type of interaction.
type FlowMode string

const (
	FlowModeRequest     FlowMode = "request"
	FlowModeResponse    FlowMode = "response"
	FlowModeRedirect    FlowMode = "redirect"
	FlowModeCallback    FlowMode = "callback"
	FlowModeInteractive FlowMode = "interactive"
	FlowModeEvent       FlowMode = "event"
	FlowModeToolCall    FlowMode = "tool_call"
	FlowModeToolResult  FlowMode = "tool_result"
)

// StateTransition represents a state transition extracted from flows.
type StateTransition struct {
	// EntityID is the entity undergoing the state change.
	EntityID string
	// FromState is the prior state (empty if not specified).
	FromState string
	// ToState is the target state.
	ToState string
	// FlowAction is the action that triggers this transition.
	FlowAction string
	// FlowLabel is the display label for the triggering flow.
	FlowLabel string
}

// LineageEdge represents a data flow connection between ports.
type LineageEdge struct {
	// SourceEntity is the ID of the entity providing the data.
	SourceEntity string `json:"source_entity"`
	// SourcePort is the name of the output port.
	SourcePort string `json:"source_port"`
	// TargetEntity is the ID of the entity consuming the data.
	TargetEntity string `json:"target_entity"`
	// TargetPort is the name of the input port.
	TargetPort string `json:"target_port"`
	// FlowIndex is the index of the flow that creates this connection.
	FlowIndex int `json:"flow_index"`
	// Transformation describes any data transformation applied.
	Transformation string `json:"transformation,omitempty"`
}

// DataLineage represents the complete data lineage graph for a protocol.
type DataLineage struct {
	// ProtocolID is the ID of the source protocol.
	ProtocolID string `json:"protocol_id"`
	// Edges are the data flow connections.
	Edges []LineageEdge `json:"edges"`
	// Sources are ports that have no upstream connections.
	Sources []PortReference `json:"sources"`
	// Sinks are ports that have no downstream connections.
	Sinks []PortReference `json:"sinks"`
	// SensitiveDataPaths tracks paths containing sensitive data.
	SensitiveDataPaths [][]PortReference `json:"sensitive_paths,omitempty"`
}

// PortReference identifies a specific port on an entity.
type PortReference struct {
	// EntityID is the ID of the entity.
	EntityID string `json:"entity_id"`
	// PortName is the name of the port.
	PortName string `json:"port_name"`
	// PortKind indicates if this is an input or output port.
	PortKind string `json:"port_kind"` // "input" or "output"
	// Sensitive indicates if the port handles sensitive data.
	Sensitive bool `json:"sensitive,omitempty"`
}

// DataPortMapping explicitly maps output ports to input ports across a flow.
type DataPortMapping struct {
	// OutputPort is the source port name on the From entity.
	OutputPort string `json:"output_port"`
	// InputPort is the target port name on the To entity.
	InputPort string `json:"input_port"`
	// Transformation describes any data transformation applied.
	Transformation string `json:"transformation,omitempty"`
}
