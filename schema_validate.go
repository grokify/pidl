package pidl

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// SchemaRegistry holds loaded JSON schemas for validation.
type SchemaRegistry struct {
	// schemas maps schema URIs to their parsed definitions.
	schemas map[string]interface{}
	// basePath is the base directory for resolving relative schema paths.
	basePath string
}

// NewSchemaRegistry creates a new schema registry.
func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]interface{}),
	}
}

// SetBasePath sets the base directory for resolving relative schema paths.
func (r *SchemaRegistry) SetBasePath(path string) {
	r.basePath = path
}

// LoadSchema loads a JSON Schema from a URI or file path.
// Supported URI formats:
//   - file:///path/to/schema.json
//   - ./relative/path/schema.json
//   - /absolute/path/schema.json
//   - #/definitions/TypeName (inline reference)
func (r *SchemaRegistry) LoadSchema(uri string) (interface{}, error) {
	// Check cache first
	if schema, ok := r.schemas[uri]; ok {
		return schema, nil
	}

	var schemaData []byte
	var err error

	// Handle different URI formats
	if strings.HasPrefix(uri, "file://") {
		// file:// URI
		path := strings.TrimPrefix(uri, "file://")
		schemaData, err = os.ReadFile(path)
	} else if strings.HasPrefix(uri, "#") {
		// Inline reference - not a loadable schema
		return nil, fmt.Errorf("inline references (#) cannot be loaded directly: %s", uri)
	} else if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		// HTTP URI - not supported in basic implementation
		return nil, fmt.Errorf("HTTP schema URIs not supported: %s", uri)
	} else {
		// Treat as file path
		path := uri
		if !filepath.IsAbs(path) && r.basePath != "" {
			path = filepath.Join(r.basePath, path)
		}
		schemaData, err = os.ReadFile(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load schema %s: %w", uri, err)
	}

	var schema interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema %s: %w", uri, err)
	}

	// Cache the schema
	r.schemas[uri] = schema
	return schema, nil
}

// RegisterSchema registers a schema with a given URI.
func (r *SchemaRegistry) RegisterSchema(uri string, schema interface{}) {
	r.schemas[uri] = schema
}

// SchemaValidationResult holds the result of schema validation.
type SchemaValidationResult struct {
	// Valid indicates if the data passed validation.
	Valid bool
	// Errors contains validation error messages.
	Errors []SchemaValidationError
}

// SchemaValidationError represents a single validation error.
type SchemaValidationError struct {
	// Path is the JSON path to the invalid field.
	Path string
	// Message describes the validation failure.
	Message string
	// SchemaPath is the path in the schema that failed.
	SchemaPath string
}

// Error returns the validation errors as a string.
func (r *SchemaValidationResult) Error() string {
	if r.Valid {
		return ""
	}
	var msgs []string
	for _, e := range r.Errors {
		if e.Path != "" {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Path, e.Message))
		} else {
			msgs = append(msgs, e.Message)
		}
	}
	return strings.Join(msgs, "; ")
}

// ValidateData validates data against a JSON Schema.
// This is a basic implementation that checks type constraints.
// For full JSON Schema validation, use a dedicated library like
// github.com/santhosh-tekuri/jsonschema/v5.
func ValidateData(data interface{}, schema interface{}) *SchemaValidationResult {
	result := &SchemaValidationResult{Valid: true}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, SchemaValidationError{
			Message: "invalid schema: expected object",
		})
		return result
	}

	validateValue("", data, schemaMap, result)
	return result
}

func validateValue(path string, data interface{}, schema map[string]interface{}, result *SchemaValidationResult) {
	// Check type constraint
	if schemaType, ok := schema["type"].(string); ok {
		if !validateType(data, schemaType) {
			result.Valid = false
			result.Errors = append(result.Errors, SchemaValidationError{
				Path:    path,
				Message: fmt.Sprintf("expected type %s, got %T", schemaType, data),
			})
			return
		}
	}

	// Check required fields for objects
	if data != nil {
		dataMap, isObject := data.(map[string]interface{})
		if isObject {
			if required, ok := schema["required"].([]interface{}); ok {
				for _, req := range required {
					reqStr, _ := req.(string)
					if _, exists := dataMap[reqStr]; !exists {
						result.Valid = false
						result.Errors = append(result.Errors, SchemaValidationError{
							Path:    joinPath(path, reqStr),
							Message: "required field missing",
						})
					}
				}
			}

			// Validate properties
			if properties, ok := schema["properties"].(map[string]interface{}); ok {
				for propName, propSchema := range properties {
					propSchemaMap, _ := propSchema.(map[string]interface{})
					if propValue, exists := dataMap[propName]; exists && propSchemaMap != nil {
						validateValue(joinPath(path, propName), propValue, propSchemaMap, result)
					}
				}
			}
		}

		// Check array items
		dataArr, isArray := data.([]interface{})
		if isArray {
			if items, ok := schema["items"].(map[string]interface{}); ok {
				for i, item := range dataArr {
					validateValue(fmt.Sprintf("%s[%d]", path, i), item, items, result)
				}
			}
		}
	}

	// Check enum constraint
	if enum, ok := schema["enum"].([]interface{}); ok {
		found := false
		for _, e := range enum {
			if data == e {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, SchemaValidationError{
				Path:    path,
				Message: fmt.Sprintf("value not in enum: %v", data),
			})
		}
	}

	// Check minimum/maximum for numbers
	if num, ok := toFloat64(data); ok {
		if min, ok := schema["minimum"].(float64); ok && num < min {
			result.Valid = false
			result.Errors = append(result.Errors, SchemaValidationError{
				Path:    path,
				Message: fmt.Sprintf("value %v is less than minimum %v", num, min),
			})
		}
		if max, ok := schema["maximum"].(float64); ok && num > max {
			result.Valid = false
			result.Errors = append(result.Errors, SchemaValidationError{
				Path:    path,
				Message: fmt.Sprintf("value %v is greater than maximum %v", num, max),
			})
		}
	}

	// Check minLength/maxLength for strings
	if str, ok := data.(string); ok {
		if minLen, ok := schema["minLength"].(float64); ok && float64(len(str)) < minLen {
			result.Valid = false
			result.Errors = append(result.Errors, SchemaValidationError{
				Path:    path,
				Message: fmt.Sprintf("string length %d is less than minLength %v", len(str), minLen),
			})
		}
		if maxLen, ok := schema["maxLength"].(float64); ok && float64(len(str)) > maxLen {
			result.Valid = false
			result.Errors = append(result.Errors, SchemaValidationError{
				Path:    path,
				Message: fmt.Sprintf("string length %d is greater than maxLength %v", len(str), maxLen),
			})
		}
	}
}

func validateType(data interface{}, schemaType string) bool {
	if data == nil {
		return schemaType == "null"
	}

	switch schemaType {
	case "string":
		_, ok := data.(string)
		return ok
	case "number":
		_, ok := toFloat64(data)
		return ok
	case "integer":
		switch v := data.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true
		case float64:
			return v == float64(int64(v))
		case float32:
			return v == float32(int32(v))
		}
		return false
	case "boolean":
		_, ok := data.(bool)
		return ok
	case "array":
		_, ok := data.([]interface{})
		return ok
	case "object":
		_, ok := data.(map[string]interface{})
		return ok
	case "null":
		return data == nil
	default:
		return true
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}

func joinPath(base, field string) string {
	if base == "" {
		return field
	}
	return base + "." + field
}

// PortSchemaValidator validates data ports against their schemas.
type PortSchemaValidator struct {
	registry *SchemaRegistry
}

// NewPortSchemaValidator creates a new port schema validator.
func NewPortSchemaValidator() *PortSchemaValidator {
	return &PortSchemaValidator{
		registry: NewSchemaRegistry(),
	}
}

// SetBasePath sets the base directory for resolving relative schema paths.
func (v *PortSchemaValidator) SetBasePath(path string) {
	v.registry.SetBasePath(path)
}

// ValidatePortData validates data against a port's schema.
func (v *PortSchemaValidator) ValidatePortData(port DataPort, data interface{}) (*SchemaValidationResult, error) {
	if port.Schema == "" {
		// No schema defined, validation passes
		return &SchemaValidationResult{Valid: true}, nil
	}

	schema, err := v.registry.LoadSchema(port.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema for port %s: %w", port.Name, err)
	}

	return ValidateData(data, schema), nil
}

// ValidateEntityInputs validates all input data for an entity.
func (v *PortSchemaValidator) ValidateEntityInputs(entity Entity, inputs map[string]interface{}) (*SchemaValidationResult, error) {
	result := &SchemaValidationResult{Valid: true}

	for _, port := range entity.Inputs {
		data, exists := inputs[port.Name]

		// Check required
		if port.Required && !exists {
			result.Valid = false
			result.Errors = append(result.Errors, SchemaValidationError{
				Path:    port.Name,
				Message: "required input missing",
			})
			continue
		}

		if !exists {
			continue
		}

		// Validate against schema if defined
		if port.Schema != "" {
			portResult, err := v.ValidatePortData(port, data)
			if err != nil {
				return nil, err
			}
			if !portResult.Valid {
				result.Valid = false
				for _, e := range portResult.Errors {
					e.Path = port.Name + "." + e.Path
					result.Errors = append(result.Errors, e)
				}
			}
		}
	}

	return result, nil
}

// ValidateEntityOutputs validates all output data from an entity.
func (v *PortSchemaValidator) ValidateEntityOutputs(entity Entity, outputs map[string]interface{}) (*SchemaValidationResult, error) {
	result := &SchemaValidationResult{Valid: true}

	for _, port := range entity.Outputs {
		data, exists := outputs[port.Name]

		if !exists {
			continue
		}

		// Validate against schema if defined
		if port.Schema != "" {
			portResult, err := v.ValidatePortData(port, data)
			if err != nil {
				return nil, err
			}
			if !portResult.Valid {
				result.Valid = false
				for _, e := range portResult.Errors {
					e.Path = port.Name + "." + e.Path
					result.Errors = append(result.Errors, e)
				}
			}
		}
	}

	return result, nil
}

// IsValidSchemaURI checks if a schema URI is syntactically valid.
func IsValidSchemaURI(uri string) bool {
	if uri == "" {
		return true // Empty is valid (no schema)
	}

	// Check for inline reference
	if strings.HasPrefix(uri, "#") {
		return true
	}

	// Check for file path
	if strings.HasPrefix(uri, "./") || strings.HasPrefix(uri, "../") || strings.HasPrefix(uri, "/") {
		return true
	}

	// Check for file:// URI
	if strings.HasPrefix(uri, "file://") {
		return true
	}

	// Check for http/https URI
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		_, err := url.Parse(uri)
		return err == nil
	}

	// Assume it's a relative file path
	return true
}
