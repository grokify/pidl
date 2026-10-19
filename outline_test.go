package pidl

import (
	"reflect"
	"testing"
)

func TestOutlineScenarios(t *testing.T) {
	o := scenarioProtocol().Outline()

	var roles []string
	for _, r := range o.Roles {
		roles = append(roles, r.ID)
	}
	if !reflect.DeepEqual(roles, []string{"person", "agent", "resource", "ps"}) {
		t.Errorf("roles = %v, want entities order", roles)
	}

	var scenarios []string
	for _, s := range o.Scenarios {
		scenarios = append(scenarios, s.ID)
	}
	if !reflect.DeepEqual(scenarios, []string{"agent-identity", "three-party", "interaction"}) {
		t.Errorf("scenarios = %v, want main scenarios then fragments", scenarios)
	}

	inc := o.Scenarios[1].Steps[2]
	if inc.Include != "interaction" || inc.IncludePosition != 3 || inc.Label != "Interaction" || inc.Number != 3 {
		t.Errorf("include step = %+v, want interaction at position 3", inc)
	}
	resp := o.Scenarios[0].Steps[1]
	if !resp.Response || resp.From != "resource" || resp.To != "agent" {
		t.Errorf("response step = %+v", resp)
	}
}

func TestOutlineLegacy(t *testing.T) {
	p := scenarioProtocol()
	p.Flows, p.Scenarios = p.Scenarios[0].Flows, nil
	o := p.Outline()
	if len(o.Scenarios) != 1 {
		t.Fatalf("legacy outline has %d scenarios, want 1", len(o.Scenarios))
	}
	s := o.Scenarios[0]
	if s.ID != "aauth" || s.Name != "AAuth" || !reflect.DeepEqual(s.Participants, []string{"agent", "resource"}) || len(s.Steps) != 2 {
		t.Errorf("legacy scenario = %+v", s)
	}
	if o.Tokens != nil {
		t.Errorf("tokens = %+v, want none", o.Tokens)
	}
}

func TestOutlineTokens(t *testing.T) {
	o := verificationProtocol().Outline()
	if len(o.Tokens) != 2 || o.Tokens[0].Name != "resource_token" || !reflect.DeepEqual(o.Tokens[0].VerifiedBy, []string{"ps"}) {
		t.Errorf("tokens = %+v", o.Tokens)
	}
}
