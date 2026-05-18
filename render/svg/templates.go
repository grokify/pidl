package svg

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*
var embeddedTemplates embed.FS

// TemplateConfig contains configuration for an SVG template.
type TemplateConfig struct {
	// Name is the template identifier.
	Name string `json:"name"`
	// Description explains what the template is for.
	Description string `json:"description"`
	// Layout contains positioning parameters.
	Layout TemplateLayout `json:"layout"`
	// Fonts specifies font families.
	Fonts TemplateFonts `json:"fonts,omitempty"`
	// LineStyle configures line rendering.
	LineStyle TemplateLineStyle `json:"line_style,omitempty"`
	// Colors overrides default color scheme.
	Colors TemplateColors `json:"colors,omitempty"`
	// Animation configures animation defaults.
	Animation TemplateAnimation `json:"animation,omitempty"`
}

// TemplateLayout contains positioning parameters.
type TemplateLayout struct {
	// Padding around the entire diagram.
	Padding int `json:"padding,omitempty"`
	// ParticipantSpacing is horizontal spacing between participants.
	ParticipantSpacing int `json:"participant_spacing,omitempty"`
	// MessageSpacing is vertical spacing between messages.
	MessageSpacing int `json:"message_spacing,omitempty"`
	// ParticipantBox configures participant box dimensions.
	ParticipantBox TemplateParticipantBox `json:"participant_box,omitempty"`
}

// TemplateParticipantBox configures participant box dimensions.
type TemplateParticipantBox struct {
	// Width of participant boxes.
	Width int `json:"width,omitempty"`
	// Height of participant boxes.
	Height int `json:"height,omitempty"`
	// CornerRadius for rounded corners. Use -1 to indicate "not set".
	CornerRadius *int `json:"corner_radius,omitempty"`
}

// TemplateFonts specifies font families.
type TemplateFonts struct {
	// Primary font for text.
	Primary string `json:"primary,omitempty"`
	// Mono font for code/technical text.
	Mono string `json:"mono,omitempty"`
}

// TemplateLineStyle configures line rendering.
type TemplateLineStyle struct {
	// StrokeWidth for lines.
	StrokeWidth float64 `json:"stroke_width,omitempty"`
	// HandDrawn enables hand-drawn effect.
	HandDrawn bool `json:"hand_drawn,omitempty"`
	// WobbleFactor controls hand-drawn wobble amount.
	WobbleFactor float64 `json:"wobble_factor,omitempty"`
}

// TemplateColors overrides default color scheme.
type TemplateColors struct {
	// Primary color.
	Primary string `json:"primary,omitempty"`
	// Secondary color.
	Secondary string `json:"secondary,omitempty"`
	// Accent color.
	Accent string `json:"accent,omitempty"`
	// Background color.
	Background string `json:"background,omitempty"`
	// Text color.
	Text string `json:"text,omitempty"`
	// Line color.
	Line string `json:"line,omitempty"`
	// Lifeline color.
	Lifeline string `json:"lifeline,omitempty"`
	// ParticipantBg is participant box background.
	ParticipantBg string `json:"participant_bg,omitempty"`
	// ParticipantText is participant text color.
	ParticipantText string `json:"participant_text,omitempty"`
}

// TemplateAnimation configures animation defaults.
type TemplateAnimation struct {
	// DotShape for animated dots (circle, square, diamond).
	DotShape string `json:"dot_shape,omitempty"`
	// DotSize is the default dot radius.
	DotSize int `json:"dot_size,omitempty"`
	// TrailEnabled shows a fading trail.
	TrailEnabled bool `json:"trail_enabled,omitempty"`
}

// Template represents a loaded SVG template.
type Template struct {
	// Config contains template configuration.
	Config TemplateConfig
	// CSS contains custom CSS styles.
	CSS string
	// source is where the template was loaded from.
	source string
}

// BuiltinTemplates lists available built-in template names.
var BuiltinTemplates = []string{"default", "minimal", "sketch", "blueprint", "dark"}

// LoadTemplate loads a template by name from embedded templates.
func LoadTemplate(name string) (*Template, error) {
	if name == "" {
		name = "default"
	}

	// Try embedded templates
	templatePath := fmt.Sprintf("templates/%s", name)

	// Check if template exists
	_, err := embeddedTemplates.ReadDir(templatePath)
	if err != nil {
		return nil, fmt.Errorf("template %q not found: %w", name, err)
	}

	return loadTemplateFromFS(embeddedTemplates, templatePath, name)
}

// LoadTemplateFromDir loads a template from a filesystem directory.
func LoadTemplateFromDir(dir string) (*Template, error) {
	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("template directory %q not found: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", dir)
	}

	return loadTemplateFromPath(dir)
}

func loadTemplateFromFS(fsys fs.FS, basePath, name string) (*Template, error) {
	tmpl := &Template{
		source: "embedded:" + name,
	}

	// Load config.json
	configPath := basePath + "/config.json"
	configData, err := fs.ReadFile(fsys, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.json: %w", err)
	}

	if err := json.Unmarshal(configData, &tmpl.Config); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	// Load styles.css (optional)
	cssPath := basePath + "/styles.css"
	cssData, err := fs.ReadFile(fsys, cssPath)
	if err == nil {
		tmpl.CSS = string(cssData)
	}

	return tmpl, nil
}

func loadTemplateFromPath(dir string) (*Template, error) {
	tmpl := &Template{
		source: "dir:" + dir,
	}

	// Load config.json
	configPath := filepath.Join(dir, "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.json: %w", err)
	}

	if err := json.Unmarshal(configData, &tmpl.Config); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	// Load styles.css (optional)
	cssPath := filepath.Join(dir, "styles.css")
	cssData, err := os.ReadFile(cssPath)
	if err == nil {
		tmpl.CSS = string(cssData)
	}

	return tmpl, nil
}

// ApplyToLayoutConfig applies template layout settings to a LayoutConfig.
func (t *Template) ApplyToLayoutConfig(config *LayoutConfig) {
	layout := t.Config.Layout

	if layout.Padding > 0 {
		config.Padding = layout.Padding
	}
	if layout.ParticipantSpacing > 0 {
		config.ParticipantSpacing = layout.ParticipantSpacing
	}
	if layout.MessageSpacing > 0 {
		config.MessageSpacing = layout.MessageSpacing
	}
	if layout.ParticipantBox.Width > 0 {
		config.ParticipantBoxWidth = layout.ParticipantBox.Width
	}
	if layout.ParticipantBox.Height > 0 {
		config.ParticipantBoxHeight = layout.ParticipantBox.Height
	}
}

// GenerateCSS generates CSS with template overrides.
func (t *Template) GenerateCSS(baseTheme Theme) string {
	var sb strings.Builder

	// Start with base theme CSS
	sb.WriteString(GenerateCSS(baseTheme))

	// Apply color overrides
	if t.hasColorOverrides() {
		sb.WriteString("\n    /* Template color overrides */\n")
		sb.WriteString("    :root {\n")
		colors := t.Config.Colors
		if colors.Primary != "" {
			sb.WriteString(fmt.Sprintf("      --color-primary: %s;\n", colors.Primary))
		}
		if colors.Secondary != "" {
			sb.WriteString(fmt.Sprintf("      --color-secondary: %s;\n", colors.Secondary))
		}
		if colors.Accent != "" {
			sb.WriteString(fmt.Sprintf("      --color-accent: %s;\n", colors.Accent))
		}
		if colors.Background != "" {
			sb.WriteString(fmt.Sprintf("      --color-bg: %s;\n", colors.Background))
		}
		if colors.Text != "" {
			sb.WriteString(fmt.Sprintf("      --color-text: %s;\n", colors.Text))
		}
		if colors.Line != "" {
			sb.WriteString(fmt.Sprintf("      --color-line: %s;\n", colors.Line))
		}
		if colors.Lifeline != "" {
			sb.WriteString(fmt.Sprintf("      --color-lifeline: %s;\n", colors.Lifeline))
		}
		if colors.ParticipantBg != "" {
			sb.WriteString(fmt.Sprintf("      --color-participant-bg: %s;\n", colors.ParticipantBg))
		}
		if colors.ParticipantText != "" {
			sb.WriteString(fmt.Sprintf("      --color-participant-text: %s;\n", colors.ParticipantText))
		}
		sb.WriteString("    }\n")
	}

	// Apply font overrides
	if t.Config.Fonts.Primary != "" {
		sb.WriteString(fmt.Sprintf("\n    .participant-text, .message-text { font-family: %s; }\n", t.Config.Fonts.Primary))
	}

	// Apply line style overrides
	if t.Config.LineStyle.StrokeWidth > 0 {
		sb.WriteString(fmt.Sprintf("\n    .message-line, .message-line-response { stroke-width: %.1f; }\n", t.Config.LineStyle.StrokeWidth))
	}

	// Append custom CSS from template
	if t.CSS != "" {
		sb.WriteString("\n    /* Template custom CSS */\n")
		sb.WriteString(t.CSS)
	}

	return sb.String()
}

func (t *Template) hasColorOverrides() bool {
	c := t.Config.Colors
	return c.Primary != "" || c.Secondary != "" || c.Accent != "" ||
		c.Background != "" || c.Text != "" || c.Line != "" ||
		c.Lifeline != "" || c.ParticipantBg != "" || c.ParticipantText != ""
}

// GetCornerRadius returns the corner radius for participant boxes.
func (t *Template) GetCornerRadius() int {
	if t.Config.Layout.ParticipantBox.CornerRadius != nil {
		return *t.Config.Layout.ParticipantBox.CornerRadius
	}
	return 4 // default
}
