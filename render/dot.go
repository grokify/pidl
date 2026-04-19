package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// DOTRenderer renders PIDL protocols as Graphviz DOT data flow diagrams.
type DOTRenderer struct {
	// Title includes the protocol name as graph label.
	Title bool

	// RankDir sets the graph direction (LR, TB, RL, BT).
	RankDir string

	// MergeEdges combines multiple flows between the same entities.
	MergeEdges bool

	// ShowPhases groups nodes by phase using subgraphs.
	ShowPhases bool

	// ShowConditions includes condition text in edge labels.
	ShowConditions bool

	// ShowAnnotations includes annotation counts in edge labels.
	ShowAnnotations bool
}

// NewDOT creates a new DOT renderer with default options.
func NewDOT() *DOTRenderer {
	return &DOTRenderer{
		Title:           true,
		RankDir:         "LR",
		MergeEdges:      true,
		ShowPhases:      false,
		ShowConditions:  true,
		ShowAnnotations: false,
	}
}

// Format returns the output format.
func (r *DOTRenderer) Format() Format {
	return FormatDOT
}

// Render writes the DOT diagram to the writer.
func (r *DOTRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the DOT diagram as a string.
func (r *DOTRenderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *DOTRenderer) render(p *pidl.Protocol) string {
	var sb strings.Builder

	sb.WriteString("digraph Protocol {\n")
	fmt.Fprintf(&sb, "    rankdir=%s;\n", r.RankDir)
	sb.WriteString("    node [shape=box, style=rounded];\n")
	sb.WriteString("    edge [fontsize=10];\n")

	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "    label=\"%s\";\n", r.escapeString(p.ProtocolMeta.Name))
		sb.WriteString("    labelloc=t;\n")
		sb.WriteString("    fontsize=14;\n")
	}

	sb.WriteString("\n")

	// Declare nodes with shapes based on entity type
	for _, e := range p.Entities {
		shape := r.entityShape(e.Type)
		fmt.Fprintf(&sb, "    %s [label=\"%s\", shape=%s];\n",
			e.ID, r.escapeString(e.Name), shape)
	}

	sb.WriteString("\n")

	// Render edges
	if r.MergeEdges {
		r.renderMergedEdges(&sb, p)
	} else {
		r.renderAllEdges(&sb, p)
	}

	sb.WriteString("}\n")

	return sb.String()
}

func (r *DOTRenderer) renderMergedEdges(sb *strings.Builder, p *pidl.Protocol) {
	// Group flows by (from, to) pair
	type edgeKey struct{ from, to string }
	edges := make(map[edgeKey][]string)
	edgeOrder := make([]edgeKey, 0)

	for _, f := range p.Flows {
		key := edgeKey{f.From, f.To}
		if _, exists := edges[key]; !exists {
			edgeOrder = append(edgeOrder, key)
		}
		edges[key] = append(edges[key], f.DisplayLabel())
	}

	for _, key := range edgeOrder {
		labels := edges[key]
		var label string
		if len(labels) == 1 {
			label = labels[0]
		} else {
			// Truncate if too many
			if len(labels) > 3 {
				label = fmt.Sprintf("%s\\n%s\\n... (+%d more)",
					r.escapeString(labels[0]),
					r.escapeString(labels[1]),
					len(labels)-2)
			} else {
				escaped := make([]string, len(labels))
				for i, l := range labels {
					escaped[i] = r.escapeString(l)
				}
				label = strings.Join(escaped, "\\n")
			}
		}
		fmt.Fprintf(sb, "    %s -> %s [label=\"%s\"];\n",
			key.from, key.to, label)
	}
}

func (r *DOTRenderer) renderAllEdges(sb *strings.Builder, p *pidl.Protocol) {
	for i, f := range p.Flows {
		style := r.modeToStyle(f.EffectiveMode())
		label := r.escapeString(f.DisplayLabel())

		// Add condition prefix if present and enabled
		if r.ShowConditions && f.HasCondition() {
			label = fmt.Sprintf("[%s]\\n%s", r.escapeString(f.Condition), label)
		}

		// Add annotation indicator if present and enabled
		if r.ShowAnnotations && f.HasAnnotations() {
			label = fmt.Sprintf("%s\\n(%d annotations)", label, len(f.Annotations))
		}

		fmt.Fprintf(sb, "    %s -> %s [label=\"%d. %s\"%s];\n",
			f.From, f.To, i+1, label, style)

		// Render alternative edges if present
		for _, alt := range f.Alternatives {
			for j, altFlow := range alt.Flows {
				altStyle := r.modeToStyle(altFlow.EffectiveMode())
				altLabel := r.escapeString(altFlow.DisplayLabel())
				if r.ShowConditions {
					altLabel = fmt.Sprintf("[ALT: %s]\\n%s", r.escapeString(alt.Condition), altLabel)
				}
				fmt.Fprintf(sb, "    %s -> %s [label=\"%d.%d. %s\"%s, color=gray];\n",
					altFlow.From, altFlow.To, i+1, j+1, altLabel, altStyle)
			}
		}
	}
}

func (r *DOTRenderer) entityShape(t pidl.EntityType) string {
	switch t {
	case pidl.EntityTypeUser:
		return "oval"
	case pidl.EntityTypeBrowser:
		return "rectangle"
	case pidl.EntityTypeClient:
		return "box"
	case pidl.EntityTypeAuthorizationServer, pidl.EntityTypeIdentityProvider:
		return "ellipse"
	case pidl.EntityTypeResourceServer, pidl.EntityTypeServer:
		return "box3d"
	case pidl.EntityTypeAgent, pidl.EntityTypeDelegatedAgent:
		return "hexagon"
	case pidl.EntityTypeToolServer:
		return "component"
	case pidl.EntityTypeTool:
		return "cds"
	default:
		return "box"
	}
}

func (r *DOTRenderer) modeToStyle(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult:
		return ", style=dashed"
	case pidl.FlowModeRedirect, pidl.FlowModeCallback:
		return ", style=bold"
	case pidl.FlowModeEvent:
		return ", style=dotted"
	default:
		return ""
	}
}

func (r *DOTRenderer) escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
