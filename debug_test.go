package pidl

import (
	"strings"
	"testing"
)

func createDebugTestProtocol() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "debug-test",
			Name: "Debug Test Protocol",
		},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle", Name: "Idle", Initial: true},
					{ID: "waiting", Name: "Waiting"},
					{ID: "done", Name: "Done", Final: true},
				},
			},
			{
				ID:   "server",
				Name: "Server",
				Type: EntityTypeServer,
			},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Sets:   []StateMutation{{Entity: "client", From: "idle", To: "waiting"}},
			},
			{
				From:   "server",
				To:     "client",
				Action: "process",
			},
			{
				From:   "server",
				To:     "client",
				Action: "response",
				Sets:   []StateMutation{{Entity: "client", From: "waiting", To: "done"}},
			},
		},
	}
}

func TestNewDebugSession(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	if session == nil {
		t.Fatal("expected non-nil session")
	}

	if session.Executor == nil {
		t.Error("expected non-nil executor")
	}

	if session.Context == nil {
		t.Error("expected non-nil context")
	}

	if len(session.Breakpoints) != 0 {
		t.Error("expected empty breakpoints")
	}
}

func TestDebugSession_Step(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	// Step once
	step, err := session.Step()
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}

	if step == nil {
		t.Fatal("expected non-nil step")
	}

	if step.StepNumber != 1 {
		t.Errorf("expected step number 1, got %d", step.StepNumber)
	}

	if step.From != "client" || step.To != "server" {
		t.Errorf("expected client->server, got %s->%s", step.From, step.To)
	}

	// Check state changed
	state := session.Inspect()
	if state.EntityStates["client"] != "waiting" {
		t.Errorf("expected client state 'waiting', got %s", state.EntityStates["client"])
	}
}

func TestDebugSession_StepToCompletion(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	// Step through all flows (3 flows + 1 nil to complete)
	stepCount := 0
	for !session.Inspect().IsCompleted {
		step, err := session.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}
		if step != nil {
			stepCount++
		}
		// Safety limit
		if stepCount > 10 {
			t.Fatal("too many steps")
		}
	}

	if stepCount != 3 {
		t.Errorf("expected 3 steps, got %d", stepCount)
	}

	// Should be complete now
	state := session.Inspect()
	if !state.IsCompleted {
		t.Error("expected execution to be completed")
	}

	// Stepping again should return nil
	step, err := session.Step()
	if err != nil {
		t.Fatalf("Step on completed session failed: %v", err)
	}
	if step != nil {
		t.Error("expected nil step after completion")
	}
}

func TestDebugSession_SetBreakpoint(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	// Set breakpoint
	err := session.SetBreakpoint(1, "")
	if err != nil {
		t.Fatalf("SetBreakpoint failed: %v", err)
	}

	if len(session.Breakpoints) != 1 {
		t.Errorf("expected 1 breakpoint, got %d", len(session.Breakpoints))
	}

	bp := session.Breakpoints[1]
	if bp == nil {
		t.Fatal("expected breakpoint at index 1")
	}

	if !bp.Enabled {
		t.Error("expected breakpoint to be enabled")
	}
}

func TestDebugSession_SetBreakpoint_OutOfRange(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	err := session.SetBreakpoint(100, "")
	if err == nil {
		t.Error("expected error for out of range breakpoint")
	}

	err = session.SetBreakpoint(-1, "")
	if err == nil {
		t.Error("expected error for negative breakpoint")
	}
}

func TestDebugSession_Continue(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	// Set breakpoint at flow 2
	err := session.SetBreakpoint(2, "")
	if err != nil {
		t.Fatalf("SetBreakpoint failed: %v", err)
	}

	// Continue should stop at breakpoint
	_, err = session.Continue()
	if err != nil {
		t.Fatalf("Continue failed: %v", err)
	}

	state := session.Inspect()
	if state.FlowIndex != 2 {
		t.Errorf("expected to stop at flow 2, got %d", state.FlowIndex)
	}

	if !state.AtBreakpoint {
		t.Error("expected to be at breakpoint")
	}

	// Continue should complete
	_, err = session.Continue()
	if err != nil {
		t.Fatalf("Continue failed: %v", err)
	}

	state = session.Inspect()
	if !state.IsCompleted {
		t.Error("expected execution to be completed")
	}
}

func TestDebugSession_RemoveBreakpoint(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	err := session.SetBreakpoint(1, "")
	if err != nil {
		t.Fatalf("SetBreakpoint failed: %v", err)
	}

	err = session.RemoveBreakpoint(1)
	if err != nil {
		t.Fatalf("RemoveBreakpoint failed: %v", err)
	}

	if len(session.Breakpoints) != 0 {
		t.Error("expected breakpoint to be removed")
	}

	// Removing non-existent breakpoint should error
	err = session.RemoveBreakpoint(1)
	if err == nil {
		t.Error("expected error for non-existent breakpoint")
	}
}

func TestDebugSession_EnableBreakpoint(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	err := session.SetBreakpoint(1, "")
	if err != nil {
		t.Fatalf("SetBreakpoint failed: %v", err)
	}

	err = session.EnableBreakpoint(1, false)
	if err != nil {
		t.Fatalf("EnableBreakpoint failed: %v", err)
	}

	if session.Breakpoints[1].Enabled {
		t.Error("expected breakpoint to be disabled")
	}

	err = session.EnableBreakpoint(1, true)
	if err != nil {
		t.Fatalf("EnableBreakpoint failed: %v", err)
	}

	if !session.Breakpoints[1].Enabled {
		t.Error("expected breakpoint to be enabled")
	}
}

func TestDebugSession_Inspect(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	state := session.Inspect()

	if state.FlowIndex != 0 {
		t.Errorf("expected flow index 0, got %d", state.FlowIndex)
	}

	if state.StepsExecuted != 0 {
		t.Errorf("expected 0 steps executed, got %d", state.StepsExecuted)
	}

	if state.IsCompleted {
		t.Error("expected not completed")
	}

	if state.CurrentFlow == nil {
		t.Error("expected non-nil current flow")
	}

	// Initial state should be set
	if state.EntityStates["client"] != "idle" {
		t.Errorf("expected client state 'idle', got %s", state.EntityStates["client"])
	}
}

func TestDebugSession_InspectEntity(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	entity, state, err := session.InspectEntity("client")
	if err != nil {
		t.Fatalf("InspectEntity failed: %v", err)
	}

	if entity == nil {
		t.Fatal("expected non-nil entity")
	}

	if entity.ID != "client" {
		t.Errorf("expected entity ID 'client', got %s", entity.ID)
	}

	if state != "idle" {
		t.Errorf("expected state 'idle', got %s", state)
	}

	// Non-existent entity
	_, _, err = session.InspectEntity("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent entity")
	}
}

func TestDebugSession_InspectFlow(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	flow, err := session.InspectFlow(0)
	if err != nil {
		t.Fatalf("InspectFlow failed: %v", err)
	}

	if flow == nil {
		t.Fatal("expected non-nil flow")
	}

	if flow.Action != "request" {
		t.Errorf("expected action 'request', got %s", flow.Action)
	}

	// Out of range
	_, err = session.InspectFlow(100)
	if err == nil {
		t.Error("expected error for out of range flow")
	}
}

func TestDebugSession_SetEntityState(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	err := session.SetEntityState("client", "waiting")
	if err != nil {
		t.Fatalf("SetEntityState failed: %v", err)
	}

	state := session.Inspect()
	if state.EntityStates["client"] != "waiting" {
		t.Errorf("expected client state 'waiting', got %s", state.EntityStates["client"])
	}

	// Invalid entity
	err = session.SetEntityState("nonexistent", "state")
	if err == nil {
		t.Error("expected error for non-existent entity")
	}

	// Invalid state
	err = session.SetEntityState("client", "invalid_state")
	if err == nil {
		t.Error("expected error for invalid state")
	}
}

func TestDebugSession_Reset(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	// Execute some steps
	_, _ = session.Step()
	_, _ = session.Step()

	// Set a breakpoint
	_ = session.SetBreakpoint(1, "")

	// Reset
	session.Reset()

	state := session.Inspect()
	if state.FlowIndex != 0 {
		t.Errorf("expected flow index 0 after reset, got %d", state.FlowIndex)
	}

	if state.StepsExecuted != 0 {
		t.Errorf("expected 0 steps after reset, got %d", state.StepsExecuted)
	}

	// Breakpoints should be preserved
	if len(session.Breakpoints) != 1 {
		t.Error("expected breakpoint to be preserved after reset")
	}
}

func TestDebugSession_ListFlows(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	// Set a breakpoint
	_ = session.SetBreakpoint(1, "")

	// Execute first step
	_, _ = session.Step()

	flows := session.ListFlows()
	if len(flows) != 3 {
		t.Errorf("expected 3 flows, got %d", len(flows))
	}

	// First flow should be executed
	if !flows[0].IsExecuted {
		t.Error("expected first flow to be executed")
	}

	// Second flow should be current and have breakpoint
	if !flows[1].IsCurrent {
		t.Error("expected second flow to be current")
	}
	if !flows[1].HasBreakpoint {
		t.Error("expected second flow to have breakpoint")
	}

	// Third flow should not be executed
	if flows[2].IsExecuted {
		t.Error("expected third flow to not be executed")
	}
}

func TestDebugSession_FormatFlowList(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	_ = session.SetBreakpoint(1, "")

	output := session.FormatFlowList()

	if output == "" {
		t.Error("expected non-empty output")
	}

	if !strings.Contains(output, "=>") {
		t.Error("expected current flow marker")
	}

	if !strings.Contains(output, "*") {
		t.Error("expected breakpoint marker")
	}
}

func TestDebugSession_ConditionalBreakpoint(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	// Set conditional breakpoint that should trigger
	err := session.SetBreakpoint(2, "client.state == \"waiting\"")
	if err != nil {
		t.Fatalf("SetBreakpoint failed: %v", err)
	}

	// Continue - should stop at breakpoint
	_, _ = session.Continue()

	state := session.Inspect()
	if state.FlowIndex != 2 {
		t.Errorf("expected to stop at flow 2, got %d", state.FlowIndex)
	}
}

func TestDebugState_String(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	state := session.Inspect()
	output := state.String()

	if output == "" {
		t.Error("expected non-empty output")
	}

	if !strings.Contains(output, "Flow Index") {
		t.Error("expected flow index in output")
	}

	if !strings.Contains(output, "Entity States") {
		t.Error("expected entity states in output")
	}
}

func TestDebugSession_ListBreakpoints(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	_ = session.SetBreakpoint(0, "")
	_ = session.SetBreakpoint(2, "condition")

	bps := session.ListBreakpoints()
	if len(bps) != 2 {
		t.Errorf("expected 2 breakpoints, got %d", len(bps))
	}
}

func TestDebugSession_Protocol(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	protocol := session.Protocol()
	if protocol == nil {
		t.Error("expected non-nil protocol")
	}

	if protocol.ProtocolMeta.ID != "debug-test" {
		t.Errorf("expected protocol ID 'debug-test', got %s", protocol.ProtocolMeta.ID)
	}
}

func TestDebugSession_Trace(t *testing.T) {
	p := createDebugTestProtocol()
	session := NewDebugSession(p)

	_, _ = session.Step()
	_, _ = session.Step()

	trace := session.Trace()
	if trace == nil {
		t.Error("expected non-nil trace")
	}

	if trace.StepCount() != 2 {
		t.Errorf("expected 2 steps in trace, got %d", trace.StepCount())
	}
}
