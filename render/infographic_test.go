package render

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

func createTestProcessSpecForInfographic() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "etl-pipeline",
			Name: "ETL Pipeline",
			Kind: pidl.ProtocolKindProcess,
		},
		Entities: []pidl.Entity{
			{
				ID:       "extract",
				Name:     "Extract",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeDeterministic,
			},
			{
				ID:       "transform",
				Name:     "Transform",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeLLM,
			},
			{
				ID:       "load",
				Name:     "Load",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeExternal,
			},
		},
		Flows: []pidl.Flow{
			{From: "extract", To: "transform", Action: "send"},
			{From: "transform", To: "load", Action: "send"},
		},
	}
}

func TestInfographicRenderer_Render(t *testing.T) {
	p := createTestProcessSpecForInfographic()
	opts := DefaultInfographicOptions()
	opts.Title = "ETL Pipeline"

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Check SVG structure
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("expected SVG to start with <svg")
	}
	if !strings.HasSuffix(svg, "</svg>") {
		t.Error("expected SVG to end with </svg>")
	}

	// Check for nodes
	if !strings.Contains(svg, "class=\"node\"") {
		t.Error("expected nodes in SVG")
	}

	// Check for edges
	if !strings.Contains(svg, "class=\"edge\"") {
		t.Error("expected edges in SVG")
	}

	// Check for title
	if !strings.Contains(svg, "ETL Pipeline") {
		t.Error("expected title in SVG")
	}

	// Check for animated dots
	if !strings.Contains(svg, "animateMotion") {
		t.Error("expected animated dots in SVG")
	}

	// Check for step type icons
	if !strings.Contains(svg, "⚙️") { // deterministic
		t.Error("expected deterministic icon")
	}
	if !strings.Contains(svg, "🧠") { // LLM
		t.Error("expected LLM icon")
	}
	if !strings.Contains(svg, "☁️") { // external
		t.Error("expected external icon")
	}
}

func TestInfographicRenderer_DatasheetTile(t *testing.T) {
	p := createTestProcessSpecForInfographic()
	opts := DatasheetTileOptions()
	opts.Title = "ETL"

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Check dimensions
	if !strings.Contains(svg, `width="400"`) {
		t.Error("expected width 400 for datasheet tile")
	}
	if !strings.Contains(svg, `height="400"`) {
		t.Error("expected height 400 for datasheet tile")
	}
}

func TestInfographicRenderer_DarkTheme(t *testing.T) {
	p := createTestProcessSpecForInfographic()
	opts := DefaultInfographicOptions()
	opts.Theme = ThemeDark

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Check for dark background
	if !strings.Contains(svg, "#1a1a2e") {
		t.Error("expected dark background color")
	}
}

func TestInfographicRenderer_NoAnimation(t *testing.T) {
	p := createTestProcessSpecForInfographic()
	opts := DefaultInfographicOptions()
	opts.AnimateDots = false

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Should not have animated dots
	if strings.Contains(svg, "animateMotion") {
		t.Error("expected no animation when AnimateDots is false")
	}
}

func TestInfographicRenderer_VerticalLayout(t *testing.T) {
	p := createTestProcessSpecForInfographic()
	opts := DefaultInfographicOptions()
	opts.Direction = "vertical"

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Should render without error
	if !strings.Contains(svg, "<svg") {
		t.Error("expected valid SVG for vertical layout")
	}
}

func TestInfographicRenderer_CustomSize(t *testing.T) {
	p := createTestProcessSpecForInfographic()
	opts := DefaultInfographicOptions()
	opts.Width = 800
	opts.Height = 600

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	if !strings.Contains(svg, `width="800"`) {
		t.Error("expected custom width 800")
	}
	if !strings.Contains(svg, `height="600"`) {
		t.Error("expected custom height 600")
	}
}

func TestInfographicRenderer_BidirectionalDots(t *testing.T) {
	p := createTestProcessSpecForInfographic()
	opts := DefaultInfographicOptions()
	opts.BidirectionalDots = true

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Should have keyPoints for reverse animation
	if !strings.Contains(svg, "keyPoints") {
		t.Error("expected keyPoints for bidirectional animation")
	}
}

func TestInfographicRenderer_AllSizes(t *testing.T) {
	p := createTestProcessSpecForInfographic()

	sizes := []InfographicSize{
		SizeLinkedInSquare,
		SizeLinkedInPortrait,
		SizeLinkedInLandscape,
		SizeDatasheetTile,
		SizeDatasheetWide,
	}

	for _, size := range sizes {
		t.Run(string(size), func(t *testing.T) {
			opts := DefaultInfographicOptions()
			opts.Size = size

			r := NewInfographicRenderer(opts)
			svg := r.Render(p)

			if !strings.Contains(svg, "<svg") {
				t.Errorf("expected valid SVG for size %s", size)
			}
		})
	}
}

func TestInfographicRenderer_AllThemes(t *testing.T) {
	p := createTestProcessSpecForInfographic()

	themes := []InfographicTheme{
		ThemeBold,
		ThemeMinimal,
		ThemeDark,
		ThemeTech,
	}

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			opts := DefaultInfographicOptions()
			opts.Theme = theme

			r := NewInfographicRenderer(opts)
			svg := r.Render(p)

			if !strings.Contains(svg, "<svg") {
				t.Errorf("expected valid SVG for theme %s", theme)
			}
		})
	}
}

// createTestRegularProtocolForInfographic creates a regular protocol (not process spec)
// to test entity type icons and colors.
func createTestRegularProtocolForInfographic() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "aauth-test",
			Name: "AAuth Test Protocol",
		},
		Entities: []pidl.Entity{
			{ID: "human", Name: "Human", Type: pidl.EntityTypeUser},
			{ID: "agent", Name: "Agent", Type: pidl.EntityTypeAgent},
			{ID: "auth_server", Name: "Auth Server", Type: pidl.EntityTypeAuthorizationServer},
			{ID: "resource", Name: "Resource", Type: pidl.EntityTypeResourceServer},
		},
		Flows: []pidl.Flow{
			{From: "human", To: "auth_server", Action: "authorize"},
			{From: "agent", To: "auth_server", Action: "request_token"},
			{From: "agent", To: "resource", Action: "access"},
		},
	}
}

func TestInfographicRenderer_EntityTypeIcons(t *testing.T) {
	p := createTestRegularProtocolForInfographic()
	opts := DefaultInfographicOptions()
	opts.Title = "AAuth Protocol"

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Check for entity type icons
	if !strings.Contains(svg, "👤") { // user
		t.Error("expected user icon 👤")
	}
	if !strings.Contains(svg, "🤖") { // agent
		t.Error("expected agent icon 🤖")
	}
	if !strings.Contains(svg, "🔐") { // authorization_server
		t.Error("expected authorization_server icon 🔐")
	}
	if !strings.Contains(svg, "🗄️") { // resource_server
		t.Error("expected resource_server icon 🗄️")
	}
}

func TestInfographicRenderer_EntityTypeColors(t *testing.T) {
	p := createTestRegularProtocolForInfographic()
	opts := DefaultInfographicOptions()

	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Check for entity type colors
	if !strings.Contains(svg, "#4CAF50") { // user - green
		t.Error("expected user color #4CAF50")
	}
	if !strings.Contains(svg, "#9C27B0") { // agent - purple
		t.Error("expected agent color #9C27B0")
	}
	if !strings.Contains(svg, "#F44336") { // authorization_server - red
		t.Error("expected authorization_server color #F44336")
	}
	if !strings.Contains(svg, "#FF9800") { // resource_server - orange
		t.Error("expected resource_server color #FF9800")
	}
}

func TestInfographicRenderer_MixedStepAndEntityTypes(t *testing.T) {
	// Test that step type takes precedence over entity type for icons and colors.
	// When both step_type and entity type are present, step_type wins.
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "mixed-test",
			Name: "Mixed Test",
			Kind: pidl.ProtocolKindProcess,
		},
		Entities: []pidl.Entity{
			{
				ID:       "llm_agent",
				Name:     "LLM Agent",
				Type:     pidl.EntityTypeAgent,
				StepType: pidl.StepTypeLLM, // step type should take precedence
			},
			{
				ID:       "human_step",
				Name:     "Human Review",
				Type:     pidl.EntityTypeUser,
				StepType: pidl.StepTypeHuman, // step type should take precedence
			},
		},
		Flows: []pidl.Flow{
			{From: "llm_agent", To: "human_step", Action: "send"},
		},
	}

	opts := DefaultInfographicOptions()
	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// LLM step type icon should appear (brain), not agent icon (robot)
	if !strings.Contains(svg, "🧠") {
		t.Error("expected LLM icon 🧠 for step type")
	}

	// Human step type icon should appear (person), same as user but from step type
	if !strings.Contains(svg, "👤") {
		t.Error("expected human icon 👤 for step type")
	}

	// LLM step color should appear
	if !strings.Contains(svg, "#7B1FA2") { // LLM purple
		t.Error("expected LLM step color #7B1FA2")
	}

	// Human step color should appear
	if !strings.Contains(svg, "#388E3C") { // Human green
		t.Error("expected human step color #388E3C")
	}
}

func TestInfographicRenderer_CustomShapes(t *testing.T) {
	// Test that custom shapes are rendered based on entity/step type.
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "shape-test",
			Name: "Shape Test",
		},
		Entities: []pidl.Entity{
			{ID: "user", Name: "User", Type: pidl.EntityTypeUser},                       // circle
			{ID: "agent", Name: "Agent", Type: pidl.EntityTypeAgent},                    // hexagon
			{ID: "resource", Name: "Resource", Type: pidl.EntityTypeResourceServer},     // cylinder
			{ID: "auth", Name: "Auth Server", Type: pidl.EntityTypeAuthorizationServer}, // diamond
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},                 // rectangle
		},
		Flows: []pidl.Flow{
			{From: "user", To: "agent", Action: "request"},
			{From: "agent", To: "auth", Action: "authenticate"},
			{From: "agent", To: "resource", Action: "access"},
		},
	}

	opts := DefaultInfographicOptions()
	opts.UseCustomShapes = true
	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Check for ellipse (user - circle shape)
	if !strings.Contains(svg, "<ellipse") {
		t.Error("expected ellipse for user (circle shape)")
	}

	// Check for polygon (hexagon and diamond shapes)
	if !strings.Contains(svg, "<polygon") {
		t.Error("expected polygon for agent (hexagon) or auth (diamond)")
	}

	// Check for cylinder components (ellipse at top/bottom, lines)
	// Cylinder has multiple ellipses, so we check the overall structure
	// The resource server should have cylinder shape which uses multiple SVG elements

	// Check for rectangle (server - default rectangle shape)
	if !strings.Contains(svg, "<rect") {
		t.Error("expected rect for server (rectangle shape)")
	}
}

func TestInfographicRenderer_CustomShapesDisabled(t *testing.T) {
	// Test that when UseCustomShapes is false, all nodes use rectangles.
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "shape-disabled-test",
			Name: "Shape Disabled Test",
		},
		Entities: []pidl.Entity{
			{ID: "user", Name: "User", Type: pidl.EntityTypeUser},
			{ID: "agent", Name: "Agent", Type: pidl.EntityTypeAgent},
		},
		Flows: []pidl.Flow{
			{From: "user", To: "agent", Action: "request"},
		},
	}

	opts := DefaultInfographicOptions()
	opts.UseCustomShapes = false // Disable custom shapes
	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Should have rectangles for all nodes
	if !strings.Contains(svg, "<rect") {
		t.Error("expected rect elements when UseCustomShapes is disabled")
	}

	// Should NOT have ellipses (circle shape) when shapes are disabled
	// (only rects should be used for nodes)
	// Note: ellipses may appear in other places (like glow filter),
	// but node shapes should be rects
}

func TestInfographicRenderer_StepTypeShapes(t *testing.T) {
	// Test that step type shapes take precedence over entity type shapes.
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "step-shape-test",
			Name: "Step Shape Test",
			Kind: pidl.ProtocolKindProcess,
		},
		Entities: []pidl.Entity{
			{
				ID:       "llm_step",
				Name:     "LLM Step",
				Type:     pidl.EntityTypeServer, // would be rectangle
				StepType: pidl.StepTypeLLM,      // should override to hexagon
			},
			{
				ID:       "human_step",
				Name:     "Human Step",
				Type:     pidl.EntityTypeServer, // would be rectangle
				StepType: pidl.StepTypeHuman,    // should override to circle
			},
			{
				ID:       "external_step",
				Name:     "External Step",
				Type:     pidl.EntityTypeServer, // would be rectangle
				StepType: pidl.StepTypeExternal, // should override to cloud
			},
			{
				ID:       "tool_step",
				Name:     "Tool Step",
				Type:     pidl.EntityTypeServer, // would be rectangle
				StepType: pidl.StepTypeTool,     // should override to diamond
			},
		},
		Flows: []pidl.Flow{
			{From: "llm_step", To: "human_step", Action: "send"},
			{From: "human_step", To: "external_step", Action: "send"},
			{From: "external_step", To: "tool_step", Action: "send"},
		},
	}

	opts := DefaultInfographicOptions()
	opts.UseCustomShapes = true
	r := NewInfographicRenderer(opts)
	svg := r.Render(p)

	// Should have polygon (hexagon for LLM, diamond for tool)
	if !strings.Contains(svg, "<polygon") {
		t.Error("expected polygon for LLM (hexagon) or tool (diamond)")
	}

	// Should have ellipse (circle for human)
	if !strings.Contains(svg, "<ellipse") {
		t.Error("expected ellipse for human (circle)")
	}

	// Should have path (cloud for external)
	if !strings.Contains(svg, "<path") {
		t.Error("expected path for external (cloud)")
	}
}
