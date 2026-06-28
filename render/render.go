// Package render provides diagram renderers for PIDL protocols.
package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// Format represents a diagram output format.
type Format string

const (
	FormatPlantUML         Format = "plantuml"
	FormatMermaid          Format = "mermaid"
	FormatMermaidState     Format = "mermaid-state"
	FormatMermaidComponent Format = "mermaid-component"
	FormatMermaidTrust     Format = "mermaid-trust"
	FormatMarkdownMatrix   Format = "markdown-matrix"
	FormatDOT              Format = "dot"
	FormatD2               Format = "d2"
	FormatD2Flow           Format = "d2-flow"
	FormatD2Arch           Format = "d2-arch"
	FormatSVG              Format = "svg"
	FormatSVGAnimated      Format = "svg-animated"
	FormatSVGNetwork       Format = "svg-network"
	FormatSVGComponent     Format = "svg-component"
	FormatSVGTrust         Format = "svg-trust"
)

// String returns the format as a string.
func (f Format) String() string {
	return string(f)
}

// FileExtension returns the conventional file extension for this format.
func (f Format) FileExtension() string {
	switch f {
	case FormatPlantUML:
		return ".puml"
	case FormatMermaid, FormatMermaidState, FormatMermaidComponent, FormatMermaidTrust:
		return ".mmd"
	case FormatMarkdownMatrix:
		return ".md"
	case FormatDOT:
		return ".dot"
	case FormatD2, FormatD2Flow, FormatD2Arch:
		return ".d2"
	case FormatSVG, FormatSVGAnimated, FormatSVGNetwork, FormatSVGComponent, FormatSVGTrust:
		return ".svg"
	default:
		return ".txt"
	}
}

// ParseFormat parses a format string, returning an error if invalid.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "plantuml", "puml", "uml":
		return FormatPlantUML, nil
	case "mermaid", "mmd":
		return FormatMermaid, nil
	case "mermaid-state", "mmd-state":
		return FormatMermaidState, nil
	case "mermaid-component", "mmd-component", "component":
		return FormatMermaidComponent, nil
	case "mermaid-trust", "mmd-trust", "trust":
		return FormatMermaidTrust, nil
	case "markdown-matrix", "md-matrix", "matrix":
		return FormatMarkdownMatrix, nil
	case "dot", "graphviz", "gv":
		return FormatDOT, nil
	case "d2", "d2-sequence", "d2-seq":
		return FormatD2, nil
	case "d2-flow", "d2-dataflow":
		return FormatD2Flow, nil
	case "d2-arch", "d2-architecture":
		return FormatD2Arch, nil
	case "svg":
		return FormatSVG, nil
	case "svg-animated", "svg-anim":
		return FormatSVGAnimated, nil
	case "svg-network", "svg-net":
		return FormatSVGNetwork, nil
	case "svg-component", "svg-comp":
		return FormatSVGComponent, nil
	case "svg-trust":
		return FormatSVGTrust, nil
	default:
		return "", fmt.Errorf("unknown format %q: valid formats are plantuml, mermaid, mermaid-state, mermaid-component, mermaid-trust, markdown-matrix, dot, d2, d2-flow, d2-arch, svg, svg-animated, svg-network, svg-component, svg-trust", s)
	}
}

// MustParseFormat parses a format string, panicking on error.
func MustParseFormat(s string) Format {
	f, err := ParseFormat(s)
	if err != nil {
		panic(err)
	}
	return f
}

// Renderer generates diagram output from a Protocol.
type Renderer interface {
	// Render writes the diagram to the writer.
	Render(w io.Writer, p *pidl.Protocol) error

	// RenderString returns the diagram as a string.
	RenderString(p *pidl.Protocol) (string, error)

	// Format returns the output format name.
	Format() Format
}

// SequenceRenderOptions contains common options for sequence diagram renderers.
// Renderers can embed this struct to get standard options.
type SequenceRenderOptions struct {
	// Title includes the protocol name as diagram title.
	Title bool

	// ShowNotes renders flow notes.
	ShowNotes bool

	// ShowAnnotations renders flow annotations.
	ShowAnnotations bool

	// ShowConditions wraps conditional flows in opt blocks.
	ShowConditions bool

	// ShowAlternatives renders alternative paths.
	ShowAlternatives bool

	// ShowSecurity renders security requirement notes/badges.
	ShowSecurity bool
}

// DefaultSequenceRenderOptions returns the default sequence render options.
func DefaultSequenceRenderOptions() SequenceRenderOptions {
	return SequenceRenderOptions{
		Title:            true,
		ShowNotes:        true,
		ShowAnnotations:  true,
		ShowConditions:   true,
		ShowAlternatives: true,
		ShowSecurity:     true,
	}
}

// New creates a Renderer for the specified format.
func New(format Format) (Renderer, error) {
	switch format {
	case FormatPlantUML:
		return NewPlantUML(), nil
	case FormatMermaid:
		return NewMermaid(), nil
	case FormatMermaidState:
		return NewMermaidState(), nil
	case FormatMermaidComponent:
		return NewMermaidComponent(), nil
	case FormatMermaidTrust:
		return NewMermaidTrust(), nil
	case FormatMarkdownMatrix:
		return NewMarkdownMatrix(), nil
	case FormatDOT:
		return NewDOT(), nil
	case FormatD2:
		return NewD2(), nil
	case FormatD2Flow:
		return NewD2Flow(), nil
	case FormatD2Arch:
		return NewD2Arch(), nil
	case FormatSVG:
		return NewSVG(), nil
	case FormatSVGAnimated:
		return NewSVGAnimated(), nil
	case FormatSVGNetwork:
		return NewSVGNetwork(), nil
	case FormatSVGComponent:
		return NewSVGComponent(), nil
	case FormatSVGTrust:
		return NewSVGTrust(), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

// MustNew creates a Renderer for the specified format, panicking on error.
func MustNew(format Format) Renderer {
	r, err := New(format)
	if err != nil {
		panic(err)
	}
	return r
}

// RenderString renders a protocol using the specified format.
func RenderString(format Format, p *pidl.Protocol) (string, error) {
	r, err := New(format)
	if err != nil {
		return "", err
	}
	return r.RenderString(p)
}

// SupportedFormats returns all supported output formats.
func SupportedFormats() []Format {
	return []Format{
		FormatPlantUML,
		FormatMermaid,
		FormatMermaidState,
		FormatMermaidComponent,
		FormatMermaidTrust,
		FormatMarkdownMatrix,
		FormatDOT,
		FormatD2,
		FormatD2Flow,
		FormatD2Arch,
		FormatSVG,
		FormatSVGAnimated,
		FormatSVGNetwork,
		FormatSVGComponent,
		FormatSVGTrust,
	}
}
