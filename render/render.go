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
	FormatPlantUML Format = "plantuml"
	FormatMermaid  Format = "mermaid"
	FormatDOT      Format = "dot"
	FormatD2       Format = "d2"
	FormatD2Flow   Format = "d2-flow"
	FormatD2Arch   Format = "d2-arch"
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
	case FormatMermaid:
		return ".mmd"
	case FormatDOT:
		return ".dot"
	case FormatD2, FormatD2Flow, FormatD2Arch:
		return ".d2"
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
	case "dot", "graphviz", "gv":
		return FormatDOT, nil
	case "d2", "d2-sequence", "d2-seq":
		return FormatD2, nil
	case "d2-flow", "d2-dataflow":
		return FormatD2Flow, nil
	case "d2-arch", "d2-architecture":
		return FormatD2Arch, nil
	default:
		return "", fmt.Errorf("unknown format %q: valid formats are plantuml, mermaid, dot, d2, d2-flow, d2-arch", s)
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

// New creates a Renderer for the specified format.
func New(format Format) (Renderer, error) {
	switch format {
	case FormatPlantUML:
		return NewPlantUML(), nil
	case FormatMermaid:
		return NewMermaid(), nil
	case FormatDOT:
		return NewDOT(), nil
	case FormatD2:
		return NewD2(), nil
	case FormatD2Flow:
		return NewD2Flow(), nil
	case FormatD2Arch:
		return NewD2Arch(), nil
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
	return []Format{FormatPlantUML, FormatMermaid, FormatDOT, FormatD2, FormatD2Flow, FormatD2Arch}
}
