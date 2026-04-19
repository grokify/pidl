package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// D2Style represents different D2 diagram styles.
type D2Style string

const (
	// D2StyleSequence renders as a sequence diagram.
	D2StyleSequence D2Style = "sequence"
	// D2StyleFlow renders as a data flow diagram.
	D2StyleFlow D2Style = "flow"
	// D2StyleArch renders as an architecture diagram with grouped entities.
	D2StyleArch D2Style = "arch"
)

// D2Renderer renders PIDL protocols as D2 diagrams.
type D2Renderer struct {
	// Style determines the diagram style (sequence, flow, or arch).
	Style D2Style

	// Title includes the protocol name as diagram title.
	Title bool

	// ShowDescriptions includes flow descriptions as tooltips.
	ShowDescriptions bool

	// Direction sets the diagram direction (down, right, left, up).
	Direction string
}

// NewD2 creates a new D2 renderer with default options (sequence diagram).
func NewD2() *D2Renderer {
	return &D2Renderer{
		Style:            D2StyleSequence,
		Title:            true,
		ShowDescriptions: false,
		Direction:        "right",
	}
}

// NewD2Flow creates a new D2 renderer for data flow diagrams.
func NewD2Flow() *D2Renderer {
	return &D2Renderer{
		Style:            D2StyleFlow,
		Title:            true,
		ShowDescriptions: false,
		Direction:        "right",
	}
}

// NewD2Arch creates a new D2 renderer for architecture diagrams.
func NewD2Arch() *D2Renderer {
	return &D2Renderer{
		Style:            D2StyleArch,
		Title:            true,
		ShowDescriptions: false,
		Direction:        "right",
	}
}

// Format returns the output format.
func (r *D2Renderer) Format() Format {
	switch r.Style {
	case D2StyleFlow:
		return FormatD2Flow
	case D2StyleArch:
		return FormatD2Arch
	default:
		return FormatD2
	}
}

// Render writes the D2 diagram to the writer.
func (r *D2Renderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the D2 diagram as a string.
func (r *D2Renderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *D2Renderer) render(p *pidl.Protocol) string {
	switch r.Style {
	case D2StyleFlow:
		return r.renderFlow(p)
	case D2StyleArch:
		return r.renderArch(p)
	default:
		return r.renderSequence(p)
	}
}

// renderSequence renders a D2 sequence diagram.
func (r *D2Renderer) renderSequence(p *pidl.Protocol) string {
	var sb strings.Builder

	// Title as a label
	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "title: %s {\n  shape: text\n  near: top-center\n  style.font-size: 24\n}\n\n", p.ProtocolMeta.Name)
	}

	// Declare the sequence diagram shape
	sb.WriteString("sequence: {\n")
	sb.WriteString("  shape: sequence_diagram\n\n")

	// Declare actors
	for _, e := range p.Entities {
		fmt.Fprintf(&sb, "  %s: %s\n", r.sanitizeID(e.ID), e.Name)
	}

	sb.WriteString("\n")

	// Track sequence number for ordering
	seq := 1

	// Track current phase for grouping
	currentPhase := ""
	inPhaseGroup := false

	for _, f := range p.Flows {
		// Handle phase changes
		if f.Phase != "" && f.Phase != currentPhase {
			if inPhaseGroup {
				sb.WriteString("  }\n\n")
			}
			phase := p.PhaseByID(f.Phase)
			if phase != nil {
				fmt.Fprintf(&sb, "  %s: %s {\n", r.sanitizeID(f.Phase), phase.Name)
				inPhaseGroup = true
			}
			currentPhase = f.Phase
		}

		// Render the flow
		indent := "  "
		if inPhaseGroup {
			indent = "    "
		}

		from := r.sanitizeID(f.From)
		to := r.sanitizeID(f.To)
		label := f.DisplayLabel()

		// Add mode annotation
		if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
			label = fmt.Sprintf("%s (%s)", label, ann)
		}

		// D2 sequence diagram message syntax
		arrow := r.modeToArrow(f.EffectiveMode())
		fmt.Fprintf(&sb, "%sseq%d: %s %s %s: %s\n", indent, seq, from, arrow, to, label)
		seq++
	}

	if inPhaseGroup {
		sb.WriteString("  }\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}

// renderFlow renders a D2 data flow diagram.
func (r *D2Renderer) renderFlow(p *pidl.Protocol) string {
	var sb strings.Builder

	// Direction
	if r.Direction != "" {
		fmt.Fprintf(&sb, "direction: %s\n\n", r.Direction)
	}

	// Title
	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "title: %s {\n  shape: text\n  near: top-center\n  style.font-size: 24\n}\n\n", p.ProtocolMeta.Name)
	}

	// Declare entities with shapes based on type
	for _, e := range p.Entities {
		id := r.sanitizeID(e.ID)
		shape := r.entityTypeToShape(e.Type)
		fmt.Fprintf(&sb, "%s: %s {\n  shape: %s\n", id, e.Name, shape)
		if e.Description != "" && r.ShowDescriptions {
			fmt.Fprintf(&sb, "  tooltip: %s\n", e.Description)
		}
		sb.WriteString("}\n")
	}

	sb.WriteString("\n")

	// Render flows as connections
	for i, f := range p.Flows {
		from := r.sanitizeID(f.From)
		to := r.sanitizeID(f.To)
		label := f.DisplayLabel()

		// Add mode annotation
		if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
			label = fmt.Sprintf("%s (%s)", label, ann)
		}

		// Connection with label
		arrow := r.modeToD2Arrow(f.EffectiveMode())
		fmt.Fprintf(&sb, "%s %s %s: %d. %s\n", from, arrow, to, i+1, label)
	}

	return sb.String()
}

// renderArch renders a D2 architecture diagram with phase groupings.
func (r *D2Renderer) renderArch(p *pidl.Protocol) string {
	var sb strings.Builder

	// Direction
	if r.Direction != "" {
		fmt.Fprintf(&sb, "direction: %s\n\n", r.Direction)
	}

	// Title
	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "title: %s {\n  shape: text\n  near: top-center\n  style.font-size: 24\n}\n\n", p.ProtocolMeta.Name)
	}

	// Group entities by type for architecture view
	entityGroups := make(map[string][]pidl.Entity)
	for _, e := range p.Entities {
		group := r.entityTypeToGroup(e.Type)
		entityGroups[group] = append(entityGroups[group], e)
	}

	// Render grouped entities
	for group, entities := range entityGroups {
		if group != "" {
			fmt.Fprintf(&sb, "%s: %s {\n", r.sanitizeID(group), group)
			for _, e := range entities {
				id := r.sanitizeID(e.ID)
				shape := r.entityTypeToShape(e.Type)
				fmt.Fprintf(&sb, "  %s: %s {\n    shape: %s\n  }\n", id, e.Name, shape)
			}
			sb.WriteString("}\n\n")
		} else {
			// Ungrouped entities at top level
			for _, e := range entities {
				id := r.sanitizeID(e.ID)
				shape := r.entityTypeToShape(e.Type)
				fmt.Fprintf(&sb, "%s: %s {\n  shape: %s\n}\n", id, e.Name, shape)
			}
			sb.WriteString("\n")
		}
	}

	// Render flows as connections
	for i, f := range p.Flows {
		from := r.qualifiedID(p, f.From)
		to := r.qualifiedID(p, f.To)
		label := f.DisplayLabel()

		if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
			label = fmt.Sprintf("%s (%s)", label, ann)
		}

		arrow := r.modeToD2Arrow(f.EffectiveMode())
		fmt.Fprintf(&sb, "%s %s %s: %d. %s\n", from, arrow, to, i+1, label)
	}

	return sb.String()
}

func (r *D2Renderer) sanitizeID(id string) string {
	// D2 IDs: replace hyphens with underscores, ensure valid identifier
	result := strings.ReplaceAll(id, "-", "_")
	return result
}

func (r *D2Renderer) modeToArrow(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult, pidl.FlowModeEvent:
		return "<-"
	default:
		return "->"
	}
}

func (r *D2Renderer) modeToD2Arrow(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult:
		return "<--"
	case pidl.FlowModeEvent:
		return "<-"
	default:
		return "->"
	}
}

func (r *D2Renderer) modeAnnotation(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeRedirect:
		return "redirect"
	case pidl.FlowModeCallback:
		return "callback"
	case pidl.FlowModeToolCall:
		return "tool"
	case pidl.FlowModeToolResult:
		return "result"
	default:
		return ""
	}
}

func (r *D2Renderer) entityTypeToShape(t pidl.EntityType) string {
	switch t {
	case pidl.EntityTypeUser:
		return "person"
	case pidl.EntityTypeBrowser:
		return "rectangle"
	case pidl.EntityTypeClient:
		return "rectangle"
	case pidl.EntityTypeServer, pidl.EntityTypeResourceServer, pidl.EntityTypeAuthorizationServer:
		return "cylinder"
	case pidl.EntityTypeAgent, pidl.EntityTypeDelegatedAgent:
		return "hexagon"
	case pidl.EntityTypeToolServer, pidl.EntityTypeTool:
		return "package"
	case pidl.EntityTypeIdentityProvider, pidl.EntityTypeServiceProvider:
		return "cloud"
	default:
		return "rectangle"
	}
}

func (r *D2Renderer) entityTypeToGroup(t pidl.EntityType) string {
	switch t {
	case pidl.EntityTypeUser, pidl.EntityTypeBrowser:
		return "Users"
	case pidl.EntityTypeClient:
		return "Clients"
	case pidl.EntityTypeServer, pidl.EntityTypeResourceServer, pidl.EntityTypeAuthorizationServer:
		return "Servers"
	case pidl.EntityTypeAgent, pidl.EntityTypeDelegatedAgent:
		return "Agents"
	case pidl.EntityTypeToolServer, pidl.EntityTypeTool:
		return "Tools"
	case pidl.EntityTypeIdentityProvider, pidl.EntityTypeServiceProvider:
		return "Providers"
	default:
		return ""
	}
}

func (r *D2Renderer) qualifiedID(p *pidl.Protocol, entityID string) string {
	entity := p.EntityByID(entityID)
	if entity == nil {
		return r.sanitizeID(entityID)
	}

	group := r.entityTypeToGroup(entity.Type)
	if group != "" {
		return fmt.Sprintf("%s.%s", r.sanitizeID(group), r.sanitizeID(entityID))
	}
	return r.sanitizeID(entityID)
}
