package pidl

import (
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
