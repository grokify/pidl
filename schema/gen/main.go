// Command gen writes schema/walkthrough.schema.json from the Go types of
// the walkthrough format (pidl.Walkthrough). It is a separate module so
// the pidl module itself keeps no dependencies.
//
// Run it from this directory:
//
//	go run .
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/grokify/pidl"
)

const (
	// Output is the schema file, relative to this directory.
	Output = "../walkthrough.schema.json"
	// RepoRoot holds the pidl package, relative to this directory.
	RepoRoot = "../.."
)

func main() {
	b, err := Generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(Output, b, 0o644); err != nil { //nolint:gosec // a schema file in the repository is meant to be readable
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Generate returns the walkthrough schema as indented JSON.
func Generate() ([]byte, error) {
	// Readers ignore fields they do not know, so a walkthrough from a
	// newer v1 producer still validates against an older v1 schema.
	r := &jsonschema.Reflector{AllowAdditionalProperties: true}
	comments, err := rootComments()
	if err != nil {
		return nil, err
	}
	r.CommentMap = comments
	s := r.Reflect(&pidl.Walkthrough{})
	s.ID = "https://github.com/grokify/pidl/schema/walkthrough.schema.json"
	s.Title = "PIDL Walkthrough"
	s.Description = "A protocol prepared for step-by-step presentation, produced by `pidl generate -f walkthrough-json`."

	def, ok := s.Definitions["Walkthrough"]
	if !ok {
		return nil, fmt.Errorf("no Walkthrough definition in the reflected schema")
	}
	format, ok := def.Properties.Get("format")
	if !ok {
		return nil, fmt.Errorf("Walkthrough has no format property")
	}
	format.Const = pidl.WalkthroughFormat

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// rootComments returns the doc comments of the pidl package's types, keyed
// as the reflector expects. The extractor names packages after the walked
// path, so it reads a directory holding only the root package's files.
func rootComments() (map[string]string, error) {
	files, err := filepath.Glob(filepath.Join(RepoRoot, "*.go"))
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "pidl-schema-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir) //nolint:errcheck // a leftover temporary directory is harmless
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.Base(f)), b, 0o600); err != nil { //nolint:gosec // f comes from a glob of the repository root, and only its base name is used
			return nil, err
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if err := os.Chdir(dir); err != nil {
		return nil, err
	}
	r := &jsonschema.Reflector{}
	extractErr := r.AddGoComments("github.com/grokify/pidl", ".")
	if err := os.Chdir(wd); err != nil {
		return nil, err
	}
	if extractErr != nil {
		return nil, fmt.Errorf("reading doc comments: %w", extractErr)
	}
	return r.CommentMap, nil
}
