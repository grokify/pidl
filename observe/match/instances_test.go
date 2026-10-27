package match

import (
	"slices"
	"testing"

	"github.com/grokify/pidl"
)

// petstore loads the AAuth document (aauth_protocol.json, from
// aistandardsio/agent-protocols), the sandbox capture of the Petstore run,
// and the binding for that deployment.
func petstore(t *testing.T) (*pidl.Walkthrough, *pidl.Observation, *pidl.Binding) {
	t.Helper()
	p, err := pidl.ParseFile("testdata/aauth_protocol.json")
	if err != nil {
		t.Fatal(err)
	}
	w := p.Walkthrough(nil)
	o, err := pidl.ReadObservation("../adapter/sandbox/testdata/petstore.observation.json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := pidl.ReadBinding("testdata/petstore.binding.json")
	if err != nil {
		t.Fatal(err)
	}
	if errs := b.Validate(); len(errs) > 0 {
		t.Fatalf("binding does not validate: %v", errs)
	}
	return &w, o, b
}

func scenarios(r *Report) []string {
	var ids []string
	for _, f := range r.Instances {
		ids = append(ids, f.Scenario)
	}
	return ids
}

// checkOrderFlow checks the order flow's instance: every observable step
// matched, the person's steps unobservable, and only the poll and its
// response weak (the poll's example path holds a literal identifier).
func checkOrderFlow(t *testing.T, f Found) {
	t.Helper()
	if f.Scenario != "ps-authorization-challenge" || f.RunnerUp != "ps-authorization" {
		t.Errorf("order flow = %s (runner-up %s)", f.Scenario, f.RunnerUp)
	}
	var weak []string
	for _, r := range f.Steps {
		switch r.Status {
		case StatusMatched:
			if r.Strength == Weak {
				weak = append(weak, r.Step.Step.ID)
			}
		case StatusUnobservable:
		default:
			t.Errorf("%s = %s", r.Step.Step.ID, r.Status)
		}
	}
	if !slices.Equal(weak, []string{"consent.poll", "auth-token"}) {
		t.Errorf("weak matches = %v, want the poll and its response", weak)
	}
	if f.Unobservable != 2 || f.NotSeen != 0 || f.Explained != 7 || len(f.Extra) != 7 {
		t.Errorf("counts: unobservable %d, notSeen %d, explained %d, extra %d", f.Unobservable, f.NotSeen, f.Explained, len(f.Extra))
	}
}

func TestMatchPetstoreCapture(t *testing.T) {
	w, o, b := petstore(t)
	r, err := Match(w, o, b, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := scenarios(r); !slices.Equal(got, []string{"agent-identity", "ps-authorization-challenge"}) {
		t.Fatalf("instances = %v, want browsing then ordering", got)
	}
	if !slices.Equal(r.Outside, []int{0}) || r.LeftOut != 0 {
		t.Errorf("outside = %v, left out = %d; want the health check outside any instance", r.Outside, r.LeftOut)
	}
	browse := r.Instances[0]
	if browse.Matched != 4 || browse.Strong != 4 || len(browse.Extra) != 2 {
		t.Errorf("browsing: matched %d, strong %d, extra %d", browse.Matched, browse.Strong, len(browse.Extra))
	}
	checkOrderFlow(t, r.Instances[1])
	var warned []string
	for _, wn := range r.Warnings {
		warned = append(warned, wn.Scenario+"/"+wn.Step)
	}
	if !slices.Equal(warned, []string{"resource-managed/poll", "consent/poll"}) {
		t.Errorf("warnings = %v, want each literal identifier once, where it is defined", warned)
	}
}

func TestMatchTimeWindow(t *testing.T) {
	w, o, b := petstore(t)
	r, err := Match(w, o, b, Options{From: o.Exchanges[1].Time})
	if err != nil {
		t.Fatal(err)
	}
	if r.LeftOut != 1 || len(r.Outside) != 0 {
		t.Errorf("left out = %d, outside = %v; want the health check left out", r.LeftOut, r.Outside)
	}
	if got := scenarios(r); !slices.Equal(got, []string{"agent-identity", "ps-authorization-challenge"}) {
		t.Errorf("instances = %v", got)
	}
}

func TestMatchNamedScenarios(t *testing.T) {
	w, o, b := petstore(t)
	r, err := Match(w, o, b, Options{Scenarios: []string{"agent-identity", "ps-authorization-challenge"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := scenarios(r); !slices.Equal(got, []string{"agent-identity", "ps-authorization-challenge"}) {
		t.Fatalf("instances = %v", got)
	}
	if r.Instances[1].RunnerUp != "" {
		t.Errorf("a named scenario has a runner-up: %s", r.Instances[1].RunnerUp)
	}

	// Only the authorization-endpoint scenario: the health check and the
	// browsing journey explain too little, so they fall outside, and the
	// order flow matches with its challenge-first steps as extra.
	r, err = Match(w, o, b, Options{Scenarios: []string{"ps-authorization"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := scenarios(r); !slices.Equal(got, []string{"ps-authorization"}) {
		t.Fatalf("instances = %v", got)
	}
	if !slices.Contains(r.Outside, 0) || r.Instances[0].Weak == 0 {
		t.Errorf("outside = %v, weak = %d", r.Outside, r.Instances[0].Weak)
	}

	if _, err := Match(w, o, b, Options{Scenarios: []string{"nope"}}); err == nil {
		t.Error("an unknown scenario was accepted")
	}
}

func TestMatchTwoJourneys(t *testing.T) {
	w, _ := demo(t)
	o := &pidl.Observation{Format: pidl.ObservationFormat, Source: pidl.ObservationSource{Tool: "test"},
		Exchanges: []pidl.Exchange{attest, call, polled, confirm, attest, call, polled}}
	r, err := Match(w, o, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := scenarios(r); !slices.Equal(got, []string{"challenge", "challenge"}) {
		t.Fatalf("instances = %v, want two journeys", got)
	}
	if r.Instances[1].NotSeen != 1 || r.Instances[1].Steps[9].Status != StatusNotSeen {
		t.Errorf("second journey lacks its confirm: notSeen = %d", r.Instances[1].NotSeen)
	}
}

func TestRankLess(t *testing.T) {
	tests := []struct {
		a, b Rank
		want bool
	}{
		{Rank{Net: 1}, Rank{Net: 2}, true},
		{Rank{Net: 2, Strong: 1}, Rank{Net: 2, Strong: 3}, true},
		{Rank{Net: 2, Strong: 3, Score: 9}, Rank{Net: 2, Strong: 3, Score: 10}, true},
		{Rank{Net: 3}, Rank{Net: 2, Strong: 9, Score: 99}, false},
		{Rank{Net: 2}, Rank{Net: 2}, false},
	}
	for _, tt := range tests {
		if got := tt.a.Less(tt.b); got != tt.want {
			t.Errorf("%+v.Less(%+v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
