package main

import (
	"bytes"
	"os"
	"testing"
)

// TestSchemaIsCurrent fails when the committed schema differs from what
// the Go types generate.
func TestSchemaIsCurrent(t *testing.T) {
	want, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(Output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s is out of date; run go run . in schema/gen and commit it", Output)
	}
}
