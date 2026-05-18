package render_test

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/render"
)

func TestSVGNetworkRenderer_Format(t *testing.T) {
	renderer := render.NewSVGNetwork()
	if got := renderer.Format(); got != render.FormatSVGNetwork {
		t.Errorf("Format() = %v, want %v", got, render.FormatSVGNetwork)
	}
}

func TestSVGNetworkRenderer_RenderString(t *testing.T) {
	protocol := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test-protocol",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{ID: "client", Name: "Client", Type: pidl.EntityTypeClient},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request", Label: "Send Request"},
			{From: "server", To: "client", Action: "response", Label: "Return Response"},
		},
	}

	renderer := render.NewSVGNetwork()
	svg, err := renderer.RenderString(protocol)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Verify SVG structure
	checks := []struct {
		name     string
		contains string
	}{
		{"SVG header", "<svg viewBox="},
		{"style block", "<style>"},
		{"boundaries group", `class="boundaries"`},
		{"entities group", `class="entities"`},
		{"connections group", `class="connections"`},
		{"entity box", `class="entity-box"`},
		{"boundary", `class="boundary`},
		{"connection", `class="connection"`},
		{"network CSS", ".boundary-trusted"},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if !strings.Contains(svg, check.contains) {
				t.Errorf("SVG should contain %q", check.contains)
			}
		})
	}
}

func TestSVGNetworkRenderer_WithMetadata(t *testing.T) {
	protocol := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test-protocol",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:       "client",
				Name:     "Client",
				Type:     pidl.EntityTypeClient,
				Metadata: &pidl.EntityMetadata{Network: "external"},
			},
			{
				ID:       "gateway",
				Name:     "API Gateway",
				Type:     pidl.EntityTypeServer,
				Metadata: &pidl.EntityMetadata{Network: "dmz"},
			},
			{
				ID:       "api",
				Name:     "Backend API",
				Type:     pidl.EntityTypeServer,
				Metadata: &pidl.EntityMetadata{Network: "internal"},
			},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "gateway", Action: "request"},
			{From: "gateway", To: "api", Action: "forward"},
		},
		Metadata: &pidl.ProtocolMetadata{
			Networks: map[string]*pidl.NetworkConfig{
				"external": {Name: "External Network", Style: "external"},
				"dmz":      {Name: "DMZ", Style: "dmz"},
				"internal": {Name: "Internal Network", Style: "trusted"},
			},
			NetworkLayout: &pidl.NetworkLayoutConfig{
				Direction: "horizontal",
				Order:     []string{"external", "dmz", "internal"},
			},
		},
	}

	renderer := render.NewSVGNetwork()
	svg, err := renderer.RenderString(protocol)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Verify boundaries
	if !strings.Contains(svg, "External Network") {
		t.Error("SVG should contain External Network boundary")
	}
	if !strings.Contains(svg, "DMZ") {
		t.Error("SVG should contain DMZ boundary")
	}
	if !strings.Contains(svg, "Internal Network") {
		t.Error("SVG should contain Internal Network boundary")
	}

	// Verify boundary styles
	if !strings.Contains(svg, "boundary-external") {
		t.Error("SVG should have external boundary style")
	}
	if !strings.Contains(svg, "boundary-dmz") {
		t.Error("SVG should have dmz boundary style")
	}
	if !strings.Contains(svg, "boundary-trusted") {
		t.Error("SVG should have trusted boundary style")
	}
}

func TestSVGNetworkRenderer_BoundaryOverrides(t *testing.T) {
	protocol := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []pidl.Entity{
			{ID: "a", Name: "A"},
			{ID: "b", Name: "B"},
			{ID: "c", Name: "C"},
		},
		Flows: []pidl.Flow{},
	}

	renderer := render.NewSVGNetwork()
	renderer.AddBoundaryOverride("zone1", []string{"a", "b"})
	renderer.AddBoundaryOverride("zone2", []string{"c"})

	svg, err := renderer.RenderString(protocol)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should have two boundaries from overrides
	if !strings.Contains(svg, "Zone1") || !strings.Contains(svg, "Zone2") {
		t.Error("SVG should contain override boundary names")
	}
}

func TestSVGNetworkRenderer_VerticalLayout(t *testing.T) {
	protocol := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []pidl.Entity{
			{ID: "a", Name: "A", Metadata: &pidl.EntityMetadata{Network: "zone1"}},
			{ID: "b", Name: "B", Metadata: &pidl.EntityMetadata{Network: "zone2"}},
		},
		Flows: []pidl.Flow{
			{From: "a", To: "b", Action: "call"},
		},
		Metadata: &pidl.ProtocolMetadata{
			NetworkLayout: &pidl.NetworkLayoutConfig{
				Direction: "vertical",
			},
		},
	}

	renderer := render.NewSVGNetwork()
	svg, err := renderer.RenderString(protocol)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Just verify it renders without error
	if !strings.Contains(svg, "<svg") {
		t.Error("Should produce valid SVG")
	}
}

func TestParseFormat_SVGNetwork(t *testing.T) {
	tests := []struct {
		input string
		want  render.Format
	}{
		{"svg-network", render.FormatSVGNetwork},
		{"svg-net", render.FormatSVGNetwork},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := render.ParseFormat(tt.input)
			if err != nil {
				t.Fatalf("ParseFormat(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNew_SVGNetwork(t *testing.T) {
	renderer, err := render.New(render.FormatSVGNetwork)
	if err != nil {
		t.Fatalf("New(FormatSVGNetwork) error = %v", err)
	}
	if renderer.Format() != render.FormatSVGNetwork {
		t.Errorf("Format() = %v, want %v", renderer.Format(), render.FormatSVGNetwork)
	}
}
