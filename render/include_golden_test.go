package render_test

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/render"
)

// TestIncludeGolden pins how each sequence format draws a scenario include
// step, collapsed into a box and expanded into the fragment's flows.
func TestIncludeGolden(t *testing.T) {
	base, err := pidl.ParseFile("testdata/scenarios_demo.json")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if errs := base.Validate(); len(errs) > 0 {
		t.Fatalf("fixture does not validate: %v", errs)
	}

	formats := []struct {
		format render.Format
		ext    string
	}{
		{render.FormatPlantUML, ".puml"},
		{render.FormatMermaid, ".mmd"},
		{render.FormatD2, ".d2"},
		{render.FormatSVG, ".svg"},
	}
	for _, expand := range []bool{false, true} {
		mode := "collapsed"
		if expand {
			mode = "expanded"
		}
		proj, err := base.Scenario("three-party", pidl.ProjectOptions{Expand: expand})
		if err != nil {
			t.Fatalf("Scenario: %v", err)
		}
		for _, f := range formats {
			name := "include_" + mode + "_" + string(f.format)
			t.Run(name, func(t *testing.T) {
				if !render.SupportsIncludes(f.format) {
					t.Fatalf("SupportsIncludes(%s) = false", f.format)
				}
				got, err := render.RenderString(f.format, proj)
				if err != nil {
					t.Fatalf("RenderString: %v", err)
				}
				hasBox := strings.Contains(got, "ref") && strings.Contains(got, "Consent")
				hasFragmentFlows := strings.Contains(got, "Consent prompt")
				if expand && (!hasFragmentFlows || strings.Contains(got, "ref: Consent") || strings.Contains(got, "ref over")) {
					t.Errorf("expanded output should inline the fragment and draw no include box")
				}
				if !expand && (!hasBox || hasFragmentFlows) {
					t.Errorf("collapsed output should draw an include box and not the fragment's flows")
				}
				checkGoldenFile(t, name, f.ext, got)
			})
		}
	}
}

func TestSupportsIncludes(t *testing.T) {
	for _, f := range []render.Format{render.FormatDOT, render.FormatD2Flow, render.FormatSVGNetwork, render.FormatMermaidState} {
		if render.SupportsIncludes(f) {
			t.Errorf("SupportsIncludes(%s) = true, want false", f)
		}
	}
}
