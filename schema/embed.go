// Package schema provides the embedded JSON Schemas: PIDL, the source
// format people write, and the walkthrough format PIDL generates.
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
