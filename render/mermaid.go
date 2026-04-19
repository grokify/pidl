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

	// ShowNotes renders flow notes as Mermaid notes.
	ShowNotes bool

	// ShowAnnotations renders flow annotations as notes with type prefixes.
	ShowAnnotations bool

	// ShowConditions wraps conditional flows in opt blocks.
	ShowConditions bool

	// ShowAlternatives renders alternative paths as alt/else blocks.
	ShowAlternatives bool
}

// NewMermaid creates a new Mermaid renderer with default options.
func NewMermaid() *MermaidRenderer {
	return &MermaidRenderer{
		Title:            true,
		Autonumber:       true,
		ShowNotes:        true,
		ShowAnnotations:  true,
		ShowConditions:   true,
		ShowAlternatives: true,
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
	phaseDepth := 0

	for _, f := range p.Flows {
		// Handle phase transitions
		if f.Phase != currentPhase {
			// Close previous phase boxes
			for i := 0; i < phaseDepth; i++ {
				sb.WriteString("    end\n")
			}
			if currentPhase != "" {
				sb.WriteString("\n")
			}
			phaseDepth = 0

			// Open new phase boxes (including parent hierarchy)
			if f.Phase != "" {
				phase := p.PhaseByID(f.Phase)
				if phase != nil {
					phaseDepth = r.openPhaseBoxes(&sb, p, phase)
				}
			}
			currentPhase = f.Phase
		}

		// Determine indentation based on phase depth
		indent := "    "
		for i := 0; i < phaseDepth; i++ {
			indent += "    "
		}

		// Render conditional wrapper if present
		if r.ShowConditions && f.HasCondition() {
			fmt.Fprintf(&sb, "%sopt %s\n", indent, r.escapeLabel(f.Condition))
			indent += "    "
		}

		// Render the flow
		r.renderFlow(&sb, p, f, indent)

		// Render alternatives if present
		if r.ShowAlternatives && f.HasAlternatives() {
			r.renderAlternatives(&sb, p, f, indent)
		}

		// Close conditional wrapper
		if r.ShowConditions && f.HasCondition() {
			indent = indent[:len(indent)-4]
			fmt.Fprintf(&sb, "%send\n", indent)
		}
	}

	// Close final phase boxes
	for i := 0; i < phaseDepth; i++ {
		sb.WriteString("    end\n")
	}

	return sb.String()
}

// openPhaseBoxes opens rect boxes for a phase and its parent hierarchy, returns depth.
func (r *MermaidRenderer) openPhaseBoxes(sb *strings.Builder, p *pidl.Protocol, phase *pidl.Phase) int {
	// Build the hierarchy from root to current phase
	var hierarchy []*pidl.Phase
	current := phase
	for current != nil {
		hierarchy = append([]*pidl.Phase{current}, hierarchy...)
		if current.Parent == "" {
			break
		}
		current = p.PhaseByID(current.Parent)
	}

	// Open boxes from root to leaf
	for i, ph := range hierarchy {
		indent := "    "
		for j := 0; j < i; j++ {
			indent += "    "
		}
		// Use different colors for different nesting levels
		color := r.phaseColor(i)
		fmt.Fprintf(sb, "%srect %s\n", indent, color)
		fmt.Fprintf(sb, "%snote right of %s: %s\n", indent+"    ", p.Entities[0].ID, ph.Name)
	}

	return len(hierarchy)
}

// phaseColor returns a color for a given nesting depth.
func (r *MermaidRenderer) phaseColor(depth int) string {
	colors := []string{
		"rgb(240, 240, 240)",
		"rgb(230, 240, 250)",
		"rgb(250, 240, 230)",
		"rgb(240, 250, 240)",
	}
	return colors[depth%len(colors)]
}

// renderFlow renders a single flow with its notes and annotations.
func (r *MermaidRenderer) renderFlow(sb *strings.Builder, _ *pidl.Protocol, f pidl.Flow, indent string) {
	arrow := r.modeToArrow(f.EffectiveMode())
	label := r.escapeLabel(f.DisplayLabel())

	// Add mode annotation for special modes
	if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
		label = fmt.Sprintf("%s (%s)", label, ann)
	}

	fmt.Fprintf(sb, "%s%s%s%s: %s\n", indent, f.From, arrow, f.To, label)

	// Render note if present
	if r.ShowNotes && f.HasNote() {
		fmt.Fprintf(sb, "%snote right of %s: %s\n", indent, f.To, r.escapeLabel(f.Note))
	}

	// Render annotations as notes
	if r.ShowAnnotations && f.HasAnnotations() {
		for _, ann := range f.Annotations {
			prefix := r.annotationPrefix(ann.Type)
			fmt.Fprintf(sb, "%snote right of %s: %s%s\n", indent, f.To, prefix, r.escapeLabel(ann.Text))
		}
	}
}

// renderAlternatives renders alternative paths using alt/else blocks.
func (r *MermaidRenderer) renderAlternatives(sb *strings.Builder, _ *pidl.Protocol, f pidl.Flow, indent string) {
	for i, alt := range f.Alternatives {
		if i == 0 {
			fmt.Fprintf(sb, "%salt %s\n", indent, r.escapeLabel(alt.Condition))
		} else {
			fmt.Fprintf(sb, "%selse %s\n", indent, r.escapeLabel(alt.Condition))
		}

		altIndent := indent + "    "
		for _, altFlow := range alt.Flows {
			r.renderFlow(sb, nil, altFlow, altIndent)
		}
	}

	if len(f.Alternatives) > 0 {
		fmt.Fprintf(sb, "%send\n", indent)
	}
}

// annotationPrefix returns a visual prefix for annotation types.
func (r *MermaidRenderer) annotationPrefix(t pidl.AnnotationType) string {
	switch t {
	case pidl.AnnotationTypeSecurity:
		return "[SECURITY] "
	case pidl.AnnotationTypePerformance:
		return "[PERF] "
	case pidl.AnnotationTypeDeprecated:
		return "[DEPRECATED] "
	case pidl.AnnotationTypeWarning:
		return "[WARNING] "
	case pidl.AnnotationTypeError:
		return "[ERROR] "
	default:
		return ""
	}
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
