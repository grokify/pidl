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
}

// NewPlantUML creates a new PlantUML renderer with default options.
func NewPlantUML() *PlantUMLRenderer {
	return &PlantUMLRenderer{
		Title:            true,
		ShowDescriptions: false,
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

	for _, f := range p.Flows {
		// Add phase separator if phase changed
		if f.Phase != "" && f.Phase != currentPhase {
			phase := p.PhaseByID(f.Phase)
			if phase != nil {
				fmt.Fprintf(&sb, "\n== %s ==\n", phase.Name)
			}
			currentPhase = f.Phase
		}

		// Render the flow
		arrow := r.modeToArrow(f.EffectiveMode())
		label := f.DisplayLabel()

		// Add mode annotation for special modes
		if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
			label = fmt.Sprintf("%s (%s)", label, ann)
		}

		fmt.Fprintf(&sb, "%s %s %s: %s\n", f.From, arrow, f.To, label)

		// Add description as note if enabled
		if r.ShowDescriptions && f.Description != "" {
			fmt.Fprintf(&sb, "note right: %s\n", f.Description)
		}
	}

	sb.WriteString("\n@enduml\n")

	return sb.String()
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
