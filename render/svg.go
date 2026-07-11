package render

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/render/svg"
)

// SVGRenderer renders PIDL protocols as SVG sequence diagrams.
type SVGRenderer struct {
	// SequenceRenderOptions contains common sequence diagram options.
	SequenceRenderOptions

	// Animated enables CSS animations (moving dots along message lines).
	// This is used by FormatSVGAnimated.
	Animated bool

	// Theme selects the color scheme ("light", "dark", "auto").
	Theme string

	// ShowStepNumbers shows step numbers on messages.
	ShowStepNumbers bool

	// ShowPhases renders phase boxes around grouped flows.
	ShowPhases bool

	// ShowStepTypes includes visual styling for process step types.
	ShowStepTypes bool

	// ParticipantSpacing sets horizontal spacing between participants.
	ParticipantSpacing int

	// MessageSpacing sets vertical spacing between messages.
	MessageSpacing int

	// Template is the name of a built-in template to use.
	Template string

	// TemplateDir is a path to a custom template directory.
	TemplateDir string

	// AnimationConfig contains global animation settings.
	AnimationConfig svg.AnimationConfig

	// layout configuration
	layoutConfig svg.LayoutConfig

	// template is the loaded template (cached)
	template *svg.Template
}

// NewSVG creates a new SVG renderer with default options.
func NewSVG() *SVGRenderer {
	return &SVGRenderer{
		SequenceRenderOptions: DefaultSequenceRenderOptions(),
		Animated:              false,
		Theme:                 "light",
		ShowStepNumbers:       true,
		ShowPhases:            true,
		ShowStepTypes:         true,
		ParticipantSpacing:    0, // 0 means use default
		MessageSpacing:        0, // 0 means use default
		AnimationConfig:       svg.DefaultAnimationConfig(),
		layoutConfig:          svg.DefaultLayoutConfig(),
	}
}

// NewSVGAnimated creates a new SVG renderer with animations enabled.
func NewSVGAnimated() *SVGRenderer {
	r := NewSVG()
	r.Animated = true
	return r
}

// NewSVGWithTemplate creates an SVG renderer using the specified template.
func NewSVGWithTemplate(templateName string) (*SVGRenderer, error) {
	r := NewSVG()
	r.Template = templateName
	if err := r.loadTemplate(); err != nil {
		return nil, err
	}
	return r, nil
}

// NewSVGWithTemplateDir creates an SVG renderer using a template from a directory.
func NewSVGWithTemplateDir(dir string) (*SVGRenderer, error) {
	r := NewSVG()
	r.TemplateDir = dir
	if err := r.loadTemplate(); err != nil {
		return nil, err
	}
	return r, nil
}

// loadTemplate loads and caches the template.
func (r *SVGRenderer) loadTemplate() error {
	if r.TemplateDir != "" {
		tmpl, err := svg.LoadTemplateFromDir(r.TemplateDir)
		if err != nil {
			return err
		}
		r.template = tmpl
	} else if r.Template != "" {
		tmpl, err := svg.LoadTemplate(r.Template)
		if err != nil {
			return err
		}
		r.template = tmpl
	}
	return nil
}

// getTemplate returns the loaded template, loading it if needed.
func (r *SVGRenderer) getTemplate() *svg.Template {
	if r.template != nil {
		return r.template
	}
	// Try to load template if name is set
	if r.Template != "" || r.TemplateDir != "" {
		_ = r.loadTemplate() // ignore errors, fall back to nil
	}
	return r.template
}

// Format returns the output format.
func (r *SVGRenderer) Format() Format {
	if r.Animated {
		return FormatSVGAnimated
	}
	return FormatSVG
}

// Render writes the SVG diagram to the writer.
func (r *SVGRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the SVG diagram as a string.
func (r *SVGRenderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *SVGRenderer) render(p *pidl.Protocol) string {
	// Get template if configured
	tmpl := r.getTemplate()

	// Apply custom spacing if set
	config := r.layoutConfig

	// Apply template layout settings first
	if tmpl != nil {
		tmpl.ApplyToLayoutConfig(&config)
	}

	// Then apply explicit overrides
	if r.ParticipantSpacing > 0 {
		config.ParticipantSpacing = r.ParticipantSpacing
	}
	if r.MessageSpacing > 0 {
		config.MessageSpacing = r.MessageSpacing
	}

	// Build entity ID to index map
	entityIndex := make(map[string]int)
	for i, e := range p.Entities {
		entityIndex[e.ID] = i
	}

	// Calculate layout
	layout := svg.CalculateLayout(len(p.Entities), len(p.Flows), config)

	// Populate participant data
	for i, e := range p.Entities {
		layout.Participants[i].ID = e.ID
		layout.Participants[i].Name = e.Name
		// Add step type info for process specs
		if r.ShowStepTypes && p.IsProcessSpec() && e.IsProcessStep() {
			layout.Participants[i].StepType = string(e.StepType)
			layout.Participants[i].IsProcessStep = true
		}
	}

	// Populate message data and calculate endpoints
	for i, f := range p.Flows {
		fromIdx, fromOK := entityIndex[f.From]
		toIdx, toOK := entityIndex[f.To]
		if !fromOK || !toOK {
			continue
		}

		layout.SetMessageEndpoints(i, fromIdx, toIdx)
		layout.Messages[i].Label = f.DisplayLabel()
		layout.Messages[i].IsDashed = r.isResponseMode(f.EffectiveMode())

		// Populate alternative paths info
		if r.ShowAlternatives && f.HasAlternatives() {
			layout.Messages[i].HasAlternatives = true
			layout.Messages[i].AlternativeCount = len(f.Alternatives)
			conditions := make([]string, len(f.Alternatives))
			for j, alt := range f.Alternatives {
				conditions[j] = alt.Condition
			}
			layout.Messages[i].AlternativeConditions = conditions
		}

		// Populate note info
		if r.ShowNotes && f.HasNote() {
			layout.Messages[i].HasNote = true
			layout.Messages[i].Note = f.Note
		}
	}

	// Build SVG
	var sb strings.Builder

	// SVG header
	themeClass := svg.ThemeClass(svg.Theme(r.Theme))
	sb.WriteString(fmt.Sprintf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg"`, layout.Width, layout.Height))
	if themeClass != "" {
		sb.WriteString(fmt.Sprintf(` class="%s"`, themeClass))
	}
	sb.WriteString(">\n")

	// Resolve animation styles for each flow
	var flowStyles []svg.FlowAnimationStyle
	if r.Animated {
		flowStyles = make([]svg.FlowAnimationStyle, len(p.Flows))
		for i, f := range p.Flows {
			flowStyles[i] = svg.ResolveFlowAnimation(f.Animation, i, r.AnimationConfig)
		}
	}

	// Embedded CSS
	sb.WriteString("  <style>\n")
	if tmpl != nil {
		sb.WriteString(tmpl.GenerateCSS(svg.Theme(r.Theme)))
	} else {
		sb.WriteString(svg.GenerateCSS(svg.Theme(r.Theme)))
	}
	if r.Animated {
		sb.WriteString(r.generateAnimationCSS(flowStyles))
	}
	sb.WriteString("  </style>\n\n")

	// Defs for arrow markers
	sb.WriteString(r.generateDefs())

	// Title (if enabled)
	if r.Title && p.ProtocolMeta.Name != "" {
		sb.WriteString(fmt.Sprintf(`  <text x="%d" y="12" class="message-text" style="font-size:14px;font-weight:600;text-anchor:middle;">%s</text>`+"\n\n",
			layout.Width/2, html.EscapeString(p.ProtocolMeta.Name)))
	}

	// Phase boxes (render first, behind everything)
	if r.ShowPhases && len(p.Phases) > 0 {
		phaseRegions := r.calculatePhaseRegions(p, layout, entityIndex)
		if len(phaseRegions) > 0 {
			sb.WriteString("  <!-- Phase boxes -->\n")
			sb.WriteString("  <g class=\"phase-boxes\">\n")
			for _, region := range phaseRegions {
				width := region.MaxX - region.MinX
				height := region.MaxY - region.MinY
				sb.WriteString(fmt.Sprintf(`    <rect class="phase-box" x="%d" y="%d" width="%d" height="%d" rx="4" style="fill: %s;"/>`+"\n",
					region.MinX, region.MinY, width, height, region.Color))
				sb.WriteString(fmt.Sprintf(`    <text class="phase-label" x="%d" y="%d">%s</text>`+"\n",
					region.MinX+8, region.MinY+14, html.EscapeString(region.Name)))
			}
			sb.WriteString("  </g>\n\n")
		}
	}

	// Lifelines (render behind messages but above phase boxes)
	sb.WriteString("  <!-- Lifelines -->\n")
	sb.WriteString("  <g class=\"lifelines\">\n")
	for _, part := range layout.Participants {
		sb.WriteString(fmt.Sprintf(`    <line class="lifeline" x1="%d" y1="%d" x2="%d" y2="%d"/>`+"\n",
			part.CenterX, part.LifelineStartY, part.CenterX, part.LifelineEndY))
	}
	sb.WriteString("  </g>\n\n")

	// Participant boxes
	cornerRadius := 4
	if tmpl != nil {
		cornerRadius = tmpl.GetCornerRadius()
	}
	sb.WriteString("  <!-- Participants -->\n")
	sb.WriteString("  <g class=\"participants\">\n")
	for _, part := range layout.Participants {
		// Apply step type styling for process specs
		styleAttr := ""
		if part.IsProcessStep {
			fill, stroke := r.stepTypeColors(part.StepType)
			styleAttr = fmt.Sprintf(` style="fill:%s;stroke:%s;"`, fill, stroke)
		}
		sb.WriteString(fmt.Sprintf(`    <rect class="participant-box" x="%d" y="%d" width="%d" height="%d" rx="%d"%s/>`+"\n",
			part.BoxX, part.BoxY, part.BoxWidth, part.BoxHeight, cornerRadius, styleAttr))

		// Add step type badge for process steps
		displayName := part.Name
		if part.IsProcessStep {
			badge := r.stepTypeBadge(part.StepType)
			if badge != "" {
				displayName = badge + " " + displayName
			}
		}
		sb.WriteString(fmt.Sprintf(`    <text class="participant-text" x="%d" y="%d">%s</text>`+"\n",
			part.CenterX, part.BoxY+part.BoxHeight/2, html.EscapeString(displayName)))
	}
	sb.WriteString("  </g>\n\n")

	// Messages
	sb.WriteString("  <!-- Messages -->\n")
	sb.WriteString("  <g class=\"messages\">\n")
	for i, msg := range layout.Messages {
		var style *svg.FlowAnimationStyle
		if r.Animated && i < len(flowStyles) {
			style = &flowStyles[i]
		}
		var security *pidl.FlowSecurity
		if r.ShowSecurity && i < len(p.Flows) {
			security = p.Flows[i].Security
		}
		r.renderMessage(&sb, msg, i, style, security)
	}
	sb.WriteString("  </g>\n")

	sb.WriteString("</svg>\n")

	return sb.String()
}

func (r *SVGRenderer) renderMessage(sb *strings.Builder, msg svg.MessageLayout, index int, style *svg.FlowAnimationStyle, security *pidl.FlowSecurity) {
	lineClass := "message-line"
	arrowClass := "message-arrow"
	if msg.IsDashed {
		lineClass = "message-line-response"
		arrowClass = "message-arrow-response"
	}

	// Message line
	markerEnd := fmt.Sprintf("url(#arrow-%s)", arrowClass)
	if msg.IsReverse {
		markerEnd = fmt.Sprintf("url(#arrow-%s-reverse)", arrowClass)
	}

	pathID := fmt.Sprintf("msg-path-%d", index)
	sb.WriteString(fmt.Sprintf(`    <path id="%s" class="%s" d="%s" marker-end="%s"/>`+"\n",
		pathID, lineClass, msg.PathD, markerEnd))

	// Step number circle
	if r.ShowStepNumbers {
		// Position step number at the start of the arrow
		stepX := msg.FromX
		if msg.IsReverse {
			stepX = msg.FromX // Still at fromX for reverse
		}
		// Offset slightly toward the direction of the arrow
		if !msg.IsReverse {
			stepX += 10
		} else {
			stepX -= 10
		}

		sb.WriteString(fmt.Sprintf(`    <circle class="step-circle" cx="%d" cy="%d" r="8"/>`+"\n", stepX, msg.Y))
		sb.WriteString(fmt.Sprintf(`    <text class="step-number" x="%d" y="%d">%d</text>`+"\n", stepX, msg.Y, msg.Step))
	}

	// Message label
	labelX := (msg.FromX + msg.ToX) / 2
	labelY := msg.Y - 8
	sb.WriteString(fmt.Sprintf(`    <text class="message-text" x="%d" y="%d" text-anchor="middle">%s</text>`+"\n",
		labelX, labelY, html.EscapeString(msg.Label)))

	// Security badge (before alternatives badge)
	if security != nil && (len(security.Requires) > 0 || security.Confidential) {
		secBadgeX := labelX + 60
		secBadgeY := msg.Y - 10

		// Adjust position if there will be an alt badge
		if msg.HasAlternatives {
			secBadgeX += 35
		}

		sb.WriteString(fmt.Sprintf(`    <g class="security-badge" transform="translate(%d, %d)">`+"\n", secBadgeX, secBadgeY))
		sb.WriteString(`      <rect class="security-badge-bg" x="0" y="-9" width="24" height="14" rx="3"/>` + "\n")
		sb.WriteString(`      <text class="security-badge-text" x="12" y="1">🔒</text>` + "\n")
		sb.WriteString("    </g>\n")
	}

	// Alternative paths indicator
	if msg.HasAlternatives {
		// Position badge to the right of the message midpoint
		altBadgeX := labelX + 60
		altBadgeY := msg.Y - 10

		// Draw alt badge
		sb.WriteString(fmt.Sprintf(`    <g class="alt-badge" transform="translate(%d, %d)">`+"\n", altBadgeX, altBadgeY))
		sb.WriteString(fmt.Sprintf(`      <rect class="alt-badge-bg" x="0" y="-9" width="30" height="14" rx="3"/>` + "\n"))
		sb.WriteString(fmt.Sprintf(`      <text class="alt-badge-text" x="15" y="1">ALT %d</text>`+"\n", msg.AlternativeCount))
		sb.WriteString("    </g>\n")
	}

	// Note indicator
	if msg.HasNote && msg.Note != "" {
		// Position note below the message line
		noteX := labelX
		noteY := msg.Y + 14

		sb.WriteString(fmt.Sprintf(`    <text class="message-note" x="%d" y="%d" text-anchor="middle">📝 %s</text>`+"\n",
			noteX, noteY, html.EscapeString(msg.Note)))
	}

	// Animated dot (if style is provided and enabled)
	if style != nil && style.Enabled {
		sb.WriteString(fmt.Sprintf(`    <circle class="flow-dot flow-dot-%d" r="%d" style="offset-path: path('%s');"/>`+"\n",
			index, style.DotSize, msg.PathD))
	}
}

func (r *SVGRenderer) generateDefs() string {
	return `  <defs>
    <!-- Arrow markers -->
    <marker id="arrow-message-arrow" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M0,0 L0,6 L9,3 z" class="message-arrow"/>
    </marker>
    <marker id="arrow-message-arrow-reverse" markerWidth="10" markerHeight="10" refX="0" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M9,0 L9,6 L0,3 z" class="message-arrow"/>
    </marker>
    <marker id="arrow-message-arrow-response" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M0,0 L0,6 L9,3 z" class="message-arrow-response"/>
    </marker>
    <marker id="arrow-message-arrow-response-reverse" markerWidth="10" markerHeight="10" refX="0" refY="3" orient="auto" markerUnits="strokeWidth">
      <path d="M9,0 L9,6 L0,3 z" class="message-arrow-response"/>
    </marker>
  </defs>

`
}

func (r *SVGRenderer) generateAnimationCSS(styles []svg.FlowAnimationStyle) string {
	return svg.GenerateAnimationCSS(styles, nil)
}

func (r *SVGRenderer) isResponseMode(mode pidl.FlowMode) bool {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult, pidl.FlowModeEvent:
		return true
	default:
		return false
	}
}

// phaseRegion represents the bounding box for a phase.
type phaseRegion struct {
	ID    string
	Name  string
	MinX  int
	MaxX  int
	MinY  int
	MaxY  int
	Depth int
	Color string
}

// phaseColors are the default colors for phase boxes at different depths.
var phaseColors = []string{
	"rgba(49, 130, 206, 0.08)",  // blue tint (depth 0)
	"rgba(72, 187, 120, 0.08)",  // green tint (depth 1)
	"rgba(237, 137, 54, 0.08)",  // orange tint (depth 2)
	"rgba(159, 122, 234, 0.08)", // purple tint (depth 3)
}

// calculatePhaseRegions calculates bounding boxes for each phase based on flow positions.
func (r *SVGRenderer) calculatePhaseRegions(p *pidl.Protocol, layout svg.Layout, entityIndex map[string]int) []phaseRegion {
	if len(p.Phases) == 0 {
		return nil
	}

	// Create a map from phase ID to region
	regions := make(map[string]*phaseRegion)

	// Initialize regions for all phases
	for _, ph := range p.Phases {
		depth := p.PhaseDepth(ph.ID)
		colorIdx := depth % len(phaseColors)
		regions[ph.ID] = &phaseRegion{
			ID:    ph.ID,
			Name:  ph.Name,
			MinX:  layout.Width,
			MaxX:  0,
			MinY:  layout.Height,
			MaxY:  0,
			Depth: depth,
			Color: phaseColors[colorIdx],
		}
	}

	// Calculate bounds based on flows
	for i, f := range p.Flows {
		if f.Phase == "" {
			continue
		}
		region, ok := regions[f.Phase]
		if !ok {
			continue
		}

		// Get message Y position
		msgY := layout.Messages[i].Y

		// Get participant X positions
		fromIdx, fromOK := entityIndex[f.From]
		toIdx, toOK := entityIndex[f.To]
		if !fromOK || !toOK {
			continue
		}

		fromX := layout.Participants[fromIdx].CenterX
		toX := layout.Participants[toIdx].CenterX

		// Update bounds
		if fromX < region.MinX {
			region.MinX = fromX
		}
		if toX < region.MinX {
			region.MinX = toX
		}
		if fromX > region.MaxX {
			region.MaxX = fromX
		}
		if toX > region.MaxX {
			region.MaxX = toX
		}
		if msgY-20 < region.MinY {
			region.MinY = msgY - 20
		}
		if msgY+15 > region.MaxY {
			region.MaxY = msgY + 15
		}
	}

	// Convert to slice and filter out empty regions
	var result []phaseRegion
	for _, ph := range p.Phases {
		region := regions[ph.ID]
		// Only include regions that have at least one flow
		if region.MaxX > region.MinX && region.MaxY > region.MinY {
			// Add padding
			region.MinX -= 30
			region.MaxX += 30
			region.MinY -= 10
			region.MaxY += 10
			result = append(result, *region)
		}
	}

	return result
}

// stepTypeColors returns fill and stroke colors for a process step type.
func (r *SVGRenderer) stepTypeColors(st string) (fill, stroke string) {
	switch st {
	case string(pidl.StepTypeDeterministic):
		return "#E3F2FD", "#1976D2" // blue
	case string(pidl.StepTypeLLM):
		return "#F3E5F5", "#7B1FA2" // purple
	case string(pidl.StepTypeHuman):
		return "#E8F5E9", "#388E3C" // green
	case string(pidl.StepTypeExternal):
		return "#FFF3E0", "#F57C00" // orange
	case string(pidl.StepTypeTool):
		return "#ECEFF1", "#607D8B" // gray
	default:
		return "#FFFFFF", "#333333"
	}
}

// stepTypeBadge returns an emoji badge for a process step type.
func (r *SVGRenderer) stepTypeBadge(st string) string {
	switch st {
	case string(pidl.StepTypeDeterministic):
		return "⚙️"
	case string(pidl.StepTypeLLM):
		return "🧠"
	case string(pidl.StepTypeHuman):
		return "👤"
	case string(pidl.StepTypeExternal):
		return "☁️"
	case string(pidl.StepTypeTool):
		return "🔧"
	default:
		return ""
	}
}
