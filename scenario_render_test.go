package pidl_test

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/render"
)

// TestScenarioProjectionRenders checks that an expanded projection renders
// with existing renderers, with participants in entities order.
func TestScenarioProjectionRenders(t *testing.T) {
	p, err := pidl.Parse([]byte(`{
  "protocol": {"id": "demo", "name": "Demo"},
  "entities": [
    {"id": "person", "name": "Person", "type": "user"},
    {"id": "agent", "name": "Agent", "type": "agent"},
    {"id": "ps", "name": "PS", "type": "authorization_server"}
  ],
  "scenarios": [
    {"id": "consent", "name": "Consent", "fragment": true, "flows": [
      {"from": "ps", "to": "person", "action": "prompt"}
    ]},
    {"id": "main", "name": "Main", "flows": [
      {"from": "agent", "to": "ps", "action": "request"},
      {"include": "consent"}
    ]}
  ]
}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if errs := p.Validate(); len(errs) > 0 {
		t.Fatalf("Validate: %v", errs)
	}
	proj, err := p.Scenario("main", pidl.ProjectOptions{Expand: true})
	if err != nil {
		t.Fatalf("Scenario: %v", err)
	}
	out, err := render.RenderString(render.FormatMermaid, proj)
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}
	iPerson := strings.Index(out, "participant person")
	iAgent := strings.Index(out, "participant agent")
	iPS := strings.Index(out, "participant ps")
	if iPerson < 0 || iAgent < 0 || iPS < 0 || !(iPerson < iAgent && iAgent < iPS) {
		t.Fatalf("participants not declared in entities order:\n%s", out)
	}
	if !strings.Contains(out, "prompt") {
		t.Fatalf("expanded fragment flow missing:\n%s", out)
	}
}
