package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// MermaidRenderer renders PIDL protocols as Mermaid sequence diagrams.
type MermaidRenderer struct {
	// Title includes the protocol name as diagram title.
	Title bool

	// Autonumber adds sequence numbers to messages.
	Autonumber bool
}

// NewMermaid creates a new Mermaid renderer with default options.
func NewMermaid() *MermaidRenderer {
	return &MermaidRenderer{
		Title:      true,
		Autonumber: true,
	}
}

// Format returns the output format.
func (r *MermaidRenderer) Format() Format {
	return FormatMermaid
}

// Render writes the Mermaid diagram to the writer.
func (r *MermaidRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the Mermaid diagram as a string.
func (r *MermaidRenderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *MermaidRenderer) render(p *pidl.Protocol) string {
	var sb strings.Builder

	sb.WriteString("sequenceDiagram\n")

	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "    title %s\n", p.ProtocolMeta.Name)
	}

	if r.Autonumber {
		sb.WriteString("    autonumber\n")
	}

	sb.WriteString("\n")

	// Declare participants
	for _, e := range p.Entities {
		fmt.Fprintf(&sb, "    participant %s as %s\n", e.ID, e.Name)
	}

	sb.WriteString("\n")

	// Track current phase for boxes
	currentPhase := ""

	for _, f := range p.Flows {
		// Add phase separator if phase changed
		if f.Phase != "" && f.Phase != currentPhase {
			// Close previous box if any
			if currentPhase != "" {
				sb.WriteString("    end\n\n")
			}
			phase := p.PhaseByID(f.Phase)
			if phase != nil {
				sb.WriteString("    rect rgb(240, 240, 240)\n")
				fmt.Fprintf(&sb, "    note right of %s: %s\n", p.Entities[0].ID, phase.Name)
			}
			currentPhase = f.Phase
		}

		// Render the flow
		arrow := r.modeToArrow(f.EffectiveMode())
		label := r.escapeLabel(f.DisplayLabel())

		// Add mode annotation for special modes
		if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
			label = fmt.Sprintf("%s (%s)", label, ann)
		}

		fmt.Fprintf(&sb, "    %s%s%s: %s\n", f.From, arrow, f.To, label)
	}

	// Close final box if any
	if currentPhase != "" {
		sb.WriteString("    end\n")
	}

	return sb.String()
}

func (r *MermaidRenderer) modeToArrow(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult, pidl.FlowModeEvent:
		return "-->>"
	default:
		return "->>"
	}
}

func (r *MermaidRenderer) modeAnnotation(mode pidl.FlowMode) string {
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

func (r *MermaidRenderer) escapeLabel(label string) string {
	// Mermaid requires escaping certain characters
	label = strings.ReplaceAll(label, ":", "&#58;")
	label = strings.ReplaceAll(label, "#", "&#35;")
	return label
}
