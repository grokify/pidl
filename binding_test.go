package pidl

import (
	"encoding/json"
	"path/filepath"
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

func TestBindingStepRule(t *testing.T) {
	b, err := ParseBinding([]byte(bindingJSON))
	if err != nil {
		t.Fatal(err)
	}
	m, ok := b.StepRule("challenge", "request")
	if !ok || m.Method != "POST" || m.Status != 401 {
		t.Errorf("scenario key: %+v, %v; want the scenario's own rule", m, ok)
	}
	m, ok = b.StepRule("agent-identity", "request")
	if !ok || m.Method != "GET" {
		t.Errorf("wildcard key: %+v, %v; want the wildcard rule", m, ok)
	}
	m, ok = b.StepRule("ps-authorization", "consent.poll")
	if !ok || !m.Repeat {
		t.Errorf("wildcard repeat: %+v, %v", m, ok)
	}
	if _, ok := b.StepRule("x", "missing"); ok {
		t.Error("found a rule that is not there")
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
