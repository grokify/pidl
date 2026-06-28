package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// MermaidStateRenderer renders PIDL protocols as Mermaid state diagrams.
type MermaidStateRenderer struct {
	// Title includes the protocol name as diagram title.
	Title bool

	// EntityFilter limits output to a specific entity (empty for all).
	EntityFilter string

	// ShowDescriptions includes state descriptions as notes.
	ShowDescriptions bool
}

// NewMermaidState creates a new Mermaid state diagram renderer with default options.
func NewMermaidState() *MermaidStateRenderer {
	return &MermaidStateRenderer{
		Title:            true,
		ShowDescriptions: true,
	}
}

// Format returns the output format.
func (r *MermaidStateRenderer) Format() Format {
	return FormatMermaidState
}

// Render writes the Mermaid state diagram to the writer.
func (r *MermaidStateRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the Mermaid state diagram as a string.
func (r *MermaidStateRenderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *MermaidStateRenderer) render(p *pidl.Protocol) string {
	var sb strings.Builder

	sb.WriteString("stateDiagram-v2\n")

	// Collect entities to render
	var entities []pidl.Entity
	if r.EntityFilter != "" {
		e := p.EntityByID(r.EntityFilter)
		if e != nil && e.HasStates() {
			entities = []pidl.Entity{*e}
		}
	} else {
		entities = p.EntitiesWithStates()
	}

	if len(entities) == 0 {
		sb.WriteString("    note: No entities with states defined\n")
		return sb.String()
	}

	// Get all transitions
	allTransitions := p.StateTransitions()

	for idx, e := range entities {
		// Add separator between entities (except for filtered single entity)
		if idx > 0 {
			sb.WriteString("\n")
		}

		// If multiple entities, wrap each in a state as a container
		useContainer := len(entities) > 1
		containerID := ""

		if useContainer {
			containerID = e.ID
			displayName := e.Name
			if displayName == "" {
				displayName = e.ID
			}
			fmt.Fprintf(&sb, "    state \"%s\" as %s {\n", r.escapeLabel(displayName), containerID)
		}

		indent := "    "
		if useContainer {
			indent = "        "
		}

		// Declare states
		for _, s := range e.States {
			stateID := r.stateID(containerID, s.ID)
			displayName := s.Name
			if displayName == "" {
				displayName = s.ID
			}
			fmt.Fprintf(&sb, "%sstate \"%s\" as %s\n", indent, r.escapeLabel(displayName), stateID)

			// Add description as note if enabled
			if r.ShowDescriptions && s.Description != "" {
				fmt.Fprintf(&sb, "%snote right of %s: %s\n", indent, stateID, r.escapeLabel(s.Description))
			}
		}

		sb.WriteString("\n")

		// Initial state transition
		initial := e.InitialState()
		if initial != nil {
			fmt.Fprintf(&sb, "%s[*] --> %s\n", indent, r.stateID(containerID, initial.ID))
		}

		// State transitions from flows
		entityTransitions := r.transitionsForEntity(allTransitions, e.ID)
		for _, t := range entityTransitions {
			fromState := t.FromState
			toState := t.ToState
			label := t.FlowLabel

			fromID := r.stateID(containerID, fromState)
			toID := r.stateID(containerID, toState)

			if fromState == "" {
				// No specific from state - show as transition from any state
				// We'll skip these as they're ambiguous in state diagrams
				// Or we could show them from [*], but that's not semantically correct
				fmt.Fprintf(&sb, "%s%s --> %s: %s\n", indent, toID, toID, r.escapeLabel(label))
			} else {
				fmt.Fprintf(&sb, "%s%s --> %s: %s\n", indent, fromID, toID, r.escapeLabel(label))
			}
		}

		// Final state transitions
		for _, s := range e.States {
			if s.Final {
				fmt.Fprintf(&sb, "%s%s --> [*]\n", indent, r.stateID(containerID, s.ID))
			}
		}

		if useContainer {
			sb.WriteString("    }\n")
		}
	}

	return sb.String()
}

// transitionsForEntity filters transitions for a specific entity.
func (r *MermaidStateRenderer) transitionsForEntity(transitions []pidl.StateTransition, entityID string) []pidl.StateTransition {
	var result []pidl.StateTransition
	for _, t := range transitions {
		if t.EntityID == entityID {
			result = append(result, t)
		}
	}
	return result
}

// stateID generates a unique state ID, optionally prefixed with container ID.
func (r *MermaidStateRenderer) stateID(containerID, stateID string) string {
	if containerID == "" {
		return stateID
	}
	return containerID + "_" + stateID
}

// escapeLabel escapes special characters for Mermaid labels.
func (r *MermaidStateRenderer) escapeLabel(label string) string {
	// Mermaid requires escaping certain characters.
	// Order matters: escape # first to avoid double-escaping &#58; etc.
	label = strings.ReplaceAll(label, "#", "&#35;")
	label = strings.ReplaceAll(label, ":", "&#58;")
	label = strings.ReplaceAll(label, "\"", "&quot;")
	return label
}
