package render

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/render/svg"
)

// SVGNetworkRenderer renders PIDL protocols as SVG network boundary diagrams.
type SVGNetworkRenderer struct {
	// Animated enables CSS animations.
	Animated bool

	// Theme selects the color scheme ("light", "dark", "auto").
	Theme string

	// Title includes the protocol name in the diagram.
	Title bool

	// Direction is the layout direction ("horizontal" or "vertical").
	Direction string

	// BoundaryOverrides allows CLI-specified boundary assignments.
	// Map of boundary ID to entity IDs.
	BoundaryOverrides map[string][]string

	// layoutConfig contains layout configuration.
	layoutConfig svg.NetworkLayoutConfig
}

// NewSVGNetwork creates a new SVG network renderer with default options.
func NewSVGNetwork() *SVGNetworkRenderer {
	return &SVGNetworkRenderer{
		Animated:          false,
		Theme:             "light",
		Title:             true,
		Direction:         "horizontal",
		BoundaryOverrides: make(map[string][]string),
		layoutConfig:      svg.DefaultNetworkLayoutConfig(),
	}
}

// Format returns the output format.
func (r *SVGNetworkRenderer) Format() Format {
	return FormatSVGNetwork
}

// Render writes the SVG diagram to the writer.
func (r *SVGNetworkRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the SVG diagram as a string.
func (r *SVGNetworkRenderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

// AddBoundaryOverride adds a CLI-specified boundary assignment.
func (r *SVGNetworkRenderer) AddBoundaryOverride(boundaryID string, entityIDs []string) {
	r.BoundaryOverrides[boundaryID] = append(r.BoundaryOverrides[boundaryID], entityIDs...)
}

func (r *SVGNetworkRenderer) render(p *pidl.Protocol) string {
	// Apply direction from protocol metadata if not overridden
	direction := r.Direction
	if p.Metadata != nil && p.Metadata.NetworkLayout != nil && p.Metadata.NetworkLayout.Direction != "" {
		direction = p.Metadata.NetworkLayout.Direction
	}
	r.layoutConfig.Direction = direction

	// Resolve boundaries
	boundaries := svg.ResolveBoundaries(p, r.BoundaryOverrides)

	// Build entity name map
	entityNames := make(map[string]string)
	for _, e := range p.Entities {
		entityNames[e.ID] = e.Name
	}

	// Build entity to boundary map for connection analysis
	entityToBoundary := make(map[string]string)
	for _, b := range boundaries {
		for _, entityID := range b.Entities {
			entityToBoundary[entityID] = b.ID
		}
	}

	// Aggregate connections
	connections := svg.AggregateConnections(p, entityToBoundary)

	// Calculate layout
	layout := svg.CalculateNetworkLayout(boundaries, connections, r.layoutConfig)

	// Build SVG
	var sb strings.Builder

	// SVG header
	themeClass := svg.ThemeClass(svg.Theme(r.Theme))
	sb.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg"`, layout.Width, layout.Height))
	if themeClass != "" {
		sb.WriteString(fmt.Sprintf(` class="%s"`, themeClass))
	}
	sb.WriteString(">\n")

	// Embedded CSS
	sb.WriteString("  <style>\n")
	sb.WriteString(svg.GenerateNetworkCSS(svg.Theme(r.Theme)))
	sb.WriteString("  </style>\n\n")

	// Defs for arrow markers
	sb.WriteString(r.generateDefs())

	// Title (if enabled)
	if r.Title && p.ProtocolMeta.Name != "" {
		sb.WriteString(fmt.Sprintf(`  <text x="%d" y="15" class="message-text" style="font-size:14px;font-weight:600;text-anchor:middle;">%s</text>`+"\n\n",
			layout.Width/2, html.EscapeString(p.ProtocolMeta.Name)))
	}

	// Boundaries
	sb.WriteString("  <!-- Network Boundaries -->\n")
	sb.WriteString("  <g class=\"boundaries\">\n")
	for _, bl := range layout.Boundaries {
		r.renderBoundary(&sb, bl)
	}
	sb.WriteString("  </g>\n\n")

	// Entities
	sb.WriteString("  <!-- Entities -->\n")
	sb.WriteString("  <g class=\"entities\">\n")
	for _, bl := range layout.Boundaries {
		for _, el := range bl.Entities {
			name := entityNames[el.ID]
			if name == "" {
				name = el.ID
			}
			r.renderEntity(&sb, el, name)
		}
	}
	sb.WriteString("  </g>\n\n")

	// Connections
	sb.WriteString("  <!-- Connections -->\n")
	sb.WriteString("  <g class=\"connections\">\n")
	for _, conn := range layout.Connections {
		r.renderConnection(&sb, conn)
	}
	sb.WriteString("  </g>\n")

	sb.WriteString("</svg>\n")

	return sb.String()
}

func (r *SVGNetworkRenderer) renderBoundary(sb *strings.Builder, bl svg.BoundaryLayout) {
	styleClass := fmt.Sprintf("boundary boundary-%s", bl.Style)

	// Boundary rectangle
	sb.WriteString(fmt.Sprintf(`    <rect class="%s" x="%d" y="%d" width="%d" height="%d" rx="8"/>`+"\n",
		styleClass, bl.X, bl.Y, bl.Width, bl.Height))

	// Boundary label
	labelX := bl.X + 10
	labelY := bl.Y + 18
	sb.WriteString(fmt.Sprintf(`    <text class="boundary-label" x="%d" y="%d">%s</text>`+"\n",
		labelX, labelY, html.EscapeString(bl.Name)))
}

func (r *SVGNetworkRenderer) renderEntity(sb *strings.Builder, el svg.EntityLayout, name string) {
	boxX := el.X - el.Width/2
	boxY := el.Y - el.Height/2

	sb.WriteString(fmt.Sprintf(`    <rect class="entity-box" x="%d" y="%d" width="%d" height="%d" rx="4"/>`+"\n",
		boxX, boxY, el.Width, el.Height))
	sb.WriteString(fmt.Sprintf(`    <text class="entity-text" x="%d" y="%d">%s</text>`+"\n",
		el.X, el.Y, html.EscapeString(name)))
}

func (r *SVGNetworkRenderer) renderConnection(sb *strings.Builder, conn svg.ConnectionLayout) {
	if conn.FromX == 0 && conn.FromY == 0 {
		return // Skip if no position data
	}

	connClass := "connection"
	if conn.CrossesBoundary {
		connClass = "connection connection-cross-boundary"
	}

	// Draw curved path
	path := r.generateConnectionPath(conn.FromX, conn.FromY, conn.ToX, conn.ToY)
	sb.WriteString(fmt.Sprintf(`    <path class="%s" d="%s" marker-end="url(#arrow-connection)"/>`+"\n",
		connClass, path))

	// Connection label
	if conn.Label != "" {
		midX := (conn.FromX + conn.ToX) / 2
		midY := (conn.FromY + conn.ToY) / 2
		sb.WriteString(fmt.Sprintf(`    <text class="connection-label" x="%d" y="%d" text-anchor="middle">%s</text>`+"\n",
			midX, midY-5, html.EscapeString(conn.Label)))
	}
}

func (r *SVGNetworkRenderer) generateConnectionPath(fromX, fromY, toX, toY int) string {
	// Use a simple curved path
	midX := (fromX + toX) / 2

	// If vertical layout, curve horizontally
	if r.Direction == "vertical" {
		midY := (fromY + toY) / 2
		return fmt.Sprintf("M%d,%d Q%d,%d %d,%d", fromX, fromY, midX, midY, toX, toY)
	}

	// Horizontal layout - curve vertically
	return fmt.Sprintf("M%d,%d Q%d,%d %d,%d", fromX, fromY, midX, fromY, midX, toY)
}

func (r *SVGNetworkRenderer) generateDefs() string {
	return `  <defs>
    <marker id="arrow-connection" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M0,0 L0,6 L9,3 z" fill="currentColor"/>
    </marker>
  </defs>

`
}
