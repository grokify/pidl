package analyze

import (
	"reflect"
	"sort"
	"testing"

	"github.com/grokify/pidl"
)

// fourParty models AAuth-style federated authorization: the PS signs the
// request to the AS on the agent's behalf.
func fourParty() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "four-party", Name: "Four party"},
		Entities: []pidl.Entity{
			{ID: "agent", Name: "Agent", Type: pidl.EntityTypeAgent},
			{ID: "resource", Name: "Resource", Type: pidl.EntityTypeResourceServer},
			{ID: "ps", Name: "PS", Type: pidl.EntityTypeAuthorizationServer},
			{ID: "as", Name: "AS", Type: pidl.EntityTypeAuthorizationServer},
			{ID: "auditor", Name: "Auditor", Type: pidl.EntityTypeOther},
		},
		Metadata: &pidl.ProtocolMetadata{Tokens: []pidl.TokenDefinition{
			{ID: "resource_token", Issuer: "resource", VerifiedBy: []string{"ps", "as"}},
			{ID: "auth_token", Issuer: "as", VerifiedBy: []string{"resource"}},
		}},
		Scenarios: []pidl.Scenario{
			{ID: "federated", Name: "Federated", Flows: []pidl.Flow{
				{From: "agent", To: "resource", Action: "request"},
				{From: "resource", To: "agent", Action: "challenge", Mode: pidl.FlowModeResponse,
					Security: &pidl.FlowSecurity{Issues: []string{"resource_token"}}},
				{Include: "exchange"},
				{From: "ps", To: "agent", Action: "deliver_auth_token", Mode: pidl.FlowModeResponse,
					Security: &pidl.FlowSecurity{Issues: []string{"auth_token"}}},
				{From: "agent", To: "resource", Action: "retry",
					Security: &pidl.FlowSecurity{Verifies: []string{"auth_token"}}},
			}},
			{ID: "exchange", Name: "PS to AS", Fragment: true, Flows: []pidl.Flow{
				{From: "ps", To: "as", Action: "token_request",
					Security: &pidl.FlowSecurity{SignedBy: "ps", Verifies: []string{"resource_token"}}},
				{From: "as", To: "ps", Action: "auth_token", Mode: pidl.FlowModeResponse,
					Security: &pidl.FlowSecurity{Issues: []string{"auth_token"}}},
			}},
			{ID: "unused", Name: "Unused", Fragment: true, Flows: []pidl.Flow{
				{From: "agent", To: "ps", Action: "ping"},
			}},
		},
	}
}

func scenarioRiskIDs(t *testing.T, p *pidl.Protocol) map[string][]string {
	t.Helper()
	opts := DefaultAnalysisOptions()
	opts.MinSeverity = SeverityInfo
	got := map[string][]string{}
	for _, r := range Analyze(p, opts).Risks {
		switch r.ID {
		case "SEC017", "SEC018", "SEC019", "SEC020", "SEC021":
			got[r.ID] = append(got[r.ID], r.Location)
		}
	}
	for _, locs := range got {
		sort.Strings(locs)
	}
	return got
}

func TestScenarioRules(t *testing.T) {
	p := fourParty()
	if errs := p.Validate(); len(errs) > 0 {
		t.Fatalf("fixture does not validate: %v", errs)
	}
	want := map[string][]string{
		"SEC017": {"entities[4]"},
		"SEC019": {"scenarios[1].flows[0] (exchange)"},
		"SEC021": {"scenarios[2]"},
	}
	if got := scenarioRiskIDs(t, p); !reflect.DeepEqual(got, want) {
		t.Fatalf("risks = %v\nwant    %v", got, want)
	}
}

func TestTokenCheckedBeforeIssued(t *testing.T) {
	p := fourParty()
	// Check the auth token before the fragment that issues it.
	fl := p.Scenarios[0].Flows
	retry := fl[4]
	p.Scenarios[0].Flows = []pidl.Flow{fl[0], fl[1], retry, fl[2], fl[3]}
	if got := scenarioRiskIDs(t, p)["SEC018"]; !reflect.DeepEqual(got, []string{"scenarios[0].flows[2] (federated)"}) {
		t.Fatalf("SEC018 = %v", got)
	}

	// A token that no flow issues comes from outside and is not reported.
	p = fourParty()
	p.Scenarios[0].Flows[1].Security = nil
	if got := scenarioRiskIDs(t, p)["SEC018"]; got != nil {
		t.Fatalf("SEC018 for an externally issued token = %v", got)
	}
}

func TestSignerDiffersFromSender(t *testing.T) {
	p := fourParty()
	p.Scenarios[1].Flows[0].Security.SignedBy = ""
	p.Scenarios[0].Flows[0].Security = &pidl.FlowSecurity{SignedBy: "ps"}
	got := scenarioRiskIDs(t, p)["SEC019"]
	want := []string{"scenarios[0].flows[0] (federated)", "scenarios[1].flows[0] (exchange)"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SEC019 = %v, want %v", got, want)
	}
}

func TestVerifierMismatchAndAnnotatedSigner(t *testing.T) {
	p := fourParty()
	p.Metadata.Tokens[0].VerifiedBy = []string{"ps"}
	p.Scenarios[1].Flows[0].Annotations = []pidl.Annotation{{Type: pidl.AnnotationTypeSecurity, Text: "PS signs for the agent"}}
	got := scenarioRiskIDs(t, p)
	if !reflect.DeepEqual(got["SEC020"], []string{"scenarios[1].flows[0] (exchange)"}) {
		t.Fatalf("SEC020 = %v", got["SEC020"])
	}
	if got["SEC019"] != nil {
		t.Fatalf("SEC019 should skip an annotated flow, got %v", got["SEC019"])
	}
}

func TestScenarioRulesOnSingleFlowList(t *testing.T) {
	p := fourParty()
	proj, err := p.Scenario("federated", pidl.ProjectOptions{Expand: true})
	if err != nil {
		t.Fatalf("Scenario: %v", err)
	}
	got := scenarioRiskIDs(t, proj)
	if !reflect.DeepEqual(got["SEC019"], []string{"flows[2]"}) {
		t.Fatalf("SEC019 on projection = %v", got["SEC019"])
	}
	if got["SEC021"] != nil || got["SEC018"] != nil {
		t.Fatalf("unexpected risks on projection: %v", got)
	}
}

func TestFlowRulesSeeScenarioFlows(t *testing.T) {
	p := fourParty()
	p.Entities[0].TrustLevel = pidl.TrustLevelUntrusted
	p.Entities[1].TrustLevel = pidl.TrustLevelTrusted
	opts := DefaultAnalysisOptions()
	opts.MinSeverity = SeverityInfo
	var locs []string
	for _, r := range Analyze(p, opts).Risks {
		if r.ID == "SEC001" {
			locs = append(locs, r.Location)
		}
	}
	if !reflect.DeepEqual(locs, []string{"scenarios[0].flows[0] (federated)", "scenarios[0].flows[3] (federated)", "scenarios[0].flows[4] (federated)"}) {
		t.Fatalf("SEC001 locations = %v, want written scenario locations", locs)
	}
}
