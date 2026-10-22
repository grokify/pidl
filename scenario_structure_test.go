package pidl

import (
	"reflect"
	"strings"
	"testing"
)

// structuredProtocol returns a valid protocol with groups, a variant and
// flow IDs. Document order deliberately differs from output order.
func structuredProtocol() *Protocol {
	p := scenarioProtocol()
	p.Groups = []Group{
		{ID: "access", Name: "Access modes"},
		{ID: "identity", Name: "Identity"},
	}
	p.Scenarios[0].Group = "identity" // agent-identity
	p.Scenarios[2].Group = "access"   // three-party
	p.Scenarios[0].Flows[0].ID = "sign"
	p.Scenarios[0].Flows[1].ID = "ok"
	p.Scenarios[1].Flows[0].ID = "prompt" // interaction fragment
	p.Scenarios[1].Flows[1].ID = "approve"
	p.Scenarios[2].Flows[2].ID = "consent" // the include step
	p.Scenarios = append(p.Scenarios, Scenario{
		ID: "agent-identity-approval", Name: "Agent identity with approval",
		Group: "identity", VariantOf: "agent-identity", VariantLabel: "With user approval",
		Flows: []Flow{{ID: "sign", From: "agent", To: "resource", Action: "signed_request"}},
	}, Scenario{
		ID: "ungrouped", Name: "Ungrouped",
		Flows: []Flow{{From: "agent", To: "resource", Action: "ping"}},
	})
	return p
}

func TestStructuredProtocolValid(t *testing.T) {
	if errs := structuredProtocol().Validate(); len(errs) > 0 {
		t.Fatalf("Validate() = %v, want no errors", errs)
	}
}

func TestOutputOrder(t *testing.T) {
	tests := []struct {
		name string
		p    *Protocol
		want []string
	}{
		{"no groups or variants keeps document order", scenarioProtocol(),
			[]string{"agent-identity", "three-party", "interaction"}},
		{"groups, then variants after their base, then ungrouped, then fragments", structuredProtocol(),
			[]string{"three-party", "agent-identity", "agent-identity-approval", "ungrouped", "interaction"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.ScenarioIDs(true); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ScenarioIDs(true) = %v, want %v", got, tt.want)
			}
			var summaryIDs []string
			for i, s := range tt.p.ScenarioSummaries() {
				if s.Position != i+1 {
					t.Errorf("summary %q position = %d, want %d", s.ID, s.Position, i+1)
				}
				summaryIDs = append(summaryIDs, s.ID)
			}
			if !reflect.DeepEqual(summaryIDs, tt.want) {
				t.Fatalf("ScenarioSummaries order = %v, want %v", summaryIDs, tt.want)
			}
		})
	}
}

func TestValidateStructureErrors(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(p *Protocol)
		wantField string
		wantMsg   string
	}{
		{"group id pattern", func(p *Protocol) { p.Groups[0].ID = "Access" }, "groups[0].id", "must match pattern"},
		{"duplicate group", func(p *Protocol) {
			p.Groups[1].ID = "access"
			p.Scenarios[0].Group = "access"
			p.Scenarios[3].Group = "access"
		}, "groups[1].id", "duplicate group ID"},
		{"group name required", func(p *Protocol) { p.Groups[0].Name = "" }, "groups[0].name", "required"},
		{"unknown group", func(p *Protocol) { p.Scenarios[2].Group = "nope" }, "scenarios[2].group", `unknown group "nope"`},
		{"fragment with group", func(p *Protocol) { p.Scenarios[1].Group = "access" }, "scenarios[1].group", "cannot have a group"},
		{"label without variant_of", func(p *Protocol) { p.Scenarios[2].VariantLabel = "x" }, "scenarios[2].variant_label", "requires variant_of"},
		{"variant_of without label", func(p *Protocol) { p.Scenarios[3].VariantLabel = "" }, "scenarios[3].variant_label", "required with variant_of"},
		{"unknown base", func(p *Protocol) { p.Scenarios[3].VariantOf = "nope" }, "scenarios[3].variant_of", `unknown scenario "nope"`},
		{"self variant", func(p *Protocol) { p.Scenarios[3].VariantOf = p.Scenarios[3].ID }, "scenarios[3].variant_of", "variant of itself"},
		{"fragment base", func(p *Protocol) { p.Scenarios[3].VariantOf = "interaction" }, "scenarios[3].variant_of", "is a fragment"},
		{"variant of a variant", func(p *Protocol) {
			p.Scenarios = append(p.Scenarios, Scenario{ID: "v2", Name: "V2", Group: "identity", VariantOf: "agent-identity-approval", VariantLabel: "Again",
				Flows: []Flow{{From: "agent", To: "resource", Action: "x"}}})
		}, "scenarios[5].variant_of", "is itself a variant"},
		{"variant in another group", func(p *Protocol) { p.Scenarios[3].Group = "access" }, "scenarios[3].group", "must match the group of its base"},
		{"flow id pattern", func(p *Protocol) { p.Scenarios[0].Flows[0].ID = "Sign" }, "scenarios[0].flows[0].id", "must match pattern"},
		{"flow id has a dot", func(p *Protocol) { p.Scenarios[0].Flows[0].ID = "a.b" }, "scenarios[0].flows[0].id", "must match pattern"},
		{"duplicate flow id", func(p *Protocol) { p.Scenarios[0].Flows[1].ID = "sign" }, "scenarios[0].flows[1].id", `duplicate flow ID "sign"`},
		{"fragment included twice without ids", func(p *Protocol) {
			p.Scenarios[2].Flows[2].ID = ""
			p.Scenarios[2].Flows = append(p.Scenarios[2].Flows, Flow{Include: "interaction"})
		}, "scenarios[2].flows", `step ID "interaction.prompt" repeats`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := structuredProtocol()
			tt.mutate(p)
			errs := p.Validate()
			for _, e := range errs {
				if e.Field == tt.wantField && strings.Contains(e.Message, tt.wantMsg) {
					return
				}
			}
			t.Fatalf("Validate() = %v, want %s: %s", errs, tt.wantField, tt.wantMsg)
		})
	}
}

func TestFragmentIncludedTwiceWithDistinctIDs(t *testing.T) {
	p := structuredProtocol()
	p.Scenarios[2].Flows = append(p.Scenarios[2].Flows, Flow{ID: "reconsent", Include: "interaction"})
	if errs := p.Validate(); len(errs) > 0 {
		t.Fatalf("Validate() = %v, want no errors", errs)
	}
}

func TestWalkthroughStructure(t *testing.T) {
	w := structuredProtocol().Walkthrough(nil)
	if len(w.Groups) != 2 || w.Groups[0].ID != "access" {
		t.Fatalf("groups = %+v", w.Groups)
	}
	byID := map[string]WalkthroughScenario{}
	for _, sc := range w.Scenarios {
		byID[sc.ID] = sc
	}
	v := byID["agent-identity-approval"]
	if v.Group != "identity" || v.VariantOf != "agent-identity" || v.VariantLabel != "With user approval" {
		t.Errorf("variant = %+v", v)
	}
	var ids []string
	for _, st := range byID["three-party"].Steps {
		ids = append(ids, st.ID)
	}
	// three-party: request, auth_token_request, consent.prompt, consent.approve, auth_token
	want := []string{"", "", "consent.prompt", "consent.approve", ""}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("three-party step IDs = %q, want %q", ids, want)
	}
	if got := byID["agent-identity"].Steps[0].ID; got != "sign" {
		t.Errorf("agent-identity step 1 ID = %q, want sign", got)
	}
}

func TestOutlineStructure(t *testing.T) {
	o := structuredProtocol().Outline()
	if len(o.Groups) != 2 {
		t.Fatalf("outline groups = %+v", o.Groups)
	}
	for _, sc := range o.Scenarios {
		if sc.ID == "agent-identity-approval" && (sc.Group != "identity" || sc.VariantOf != "agent-identity") {
			t.Errorf("outline variant = %+v", sc)
		}
	}
}

func TestResolveExtendsGroups(t *testing.T) {
	dir := t.TempDir()
	writeProtocol(t, dir, "base.json", structuredProtocol())
	derived := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "derived", Name: "Derived"},
		Extends:      &ProtocolExtends{Path: "base.json"},
		Groups: []Group{
			{ID: "identity", Name: "Agent identity"},
			{ID: "advanced", Name: "Advanced"},
		},
	}
	got := resolveIn(t, dir, derived)
	want := []Group{{ID: "access", Name: "Access modes"}, {ID: "identity", Name: "Agent identity"}, {ID: "advanced", Name: "Advanced"}}
	if !reflect.DeepEqual(got.Groups, want) {
		t.Fatalf("groups = %+v, want %+v", got.Groups, want)
	}
	if errs := got.Validate(); len(errs) > 0 {
		t.Fatalf("resolved protocol does not validate: %v", errs)
	}
}

func TestResolveImportGroupsAndVariants(t *testing.T) {
	dir := t.TempDir()
	writeProtocol(t, dir, "lib.json", structuredProtocol())
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "app", Name: "App"},
		Entities:     []Entity{{ID: "client", Name: "Client", Type: EntityTypeClient}},
		Flows:        []Flow{{From: "client", To: "client", Action: "noop"}},
		Imports: []ProtocolImport{{
			Path: "lib.json", Alias: "lib_",
			Scenarios: []string{"agent-identity", "agent-identity-approval"},
		}},
	}
	got := resolveIn(t, dir, p)
	if len(got.Groups) != 1 || got.Groups[0].ID != "lib_identity" {
		t.Fatalf("imported groups = %+v, want only lib_identity", got.Groups)
	}
	v := got.ScenarioByID("lib_agent-identity-approval")
	if v == nil || v.VariantOf != "lib_agent-identity" || v.Group != "lib_identity" {
		t.Fatalf("imported variant = %+v", v)
	}
}
