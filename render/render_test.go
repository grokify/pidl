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
		{"mermaid-component", FormatMermaidComponent, false},
		{"mmd-component", FormatMermaidComponent, false},
		{"component", FormatMermaidComponent, false},
		{"mermaid-trust", FormatMermaidTrust, false},
		{"mmd-trust", FormatMermaidTrust, false},
		{"trust", FormatMermaidTrust, false},
		{"markdown-matrix", FormatMarkdownMatrix, false},
		{"md-matrix", FormatMarkdownMatrix, false},
		{"matrix", FormatMarkdownMatrix, false},
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
		{"svg-component", FormatSVGComponent, false},
		{"svg-comp", FormatSVGComponent, false},
		{"svg-trust", FormatSVGTrust, false},
		{"infographic", FormatInfographic, false},
		{"infog", FormatInfographic, false},
		{"ig", FormatInfographic, false},
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
		{FormatMermaidComponent, false},
		{FormatMermaidTrust, false},
		{FormatMarkdownMatrix, false},
		{FormatDOT, false},
		{FormatD2, false},
		{FormatD2Flow, false},
		{FormatD2Arch, false},
		{FormatSVGComponent, false},
		{FormatSVGTrust, false},
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
	if len(formats) != 16 {
		t.Errorf("SupportedFormats() = %d formats, want 16", len(formats))
	}
}

func TestRenderString(t *testing.T) {
	p := testProtocol()

	for _, format := range SupportedFormats() {
		// Skip infographic - it uses a different renderer API
		if format == FormatInfographic {
			continue
		}
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

	// Check phase box
	if !strings.Contains(s, `box "Authorization"`) {
		t.Error("PlantUML should contain phase box")
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

// testProcessSpec returns a sample process spec for testing.
func testProcessSpec() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "vision-pipeline",
			Name: "VisionSpec Execution Pipeline",
			Kind: pidl.ProtocolKindProcess,
		},
		Entities: []pidl.Entity{
			{ID: "loader", Name: "Document Loader", Type: pidl.EntityTypeServer, StepType: pidl.StepTypeDeterministic},
			{ID: "analyzer", Name: "LLM Analyzer", Type: pidl.EntityTypeServer, StepType: pidl.StepTypeLLM},
			{ID: "reviewer", Name: "Human Review", Type: pidl.EntityTypeUser, StepType: pidl.StepTypeHuman},
			{ID: "extractor", Name: "Data Extractor", Type: pidl.EntityTypeServer, StepType: pidl.StepTypeTool},
		},
		Flows: []pidl.Flow{
			{From: "loader", To: "analyzer", Action: "send", Label: "Document", Mode: pidl.FlowModeRequest},
			{From: "analyzer", To: "reviewer", Action: "submit", Label: "Analysis", Mode: pidl.FlowModeRequest},
			{From: "reviewer", To: "extractor", Action: "approve", Label: "Reviewed Data", Mode: pidl.FlowModeRequest},
		},
	}
}

// testProcessSpecWithDataPorts returns a process spec with data ports for testing.
func testProcessSpecWithDataPorts() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "etl-pipeline",
			Name: "ETL Pipeline",
			Kind: pidl.ProtocolKindProcess,
		},
		Entities: []pidl.Entity{
			{
				ID:       "extractor",
				Name:     "Data Extractor",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeDeterministic,
				Inputs: []pidl.DataPort{
					{Kind: pidl.DataPortKindFile, Name: "input.csv", Description: "Raw CSV data", Required: true},
					{Kind: pidl.DataPortKindAPI, Name: "config_api", Description: "Configuration endpoint"},
				},
				Outputs: []pidl.DataPort{
					{Kind: pidl.DataPortKindObject, Name: "raw_data", Description: "Parsed raw data"},
				},
			},
			{
				ID:       "transformer",
				Name:     "Data Transformer",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeLLM,
				Inputs: []pidl.DataPort{
					{Kind: pidl.DataPortKindObject, Name: "raw_data", Required: true},
				},
				Outputs: []pidl.DataPort{
					{Kind: pidl.DataPortKindObject, Name: "transformed_data"},
				},
			},
			{
				ID:       "loader",
				Name:     "Data Loader",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeDeterministic,
				Inputs: []pidl.DataPort{
					{Kind: pidl.DataPortKindObject, Name: "transformed_data", Required: true},
				},
				Outputs: []pidl.DataPort{
					{Kind: pidl.DataPortKindDatabase, Name: "results_db", Description: "Output database"},
					{Kind: pidl.DataPortKindQueue, Name: "notifications", Description: "Notification queue"},
				},
			},
		},
		Flows: []pidl.Flow{
			{From: "extractor", To: "transformer", Action: "send", Label: "Raw Data", Mode: pidl.FlowModeRequest},
			{From: "transformer", To: "loader", Action: "send", Label: "Transformed Data", Mode: pidl.FlowModeRequest},
		},
	}
}

func TestPlantUMLProcessSpec(t *testing.T) {
	p := testProcessSpec()
	r := NewPlantUML()
	r.ShowStepTypes = true

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for step type skinparams
	if !strings.Contains(s, "skinparam participant") {
		t.Error("PlantUML process spec should have participant skinparams")
	}

	// Check for step type stereotypes
	if !strings.Contains(s, "<<deterministic>>") {
		t.Error("PlantUML process spec should have deterministic stereotype")
	}
	if !strings.Contains(s, "<<llm>>") {
		t.Error("PlantUML process spec should have llm stereotype")
	}
	if !strings.Contains(s, "<<human>>") {
		t.Error("PlantUML process spec should have human stereotype")
	}
	if !strings.Contains(s, "<<tool>>") {
		t.Error("PlantUML process spec should have tool stereotype")
	}
}

func TestMermaidProcessSpec(t *testing.T) {
	p := testProcessSpec()
	r := NewMermaid()
	r.ShowStepTypes = true

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for step type emoji badges
	if !strings.Contains(s, "⚙️") {
		t.Error("Mermaid process spec should have deterministic badge (⚙️)")
	}
	if !strings.Contains(s, "🧠") {
		t.Error("Mermaid process spec should have LLM badge (🧠)")
	}
	if !strings.Contains(s, "👤") {
		t.Error("Mermaid process spec should have human badge (👤)")
	}
	if !strings.Contains(s, "🔧") {
		t.Error("Mermaid process spec should have tool badge (🔧)")
	}
}

func TestD2ProcessSpec(t *testing.T) {
	p := testProcessSpec()
	// Use D2Flow style which supports entity styling
	r := NewD2Flow()
	r.ShowStepTypes = true

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for step type color styling
	if !strings.Contains(s, "#E3F2FD") || !strings.Contains(s, "#1976D2") {
		t.Error("D2 process spec should have deterministic step colors (blue)")
	}
	if !strings.Contains(s, "#F3E5F5") || !strings.Contains(s, "#7B1FA2") {
		t.Error("D2 process spec should have LLM step colors (purple)")
	}
	if !strings.Contains(s, "#E8F5E9") || !strings.Contains(s, "#388E3C") {
		t.Error("D2 process spec should have human step colors (green)")
	}
	if !strings.Contains(s, "#ECEFF1") || !strings.Contains(s, "#607D8B") {
		t.Error("D2 process spec should have tool step colors (gray)")
	}
}

func TestSVGProcessSpec(t *testing.T) {
	p := testProcessSpec()
	r := NewSVG()
	r.ShowStepTypes = true

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check SVG structure
	if !strings.HasPrefix(s, "<svg") {
		t.Error("SVG should start with <svg")
	}
	if !strings.Contains(s, "</svg>") {
		t.Error("SVG should contain closing </svg>")
	}

	// Check for step type emoji badges in participant text
	if !strings.Contains(s, "⚙️") {
		t.Error("SVG process spec should have deterministic badge (⚙️)")
	}
	if !strings.Contains(s, "🧠") {
		t.Error("SVG process spec should have LLM badge (🧠)")
	}
	if !strings.Contains(s, "👤") {
		t.Error("SVG process spec should have human badge (👤)")
	}
	if !strings.Contains(s, "🔧") {
		t.Error("SVG process spec should have tool badge (🔧)")
	}

	// Check for step type inline styles on participant boxes
	if !strings.Contains(s, "fill:#E3F2FD") {
		t.Error("SVG process spec should have deterministic fill color")
	}
	if !strings.Contains(s, "fill:#F3E5F5") {
		t.Error("SVG process spec should have LLM fill color")
	}
	if !strings.Contains(s, "fill:#E8F5E9") {
		t.Error("SVG process spec should have human fill color")
	}
	if !strings.Contains(s, "fill:#ECEFF1") {
		t.Error("SVG process spec should have tool fill color")
	}
}

func TestProcessSpecRenderAllFormats(t *testing.T) {
	p := testProcessSpec()

	// Test all sequence diagram formats can render process specs
	formats := []Format{
		FormatPlantUML,
		FormatMermaid,
		FormatD2,
		FormatSVG,
	}

	for _, format := range formats {
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

func TestD2FlowDataPorts(t *testing.T) {
	p := testProcessSpecWithDataPorts()
	r := NewD2Flow()
	r.ShowDataPorts = true

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for data port shapes
	if !strings.Contains(s, "shape: page") {
		t.Error("D2 flow should have page shape for file ports")
	}
	if !strings.Contains(s, "shape: cylinder") {
		t.Error("D2 flow should have cylinder shape for database ports")
	}
	if !strings.Contains(s, "shape: cloud") {
		t.Error("D2 flow should have cloud shape for API ports")
	}
	if !strings.Contains(s, "shape: queue") {
		t.Error("D2 flow should have queue shape for queue ports")
	}

	// Check for data port icons
	if !strings.Contains(s, "📄") {
		t.Error("D2 flow should have file icon (📄)")
	}
	if !strings.Contains(s, "🗄️") {
		t.Error("D2 flow should have database icon (🗄️)")
	}
	if !strings.Contains(s, "🌐") {
		t.Error("D2 flow should have API icon (🌐)")
	}
	if !strings.Contains(s, "📬") {
		t.Error("D2 flow should have queue icon (📬)")
	}

	// Check for data port names
	if !strings.Contains(s, "input.csv") {
		t.Error("D2 flow should contain input.csv port")
	}
	if !strings.Contains(s, "results_db") {
		t.Error("D2 flow should contain results_db port")
	}

	// Check for data port connections section
	if !strings.Contains(s, "# Data Port Connections") {
		t.Error("D2 flow should have data port connections section")
	}
}

func TestD2ArchDataPorts(t *testing.T) {
	p := testProcessSpecWithDataPorts()
	r := NewD2Arch()
	r.ShowDataPorts = true

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for grouped data ports
	if !strings.Contains(s, "Files:") {
		t.Error("D2 arch should have Files group")
	}
	if !strings.Contains(s, "Databases:") {
		t.Error("D2 arch should have Databases group")
	}
	if !strings.Contains(s, "APIs:") {
		t.Error("D2 arch should have APIs group")
	}
	if !strings.Contains(s, "Queues:") {
		t.Error("D2 arch should have Queues group")
	}

	// Check for qualified port IDs in connections
	if !strings.Contains(s, "Files.") || !strings.Contains(s, "Databases.") {
		t.Error("D2 arch should use qualified IDs for port connections")
	}
}

func TestD2FlowDataPortsDisabled(t *testing.T) {
	p := testProcessSpecWithDataPorts()
	r := NewD2Flow()
	r.ShowDataPorts = false

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should not contain data port elements
	if strings.Contains(s, "# Data Ports") {
		t.Error("D2 flow with ShowDataPorts=false should not have data ports section")
	}
	if strings.Contains(s, "port_input") {
		t.Error("D2 flow with ShowDataPorts=false should not have port IDs")
	}
}
