package examples

import (
	"testing"
)

func TestList(t *testing.T) {
	names := List()

	if len(names) == 0 {
		t.Error("List() should return at least one example")
	}

	// Check for expected examples
	expected := []string{
		"a2a_agent_delegation",
		"mcp_tool_invocation",
		"oauth2_authorization_code",
		"oauth2_pkce",
		"oidc_authentication",
	}

	for _, name := range expected {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List() should contain %q", name)
		}
	}
}

func TestGet(t *testing.T) {
	ex, err := Get("oauth2_authorization_code")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if ex.Name != "oauth2_authorization_code" {
		t.Errorf("Name = %q, want %q", ex.Name, "oauth2_authorization_code")
	}

	p, err := ex.Protocol()
	if err != nil {
		t.Fatalf("Protocol() error = %v", err)
	}

	if p.ProtocolMeta.ID != "oauth2-authorization-code" {
		t.Errorf("Protocol ID = %q, want %q", p.ProtocolMeta.ID, "oauth2-authorization-code")
	}
}

func TestGetWithExtension(t *testing.T) {
	ex, err := Get("oauth2_authorization_code.json")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if ex.Name != "oauth2_authorization_code" {
		t.Errorf("Name = %q, want %q", ex.Name, "oauth2_authorization_code")
	}
}

func TestGetNotFound(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Error("Get() should error on nonexistent example")
	}
}

func TestGetProtocol(t *testing.T) {
	p, err := GetProtocol("mcp_tool_invocation")
	if err != nil {
		t.Fatalf("GetProtocol() error = %v", err)
	}

	if p.ProtocolMeta.ID != "mcp-tool-invocation" {
		t.Errorf("Protocol ID = %q, want %q", p.ProtocolMeta.ID, "mcp-tool-invocation")
	}
}

func TestGetJSON(t *testing.T) {
	data, err := GetJSON("oauth2_pkce")
	if err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("GetJSON() returned empty data")
	}

	// Should start with {
	if data[0] != '{' {
		t.Errorf("GetJSON() should return JSON, got first byte %q", data[0])
	}
}

func TestGetJSONNotFound(t *testing.T) {
	_, err := GetJSON("nonexistent")
	if err == nil {
		t.Error("GetJSON() should error on nonexistent example")
	}
}

func TestAll(t *testing.T) {
	examples, err := All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}

	if len(examples) < 5 {
		t.Errorf("All() returned %d examples, want at least 5", len(examples))
	}

	// Verify each example is valid
	for _, ex := range examples {
		p, err := ex.Protocol()
		if err != nil {
			t.Errorf("Example %s Protocol() error = %v", ex.Name, err)
			continue
		}

		if !p.IsValid() {
			errs := p.Validate()
			t.Errorf("Example %s is not valid: %v", ex.Name, errs)
		}
	}
}

func TestExampleProtocolCaching(t *testing.T) {
	ex, err := Get("oauth2_authorization_code")
	if err != nil {
		t.Fatal(err)
	}

	// First call
	p1, err := ex.Protocol()
	if err != nil {
		t.Fatal(err)
	}

	// Second call should return cached
	p2, err := ex.Protocol()
	if err != nil {
		t.Fatal(err)
	}

	// Should be same pointer
	if p1 != p2 {
		t.Error("Protocol() should return cached instance")
	}
}
