package pidl

import (
	"reflect"
	"strings"
	"testing"
)

// scenarioProtocol returns a valid protocol with two main scenarios and
// one fragment.
func scenarioProtocol() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "aauth", Name: "AAuth"},
		Entities: []Entity{
			{ID: "person", Name: "Person", Type: EntityTypeUser},
			{ID: "agent", Name: "Agent", Type: EntityTypeAgent},
			{ID: "resource", Name: "Resource", Type: EntityTypeResourceServer},
			{ID: "ps", Name: "Person Server", Type: EntityTypeAuthorizationServer},
		},
		Phases: []Phase{{ID: "consent", Name: "Consent"}},
		Scenarios: []Scenario{
			{
				ID: "agent-identity", Name: "Agent identity",
				Flows: []Flow{
					{From: "agent", To: "resource", Action: "signed_request"},
					{From: "resource", To: "agent", Action: "response", Mode: FlowModeResponse},
				},
			},
			{
				ID: "interaction", Name: "Interaction", Fragment: true,
				Flows: []Flow{
					{From: "ps", To: "person", Action: "consent_prompt"},
					{From: "person", To: "ps", Action: "approve"},
				},
			},
			{
				ID: "three-party", Name: "PS authorization",
				Participants: []string{"person", "agent", "resource", "ps"},
				Flows: []Flow{
					{From: "agent", To: "resource", Action: "request"},
					{From: "agent", To: "ps", Action: "auth_token_request"},
					{Include: "interaction", Phase: "consent"},
					{From: "ps", To: "agent", Action: "auth_token", Mode: FlowModeResponse},
				},
			},
		},
	}
}

func TestValidateScenariosValid(t *testing.T) {
	if errs := scenarioProtocol().Validate(); len(errs) > 0 {
		t.Fatalf("Validate() = %v, want no errors", errs)
	}
}

func TestValidateScenariosErrors(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(p *Protocol)
		wantField string
		wantMsg   string
	}{
		{
			name:      "flows and scenarios both set",
			mutate:    func(p *Protocol) { p.Flows = []Flow{{From: "agent", To: "resource", Action: "x"}} },
			wantField: "flows",
			wantMsg:   "must be empty when scenarios are present",
		},
		{
			name:      "duplicate id",
			mutate:    func(p *Protocol) { p.Scenarios[2].ID = "agent-identity" },
			wantField: "scenarios[2].id",
			wantMsg:   "duplicate scenario ID",
		},
		{
			name:      "bad id pattern",
			mutate:    func(p *Protocol) { p.Scenarios[0].ID = "Agent Identity" },
			wantField: "scenarios[0].id",
			wantMsg:   "must match pattern",
		},
		{
			name:      "missing name",
			mutate:    func(p *Protocol) { p.Scenarios[0].Name = "" },
			wantField: "scenarios[0].name",
			wantMsg:   "required",
		},
		{
			name:      "empty flows",
			mutate:    func(p *Protocol) { p.Scenarios[0].Flows = nil },
			wantField: "scenarios[0].flows",
			wantMsg:   "must have at least 1 flow",
		},
		{
			name: "only fragments",
			mutate: func(p *Protocol) {
				for i := range p.Scenarios {
					p.Scenarios[i].Fragment = true
				}
			},
			wantField: "scenarios",
			wantMsg:   "not a fragment",
		},
		{
			name:      "participants out of order",
			mutate:    func(p *Protocol) { p.Scenarios[2].Participants = []string{"agent", "person", "resource", "ps"} },
			wantField: "scenarios[2].participants",
			wantMsg:   "must follow entities order: [person, agent, resource, ps]",
		},
		{
			name:      "unknown participant",
			mutate:    func(p *Protocol) { p.Scenarios[2].Participants = []string{"person", "agent", "resource", "ps", "as"} },
			wantField: "scenarios[2].participants[4]",
			wantMsg:   `unknown entity "as"`,
		},
		{
			name: "duplicate participant",
			mutate: func(p *Protocol) {
				p.Scenarios[2].Participants = []string{"person", "agent", "agent", "resource", "ps"}
			},
			wantField: "scenarios[2].participants[2]",
			wantMsg:   "duplicate participant",
		},
		{
			name:      "participants miss a fragment entity",
			mutate:    func(p *Protocol) { p.Scenarios[2].Participants = []string{"agent", "resource", "ps"} },
			wantField: "scenarios[2].participants",
			wantMsg:   "missing entities used by the scenario: [person]",
		},
		{
			name:      "include unknown scenario",
			mutate:    func(p *Protocol) { p.Scenarios[2].Flows[2].Include = "nope" },
			wantField: "scenarios[2].flows[2].include",
			wantMsg:   `unknown scenario "nope"`,
		},
		{
			name:      "include a main scenario",
			mutate:    func(p *Protocol) { p.Scenarios[2].Flows[2].Include = "agent-identity" },
			wantField: "scenarios[2].flows[2].include",
			wantMsg:   "is not a fragment",
		},
		{
			name: "include step with message fields",
			mutate: func(p *Protocol) {
				f := &p.Scenarios[2].Flows[2]
				f.From, f.Action = "agent", "x"
			},
			wantField: "scenarios[2].flows[2]",
			wantMsg:   "include step cannot set from, action",
		},
		{
			name:      "include step unknown phase",
			mutate:    func(p *Protocol) { p.Scenarios[2].Flows[2].Phase = "nope" },
			wantField: "scenarios[2].flows[2].phase",
			wantMsg:   `unknown phase "nope"`,
		},
		{
			name: "include cycle",
			mutate: func(p *Protocol) {
				p.Scenarios[1].Flows = append(p.Scenarios[1].Flows, Flow{Include: "interaction"})
			},
			wantField: "scenarios",
			wantMsg:   "include cycle: interaction -> interaction",
		},
		{
			name:      "message flow error path in scenario",
			mutate:    func(p *Protocol) { p.Scenarios[0].Flows[1].To = "nobody" },
			wantField: "scenarios[0].flows[1].to",
			wantMsg:   `unknown entity "nobody"`,
		},
		{
			name: "state mutation path in scenario",
			mutate: func(p *Protocol) {
				p.Scenarios[0].Flows[0].Sets = []StateMutation{{Entity: "agent", To: "x"}}
			},
			wantField: "scenarios[0].flows[0].sets[0].entity",
			wantMsg:   "has no states defined",
		},
		{
			name: "security path in scenario",
			mutate: func(p *Protocol) {
				p.Scenarios[0].Flows[0].Security = &FlowSecurity{Token: "missing"}
			},
			wantField: "scenarios[0].flows[0].security.token",
			wantMsg:   "unknown token definition",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := scenarioProtocol()
			tt.mutate(p)
			errs := p.Validate()
			for _, e := range errs {
				if e.Field == tt.wantField && strings.Contains(e.Message, tt.wantMsg) {
					return
				}
			}
			t.Fatalf("Validate() = %v, want %s: ...%s...", errs, tt.wantField, tt.wantMsg)
		})
	}
}

func TestValidateIncludeCycleAcrossFragments(t *testing.T) {
	p := scenarioProtocol()
	p.Scenarios = append(p.Scenarios, Scenario{
		ID: "step-up", Name: "Step-up", Fragment: true,
		Flows: []Flow{{Include: "interaction"}},
	})
	p.Scenarios[1].Flows = append(p.Scenarios[1].Flows, Flow{Include: "step-up"})
	var cycles []string
	for _, e := range p.Validate() {
		if strings.HasPrefix(e.Message, "include cycle") {
			cycles = append(cycles, e.Message)
		}
	}
	want := []string{"include cycle: interaction -> step-up -> interaction"}
	if !reflect.DeepEqual(cycles, want) {
		t.Fatalf("cycle errors = %v, want %v", cycles, want)
	}
}

func TestValidateIncludeInLegacyFlows(t *testing.T) {
	p := scenarioProtocol()
	p.Flows = []Flow{
		{From: "agent", To: "resource", Action: "x"},
		{Include: "interaction"},
	}
	p.Scenarios = nil
	for _, e := range p.Validate() {
		if e.Field == "flows[1].include" && strings.Contains(e.Message, "only allowed in scenarios") {
			return
		}
	}
	t.Fatalf("Validate() did not reject include step in top-level flows")
}

func TestScenarioParticipants(t *testing.T) {
	p := scenarioProtocol()
	tests := []struct {
		id   string
		want []string
	}{
		{"agent-identity", []string{"agent", "resource"}},
		{"interaction", []string{"person", "ps"}},
		{"three-party", []string{"person", "agent", "resource", "ps"}},
		{"missing", nil},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := p.ScenarioParticipants(tt.id); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ScenarioParticipants(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}

	// Without declared participants, includes contribute fragment entities.
	p.Scenarios[2].Participants = nil
	want := []string{"person", "agent", "resource", "ps"}
	if got := p.ScenarioParticipants("three-party"); !reflect.DeepEqual(got, want) {
		t.Fatalf("derived ScenarioParticipants = %v, want %v", got, want)
	}
}
