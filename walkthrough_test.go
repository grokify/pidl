package pidl

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func stepLabels(steps []WalkthroughStep) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = s.Label
	}
	return out
}

func TestWalkthroughExpandsFragments(t *testing.T) {
	p := exampleProtocol()
	w := p.Walkthrough(nil)

	var ids []string
	for _, s := range w.Scenarios {
		ids = append(ids, s.ID)
	}
	if !reflect.DeepEqual(ids, []string{"agent-identity", "three-party", "interaction"}) {
		t.Fatalf("scenarios = %v", ids)
	}

	tp := w.Scenarios[1]
	if got := stepLabels(tp.Steps); !reflect.DeepEqual(got, []string{"request", "auth_token_request", "consent_prompt", "approve", "auth_token"}) {
		t.Fatalf("steps = %v", got)
	}
	for i, s := range tp.Steps {
		if s.Number != i+1 {
			t.Errorf("step %d numbered %d", i, s.Number)
		}
	}
	consent := tp.Steps[2]
	if !reflect.DeepEqual(consent.Fragments, []string{"Interaction"}) || consent.Phase != "consent" {
		t.Errorf("fragment step = %+v, want trail [Interaction] and inherited phase", consent)
	}
	if consent.Location != "scenarios[1].flows[0] (interaction)" {
		t.Errorf("fragment step location = %q", consent.Location)
	}
	if tp.Steps[1].Example == nil || tp.Steps[1].Example.Source != "draft-11 §7.2.1" {
		t.Errorf("example not carried: %+v", tp.Steps[1].Example)
	}
	if !reflect.DeepEqual(tp.Steps[1].Verifies, []string{"resource_token"}) || tp.Steps[1].SignedBy != "agent" {
		t.Errorf("security not carried: %+v", tp.Steps[1])
	}
	if frag := w.Scenarios[2]; !frag.Fragment || frag.Steps[0].Fragments != nil {
		t.Errorf("fragment scenario = %+v, want no trail on its own steps", frag)
	}
	if len(w.Tokens) != 2 || len(w.Tokens[0].Claims) != 2 {
		t.Errorf("tokens = %+v", w.Tokens)
	}
}

func TestWalkthroughAttachesFindings(t *testing.T) {
	p := exampleProtocol()
	findings := []Finding{
		{ID: "A", Location: "scenarios[1].flows[0] (interaction)"},
		{ID: "B", Location: "scenarios[2].flows[1] (three-party)"},
		{ID: "C", Location: "entities[0]"},
	}
	w := p.Walkthrough(findings)

	// A is inside the fragment, so it appears where the fragment is
	// included and in the fragment's own walkthrough.
	if got := w.Scenarios[1].Steps[2].Findings; len(got) != 1 || got[0].ID != "A" {
		t.Errorf("included fragment step findings = %+v", got)
	}
	if got := w.Scenarios[2].Steps[0].Findings; len(got) != 1 || got[0].ID != "A" {
		t.Errorf("fragment scenario step findings = %+v", got)
	}
	if got := w.Scenarios[1].Steps[1].Findings; len(got) != 1 || got[0].ID != "B" {
		t.Errorf("step findings = %+v", got)
	}
	if len(w.Findings) != 1 || w.Findings[0].ID != "C" {
		t.Errorf("general findings = %+v", w.Findings)
	}
}

func TestWalkthroughLegacy(t *testing.T) {
	p := exampleProtocol()
	p.Flows, p.Scenarios = p.Scenarios[0].Flows, nil
	w := p.Walkthrough([]Finding{{ID: "X", Location: "flows[1]"}})
	if len(w.Scenarios) != 1 || len(w.Scenarios[0].Steps) != 2 {
		t.Fatalf("legacy walkthrough = %+v", w.Scenarios)
	}
	if s := w.Scenarios[0].Steps[1]; s.Location != "flows[1]" || len(s.Findings) != 1 {
		t.Errorf("legacy step = %+v", s)
	}
}

func TestWalkthroughFormat(t *testing.T) {
	w := exampleProtocol().Walkthrough(nil)
	if w.Format != WalkthroughFormat {
		t.Fatalf("format = %q, want %q", w.Format, WalkthroughFormat)
	}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), `{"format":"pidl-walkthrough/v1",`) {
		t.Errorf("format is not the first field: %.60s", data)
	}
}
