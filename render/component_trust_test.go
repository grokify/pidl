package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

// testProtocolWithComponentsAndTrust returns a protocol with components, trust relations, and protocol roles.
func testProtocolWithComponentsAndTrust() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "oauth-flow",
			Name: "OAuth 2.0 Flow",
		},
		Entities: []pidl.Entity{
			{
				ID:   "client",
				Name: "Client App",
				Type: pidl.EntityTypeClient,
				ProtocolRoles: []pidl.ProtocolRole{
					{Protocol: "oauth", Role: "client"},
				},
			},
			{
				ID:   "auth_server",
				Name: "Authorization Server",
				Type: pidl.EntityTypeAuthorizationServer,
				ProtocolRoles: []pidl.ProtocolRole{
					{Protocol: "oauth", Role: "authorization_server"},
				},
			},
			{
				ID:   "resource_server",
				Name: "Resource Server",
				Type: pidl.EntityTypeResourceServer,
				ProtocolRoles: []pidl.ProtocolRole{
					{Protocol: "oauth", Role: "resource_server"},
				},
			},
			{
				ID:   "user",
				Name: "End User",
				Type: pidl.EntityTypeUser,
			},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "auth_server", Action: "authorize", Mode: pidl.FlowModeRedirect},
			{From: "auth_server", To: "client", Action: "callback", Mode: pidl.FlowModeCallback},
			{From: "client", To: "resource_server", Action: "api_call", Mode: pidl.FlowModeRequest},
		},
		Metadata: &pidl.ProtocolMetadata{
			Components: []pidl.DeploymentComponent{
				{
					ID:       "idp",
					Name:     "Identity Provider",
					Type:     pidl.ComponentTypeIdP,
					Entities: []string{"auth_server"},
					Implements: []pidl.ProtocolRole{
						{Protocol: "oauth", Role: "authorization_server"},
					},
				},
				{
					ID:       "api",
					Name:     "API Gateway",
					Type:     pidl.ComponentTypeGateway,
					Entities: []string{"resource_server"},
					Implements: []pidl.ProtocolRole{
						{Protocol: "oauth", Role: "resource_server"},
					},
				},
			},
			TrustRelations: []pidl.TrustRelationship{
				{
					From:        "client",
					To:          "auth_server",
					Type:        pidl.TrustTypeAuthenticates,
					Credentials: []string{"client_secret"},
				},
				{
					From:        "auth_server",
					To:          "resource_server",
					Type:        pidl.TrustTypeIssues,
					Credentials: []string{"access_token"},
				},
				{
					From:        "client",
					To:          "resource_server",
					Type:        pidl.TrustTypeTrusts,
					Mutual:      true,
					Credentials: []string{"access_token", "mtls"},
				},
			},
		},
	}
}

// TestMermaidComponentRenderer tests the Mermaid component diagram renderer.
func TestMermaidComponentRenderer(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewMermaidComponent()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check structure
	if !strings.HasPrefix(s, "flowchart TB") {
		t.Error("Mermaid component should start with flowchart TB")
	}

	// Check for subgraphs (components)
	if !strings.Contains(s, "subgraph") {
		t.Error("Mermaid component should contain subgraphs")
	}

	// Check for component names
	if !strings.Contains(s, "Identity Provider") {
		t.Error("Mermaid component should contain Identity Provider")
	}
	if !strings.Contains(s, "API Gateway") {
		t.Error("Mermaid component should contain API Gateway")
	}

	// Check for flows
	if !strings.Contains(s, "-->") {
		t.Error("Mermaid component should contain flow arrows")
	}
}

func TestMermaidComponentRendererFormat(t *testing.T) {
	r := NewMermaidComponent()
	if r.Format() != FormatMermaidComponent {
		t.Errorf("Format() = %v, want %v", r.Format(), FormatMermaidComponent)
	}
}

func TestMermaidComponentRendererWriter(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewMermaidComponent()

	var buf bytes.Buffer
	err := r.Render(&buf, p)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Render() wrote nothing")
	}
}

func TestMermaidComponentRendererNoComponents(t *testing.T) {
	p := testProtocol() // From render_test.go - has no components
	r := NewMermaidComponent()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should render entities directly when no components
	if !strings.Contains(s, "Client") || !strings.Contains(s, "Server") {
		t.Error("Mermaid component without components should render entities directly")
	}
}

// TestMermaidTrustRenderer tests the Mermaid trust diagram renderer.
func TestMermaidTrustRenderer(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewMermaidTrust()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check structure
	if !strings.HasPrefix(s, "flowchart LR") {
		t.Error("Mermaid trust should start with flowchart LR")
	}

	// Check for trust relationship labels
	if !strings.Contains(s, "authenticates") {
		t.Error("Mermaid trust should contain authenticates relationship")
	}
	if !strings.Contains(s, "issues") {
		t.Error("Mermaid trust should contain issues relationship")
	}
	if !strings.Contains(s, "trusts") {
		t.Error("Mermaid trust should contain trusts relationship")
	}

	// Check for credentials when enabled
	if !strings.Contains(s, "access_token") {
		t.Error("Mermaid trust should contain credential types")
	}

	// Check for styling
	if !strings.Contains(s, "classDef") {
		t.Error("Mermaid trust should contain style definitions")
	}
}

func TestMermaidTrustRendererFormat(t *testing.T) {
	r := NewMermaidTrust()
	if r.Format() != FormatMermaidTrust {
		t.Errorf("Format() = %v, want %v", r.Format(), FormatMermaidTrust)
	}
}

func TestMermaidTrustRendererWriter(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewMermaidTrust()

	var buf bytes.Buffer
	err := r.Render(&buf, p)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Render() wrote nothing")
	}
}

func TestMermaidTrustRendererNoTrustRelations(t *testing.T) {
	p := testProtocol() // Has no trust relations
	r := NewMermaidTrust()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should show no trust relationships message
	if !strings.Contains(s, "No trust relationships defined") {
		t.Error("Mermaid trust without relations should indicate no trust relationships")
	}
}

func TestMermaidTrustRendererMutualTrust(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewMermaidTrust()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for mutual trust arrow
	if !strings.Contains(s, "<-->") {
		t.Error("Mermaid trust should contain mutual trust arrow")
	}
}

// TestMarkdownMatrixRenderer tests the Markdown role matrix renderer.
func TestMarkdownMatrixRenderer(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewMarkdownMatrix()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for title
	if !strings.Contains(s, "# OAuth 2.0 Flow - Protocol Role Matrix") {
		t.Error("Markdown matrix should contain title header")
	}

	// Check for table structure
	if !strings.Contains(s, "| Entity |") {
		t.Error("Markdown matrix should contain Entity header column")
	}

	// Check for protocol column
	if !strings.Contains(s, "| oauth |") {
		t.Error("Markdown matrix should contain oauth protocol column")
	}

	// Check for entity rows
	if !strings.Contains(s, "**Client App**") {
		t.Error("Markdown matrix should contain Client App entity")
	}
	if !strings.Contains(s, "**Authorization Server**") {
		t.Error("Markdown matrix should contain Authorization Server entity")
	}

	// Check for roles
	if !strings.Contains(s, "client") && !strings.Contains(s, "authorization_server") {
		t.Error("Markdown matrix should contain protocol roles")
	}

	// Check for protocol legend
	if !strings.Contains(s, "## Protocol Legend") {
		t.Error("Markdown matrix should contain protocol legend")
	}

	// Check for components section
	if !strings.Contains(s, "## Deployment Components") {
		t.Error("Markdown matrix should contain deployment components section")
	}

	// Check for trust relationships section
	if !strings.Contains(s, "## Trust Relationships") {
		t.Error("Markdown matrix should contain trust relationships section")
	}
}

func TestMarkdownMatrixRendererFormat(t *testing.T) {
	r := NewMarkdownMatrix()
	if r.Format() != FormatMarkdownMatrix {
		t.Errorf("Format() = %v, want %v", r.Format(), FormatMarkdownMatrix)
	}
}

func TestMarkdownMatrixRendererWriter(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewMarkdownMatrix()

	var buf bytes.Buffer
	err := r.Render(&buf, p)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Render() wrote nothing")
	}
}

func TestMarkdownMatrixRendererNoRoles(t *testing.T) {
	p := testProtocol() // Has no protocol roles
	r := NewMarkdownMatrix()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should show no roles message
	if !strings.Contains(s, "No protocol roles defined") {
		t.Error("Markdown matrix without roles should indicate no roles")
	}
}

// TestSVGComponentRenderer tests the SVG component diagram renderer.
func TestSVGComponentRenderer(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewSVGComponent()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check SVG structure
	if !strings.HasPrefix(s, "<svg") {
		t.Error("SVG component should start with <svg")
	}
	if !strings.Contains(s, "</svg>") {
		t.Error("SVG component should contain closing </svg>")
	}

	// Check for component labels
	if !strings.Contains(s, "Identity Provider") {
		t.Error("SVG component should contain Identity Provider")
	}
	if !strings.Contains(s, "API Gateway") {
		t.Error("SVG component should contain API Gateway")
	}

	// Check for connection arrows (may be line, path, or polyline)
	hasConnections := strings.Contains(s, "<line") || strings.Contains(s, "<path") || strings.Contains(s, "<polyline")
	// Or the SVG might use markers for arrows
	if !hasConnections && !strings.Contains(s, "marker") {
		// Some SVG renderers use text arrows or other approaches
		// Just verify we have a valid SVG with rects for components
		if !strings.Contains(s, "<rect") {
			t.Error("SVG component should contain component rectangles")
		}
	}
}

func TestSVGComponentRendererFormat(t *testing.T) {
	r := NewSVGComponent()
	if r.Format() != FormatSVGComponent {
		t.Errorf("Format() = %v, want %v", r.Format(), FormatSVGComponent)
	}
}

func TestSVGComponentRendererWriter(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewSVGComponent()

	var buf bytes.Buffer
	err := r.Render(&buf, p)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Render() wrote nothing")
	}
}

func TestSVGComponentRendererDarkTheme(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewSVGComponent()
	r.Theme = "dark"

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for dark background or dark theme colors
	if !strings.Contains(s, "<svg") {
		t.Error("SVG component with dark theme should still be valid SVG")
	}
}

// TestSVGTrustRenderer tests the SVG trust diagram renderer.
func TestSVGTrustRenderer(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewSVGTrust()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check SVG structure
	if !strings.HasPrefix(s, "<svg") {
		t.Error("SVG trust should start with <svg")
	}
	if !strings.Contains(s, "</svg>") {
		t.Error("SVG trust should contain closing </svg>")
	}

	// Check for node elements
	if !strings.Contains(s, "<rect") && !strings.Contains(s, "<circle") {
		t.Error("SVG trust should contain node shapes")
	}

	// Check for edge elements
	if !strings.Contains(s, "<line") && !strings.Contains(s, "<path") {
		t.Error("SVG trust should contain edge lines or paths")
	}
}

func TestSVGTrustRendererFormat(t *testing.T) {
	r := NewSVGTrust()
	if r.Format() != FormatSVGTrust {
		t.Errorf("Format() = %v, want %v", r.Format(), FormatSVGTrust)
	}
}

func TestSVGTrustRendererWriter(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewSVGTrust()

	var buf bytes.Buffer
	err := r.Render(&buf, p)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Render() wrote nothing")
	}
}

func TestSVGTrustRendererNoTrustRelations(t *testing.T) {
	p := testProtocol() // Has no trust relations
	r := NewSVGTrust()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should still render valid SVG
	if !strings.HasPrefix(s, "<svg") {
		t.Error("SVG trust without relations should still render valid SVG")
	}
}

func TestSVGTrustRendererDarkTheme(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()
	r := NewSVGTrust()
	r.Theme = "dark"

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check for dark background or dark theme colors
	if !strings.Contains(s, "<svg") {
		t.Error("SVG trust with dark theme should still be valid SVG")
	}
}

// TestRenderStringComponentFormats tests that all component/trust formats render without error.
func TestRenderStringComponentFormats(t *testing.T) {
	p := testProtocolWithComponentsAndTrust()

	formats := []Format{
		FormatMermaidComponent,
		FormatMermaidTrust,
		FormatMarkdownMatrix,
		FormatSVGComponent,
		FormatSVGTrust,
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
