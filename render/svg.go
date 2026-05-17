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
	// Animated enables CSS animations (moving dots along message lines).
	// This is used by FormatSVGAnimated.
	Animated bool

	// Theme selects the color scheme ("light", "dark", "auto").
	Theme string

	// Title includes the protocol name in the diagram.
	Title bool

	// ShowNotes renders flow notes.
	ShowNotes bool

	// ShowStepNumbers shows step numbers on messages.
	ShowStepNumbers bool

	// ParticipantSpacing sets horizontal spacing between participants.
	ParticipantSpacing int

	// MessageSpacing sets vertical spacing between messages.
	MessageSpacing int

	// layout configuration
	layoutConfig svg.LayoutConfig
}

// NewSVG creates a new SVG renderer with default options.
func NewSVG() *SVGRenderer {
	return &SVGRenderer{
		Animated:           false,
		Theme:              "light",
		Title:              true,
		ShowNotes:          true,
		ShowStepNumbers:    true,
		ParticipantSpacing: 0, // 0 means use default
		MessageSpacing:     0, // 0 means use default
		layoutConfig:       svg.DefaultLayoutConfig(),
	}
}

// NewSVGAnimated creates a new SVG renderer with animations enabled.
func NewSVGAnimated() *SVGRenderer {
	r := NewSVG()
	r.Animated = true
	return r
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
	// Apply custom spacing if set
	config := r.layoutConfig
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

	// Embedded CSS
	sb.WriteString("  <style>\n")
	sb.WriteString(svg.GenerateCSS(svg.Theme(r.Theme)))
	if r.Animated {
		sb.WriteString(r.generateAnimationCSS(len(p.Flows)))
	}
	sb.WriteString("  </style>\n\n")

	// Defs for arrow markers
	sb.WriteString(r.generateDefs())

	// Title (if enabled)
	if r.Title && p.ProtocolMeta.Name != "" {
		sb.WriteString(fmt.Sprintf(`  <text x="%d" y="12" class="message-text" style="font-size:14px;font-weight:600;text-anchor:middle;">%s</text>`+"\n\n",
			layout.Width/2, html.EscapeString(p.ProtocolMeta.Name)))
	}

	// Lifelines (render first, behind everything)
	sb.WriteString("  <!-- Lifelines -->\n")
	sb.WriteString("  <g class=\"lifelines\">\n")
	for _, part := range layout.Participants {
		sb.WriteString(fmt.Sprintf(`    <line class="lifeline" x1="%d" y1="%d" x2="%d" y2="%d"/>`+"\n",
			part.CenterX, part.LifelineStartY, part.CenterX, part.LifelineEndY))
	}
	sb.WriteString("  </g>\n\n")

	// Participant boxes
	sb.WriteString("  <!-- Participants -->\n")
	sb.WriteString("  <g class=\"participants\">\n")
	for _, part := range layout.Participants {
		sb.WriteString(fmt.Sprintf(`    <rect class="participant-box" x="%d" y="%d" width="%d" height="%d" rx="4"/>`+"\n",
			part.BoxX, part.BoxY, part.BoxWidth, part.BoxHeight))
		sb.WriteString(fmt.Sprintf(`    <text class="participant-text" x="%d" y="%d">%s</text>`+"\n",
			part.CenterX, part.BoxY+part.BoxHeight/2, html.EscapeString(part.Name)))
	}
	sb.WriteString("  </g>\n\n")

	// Messages
	sb.WriteString("  <!-- Messages -->\n")
	sb.WriteString("  <g class=\"messages\">\n")
	for i, msg := range layout.Messages {
		r.renderMessage(&sb, msg, i, r.Animated)
	}
	sb.WriteString("  </g>\n")

	sb.WriteString("</svg>\n")

	return sb.String()
}

func (r *SVGRenderer) renderMessage(sb *strings.Builder, msg svg.MessageLayout, index int, animated bool) {
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

	// Animated dot (if enabled)
	if animated {
		sb.WriteString(fmt.Sprintf(`    <circle class="flow-dot flow-dot-%d" r="4" style="offset-path: path('%s');"/>`+"\n",
			index, msg.PathD))
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

func (r *SVGRenderer) generateAnimationCSS(messageCount int) string {
	var sb strings.Builder

	sb.WriteString(`
    /* Animation keyframes */
    @keyframes flow {
      0% { offset-distance: 0%; }
      100% { offset-distance: 100%; }
    }

    .flow-dot {
      fill: var(--color-dot);
      offset-rotate: 0deg;
      animation: flow 2s linear infinite;
    }
`)

	// Staggered delays for each message
	for i := 0; i < messageCount; i++ {
		delay := float64(i) * 0.3
		sb.WriteString(fmt.Sprintf("    .flow-dot-%d { animation-delay: %.1fs; }\n", i, delay))
	}

	return sb.String()
}

func (r *SVGRenderer) isResponseMode(mode pidl.FlowMode) bool {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult, pidl.FlowModeEvent:
		return true
	default:
		return false
	}
}
