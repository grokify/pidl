package match

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

func demo(t *testing.T) (*pidl.Walkthrough, *pidl.WalkthroughScenario) {
	t.Helper()
	p, err := pidl.ParseFile("testdata/demo.pidl.json")
	if err != nil {
		t.Fatal(err)
	}
	if errs := p.Validate(); len(errs) > 0 {
		t.Fatalf("demo document does not validate: %v", errs)
	}
	w := p.Walkthrough(nil)
	for i := range w.Scenarios {
		if w.Scenarios[i].ID == "challenge" {
			return &w, &w.Scenarios[i]
		}
	}
	t.Fatal("no challenge scenario")
	return nil, nil
}

func byID(steps []Step, id string) Step {
	for _, s := range steps {
		if s.Step.ID == id {
			return s
		}
	}
	return Step{Index: -1}
}

func TestPrepareKinds(t *testing.T) {
	w, sc := demo(t)
	steps, _ := Prepare(w, sc, nil)
	want := map[string]Kind{
		"agent-token.attest":      KindRequest,
		"agent-token.agent-token": KindResponse,
		"request":                 KindRequest,
		"person-token-required":   KindResponse,
		"direct-user":             KindUnobservable,
		"approve":                 KindUnobservable,
		"poll":                    KindRequest,
		"polled":                  KindResponse,
		"audit":                   KindStandalone,
		"confirm":                 KindStandalone,
	}
	if len(steps) != len(want) {
		t.Fatalf("steps = %d, want %d", len(steps), len(want))
	}
	for _, s := range steps {
		if s.Kind != want[s.Step.ID] {
			t.Errorf("%s: kind = %s, want %s", s.Step.ID, s.Kind, want[s.Step.ID])
		}
	}
	if r := byID(steps, "agent-token.agent-token"); r.PairedWith != byID(steps, "agent-token.attest").Index {
		t.Errorf("response paired with %d", r.PairedWith)
	}
}

func TestPrepareRules(t *testing.T) {
	w, sc := demo(t)
	steps, warnings := Prepare(w, sc, nil)

	attest := byID(steps, "agent-token.attest")
	if !attest.HasRule || attest.Rule.Method != "POST" || attest.Rule.Path != "/agent-token" || attest.Rule.Status != 200 {
		t.Errorf("example defaults: %+v", attest.Rule)
	}
	if attest.Step.DefinedIn != "agent-token" || attest.Step.FlowID != "attest" {
		t.Errorf("fragment step defined in %q as %q", attest.Step.DefinedIn, attest.Step.FlowID)
	}
	if audit := byID(steps, "audit"); audit.HasRule {
		t.Errorf("a standalone step without a rule has one: %+v", audit.Rule)
	}
	confirm := byID(steps, "confirm")
	if !confirm.HasRule || confirm.Rule.Path != "/confirm/{id}" || strings.Join(confirm.RuleSources, ",") != "document" {
		t.Errorf("document rule: %+v from %v", confirm.Rule, confirm.RuleSources)
	}
	if len(warnings) != 1 || warnings[0].Step != "poll" || !strings.Contains(warnings[0].Message, "/pending/{id}") {
		t.Errorf("warnings = %v, want one for the poll's literal identifier", warnings)
	}
}

func TestPrepareBinding(t *testing.T) {
	w, sc := demo(t)
	b := &pidl.Binding{
		Format:       pidl.BindingFormat,
		Participants: map[string]pidl.BindingParticipant{"ap": {Hosts: []string{"ap.example.test"}}},
		Paths:        map[string]map[string]string{"resource": {"*/request": "/v2/store/order"}},
		Unobservable: []string{"ap"},
		Steps: map[string]pidl.BindingStep{
			"challenge/request": {Match: pidl.StepMatch{Method: "POST", Status: 401}},
			"*/poll":            {Match: pidl.StepMatch{Path: "^/pending/", Repeat: true}},
			"challenge/audit":   {Match: pidl.StepMatch{Method: "POST", Path: "/audit"}},
		},
	}
	steps, warnings := Prepare(w, sc, b)
	req := byID(steps, "request")
	if req.Rule.Method != "POST" || req.Rule.Status != 401 || req.Rule.Path != "/v2/store/order" {
		t.Errorf("binding over example: %+v", req.Rule)
	}
	if got := strings.Join(req.RuleSources, ","); got != "example,binding challenge/request,binding path */request" {
		t.Errorf("rule sources = %s", got)
	}
	poll := byID(steps, "poll")
	if poll.Rule.Path != "^/pending/" || !poll.Rule.Repeat || poll.Rule.Method != "GET" {
		t.Errorf("binding path pattern: %+v", poll.Rule)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v; the binding templated the poll's path", warnings)
	}
	if audit := byID(steps, "audit"); !audit.HasRule || audit.Kind != KindStandalone {
		t.Errorf("binding rule for a standalone step: %+v", audit)
	}
	if attest := byID(steps, "agent-token.attest"); attest.Kind != KindUnobservable {
		t.Errorf("binding unobservable participant: kind = %s", attest.Kind)
	}
}

func TestResolve(t *testing.T) {
	b := &pidl.Binding{
		Format: pidl.BindingFormat,
		Participants: map[string]pidl.BindingParticipant{
			"agent":    {Labels: []string{"agent"}},
			"ap":       {Labels: []string{"ap"}, Hosts: []string{"ap.example.test"}},
			"resource": {Labels: []string{"petstore.example.test"}, Hosts: []string{"petstore.example.test"}},
		},
		LabelsByPath: []pidl.BindingLabelRule{{Label: "client", Path: `^/(\.well-known/.*|jwks\.json)$`, Participant: "resource"}},
		Rewrite:      []pidl.BindingRewrite{{Match: `^http://sandbox:8099/(https://.*)$`, To: "$1"}},
	}
	r, err := NewResolver(b)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, from, to, url        string
		wantFrom, wantTo, wantPath string
	}{
		{"labels", "agent", "ap", "https://ap.example.test/agent-token", "agent", "ap", "/agent-token"},
		{"host when the label is unknown", "agent", "edge", "https://petstore.example.test/v2/store/order", "agent", "resource", "/v2/store/order"},
		{"sender by path", "client", "ap", "https://ap.example.test/jwks.json", "resource", "ap", "/jwks.json"},
		{"unresolved sender", "client", "ap", "https://ap.example.test/agent-token", "", "ap", "/agent-token"},
		{"rewrite first", "", "", "http://sandbox:8099/https://ap.example.test/agent-token", "", "ap", "/agent-token"},
		{"unknown receiver", "agent", "", "https://elsewhere.test/x", "agent", "", "/x"},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &pidl.Exchange{From: tt.from, To: tt.to, Request: pidl.ObservedRequest{Method: "GET", URL: tt.url}}
			got := r.Resolve(i, e)
			if got.From != tt.wantFrom || got.To != tt.wantTo || got.Path != tt.wantPath || got.Index != i {
				t.Errorf("Resolve = from %q to %q path %q; want %q %q %q", got.From, got.To, got.Path, tt.wantFrom, tt.wantTo, tt.wantPath)
			}
		})
	}
	if _, err := NewResolver(&pidl.Binding{Rewrite: []pidl.BindingRewrite{{Match: "("}}}); err == nil {
		t.Error("an invalid rewrite compiled")
	}
	plain, err := NewResolver(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := plain.Resolve(0, &pidl.Exchange{From: "agent", To: "ap", Request: pidl.ObservedRequest{URL: "https://x.test/"}}); got.From != "agent" || got.To != "ap" {
		t.Errorf("without a binding, labels are participants: %+v", got)
	}
}

func TestTest(t *testing.T) {
	step := Step{Step: &pidl.WalkthroughStep{From: "agent", To: "resource"}, Rule: pidl.StepMatch{
		Method: "POST", Path: "/v2/store/order", Status: 401,
		Headers: map[string]string{"AAuth-Requirement": "requirement=person-token*"},
	}}
	ex := func(method string, status int, hdr string) *pidl.Exchange {
		return &pidl.Exchange{
			Request:  pidl.ObservedRequest{Method: method},
			Response: pidl.ObservedResponse{Status: status, Headers: pidl.ObservedHeaders{{"Aauth-Requirement", hdr}}},
		}
	}
	tests := []struct {
		name      string
		r         Resolved
		candidate bool
		want      Tests
	}{
		{"all pass", Resolved{From: "agent", To: "resource", Path: "/v2/store/order", Exchange: ex("POST", 401, "requirement=person-token")}, true,
			Tests{Participants: Passed, Method: Passed, SenderKnown: true, Path: Passed, Status: Passed, Headers: Passed}},
		{"unknown sender agrees", Resolved{To: "resource", Path: "/v2/store/order", Exchange: ex("post", 401, "requirement=person-token")}, true,
			Tests{Participants: Passed, Method: Passed, Path: Passed, Status: Passed, Headers: Passed}},
		{"scored tests fail", Resolved{From: "agent", To: "resource", Path: "/v2/pet/1", Exchange: ex("POST", 200, "requirement=auth-token")}, true,
			Tests{Participants: Passed, Method: Passed, SenderKnown: true, Path: Failed, Status: Failed, Headers: Failed}},
		{"wrong method", Resolved{From: "agent", To: "resource", Path: "/v2/store/order", Exchange: ex("GET", 401, "")}, false,
			Tests{Participants: Passed, Method: Failed, SenderKnown: true, Path: Passed, Status: Passed, Headers: Failed}},
		{"wrong receiver", Resolved{From: "agent", To: "ap", Exchange: ex("POST", 401, "")}, false,
			Tests{Participants: Failed, Method: Passed, Path: Failed, Status: Passed, Headers: Failed}},
		{"wrong sender", Resolved{From: "ps", To: "resource", Exchange: ex("POST", 401, "")}, false,
			Tests{Participants: Failed, Method: Passed, Path: Failed, Status: Passed, Headers: Failed}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Test(step, tt.r)
			if got != tt.want || got.Candidate() != tt.candidate {
				t.Errorf("Test = %+v (candidate %v), want %+v (%v)", got, got.Candidate(), tt.want, tt.candidate)
			}
		})
	}
}

func TestPathMatches(t *testing.T) {
	tests := []struct {
		rule, path string
		want       bool
	}{
		{"/pending/{id}", "/pending/2U5_3Rqrgp425", true},
		{"/pending/{id}", "/pending/", false},
		{"/pending/{id}", "/pending/a/b", false},
		{"/agent-token", "/agent-token/", true},
		{"/agent-token", "/agent", false},
		{"^/pending/", "/pending/abc", true},
		{"^/pending/", "/other", false},
		{"^(", "/x", false},
	}
	for _, tt := range tests {
		if got := PathMatches(tt.rule, tt.path); got != tt.want {
			t.Errorf("PathMatches(%q, %q) = %v, want %v", tt.rule, tt.path, got, tt.want)
		}
	}
}

func TestValueMatches(t *testing.T) {
	tests := []struct {
		rule, value string
		want        bool
	}{
		{"requirement=person-token", "requirement=person-token", true},
		{"requirement=person-token", "requirement=person-token; x=1", false},
		{"requirement=auth-token*", `requirement=auth-token; resource-token="eyJ…"`, true},
		{`^requirement=(auth|person)-token`, "requirement=person-token", true},
		{`^requirement=interaction`, "requirement=person-token", false},
	}
	for _, tt := range tests {
		if got := ValueMatches(tt.rule, tt.value); got != tt.want {
			t.Errorf("ValueMatches(%q, %q) = %v, want %v", tt.rule, tt.value, got, tt.want)
		}
	}
}

func TestLiteralSegment(t *testing.T) {
	for path, want := range map[string]string{
		"/pending/f7a3b9c":   "f7a3b9c",
		"/pending/{id}":      "",
		"/api/documents":     "",
		"/v2/store/order/1":  "",
		"/oauth2/authorize":  "",
		"/pending/2U5_3Rqrg": "2U5_3Rqrg",
		"/x/abc123/def456":   "abc123",
		"/.well-known/jwks1": "",
	} {
		if got := literalSegment(path); got != want {
			t.Errorf("literalSegment(%q) = %q, want %q", path, got, want)
		}
	}
}
