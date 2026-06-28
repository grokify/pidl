package render

import (
	"fmt"
	"html"
	"io"
	"math"
	"strings"

	"github.com/grokify/pidl"
)

// trustNodePos represents a node's position in the trust diagram.
type trustNodePos struct {
	x, y int
	name string
}

// SVGTrustRenderer renders trust relationship diagrams as SVG.
type SVGTrustRenderer struct {
	// Title includes a title header
	Title bool
	// Theme selects the color scheme ("light", "dark")
	Theme string
	// ShowCredentials shows credential types on edges
	ShowCredentials bool
	// ShowComponents uses component names instead of entity/component IDs
	ShowComponents bool
}

// NewSVGTrust creates a new SVGTrustRenderer with default settings.
func NewSVGTrust() *SVGTrustRenderer {
	return &SVGTrustRenderer{
		Title:           true,
		Theme:           "light",
		ShowCredentials: true,
		ShowComponents:  true,
	}
}

// Format returns the output format identifier.
func (r *SVGTrustRenderer) Format() Format {
	return FormatSVGTrust
}

// Render writes the SVG diagram to the given writer.
func (r *SVGTrustRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

// RenderString returns the SVG diagram as a string.
func (r *SVGTrustRenderer) RenderString(p *pidl.Protocol) (string, error) {
	if p.Metadata == nil || len(p.Metadata.TrustRelations) == 0 {
		return r.renderEmptyDiagram(p), nil
	}
	relations := p.Metadata.TrustRelations

	// Collect all unique nodes (entities and components)
	nodeSet := make(map[string]bool)
	for _, rel := range relations {
		nodeSet[rel.From] = true
		nodeSet[rel.To] = true
	}

	// Convert to slice and create index map
	var nodes []string
	nodeIndex := make(map[string]int)
	for node := range nodeSet {
		nodeIndex[node] = len(nodes)
		nodes = append(nodes, node)
	}

	// Layout configuration
	const (
		padding     = 60
		nodeW       = 140
		nodeH       = 50
		titleHeight = 40
		minWidth    = 600
		minHeight   = 400
	)

	// Calculate circular layout
	numNodes := len(nodes)
	centerX := minWidth / 2
	centerY := (minHeight + titleHeight) / 2
	radius := float64(minWidth/2 - padding - nodeW/2)
	if radius < 100 {
		radius = 100
	}

	// Calculate node positions
	positions := make([]trustNodePos, numNodes)
	for i, node := range nodes {
		angle := 2*math.Pi*float64(i)/float64(numNodes) - math.Pi/2
		x := centerX + int(radius*math.Cos(angle))
		y := centerY + int(radius*math.Sin(angle))

		// Get display name
		name := node
		if r.ShowComponents {
			if comp := p.ComponentByID(node); comp != nil {
				name = comp.Name
			} else if entity := p.EntityByID(node); entity != nil {
				name = entity.Name
			}
		}
		positions[i] = trustNodePos{x: x, y: y, name: name}
	}

	// Calculate SVG dimensions
	width := minWidth
	height := minHeight
	if r.Title {
		height += titleHeight
	}

	var sb strings.Builder

	// SVG header
	sb.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	sb.WriteString("\n")

	// CSS styles
	sb.WriteString(r.generateCSS())

	// Defs for arrow markers
	sb.WriteString(r.generateDefs())

	// Title
	if r.Title {
		sb.WriteString(fmt.Sprintf(`  <text class="title" x="%d" y="%d">%s - Trust Relationships</text>`+"\n",
			width/2, 30, html.EscapeString(p.ProtocolMeta.Name)))
	}

	// Render edges (before nodes so nodes appear on top)
	sb.WriteString("  <!-- Trust relationships -->\n")
	sb.WriteString("  <g class=\"edges\">\n")
	for _, rel := range relations {
		fromIdx := nodeIndex[rel.From]
		toIdx := nodeIndex[rel.To]
		r.renderEdge(&sb, positions[fromIdx], positions[toIdx], rel)
	}
	sb.WriteString("  </g>\n")

	// Render nodes
	sb.WriteString("  <!-- Nodes -->\n")
	sb.WriteString("  <g class=\"nodes\">\n")
	for i, pos := range positions {
		nodeID := nodes[i]
		nodeType := r.getNodeType(p, nodeID)
		r.renderNode(&sb, pos.x, pos.y, nodeW, nodeH, pos.name, nodeType)
	}
	sb.WriteString("  </g>\n")

	sb.WriteString("</svg>\n")
	return sb.String(), nil
}

func (r *SVGTrustRenderer) renderNode(sb *strings.Builder, x, y, w, h int, name, nodeType string) {
	color := r.nodeTypeColor(nodeType)

	// Node box
	sb.WriteString(fmt.Sprintf(`    <rect class="node-box" x="%d" y="%d" width="%d" height="%d" rx="6" fill="%s"/>`,
		x-w/2, y-h/2, w, h, color))
	sb.WriteString("\n")

	// Node name
	displayName := name
	if len(displayName) > 18 {
		displayName = displayName[:15] + "..."
	}
	sb.WriteString(fmt.Sprintf(`    <text class="node-name" x="%d" y="%d">%s</text>`,
		x, y+4, html.EscapeString(displayName)))
	sb.WriteString("\n")
}

func (r *SVGTrustRenderer) renderEdge(sb *strings.Builder, from, to trustNodePos, rel pidl.TrustRelationship) {
	// Calculate edge endpoints (offset from node center to edge)
	dx := float64(to.x - from.x)
	dy := float64(to.y - from.y)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist == 0 {
		return
	}

	// Offset from node centers
	nodeRadius := 35.0 // Approximate node radius
	fromX := float64(from.x) + dx/dist*nodeRadius
	fromY := float64(from.y) + dy/dist*nodeRadius
	toX := float64(to.x) - dx/dist*(nodeRadius+8) // Extra offset for arrow
	toY := float64(to.y) - dy/dist*(nodeRadius+8)

	// Edge line
	edgeClass := r.edgeTypeClass(rel.Type)
	markerEnd := "url(#arrow-trust)"
	if rel.Mutual {
		markerEnd = "url(#arrow-trust-mutual)"
	}

	sb.WriteString(fmt.Sprintf(`    <line class="%s" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f" marker-end="%s"/>`,
		edgeClass, fromX, fromY, toX, toY, markerEnd))
	sb.WriteString("\n")

	// Edge label (relationship type)
	labelX := (from.x + to.x) / 2
	labelY := (from.y + to.y) / 2

	// Offset label perpendicular to edge
	perpX := -dy / dist * 12
	perpY := dx / dist * 12
	labelX += int(perpX)
	labelY += int(perpY)

	sb.WriteString(fmt.Sprintf(`    <text class="edge-label" x="%d" y="%d">%s</text>`,
		labelX, labelY, html.EscapeString(rel.Type)))
	sb.WriteString("\n")

	// Credentials (if enabled and present)
	if r.ShowCredentials && len(rel.Credentials) > 0 {
		credsStr := strings.Join(rel.Credentials, ", ")
		if len(credsStr) > 25 {
			credsStr = credsStr[:22] + "..."
		}
		sb.WriteString(fmt.Sprintf(`    <text class="edge-creds" x="%d" y="%d">%s</text>`,
			labelX, labelY+12, html.EscapeString(credsStr)))
		sb.WriteString("\n")
	}
}

func (r *SVGTrustRenderer) getNodeType(p *pidl.Protocol, nodeID string) string {
	// Check if it's a component
	if comp := p.ComponentByID(nodeID); comp != nil {
		return comp.Type
	}
	// Check if it's an entity
	if entity := p.EntityByID(nodeID); entity != nil {
		return string(entity.Type)
	}
	return "unknown"
}

func (r *SVGTrustRenderer) nodeTypeColor(t string) string {
	colors := map[string]string{
		"idp":                  "#cce5ff",
		"iga":                  "#d4edda",
		"agent_provider":       "#fff3cd",
		"person_server":        "#f8d7da",
		"access_server":        "#d1ecf1",
		"pdp":                  "#e2d5f0",
		"gateway":              "#ffeeba",
		"mcp_client":           "#c3e6cb",
		"mcp_server":           "#bee5eb",
		"resource_api":         "#f5c6cb",
		"spire":                "#d6d8db",
		"user":                 "#fff3cd",
		"agent":                "#c3e6cb",
		"authorization_server": "#cce5ff",
		"resource_server":      "#f5c6cb",
	}
	if c, ok := colors[t]; ok {
		return c
	}
	return "#e9ecef"
}

func (r *SVGTrustRenderer) edgeTypeClass(relType string) string {
	switch relType {
	case "authenticates":
		return "edge-auth"
	case "issues":
		return "edge-issues"
	case "delegates":
		return "edge-delegates"
	case "trusts":
		return "edge-trusts"
	case "authorizes":
		return "edge-authorizes"
	case "validates":
		return "edge-validates"
	case "provisions":
		return "edge-provisions"
	case "attests":
		return "edge-attests"
	default:
		return "edge-default"
	}
}

func (r *SVGTrustRenderer) renderEmptyDiagram(p *pidl.Protocol) string {
	return fmt.Sprintf(`<svg viewBox="0 0 400 100" xmlns="http://www.w3.org/2000/svg">
  <text x="200" y="50" text-anchor="middle" font-family="sans-serif" fill="#666">No trust relationships defined in %s</text>
</svg>
`, html.EscapeString(p.ProtocolMeta.Name))
}

func (r *SVGTrustRenderer) generateDefs() string {
	return `  <defs>
    <marker id="arrow-trust" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M0,0 L0,6 L9,3 z" fill="#4a5568"/>
    </marker>
    <marker id="arrow-trust-mutual" markerWidth="10" markerHeight="10" refX="8" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M0,0 L0,6 L9,3 z" fill="#3182ce"/>
    </marker>
  </defs>
`
}

func (r *SVGTrustRenderer) generateCSS() string {
	textColor := "#1a202c"
	if r.Theme == "dark" {
		textColor = "#e2e8f0"
	}

	return fmt.Sprintf(`  <style>
    .title { font-family: sans-serif; font-size: 16px; font-weight: 600; fill: %s; text-anchor: middle; }
    .node-box { stroke: #4a5568; stroke-width: 2; }
    .node-name { font-family: sans-serif; font-size: 11px; font-weight: 500; fill: %s; text-anchor: middle; }
    .edge-auth { stroke: #3182ce; stroke-width: 2; }
    .edge-issues { stroke: #38a169; stroke-width: 2; stroke-dasharray: 5,3; }
    .edge-delegates { stroke: #d69e2e; stroke-width: 2; }
    .edge-trusts { stroke: #805ad5; stroke-width: 2; stroke-dasharray: 3,3; }
    .edge-authorizes { stroke: #38a169; stroke-width: 2; }
    .edge-validates { stroke: #3182ce; stroke-width: 2; stroke-dasharray: 8,4; }
    .edge-provisions { stroke: #718096; stroke-width: 2; }
    .edge-attests { stroke: #e53e3e; stroke-width: 2; }
    .edge-default { stroke: #718096; stroke-width: 2; }
    .edge-label { font-family: sans-serif; font-size: 9px; fill: #4a5568; text-anchor: middle; }
    .edge-creds { font-family: monospace; font-size: 8px; fill: #718096; text-anchor: middle; }
  </style>
`, textColor, textColor)
}
