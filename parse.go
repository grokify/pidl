package pidl

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ParseFile reads and parses a PIDL JSON file.
func ParseFile(filename string) (*Protocol, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return Parse(data)
}

// Parse parses PIDL JSON data into a Protocol.
func Parse(data []byte) (*Protocol, error) {
	var p Protocol
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return &p, nil
}

// ParseReader parses PIDL JSON from an io.Reader.
func ParseReader(r io.Reader) (*Protocol, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading: %w", err)
	}
	return Parse(data)
}

// MustParse parses PIDL JSON data and panics on error.
// Use only in tests or initialization code.
func MustParse(data []byte) *Protocol {
	p, err := Parse(data)
	if err != nil {
		panic(err)
	}
	return p
}

// MustParseFile reads and parses a PIDL JSON file, panicking on error.
// Use only in tests or initialization code.
func MustParseFile(filename string) *Protocol {
	p, err := ParseFile(filename)
	if err != nil {
		panic(err)
	}
	return p
}

// ToJSON serializes a Protocol to JSON with indentation.
func (p *Protocol) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// ToJSONCompact serializes a Protocol to compact JSON.
func (p *Protocol) ToJSONCompact() ([]byte, error) {
	return json.Marshal(p)
}

// WriteFile writes the Protocol to a file as JSON.
func (p *Protocol) WriteFile(filename string) error {
	data, err := p.ToJSON()
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}
