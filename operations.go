package pidl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateFile parses and validates a PIDL file, returning any validation errors.
func ValidateFile(filename string) (*Protocol, ValidationErrors, error) {
	p, err := ParseFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	errs := p.Validate()
	return p, errs, nil
}

// ValidateFiles validates multiple PIDL files, returning results for each.
func ValidateFiles(filenames []string) []FileValidationResult {
	results := make([]FileValidationResult, len(filenames))
	for i, f := range filenames {
		p, errs, err := ValidateFile(f)
		results[i] = FileValidationResult{
			Filename: f,
			Protocol: p,
			Errors:   errs,
			ParseErr: err,
		}
	}
	return results
}

// FileValidationResult contains the result of validating a single file.
type FileValidationResult struct {
	Filename string
	Protocol *Protocol
	Errors   ValidationErrors
	ParseErr error
}

// IsValid returns true if the file parsed and validated successfully.
func (r FileValidationResult) IsValid() bool {
	return r.ParseErr == nil && !r.Errors.HasErrors()
}

// NewProtocol creates a new Protocol with the given ID and name.
func NewProtocol(id, name string) *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   id,
			Name: name,
		},
		Entities: []Entity{},
		Phases:   []Phase{},
		Flows:    []Flow{},
	}
}

// NewMinimalProtocol creates a minimal valid protocol scaffold.
func NewMinimalProtocol(id, name string) *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:          id,
			Name:        name,
			Version:     "1.0",
			Description: "TODO: Add description",
			Category:    CategoryOther,
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "main", Name: "Main", Description: "Main protocol flow"},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request", Label: "Request", Mode: FlowModeRequest, Phase: "main"},
			{From: "server", To: "client", Action: "response", Label: "Response", Mode: FlowModeResponse, Phase: "main"},
		},
	}
}

// WriteProtocolFile writes a protocol to a file as formatted JSON.
func WriteProtocolFile(filename string, p *Protocol) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
	}

	data, err := p.ToJSON()
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	// Add trailing newline
	data = append(data, '\n')

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// SanitizeID converts a string to a valid PIDL ID (lowercase, alphanumeric, underscores/hyphens).
func SanitizeID(s string) string {
	var result []byte
	lastWasSeparator := true // Treat start as separator to avoid leading separators

	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z':
			result = append(result, byte(c))
			lastWasSeparator = false
		case c >= 'A' && c <= 'Z':
			result = append(result, byte(c-'A'+'a'))
			lastWasSeparator = false
		case c >= '0' && c <= '9':
			if len(result) == 0 {
				// IDs must start with letter
				result = append(result, 'p')
			}
			result = append(result, byte(c))
			lastWasSeparator = false
		case c == '-':
			// Preserve hyphens (valid in protocol IDs)
			if !lastWasSeparator && len(result) > 0 {
				result = append(result, '-')
				lastWasSeparator = true
			}
		case c == '_' || c == ' ' || c == '.':
			// Convert to underscore
			if !lastWasSeparator && len(result) > 0 {
				result = append(result, '_')
				lastWasSeparator = true
			}
		}
	}

	// Trim trailing separators
	for len(result) > 0 && (result[len(result)-1] == '_' || result[len(result)-1] == '-') {
		result = result[:len(result)-1]
	}

	if len(result) == 0 {
		return "protocol"
	}
	return string(result)
}

// TitleCase converts a string to title case (first letter of each word capitalized).
func TitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}
