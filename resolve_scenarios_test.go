package pidl

import (
	"path/filepath"
	"reflect"
	"testing"
)

func scenarioIDList(p *Protocol) []string {
	ids := make([]string, len(p.Scenarios))
	for i, sc := range p.Scenarios {
		ids[i] = sc.ID
	}
	return ids
}

func writeProtocol(t *testing.T, dir, name string, p *Protocol) {
	t.Helper()
	if err := p.WriteFile(filepath.Join(dir, name)); err != nil {
		t.Fatal(err)
	}
}

func resolveIn(t *testing.T, dir string, p *Protocol) *Protocol {
	t.Helper()
	opts := DefaultResolveOptions()
	opts.BasePath = dir
	got, err := p.Resolve(opts)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	return got
}

func TestResolveExtendsScenarios(t *testing.T) {
	dir := t.TempDir()
	base := scenarioProtocol()
	writeProtocol(t, dir, "base.json", base)

	derived := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "derived", Name: "Derived"},
		Extends:      &ProtocolExtends{Path: "base.json"},
		Entities:     []Entity{{ID: "as", Name: "Access Server", Type: EntityTypeAuthorizationServer}},
		Scenarios: []Scenario{
			{ID: "agent-identity", Name: "Agent identity (revised)", Flows: []Flow{
				{From: "agent", To: "resource", Action: "signed_request_v2"},
			}},
			{ID: "four-party", Name: "Federated", Flows: []Flow{
				{From: "agent", To: "as", Action: "token_request"},
				{Include: "interaction"},
			}},
		},
	}
	got := resolveIn(t, dir, derived)

	want := []string{"agent-identity", "interaction", "three-party", "four-party"}
	if ids := scenarioIDList(got); !reflect.DeepEqual(ids, want) {
		t.Fatalf("scenarios = %v, want %v (override keeps base position)", ids, want)
	}
	if got.Scenarios[0].Name != "Agent identity (revised)" {
		t.Errorf("override not applied: %+v", got.Scenarios[0])
	}
	if errs := got.Validate(); len(errs) > 0 {
		t.Fatalf("resolved protocol does not validate: %v", errs)
	}
	if base.Scenarios[0].Name != "Agent identity" {
		t.Error("resolve mutated the base protocol")
	}
}

func TestResolveExtendsExclusions(t *testing.T) {
	dir := t.TempDir()
	writeProtocol(t, dir, "base.json", scenarioProtocol())

	derived := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "derived", Name: "Derived"},
		Extends: &ProtocolExtends{
			Path:             "base.json",
			ExcludeScenarios: []string{"agent-identity"},
			ExcludeEntities:  []string{"resource"},
			ExcludePhases:    []string{"consent"},
		},
	}
	got := resolveIn(t, dir, derived)

	if ids := scenarioIDList(got); !reflect.DeepEqual(ids, []string{"interaction", "three-party"}) {
		t.Fatalf("scenarios = %v", ids)
	}
	tp := got.ScenarioByID("three-party")
	if !reflect.DeepEqual(tp.Participants, []string{"person", "agent", "ps"}) {
		t.Errorf("participants = %v, want resource removed", tp.Participants)
	}
	for _, f := range tp.Flows {
		if f.From == "resource" || f.To == "resource" {
			t.Errorf("flow still references excluded entity: %+v", f)
		}
		if f.Phase == "consent" {
			t.Errorf("flow still references excluded phase: %+v", f)
		}
	}
}

func TestResolveImportScenariosWithAlias(t *testing.T) {
	dir := t.TempDir()
	writeProtocol(t, dir, "lib.json", scenarioProtocol())

	target := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "target", Name: "Target"},
		Imports: []ProtocolImport{{
			Path: "lib.json", Alias: "lib_",
			Phases:    []string{"consent"},
			Scenarios: []string{"three-party"},
		}},
	}
	got := resolveIn(t, dir, target)

	// The included fragment comes along, in the imported document's order.
	if ids := scenarioIDList(got); !reflect.DeepEqual(ids, []string{"lib_interaction", "lib_three-party"}) {
		t.Fatalf("scenarios = %v", ids)
	}
	tp := got.ScenarioByID("lib_three-party")
	if tp.Flows[0].From != "lib_agent" || tp.Flows[2].Include != "lib_interaction" || tp.Flows[2].Phase != "lib_consent" {
		t.Errorf("alias not applied to flows: %+v", tp.Flows)
	}
	if tp.Participants[0] != "lib_person" {
		t.Errorf("alias not applied to participants: %v", tp.Participants)
	}
	if errs := got.Validate(); len(errs) > 0 {
		t.Fatalf("resolved protocol does not validate: %v", errs)
	}
}

func TestResolveMixesFlowsIntoFirstScenario(t *testing.T) {
	dir := t.TempDir()
	legacy := scenarioProtocol()
	legacy.ProtocolMeta = ProtocolMeta{ID: "legacy", Name: "Legacy"}
	legacy.Flows, legacy.Scenarios = legacy.Scenarios[0].Flows, nil
	writeProtocol(t, dir, "legacy.json", legacy)

	derived := scenarioProtocol()
	derived.Extends = &ProtocolExtends{Path: "legacy.json"}
	got := resolveIn(t, dir, derived)

	if len(got.Flows) != 0 {
		t.Fatalf("flows = %d, want 0 after normalization", len(got.Flows))
	}
	want := []string{"aauth", "agent-identity", "interaction", "three-party"}
	if ids := scenarioIDList(got); !reflect.DeepEqual(ids, want) {
		t.Fatalf("scenarios = %v, want %v", ids, want)
	}
	if got.Scenarios[0].Name != "AAuth" || len(got.Scenarios[0].Flows) != 2 {
		t.Errorf("first scenario = %+v", got.Scenarios[0])
	}
	if errs := got.Validate(); len(errs) > 0 {
		t.Fatalf("resolved protocol does not validate: %v", errs)
	}
}

func TestNormalizeScenariosIDCollision(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "x", Name: "X"},
		Flows:        []Flow{{From: "a", To: "b", Action: "f"}},
		Scenarios:    []Scenario{{ID: "x", Name: "X scenario"}},
	}
	got := normalizeScenarios(p)
	if ids := scenarioIDList(got); !reflect.DeepEqual(ids, []string{"x-flows", "x"}) {
		t.Fatalf("scenarios = %v", ids)
	}
}
