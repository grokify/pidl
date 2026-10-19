package render_test

import (
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/render"
)

// TestDocumentGolden pins the whole-document outputs for the scenarios
// fixture.
func TestDocumentGolden(t *testing.T) {
	p, err := pidl.ParseFile("testdata/scenarios_demo.json")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	cases := []struct {
		name   string
		format render.Format
		ext    string
	}{
		{"document_scenario_matrix", render.FormatScenarioMatrix, ".md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !render.IsDocumentFormat(tc.format) {
				t.Fatalf("IsDocumentFormat(%s) = false", tc.format)
			}
			got, err := render.RenderString(tc.format, p)
			if err != nil {
				t.Fatalf("RenderString: %v", err)
			}
			checkGoldenFile(t, tc.name, tc.ext, got)
		})
	}
}

func TestScenarioMatrixOptions(t *testing.T) {
	p, err := pidl.ParseFile("testdata/scenarios_demo.json")
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	r := render.NewScenarioMatrix()
	r.IncludeFragments = false
	r.Title = false
	got, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}
	if want := "| Role | 1. PS authorization |\n"; got[:len(want)] != want {
		t.Fatalf("header = %q, want %q", got[:len(want)], want)
	}

	legacy := &pidl.Protocol{ProtocolMeta: pidl.ProtocolMeta{Name: "Legacy"}}
	if got, _ := render.NewScenarioMatrix().RenderString(legacy); got != "# Legacy - Roles by Scenario\n\n*No scenarios defined.*\n" {
		t.Fatalf("legacy output = %q", got)
	}
}
