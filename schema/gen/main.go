// Command gen writes the JSON Schemas of the formats PIDL reads and
// writes from their Go types: schema/walkthrough.schema.json from
// pidl.Walkthrough, and schema/samples.schema.json from pidl.SampleSet. It
// is a separate module so the pidl module itself keeps no dependencies.
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
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/grokify/pidl"
)

// RepoRoot holds the pidl package, relative to this directory.
const RepoRoot = "../.."

// Spec describes one generated schema.
type Spec struct {
	// Output is the schema file, relative to this directory.
	Output string
	// Value is a value of the root type.
	Value any
	// Definition names the root type's definition.
	Definition string
	// Format is the value the root's format property must have.
	Format      string
	ID          string
	Title       string
	Description string
}

// Specs lists the schemas gen writes.
var Specs = []Spec{
	{
		Output: "../walkthrough.schema.json", Value: &pidl.Walkthrough{}, Definition: "Walkthrough",
		Format:      pidl.WalkthroughFormat,
		ID:          "https://github.com/grokify/pidl/schema/walkthrough.schema.json",
		Title:       "PIDL Walkthrough",
		Description: "A protocol prepared for step-by-step presentation, produced by `pidl generate -f walkthrough-json`.",
	},
	{
		Output: "../samples.schema.json", Value: &pidl.SampleSet{}, Definition: "SampleSet",
		Format:      pidl.SampleSetFormat,
		ID:          "https://github.com/grokify/pidl/schema/samples.schema.json",
		Title:       "PIDL Sample Set",
		Description: "Code samples an implementation keeps apart from a protocol document, each attached to a step by its address, read by `pidl samples` and `pidl generate --samples`.",
	},
}

func main() {
	for _, spec := range Specs {
		b, err := Generate(spec)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(spec.Output, b, 0o644); err != nil { //nolint:gosec // a schema file in the repository is meant to be readable
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

// Generate returns a schema as indented JSON.
func Generate(spec Spec) ([]byte, error) {
	// Readers ignore fields they do not know, so a file from a newer v1
	// producer still validates against an older v1 schema.
	r := &jsonschema.Reflector{AllowAdditionalProperties: true, Mapper: enums}
	comments, err := rootComments()
	if err != nil {
		return nil, err
	}
	r.CommentMap = comments
	s := r.Reflect(spec.Value)
	s.ID = jsonschema.ID(spec.ID)
	s.Title = spec.Title
	s.Description = spec.Description

	def, ok := s.Definitions[spec.Definition]
	if !ok {
		return nil, fmt.Errorf("no %s definition in the reflected schema", spec.Definition)
	}
	format, ok := def.Properties.Get("format")
	if !ok {
		return nil, fmt.Errorf("%s has no format property", spec.Definition)
	}
	format.Const = spec.Format

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	// The schema types marshal their own fields, escaping <, > and & in
	// descriptions; keep those readable.
	out := strings.NewReplacer(`\u003c`, "<", `\u003e`, ">", `\u0026`, "&").Replace(buf.String())
	return []byte(out), nil
}

// enums gives string types with a closed set of values their enum, which
// reflection cannot see.
func enums(t reflect.Type) *jsonschema.Schema {
	if t == reflect.TypeFor[pidl.ExampleKind]() {
		return &jsonschema.Schema{Type: "string", Enum: []any{pidl.ExampleKindDraft, pidl.ExampleKindDerived}}
	}
	return nil
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
