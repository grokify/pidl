package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// MermaidComponentRenderer renders PIDL protocols as Mermaid flowcharts
// showing deployment components and their relationships.
type MermaidComponentRenderer struct {
	// Title includes the protocol name as diagram title.
	Title bool

	// ShowEntities includes entity nodes within components.
	ShowEntities bool

	// ShowRoles includes protocol roles on components.
	ShowRoles bool
}

// NewMermaidComponent creates a new Mermaid component diagram renderer.
func NewMermaidComponent() *MermaidComponentRenderer {
	return &MermaidComponentRenderer{
		Title:        true,
		ShowEntities: true,
		ShowRoles:    true,
	}
}

// Format returns the output format.
func (r *MermaidComponentRenderer) Format() Format {
	return FormatMermaidComponent
}

// Render writes the Mermaid diagram to the writer.
func (r *MermaidComponentRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the Mermaid diagram as a string.
func (r *MermaidComponentRenderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *MermaidComponentRenderer) render(p *pidl.Protocol) string {
	var sb strings.Builder

	sb.WriteString("flowchart TB\n")

	if r.Title && p.ProtocolMeta.Name != "" {
		// Use a note for the title since flowchart doesn't have native title
		sb.WriteString("    %% " + p.ProtocolMeta.Name + "\n\n")
	}

	// Check if we have components
	if p.Metadata == nil || len(p.Metadata.Components) == 0 {
		// No components - render entities directly
		r.renderEntitiesOnly(&sb, p)
		return sb.String()
	}

	// Track which entities are in components
	entityInComponent := make(map[string]bool)

	// Render each component as a subgraph
	for _, c := range p.Metadata.Components {
		r.renderComponent(&sb, p, c, entityInComponent)
	}

	// Render entities not in any component
	r.renderUnassignedEntities(&sb, p, entityInComponent)

	// Render flows between entities/components
	r.renderFlows(&sb, p)

	return sb.String()
}

func (r *MermaidComponentRenderer) renderComponent(sb *strings.Builder, p *pidl.Protocol, c pidl.DeploymentComponent, entityInComponent map[string]bool) {
	compID := r.sanitizeID(c.ID)

	// Start subgraph for the component
	fmt.Fprintf(sb, "    subgraph %s[\"%s\"]\n", compID, r.componentLabel(c))

	// Add component type styling via class
	fmt.Fprintf(sb, "        direction TB\n")

	if r.ShowEntities {
		// Render entities within the component
		for _, eid := range c.Entities {
			entityInComponent[eid] = true
			e := p.EntityByID(eid)
			if e != nil {
				nodeID := r.sanitizeID(eid)
				shape := r.entityTypeToShape(e.Type)
				fmt.Fprintf(sb, "        %s%s%s\n", nodeID, shape[0], e.Name+shape[1])
			}
		}
	}

	sb.WriteString("    end\n\n")
}

func (r *MermaidComponentRenderer) componentLabel(c pidl.DeploymentComponent) string {
	label := c.Name
	if r.ShowRoles && len(c.Implements) > 0 {
		var roles []string
		for _, role := range c.Implements {
			roles = append(roles, fmt.Sprintf("%s:%s", role.Protocol, role.Role))
		}
		label = fmt.Sprintf("%s<br/><small>%s</small>", c.Name, strings.Join(roles, ", "))
	}
	return label
}

func (r *MermaidComponentRenderer) renderUnassignedEntities(sb *strings.Builder, p *pidl.Protocol, entityInComponent map[string]bool) {
	var unassigned []pidl.Entity
	for _, e := range p.Entities {
		if !entityInComponent[e.ID] {
			unassigned = append(unassigned, e)
		}
	}

	if len(unassigned) == 0 {
		return
	}

	sb.WriteString("    subgraph other[\"Other\"]\n")
	sb.WriteString("        direction TB\n")
	for _, e := range unassigned {
		nodeID := r.sanitizeID(e.ID)
		shape := r.entityTypeToShape(e.Type)
		fmt.Fprintf(sb, "        %s%s%s\n", nodeID, shape[0], e.Name+shape[1])
	}
	sb.WriteString("    end\n\n")
}

func (r *MermaidComponentRenderer) renderEntitiesOnly(sb *strings.Builder, p *pidl.Protocol) {
	for _, e := range p.Entities {
		nodeID := r.sanitizeID(e.ID)
		shape := r.entityTypeToShape(e.Type)
		fmt.Fprintf(sb, "    %s%s%s\n", nodeID, shape[0], e.Name+shape[1])
	}
	sb.WriteString("\n")

	r.renderFlows(sb, p)
}

func (r *MermaidComponentRenderer) renderFlows(sb *strings.Builder, p *pidl.Protocol) {
	// Track unique connections to avoid duplicates
	seen := make(map[string]bool)

	for _, f := range p.Flows {
		fromID := r.sanitizeID(f.From)
		toID := r.sanitizeID(f.To)
		key := fromID + "->" + toID

		if seen[key] {
			continue
		}
		seen[key] = true

		arrow := r.modeToArrow(f.EffectiveMode())
		fmt.Fprintf(sb, "    %s %s %s\n", fromID, arrow, toID)
	}
}

func (r *MermaidComponentRenderer) sanitizeID(id string) string {
	// Mermaid IDs: replace hyphens with underscores
	return strings.ReplaceAll(id, "-", "_")
}

func (r *MermaidComponentRenderer) entityTypeToShape(t pidl.EntityType) [2]string {
	// Returns [opening, closing] for Mermaid shapes
	switch t {
	case pidl.EntityTypeUser:
		return [2]string{"((", "))"} // Circle (person)
	case pidl.EntityTypeBrowser:
		return [2]string{"[", "]"} // Rectangle
	case pidl.EntityTypeClient:
		return [2]string{"[", "]"} // Rectangle
	case pidl.EntityTypeServer, pidl.EntityTypeResourceServer, pidl.EntityTypeAuthorizationServer:
		return [2]string{"[(", ")]"} // Cylinder (database-like)
	case pidl.EntityTypeAgent, pidl.EntityTypeDelegatedAgent:
		return [2]string{"{{", "}}"} // Hexagon
	case pidl.EntityTypeToolServer, pidl.EntityTypeTool:
		return [2]string{"[/", "/]"} // Parallelogram
	case pidl.EntityTypeIdentityProvider, pidl.EntityTypeServiceProvider:
		return [2]string{"([", "])"} // Stadium
	default:
		return [2]string{"[", "]"} // Rectangle
	}
}

func (r *MermaidComponentRenderer) modeToArrow(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult:
		return "-.->|response|"
	case pidl.FlowModeRedirect:
		return "-->|redirect|"
	case pidl.FlowModeCallback:
		return "-->|callback|"
	case pidl.FlowModeEvent:
		return "-.->|event|"
	default:
		return "-->"
	}
}
