package main

import (
	"bytes"
	"os"
	"testing"
)

// TestSchemaIsCurrent fails when a committed schema differs from what the
// Go types generate.
func TestSchemaIsCurrent(t *testing.T) {
	for _, spec := range Specs {
		want, err := Generate(spec)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(spec.Output)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s is out of date; run go run . in schema/gen and commit it", spec.Output)
		}
	}
}
