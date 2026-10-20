package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/schema"
)

// TestWalkthroughSchemaFormat checks that the embedded schema describes
// the format version the library writes.
func TestWalkthroughSchemaFormat(t *testing.T) {
	var s struct {
		ID string `json:"$id"`
		// A property schema may be the boolean true, so stay raw.
		Defs map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(schema.WalkthroughSchemaJSON, &s); err != nil {
		t.Fatal(err)
	}
	var format struct {
		Const string `json:"const"`
	}
	if err := json.Unmarshal(s.Defs["Walkthrough"].Properties["format"], &format); err != nil {
		t.Fatal(err)
	}
	if got := format.Const; got != pidl.WalkthroughFormat {
		t.Fatalf("schema format const = %q, want %q", got, pidl.WalkthroughFormat)
	}
	if s.ID != "https://github.com/grokify/pidl/schema/walkthrough.schema.json" {
		t.Errorf("$id = %q", s.ID)
	}
}
