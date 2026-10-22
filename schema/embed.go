// Package schema provides the embedded JSON Schemas: PIDL, the source
// format people write; the walkthrough format PIDL generates; and the
// sample set format implementations keep their code samples in.
package schema

import (
	_ "embed"
)

// SchemaJSON is the embedded PIDL JSON Schema.
//
//go:embed pidl.schema.json
var SchemaJSON []byte

// SchemaVersion is the version of the embedded schema.
const SchemaVersion = "1.0"

// WalkthroughSchemaJSON is the JSON Schema of the walkthrough format
// (pidl-walkthrough/v1), generated from pidl.Walkthrough by schema/gen.
//
//go:embed walkthrough.schema.json
var WalkthroughSchemaJSON []byte

// SampleSetSchemaJSON is the JSON Schema of the sample set format
// (pidl-samples/v1), generated from pidl.SampleSet by schema/gen.
//
//go:embed samples.schema.json
var SampleSetSchemaJSON []byte
