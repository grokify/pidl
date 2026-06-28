package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// MermaidTrustRenderer renders PIDL protocols as Mermaid flowcharts
// showing trust relationships between entities and components.
type MermaidTrustRenderer struct {
	// Title includes the protocol name as diagram title.
	Title bool

	// ShowCredentials includes credential types on edges.
	ShowCredentials bool

	// ShowComponents groups entities by their components.
	ShowComponents bool
}

// NewMermaidTrust creates a new Mermaid trust diagram renderer.
func NewMermaidTrust() *MermaidTrustRenderer {
	return &MermaidTrustRenderer{
		Title:           true,
		ShowCredentials: true,
		ShowComponents:  true,
	}
}

// Format returns the output format.
func (r *MermaidTrustRenderer) Format() Format {
	return FormatMermaidTrust
}

// Render writes the Mermaid diagram to the writer.
func (r *MermaidTrustRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the Mermaid diagram as a string.
func (r *MermaidTrustRenderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *MermaidTrustRenderer) render(p *pidl.Protocol) string {
	var sb strings.Builder

	sb.WriteString("flowchart LR\n")

	if r.Title && p.ProtocolMeta.Name != "" {
		sb.WriteString("    %% " + p.ProtocolMeta.Name + " - Trust Relationships\n\n")
	}

	// Check if we have trust relations
	if p.Metadata == nil || len(p.Metadata.TrustRelations) == 0 {
		sb.WriteString("    %% No trust relationships defined\n")
		return sb.String()
	}

	// Collect all nodes referenced in trust relations
	nodes := r.collectNodes(p)

	// Render nodes (components or entities)
	if r.ShowComponents && p.Metadata != nil && len(p.Metadata.Components) > 0 {
		r.renderComponentNodes(&sb, p, nodes)
	} else {
		r.renderEntityNodes(&sb, p, nodes)
	}

	sb.WriteString("\n")

	// Render trust relationships as edges
	r.renderTrustEdges(&sb, p)

	// Add styling
	r.renderStyles(&sb)

	return sb.String()
}

func (r *MermaidTrustRenderer) collectNodes(p *pidl.Protocol) map[string]bool {
	nodes := make(map[string]bool)
	for _, tr := range p.Metadata.TrustRelations {
		nodes[tr.From] = true
		nodes[tr.To] = true
	}
	return nodes
}

func (r *MermaidTrustRenderer) renderComponentNodes(sb *strings.Builder, p *pidl.Protocol, nodes map[string]bool) {
	// Track which nodes are components
	componentIDs := make(map[string]bool)
	for _, c := range p.Metadata.Components {
		componentIDs[c.ID] = true
	}

	// Render components
	for _, c := range p.Metadata.Components {
		if !nodes[c.ID] {
			continue
		}
		nodeID := r.sanitizeID(c.ID)
		fmt.Fprintf(sb, "    %s[\"%s<br/><small>%s</small>\"]\n", nodeID, c.Name, c.Type)
	}

	// Render entities that are referenced but not in components
	for _, e := range p.Entities {
		if !nodes[e.ID] {
			continue
		}
		if componentIDs[e.ID] {
			continue // It's a component, already rendered
		}
		// Check if entity is in a component
		if p.ComponentForEntity(e.ID) != nil {
			continue // Entity is in a component, skip
		}
		nodeID := r.sanitizeID(e.ID)
		shape := r.entityTypeToShape(e.Type)
		fmt.Fprintf(sb, "    %s%s%s\n", nodeID, shape[0], e.Name+shape[1])
	}
}

func (r *MermaidTrustRenderer) renderEntityNodes(sb *strings.Builder, p *pidl.Protocol, nodes map[string]bool) {
	for _, e := range p.Entities {
		if !nodes[e.ID] {
			continue
		}
		nodeID := r.sanitizeID(e.ID)
		shape := r.entityTypeToShape(e.Type)
		fmt.Fprintf(sb, "    %s%s%s\n", nodeID, shape[0], e.Name+shape[1])
	}
}

func (r *MermaidTrustRenderer) renderTrustEdges(sb *strings.Builder, p *pidl.Protocol) {
	for _, tr := range p.Metadata.TrustRelations {
		fromID := r.sanitizeID(tr.From)
		toID := r.sanitizeID(tr.To)

		label := r.trustTypeLabel(tr.Type)
		if r.ShowCredentials && len(tr.Credentials) > 0 {
			label = fmt.Sprintf("%s<br/><small>%s</small>", label, strings.Join(tr.Credentials, ", "))
		}

		arrow := r.trustTypeToArrow(tr.Type, tr.Mutual)
		fmt.Fprintf(sb, "    %s %s|%s| %s\n", fromID, arrow, label, toID)
	}
}

func (r *MermaidTrustRenderer) renderStyles(sb *strings.Builder) {
	sb.WriteString("\n")
	sb.WriteString("    %% Styling\n")
	sb.WriteString("    classDef trusted fill:#d4edda,stroke:#28a745\n")
	sb.WriteString("    classDef authoritative fill:#cce5ff,stroke:#004085\n")
	sb.WriteString("    classDef untrusted fill:#f8d7da,stroke:#721c24\n")
}

func (r *MermaidTrustRenderer) sanitizeID(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}

func (r *MermaidTrustRenderer) entityTypeToShape(t pidl.EntityType) [2]string {
	switch t {
	case pidl.EntityTypeUser:
		return [2]string{"((", "))"}
	case pidl.EntityTypeBrowser:
		return [2]string{"[", "]"}
	case pidl.EntityTypeClient:
		return [2]string{"[", "]"}
	case pidl.EntityTypeServer, pidl.EntityTypeResourceServer, pidl.EntityTypeAuthorizationServer:
		return [2]string{"[(", ")]"}
	case pidl.EntityTypeAgent, pidl.EntityTypeDelegatedAgent:
		return [2]string{"{{", "}}"}
	case pidl.EntityTypeToolServer, pidl.EntityTypeTool:
		return [2]string{"[/", "/]"}
	case pidl.EntityTypeIdentityProvider, pidl.EntityTypeServiceProvider:
		return [2]string{"([", "])"}
	default:
		return [2]string{"[", "]"}
	}
}

func (r *MermaidTrustRenderer) trustTypeLabel(t string) string {
	switch t {
	case pidl.TrustTypeAuthenticates:
		return "authenticates"
	case pidl.TrustTypeValidates:
		return "validates"
	case pidl.TrustTypeDelegates:
		return "delegates"
	case pidl.TrustTypeAuthorizes:
		return "authorizes"
	case pidl.TrustTypeIssues:
		return "issues"
	case pidl.TrustTypeTrusts:
		return "trusts"
	case pidl.TrustTypeProvisions:
		return "provisions"
	case pidl.TrustTypeAttests:
		return "attests"
	default:
		return t
	}
}

func (r *MermaidTrustRenderer) trustTypeToArrow(t string, mutual bool) string {
	if mutual {
		return "<-->"
	}

	switch t {
	case pidl.TrustTypeIssues:
		return "==>" // Thick arrow for credential issuance
	case pidl.TrustTypeTrusts:
		return "-.->|" // Dashed for trust relationship
	default:
		return "-->"
	}
}
