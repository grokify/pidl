package svg_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/pidl/render/svg"
)

func TestLoadTemplate_Default(t *testing.T) {
	tmpl, err := svg.LoadTemplate("default")
	if err != nil {
		t.Fatalf("LoadTemplate(default) error = %v", err)
	}

	if tmpl.Config.Name != "default" {
		t.Errorf("Name = %q, want %q", tmpl.Config.Name, "default")
	}
	if tmpl.Config.Layout.Padding == 0 {
		t.Error("Padding should not be zero")
	}
	if tmpl.Config.Layout.ParticipantSpacing == 0 {
		t.Error("ParticipantSpacing should not be zero")
	}
}

func TestLoadTemplate_AllBuiltins(t *testing.T) {
	for _, name := range svg.BuiltinTemplates {
		t.Run(name, func(t *testing.T) {
			tmpl, err := svg.LoadTemplate(name)
			if err != nil {
				t.Fatalf("LoadTemplate(%q) error = %v", name, err)
			}
			if tmpl.Config.Name != name {
				t.Errorf("Name = %q, want %q", tmpl.Config.Name, name)
			}
		})
	}
}

func TestLoadTemplate_Empty(t *testing.T) {
	tmpl, err := svg.LoadTemplate("")
	if err != nil {
		t.Fatalf("LoadTemplate('') should return default, got error = %v", err)
	}
	if tmpl.Config.Name != "default" {
		t.Errorf("Empty string should load default template, got %q", tmpl.Config.Name)
	}
}

func TestLoadTemplate_NotFound(t *testing.T) {
	_, err := svg.LoadTemplate("nonexistent")
	if err == nil {
		t.Error("LoadTemplate(nonexistent) should error")
	}
}

func TestLoadTemplateFromDir(t *testing.T) {
	// Create a temporary directory with a template
	dir := t.TempDir()

	configJSON := `{
		"name": "custom",
		"description": "Custom test template",
		"layout": {
			"padding": 30,
			"participant_spacing": 200
		}
	}`

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(configJSON), 0600); err != nil {
		t.Fatal(err)
	}

	tmpl, err := svg.LoadTemplateFromDir(dir)
	if err != nil {
		t.Fatalf("LoadTemplateFromDir() error = %v", err)
	}

	if tmpl.Config.Name != "custom" {
		t.Errorf("Name = %q, want %q", tmpl.Config.Name, "custom")
	}
	if tmpl.Config.Layout.Padding != 30 {
		t.Errorf("Padding = %d, want 30", tmpl.Config.Layout.Padding)
	}
	if tmpl.Config.Layout.ParticipantSpacing != 200 {
		t.Errorf("ParticipantSpacing = %d, want 200", tmpl.Config.Layout.ParticipantSpacing)
	}
}

func TestLoadTemplateFromDir_WithCSS(t *testing.T) {
	dir := t.TempDir()

	configJSON := `{"name": "with-css"}`
	customCSS := ".custom-class { color: red; }"

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(configJSON), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "styles.css"), []byte(customCSS), 0600); err != nil {
		t.Fatal(err)
	}

	tmpl, err := svg.LoadTemplateFromDir(dir)
	if err != nil {
		t.Fatalf("LoadTemplateFromDir() error = %v", err)
	}

	if tmpl.CSS != customCSS {
		t.Errorf("CSS = %q, want %q", tmpl.CSS, customCSS)
	}
}

func TestLoadTemplateFromDir_NotFound(t *testing.T) {
	_, err := svg.LoadTemplateFromDir("/nonexistent/path")
	if err == nil {
		t.Error("LoadTemplateFromDir should error for nonexistent path")
	}
}

func TestTemplate_ApplyToLayoutConfig(t *testing.T) {
	tmpl := &svg.Template{
		Config: svg.TemplateConfig{
			Layout: svg.TemplateLayout{
				Padding:            50,
				ParticipantSpacing: 180,
				MessageSpacing:     60,
				ParticipantBox: svg.TemplateParticipantBox{
					Width:  120,
					Height: 44,
				},
			},
		},
	}

	config := svg.DefaultLayoutConfig()
	tmpl.ApplyToLayoutConfig(&config)

	if config.Padding != 50 {
		t.Errorf("Padding = %d, want 50", config.Padding)
	}
	if config.ParticipantSpacing != 180 {
		t.Errorf("ParticipantSpacing = %d, want 180", config.ParticipantSpacing)
	}
	if config.MessageSpacing != 60 {
		t.Errorf("MessageSpacing = %d, want 60", config.MessageSpacing)
	}
	if config.ParticipantBoxWidth != 120 {
		t.Errorf("ParticipantBoxWidth = %d, want 120", config.ParticipantBoxWidth)
	}
	if config.ParticipantBoxHeight != 44 {
		t.Errorf("ParticipantBoxHeight = %d, want 44", config.ParticipantBoxHeight)
	}
}

func TestTemplate_GenerateCSS(t *testing.T) {
	tmpl := &svg.Template{
		Config: svg.TemplateConfig{
			Colors: svg.TemplateColors{
				Primary: "#ff0000",
				Text:    "#00ff00",
			},
			Fonts: svg.TemplateFonts{
				Primary: "Comic Sans, cursive",
			},
			LineStyle: svg.TemplateLineStyle{
				StrokeWidth: 3.0,
			},
		},
		CSS: ".custom { display: block; }",
	}

	css := tmpl.GenerateCSS(svg.ThemeLight)

	// Check base theme is included
	if !strings.Contains(css, "--color-primary:") {
		t.Error("CSS should contain base theme")
	}

	// Check color overrides
	if !strings.Contains(css, "--color-primary: #ff0000") {
		t.Error("CSS should contain primary color override")
	}
	if !strings.Contains(css, "--color-text: #00ff00") {
		t.Error("CSS should contain text color override")
	}

	// Check font override
	if !strings.Contains(css, "Comic Sans, cursive") {
		t.Error("CSS should contain font override")
	}

	// Check line style override
	if !strings.Contains(css, "stroke-width: 3.0") {
		t.Error("CSS should contain stroke width override")
	}

	// Check custom CSS
	if !strings.Contains(css, ".custom { display: block; }") {
		t.Error("CSS should contain custom CSS")
	}
}

func TestTemplate_GetCornerRadius(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name   string
		radius *int
		want   int
	}{
		{"nil uses default", nil, 4},
		{"explicit zero", intPtr(0), 0},
		{"custom value", intPtr(8), 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := &svg.Template{
				Config: svg.TemplateConfig{
					Layout: svg.TemplateLayout{
						ParticipantBox: svg.TemplateParticipantBox{
							CornerRadius: tt.radius,
						},
					},
				},
			}

			if got := tmpl.GetCornerRadius(); got != tt.want {
				t.Errorf("GetCornerRadius() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBuiltinTemplates_HaveCorrectStructure(t *testing.T) {
	for _, name := range svg.BuiltinTemplates {
		t.Run(name, func(t *testing.T) {
			tmpl, err := svg.LoadTemplate(name)
			if err != nil {
				t.Fatalf("LoadTemplate(%q) error = %v", name, err)
			}

			// All templates should have a name matching
			if tmpl.Config.Name != name {
				t.Errorf("Name = %q, want %q", tmpl.Config.Name, name)
			}

			// All templates should have a description
			if tmpl.Config.Description == "" {
				t.Error("Description should not be empty")
			}

			// Layout values should be reasonable
			if tmpl.Config.Layout.ParticipantSpacing > 0 && tmpl.Config.Layout.ParticipantSpacing < 50 {
				t.Errorf("ParticipantSpacing = %d seems too small", tmpl.Config.Layout.ParticipantSpacing)
			}
		})
	}
}
