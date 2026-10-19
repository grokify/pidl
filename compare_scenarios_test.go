package pidl

import (
	"reflect"
	"strings"
	"testing"
)

func diffPaths(d *ProtocolDiff, category DiffCategory) []string {
	var out []string
	for _, it := range d.Items {
		if it.Category == category {
			out = append(out, string(it.Type)+" "+it.Path)
		}
	}
	return out
}

func TestCompareScenarios(t *testing.T) {
	base := scenarioProtocol()
	next := scenarioProtocol()
	next.Scenarios[0].Name = "Agent identity v2"                       // renamed
	next.Scenarios[2].Flows[1].Action = "token_request"                // changed flow
	next.Scenarios = append(next.Scenarios[:1], next.Scenarios[2:]...) // drop interaction...
	next.Scenarios = append(next.Scenarios, Scenario{ID: "interaction", Name: "Interaction", Fragment: true,
		Flows: []Flow{{From: "ps", To: "person", Action: "consent_prompt"}}}) // ...and re-add it last
	next.Scenarios = append(next.Scenarios, Scenario{ID: "four-party", Name: "Federated", Flows: []Flow{{Include: "interaction"}}})

	d := Compare(base, next, DefaultDiffOptions())

	wantScenario := []string{
		"added scenarios[four-party]",
		"modified scenarios[agent-identity].name",
		"removed scenarios[interaction].flows[1]",
		"modified scenarios[three-party].flows[1].action",
	}
	if got := diffPaths(d, DiffCategoryScenario); !reflect.DeepEqual(got, wantScenario[:2]) {
		t.Errorf("scenario items = %v, want %v", got, wantScenario[:2])
	}
	wantFlow := []string{"removed scenarios[interaction].flows[1]", "modified scenarios[three-party].flows[1].action"}
	if got := diffPaths(d, DiffCategoryFlow); !reflect.DeepEqual(got, wantFlow) {
		t.Errorf("flow items = %v, want %v", got, wantFlow)
	}
	if got := diffPaths(d, DiffCategoryOrder); !reflect.DeepEqual(got, []string{"modified scenarios[three-party]", "modified scenarios[interaction]"}) {
		t.Errorf("order items = %v", got)
	}
}

func TestCompareOrderIgnoresInsertions(t *testing.T) {
	base := scenarioProtocol()
	next := scenarioProtocol()
	next.Entities = append([]Entity{{ID: "ap", Name: "Agent Provider", Type: EntityTypeServer}}, next.Entities...)
	d := Compare(base, next, DefaultDiffOptions())
	if got := diffPaths(d, DiffCategoryOrder); got != nil {
		t.Fatalf("an insertion should not count as a move, got %v", got)
	}

	next.Entities[1], next.Entities[2] = next.Entities[2], next.Entities[1] // swap person and agent
	d = Compare(base, next, DefaultDiffOptions())
	var summaries []string
	for _, it := range d.Items {
		if it.Category == DiffCategoryOrder {
			summaries = append(summaries, it.Summary)
		}
	}
	want := []string{`entity "person" moved from position 1 to 3`}
	if !reflect.DeepEqual(summaries, want) {
		t.Fatalf("order summaries = %v, want %v", summaries, want)
	}
}

func TestCompareIncludeChange(t *testing.T) {
	base := scenarioProtocol()
	next := scenarioProtocol()
	next.Scenarios = append(next.Scenarios, Scenario{ID: "consent-v2", Name: "Consent v2", Fragment: true,
		Flows: []Flow{{From: "ps", To: "person", Action: "prompt"}}})
	next.Scenarios[2].Flows[2].Include = "consent-v2"
	d := Compare(base, next, DefaultDiffOptions())
	found := false
	for _, it := range d.Items {
		if it.Path == "scenarios[three-party].flows[2].include" {
			found = true
		}
	}
	if !found {
		t.Fatalf("include change not reported: %+v", d.Items)
	}
}

func TestCompareDeterministic(t *testing.T) {
	base := scenarioProtocol()
	next := scenarioProtocol()
	next.Entities = next.Entities[1:]
	next.Entities = append(next.Entities, Entity{ID: "as", Name: "AS"}, Entity{ID: "ap", Name: "AP"})
	next.Phases = []Phase{{ID: "p1", Name: "P1"}, {ID: "p2", Name: "P2"}}
	first := Compare(base, next, DefaultDiffOptions()).Items
	for i := 0; i < 30; i++ {
		if got := Compare(base, next, DefaultDiffOptions()).Items; !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d produced a different item order", i)
		}
	}
}

func TestCompareOutputIncludesScenarioCategories(t *testing.T) {
	base := scenarioProtocol()
	next := scenarioProtocol()
	next.Scenarios[0].Name = "Renamed"
	next.Scenarios[0], next.Scenarios[2] = next.Scenarios[2], next.Scenarios[0]
	d := Compare(base, next, DefaultDiffOptions())
	for name, out := range map[string]string{"String": d.String(), "Markdown": d.ToMarkdown()} {
		for _, want := range []string{"scenario renamed", "moved from position"} {
			if !strings.Contains(out, want) {
				t.Errorf("%s output missing %q:\n%s", name, want, out)
			}
		}
	}
}
