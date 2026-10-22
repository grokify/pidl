package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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
