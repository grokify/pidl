package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/grokify/pidl"
)

// TraceTextOptions configures text trace rendering.
type TraceTextOptions struct {
	// ShowTimestamps includes step timestamps.
	ShowTimestamps bool
	// ShowStates shows entity states after each step.
	ShowStates bool
	// Compact uses single-line output per step.
	Compact bool
	// UseColors enables ANSI color codes (for terminals).
	UseColors bool
}

// DefaultTraceTextOptions returns default text rendering options.
func DefaultTraceTextOptions() TraceTextOptions {
	return TraceTextOptions{
		ShowTimestamps: false,
		ShowStates:     true,
		Compact:        false,
		UseColors:      false,
	}
}

// TraceRenderer renders execution traces in various formats.
type TraceRenderer struct {
	// ShowStates includes entity states in output.
	ShowStates bool
	// ShowTimings includes timing information.
	ShowTimings bool
	// HighlightSkipped emphasizes skipped flows.
	HighlightSkipped bool
}

// NewTraceRenderer creates a new TraceRenderer with default options.
func NewTraceRenderer() *TraceRenderer {
	return &TraceRenderer{
		ShowStates:       true,
		ShowTimings:      false,
		HighlightSkipped: true,
	}
}

// RenderText renders the trace as formatted text.
func (r *TraceRenderer) RenderText(trace *pidl.ExecutionTrace, opts TraceTextOptions) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("Execution Trace: %s\n", trace.ProtocolName))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	if opts.ShowTimestamps {
		sb.WriteString(fmt.Sprintf("Started: %s\n", trace.StartTime.Format(time.RFC3339)))
		if !trace.EndTime.IsZero() {
			sb.WriteString(fmt.Sprintf("Ended:   %s\n", trace.EndTime.Format(time.RFC3339)))
			sb.WriteString(fmt.Sprintf("Duration: %v\n", trace.Duration()))
		}
		sb.WriteString("\n")
	}

	// Initial states
	if opts.ShowStates && len(trace.InitialStates) > 0 {
		sb.WriteString("Initial States:\n")
		for entity, state := range trace.InitialStates {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", entity, state))
		}
		sb.WriteString("\n")
	}

	// Steps
	sb.WriteString("Steps:\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, step := range trace.Steps {
		r.renderTextStep(&sb, &step, opts)
	}

	// Summary
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString("Summary:\n")
	sb.WriteString(fmt.Sprintf("  Steps executed: %d\n", trace.StepCount()))
	sb.WriteString(fmt.Sprintf("  Steps skipped:  %d\n", trace.SkippedCount()))
	sb.WriteString(fmt.Sprintf("  State changes:  %d\n", trace.StateChangeCount()))

	if trace.Completed {
		sb.WriteString("  Status: Completed\n")
	} else {
		sb.WriteString("  Status: Partial\n")
	}

	if trace.Error != "" {
		sb.WriteString(fmt.Sprintf("  Error: %s\n", trace.Error))
	}

	// Final states
	if opts.ShowStates && len(trace.FinalStates) > 0 {
		sb.WriteString("\nFinal States:\n")
		for entity, state := range trace.FinalStates {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", entity, state))
		}
	}

	return sb.String()
}

func (r *TraceRenderer) renderTextStep(sb *strings.Builder, step *pidl.ExecutionStep, opts TraceTextOptions) {
	prefix := "  "
	symbol := "[OK]"
	if step.Skipped {
		symbol = "[--]"
	}

	label := step.Action
	if step.Label != "" {
		label = step.Label
	}

	if opts.Compact {
		sb.WriteString(fmt.Sprintf("%s%s %d: %s -> %s: %s",
			prefix, symbol, step.StepNumber, step.From, step.To, label))
		if step.Skipped {
			sb.WriteString(fmt.Sprintf(" (skipped: %s)", step.SkipReason))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString(fmt.Sprintf("%s%s Step %d\n", prefix, symbol, step.StepNumber))
		sb.WriteString(fmt.Sprintf("%s    Flow: %s -> %s\n", prefix, step.From, step.To))
		sb.WriteString(fmt.Sprintf("%s    Action: %s\n", prefix, label))

		if step.Mode != "" {
			sb.WriteString(fmt.Sprintf("%s    Mode: %s\n", prefix, step.Mode))
		}

		if step.Phase != "" {
			sb.WriteString(fmt.Sprintf("%s    Phase: %s\n", prefix, step.Phase))
		}

		if step.Condition != "" {
			condMet := "unknown"
			if step.ConditionMet != nil {
				if *step.ConditionMet {
					condMet = "true"
				} else {
					condMet = "false"
				}
			}
			sb.WriteString(fmt.Sprintf("%s    Condition: %s (%s)\n", prefix, step.Condition, condMet))
		}

		if step.Skipped {
			sb.WriteString(fmt.Sprintf("%s    Skipped: %s\n", prefix, step.SkipReason))
		}

		if len(step.StateChanges) > 0 && opts.ShowStates {
			sb.WriteString(fmt.Sprintf("%s    State Changes:\n", prefix))
			for _, sc := range step.StateChanges {
				if sc.FromState != "" {
					sb.WriteString(fmt.Sprintf("%s      %s: %s -> %s\n", prefix, sc.Entity, sc.FromState, sc.ToState))
				} else {
					sb.WriteString(fmt.Sprintf("%s      %s: -> %s\n", prefix, sc.Entity, sc.ToState))
				}
			}
		}

		sb.WriteString("\n")
	}
}

// RenderMermaid renders the trace as a Mermaid sequence diagram.
func (r *TraceRenderer) RenderMermaid(trace *pidl.ExecutionTrace, p *pidl.Protocol) string {
	var sb strings.Builder

	sb.WriteString("sequenceDiagram\n")
	sb.WriteString(fmt.Sprintf("    %%{init: {'sequence': {'mirrorActors': false}}}%%\n"))
	sb.WriteString(fmt.Sprintf("    title Execution Trace: %s\n\n", escapeLabel(trace.ProtocolName)))

	// Declare participants based on protocol entities
	participants := make(map[string]bool)
	for _, step := range trace.Steps {
		participants[step.From] = true
		participants[step.To] = true
	}

	for _, e := range p.Entities {
		if participants[e.ID] {
			displayName := e.Name
			if displayName == "" {
				displayName = e.ID
			}
			sb.WriteString(fmt.Sprintf("    participant %s as %s\n", e.ID, escapeLabel(displayName)))
		}
	}
	sb.WriteString("\n")

	// Render steps
	for _, step := range trace.Steps {
		arrow := "->>"
		if step.Mode == pidl.FlowModeResponse || step.Mode == pidl.FlowModeToolResult {
			arrow = "-->>"
		}
		if step.Mode == pidl.FlowModeRedirect || step.Mode == pidl.FlowModeCallback {
			arrow = "--))"
		}

		label := step.Action
		if step.Label != "" {
			label = step.Label
		}

		if step.Skipped && r.HighlightSkipped {
			sb.WriteString(fmt.Sprintf("    note over %s,%s: SKIPPED: %s\n", step.From, step.To, escapeLabel(step.SkipReason)))
			sb.WriteString(fmt.Sprintf("    rect rgba(200,200,200,0.3)\n"))
			sb.WriteString(fmt.Sprintf("        %s%s%s: %s (skipped)\n", step.From, arrow, step.To, escapeLabel(label)))
			sb.WriteString(fmt.Sprintf("    end\n"))
		} else {
			sb.WriteString(fmt.Sprintf("    %s%s%s: %s\n", step.From, arrow, step.To, escapeLabel(label)))

			// Show state changes as notes
			if r.ShowStates && len(step.StateChanges) > 0 {
				for _, sc := range step.StateChanges {
					change := fmt.Sprintf("%s -> %s", sc.FromState, sc.ToState)
					if sc.FromState == "" {
						change = fmt.Sprintf("-> %s", sc.ToState)
					}
					sb.WriteString(fmt.Sprintf("    note right of %s: State: %s\n", sc.Entity, escapeLabel(change)))
				}
			}
		}
	}

	// Add summary note
	sb.WriteString("\n")
	status := "Completed"
	if !trace.Completed {
		status = "Partial"
	}
	sb.WriteString(fmt.Sprintf("    note over %s: Steps: %d, Skipped: %d, Status: %s\n",
		trace.Steps[0].From, trace.StepCount(), trace.SkippedCount(), status))

	return sb.String()
}

// RenderSVG renders the trace as an SVG sequence diagram.
func (r *TraceRenderer) RenderSVG(trace *pidl.ExecutionTrace, p *pidl.Protocol) (string, error) {
	var sb strings.Builder

	// Collect participants
	participants := make([]string, 0)
	participantMap := make(map[string]bool)
	participantNames := make(map[string]string)

	for _, step := range trace.Steps {
		if !participantMap[step.From] {
			participantMap[step.From] = true
			participants = append(participants, step.From)
		}
		if !participantMap[step.To] {
			participantMap[step.To] = true
			participants = append(participants, step.To)
		}
	}

	// Get display names from protocol
	for _, e := range p.Entities {
		if participantMap[e.ID] {
			name := e.Name
			if name == "" {
				name = e.ID
			}
			participantNames[e.ID] = name
		}
	}

	// Layout calculations
	const (
		headerHeight     = 80
		participantWidth = 150
		participantGap   = 50
		stepHeight       = 60
		padding          = 40
		actorHeight      = 50
		actorWidth       = 120
	)

	numParticipants := len(participants)
	numSteps := len(trace.Steps)

	width := padding*2 + numParticipants*participantWidth + (numParticipants-1)*participantGap
	height := headerHeight + actorHeight + numSteps*stepHeight + padding*2

	// SVG header
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
<defs>
  <style>
    .title { font: bold 18px sans-serif; fill: #333; }
    .actor { font: 14px sans-serif; fill: #333; }
    .actor-box { fill: #fff; stroke: #333; stroke-width: 2; }
    .lifeline { stroke: #999; stroke-width: 1; stroke-dasharray: 5,5; }
    .message { font: 12px sans-serif; fill: #333; }
    .arrow { fill: none; stroke: #333; stroke-width: 2; }
    .arrow-head { fill: #333; }
    .skipped { fill: #ccc; stroke: #999; }
    .skipped-text { fill: #666; font-style: italic; }
    .state-change { font: 11px sans-serif; fill: #007acc; }
  </style>
  <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">
    <polygon class="arrow-head" points="0 0, 10 3.5, 0 7" />
  </marker>
  <marker id="arrowhead-dashed" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">
    <polygon fill="#666" points="0 0, 10 3.5, 0 7" />
  </marker>
</defs>
`, width, height, width, height))

	// Title
	sb.WriteString(fmt.Sprintf(`<text class="title" x="%d" y="30" text-anchor="middle">Execution Trace: %s</text>
`, width/2, escapeXML(trace.ProtocolName)))

	// Calculate participant positions
	participantX := make(map[string]int)
	for i, pid := range participants {
		x := padding + i*(participantWidth+participantGap) + participantWidth/2
		participantX[pid] = x
	}

	// Draw actors
	actorY := headerHeight
	for _, pid := range participants {
		x := participantX[pid]
		name := participantNames[pid]
		if name == "" {
			name = pid
		}

		sb.WriteString(fmt.Sprintf(`<rect class="actor-box" x="%d" y="%d" width="%d" height="%d" rx="5"/>
`, x-actorWidth/2, actorY, actorWidth, actorHeight))
		sb.WriteString(fmt.Sprintf(`<text class="actor" x="%d" y="%d" text-anchor="middle">%s</text>
`, x, actorY+actorHeight/2+5, escapeXML(name)))
	}

	// Draw lifelines
	lifelineStartY := actorY + actorHeight
	lifelineEndY := height - padding
	for _, pid := range participants {
		x := participantX[pid]
		sb.WriteString(fmt.Sprintf(`<line class="lifeline" x1="%d" y1="%d" x2="%d" y2="%d"/>
`, x, lifelineStartY, x, lifelineEndY))
	}

	// Draw steps
	stepStartY := lifelineStartY + 20
	for i, step := range trace.Steps {
		y := stepStartY + i*stepHeight

		fromX := participantX[step.From]
		toX := participantX[step.To]

		label := step.Action
		if step.Label != "" {
			label = step.Label
		}

		arrowClass := "arrow"
		markerEnd := "url(#arrowhead)"
		if step.Mode == pidl.FlowModeResponse || step.Mode == pidl.FlowModeToolResult {
			arrowClass = "arrow"
			markerEnd = "url(#arrowhead-dashed)"
		}

		if step.Skipped {
			// Skipped step - draw grayed out
			sb.WriteString(fmt.Sprintf(`<line class="%s skipped" x1="%d" y1="%d" x2="%d" y2="%d" marker-end="%s" stroke-dasharray="5,3"/>
`, arrowClass, fromX, y, toX, y, markerEnd))
			sb.WriteString(fmt.Sprintf(`<text class="message skipped-text" x="%d" y="%d" text-anchor="middle">%s (skipped)</text>
`, (fromX+toX)/2, y-10, escapeXML(label)))
		} else {
			// Normal step
			sb.WriteString(fmt.Sprintf(`<line class="%s" x1="%d" y1="%d" x2="%d" y2="%d" marker-end="%s"/>
`, arrowClass, fromX, y, toX, y, markerEnd))
			sb.WriteString(fmt.Sprintf(`<text class="message" x="%d" y="%d" text-anchor="middle">%s</text>
`, (fromX+toX)/2, y-10, escapeXML(label)))

			// Show state changes
			if r.ShowStates && len(step.StateChanges) > 0 {
				for j, sc := range step.StateChanges {
					entityX := participantX[sc.Entity]
					change := fmt.Sprintf("%s->%s", sc.FromState, sc.ToState)
					if sc.FromState == "" {
						change = fmt.Sprintf("->%s", sc.ToState)
					}
					sb.WriteString(fmt.Sprintf(`<text class="state-change" x="%d" y="%d">%s</text>
`, entityX+10, y+15+j*12, escapeXML(change)))
				}
			}
		}
	}

	sb.WriteString("</svg>\n")
	return sb.String(), nil
}

// RenderJSON renders the trace as JSON (simply marshals the trace).
func (r *TraceRenderer) RenderJSON(trace *pidl.ExecutionTrace) ([]byte, error) {
	return trace.ToJSON()
}

// TraceRendererInterface defines the interface for trace renderers.
type TraceRendererInterface interface {
	RenderText(trace *pidl.ExecutionTrace, opts TraceTextOptions) string
	RenderMermaid(trace *pidl.ExecutionTrace, p *pidl.Protocol) string
	RenderSVG(trace *pidl.ExecutionTrace, p *pidl.Protocol) (string, error)
}

// Compile-time interface check.
var _ TraceRendererInterface = (*TraceRenderer)(nil)

// escapeLabel escapes special characters for Mermaid labels.
func escapeLabel(label string) string {
	label = strings.ReplaceAll(label, "#", "&#35;")
	label = strings.ReplaceAll(label, ":", "&#58;")
	label = strings.ReplaceAll(label, "\"", "&quot;")
	return label
}

// escapeXML escapes special characters for XML/SVG.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// Render implements the Renderer interface for trace format.
func (r *TraceRenderer) Render(w io.Writer, trace *pidl.ExecutionTrace, opts TraceTextOptions) error {
	_, err := w.Write([]byte(r.RenderText(trace, opts)))
	return err
}
