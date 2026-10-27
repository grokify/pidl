package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/examples"
	"github.com/grokify/pidl/render"
	"github.com/grokify/pidl/schema"
)

// TestExamplesValidate renders every built-in example as a walkthrough
// and validates it against the embedded schema.
func TestExamplesValidate(t *testing.T) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema.WalkthroughSchemaJSON))
	if err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	const url = "https://github.com/grokify/pidl/schema/walkthrough.schema.json"
	if err := c.AddResource(url, doc); err != nil {
		t.Fatal(err)
	}
	sch, err := c.Compile(url)
	if err != nil {
		t.Fatal(err)
	}
	all, err := examples.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("no built-in examples")
	}
	r := render.NewWalkthroughJSON()
	for _, ex := range all {
		t.Run(ex.Name, func(t *testing.T) {
			p, err := ex.Protocol()
			if err != nil {
				t.Fatal(err)
			}
			b, err := json.Marshal(r.Walkthrough(p))
			if err != nil {
				t.Fatal(err)
			}
			v, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			if err := sch.Validate(v); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

// TestSampleSetValidates checks a sample set, as resolved and written by
// the library, against the sample set schema.
func TestSampleSetValidates(t *testing.T) {
	c := jsonschema.NewCompiler()
	if err := c.AddResource("samples.schema.json", mustJSON(t, "../samples.schema.json")); err != nil {
		t.Fatal(err)
	}
	sch, err := c.Compile("samples.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	set := &pidl.SampleSet{Format: pidl.SampleSetFormat, Protocol: "demo", Name: "demo-sdk-go", Samples: []pidl.SampleSetEntry{
		{Step: "sign/request", Sample: pidl.Sample{Language: "go", Role: "agent", Code: "sdk.Sign(req)",
			Source: &pidl.SampleSource{Repo: "example.com/sdk", Path: "sign_test.go", Region: "sign"}}, SHA256: "ab"},
		{Step: "sign", Sample: pidl.Sample{Language: "python", Code: "sdk.sign(req)"}},
	}}
	path := filepath.Join(t.TempDir(), "set.json")
	if err := set.WriteFile(path); err != nil {
		t.Fatal(err)
	}
	if err := sch.Validate(mustJSON(t, path)); err != nil {
		t.Fatalf("sample set does not validate: %v", err)
	}
	var missing map[string]any
	if err := json.Unmarshal([]byte(`{"format": "pidl-samples/v1"}`), &missing); err != nil {
		t.Fatal(err)
	}
	if err := sch.Validate(missing); err == nil {
		t.Error("a set without samples validated")
	}
}

// TestObservationAndBindingValidate checks an observation and a binding,
// as written by the library, against their schemas, and that each schema
// refuses a document missing a required member.
func TestObservationAndBindingValidate(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ok := true
	body, _, err := pidl.NewObservedBody([]byte(`{"agent_token":"eyJ…"}`), "application/json", 0)
	if err != nil {
		t.Fatal(err)
	}
	obs := &pidl.Observation{
		Format: pidl.ObservationFormat,
		Source: pidl.ObservationSource{Tool: "aauth-sandbox", RequestFidelity: pidl.RequestFidelityWire, From: &now, To: &now},
		Exchanges: []pidl.Exchange{{
			Time: &now, From: "agent", To: "ap", RequestID: "r1", Label: "agent token",
			Request:     pidl.ObservedRequest{Method: "POST", URL: "https://ap.example.test/agent-token", Headers: pidl.ObservedHeaders{{"Content-Type", "application/json"}}},
			Response:    pidl.ObservedResponse{Status: 200, Body: body},
			Annotations: []pidl.ExchangeAnnotation{{Kind: "assertion", Text: "status 200", OK: &ok}},
		}},
	}
	bind := &pidl.Binding{
		Format:       pidl.BindingFormat,
		Participants: map[string]pidl.BindingParticipant{"ap": {Labels: []string{"ap"}, Hosts: []string{"ap.example.test"}}},
		LabelsByPath: []pidl.BindingLabelRule{{Label: "client", Path: "^/jwks\\.json$", Participant: "ap"}},
		Rewrite:      []pidl.BindingRewrite{{Match: "^http://sandbox:8099/(https://.*)$", To: "$1"}},
		Steps: map[string]pidl.BindingStep{"*/poll": {Match: pidl.StepMatch{
			Method: "GET", Path: "/pending/{id}", Status: 200, Repeat: true,
			Headers: map[string]string{"Location": "^https://"}, Body: map[string]json.RawMessage{"/a": json.RawMessage(`true`)},
		}}},
	}
	dir := t.TempDir()
	for _, tc := range []struct {
		schema, missing string
		write           func(string) error
	}{
		{"observation.schema.json", `{"format": "pidl-observation/v1", "source": {"tool": "x"}}`, obs.WriteFile},
		{"binding.schema.json", `{"format": "pidl-binding/v1"}`, bind.WriteFile},
	} {
		c := jsonschema.NewCompiler()
		if err := c.AddResource(tc.schema, mustJSON(t, "../"+tc.schema)); err != nil {
			t.Fatal(err)
		}
		sch, err := c.Compile(tc.schema)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, tc.schema)
		if err := tc.write(path); err != nil {
			t.Fatal(err)
		}
		if err := sch.Validate(mustJSON(t, path)); err != nil {
			t.Fatalf("%s: written document does not validate: %v", tc.schema, err)
		}
		var missing map[string]any
		if err := json.Unmarshal([]byte(tc.missing), &missing); err != nil {
			t.Fatal(err)
		}
		if err := sch.Validate(missing); err == nil {
			t.Errorf("%s: a document missing a required member validated", tc.schema)
		}
	}
}

func mustJSON(t *testing.T, path string) any {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() //nolint:errcheck // read-only file
	v, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
