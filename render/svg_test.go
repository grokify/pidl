package render_test

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/render"
)

func TestSVGRenderer_Format(t *testing.T) {
	tests := []struct {
		name     string
		renderer *render.SVGRenderer
		want     render.Format
	}{
		{
			name:     "static SVG",
			renderer: render.NewSVG(),
			want:     render.FormatSVG,
		},
		{
			name:     "animated SVG",
			renderer: render.NewSVGAnimated(),
			want:     render.FormatSVGAnimated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.renderer.Format(); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSVGRenderer_RenderString(t *testing.T) {
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
			{From: "server", To: "client", Action: "response", Label: "Return Response", Mode: pidl.FlowModeResponse},
		},
	}

	renderer := render.NewSVG()
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
		{"lifelines group", `class="lifelines"`},
		{"participants group", `class="participants"`},
		{"messages group", `class="messages"`},
		{"client participant", "Client"},
		{"server participant", "Server"},
		{"request message", "Send Request"},
		{"response message", "Return Response"},
		{"participant box", `class="participant-box"`},
		{"lifeline", `class="lifeline"`},
		{"message line", `class="message-line"`},
		{"response dashed", `class="message-line-response"`},
		{"CSS variables", "--color-primary"},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if !strings.Contains(svg, check.contains) {
				t.Errorf("SVG should contain %q", check.contains)
			}
		})
	}
}

func TestSVGRenderer_Animated(t *testing.T) {
	protocol := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test-protocol",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{ID: "a", Name: "A"},
			{ID: "b", Name: "B"},
		},
		Flows: []pidl.Flow{
			{From: "a", To: "b", Action: "msg", Label: "Message"},
		},
	}

	t.Run("static has no animation elements", func(t *testing.T) {
		renderer := render.NewSVG()
		svg, _ := renderer.RenderString(protocol)
		// Static SVG should not have animated circle elements
		if strings.Contains(svg, `class="flow-dot flow-dot-`) {
			t.Error("Static SVG should not contain animated flow-dot elements")
		}
		if strings.Contains(svg, "@keyframes flow") {
			t.Error("Static SVG should not contain animation keyframes")
		}
	})

	t.Run("animated has animation elements", func(t *testing.T) {
		renderer := render.NewSVGAnimated()
		svg, _ := renderer.RenderString(protocol)
		// Animated SVG should have animated circle elements
		if !strings.Contains(svg, `class="flow-dot flow-dot-`) {
			t.Error("Animated SVG should contain animated flow-dot elements")
		}
		if !strings.Contains(svg, "@keyframes flow") {
			t.Error("Animated SVG should contain animation keyframes")
		}
		if !strings.Contains(svg, "offset-path") {
			t.Error("Animated SVG should use CSS offset-path")
		}
	})
}

func TestSVGRenderer_Theme(t *testing.T) {
	protocol := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "test", Name: "Test"},
		Entities:     []pidl.Entity{{ID: "a", Name: "A"}},
		Flows:        []pidl.Flow{},
	}

	tests := []struct {
		name      string
		theme     string
		wantClass string
	}{
		{"light theme", "light", "theme-light"},
		{"dark theme", "dark", "theme-dark"},
		{"auto theme", "auto", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := render.NewSVG()
			renderer.Theme = tt.theme
			svg, _ := renderer.RenderString(protocol)

			if tt.wantClass != "" {
				if !strings.Contains(svg, tt.wantClass) {
					t.Errorf("SVG should contain class %q for theme %q", tt.wantClass, tt.theme)
				}
			}
		})
	}
}

func TestSVGRenderer_EmptyProtocol(t *testing.T) {
	protocol := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "empty", Name: "Empty"},
		Entities:     []pidl.Entity{},
		Flows:        []pidl.Flow{},
	}

	renderer := render.NewSVG()
	svg, err := renderer.RenderString(protocol)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if !strings.Contains(svg, "<svg") {
		t.Error("Should produce valid SVG even with empty protocol")
	}
}

func TestParseFormat_SVG(t *testing.T) {
	tests := []struct {
		input string
		want  render.Format
	}{
		{"svg", render.FormatSVG},
		{"SVG", render.FormatSVG},
		{"svg-animated", render.FormatSVGAnimated},
		{"svg-anim", render.FormatSVGAnimated},
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

func TestSVGFormat_FileExtension(t *testing.T) {
	if ext := render.FormatSVG.FileExtension(); ext != ".svg" {
		t.Errorf("FormatSVG.FileExtension() = %q, want .svg", ext)
	}
	if ext := render.FormatSVGAnimated.FileExtension(); ext != ".svg" {
		t.Errorf("FormatSVGAnimated.FileExtension() = %q, want .svg", ext)
	}
}

func TestNew_SVG(t *testing.T) {
	renderer, err := render.New(render.FormatSVG)
	if err != nil {
		t.Fatalf("New(FormatSVG) error = %v", err)
	}
	if renderer.Format() != render.FormatSVG {
		t.Errorf("New(FormatSVG).Format() = %v, want %v", renderer.Format(), render.FormatSVG)
	}

	renderer, err = render.New(render.FormatSVGAnimated)
	if err != nil {
		t.Fatalf("New(FormatSVGAnimated) error = %v", err)
	}
	if renderer.Format() != render.FormatSVGAnimated {
		t.Errorf("New(FormatSVGAnimated).Format() = %v, want %v", renderer.Format(), render.FormatSVGAnimated)
	}
}
