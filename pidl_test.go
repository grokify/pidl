package pidl

import (
	"testing"
)

func TestFlowDisplayLabel(t *testing.T) {
	tests := []struct {
		name     string
		flow     Flow
		expected string
	}{
		{
			name:     "label takes precedence",
			flow:     Flow{Action: "action", Label: "Label"},
			expected: "Label",
		},
		{
			name:     "falls back to action",
			flow:     Flow{Action: "action"},
			expected: "action",
		},
		{
			name:     "empty returns empty action",
			flow:     Flow{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flow.DisplayLabel(); got != tt.expected {
				t.Errorf("DisplayLabel() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFlowEffectiveMode(t *testing.T) {
	tests := []struct {
		name     string
		flow     Flow
		expected FlowMode
	}{
		{
			name:     "explicit mode",
			flow:     Flow{Mode: FlowModeResponse},
			expected: FlowModeResponse,
		},
		{
			name:     "defaults to request",
			flow:     Flow{},
			expected: FlowModeRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flow.EffectiveMode(); got != tt.expected {
				t.Errorf("EffectiveMode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProtocolEntityByID(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "client", Name: "Client"},
			{ID: "server", Name: "Server"},
		},
	}

	if e := p.EntityByID("client"); e == nil || e.Name != "Client" {
		t.Error("EntityByID should find client")
	}

	if e := p.EntityByID("unknown"); e != nil {
		t.Error("EntityByID should return nil for unknown ID")
	}
}

func TestProtocolPhaseByID(t *testing.T) {
	p := &Protocol{
		Phases: []Phase{
			{ID: "auth", Name: "Authorization"},
			{ID: "token", Name: "Token Exchange"},
		},
	}

	if ph := p.PhaseByID("auth"); ph == nil || ph.Name != "Authorization" {
		t.Error("PhaseByID should find auth phase")
	}

	if ph := p.PhaseByID("unknown"); ph != nil {
		t.Error("PhaseByID should return nil for unknown ID")
	}
}

func TestProtocolFlowsByPhase(t *testing.T) {
	p := &Protocol{
		Flows: []Flow{
			{From: "a", To: "b", Action: "x", Phase: "auth"},
			{From: "b", To: "c", Action: "y", Phase: "token"},
			{From: "a", To: "c", Action: "z", Phase: "auth"},
		},
	}

	authFlows := p.FlowsByPhase("auth")
	if len(authFlows) != 2 {
		t.Errorf("FlowsByPhase(auth) = %d flows, want 2", len(authFlows))
	}

	tokenFlows := p.FlowsByPhase("token")
	if len(tokenFlows) != 1 {
		t.Errorf("FlowsByPhase(token) = %d flows, want 1", len(tokenFlows))
	}

	unknownFlows := p.FlowsByPhase("unknown")
	if len(unknownFlows) != 0 {
		t.Errorf("FlowsByPhase(unknown) = %d flows, want 0", len(unknownFlows))
	}
}

func TestProtocolEntityIDs(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "a"},
			{ID: "b"},
			{ID: "c"},
		},
	}

	ids := p.EntityIDs()
	if len(ids) != 3 {
		t.Errorf("EntityIDs() = %d IDs, want 3", len(ids))
	}
	if ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
		t.Errorf("EntityIDs() = %v, want [a b c]", ids)
	}
}
