package pidl

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

const bindingJSON = `{
  "format": "pidl-binding/v1",
  "participants": {
    "agent":    {"labels": ["agent"]},
    "ap":       {"labels": ["ap"], "hosts": ["ap.example.test"]},
    "resource": {"labels": ["petstore.example.test"], "hosts": ["petstore.example.test"]}
  },
  "labelsByPath": [
    {"label": "client", "path": "^/(\\.well-known/.*|jwks\\.json)$", "participant": "resource"}
  ],
  "paths": {"ap": {"*/agent-token.attest": "/agent-token"}},
  "rewrite": [{"match": "^http://sandbox:8099/(https://.*)$", "to": "$1"}],
  "unobservable": ["person"],
  "steps": {
    "*/consent.poll": {"match": {"method": "GET", "path": "^/pending/", "repeat": true}},
    "challenge/request": {"match": {"method": "POST", "path": "/v2/store/order", "status": 401,
                                    "headers": {"AAuth-Requirement": "requirement=person-token*"},
                                    "body": {"/petId": 1, "/quantity": true}}},
    "*/request": {"match": {"method": "GET"}}
  }
}`

func TestParseBinding(t *testing.T) {
	b, err := ParseBinding([]byte(bindingJSON))
	if err != nil {
		t.Fatal(err)
	}
	if errs := b.Validate(); len(errs) > 0 {
		t.Fatalf("valid binding reported: %v", errs)
	}
	if b.Participants["ap"].Hosts[0] != "ap.example.test" || b.LabelsByPath[0].Participant != "resource" {
		t.Errorf("unexpected binding: %+v", b)
	}
}

func TestParseBindingRefusesOtherVersion(t *testing.T) {
	if _, err := ParseBinding([]byte(`{"format": "pidl-binding/v2"}`)); err == nil {
		t.Fatal("another major version was accepted")
	}
}

func TestStepKeys(t *testing.T) {
	tests := []struct {
		name                                  string
		scenario, expanded, definedIn, flowID string
		want                                  []string
	}{
		{"flow in its own scenario", "challenge", "request", "challenge", "request", []string{"challenge/request", "*/request"}},
		{"flow from a fragment", "ps-authorization", "consent.poll", "consent", "poll", []string{"ps-authorization/consent.poll", "consent/poll", "*/poll"}},
		{"no ID", "s", "", "s", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StepKeys(tt.scenario, tt.expanded, tt.definedIn, tt.flowID)
			if strings.Join(got, " ") != strings.Join(tt.want, " ") {
				t.Errorf("StepKeys = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindingStepRule(t *testing.T) {
	b := &Binding{Format: BindingFormat, Steps: map[string]BindingStep{
		"ps-authorization/consent.poll": {Match: StepMatch{Status: 202}},
		"consent/poll":                  {Match: StepMatch{Status: 200}},
		"*/poll":                        {Match: StepMatch{Method: "GET", Repeat: true}},
		"challenge/request":             {Match: StepMatch{Method: "POST"}},
		"*/request":                     {Match: StepMatch{Method: "GET"}},
	}}
	tests := []struct {
		name                                  string
		scenario, expanded, definedIn, flowID string
		wantKey                               string
	}{
		{"scenario key wins", "ps-authorization", "consent.poll", "consent", "poll", "ps-authorization/consent.poll"},
		{"defining scenario next", "ps-authorization-challenge", "consent.poll", "consent", "poll", "consent/poll"},
		{"wildcard last", "other", "poll", "other", "poll", "*/poll"},
		{"flow in its scenario", "challenge", "request", "challenge", "request", "challenge/request"},
		{"same ID elsewhere takes the wildcard", "agent-identity", "request", "agent-identity", "request", "*/request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, key, ok := b.StepRule(tt.scenario, tt.expanded, tt.definedIn, tt.flowID)
			if !ok || key != tt.wantKey {
				t.Errorf("StepRule key = %q, %v; want %q", key, ok, tt.wantKey)
			}
		})
	}
	if _, _, ok := b.StepRule("x", "missing", "x", "missing"); ok {
		t.Error("found a rule that is not there")
	}
}

func TestBindingStepPath(t *testing.T) {
	b := &Binding{Format: BindingFormat, Paths: map[string]map[string]string{
		"resource": {"*/api-request": "/v2/store/order", "agent-identity/request": "/v2/pet/findByStatus"},
	}}
	if p, key, ok := b.StepPath("resource", "agent-identity", "request", "agent-identity", "request"); !ok || p != "/v2/pet/findByStatus" || key != "agent-identity/request" {
		t.Errorf("StepPath = %q, %q, %v", p, key, ok)
	}
	if p, _, ok := b.StepPath("resource", "ps-authorization", "api-request", "ps-authorization", "api-request"); !ok || p != "/v2/store/order" {
		t.Errorf("wildcard StepPath = %q, %v", p, ok)
	}
	if _, _, ok := b.StepPath("ap", "s", "x", "s", "x"); ok {
		t.Error("found a path for a participant with none")
	}
}

func TestBindingRoundTrip(t *testing.T) {
	b, err := ParseBinding([]byte(bindingJSON))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "sandbox.binding.json")
	if err := b.WriteFile(path); err != nil {
		t.Fatal(err)
	}
	back, err := ReadBinding(path)
	if err != nil {
		t.Fatal(err)
	}
	x, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	y, err := json.Marshal(back)
	if err != nil {
		t.Fatal(err)
	}
	if string(x) != string(y) {
		t.Errorf("round trip changed the binding:\n%s\n%s", x, y)
	}
}

func TestBindingValidate(t *testing.T) {
	b := &Binding{
		Format:       "pidl-binding/v1",
		Participants: map[string]BindingParticipant{"ap": {Hosts: []string{"https://ap.example.test"}}},
		LabelsByPath: []BindingLabelRule{{Path: "(", Participant: "nobody"}},
		Paths:        map[string]map[string]string{"ghost": {"no-slash": "/x"}},
		Rewrite:      []BindingRewrite{{Match: "["}},
		Steps: map[string]BindingStep{
			"bad-key": {},
			"s/a/b":   {},
			"s/step": {Match: StepMatch{
				Path: "^(", Status: 700,
				Headers: map[string]string{"X": "^["},
				Body:    map[string]json.RawMessage{"nopointer": json.RawMessage(`1`), "/bad": json.RawMessage(`{`)},
			}},
		},
	}
	got := map[string]bool{}
	for _, e := range b.Validate() {
		got[e.Field] = true
	}
	for _, f := range []string{
		"participants[ap].hosts[0]", "labelsByPath[0].label", "labelsByPath[0].path", "labelsByPath[0].participant",
		"paths[ghost]", "paths[ghost][no-slash]", "rewrite[0].match", "steps[bad-key]", "steps[s/a/b]",
		"steps[s/step].match.path", "steps[s/step].match.status", "steps[s/step].match.headers[X]",
		"steps[s/step].match.body[nopointer]", "steps[s/step].match.body[/bad]",
	} {
		if !got[f] {
			t.Errorf("no error for %s; got %v", f, got)
		}
	}
	if errs := (&Binding{Format: BindingFormat}).Validate(); len(errs) != 1 || errs[0].Field != "participants" {
		t.Errorf("empty binding: %v", errs)
	}
}

func TestProtocolValidatesStepMatch(t *testing.T) {
	p, err := Parse([]byte(`{
	  "protocol": {"id": "m", "name": "M", "version": "1.0"},
	  "entities": [{"id": "a", "name": "A", "type": "agent"}, {"id": "b", "name": "B", "type": "server"}],
	  "scenarios": [{"id": "s", "name": "S", "flows": [
	    {"id": "x", "from": "a", "to": "b", "action": "call", "label": "Call", "match": {"path": "^(", "status": 42}}
	  ]}]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, e := range p.Validate() {
		if strings.Contains(e.Field, ".match.") {
			got[e.Field[strings.Index(e.Field, ".match."):]] = true
		}
	}
	if !got[".match.path"] || !got[".match.status"] {
		t.Errorf("match errors = %v, want path and status", got)
	}
}
