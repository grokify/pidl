// Package render provides diagram rendering for PIDL protocols.
package render

import (
	"fmt"
	"strings"

	"github.com/grokify/pidl"
)

// InfographicSize defines preset canvas sizes.
type InfographicSize string

const (
	// SizeLinkedInSquare is 1200x1200 for LinkedIn feed posts.
	SizeLinkedInSquare InfographicSize = "linkedin-square"
	// SizeLinkedInPortrait is 1080x1350 for LinkedIn portrait posts.
	SizeLinkedInPortrait InfographicSize = "linkedin-portrait"
	// SizeLinkedInLandscape is 1200x627 for LinkedIn link previews.
	SizeLinkedInLandscape InfographicSize = "linkedin-landscape"
	// SizeDatasheetTile is 400x400 for datasheet grid layouts.
	SizeDatasheetTile InfographicSize = "datasheet-tile"
	// SizeDatasheetWide is 600x300 for wide datasheet tiles.
	SizeDatasheetWide InfographicSize = "datasheet-wide"
)

// InfographicTheme defines color themes.
type InfographicTheme string

const (
	ThemeBold    InfographicTheme = "bold"    // High contrast, saturated
	ThemeMinimal InfographicTheme = "minimal" // Clean, subtle
	ThemeDark    InfographicTheme = "dark"    // Dark background
	ThemeTech    InfographicTheme = "tech"    // Tech/engineering feel
)

// NodeShape defines the shape of a node.
type NodeShape string

const (
	ShapeRectangle NodeShape = "rectangle" // Rounded rectangle (default)
	ShapeCircle    NodeShape = "circle"    // Circle for users/humans
	ShapeHexagon   NodeShape = "hexagon"   // Hexagon for agents/LLMs
	ShapeCylinder  NodeShape = "cylinder"  // Cylinder for databases/resources
	ShapeCloud     NodeShape = "cloud"     // Cloud for external services
	ShapeDiamond   NodeShape = "diamond"   // Diamond for tools/decisions
)

// InfographicOptions configures the infographic renderer.
type InfographicOptions struct {
	// Canvas size preset
	Size InfographicSize

	// Custom dimensions (override Size preset)
	Width  int
	Height int

	// Theme
	Theme InfographicTheme

	// Title to display at top (optional)
	Title string

	// Animation
	AnimateDots       bool    // Enable animated dots on edges
	DotSpeed          float64 // Animation duration in seconds
	DotsPerEdge       int     // Number of dots per edge
	BidirectionalDots bool    // Dots go both directions

	// Layout
	Direction string // "horizontal" or "vertical"
	Padding   int    // Padding around content

	// Simplification
	MaxNodes       int  // Warn if exceeded
	TruncateLabels int  // Max label chars (0 = no truncation)
	ShowLabels     bool // Show node labels
	ShowEdgeLabels bool // Show edge labels

	// Styling
	NodeRadius      int  // Corner radius for rectangle nodes
	StrokeWidth     int  // Line thickness
	FontSize        int  // Base font size
	TitleFontSize   int  // Title font size
	UseCustomShapes bool // Use shape based on entity/step type
}

// DefaultInfographicOptions returns sensible defaults for LinkedIn.
func DefaultInfographicOptions() InfographicOptions {
	return InfographicOptions{
		Size:              SizeLinkedInSquare,
		Theme:             ThemeBold,
		AnimateDots:       true,
		DotSpeed:          2.0,
		DotsPerEdge:       3,
		BidirectionalDots: false,
		Direction:         "horizontal",
		Padding:           60,
		MaxNodes:          7,
		TruncateLabels:    12,
		ShowLabels:        true,
		ShowEdgeLabels:    false,
		NodeRadius:        12,
		StrokeWidth:       3,
		FontSize:          18,
		TitleFontSize:     28,
		UseCustomShapes:   true,
	}
}

// DatasheetTileOptions returns options optimized for datasheet tiles.
func DatasheetTileOptions() InfographicOptions {
	opts := DefaultInfographicOptions()
	opts.Size = SizeDatasheetTile
	opts.Padding = 30
	opts.FontSize = 12
	opts.TitleFontSize = 16
	opts.TruncateLabels = 8
	opts.StrokeWidth = 2
	opts.NodeRadius = 8
	opts.DotsPerEdge = 2
	return opts
}

// InfographicRenderer renders process specs as compact infographics.
type InfographicRenderer struct {
	opts InfographicOptions
}

// NewInfographicRenderer creates a new infographic renderer.
func NewInfographicRenderer(opts InfographicOptions) *InfographicRenderer {
	return &InfographicRenderer{opts: opts}
}

// Render generates an SVG infographic from a protocol.
func (r *InfographicRenderer) Render(p *pidl.Protocol) string {
	width, height := r.getCanvasSize()
	theme := r.getTheme()

	var sb strings.Builder

	// SVG header
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
`, width, height, width, height))

	// Styles
	sb.WriteString(r.renderStyles(theme))

	// Defs (gradients, filters, markers)
	sb.WriteString(r.renderDefs(theme))

	// Background
	sb.WriteString(fmt.Sprintf(`  <rect width="%d" height="%d" fill="%s"/>
`, width, height, theme.Background))

	// Title
	if r.opts.Title != "" {
		sb.WriteString(r.renderTitle(width))
	}

	// Get process steps
	steps := r.getSteps(p)
	if len(steps) == 0 {
		sb.WriteString(`  <text x="50%" y="50%" text-anchor="middle" class="label">No process steps</text>
`)
		sb.WriteString("</svg>")
		return sb.String()
	}

	// Calculate layout
	nodes := r.calculateLayout(steps, width, height)

	// Render edges first (behind nodes)
	sb.WriteString(r.renderEdges(p, nodes))

	// Render nodes
	sb.WriteString(r.renderNodes(nodes, theme))

	sb.WriteString("</svg>")
	return sb.String()
}

// infographicNode represents a positioned node.
type infographicNode struct {
	ID         string
	Label      string
	Icon       string
	Shape      NodeShape
	StepType   pidl.StepType
	EntityType pidl.EntityType
	X          int
	Y          int
	Width      int
	Height     int
}

// infographicTheme holds theme colors.
type infographicTheme struct {
	Background   string
	NodeFill     string
	NodeStroke   string
	TextColor    string
	TitleColor   string
	EdgeColor    string
	DotColor     string
	ShadowColor  string
	StepColors   map[pidl.StepType]string
	EntityColors map[pidl.EntityType]string
}

func (r *InfographicRenderer) getCanvasSize() (int, int) {
	if r.opts.Width > 0 && r.opts.Height > 0 {
		return r.opts.Width, r.opts.Height
	}

	switch r.opts.Size {
	case SizeLinkedInSquare:
		return 1200, 1200
	case SizeLinkedInPortrait:
		return 1080, 1350
	case SizeLinkedInLandscape:
		return 1200, 627
	case SizeDatasheetTile:
		return 400, 400
	case SizeDatasheetWide:
		return 600, 300
	default:
		return 1200, 1200
	}
}

func (r *InfographicRenderer) getTheme() infographicTheme {
	stepColors := map[pidl.StepType]string{
		pidl.StepTypeDeterministic: "#1976D2", // Blue
		pidl.StepTypeLLM:           "#7B1FA2", // Purple
		pidl.StepTypeHuman:         "#388E3C", // Green
		pidl.StepTypeExternal:      "#F57C00", // Orange
		pidl.StepTypeTool:          "#607D8B", // Gray
	}

	entityColors := map[pidl.EntityType]string{
		pidl.EntityTypeClient:              "#2196F3", // Blue
		pidl.EntityTypeUser:                "#4CAF50", // Green
		pidl.EntityTypeBrowser:             "#4CAF50", // Green (same as user)
		pidl.EntityTypeAgent:               "#9C27B0", // Purple
		pidl.EntityTypeDelegatedAgent:      "#7B1FA2", // Deep Purple
		pidl.EntityTypeAuthorizationServer: "#F44336", // Red
		pidl.EntityTypeResourceServer:      "#FF9800", // Orange
		pidl.EntityTypeServer:              "#607D8B", // Blue-Gray
		pidl.EntityTypeIdentityProvider:    "#3F51B5", // Indigo
		pidl.EntityTypeServiceProvider:     "#009688", // Teal
		pidl.EntityTypeToolServer:          "#795548", // Brown
		pidl.EntityTypeTool:                "#9E9E9E", // Gray
		pidl.EntityTypeOther:               "#757575", // Dark Gray
	}

	switch r.opts.Theme {
	case ThemeDark:
		return infographicTheme{
			Background:   "#1a1a2e",
			NodeFill:     "#16213e",
			NodeStroke:   "#0f3460",
			TextColor:    "#eaeaea",
			TitleColor:   "#ffffff",
			EdgeColor:    "#4a5568",
			DotColor:     "#00d9ff",
			ShadowColor:  "rgba(0,0,0,0.5)",
			StepColors:   stepColors,
			EntityColors: entityColors,
		}
	case ThemeMinimal:
		return infographicTheme{
			Background:   "#ffffff",
			NodeFill:     "#f8f9fa",
			NodeStroke:   "#dee2e6",
			TextColor:    "#495057",
			TitleColor:   "#212529",
			EdgeColor:    "#adb5bd",
			DotColor:     "#6c757d",
			ShadowColor:  "rgba(0,0,0,0.1)",
			StepColors:   stepColors,
			EntityColors: entityColors,
		}
	case ThemeTech:
		return infographicTheme{
			Background:   "#0d1117",
			NodeFill:     "#161b22",
			NodeStroke:   "#30363d",
			TextColor:    "#c9d1d9",
			TitleColor:   "#58a6ff",
			EdgeColor:    "#484f58",
			DotColor:     "#39d353",
			ShadowColor:  "rgba(0,0,0,0.6)",
			StepColors:   stepColors,
			EntityColors: entityColors,
		}
	default: // ThemeBold
		return infographicTheme{
			Background:   "#ffffff",
			NodeFill:     "#ffffff",
			NodeStroke:   "#e0e0e0",
			TextColor:    "#333333",
			TitleColor:   "#1a1a1a",
			EdgeColor:    "#90a4ae",
			DotColor:     "#7B1FA2",
			ShadowColor:  "rgba(0,0,0,0.15)",
			StepColors:   stepColors,
			EntityColors: entityColors,
		}
	}
}

func (r *InfographicRenderer) renderStyles(theme infographicTheme) string {
	return fmt.Sprintf(`  <style>
    .node { filter: drop-shadow(2px 4px 6px %s); }
    .label { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; font-weight: 600; fill: %s; }
    .title { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; font-weight: 700; fill: %s; }
    .icon { font-size: %dpx; }
    .edge { fill: none; stroke: %s; stroke-width: %d; stroke-linecap: round; }
    .dot { fill: %s; }
  </style>
`, theme.ShadowColor, theme.TextColor, theme.TitleColor, r.opts.FontSize+8, theme.EdgeColor, r.opts.StrokeWidth, theme.DotColor)
}

func (r *InfographicRenderer) renderDefs(theme infographicTheme) string {
	var sb strings.Builder
	sb.WriteString("  <defs>\n")

	// Arrow marker
	sb.WriteString(fmt.Sprintf(`    <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
      <polygon points="0 0, 10 3.5, 0 7" fill="%s"/>
    </marker>
`, theme.EdgeColor))

	// Glow filter for dots
	sb.WriteString(`    <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
      <feMerge>
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
`)

	sb.WriteString("  </defs>\n")
	return sb.String()
}

func (r *InfographicRenderer) renderTitle(width int) string {
	return fmt.Sprintf(`  <text x="%d" y="%d" text-anchor="middle" class="title" font-size="%d">%s</text>
`, width/2, r.opts.Padding, r.opts.TitleFontSize, escapeXML(r.opts.Title))
}

func (r *InfographicRenderer) getSteps(p *pidl.Protocol) []pidl.Entity {
	var steps []pidl.Entity
	for _, e := range p.Entities {
		if e.IsProcessStep() {
			steps = append(steps, e)
		}
	}

	// If no process steps, use all entities
	if len(steps) == 0 {
		steps = p.Entities
	}

	// Limit to MaxNodes
	if r.opts.MaxNodes > 0 && len(steps) > r.opts.MaxNodes {
		steps = steps[:r.opts.MaxNodes]
	}

	return steps
}

func (r *InfographicRenderer) calculateLayout(steps []pidl.Entity, canvasWidth, canvasHeight int) []infographicNode {
	nodes := make([]infographicNode, len(steps))

	// Calculate content area
	titleOffset := 0
	if r.opts.Title != "" {
		titleOffset = r.opts.TitleFontSize + 20
	}

	contentX := r.opts.Padding
	contentY := r.opts.Padding + titleOffset
	contentWidth := canvasWidth - (2 * r.opts.Padding)
	contentHeight := canvasHeight - (2 * r.opts.Padding) - titleOffset

	// Node dimensions
	nodeWidth := 120
	nodeHeight := 80

	// Adjust for small canvases
	if canvasWidth <= 500 {
		nodeWidth = 80
		nodeHeight = 60
	}

	n := len(steps)

	if r.opts.Direction == "vertical" {
		// Vertical layout
		spacing := contentHeight / (n + 1)
		for i, step := range steps {
			nodes[i] = infographicNode{
				ID:         step.ID,
				Label:      r.truncateLabel(step.Name),
				Icon:       nodeIcon(step),
				Shape:      r.nodeShape(step),
				StepType:   step.StepType,
				EntityType: step.Type,
				X:          contentX + contentWidth/2 - nodeWidth/2,
				Y:          contentY + spacing*(i+1) - nodeHeight/2,
				Width:      nodeWidth,
				Height:     nodeHeight,
			}
		}
	} else {
		// Horizontal layout (default)
		spacing := contentWidth / (n + 1)
		for i, step := range steps {
			nodes[i] = infographicNode{
				ID:         step.ID,
				Label:      r.truncateLabel(step.Name),
				Icon:       nodeIcon(step),
				Shape:      r.nodeShape(step),
				StepType:   step.StepType,
				EntityType: step.Type,
				X:          contentX + spacing*(i+1) - nodeWidth/2,
				Y:          contentY + contentHeight/2 - nodeHeight/2,
				Width:      nodeWidth,
				Height:     nodeHeight,
			}
		}
	}

	return nodes
}

func (r *InfographicRenderer) truncateLabel(label string) string {
	if r.opts.TruncateLabels <= 0 || len(label) <= r.opts.TruncateLabels {
		return label
	}
	return label[:r.opts.TruncateLabels-1] + "…"
}

func (r *InfographicRenderer) renderNodes(nodes []infographicNode, theme infographicTheme) string {
	var sb strings.Builder
	sb.WriteString("  <!-- Nodes -->\n")

	for _, node := range nodes {
		// Get node color: prefer step type, fall back to entity type, then default
		nodeColor := theme.StepColors[node.StepType]
		if nodeColor == "" {
			nodeColor = theme.EntityColors[node.EntityType]
		}
		if nodeColor == "" {
			nodeColor = theme.NodeStroke
		}

		// Node group
		sb.WriteString(fmt.Sprintf(`  <g class="node" transform="translate(%d,%d)">
`, node.X, node.Y))

		// Render shape based on node.Shape
		sb.WriteString(r.renderNodeShape(node, theme.NodeFill, nodeColor))

		// Icon - adjust position for different shapes
		iconY := node.Height/2 - 5
		if node.Shape == ShapeCylinder {
			iconY = node.Height / 2 // Center for cylinder
		}
		sb.WriteString(fmt.Sprintf(`    <text x="%d" y="%d" text-anchor="middle" class="icon">%s</text>
`, node.Width/2, iconY, node.Icon))

		// Label
		if r.opts.ShowLabels {
			labelY := node.Height - 12
			if node.Shape == ShapeCylinder {
				labelY = node.Height - 8
			} else if node.Shape == ShapeCircle {
				labelY = node.Height - 8
			}
			sb.WriteString(fmt.Sprintf(`    <text x="%d" y="%d" text-anchor="middle" class="label" font-size="%d">%s</text>
`, node.Width/2, labelY, r.opts.FontSize, escapeXML(node.Label)))
		}

		sb.WriteString("  </g>\n")
	}

	return sb.String()
}

func (r *InfographicRenderer) renderNodeShape(node infographicNode, fill, stroke string) string {
	w := node.Width
	h := node.Height
	sw := r.opts.StrokeWidth

	switch node.Shape {
	case ShapeCircle:
		// Ellipse that fits in the bounding box
		cx := w / 2
		cy := h / 2
		rx := (w - sw) / 2
		ry := (h - sw) / 2
		return fmt.Sprintf(`    <ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s" stroke="%s" stroke-width="%d"/>
`, cx, cy, rx, ry, fill, stroke, sw)

	case ShapeHexagon:
		// Pointy-top hexagon
		cx := float64(w) / 2
		cy := float64(h) / 2
		rx := float64(w-sw) / 2
		ry := float64(h-sw) / 2
		// Six points for hexagon (pointy top)
		points := fmt.Sprintf("%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f",
			cx, cy-ry, // top
			cx+rx, cy-ry*0.5, // top-right
			cx+rx, cy+ry*0.5, // bottom-right
			cx, cy+ry, // bottom
			cx-rx, cy+ry*0.5, // bottom-left
			cx-rx, cy-ry*0.5, // top-left
		)
		return fmt.Sprintf(`    <polygon points="%s" fill="%s" stroke="%s" stroke-width="%d"/>
`, points, fill, stroke, sw)

	case ShapeCylinder:
		// Cylinder (3D database shape)
		ellipseH := h / 6 // Height of ellipse caps
		bodyH := h - ellipseH
		return fmt.Sprintf(`    <ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s" stroke="%s" stroke-width="%d"/>
    <rect x="%d" y="%d" width="%d" height="%d" fill="%s" stroke="none"/>
    <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="%d"/>
    <line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="%d"/>
    <ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s" stroke="%s" stroke-width="%d"/>
`,
			w/2, ellipseH/2, (w-sw)/2, ellipseH/2, fill, stroke, sw, // top ellipse
			sw/2, ellipseH/2, w-sw, bodyH, fill, // body rect (no stroke)
			sw/2, ellipseH/2, sw/2, h-ellipseH/2, stroke, sw, // left line
			w-sw/2, ellipseH/2, w-sw/2, h-ellipseH/2, stroke, sw, // right line
			w/2, h-ellipseH/2, (w-sw)/2, ellipseH/2, fill, stroke, sw, // bottom ellipse
		)

	case ShapeCloud:
		// Simplified cloud shape using path
		return fmt.Sprintf(`    <path d="M%d,%d
      Q%d,%d %d,%d
      Q%d,%d %d,%d
      Q%d,%d %d,%d
      Q%d,%d %d,%d
      Q%d,%d %d,%d
      Z" fill="%s" stroke="%s" stroke-width="%d"/>
`,
			w/5, h*2/3, // start left
			0, h*2/3, w/10, h/3, // left bump
			0, 0, w/2, h/6, // top-left bump
			w, 0, w*9/10, h/3, // top-right bump
			w, h*2/3, w*4/5, h*2/3, // right bump
			w/2, h, w/5, h*2/3, // bottom
			fill, stroke, sw)

	case ShapeDiamond:
		// Diamond/rhombus shape
		cx := w / 2
		cy := h / 2
		points := fmt.Sprintf("%d,%d %d,%d %d,%d %d,%d",
			cx, sw/2, // top
			w-sw/2, cy, // right
			cx, h-sw/2, // bottom
			sw/2, cy, // left
		)
		return fmt.Sprintf(`    <polygon points="%s" fill="%s" stroke="%s" stroke-width="%d"/>
`, points, fill, stroke, sw)

	default: // ShapeRectangle
		return fmt.Sprintf(`    <rect width="%d" height="%d" rx="%d" fill="%s" stroke="%s" stroke-width="%d"/>
`, w, h, r.opts.NodeRadius, fill, stroke, sw)
	}
}

func (r *InfographicRenderer) renderEdges(p *pidl.Protocol, nodes []infographicNode) string {
	var sb strings.Builder
	sb.WriteString("  <!-- Edges -->\n")

	// Build node lookup
	nodeMap := make(map[string]*infographicNode)
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	edgeIndex := 0
	for _, flow := range p.Flows {
		fromNode := nodeMap[flow.From]
		toNode := nodeMap[flow.To]

		if fromNode == nil || toNode == nil {
			continue
		}

		// Calculate edge path
		path := r.calculateEdgePath(fromNode, toNode)

		// Draw edge
		sb.WriteString(fmt.Sprintf(`  <path id="edge%d" class="edge" d="%s" marker-end="url(#arrowhead)"/>
`, edgeIndex, path))

		// Animated dots
		if r.opts.AnimateDots {
			sb.WriteString(r.renderAnimatedDots(edgeIndex))
		}

		edgeIndex++
	}

	return sb.String()
}

func (r *InfographicRenderer) calculateEdgePath(from, to *infographicNode) string {
	// Calculate connection points
	var x1, y1, x2, y2 int

	if r.opts.Direction == "vertical" {
		// Connect bottom of from to top of to
		x1 = from.X + from.Width/2
		y1 = from.Y + from.Height
		x2 = to.X + to.Width/2
		y2 = to.Y
	} else {
		// Connect right of from to left of to
		x1 = from.X + from.Width
		y1 = from.Y + from.Height/2
		x2 = to.X
		y2 = to.Y + to.Height/2
	}

	// Calculate control points for bezier curve
	dx := x2 - x1
	dy := y2 - y1

	var cx1, cy1, cx2, cy2 int
	if r.opts.Direction == "vertical" {
		cx1 = x1
		cy1 = y1 + dy/3
		cx2 = x2
		cy2 = y2 - dy/3
	} else {
		cx1 = x1 + dx/3
		cy1 = y1
		cx2 = x2 - dx/3
		cy2 = y2
	}

	return fmt.Sprintf("M%d,%d C%d,%d %d,%d %d,%d", x1, y1, cx1, cy1, cx2, cy2, x2, y2)
}

func (r *InfographicRenderer) renderAnimatedDots(edgeIndex int) string {
	var sb strings.Builder

	dotRadius := 5
	if r.opts.Size == SizeDatasheetTile || r.opts.Size == SizeDatasheetWide {
		dotRadius = 3
	}

	for i := 0; i < r.opts.DotsPerEdge; i++ {
		// Stagger start times
		delay := float64(i) * (r.opts.DotSpeed / float64(r.opts.DotsPerEdge))

		sb.WriteString(fmt.Sprintf(`  <circle r="%d" class="dot" filter="url(#glow)">
    <animateMotion dur="%.1fs" repeatCount="indefinite" begin="%.1fs">
      <mpath href="#edge%d"/>
    </animateMotion>
  </circle>
`, dotRadius, r.opts.DotSpeed, delay, edgeIndex))

		// Bidirectional dots (reverse direction)
		if r.opts.BidirectionalDots {
			sb.WriteString(fmt.Sprintf(`  <circle r="%d" class="dot" filter="url(#glow)" opacity="0.6">
    <animateMotion dur="%.1fs" repeatCount="indefinite" begin="%.1fs" keyPoints="1;0" keyTimes="0;1" calcMode="linear">
      <mpath href="#edge%d"/>
    </animateMotion>
  </circle>
`, dotRadius, r.opts.DotSpeed, delay+r.opts.DotSpeed/2, edgeIndex))
		}
	}

	return sb.String()
}

// nodeShape returns the appropriate shape for an entity based on step type or entity type.
func (r *InfographicRenderer) nodeShape(e pidl.Entity) NodeShape {
	if !r.opts.UseCustomShapes {
		return ShapeRectangle
	}

	// If entity has a step type, use step type shape
	if e.StepType != "" {
		return stepTypeShape(e.StepType)
	}
	// Otherwise use entity type shape
	return entityTypeShape(e.Type)
}

func stepTypeShape(st pidl.StepType) NodeShape {
	switch st {
	case pidl.StepTypeLLM:
		return ShapeHexagon
	case pidl.StepTypeHuman:
		return ShapeCircle
	case pidl.StepTypeExternal:
		return ShapeCloud
	case pidl.StepTypeTool:
		return ShapeDiamond
	default: // StepTypeDeterministic and others
		return ShapeRectangle
	}
}

func entityTypeShape(et pidl.EntityType) NodeShape {
	switch et {
	case pidl.EntityTypeUser, pidl.EntityTypeBrowser:
		return ShapeCircle
	case pidl.EntityTypeAgent, pidl.EntityTypeDelegatedAgent:
		return ShapeHexagon
	case pidl.EntityTypeResourceServer:
		return ShapeCylinder
	case pidl.EntityTypeAuthorizationServer, pidl.EntityTypeIdentityProvider:
		return ShapeDiamond
	default: // client, server, tool, etc.
		return ShapeRectangle
	}
}

// nodeIcon returns the appropriate icon for an entity, preferring step type over entity type.
func nodeIcon(e pidl.Entity) string {
	// If entity has a step type, use step type icon
	if e.StepType != "" {
		return stepTypeIcon(e.StepType)
	}
	// Otherwise use entity type icon
	return entityTypeIcon(e.Type)
}

func stepTypeIcon(st pidl.StepType) string {
	switch st {
	case pidl.StepTypeDeterministic:
		return "⚙️"
	case pidl.StepTypeLLM:
		return "🧠"
	case pidl.StepTypeHuman:
		return "👤"
	case pidl.StepTypeExternal:
		return "☁️"
	case pidl.StepTypeTool:
		return "🔧"
	default:
		return "📦"
	}
}

func entityTypeIcon(et pidl.EntityType) string {
	switch et {
	case pidl.EntityTypeClient:
		return "💻"
	case pidl.EntityTypeUser:
		return "👤"
	case pidl.EntityTypeBrowser:
		return "🌐"
	case pidl.EntityTypeAgent:
		return "🤖"
	case pidl.EntityTypeDelegatedAgent:
		return "🤝"
	case pidl.EntityTypeAuthorizationServer:
		return "🔐"
	case pidl.EntityTypeResourceServer:
		return "🗄️"
	case pidl.EntityTypeServer:
		return "🖥️"
	case pidl.EntityTypeIdentityProvider:
		return "🪪"
	case pidl.EntityTypeServiceProvider:
		return "🏢"
	case pidl.EntityTypeToolServer:
		return "🔧"
	case pidl.EntityTypeTool:
		return "🔧"
	default:
		return "📦"
	}
}
