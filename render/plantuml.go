package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// PlantUMLRenderer renders PIDL protocols as PlantUML sequence diagrams.
type PlantUMLRenderer struct {
	// Title includes the protocol name as diagram title.
	Title bool

	// ShowDescriptions includes flow descriptions as notes.
	ShowDescriptions bool

	// ShowNotes renders flow notes as PlantUML notes.
	ShowNotes bool

	// ShowAnnotations renders flow annotations as notes with type prefixes.
	ShowAnnotations bool

	// ShowConditions wraps conditional flows in opt blocks.
	ShowConditions bool

	// ShowAlternatives renders alternative paths as alt/else blocks.
	ShowAlternatives bool
}

// NewPlantUML creates a new PlantUML renderer with default options.
func NewPlantUML() *PlantUMLRenderer {
	return &PlantUMLRenderer{
		Title:            true,
		ShowDescriptions: false,
		ShowNotes:        true,
		ShowAnnotations:  true,
		ShowConditions:   true,
		ShowAlternatives: true,
	}
}

// Format returns the output format.
func (r *PlantUMLRenderer) Format() Format {
	return FormatPlantUML
}

// Render writes the PlantUML diagram to the writer.
func (r *PlantUMLRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the PlantUML diagram as a string.
func (r *PlantUMLRenderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *PlantUMLRenderer) render(p *pidl.Protocol) string {
	var sb strings.Builder

	sb.WriteString("@startuml\n")

	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "title %s\n", p.ProtocolMeta.Name)
	}

	sb.WriteString("\n")

	// Declare participants
	for _, e := range p.Entities {
		participant := r.entityToParticipant(e)
		if e.ID != e.Name {
			fmt.Fprintf(&sb, "participant \"%s\" as %s\n", e.Name, e.ID)
		} else {
			fmt.Fprintf(&sb, "participant %s\n", participant)
		}
	}

	sb.WriteString("\n")

	// Track current phase for separators
	currentPhase := ""
	phaseDepth := 0

	for _, f := range p.Flows {
		// Handle phase transitions
		if f.Phase != currentPhase {
			// Close previous phase boxes
			for i := 0; i < phaseDepth; i++ {
				sb.WriteString("end box\n")
			}
			phaseDepth = 0

			// Add phase separator/boxes
			if f.Phase != "" {
				phase := p.PhaseByID(f.Phase)
				if phase != nil {
					phaseDepth = r.openPhaseBoxes(&sb, p, phase)
				}
			}
			currentPhase = f.Phase
		}

		// Render conditional wrapper if present
		if r.ShowConditions && f.HasCondition() {
			fmt.Fprintf(&sb, "opt %s\n", f.Condition)
		}

		// Render the flow
		r.renderFlow(&sb, p, f)

		// Render alternatives if present
		if r.ShowAlternatives && f.HasAlternatives() {
			r.renderAlternatives(&sb, p, f)
		}

		// Close conditional wrapper
		if r.ShowConditions && f.HasCondition() {
			sb.WriteString("end\n")
		}
	}

	// Close final phase boxes
	for i := 0; i < phaseDepth; i++ {
		sb.WriteString("end box\n")
	}

	sb.WriteString("\n@enduml\n")

	return sb.String()
}

// openPhaseBoxes opens box containers for a phase and its parent hierarchy, returns depth.
func (r *PlantUMLRenderer) openPhaseBoxes(sb *strings.Builder, p *pidl.Protocol, phase *pidl.Phase) int {
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
		color := r.phaseColor(i)
		fmt.Fprintf(sb, "\nbox \"%s\" %s\n", ph.Name, color)
	}

	return len(hierarchy)
}

// phaseColor returns a PlantUML color for a given nesting depth.
func (r *PlantUMLRenderer) phaseColor(depth int) string {
	colors := []string{
		"#LightGray",
		"#LightBlue",
		"#LightYellow",
		"#LightGreen",
	}
	return colors[depth%len(colors)]
}

// renderFlow renders a single flow with its notes and annotations.
func (r *PlantUMLRenderer) renderFlow(sb *strings.Builder, _ *pidl.Protocol, f pidl.Flow) {
	arrow := r.modeToArrow(f.EffectiveMode())
	label := f.DisplayLabel()

	// Add mode annotation for special modes
	if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
		label = fmt.Sprintf("%s (%s)", label, ann)
	}

	fmt.Fprintf(sb, "%s %s %s: %s\n", f.From, arrow, f.To, label)

	// Add description as note if enabled
	if r.ShowDescriptions && f.Description != "" {
		fmt.Fprintf(sb, "note right: %s\n", f.Description)
	}

	// Render note if present
	if r.ShowNotes && f.HasNote() {
		fmt.Fprintf(sb, "note right: %s\n", f.Note)
	}

	// Render annotations as notes
	if r.ShowAnnotations && f.HasAnnotations() {
		for _, ann := range f.Annotations {
			prefix := r.annotationPrefix(ann.Type)
			fmt.Fprintf(sb, "note right: %s%s\n", prefix, ann.Text)
		}
	}
}

// renderAlternatives renders alternative paths using alt/else blocks.
func (r *PlantUMLRenderer) renderAlternatives(sb *strings.Builder, _ *pidl.Protocol, f pidl.Flow) {
	for i, alt := range f.Alternatives {
		if i == 0 {
			fmt.Fprintf(sb, "alt %s\n", alt.Condition)
		} else {
			fmt.Fprintf(sb, "else %s\n", alt.Condition)
		}

		for _, altFlow := range alt.Flows {
			r.renderFlow(sb, nil, altFlow)
		}
	}

	if len(f.Alternatives) > 0 {
		sb.WriteString("end\n")
	}
}

// annotationPrefix returns a visual prefix for annotation types.
func (r *PlantUMLRenderer) annotationPrefix(t pidl.AnnotationType) string {
	switch t {
	case pidl.AnnotationTypeSecurity:
		return "<&warning> SECURITY: "
	case pidl.AnnotationTypePerformance:
		return "<&timer> PERF: "
	case pidl.AnnotationTypeDeprecated:
		return "<&ban> DEPRECATED: "
	case pidl.AnnotationTypeWarning:
		return "<&warning> WARNING: "
	case pidl.AnnotationTypeError:
		return "<&x> ERROR: "
	default:
		return ""
	}
}

func (r *PlantUMLRenderer) entityToParticipant(e pidl.Entity) string {
	// Use ID for simple cases
	return e.ID
}

func (r *PlantUMLRenderer) modeToArrow(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult, pidl.FlowModeEvent:
		return "-->"
	default:
		return "->"
	}
}

func (r *PlantUMLRenderer) modeAnnotation(mode pidl.FlowMode) string {
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
