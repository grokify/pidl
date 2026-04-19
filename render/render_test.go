package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"plantuml", FormatPlantUML, false},
		{"PlantUML", FormatPlantUML, false},
		{"PLANTUML", FormatPlantUML, false},
		{"puml", FormatPlantUML, false},
		{"uml", FormatPlantUML, false},
		{"mermaid", FormatMermaid, false},
		{"Mermaid", FormatMermaid, false},
		{"mmd", FormatMermaid, false},
		{"dot", FormatDOT, false},
		{"DOT", FormatDOT, false},
		{"graphviz", FormatDOT, false},
		{"gv", FormatDOT, false},
		{"d2", FormatD2, false},
		{"D2", FormatD2, false},
		{"d2-sequence", FormatD2, false},
		{"d2-seq", FormatD2, false},
		{"d2-flow", FormatD2Flow, false},
		{"d2-dataflow", FormatD2Flow, false},
		{"d2-arch", FormatD2Arch, false},
		{"d2-architecture", FormatD2Arch, false},
		{"  plantuml  ", FormatPlantUML, false},
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFormat(%q) should error", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseFormat(%q) error = %v", tt.input, err)
				}
				if got != tt.want {
					t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestMustParseFormat(t *testing.T) {
	f := MustParseFormat("plantuml")
	if f != FormatPlantUML {
		t.Errorf("MustParseFormat(plantuml) = %v, want %v", f, FormatPlantUML)
	}
}

func TestMustParseFormatPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseFormat should panic on invalid format")
		}
	}()
	MustParseFormat("invalid")
}

func TestFormatString(t *testing.T) {
	if FormatPlantUML.String() != "plantuml" {
		t.Errorf("FormatPlantUML.String() = %q, want %q", FormatPlantUML.String(), "plantuml")
	}
}

func TestFormatFileExtension(t *testing.T) {
	tests := []struct {
		format Format
		want   string
	}{
		{FormatPlantUML, ".puml"},
		{FormatMermaid, ".mmd"},
		{FormatDOT, ".dot"},
		{FormatD2, ".d2"},
		{FormatD2Flow, ".d2"},
		{FormatD2Arch, ".d2"},
		{Format("unknown"), ".txt"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if got := tt.format.FileExtension(); got != tt.want {
				t.Errorf("FileExtension() = %q, want %q", got, tt.want)
			}
		})
	}
}

func testProtocol() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{ID: "client", Name: "Client", Type: pidl.EntityTypeClient},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Phases: []pidl.Phase{
			{ID: "auth", Name: "Authorization"},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request", Label: "Request", Mode: pidl.FlowModeRequest, Phase: "auth"},
			{From: "server", To: "client", Action: "response", Label: "Response", Mode: pidl.FlowModeResponse, Phase: "auth"},
		},
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		format  Format
		wantErr bool
	}{
		{FormatPlantUML, false},
		{FormatMermaid, false},
		{FormatDOT, false},
		{FormatD2, false},
		{FormatD2Flow, false},
		{FormatD2Arch, false},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			r, err := New(tt.format)
			if tt.wantErr {
				if err == nil {
					t.Error("New() should error on unknown format")
				}
			} else {
				if err != nil {
					t.Errorf("New() error = %v", err)
				}
				if r == nil {
					t.Error("New() returned nil renderer")
				}
				if r.Format() != tt.format {
					t.Errorf("Format() = %v, want %v", r.Format(), tt.format)
				}
			}
		})
	}
}

func TestMustNew(t *testing.T) {
	r := MustNew(FormatPlantUML)
	if r == nil {
		t.Error("MustNew() returned nil")
	}
}

func TestMustNewPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew() should panic on unknown format")
		}
	}()
	MustNew("unknown")
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) != 6 {
		t.Errorf("SupportedFormats() = %d formats, want 6", len(formats))
	}
}

func TestRenderString(t *testing.T) {
	p := testProtocol()

	for _, format := range SupportedFormats() {
		t.Run(string(format), func(t *testing.T) {
			s, err := RenderString(format, p)
			if err != nil {
				t.Errorf("RenderString() error = %v", err)
			}
			if s == "" {
				t.Error("RenderString() returned empty string")
			}
		})
	}
}

func TestPlantUMLRenderer(t *testing.T) {
	p := testProtocol()
	r := NewPlantUML()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check structure
	if !strings.HasPrefix(s, "@startuml") {
		t.Error("PlantUML should start with @startuml")
	}
	if !strings.HasSuffix(strings.TrimSpace(s), "@enduml") {
		t.Error("PlantUML should end with @enduml")
	}

	// Check participants
	if !strings.Contains(s, "participant") {
		t.Error("PlantUML should have participant declarations")
	}

	// Check flows
	if !strings.Contains(s, "client") && !strings.Contains(s, "server") {
		t.Error("PlantUML should contain entity IDs")
	}

	// Check phase separator
	if !strings.Contains(s, "== Authorization ==") {
		t.Error("PlantUML should contain phase separator")
	}

	// Check arrows
	if !strings.Contains(s, "->") {
		t.Error("PlantUML should contain arrows")
	}
}

func TestPlantUMLRendererWriter(t *testing.T) {
	p := testProtocol()
	r := NewPlantUML()

	var buf bytes.Buffer
	err := r.Render(&buf, p)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Render() wrote nothing")
	}
}

func TestMermaidRenderer(t *testing.T) {
	p := testProtocol()
	r := NewMermaid()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check structure
	if !strings.HasPrefix(s, "sequenceDiagram") {
		t.Error("Mermaid should start with sequenceDiagram")
	}

	// Check participants
	if !strings.Contains(s, "participant") {
		t.Error("Mermaid should have participant declarations")
	}

	// Check arrows
	if !strings.Contains(s, "->>") || !strings.Contains(s, "-->>") {
		t.Error("Mermaid should contain arrows")
	}
}

func TestDOTRenderer(t *testing.T) {
	p := testProtocol()
	r := NewDOT()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check structure
	if !strings.HasPrefix(s, "digraph") {
		t.Error("DOT should start with digraph")
	}
	if !strings.HasSuffix(strings.TrimSpace(s), "}") {
		t.Error("DOT should end with }")
	}

	// Check nodes
	if !strings.Contains(s, "client") || !strings.Contains(s, "server") {
		t.Error("DOT should contain node IDs")
	}

	// Check edges
	if !strings.Contains(s, "->") {
		t.Error("DOT should contain edges")
	}

	// Check rankdir
	if !strings.Contains(s, "rankdir=LR") {
		t.Error("DOT should contain rankdir")
	}
}

func TestDOTRendererMergedEdges(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []pidl.Entity{
			{ID: "a", Name: "A", Type: pidl.EntityTypeClient},
			{ID: "b", Name: "B", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "a", To: "b", Action: "x", Label: "First"},
			{From: "a", To: "b", Action: "y", Label: "Second"},
			{From: "a", To: "b", Action: "z", Label: "Third"},
		},
	}

	r := NewDOT()
	r.MergeEdges = true

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should have merged edges
	arrows := strings.Count(s, "a -> b")
	if arrows != 1 {
		t.Errorf("Merged edges should produce 1 arrow, got %d", arrows)
	}
}

func TestDOTRendererUnmergedEdges(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []pidl.Entity{
			{ID: "a", Name: "A", Type: pidl.EntityTypeClient},
			{ID: "b", Name: "B", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "a", To: "b", Action: "x"},
			{From: "a", To: "b", Action: "y"},
		},
	}

	r := NewDOT()
	r.MergeEdges = false

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should have separate edges
	arrows := strings.Count(s, "a -> b")
	if arrows != 2 {
		t.Errorf("Unmerged edges should produce 2 arrows, got %d", arrows)
	}
}

func TestD2SequenceRenderer(t *testing.T) {
	p := testProtocol()
	r := NewD2()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check structure
	if !strings.Contains(s, "shape: sequence_diagram") {
		t.Error("D2 sequence should contain shape: sequence_diagram")
	}

	// Check title
	if !strings.Contains(s, "title: Test Protocol") {
		t.Error("D2 should contain title")
	}

	// Check actors
	if !strings.Contains(s, "client: Client") || !strings.Contains(s, "server: Server") {
		t.Error("D2 should contain actor declarations")
	}

	// Check arrows
	if !strings.Contains(s, "->") {
		t.Error("D2 should contain arrows")
	}
}

func TestD2FlowRenderer(t *testing.T) {
	p := testProtocol()
	r := NewD2Flow()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check direction
	if !strings.Contains(s, "direction: right") {
		t.Error("D2 flow should contain direction")
	}

	// Check nodes have shapes
	if !strings.Contains(s, "shape:") {
		t.Error("D2 flow should contain shape declarations")
	}

	// Check connections with numbers
	if !strings.Contains(s, "1. Request") {
		t.Error("D2 flow should contain numbered flows")
	}
}

func TestD2ArchRenderer(t *testing.T) {
	p := testProtocol()
	r := NewD2Arch()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check that entities are grouped
	if !strings.Contains(s, "Clients:") || !strings.Contains(s, "Servers:") {
		t.Error("D2 arch should group entities by type")
	}

	// Check connections use qualified IDs
	if !strings.Contains(s, "Clients.client") || !strings.Contains(s, "Servers.server") {
		t.Error("D2 arch should use qualified IDs for connections")
	}
}

func TestD2RendererWriter(t *testing.T) {
	p := testProtocol()
	r := NewD2()

	var buf bytes.Buffer
	err := r.Render(&buf, p)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Render() wrote nothing")
	}
}

func TestD2RendererFormat(t *testing.T) {
	tests := []struct {
		renderer *D2Renderer
		want     Format
	}{
		{NewD2(), FormatD2},
		{NewD2Flow(), FormatD2Flow},
		{NewD2Arch(), FormatD2Arch},
	}

	for _, tt := range tests {
		if got := tt.renderer.Format(); got != tt.want {
			t.Errorf("Format() = %v, want %v", got, tt.want)
		}
	}
}
