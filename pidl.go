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
