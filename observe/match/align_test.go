package match

import (
	"strconv"
	"testing"

	"github.com/grokify/pidl"
)

// ex builds an exchange whose labels are participant IDs, as a resolver
// without a binding takes them.
func ex(from, to, method, url string, status int) pidl.Exchange {
	return pidl.Exchange{From: from, To: to,
		Request: pidl.ObservedRequest{Method: method, URL: url}, Response: pidl.ObservedResponse{Status: status}}
}

var (
	attest  = ex("agent", "ap", "POST", "https://ap.test/agent-token", 200)
	call    = ex("agent", "resource", "GET", "https://resource.test/api/documents", 401)
	poll    = ex("agent", "ps", "GET", "https://ps.test/pending/2U5_3Rqrgp", 202)
	polled  = ex("agent", "ps", "GET", "https://ps.test/pending/2U5_3Rqrgp", 200)
	confirm = ex("agent", "ps", "POST", "https://ps.test/confirm/77", 204)
	jwks    = ex("ps", "ap", "GET", "https://ap.test/jwks.json", 200)
)

func run(t *testing.T, b *pidl.Binding, xs ...pidl.Exchange) (Instance, []Step) {
	t.Helper()
	w, sc := demo(t)
	steps, _ := Prepare(w, sc, b)
	r, err := NewResolver(nil)
	if err != nil {
		t.Fatal(err)
	}
	resolved := make([]Resolved, len(xs))
	for i := range xs {
		resolved[i] = r.Resolve(i, &xs[i])
	}
	return Align(sc.ID, steps, resolved), steps
}

func result(t *testing.T, inst Instance, id string) StepResult {
	t.Helper()
	for _, r := range inst.Steps {
		if r.Step.Step.ID == id {
			return r
		}
	}
	t.Fatalf("no step %s", id)
	return StepResult{}
}

func TestAlignInOrder(t *testing.T) {
	inst, _ := run(t, nil, attest, call, polled, confirm)
	want := map[string]struct {
		status    Status
		strength  Strength
		exchanges string
	}{
		"agent-token.attest":      {StatusMatched, Strong, "0"},
		"agent-token.agent-token": {StatusMatched, Strong, "0"},
		"request":                 {StatusMatched, Strong, "1"},
		"person-token-required":   {StatusMatched, Strong, "1"},
		"direct-user":             {StatusUnobservable, "", ""},
		"approve":                 {StatusUnobservable, "", ""},
		"poll":                    {StatusMatched, Strong, "2"},
		"polled":                  {StatusMatched, Strong, "2"},
		"audit":                   {StatusUnmatchable, "", ""},
		"confirm":                 {StatusMatched, Strong, "3"},
	}
	for id, w := range want {
		r := result(t, inst, id)
		if r.Status != w.status || r.Strength != w.strength || joinInts(r.Exchanges) != w.exchanges {
			t.Errorf("%s = %s %s [%s], want %s %s [%s]", id, r.Status, r.Strength, joinInts(r.Exchanges), w.status, w.strength, w.exchanges)
		}
	}
	if len(inst.Extra) != 0 || inst.Explained != 4 {
		t.Errorf("extra = %v, explained = %d", inst.Extra, inst.Explained)
	}
	if inst.Matched != 7 || inst.Strong != 7 || inst.Weak != 0 || inst.Unobservable != 2 || inst.Unmatchable != 1 || inst.NotSeen != 0 {
		t.Errorf("counts = %+v", inst)
	}
}

func TestAlignWeakMatch(t *testing.T) {
	// The example's poll path has a literal ID and its status is 200; a
	// pending 202 at another ID agrees on participants and method only.
	inst, _ := run(t, nil, attest, call, poll, confirm)
	if r := result(t, inst, "poll"); r.Status != StatusMatched || r.Strength != Weak {
		t.Errorf("poll = %s %s, want a weak match", r.Status, r.Strength)
	}
	if inst.Weak != 2 {
		t.Errorf("weak = %d, want the poll and its response", inst.Weak)
	}
}

func TestAlignSwap(t *testing.T) {
	inst, _ := run(t, nil, call, attest, polled, confirm)
	a, c := result(t, inst, "agent-token.attest"), result(t, inst, "request")
	if a.Status == StatusMatched && c.Status == StatusMatched {
		t.Fatalf("both swapped steps matched: order was not kept")
	}
	if len(inst.Extra) != 1 {
		t.Errorf("extra = %v, want the exchange out of order", inst.Extra)
	}
}

func TestAlignDuplicateAndTie(t *testing.T) {
	inst, _ := run(t, nil, attest, attest, call, polled, confirm)
	if r := result(t, inst, "agent-token.attest"); joinInts(r.Exchanges) != "0" {
		t.Errorf("attest took exchange %s, want the earlier of two equal ones", joinInts(r.Exchanges))
	}
	if len(inst.Extra) != 1 || inst.Extra[0].Exchange != 1 {
		t.Errorf("extra = %v, want the duplicate", inst.Extra)
	}
}

func TestAlignMissingAndOptional(t *testing.T) {
	inst, _ := run(t, nil, attest, call, polled)
	if r := result(t, inst, "confirm"); r.Status != StatusNotSeen {
		t.Errorf("confirm = %s, want notSeen", r.Status)
	}
	b := &pidl.Binding{Format: pidl.BindingFormat, Steps: map[string]pidl.BindingStep{"*/confirm": {Match: pidl.StepMatch{Optional: true}}}}
	inst, _ = run(t, b, attest, call, polled)
	if r := result(t, inst, "confirm"); r.Status != StatusSkipped {
		t.Errorf("optional confirm = %s, want skipped", r.Status)
	}
	if inst.Skipped != 1 || inst.NotSeen != 0 {
		t.Errorf("counts = %+v", inst)
	}
}

func TestAlignRepeat(t *testing.T) {
	b := &pidl.Binding{Format: pidl.BindingFormat, Steps: map[string]pidl.BindingStep{
		"*/poll": {Match: pidl.StepMatch{Path: "/pending/{id}", Status: 0, Repeat: true}},
	}}
	inst, _ := run(t, b, attest, call, poll, jwks, poll, polled, confirm)
	r := result(t, inst, "poll")
	if r.Status != StatusMatched || joinInts(r.Exchanges) != "2,4,5" {
		t.Errorf("poll took [%s], want every poll before the next step, past the key fetch", joinInts(r.Exchanges))
	}
	if resp := result(t, inst, "polled"); joinInts(resp.Exchanges) != "2,4,5" {
		t.Errorf("response took [%s], want its request's exchanges", joinInts(resp.Exchanges))
	}
	if len(inst.Extra) != 1 || inst.Extra[0].Exchange != 3 || inst.Explained != 6 {
		t.Errorf("extra = %v, explained = %d; want the key fetch extra and six explained", inst.Extra, inst.Explained)
	}
	without, _ := run(t, nil, attest, call, poll, jwks, poll, polled, confirm)
	if r := result(t, without, "poll"); len(r.Exchanges) != 1 {
		t.Errorf("without repeat the poll took [%s]", joinInts(r.Exchanges))
	}
}

func TestAlignExtraPlacement(t *testing.T) {
	inst, steps := run(t, nil, jwks, attest, jwks, call, polled, confirm)
	if len(inst.Extra) != 2 {
		t.Fatalf("extra = %v", inst.Extra)
	}
	if inst.Extra[0].AfterStep != -1 {
		t.Errorf("first extra after step %d, want before the first step", inst.Extra[0].AfterStep)
	}
	if got := steps[inst.Extra[1].AfterStep].Step.ID; got != "agent-token.agent-token" {
		t.Errorf("second extra after %s, want after the agent token's response", got)
	}
}

func TestAlignEmpty(t *testing.T) {
	inst, _ := run(t, nil)
	// Seven not seen: four steps with rules and the three responses that
	// take their status.
	if inst.Matched != 0 || inst.NotSeen != 7 || inst.Explained != 0 || len(inst.Extra) != 0 {
		t.Errorf("empty run = %+v", inst)
	}
}

func TestScore(t *testing.T) {
	s := Step{Rule: pidl.StepMatch{Headers: map[string]string{"a": "b", "c": "d"}}}
	if score, strength := Score(Tests{Participants: Passed, Method: Passed, SenderKnown: true}, s); score != WeightSender || strength != Weak {
		t.Errorf("required tests only: %d %s", score, strength)
	}
	if score, strength := Score(Tests{Path: Passed, Status: Passed, Headers: Passed}, s); score != WeightPath+WeightStatus+2*WeightHeader || strength != Strong {
		t.Errorf("scored tests: %d %s", score, strength)
	}
}

func joinInts(xs []int) string {
	s := ""
	for i, x := range xs {
		if i > 0 {
			s += ","
		}
		s += strconv.Itoa(x)
	}
	return s
}
