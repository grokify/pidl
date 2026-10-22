package pidl

import (
	"reflect"
	"testing"
)

func orderTestProtocol() *Protocol {
	return &Protocol{
		Entities: []Entity{{ID: "person"}, {ID: "agent"}, {ID: "resource"}, {ID: "ps"}},
		Scenarios: []Scenario{
			{ID: "agent-identity", Name: "Agent identity"},
			{ID: "interaction", Name: "Interaction", Fragment: true},
			{ID: "three-party", Name: "PS authorization"},
		},
	}
}

func TestScenarioIDs(t *testing.T) {
	p := orderTestProtocol()
	tests := []struct {
		name      string
		fragments bool
		want      []string
	}{
		{"main only", false, []string{"agent-identity", "three-party"}},
		{"with fragments", true, []string{"agent-identity", "three-party", "interaction"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.ScenarioIDs(tt.fragments); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ScenarioIDs(%v) = %v, want %v", tt.fragments, got, tt.want)
			}
		})
	}
}

func TestScenarioByID(t *testing.T) {
	p := orderTestProtocol()
	if s := p.ScenarioByID("interaction"); s == nil || !s.Fragment {
		t.Fatalf("ScenarioByID(interaction) = %+v, want fragment scenario", s)
	}
	if s := p.ScenarioByID("missing"); s != nil {
		t.Fatalf("ScenarioByID(missing) = %+v, want nil", s)
	}
	if !p.HasScenarios() {
		t.Fatal("HasScenarios() = false, want true")
	}
	if (&Protocol{}).HasScenarios() {
		t.Fatal("HasScenarios() on empty protocol = true, want false")
	}
}

func TestEntityIndex(t *testing.T) {
	p := orderTestProtocol()
	if got := p.EntityIndex("resource"); got != 2 {
		t.Fatalf("EntityIndex(resource) = %d, want 2", got)
	}
	if got := p.EntityIndex("missing"); got != -1 {
		t.Fatalf("EntityIndex(missing) = %d, want -1", got)
	}
}

func TestOrderEntities(t *testing.T) {
	p := orderTestProtocol()
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"reorders", []string{"ps", "agent", "person"}, []string{"person", "agent", "ps"}},
		{"dedupes", []string{"agent", "agent", "person"}, []string{"person", "agent"}},
		{"unknown last", []string{"zeta", "ps", "alpha", "agent"}, []string{"agent", "ps", "zeta", "alpha"}},
		{"empty", nil, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.OrderEntities(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("OrderEntities(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestScenarioJSONRoundTrip(t *testing.T) {
	data := []byte(`{
  "protocol": {"id": "x", "name": "X"},
  "entities": [{"id": "a", "name": "A", "type": "client"}, {"id": "b", "name": "B", "type": "server"}],
  "scenarios": [
    {"id": "frag", "name": "Frag", "fragment": true, "flows": [{"from": "a", "to": "b", "action": "ping"}]},
    {"id": "main", "name": "Main", "participants": ["a", "b"], "flows": [{"include": "frag"}]}
  ]
}`)
	p, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(p.Flows) != 0 || len(p.Scenarios) != 2 {
		t.Fatalf("got %d flows, %d scenarios; want 0, 2", len(p.Flows), len(p.Scenarios))
	}
	inc := p.Scenarios[1].Flows[0]
	if !inc.IsInclude() || inc.Include != "frag" {
		t.Fatalf("include step = %+v, want include of frag", inc)
	}
	out, err := p.ToJSONCompact()
	if err != nil {
		t.Fatalf("ToJSONCompact: %v", err)
	}
	p2, err := Parse(out)
	if err != nil {
		t.Fatalf("re-Parse: %v", err)
	}
	if !reflect.DeepEqual(p.Scenarios, p2.Scenarios) {
		t.Fatalf("round trip changed scenarios:\n got %+v\nwant %+v", p2.Scenarios, p.Scenarios)
	}
}
