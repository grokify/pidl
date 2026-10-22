package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

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
