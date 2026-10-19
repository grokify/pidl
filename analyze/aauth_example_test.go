package analyze_test

import (
	"reflect"
	"testing"

	"github.com/grokify/pidl/analyze"
	"github.com/grokify/pidl/examples"
)

// TestAAuthExample pins the AAuth example's shape and the finding it is
// meant to demonstrate: in federated access the PS signs the request to the
// AS, while the resource token it presents was delivered to the agent.
func TestAAuthExample(t *testing.T) {
	p, err := examples.GetProtocol("aauth_protocol")
	if err != nil {
		t.Fatalf("GetProtocol: %v", err)
	}
	if errs := p.Validate(); len(errs) > 0 {
		t.Fatalf("example does not validate: %v", errs)
	}
	if got := p.ScenarioIDs(false); !reflect.DeepEqual(got, []string{
		"agent-identity", "resource-managed", "person-identity", "ps-authorization", "federated-authorization",
	}) {
		t.Errorf("main scenarios = %v", got)
	}
	if got := len(p.ScenarioIDs(true)) - len(p.ScenarioIDs(false)); got != 3 {
		t.Errorf("fragments = %d, want 3", got)
	}

	opts := analyze.DefaultAnalysisOptions()
	opts.MinSeverity = analyze.SeverityInfo
	var signer []string
	for _, r := range analyze.Analyze(p, opts).Risks {
		switch r.ID {
		case "SEC019":
			signer = append(signer, r.Location)
		case "SEC017", "SEC018", "SEC020", "SEC021":
			t.Errorf("unexpected %s at %s: %s", r.ID, r.Location, r.Description)
		}
	}
	if !reflect.DeepEqual(signer, []string{"scenarios[4].flows[6] (federated-authorization)"}) {
		t.Errorf("SEC019 = %v, want the PS to AS request in federated access", signer)
	}
}
