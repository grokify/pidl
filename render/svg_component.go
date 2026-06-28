package render

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// SVGComponentRenderer renders deployment component diagrams as SVG.
type SVGComponentRenderer struct {
	// Title includes a title header
	Title bool
	// Theme selects the color scheme ("light", "dark")
	Theme string
	// ShowEntities shows entities within components
	ShowEntities bool
	// ShowRoles shows protocol roles on components
	ShowRoles bool
}

// NewSVGComponent creates a new SVGComponentRenderer with default settings.
func NewSVGComponent() *SVGComponentRenderer {
	return &SVGComponentRenderer{
		Title:        true,
		Theme:        "light",
		ShowEntities: true,
		ShowRoles:    true,
	}
}

// Format returns the output format identifier.
func (r *SVGComponentRenderer) Format() Format {
	return FormatSVGComponent
}

// Render writes the SVG diagram to the given writer.
func (r *SVGComponentRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

// RenderString returns the SVG diagram as a string.
func (r *SVGComponentRenderer) RenderString(p *pidl.Protocol) (string, error) {
	if p.Metadata == nil || len(p.Metadata.Components) == 0 {
		return r.renderEmptyDiagram(p), nil
	}
	components := p.Metadata.Components

	// Layout configuration
	const (
		padding       = 40
		componentW    = 220
		componentH    = 120
		componentGap  = 40
		entityH       = 24
		entityPadding = 8
		titleHeight   = 30
		columnsPerRow = 3
	)

	// Calculate grid dimensions
	numComponents := len(components)
	rows := (numComponents + columnsPerRow - 1) / columnsPerRow
	cols := columnsPerRow
	if numComponents < columnsPerRow {
		cols = numComponents
	}

	// Calculate SVG dimensions
	width := padding*2 + cols*componentW + (cols-1)*componentGap
	height := padding*2 + rows*componentH + (rows-1)*componentGap
	if r.Title {
		height += titleHeight
	}

	var sb strings.Builder

	// SVG header
	sb.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	sb.WriteString("\n")

	// CSS styles
	sb.WriteString(r.generateCSS())

	// Title
	startY := padding
	if r.Title {
		sb.WriteString(fmt.Sprintf(`  <text class="title" x="%d" y="%d">%s - Components</text>`+"\n",
			width/2, padding, html.EscapeString(p.ProtocolMeta.Name)))
		startY += titleHeight
	}

	// Render components
	for i, comp := range components {
		row := i / columnsPerRow
		col := i % columnsPerRow

		x := padding + col*(componentW+componentGap)
		y := startY + row*(componentH+componentGap)

		r.renderComponent(&sb, comp, x, y, componentW, componentH)
	}

	sb.WriteString("</svg>\n")
	return sb.String(), nil
}

func (r *SVGComponentRenderer) renderComponent(sb *strings.Builder, comp pidl.DeploymentComponent, x, y, w, h int) {
	// Component background
	typeColor := r.componentTypeColor(comp.Type)
	sb.WriteString(fmt.Sprintf(`  <rect class="component-box" x="%d" y="%d" width="%d" height="%d" rx="8" fill="%s"/>`,
		x, y, w, h, typeColor))
	sb.WriteString("\n")

	// Component name
	sb.WriteString(fmt.Sprintf(`  <text class="component-name" x="%d" y="%d">%s</text>`,
		x+w/2, y+20, html.EscapeString(comp.Name)))
	sb.WriteString("\n")

	// Component type badge
	sb.WriteString(fmt.Sprintf(`  <text class="component-type" x="%d" y="%d">[%s]</text>`,
		x+w/2, y+36, html.EscapeString(comp.Type)))
	sb.WriteString("\n")

	// Protocol roles (if enabled and present)
	if r.ShowRoles && len(comp.Implements) > 0 {
		roles := make([]string, len(comp.Implements))
		for i, impl := range comp.Implements {
			roles[i] = fmt.Sprintf("%s:%s", impl.Protocol, impl.Role)
		}
		roleStr := strings.Join(roles, ", ")
		if len(roleStr) > 35 {
			roleStr = roleStr[:32] + "..."
		}
		sb.WriteString(fmt.Sprintf(`  <text class="component-roles" x="%d" y="%d">%s</text>`,
			x+w/2, y+54, html.EscapeString(roleStr)))
		sb.WriteString("\n")
	}

	// Entities (if enabled and present)
	if r.ShowEntities && len(comp.Entities) > 0 {
		entitiesStr := strings.Join(comp.Entities, ", ")
		if len(entitiesStr) > 30 {
			entitiesStr = entitiesStr[:27] + "..."
		}
		sb.WriteString(fmt.Sprintf(`  <text class="component-entities" x="%d" y="%d">Entities: %s</text>`,
			x+w/2, y+h-20, html.EscapeString(entitiesStr)))
		sb.WriteString("\n")
	}

	// Examples (if present)
	if len(comp.Examples) > 0 {
		examplesStr := strings.Join(comp.Examples, ", ")
		if len(examplesStr) > 30 {
			examplesStr = examplesStr[:27] + "..."
		}
		sb.WriteString(fmt.Sprintf(`  <text class="component-examples" x="%d" y="%d">e.g., %s</text>`,
			x+w/2, y+h-6, html.EscapeString(examplesStr)))
		sb.WriteString("\n")
	}
}

func (r *SVGComponentRenderer) componentTypeColor(t string) string {
	colors := map[string]string{
		"idp":            "#cce5ff", // blue
		"iga":            "#d4edda", // green
		"agent_provider": "#fff3cd", // yellow
		"person_server":  "#f8d7da", // red
		"access_server":  "#d1ecf1", // cyan
		"pdp":            "#e2d5f0", // purple
		"gateway":        "#ffeeba", // orange
		"mcp_client":     "#c3e6cb", // light green
		"mcp_server":     "#bee5eb", // light blue
		"resource_api":   "#f5c6cb", // light red
		"spire":          "#d6d8db", // gray
	}
	if c, ok := colors[t]; ok {
		return c
	}
	return "#e9ecef" // default gray
}

func (r *SVGComponentRenderer) renderEmptyDiagram(p *pidl.Protocol) string {
	return fmt.Sprintf(`<svg viewBox="0 0 400 100" xmlns="http://www.w3.org/2000/svg">
  <text x="200" y="50" text-anchor="middle" font-family="sans-serif" fill="#666">No components defined in %s</text>
</svg>
`, html.EscapeString(p.ProtocolMeta.Name))
}

func (r *SVGComponentRenderer) generateCSS() string {
	textColor := "#1a202c"
	if r.Theme == "dark" {
		textColor = "#e2e8f0"
	}

	return fmt.Sprintf(`  <style>
    .title { font-family: sans-serif; font-size: 16px; font-weight: 600; fill: %s; text-anchor: middle; }
    .component-box { stroke: #4a5568; stroke-width: 2; }
    .component-name { font-family: sans-serif; font-size: 13px; font-weight: 600; fill: %s; text-anchor: middle; }
    .component-type { font-family: sans-serif; font-size: 10px; fill: #718096; text-anchor: middle; }
    .component-roles { font-family: monospace; font-size: 9px; fill: #4a5568; text-anchor: middle; }
    .component-entities { font-family: sans-serif; font-size: 9px; fill: #718096; text-anchor: middle; }
    .component-examples { font-family: sans-serif; font-size: 8px; font-style: italic; fill: #a0aec0; text-anchor: middle; }
  </style>
`, textColor, textColor)
}
