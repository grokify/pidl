package render_test

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/examples"
	"github.com/grokify/pidl/render"
	"github.com/grokify/pidl/render/svg"
)

// update rewrites the golden files instead of comparing against them:
//
//	go test ./render -run TestSVGGolden -update
var update = flag.Bool("update", false, "rewrite golden files")

const goldenDir = "testdata/golden"

// goldenCase describes one end-to-end render whose output is pinned to a
// golden file at testdata/golden/<name>.svg.
type goldenCase struct {
	name string
	// example is the built-in example name passed to examples.GetProtocol.
	example string
	// mutate optionally adjusts the freshly parsed example before rendering.
	mutate func(p *pidl.Protocol)
	// render produces the output for this case.
	render func(p *pidl.Protocol) (string, error)
}

// renderSequence returns a render func for the sequence SVG renderer after
// applying cfg to a fresh render.NewSVG().
func renderSequence(cfg func(r *render.SVGRenderer)) func(*pidl.Protocol) (string, error) {
	return func(p *pidl.Protocol) (string, error) {
		r := render.NewSVG()
		if cfg != nil {
			cfg(r)
		}
		return r.RenderString(p)
	}
}

// renderTemplate returns a render func that uses a built-in template, with
// an optional post-construction cfg hook.
func renderTemplate(template string, cfg func(r *render.SVGRenderer)) func(*pidl.Protocol) (string, error) {
	return func(p *pidl.Protocol) (string, error) {
		r, err := render.NewSVGWithTemplate(template)
		if err != nil {
			return "", err
		}
		if cfg != nil {
			cfg(r)
		}
		return r.RenderString(p)
	}
}

// renderNetwork returns a render func for the network boundary renderer.
// boundaries use the CLI -boundary flag format: boundary_id:entity1,entity2.
func renderNetwork(boundaries []string, cfg func(r *render.SVGNetworkRenderer)) func(*pidl.Protocol) (string, error) {
	return func(p *pidl.Protocol) (string, error) {
		r := render.NewSVGNetwork()
		for _, b := range boundaries {
			id, entities, err := parseBoundaryFlag(b)
			if err != nil {
				return "", err
			}
			r.AddBoundaryOverride(id, entities)
		}
		if cfg != nil {
			cfg(r)
		}
		return r.RenderString(p)
	}
}

func renderComponent(cfg func(r *render.SVGComponentRenderer)) func(*pidl.Protocol) (string, error) {
	return func(p *pidl.Protocol) (string, error) {
		r := render.NewSVGComponent()
		if cfg != nil {
			cfg(r)
		}
		return r.RenderString(p)
	}
}

func renderTrust(cfg func(r *render.SVGTrustRenderer)) func(*pidl.Protocol) (string, error) {
	return func(p *pidl.Protocol) (string, error) {
		r := render.NewSVGTrust()
		if cfg != nil {
			cfg(r)
		}
		return r.RenderString(p)
	}
}

func renderInfographic(opts render.InfographicOptions) func(*pidl.Protocol) (string, error) {
	return func(p *pidl.Protocol) (string, error) {
		return render.NewInfographicRenderer(opts).Render(p), nil
	}
}

// parseBoundaryFlag mirrors the CLI's -boundary parsing (boundary_id:entity1,entity2).
func parseBoundaryFlag(value string) (string, []string, error) {
	id, list, ok := strings.Cut(value, ":")
	if !ok || id == "" || list == "" {
		return "", nil, fmt.Errorf("invalid boundary flag %q: want boundary_id:entity1,entity2", value)
	}
	var entities []string
	for _, e := range strings.Split(list, ",") {
		if e = strings.TrimSpace(e); e != "" {
			entities = append(entities, e)
		}
	}
	return id, entities, nil
}

// applyAnimationPresets assigns every animation preset (plus one fully
// custom style and one explicitly disabled flow) across the protocol flows.
func applyAnimationPresets(p *pidl.Protocol) {
	presets := []pidl.AnimationPreset{
		pidl.AnimationPresetRequest,
		pidl.AnimationPresetResponse,
		pidl.AnimationPresetSuccess,
		pidl.AnimationPresetError,
		pidl.AnimationPresetWarning,
		pidl.AnimationPresetHighlight,
		pidl.AnimationPresetNone,
	}
	disabled := false
	for i := range p.Flows {
		switch {
		case i == len(presets):
			p.Flows[i].Animation = &pidl.FlowAnimation{
				Duration: "1s",
				Delay:    "0.25s",
				DotColor: "#ff6600",
				DotSize:  6,
				Pulse:    true,
				Easing:   "ease-in-out",
			}
		case i == len(presets)+1:
			p.Flows[i].Animation = &pidl.FlowAnimation{Enabled: &disabled}
		default:
			p.Flows[i].Animation = &pidl.FlowAnimation{Preset: presets[i%len(presets)]}
		}
	}
}

// applyNetworkMetadata configures explicit network boundaries via protocol
// metadata (metadata.networks, metadata.network_layout, entity.metadata.network).
func applyNetworkMetadata(p *pidl.Protocol) {
	if p.Metadata == nil {
		p.Metadata = &pidl.ProtocolMetadata{}
	}
	p.Metadata.Networks = map[string]*pidl.NetworkConfig{
		"cloud": {
			Name:        "Cloud Provider",
			Style:       "cloud",
			Entities:    []string{"auth_server", "resource_server"},
			Description: "Managed identity and API hosting",
			Color:       "#718096",
		},
		"corp": {
			Name:  "Corporate Network",
			Style: "trusted",
		},
		"perimeter": {
			Style: "dmz",
		},
	}
	p.Metadata.NetworkLayout = &pidl.NetworkLayoutConfig{
		Direction: "vertical",
		Order:     []string{"corp", "perimeter", "cloud"},
	}
	for i := range p.Entities {
		switch p.Entities[i].ID {
		case "client":
			p.Entities[i].Metadata = &pidl.EntityMetadata{Network: "corp"}
		case "browser":
			p.Entities[i].Metadata = &pidl.EntityMetadata{Network: "perimeter"}
		}
	}
}

func goldenCases() []goldenCase {
	cases := []goldenCase{
		// FormatSVG: sequence diagrams.
		{name: "svg_oauth2_with_security", example: "oauth2_with_security", render: renderSequence(nil)},
		{name: "svg_webauthn_registration", example: "webauthn_registration", render: renderSequence(nil)},
		{name: "svg_oauth2_with_states", example: "oauth2_with_states", render: renderSequence(nil)},
		{name: "svg_a2a_agent_delegation", example: "a2a_agent_delegation", render: renderSequence(nil)},
		{name: "svg_process_ci_cd_pipeline", example: "process/ci_cd_pipeline", render: renderSequence(nil)},
		{name: "svg_theme_dark_mcp_tool_invocation", example: "mcp_tool_invocation", render: renderSequence(func(r *render.SVGRenderer) {
			r.Theme = "dark"
		})},
		{name: "svg_theme_auto_oauth2_authorization_code", example: "oauth2_authorization_code", render: renderSequence(func(r *render.SVGRenderer) {
			r.Theme = "auto"
		})},
		{name: "svg_minimal_options_oauth2_pkce", example: "oauth2_pkce", render: renderSequence(func(r *render.SVGRenderer) {
			r.Title = false
			r.ShowStepNumbers = false
			r.ShowPhases = false
			r.ShowNotes = false
			r.ShowAlternatives = false
			r.ShowSecurity = false
			r.ParticipantSpacing = 150
			r.MessageSpacing = 36
		})},
		{name: "svg_template_dir_sketch_oauth2_pkce", example: "oauth2_pkce", render: func(p *pidl.Protocol) (string, error) {
			r, err := render.NewSVGWithTemplateDir(filepath.Join("svg", "templates", "sketch"))
			if err != nil {
				return "", err
			}
			return r.RenderString(p)
		}},

		// FormatSVGAnimated.
		{name: "svg_animated_mcp_tool_invocation", example: "mcp_tool_invocation", render: func(p *pidl.Protocol) (string, error) {
			return render.NewSVGAnimated().RenderString(p)
		}},
		{name: "svg_animated_presets_a2a_agent_delegation", example: "a2a_agent_delegation", mutate: applyAnimationPresets, render: func(p *pidl.Protocol) (string, error) {
			return render.NewSVGAnimated().RenderString(p)
		}},
		{name: "svg_animated_stagger_sequential_oauth2_pkce", example: "oauth2_pkce", render: renderSequence(func(r *render.SVGRenderer) {
			r.Animated = true
			r.AnimationConfig.StaggerMode = svg.StaggerSequential
		})},
		{name: "svg_animated_stagger_parallel_oauth2_pkce", example: "oauth2_pkce", render: renderSequence(func(r *render.SVGRenderer) {
			r.Animated = true
			r.AnimationConfig.StaggerMode = svg.StaggerParallel
			r.AnimationConfig.DefaultDotColor = "#2b6cb0"
			r.AnimationConfig.DefaultDotSize = 5
		})},
		{name: "svg_animated_template_dark_oidc_authentication", example: "oidc_authentication", render: renderTemplate("dark", func(r *render.SVGRenderer) {
			r.Animated = true
			r.Theme = "dark"
		})},

		// FormatSVGNetwork.
		{name: "svg_network_trust_inferred_oauth2_with_security", example: "oauth2_with_security", render: renderNetwork(nil, nil)},
		{name: "svg_network_boundaries_oauth2_authorization_code", example: "oauth2_authorization_code", render: renderNetwork([]string{
			"external:user,browser",
			"dmz:client",
			"internal:auth_server,resource_server",
		}, nil)},
		{name: "svg_network_boundaries_vertical_oauth2_with_security", example: "oauth2_with_security", render: renderNetwork([]string{
			"cloud:auth_server",
			"private:resource_server",
		}, func(r *render.SVGNetworkRenderer) {
			r.Direction = "vertical"
		})},
		{name: "svg_network_metadata_oauth2_pkce", example: "oauth2_pkce", mutate: applyNetworkMetadata, render: renderNetwork([]string{
			"perimeter:user",
		}, nil)},
		{name: "svg_network_dark_no_inference_mcp_tool_invocation", example: "mcp_tool_invocation", render: renderNetwork(nil, func(r *render.SVGNetworkRenderer) {
			r.Theme = "dark"
			r.Title = false
			r.InferBoundariesFromTrust = false
		})},
		{name: "svg_network_saml2_sso", example: "saml2_sso", render: renderNetwork(nil, nil)},

		// FormatSVGComponent.
		{name: "svg_component_oaiaf_architecture", example: "oaiaf_architecture", render: renderComponent(nil)},
		{name: "svg_component_dark_bare_oaiaf_architecture", example: "oaiaf_architecture", render: renderComponent(func(r *render.SVGComponentRenderer) {
			r.Theme = "dark"
			r.Title = false
			r.ShowEntities = false
			r.ShowRoles = false
		})},
		{name: "svg_component_empty_oauth2_pkce", example: "oauth2_pkce", render: renderComponent(nil)},

		// FormatSVGTrust.
		{name: "svg_trust_oaiaf_architecture", example: "oaiaf_architecture", render: renderTrust(nil)},
		{name: "svg_trust_dark_bare_oaiaf_architecture", example: "oaiaf_architecture", render: renderTrust(func(r *render.SVGTrustRenderer) {
			r.Theme = "dark"
			r.Title = false
			r.ShowCredentials = false
			r.ShowComponents = false
		})},
		{name: "svg_trust_empty_oauth2_pkce", example: "oauth2_pkce", render: renderTrust(nil)},

		// FormatInfographic.
		{name: "infographic_process_ci_cd_pipeline", example: "process/ci_cd_pipeline", render: renderInfographic(render.DefaultInfographicOptions())},
		{name: "infographic_datasheet_process_etl_pipeline", example: "process/etl_pipeline", render: renderInfographic(render.DatasheetTileOptions())},
	}

	// Every built-in template for FormatSVG.
	for _, tmpl := range []string{"default", "minimal", "sketch", "blueprint", "dark", "high-contrast"} {
		cases = append(cases, goldenCase{
			name:    "svg_template_" + tmpl + "_oauth2_authorization_code",
			example: "oauth2_authorization_code",
			render:  renderTemplate(tmpl, nil),
		})
	}
	return cases
}

func TestSVGGolden(t *testing.T) {
	cases := goldenCases()
	seen := make(map[string]bool, len(cases))
	for _, tc := range cases {
		if seen[tc.name] {
			t.Fatalf("duplicate golden case name %q", tc.name)
		}
		seen[tc.name] = true
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := examples.GetProtocol(tc.example)
			if err != nil {
				t.Fatalf("examples.GetProtocol(%q): %v", tc.example, err)
			}
			if tc.mutate != nil {
				tc.mutate(p)
			}
			got, err := tc.render(p)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got == "" {
				t.Fatal("render returned empty output")
			}
			if !strings.HasPrefix(got, "<svg") {
				t.Fatalf("output does not start with <svg: %.40q", got)
			}
			checkGolden(t, tc.name, got)
		})
	}
}

// checkGolden compares got with testdata/golden/<name>.svg, rewriting the
// file when -update is set.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	checkGoldenFile(t, name, ".svg", got)
}

// checkGoldenFile compares got with testdata/golden/<name><ext>, rewriting
// the file when -update is set.
func checkGoldenFile(t *testing.T, name, ext, got string) {
	t.Helper()
	path := filepath.Join(goldenDir, name+ext)

	if *update {
		if err := os.MkdirAll(goldenDir, 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", goldenDir, err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
		return
	}

	wantBytes, err := os.ReadFile(path) //nolint:gosec // path is built from a fixed test-case name under testdata
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("golden file %s missing; regenerate with: go test ./render -run Golden -update", path)
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	want := string(wantBytes)
	if got == want {
		return
	}

	line, gotLine, wantLine := firstDifferingLine(got, want)
	t.Errorf("output differs from %s at line %d\n  want: %s\n  got:  %s\n(got %d bytes, want %d bytes)\nregenerate with: go test ./render -run Golden/%s -update",
		path, line, wantLine, gotLine, len(got), len(want), name)
}

// firstDifferingLine returns the 1-based line number and contents of the first
// line that differs between got and want.
func firstDifferingLine(got, want string) (int, string, string) {
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")
	for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
		var g, w string
		if i < len(gotLines) {
			g = gotLines[i]
		} else {
			g = "<EOF>"
		}
		if i < len(wantLines) {
			w = wantLines[i]
		} else {
			w = "<EOF>"
		}
		if g != w {
			return i + 1, g, w
		}
	}
	return 0, "", ""
}
