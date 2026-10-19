package pidl

import (
	"reflect"
	"strings"
	"testing"
)

func flowActions(flows []Flow) []string {
	out := make([]string, len(flows))
	for i, f := range flows {
		if f.IsInclude() {
			out[i] = "include:" + f.Include
		} else {
			out[i] = f.Action
		}
	}
	return out
}

func entityIDs(entities []Entity) []string {
	out := make([]string, len(entities))
	for i, e := range entities {
		out[i] = e.ID
	}
	return out
}

func TestScenarioProjection(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		opts         ProjectOptions
		wantEntities []string
		wantActions  []string
		wantPhases   []string
	}{
		{
			name:         "two-party",
			id:           "agent-identity",
			wantEntities: []string{"agent", "resource"},
			wantActions:  []string{"signed_request", "response"},
			wantPhases:   []string{},
		},
		{
			name:         "collapsed include",
			id:           "three-party",
			wantEntities: []string{"person", "agent", "resource", "ps"},
			wantActions:  []string{"request", "auth_token_request", "include:interaction", "auth_token"},
			wantPhases:   []string{"consent"},
		},
		{
			name:         "expanded include",
			id:           "three-party",
			opts:         ProjectOptions{Expand: true},
			wantEntities: []string{"person", "agent", "resource", "ps"},
			wantActions:  []string{"request", "auth_token_request", "consent_prompt", "approve", "auth_token"},
			wantPhases:   []string{"consent"},
		},
		{
			name:         "fragment on its own",
			id:           "interaction",
			wantEntities: []string{"person", "ps"},
			wantActions:  []string{"consent_prompt", "approve"},
			wantPhases:   []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := scenarioProtocol()
			got, err := p.Scenario(tt.id, tt.opts)
			if err != nil {
				t.Fatalf("Scenario(%q): %v", tt.id, err)
			}
			if ids := entityIDs(got.Entities); !reflect.DeepEqual(ids, tt.wantEntities) {
				t.Errorf("entities = %v, want %v", ids, tt.wantEntities)
			}
			if acts := flowActions(got.Flows); !reflect.DeepEqual(acts, tt.wantActions) {
				t.Errorf("flows = %v, want %v", acts, tt.wantActions)
			}
			phases := []string{}
			for _, ph := range got.Phases {
				phases = append(phases, ph.ID)
			}
			if !reflect.DeepEqual(phases, tt.wantPhases) {
				t.Errorf("phases = %v, want %v", phases, tt.wantPhases)
			}
			if got.HasScenarios() {
				t.Errorf("projection still has %d scenarios", len(got.Scenarios))
			}
			if !strings.HasSuffix(got.ProtocolMeta.Name, " — "+p.ScenarioByID(tt.id).Name) {
				t.Errorf("name = %q, want scenario name suffix", got.ProtocolMeta.Name)
			}
			if tt.opts.Expand {
				if errs := got.Validate(); len(errs) > 0 {
					t.Errorf("expanded projection does not validate: %v", errs)
				}
			}
		})
	}
}

func TestScenarioProjectionInheritsIncludeFields(t *testing.T) {
	p := scenarioProtocol()
	p.Scenarios[2].Flows[2].Condition = "consent_required"
	p.Scenarios[1].Flows[1].Phase = "approval"
	p.Phases = append(p.Phases, Phase{ID: "approval", Name: "Approval", Parent: "consent"})

	got, err := p.Scenario("three-party", ProjectOptions{Expand: true})
	if err != nil {
		t.Fatalf("Scenario: %v", err)
	}
	prompt, approve := got.Flows[2], got.Flows[3]
	if prompt.Phase != "consent" || approve.Phase != "approval" {
		t.Errorf("phases = %q, %q; want consent (inherited), approval (own)", prompt.Phase, approve.Phase)
	}
	if prompt.Condition != "consent_required" || approve.Condition != "consent_required" {
		t.Errorf("conditions = %q, %q; want inherited consent_required", prompt.Condition, approve.Condition)
	}
	if len(got.Phases) != 2 || got.Phases[0].ID != "consent" || got.Phases[1].ID != "approval" {
		t.Errorf("phases = %+v, want consent then approval", got.Phases)
	}
}

func TestScenarioProjectionNestedAndAncestorPhases(t *testing.T) {
	p := scenarioProtocol()
	p.Phases = []Phase{
		{ID: "auth", Name: "Auth"},
		{ID: "unused", Name: "Unused"},
		{ID: "consent", Name: "Consent", Parent: "auth"},
	}
	p.Scenarios = append(p.Scenarios, Scenario{
		ID: "step-up", Name: "Step-up", Fragment: true,
		Flows: []Flow{
			{From: "resource", To: "agent", Action: "step_up_required"},
			{Include: "interaction"},
		},
	})
	p.Scenarios[2].Flows[2].Include = "step-up"

	got, err := p.Scenario("three-party", ProjectOptions{Expand: true})
	if err != nil {
		t.Fatalf("Scenario: %v", err)
	}
	want := []string{"request", "auth_token_request", "step_up_required", "consent_prompt", "approve", "auth_token"}
	if acts := flowActions(got.Flows); !reflect.DeepEqual(acts, want) {
		t.Errorf("flows = %v, want %v", acts, want)
	}
	if len(got.Phases) != 2 || got.Phases[0].ID != "auth" || got.Phases[1].ID != "consent" {
		t.Errorf("phases = %+v, want auth (ancestor) then consent", got.Phases)
	}
}

func TestScenarioProjectionErrors(t *testing.T) {
	tests := []struct {
		name    string
		p       func() *Protocol
		id      string
		opts    ProjectOptions
		wantErr string
	}{
		{"unknown id", scenarioProtocol, "nope", ProjectOptions{}, "unknown scenario \"nope\"; choose one of: agent-identity, interaction, three-party"},
		{"no id with scenarios", scenarioProtocol, "", ProjectOptions{}, "choose one of"},
		{"id on legacy document", func() *Protocol {
			p := scenarioProtocol()
			p.Flows, p.Scenarios = p.Scenarios[0].Flows, nil
			return p
		}, "x", ProjectOptions{}, "has no scenarios"},
		{"expand cycle", func() *Protocol {
			p := scenarioProtocol()
			p.Scenarios[1].Flows = append(p.Scenarios[1].Flows, Flow{Include: "interaction"})
			return p
		}, "three-party", ProjectOptions{Expand: true}, "include cycle: three-party -> interaction -> interaction"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.p().Scenario(tt.id, tt.opts)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Scenario(%q) error = %v, want %q", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestScenarioProjectionLegacy(t *testing.T) {
	p := scenarioProtocol()
	p.Flows, p.Scenarios = p.Scenarios[0].Flows, nil
	got, err := p.Scenario("", ProjectOptions{})
	if err != nil {
		t.Fatalf("Scenario(\"\"): %v", err)
	}
	if !reflect.DeepEqual(got.Flows, p.Flows) || !reflect.DeepEqual(got.Entities, p.Entities) || !reflect.DeepEqual(got.ProtocolMeta, p.ProtocolMeta) {
		t.Fatal("legacy projection is not a copy of the protocol")
	}
	got.Flows[0].Action = "changed"
	if p.Flows[0].Action == "changed" {
		t.Fatal("legacy projection shares the flows slice with the original")
	}
}

func TestScenarioProjectionDoesNotMutate(t *testing.T) {
	p := scenarioProtocol()
	before := scenarioProtocol()
	if _, err := p.Scenario("three-party", ProjectOptions{Expand: true}); err != nil {
		t.Fatalf("Scenario: %v", err)
	}
	if !reflect.DeepEqual(p, before) {
		t.Fatal("Scenario mutated the source protocol")
	}
}

func TestFlatten(t *testing.T) {
	p := scenarioProtocol()
	got := p.Flatten()
	want := []string{"signed_request", "response", "consent_prompt", "approve", "request", "auth_token_request", "auth_token"}
	if acts := flowActions(got.Flows); !reflect.DeepEqual(acts, want) {
		t.Fatalf("Flatten flows = %v, want %v", acts, want)
	}
	if got.HasScenarios() || !reflect.DeepEqual(got.Entities, p.Entities) {
		t.Fatal("Flatten should drop scenarios and keep all entities")
	}
	if errs := got.Validate(); len(errs) > 0 {
		t.Fatalf("flattened protocol does not validate: %v", errs)
	}

	legacy := scenarioProtocol()
	legacy.Flows, legacy.Scenarios = legacy.Scenarios[0].Flows, nil
	if !reflect.DeepEqual(legacy.Flatten().Flows, legacy.Flows) {
		t.Fatal("Flatten of a legacy document should copy its flows")
	}
}

func TestIncludeSpan(t *testing.T) {
	p := scenarioProtocol()
	proj, err := p.Scenario("three-party", ProjectOptions{})
	if err != nil {
		t.Fatalf("Scenario: %v", err)
	}
	if len(proj.Fragments) != 1 || proj.Fragments[0].ID != "interaction" {
		t.Fatalf("Fragments = %+v, want the interaction fragment", proj.Fragments)
	}
	label, span := proj.IncludeSpan(proj.Flows[2])
	if label != "Interaction" || !reflect.DeepEqual(span, []string{"person", "ps"}) {
		t.Fatalf("IncludeSpan = %q, %v; want Interaction, [person ps]", label, span)
	}

	f := proj.Flows[2]
	f.Label = "Consent"
	if label, _ := proj.IncludeSpan(f); label != "Consent" {
		t.Fatalf("IncludeSpan label = %q, want the step's own label", label)
	}
	if label, span := proj.IncludeSpan(Flow{Include: "missing"}); label != "missing" || span != nil {
		t.Fatalf("IncludeSpan(missing) = %q, %v; want missing, nil", label, span)
	}

	expanded, err := p.Scenario("three-party", ProjectOptions{Expand: true})
	if err != nil {
		t.Fatalf("Scenario: %v", err)
	}
	if expanded.Fragments != nil {
		t.Fatalf("expanded projection Fragments = %+v, want nil", expanded.Fragments)
	}
	if got := p.Flatten().Fragments; got != nil {
		t.Fatalf("Flatten Fragments = %+v, want nil", got)
	}
}
