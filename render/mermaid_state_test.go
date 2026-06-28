package render

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

func TestMermaidStateRenderer_Format(t *testing.T) {
	r := NewMermaidState()
	if r.Format() != FormatMermaidState {
		t.Errorf("Format() = %q, want %q", r.Format(), FormatMermaidState)
	}
}

func TestMermaidStateRenderer_NoStates(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{ID: "client", Name: "Client", Type: pidl.EntityTypeClient},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	r := NewMermaidState()
	output, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if !strings.Contains(output, "stateDiagram-v2") {
		t.Error("output should contain stateDiagram-v2")
	}
	if !strings.Contains(output, "No entities with states defined") {
		t.Error("output should indicate no entities with states")
	}
}

func TestMermaidStateRenderer_SingleEntity(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: pidl.EntityTypeClient,
				States: []pidl.EntityState{
					{ID: "idle", Name: "Idle", Initial: true},
					{ID: "active", Name: "Active"},
					{ID: "error", Name: "Error", Final: true},
				},
			},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{
				From:   "client",
				To:     "server",
				Action: "login",
				Label:  "Login Request",
				Sets: []pidl.StateMutation{
					{Entity: "client", From: "idle", To: "active"},
				},
			},
		},
	}

	r := NewMermaidState()
	output, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check diagram type
	if !strings.Contains(output, "stateDiagram-v2") {
		t.Error("output should contain stateDiagram-v2")
	}

	// Check states are declared
	if !strings.Contains(output, "state \"Idle\"") {
		t.Error("output should contain Idle state")
	}
	if !strings.Contains(output, "state \"Active\"") {
		t.Error("output should contain Active state")
	}
	if !strings.Contains(output, "state \"Error\"") {
		t.Error("output should contain Error state")
	}

	// Check initial state
	if !strings.Contains(output, "[*] --> idle") {
		t.Error("output should show initial state transition")
	}

	// Check transition
	if !strings.Contains(output, "idle --> active") {
		t.Error("output should contain idle to active transition")
	}
	if !strings.Contains(output, "Login Request") {
		t.Error("output should contain transition label")
	}

	// Check final state
	if !strings.Contains(output, "error --> [*]") {
		t.Error("output should show final state transition")
	}
}

func TestMermaidStateRenderer_MultipleEntities(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: pidl.EntityTypeClient,
				States: []pidl.EntityState{
					{ID: "idle", Initial: true},
					{ID: "active"},
				},
			},
			{
				ID:   "server",
				Name: "Server",
				Type: pidl.EntityTypeServer,
				States: []pidl.EntityState{
					{ID: "ready", Initial: true},
					{ID: "busy"},
				},
			},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	r := NewMermaidState()
	output, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Check both entities are present as containers
	if !strings.Contains(output, "state \"Client\" as client") {
		t.Error("output should contain Client container")
	}
	if !strings.Contains(output, "state \"Server\" as server") {
		t.Error("output should contain Server container")
	}

	// Check states use prefixed IDs to avoid collisions
	if !strings.Contains(output, "client_idle") {
		t.Error("output should contain client_idle prefixed state")
	}
	if !strings.Contains(output, "server_ready") {
		t.Error("output should contain server_ready prefixed state")
	}
}

func TestMermaidStateRenderer_EntityFilter(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: pidl.EntityTypeClient,
				States: []pidl.EntityState{
					{ID: "idle", Initial: true},
					{ID: "active"},
				},
			},
			{
				ID:   "server",
				Name: "Server",
				Type: pidl.EntityTypeServer,
				States: []pidl.EntityState{
					{ID: "ready", Initial: true},
					{ID: "busy"},
				},
			},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	r := NewMermaidState()
	r.EntityFilter = "client"
	output, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Should contain client states
	if !strings.Contains(output, "idle") {
		t.Error("output should contain client states")
	}

	// Should NOT contain server container or states
	if strings.Contains(output, "Server") {
		t.Error("output should not contain Server when filtering for client")
	}
	if strings.Contains(output, "server_ready") {
		t.Error("output should not contain server states when filtering for client")
	}
}

func TestMermaidStateRenderer_FilterNonExistentEntity(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: pidl.EntityTypeClient,
				States: []pidl.EntityState{
					{ID: "idle"},
				},
			},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	r := NewMermaidState()
	r.EntityFilter = "nonexistent"
	output, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if !strings.Contains(output, "No entities with states defined") {
		t.Error("output should indicate no matching entities")
	}
}

func TestMermaidStateRenderer_EscapeLabels(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: pidl.EntityTypeClient,
				States: []pidl.EntityState{
					{ID: "state_with_colon", Name: "State: With Colon"},
				},
			},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	r := NewMermaidState()
	output, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	// Colon should be escaped
	if !strings.Contains(output, "&#58;") {
		t.Error("output should escape colons in labels")
	}
}

func TestMermaidStateRenderer_StateDescription(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: pidl.EntityTypeClient,
				States: []pidl.EntityState{
					{ID: "idle", Name: "Idle", Description: "Waiting for user input"},
				},
			},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	r := NewMermaidState()
	r.ShowDescriptions = true
	output, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if !strings.Contains(output, "note right of") {
		t.Error("output should contain note for description")
	}
	if !strings.Contains(output, "Waiting for user input") {
		t.Error("output should contain state description")
	}
}

func TestMermaidStateRenderer_NoDescription(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: pidl.EntityTypeClient,
				States: []pidl.EntityState{
					{ID: "idle", Name: "Idle", Description: "Waiting for user input"},
				},
			},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	r := NewMermaidState()
	r.ShowDescriptions = false
	output, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if strings.Contains(output, "note right of") {
		t.Error("output should not contain notes when ShowDescriptions is false")
	}
}
