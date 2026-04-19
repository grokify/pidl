package pidl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFile(t *testing.T) {
	// Create a temp file with valid content
	dir := t.TempDir()
	filename := filepath.Join(dir, "test.json")

	p := NewMinimalProtocol("test", "Test Protocol")
	if err := WriteProtocolFile(filename, p); err != nil {
		t.Fatalf("WriteProtocolFile() error = %v", err)
	}

	// Validate it
	parsed, errs, err := ValidateFile(filename)
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}
	if errs.HasErrors() {
		t.Errorf("ValidateFile() validation errors = %v", errs)
	}
	if parsed.ProtocolMeta.ID != "test" {
		t.Errorf("Protocol ID = %q, want %q", parsed.ProtocolMeta.ID, "test")
	}
}

func TestValidateFileNotFound(t *testing.T) {
	_, _, err := ValidateFile("/nonexistent/file.json")
	if err == nil {
		t.Error("ValidateFile() should error on missing file")
	}
}

func TestValidateFileInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "invalid.json")

	if err := os.WriteFile(filename, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := ValidateFile(filename)
	if err == nil {
		t.Error("ValidateFile() should error on invalid JSON")
	}
}

func TestValidateFiles(t *testing.T) {
	dir := t.TempDir()

	// Create valid file
	valid := filepath.Join(dir, "valid.json")
	p := NewMinimalProtocol("valid", "Valid")
	if err := WriteProtocolFile(valid, p); err != nil {
		t.Fatal(err)
	}

	// Create invalid file
	invalid := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalid, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	results := ValidateFiles([]string{valid, invalid})

	if len(results) != 2 {
		t.Fatalf("ValidateFiles() returned %d results, want 2", len(results))
	}

	if !results[0].IsValid() {
		t.Errorf("results[0] should be valid")
	}

	if results[1].IsValid() {
		t.Errorf("results[1] should be invalid")
	}
}

func TestFileValidationResultIsValid(t *testing.T) {
	tests := []struct {
		name   string
		result FileValidationResult
		want   bool
	}{
		{
			name:   "valid",
			result: FileValidationResult{},
			want:   true,
		},
		{
			name: "parse error",
			result: FileValidationResult{
				ParseErr: os.ErrNotExist,
			},
			want: false,
		},
		{
			name: "validation errors",
			result: FileValidationResult{
				Errors: ValidationErrors{{Field: "x", Message: "y"}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewProtocol(t *testing.T) {
	p := NewProtocol("test-id", "Test Name")

	if p.ProtocolMeta.ID != "test-id" {
		t.Errorf("ID = %q, want %q", p.ProtocolMeta.ID, "test-id")
	}
	if p.ProtocolMeta.Name != "Test Name" {
		t.Errorf("Name = %q, want %q", p.ProtocolMeta.Name, "Test Name")
	}
}

func TestNewMinimalProtocol(t *testing.T) {
	p := NewMinimalProtocol("test", "Test")

	if !p.IsValid() {
		errs := p.Validate()
		t.Errorf("NewMinimalProtocol() should be valid, got errors: %v", errs)
	}

	if len(p.Entities) < 2 {
		t.Errorf("NewMinimalProtocol() should have at least 2 entities")
	}
	if len(p.Flows) < 1 {
		t.Errorf("NewMinimalProtocol() should have at least 1 flow")
	}
}

func TestWriteProtocolFile(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "subdir", "test.json")

	p := NewMinimalProtocol("test", "Test")
	if err := WriteProtocolFile(filename, p); err != nil {
		t.Fatalf("WriteProtocolFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); err != nil {
		t.Errorf("File should exist: %v", err)
	}

	// Read it back
	p2, err := ParseFile(filename)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if p2.ProtocolMeta.ID != p.ProtocolMeta.ID {
		t.Errorf("Round-trip ID = %q, want %q", p2.ProtocolMeta.ID, p.ProtocolMeta.ID)
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"hello", "Hello"},
		{"", ""},
		{"my protocol", "My Protocol"},
		{"oauth 2 0", "Oauth 2 0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TitleCase(tt.input)
			if got != tt.want {
				t.Errorf("TitleCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"test", "test"},
		{"Test", "test"},
		{"TEST", "test"},
		{"test-id", "test-id"},
		{"test_id", "test_id"},
		{"Test ID", "test_id"},
		{"My Protocol", "my_protocol"},
		{"123abc", "p123abc"},
		{"", "protocol"},
		{"---", "protocol"},
		{"OAuth 2.0", "oauth_2_0"},
		{"MCP Tool", "mcp_tool"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeID(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
