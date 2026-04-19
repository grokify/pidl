// Package examples provides embedded PIDL example protocols.
package examples

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/grokify/pidl"
)

//go:embed *.json
var examplesFS embed.FS

// Example represents an embedded example protocol.
type Example struct {
	// Name is the example name (filename without extension).
	Name string

	// Filename is the full filename.
	Filename string

	// Protocol is the parsed protocol (lazily loaded).
	protocol *pidl.Protocol
}

// Protocol returns the parsed protocol, loading it if necessary.
func (e *Example) Protocol() (*pidl.Protocol, error) {
	if e.protocol != nil {
		return e.protocol, nil
	}

	data, err := examplesFS.ReadFile(e.Filename)
	if err != nil {
		return nil, fmt.Errorf("reading example %s: %w", e.Name, err)
	}

	p, err := pidl.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing example %s: %w", e.Name, err)
	}

	e.protocol = p
	return p, nil
}

// List returns all available example names.
func List() []string {
	entries, err := examplesFS.ReadDir(".")
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") && !strings.HasPrefix(name, ".") {
			names = append(names, strings.TrimSuffix(name, ".json"))
		}
	}
	sort.Strings(names)
	return names
}

// Get returns an Example by name.
func Get(name string) (*Example, error) {
	filename := name
	if !strings.HasSuffix(filename, ".json") {
		filename = name + ".json"
	}

	data, err := examplesFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("example not found: %s", name)
	}

	// Parse to validate
	p, err := pidl.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing example %s: %w", name, err)
	}

	return &Example{
		Name:     strings.TrimSuffix(filepath.Base(filename), ".json"),
		Filename: filename,
		protocol: p,
	}, nil
}

// GetProtocol returns a parsed protocol by example name.
func GetProtocol(name string) (*pidl.Protocol, error) {
	ex, err := Get(name)
	if err != nil {
		return nil, err
	}
	return ex.Protocol()
}

// GetJSON returns the raw JSON for an example.
func GetJSON(name string) ([]byte, error) {
	filename := name
	if !strings.HasSuffix(filename, ".json") {
		filename = name + ".json"
	}
	return examplesFS.ReadFile(filename)
}

// All returns all examples.
func All() ([]*Example, error) {
	names := List()
	examples := make([]*Example, 0, len(names))
	for _, name := range names {
		ex, err := Get(name)
		if err != nil {
			return nil, err
		}
		examples = append(examples, ex)
	}
	return examples, nil
}
