package render

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/grokify/pidl"
)

// ScenarioMatrixRenderer generates a Markdown table of entities (rows, in
// entities order) by scenarios (columns, main scenarios then fragments).
// A cell is "●" when the scenario's own flows use the entity and "○" when
// the entity takes part only through an included fragment.
type ScenarioMatrixRenderer struct {
	// Title adds a heading with the protocol name.
	Title bool
	// IncludeFragments adds a column per fragment after the main scenarios.
	IncludeFragments bool
}

// NewScenarioMatrix creates a ScenarioMatrixRenderer with fragments shown.
func NewScenarioMatrix() *ScenarioMatrixRenderer {
	return &ScenarioMatrixRenderer{Title: true, IncludeFragments: true}
}

// Format returns the output format identifier.
func (r *ScenarioMatrixRenderer) Format() Format {
	return FormatScenarioMatrix
}

// Render writes the matrix to w.
func (r *ScenarioMatrixRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// RenderString returns the matrix as Markdown.
func (r *ScenarioMatrixRenderer) RenderString(p *pidl.Protocol) (string, error) {
	var sb strings.Builder
	if r.Title {
		fmt.Fprintf(&sb, "# %s - Roles by Scenario\n\n", p.ProtocolMeta.Name)
	}

	var cols []pidl.ScenarioSummary
	for _, s := range p.ScenarioSummaries() {
		if s.Fragment && !r.IncludeFragments {
			continue
		}
		cols = append(cols, s)
	}
	if len(cols) == 0 {
		sb.WriteString("*No scenarios defined.*\n")
		return sb.String(), nil
	}

	sb.WriteString("| Role |")
	for _, c := range cols {
		name := c.Name
		if c.Fragment {
			name += " (fragment)"
		}
		fmt.Fprintf(&sb, " %d. %s |", c.Position, escapeCell(name))
	}
	sb.WriteString("\n|------|")
	for range cols {
		sb.WriteString(":---:|")
	}
	sb.WriteString("\n")

	direct := make(map[string][]string, len(cols))
	for _, c := range cols {
		direct[c.ID] = p.DirectParticipants(c.ID)
	}
	for _, e := range p.Entities {
		fmt.Fprintf(&sb, "| **%s** |", escapeCell(e.Name))
		for _, c := range cols {
			switch {
			case slices.Contains(direct[c.ID], e.ID):
				sb.WriteString(" ● |")
			case slices.Contains(c.Participants, e.ID):
				sb.WriteString(" ○ |")
			default:
				sb.WriteString("  |")
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n● takes part directly · ○ takes part only through an included fragment\n")
	return sb.String(), nil
}

// escapeCell escapes characters that would break a Markdown table cell.
func escapeCell(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}
