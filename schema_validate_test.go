package pidl

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestValidateData_TypeChecks(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		schema  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid string",
			data: "hello",
			schema: map[string]interface{}{
				"type": "string",
			},
			wantErr: false,
		},
		{
			name: "invalid string",
			data: 123,
			schema: map[string]interface{}{
				"type": "string",
			},
			wantErr: true,
		},
		{
			name: "valid number",
			data: 42.5,
			schema: map[string]interface{}{
				"type": "number",
			},
			wantErr: false,
		},
		{
			name: "valid integer",
			data: float64(42),
			schema: map[string]interface{}{
				"type": "integer",
			},
			wantErr: false,
		},
		{
			name: "invalid integer (float)",
			data: 42.5,
			schema: map[string]interface{}{
				"type": "integer",
			},
			wantErr: true,
		},
		{
			name: "valid boolean",
			data: true,
			schema: map[string]interface{}{
				"type": "boolean",
			},
			wantErr: false,
		},
		{
			name: "valid array",
			data: []interface{}{"a", "b"},
			schema: map[string]interface{}{
				"type": "array",
			},
			wantErr: false,
		},
		{
			name: "valid object",
			data: map[string]interface{}{"key": "value"},
			schema: map[string]interface{}{
				"type": "object",
			},
			wantErr: false,
		},
		{
			name: "valid null",
			data: nil,
			schema: map[string]interface{}{
				"type": "null",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateData(tt.data, tt.schema)
			if tt.wantErr && result.Valid {
				t.Error("expected validation to fail")
			}
			if !tt.wantErr && !result.Valid {
				t.Errorf("expected validation to pass, got errors: %s", result.Error())
			}
		})
	}
}

func TestValidateData_RequiredFields(t *testing.T) {
	schema := map[string]interface{}{
		"type":     "object",
		"required": []interface{}{"name", "age"},
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
			"age":  map[string]interface{}{"type": "number"},
		},
	}

	// Missing required field
	data := map[string]interface{}{
		"name": "John",
	}
	result := ValidateData(data, schema)
	if result.Valid {
		t.Error("expected validation to fail for missing required field")
	}

	// All required fields present
	data = map[string]interface{}{
		"name": "John",
		"age":  float64(30),
	}
	result = ValidateData(data, schema)
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %s", result.Error())
	}
}

func TestValidateData_NestedObjects(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"user": map[string]interface{}{
				"type":     "object",
				"required": []interface{}{"email"},
				"properties": map[string]interface{}{
					"email": map[string]interface{}{"type": "string"},
					"name":  map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	// Valid nested object
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}
	result := ValidateData(data, schema)
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %s", result.Error())
	}

	// Invalid nested object (wrong type)
	data = map[string]interface{}{
		"user": map[string]interface{}{
			"email": 123, // Should be string
			"name":  "Test User",
		},
	}
	result = ValidateData(data, schema)
	if result.Valid {
		t.Error("expected validation to fail for wrong type in nested object")
	}
}

func TestValidateData_ArrayItems(t *testing.T) {
	schema := map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "string",
		},
	}

	// Valid array
	data := []interface{}{"a", "b", "c"}
	result := ValidateData(data, schema)
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %s", result.Error())
	}

	// Invalid array (mixed types)
	data = []interface{}{"a", 123, "c"}
	result = ValidateData(data, schema)
	if result.Valid {
		t.Error("expected validation to fail for invalid array item type")
	}
}

func TestValidateData_Enum(t *testing.T) {
	schema := map[string]interface{}{
		"type": "string",
		"enum": []interface{}{"red", "green", "blue"},
	}

	// Valid enum value
	result := ValidateData("green", schema)
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %s", result.Error())
	}

	// Invalid enum value
	result = ValidateData("yellow", schema)
	if result.Valid {
		t.Error("expected validation to fail for invalid enum value")
	}
}

func TestValidateData_MinMax(t *testing.T) {
	schema := map[string]interface{}{
		"type":    "number",
		"minimum": float64(0),
		"maximum": float64(100),
	}

	// Valid range
	result := ValidateData(float64(50), schema)
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %s", result.Error())
	}

	// Below minimum
	result = ValidateData(float64(-5), schema)
	if result.Valid {
		t.Error("expected validation to fail for value below minimum")
	}

	// Above maximum
	result = ValidateData(float64(150), schema)
	if result.Valid {
		t.Error("expected validation to fail for value above maximum")
	}
}

func TestValidateData_StringLength(t *testing.T) {
	schema := map[string]interface{}{
		"type":      "string",
		"minLength": float64(3),
		"maxLength": float64(10),
	}

	// Valid length
	result := ValidateData("hello", schema)
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %s", result.Error())
	}

	// Too short
	result = ValidateData("hi", schema)
	if result.Valid {
		t.Error("expected validation to fail for string too short")
	}

	// Too long
	result = ValidateData("hello world!", schema)
	if result.Valid {
		t.Error("expected validation to fail for string too long")
	}
}

func TestSchemaRegistry_LoadSchema(t *testing.T) {
	// Create a temp schema file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "test.schema.json")
	schemaContent := `{"type": "object", "properties": {"name": {"type": "string"}}}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0600); err != nil {
		t.Fatalf("failed to write test schema: %v", err)
	}

	registry := NewSchemaRegistry()

	// Load from file path
	schema, err := registry.LoadSchema(schemaPath)
	if err != nil {
		t.Fatalf("LoadSchema failed: %v", err)
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		t.Fatal("expected schema to be a map")
	}

	if schemaMap["type"] != "object" {
		t.Errorf("expected type 'object', got %v", schemaMap["type"])
	}

	// Load again (should use cache)
	_, err = registry.LoadSchema(schemaPath)
	if err != nil {
		t.Fatalf("LoadSchema (cached) failed: %v", err)
	}
	// Just verify no error on cache hit
}

func TestSchemaRegistry_RelativePath(t *testing.T) {
	// Create a temp schema file
	tmpDir := t.TempDir()
	schemaDir := filepath.Join(tmpDir, "schemas")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("failed to create schema dir: %v", err)
	}

	schemaPath := filepath.Join(schemaDir, "user.schema.json")
	schemaContent := `{"type": "object"}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0600); err != nil {
		t.Fatalf("failed to write test schema: %v", err)
	}

	registry := NewSchemaRegistry()
	registry.SetBasePath(tmpDir)

	// Load using relative path
	schema, err := registry.LoadSchema("schemas/user.schema.json")
	if err != nil {
		t.Fatalf("LoadSchema with relative path failed: %v", err)
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok || schemaMap["type"] != "object" {
		t.Error("expected schema to be loaded correctly")
	}
}

func TestPortSchemaValidator_ValidatePortData(t *testing.T) {
	// Create a temp schema file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "user.schema.json")
	schemaContent := `{
		"type": "object",
		"required": ["id", "name"],
		"properties": {
			"id": {"type": "integer"},
			"name": {"type": "string"}
		}
	}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0600); err != nil {
		t.Fatalf("failed to write test schema: %v", err)
	}

	validator := NewPortSchemaValidator()
	validator.SetBasePath(tmpDir)

	port := DataPort{
		Name:   "user_input",
		Kind:   DataPortKindObject,
		Schema: "user.schema.json",
	}

	// Valid data
	data := map[string]interface{}{
		"id":   float64(1),
		"name": "John",
	}
	result, err := validator.ValidatePortData(port, data)
	if err != nil {
		t.Fatalf("ValidatePortData failed: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %s", result.Error())
	}

	// Invalid data (missing required field)
	data = map[string]interface{}{
		"id": float64(1),
	}
	result, err = validator.ValidatePortData(port, data)
	if err != nil {
		t.Fatalf("ValidatePortData failed: %v", err)
	}
	if result.Valid {
		t.Error("expected validation to fail for missing required field")
	}
}

func TestPortSchemaValidator_NoSchema(t *testing.T) {
	validator := NewPortSchemaValidator()

	port := DataPort{
		Name: "data",
		Kind: DataPortKindObject,
		// No schema defined
	}

	// Should pass without schema
	result, err := validator.ValidatePortData(port, map[string]interface{}{"anything": "goes"})
	if err != nil {
		t.Fatalf("ValidatePortData failed: %v", err)
	}
	if !result.Valid {
		t.Error("expected validation to pass when no schema defined")
	}
}

func TestPortSchemaValidator_ValidateEntityInputs(t *testing.T) {
	// Create temp schema files
	tmpDir := t.TempDir()
	userSchemaPath := filepath.Join(tmpDir, "user.schema.json")
	userSchemaContent := `{"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}}`
	if err := os.WriteFile(userSchemaPath, []byte(userSchemaContent), 0600); err != nil {
		t.Fatalf("failed to write user schema: %v", err)
	}

	validator := NewPortSchemaValidator()
	validator.SetBasePath(tmpDir)

	entity := Entity{
		ID:   "processor",
		Name: "Data Processor",
		Inputs: []DataPort{
			{Name: "user", Kind: DataPortKindObject, Schema: "user.schema.json", Required: true},
			{Name: "config", Kind: DataPortKindObject, Required: false},
		},
	}

	// Valid inputs
	inputs := map[string]interface{}{
		"user":   map[string]interface{}{"name": "John"},
		"config": map[string]interface{}{"debug": true},
	}
	result, err := validator.ValidateEntityInputs(entity, inputs)
	if err != nil {
		t.Fatalf("ValidateEntityInputs failed: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected validation to pass, got errors: %s", result.Error())
	}

	// Missing required input
	inputs = map[string]interface{}{
		"config": map[string]interface{}{"debug": true},
	}
	result, err = validator.ValidateEntityInputs(entity, inputs)
	if err != nil {
		t.Fatalf("ValidateEntityInputs failed: %v", err)
	}
	if result.Valid {
		t.Error("expected validation to fail for missing required input")
	}
}

func TestIsValidSchemaURI(t *testing.T) {
	tests := []struct {
		uri   string
		valid bool
	}{
		{"", true},
		{"#/definitions/User", true},
		{"./schemas/user.json", true},
		{"../common/schema.json", true},
		{"/absolute/path/schema.json", true},
		{"file:///path/to/schema.json", true},
		{"https://example.com/schema.json", true},
		{"http://example.com/schema.json", true},
		{"user.schema.json", true},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			if got := IsValidSchemaURI(tt.uri); got != tt.valid {
				t.Errorf("IsValidSchemaURI(%q) = %v, want %v", tt.uri, got, tt.valid)
			}
		})
	}
}

func TestSchemaValidationResult_Error(t *testing.T) {
	// Valid result
	result := &SchemaValidationResult{Valid: true}
	if result.Error() != "" {
		t.Error("expected empty error for valid result")
	}

	// Invalid result with errors
	result = &SchemaValidationResult{
		Valid: false,
		Errors: []SchemaValidationError{
			{Path: "name", Message: "required field missing"},
			{Path: "age", Message: "expected type number"},
		},
	}
	errStr := result.Error()
	if errStr == "" {
		t.Error("expected non-empty error for invalid result")
	}
	if errStr != "name: required field missing; age: expected type number" {
		t.Errorf("unexpected error format: %s", errStr)
	}
}

func TestSchemaRegistry_RegisterSchema(t *testing.T) {
	tests := []struct {
		name   string
		uri    string
		schema interface{}
	}{
		{name: "inline-style uri", uri: "#/defs/user", schema: map[string]interface{}{"type": "object"}},
		{name: "http uri never fetched", uri: "https://example.invalid/never-fetched.json", schema: map[string]interface{}{"type": "string"}},
		{name: "overwrites existing", uri: "#/defs/user", schema: map[string]interface{}{"type": "integer"}},
	}

	registry := NewSchemaRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry.RegisterSchema(tt.uri, tt.schema)

			got, err := registry.LoadSchema(tt.uri)
			if err != nil {
				t.Fatalf("LoadSchema(%q) after RegisterSchema error: %v", tt.uri, err)
			}
			if !reflect.DeepEqual(got, tt.schema) {
				t.Errorf("LoadSchema(%q) = %v, want registered %v", tt.uri, got, tt.schema)
			}
		})
	}
}

func TestPortSchemaValidator_ValidateEntityOutputs(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "result.schema.json")
	schemaContent := `{"type": "object", "required": ["id"], "properties": {"id": {"type": "string"}}}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0600); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	entity := Entity{
		ID: "producer",
		Outputs: []DataPort{
			{Name: "result", Kind: DataPortKindObject, Schema: "result.schema.json"},
			{Name: "log", Kind: DataPortKindFile},
		},
	}
	brokenEntity := Entity{
		ID:      "broken",
		Outputs: []DataPort{{Name: "result", Kind: DataPortKindObject, Schema: "missing.schema.json"}},
	}

	tests := []struct {
		name       string
		entity     Entity
		outputs    map[string]interface{}
		wantErr    bool
		wantValid  bool
		wantErrLen int
	}{
		{name: "valid output", entity: entity, outputs: map[string]interface{}{"result": map[string]interface{}{"id": "r1"}}, wantValid: true},
		{name: "missing output is skipped", entity: entity, outputs: map[string]interface{}{}, wantValid: true},
		{name: "unschema'd output is not validated", entity: entity, outputs: map[string]interface{}{"log": 42}, wantValid: true},
		{name: "invalid output", entity: entity, outputs: map[string]interface{}{"result": map[string]interface{}{}}, wantValid: false, wantErrLen: 1},
		{name: "schema cannot be loaded", entity: brokenEntity, outputs: map[string]interface{}{"result": map[string]interface{}{}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewPortSchemaValidator()
			validator.SetBasePath(tmpDir)

			result, err := validator.ValidateEntityOutputs(tt.entity, tt.outputs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateEntityOutputs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v (errors: %s)", result.Valid, tt.wantValid, result.Error())
			}
			if len(result.Errors) != tt.wantErrLen {
				t.Errorf("len(Errors) = %d, want %d", len(result.Errors), tt.wantErrLen)
			}
			for _, e := range result.Errors {
				if !strings.HasPrefix(e.Path, "result.") {
					t.Errorf("error path %q should be prefixed with the port name", e.Path)
				}
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		want   float64
		wantOK bool
	}{
		{name: "float64", input: 1.5, want: 1.5, wantOK: true},
		{name: "float32", input: float32(2.5), want: 2.5, wantOK: true},
		{name: "int", input: 3, want: 3, wantOK: true},
		{name: "int64", input: int64(-4), want: -4, wantOK: true},
		{name: "int32", input: int32(5), want: 5, wantOK: true},
		{name: "string", input: "6"},
		{name: "nil", input: nil},
		{name: "bool", input: true},
		{name: "uint", input: uint(7)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat64(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
