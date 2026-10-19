package pidl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	json := `{
		"protocol": {"id": "test", "name": "Test Protocol"},
		"entities": [
			{"id": "client", "name": "Client", "type": "client"},
			{"id": "server", "name": "Server", "type": "server"}
		],
		"flows": [
			{"from": "client", "to": "server", "action": "request"}
		]
	}`

	p, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if p.ProtocolMeta.ID != "test" {
		t.Errorf("Protocol.ID = %q, want %q", p.ProtocolMeta.ID, "test")
	}

	if len(p.Entities) != 2 {
		t.Errorf("len(Entities) = %d, want 2", len(p.Entities))
	}

	if len(p.Flows) != 1 {
		t.Errorf("len(Flows) = %d, want 1", len(p.Flows))
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := Parse([]byte("not json"))
	if err == nil {
		t.Error("Parse() should error on invalid JSON")
	}
}

func TestParseReader(t *testing.T) {
	json := `{
		"protocol": {"id": "test", "name": "Test"},
		"entities": [
			{"id": "a", "name": "A", "type": "client"},
			{"id": "b", "name": "B", "type": "server"}
		],
		"flows": [{"from": "a", "to": "b", "action": "x"}]
	}`

	p, err := ParseReader(strings.NewReader(json))
	if err != nil {
		t.Fatalf("ParseReader() error = %v", err)
	}

	if p.ProtocolMeta.ID != "test" {
		t.Errorf("Protocol.ID = %q, want %q", p.ProtocolMeta.ID, "test")
	}
}

func TestProtocolToJSON(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "request"},
		},
	}

	data, err := p.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Parse it back
	p2, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(ToJSON()) error = %v", err)
	}

	if p2.ProtocolMeta.ID != p.ProtocolMeta.ID {
		t.Errorf("Round-trip ID = %q, want %q", p2.ProtocolMeta.ID, p.ProtocolMeta.ID)
	}
}

func TestMustParse(t *testing.T) {
	json := `{
		"protocol": {"id": "test", "name": "Test"},
		"entities": [
			{"id": "a", "name": "A", "type": "client"},
			{"id": "b", "name": "B", "type": "server"}
		],
		"flows": [{"from": "a", "to": "b", "action": "x"}]
	}`

	// Should not panic
	p := MustParse([]byte(json))
	if p.ProtocolMeta.ID != "test" {
		t.Errorf("MustParse ID = %q, want %q", p.ProtocolMeta.ID, "test")
	}
}

func TestMustParsePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse should panic on invalid JSON")
		}
	}()
	MustParse([]byte("invalid"))
}

func TestMustParseFile(t *testing.T) {
	validPath := filepath.Join(t.TempDir(), "valid.json")
	if err := os.WriteFile(validPath, []byte(`{"protocol": {"id": "disk", "name": "From Disk"}}`), 0600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	invalidPath := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(invalidPath, []byte(`{not json`), 0600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantPanic bool
		wantID    string
	}{
		{name: "valid file", path: validPath, wantID: "disk"},
		{name: "missing file panics", path: filepath.Join(t.TempDir(), "missing.json"), wantPanic: true},
		{name: "invalid json panics", path: invalidPath, wantPanic: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("MustParseFile(%q) panic = %v, wantPanic %v", tt.path, r, tt.wantPanic)
				}
			}()
			p := MustParseFile(tt.path)
			if p.ProtocolMeta.ID != tt.wantID {
				t.Errorf("MustParseFile(%q).ID = %q, want %q", tt.path, p.ProtocolMeta.ID, tt.wantID)
			}
		})
	}
}

func TestProtocolToJSONCompact(t *testing.T) {
	tests := []struct {
		name string
		p    *Protocol
		want string
	}{
		{
			name: "minimal protocol",
			p:    &Protocol{ProtocolMeta: ProtocolMeta{ID: "c", Name: "Compact"}},
			want: `{"protocol":{"id":"c","name":"Compact"},"entities":null}`,
		},
		{
			name: "protocol with entity",
			p: &Protocol{
				ProtocolMeta: ProtocolMeta{ID: "c", Name: "Compact"},
				Entities:     []Entity{{ID: "e", Name: "E", Type: EntityTypeClient}},
			},
			want: `{"protocol":{"id":"c","name":"Compact"},"entities":[{"id":"e","name":"E","type":"client"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.p.ToJSONCompact()
			if err != nil {
				t.Fatalf("ToJSONCompact() error: %v", err)
			}
			if got := string(data); got != tt.want {
				t.Errorf("ToJSONCompact() = %s, want %s", got, tt.want)
			}
			if strings.Contains(string(data), "\n") {
				t.Error("ToJSONCompact() output should not contain newlines")
			}
		})
	}
}
