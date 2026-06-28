// Package pidl provides types and utilities for the Protocol Interaction Description Language.
// PIDL is a JSON-based DSL for describing protocol choreography that compiles to diagrams.
package pidl

// Protocol represents a complete PIDL document describing a protocol's choreography.
type Protocol struct {
	// ProtocolMeta contains metadata about the protocol.
	ProtocolMeta ProtocolMeta `json:"protocol"`

	// Entities are the participants in the protocol (systems, actors, services).
	Entities []Entity `json:"entities"`

	// Phases provide optional logical grouping of flows.
	Phases []Phase `json:"phases,omitempty"`

	// Flows are the interactions between entities.
	Flows []Flow `json:"flows"`

	// Metadata contains additional protocol-level configuration.
	Metadata *ProtocolMetadata `json:"metadata,omitempty"`
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
	ProtocolOAuth   = "oauth"
	ProtocolSCIM    = "scim"
	ProtocolSPIFFE  = "spiffe"
	ProtocolAAuth   = "aauth"
	ProtocolIDJAG   = "idjag"
	ProtocolAuthZEN = "authzen"
	ProtocolMCP     = "mcp"
	ProtocolA2A     = "a2a"
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
	CredentialX509SVID     = "x509_svid"
	CredentialJWTSVID      = "jwt_svid"
	CredentialJWTAssertion = "jwt_assertion"
	CredentialAccessToken  = "access_token"
	CredentialIDToken      = "id_token"
	CredentialAAAgentJWT   = "aa_agent_jwt"
	CredentialAAAuthJWT    = "aa_auth_jwt"
	CredentialMTLS         = "mtls"
	CredentialAPIKey       = "api_key"
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

// IsAnimationEnabled returns whether animation is enabled for this flow.
func (f *FlowAnimation) IsAnimationEnabled() bool {
	if f == nil {
		return true // default enabled
	}
	if f.Preset == AnimationPresetNone {
		return false
	}
	if f.Enabled != nil {
		return *f.Enabled
	}
	return true
}

// EffectiveDotColor returns the dot color, applying preset defaults.
func (f *FlowAnimation) EffectiveDotColor(defaultColor string) string {
	if f == nil || f.DotColor == "" {
		if f != nil {
			switch f.Preset {
			case AnimationPresetSuccess:
				return "#68d391"
			case AnimationPresetError:
				return "#fc8181"
			case AnimationPresetWarning:
				return "#f6ad55"
			case AnimationPresetHighlight:
				return "#faf089"
			case AnimationPresetResponse:
				return "#a0aec0"
			}
		}
		return defaultColor
	}
	return f.DotColor
}

// EffectiveDotSize returns the dot size, applying preset defaults.
func (f *FlowAnimation) EffectiveDotSize(defaultSize int) int {
	if f == nil || f.DotSize == 0 {
		if f != nil && f.Preset == AnimationPresetHighlight {
			return 6
		}
		return defaultSize
	}
	return f.DotSize
}

// ShouldPulse returns whether the dot should pulse.
func (f *FlowAnimation) ShouldPulse() bool {
	if f == nil {
		return false
	}
	if f.Pulse {
		return true
	}
	return f.Preset == AnimationPresetError || f.Preset == AnimationPresetWarning
}

// ShouldGlow returns whether the dot should have a glow effect.
func (f *FlowAnimation) ShouldGlow() bool {
	if f == nil {
		return false
	}
	return f.Preset == AnimationPresetHighlight
}

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

// DisplayLabel returns the label for display, falling back to Action if Label is empty.
func (f Flow) DisplayLabel() string {
	if f.Label != "" {
		return f.Label
	}
	return f.Action
}

// EffectiveMode returns the flow mode, defaulting to FlowModeRequest if empty.
func (f Flow) EffectiveMode() FlowMode {
	if f.Mode == "" {
		return FlowModeRequest
	}
	return f.Mode
}

// EntityByID returns the entity with the given ID, or nil if not found.
func (p *Protocol) EntityByID(id string) *Entity {
	for i := range p.Entities {
		if p.Entities[i].ID == id {
			return &p.Entities[i]
		}
	}
	return nil
}

// PhaseByID returns the phase with the given ID, or nil if not found.
func (p *Protocol) PhaseByID(id string) *Phase {
	for i := range p.Phases {
		if p.Phases[i].ID == id {
			return &p.Phases[i]
		}
	}
	return nil
}

// FlowsByPhase returns all flows belonging to the given phase.
func (p *Protocol) FlowsByPhase(phaseID string) []Flow {
	var flows []Flow
	for _, f := range p.Flows {
		if f.Phase == phaseID {
			flows = append(flows, f)
		}
	}
	return flows
}

// EntityIDs returns a slice of all entity IDs.
func (p *Protocol) EntityIDs() []string {
	ids := make([]string, len(p.Entities))
	for i, e := range p.Entities {
		ids[i] = e.ID
	}
	return ids
}

// PhaseIDs returns a slice of all phase IDs.
func (p *Protocol) PhaseIDs() []string {
	ids := make([]string, len(p.Phases))
	for i, ph := range p.Phases {
		ids[i] = ph.ID
	}
	return ids
}

// RootPhases returns phases that have no parent (top-level phases).
func (p *Protocol) RootPhases() []Phase {
	var roots []Phase
	for _, ph := range p.Phases {
		if ph.Parent == "" {
			roots = append(roots, ph)
		}
	}
	return roots
}

// ChildPhases returns phases that have the given parent ID.
func (p *Protocol) ChildPhases(parentID string) []Phase {
	var children []Phase
	for _, ph := range p.Phases {
		if ph.Parent == parentID {
			children = append(children, ph)
		}
	}
	return children
}

// PhaseDepth returns the nesting depth of a phase (0 for root phases).
func (p *Protocol) PhaseDepth(phaseID string) int {
	depth := 0
	current := p.PhaseByID(phaseID)
	for current != nil && current.Parent != "" {
		depth++
		current = p.PhaseByID(current.Parent)
	}
	return depth
}

// HasCondition returns true if the flow has a condition.
func (f Flow) HasCondition() bool {
	return f.Condition != ""
}

// HasAlternatives returns true if the flow has alternative paths.
func (f Flow) HasAlternatives() bool {
	return len(f.Alternatives) > 0
}

// HasAnnotations returns true if the flow has annotations.
func (f Flow) HasAnnotations() bool {
	return len(f.Annotations) > 0
}

// HasNote returns true if the flow has a note.
func (f Flow) HasNote() bool {
	return f.Note != ""
}

// IsValidAnnotationType checks if the annotation type is valid.
func IsValidAnnotationType(t AnnotationType) bool {
	switch t {
	case AnnotationTypeSecurity, AnnotationTypePerformance, AnnotationTypeDeprecated,
		AnnotationTypeInfo, AnnotationTypeWarning, AnnotationTypeError:
		return true
	}
	return false
}

// HasStates returns true if the entity has any states defined.
func (e Entity) HasStates() bool {
	return len(e.States) > 0
}

// StateByID returns the state with the given ID, or nil if not found.
func (e Entity) StateByID(id string) *EntityState {
	for i := range e.States {
		if e.States[i].ID == id {
			return &e.States[i]
		}
	}
	return nil
}

// InitialState returns the initial state, or nil if none is marked as initial.
func (e Entity) InitialState() *EntityState {
	for i := range e.States {
		if e.States[i].Initial {
			return &e.States[i]
		}
	}
	return nil
}

// StateIDs returns a slice of all state IDs for this entity.
func (e Entity) StateIDs() []string {
	ids := make([]string, len(e.States))
	for i, s := range e.States {
		ids[i] = s.ID
	}
	return ids
}

// HasStateMutations returns true if the flow has any state mutations.
func (f Flow) HasStateMutations() bool {
	return len(f.Sets) > 0
}

// EntitiesWithStates returns all entities that have states defined.
func (p *Protocol) EntitiesWithStates() []Entity {
	var entities []Entity
	for _, e := range p.Entities {
		if e.HasStates() {
			entities = append(entities, e)
		}
	}
	return entities
}

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

// StateTransitions extracts all state transitions from the protocol's flows.
func (p *Protocol) StateTransitions() []StateTransition {
	var transitions []StateTransition
	for _, f := range p.Flows {
		for _, m := range f.Sets {
			transitions = append(transitions, StateTransition{
				EntityID:   m.Entity,
				FromState:  m.From,
				ToState:    m.To,
				FlowAction: f.Action,
				FlowLabel:  f.DisplayLabel(),
			})
		}
	}
	return transitions
}

// StateTransitionsForEntity returns state transitions for a specific entity.
func (p *Protocol) StateTransitionsForEntity(entityID string) []StateTransition {
	var transitions []StateTransition
	for _, t := range p.StateTransitions() {
		if t.EntityID == entityID {
			transitions = append(transitions, t)
		}
	}
	return transitions
}

// TokenByID returns the token definition with the given ID, or nil if not found.
func (p *Protocol) TokenByID(id string) *TokenDefinition {
	if p.Metadata == nil {
		return nil
	}
	for i := range p.Metadata.Tokens {
		if p.Metadata.Tokens[i].ID == id {
			return &p.Metadata.Tokens[i]
		}
	}
	return nil
}

// HasSecurity returns true if the flow has security requirements.
func (f Flow) HasSecurity() bool {
	return f.Security != nil && (len(f.Security.Requires) > 0 || f.Security.Token != "" || f.Security.Confidential)
}

// RequiresEncryption returns true if the flow requires encryption.
func (f Flow) RequiresEncryption() bool {
	if f.Security == nil {
		return false
	}
	for _, req := range f.Security.Requires {
		if req == SecurityRequirementEncryption {
			return true
		}
	}
	return false
}

// RequiresToken returns true if the flow requires a token.
func (f Flow) RequiresToken() bool {
	if f.Security == nil {
		return false
	}
	if f.Security.Token != "" {
		return true
	}
	for _, req := range f.Security.Requires {
		if req == SecurityRequirementToken {
			return true
		}
	}
	return false
}

// IsValidTrustLevel checks if the trust level is valid.
func IsValidTrustLevel(t TrustLevel) bool {
	switch t {
	case TrustLevelTrusted, TrustLevelSemiTrusted, TrustLevelUntrusted, TrustLevelAuthoritative:
		return true
	}
	return false
}

// IsValidSecurityRequirement checks if the security requirement is valid.
func IsValidSecurityRequirement(r SecurityRequirement) bool {
	switch r {
	case SecurityRequirementToken, SecurityRequirementSignature, SecurityRequirementEncryption,
		SecurityRequirementMTLS, SecurityRequirementMAC:
		return true
	}
	return false
}

// HasProtocolRoles returns true if the entity has protocol roles defined.
func (e Entity) HasProtocolRoles() bool {
	return len(e.ProtocolRoles) > 0
}

// RolesByProtocol returns all roles for a specific protocol.
func (e Entity) RolesByProtocol(protocol string) []ProtocolRole {
	var roles []ProtocolRole
	for _, r := range e.ProtocolRoles {
		if r.Protocol == protocol {
			roles = append(roles, r)
		}
	}
	return roles
}

// HasRole checks if the entity has a specific protocol role.
func (e Entity) HasRole(protocol, role string) bool {
	for _, r := range e.ProtocolRoles {
		if r.Protocol == protocol && r.Role == role {
			return true
		}
	}
	return false
}

// EntitiesWithProtocolRoles returns all entities that have protocol roles defined.
func (p *Protocol) EntitiesWithProtocolRoles() []Entity {
	var entities []Entity
	for _, e := range p.Entities {
		if e.HasProtocolRoles() {
			entities = append(entities, e)
		}
	}
	return entities
}

// EntitiesByProtocol returns all entities that implement a role for a specific protocol.
func (p *Protocol) EntitiesByProtocol(protocol string) []Entity {
	var entities []Entity
	for _, e := range p.Entities {
		if len(e.RolesByProtocol(protocol)) > 0 {
			entities = append(entities, e)
		}
	}
	return entities
}

// EntitiesByRole returns all entities that implement a specific protocol role.
func (p *Protocol) EntitiesByRole(protocol, role string) []Entity {
	var entities []Entity
	for _, e := range p.Entities {
		if e.HasRole(protocol, role) {
			entities = append(entities, e)
		}
	}
	return entities
}

// ComponentByID returns the component with the given ID, or nil if not found.
func (p *Protocol) ComponentByID(id string) *DeploymentComponent {
	if p.Metadata == nil {
		return nil
	}
	for i := range p.Metadata.Components {
		if p.Metadata.Components[i].ID == id {
			return &p.Metadata.Components[i]
		}
	}
	return nil
}

// ComponentsByType returns all components of a given type.
func (p *Protocol) ComponentsByType(componentType string) []DeploymentComponent {
	if p.Metadata == nil {
		return nil
	}
	var components []DeploymentComponent
	for _, c := range p.Metadata.Components {
		if c.Type == componentType {
			components = append(components, c)
		}
	}
	return components
}

// EntitiesInComponent returns all entities that belong to a component.
func (p *Protocol) EntitiesInComponent(componentID string) []Entity {
	c := p.ComponentByID(componentID)
	if c == nil {
		return nil
	}
	var entities []Entity
	for _, eid := range c.Entities {
		if e := p.EntityByID(eid); e != nil {
			entities = append(entities, *e)
		}
	}
	return entities
}

// ComponentForEntity returns the component that contains an entity, or nil if not found.
func (p *Protocol) ComponentForEntity(entityID string) *DeploymentComponent {
	if p.Metadata == nil {
		return nil
	}
	for i := range p.Metadata.Components {
		for _, eid := range p.Metadata.Components[i].Entities {
			if eid == entityID {
				return &p.Metadata.Components[i]
			}
		}
	}
	return nil
}

// TrustRelationByID returns the trust relation with the given ID, or nil if not found.
func (p *Protocol) TrustRelationByID(id string) *TrustRelationship {
	if p.Metadata == nil {
		return nil
	}
	for i := range p.Metadata.TrustRelations {
		if p.Metadata.TrustRelations[i].ID == id {
			return &p.Metadata.TrustRelations[i]
		}
	}
	return nil
}

// TrustRelationsFrom returns all trust relations originating from an entity or component.
func (p *Protocol) TrustRelationsFrom(id string) []TrustRelationship {
	if p.Metadata == nil {
		return nil
	}
	var relations []TrustRelationship
	for _, r := range p.Metadata.TrustRelations {
		if r.From == id {
			relations = append(relations, r)
		}
	}
	return relations
}

// TrustRelationsTo returns all trust relations targeting an entity or component.
func (p *Protocol) TrustRelationsTo(id string) []TrustRelationship {
	if p.Metadata == nil {
		return nil
	}
	var relations []TrustRelationship
	for _, r := range p.Metadata.TrustRelations {
		if r.To == id {
			relations = append(relations, r)
		}
	}
	return relations
}

// TrustRelationsByType returns all trust relations of a given type.
func (p *Protocol) TrustRelationsByType(relType string) []TrustRelationship {
	if p.Metadata == nil {
		return nil
	}
	var relations []TrustRelationship
	for _, r := range p.Metadata.TrustRelations {
		if r.Type == relType {
			relations = append(relations, r)
		}
	}
	return relations
}

// AllProtocols returns a unique list of all protocols referenced in entity roles.
func (p *Protocol) AllProtocols() []string {
	seen := make(map[string]bool)
	var protocols []string
	for _, e := range p.Entities {
		for _, r := range e.ProtocolRoles {
			if !seen[r.Protocol] {
				seen[r.Protocol] = true
				protocols = append(protocols, r.Protocol)
			}
		}
	}
	return protocols
}

// AllComponentTypes returns a unique list of all component types.
func (p *Protocol) AllComponentTypes() []string {
	if p.Metadata == nil {
		return nil
	}
	seen := make(map[string]bool)
	var types []string
	for _, c := range p.Metadata.Components {
		if !seen[c.Type] {
			seen[c.Type] = true
			types = append(types, c.Type)
		}
	}
	return types
}

// AllTrustRelationTypes returns a unique list of all trust relation types.
func (p *Protocol) AllTrustRelationTypes() []string {
	if p.Metadata == nil {
		return nil
	}
	seen := make(map[string]bool)
	var types []string
	for _, r := range p.Metadata.TrustRelations {
		if !seen[r.Type] {
			seen[r.Type] = true
			types = append(types, r.Type)
		}
	}
	return types
}

// IsValidProtocol checks if the protocol identifier is a known protocol.
func IsValidProtocol(protocol string) bool {
	switch protocol {
	case ProtocolOAuth, ProtocolSCIM, ProtocolSPIFFE, ProtocolAAuth,
		ProtocolIDJAG, ProtocolAuthZEN, ProtocolMCP, ProtocolA2A:
		return true
	}
	return false
}

// IsValidComponentType checks if the component type is a known type.
func IsValidComponentType(t string) bool {
	switch t {
	case ComponentTypeIdP, ComponentTypeIGA, ComponentTypeAgentProvider,
		ComponentTypePersonServer, ComponentTypeAccessServer, ComponentTypePDP,
		ComponentTypeGateway, ComponentTypeMCPClient, ComponentTypeMCPServer,
		ComponentTypeResourceAPI, ComponentTypeSPIRE:
		return true
	}
	return false
}

// IsValidTrustRelationType checks if the trust relation type is a known type.
func IsValidTrustRelationType(t string) bool {
	switch t {
	case TrustTypeAuthenticates, TrustTypeValidates, TrustTypeDelegates,
		TrustTypeAuthorizes, TrustTypeIssues, TrustTypeTrusts,
		TrustTypeProvisions, TrustTypeAttests:
		return true
	}
	return false
}

// IsValidCredential checks if the credential type is a known type.
func IsValidCredential(c string) bool {
	switch c {
	case CredentialX509SVID, CredentialJWTSVID, CredentialJWTAssertion,
		CredentialAccessToken, CredentialIDToken, CredentialAAAgentJWT,
		CredentialAAAuthJWT, CredentialMTLS, CredentialAPIKey:
		return true
	}
	return false
}
