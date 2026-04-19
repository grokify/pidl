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
