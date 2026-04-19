// Package schema provides the embedded PIDL JSON Schema.
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
