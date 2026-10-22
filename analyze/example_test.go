package analyze_test

import (
	"reflect"
	"testing"

	"github.com/grokify/pidl/analyze"
	"github.com/grokify/pidl/examples"
)

// TestWorkedExample pins the worked example's shape and the findings it
// is meant to show: access tokens are bearer tokens, and the client
// presents an authorization code it received through the browser, the
// interception PKCE defends against.
func TestWorkedExample(t *testing.T) {
	p, err := examples.GetProtocol("oauth2_api_access")
	if err != nil {
		t.Fatalf("GetProtocol: %v", err)
	}
	if errs := p.Validate(); len(errs) > 0 {
		t.Fatalf("example does not validate: %v", errs)
	}
	if got := p.ScenarioIDs(false); !reflect.DeepEqual(got, []string{
		"authorization-code", "client-credentials", "scope-challenge", "async-export", "signed-request",
	}) {
		t.Errorf("main scenarios = %v", got)
	}
	if got := len(p.ScenarioIDs(true)) - len(p.ScenarioIDs(false)); got != 2 {
		t.Errorf("fragments = %d, want 2", got)
	}

	opts := analyze.DefaultAnalysisOptions()
	opts.MinSeverity = analyze.SeverityInfo
	var got []string
	for _, r := range analyze.Analyze(p, opts).Risks {
		got = append(got, r.ID+" "+r.Location)
	}
	want := []string{
		"SEC003 metadata.tokens[authorization_code]",
		"SEC003 metadata.tokens[access_token]",
		"SEC019 scenarios[0].flows[4] (authorization-code)",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("findings = %v, want %v", got, want)
	}
}
