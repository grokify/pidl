package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/schema"
)

// TestWalkthroughSchemaFormat checks that the embedded schemas describe
// the format versions the library reads and writes.
func TestWalkthroughSchemaFormat(t *testing.T) {
	checkFormat(t, schema.WalkthroughSchemaJSON, "Walkthrough", pidl.WalkthroughFormat, "https://github.com/grokify/pidl/schema/walkthrough.schema.json")
	checkFormat(t, schema.SampleSetSchemaJSON, "SampleSet", pidl.SampleSetFormat, "https://github.com/grokify/pidl/schema/samples.schema.json")
	checkFormat(t, schema.ObservationSchemaJSON, "Observation", pidl.ObservationFormat, "https://github.com/grokify/pidl/schema/observation.schema.json")
	checkFormat(t, schema.BindingSchemaJSON, "Binding", pidl.BindingFormat, "https://github.com/grokify/pidl/schema/binding.schema.json")
}

func checkFormat(t *testing.T, data []byte, def, want, id string) {
	t.Helper()
	var s struct {
		ID string `json:"$id"`
		// A property schema may be the boolean true, so stay raw.
		Defs map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	var format struct {
		Const string `json:"const"`
	}
	if err := json.Unmarshal(s.Defs[def].Properties["format"], &format); err != nil {
		t.Fatal(err)
	}
	if got := format.Const; got != want {
		t.Fatalf("%s format const = %q, want %q", def, got, want)
	}
	if s.ID != id {
		t.Errorf("$id = %q", s.ID)
	}
}
