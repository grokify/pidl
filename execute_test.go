package pidl

import (
	"testing"
)

func TestExecutor_NewContext(t *testing.T) {
	p := NewMinimalProtocol("test", "Test")
	exec := NewExecutor(p)
	ctx := exec.NewContext()

	if ctx.Protocol != p {
		t.Error("Context should reference the protocol")
	}

	if ctx.FlowIndex != 0 {
		t.Errorf("FlowIndex = %d, want 0", ctx.FlowIndex)
	}

	if ctx.Completed {
		t.Error("New context should not be completed")
	}

	if ctx.Trace == nil {
		t.Error("Trace should be initialized")
	}

	if ctx.Trace.ProtocolID != "test" {
		t.Errorf("Trace.ProtocolID = %q, want %q", ctx.Trace.ProtocolID, "test")
	}
}

func TestExecutor_Step(t *testing.T) {
	p := NewMinimalProtocol("test", "Test")
	exec := NewExecutor(p)
	ctx := exec.NewContext()

	// Execute first step
	step, err := exec.Step(ctx)
	if err != nil {
		t.Fatalf("Step() error = %v", err)
	}

	if step == nil {
		t.Fatal("Step should not be nil")
	}

	if step.StepNumber != 1 {
		t.Errorf("StepNumber = %d, want 1", step.StepNumber)
	}

	if step.FlowIndex != 0 {
		t.Errorf("FlowIndex = %d, want 0", step.FlowIndex)
	}

	if step.From != "client" {
		t.Errorf("From = %q, want %q", step.From, "client")
	}

	if step.To != "server" {
		t.Errorf("To = %q, want %q", step.To, "server")
	}

	if ctx.FlowIndex != 1 {
		t.Errorf("Context FlowIndex = %d, want 1", ctx.FlowIndex)
	}
}

func TestExecutor_Run(t *testing.T) {
	p := NewMinimalProtocol("test", "Test")
	exec := NewExecutor(p)
	ctx := exec.NewContext()

	trace, err := exec.Run(ctx)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !ctx.Completed {
		t.Error("Context should be completed after Run")
	}

	if !trace.Completed {
		t.Error("Trace should be completed after Run")
	}

	if trace.StepCount() != len(p.Flows) {
		t.Errorf("StepCount = %d, want %d", trace.StepCount(), len(p.Flows))
	}

	if trace.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}
}

func TestExecutor_RunN(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "step1"},
			{From: "b", To: "a", Action: "step2"},
			{From: "a", To: "b", Action: "step3"},
			{From: "b", To: "a", Action: "step4"},
		},
	}

	exec := NewExecutor(p)
	ctx := exec.NewContext()

	// Run only 2 steps
	trace, err := exec.RunN(ctx, 2)
	if err != nil {
		t.Fatalf("RunN() error = %v", err)
	}

	if trace.StepCount() != 2 {
		t.Errorf("StepCount = %d, want 2", trace.StepCount())
	}

	if ctx.Completed {
		t.Error("Context should not be completed after RunN(2)")
	}

	if ctx.FlowIndex != 2 {
		t.Errorf("FlowIndex = %d, want 2", ctx.FlowIndex)
	}

	// Run remaining
	_, err = exec.Run(ctx)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !ctx.Completed {
		t.Error("Context should be completed after Run")
	}
}

func TestExecutor_StateTransitions(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle", Initial: true},
					{ID: "requesting"},
					{ID: "authorized"},
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
				Sets: []StateMutation{
					{Entity: "client", To: "requesting"},
				},
			},
			{
				From:   "server",
				To:     "client",
				Action: "authorize",
				Sets: []StateMutation{
					{Entity: "client", From: "requesting", To: "authorized"},
				},
			},
		},
	}

	exec := NewExecutor(p)
	ctx := exec.NewContext()

	// Check initial state
	if ctx.GetEntityState("client") != "idle" {
		t.Errorf("Initial state = %q, want %q", ctx.GetEntityState("client"), "idle")
	}

	// Step 1: client -> requesting
	step1, err := exec.Step(ctx)
	if err != nil {
		t.Fatalf("Step() error = %v", err)
	}

	if len(step1.StateChanges) != 1 {
		t.Errorf("StateChanges count = %d, want 1", len(step1.StateChanges))
	}

	if ctx.GetEntityState("client") != "requesting" {
		t.Errorf("State after step1 = %q, want %q", ctx.GetEntityState("client"), "requesting")
	}

	// Step 2: client -> authorized
	step2, err := exec.Step(ctx)
	if err != nil {
		t.Fatalf("Step() error = %v", err)
	}

	if len(step2.StateChanges) != 1 {
		t.Errorf("StateChanges count = %d, want 1", len(step2.StateChanges))
	}

	if step2.StateChanges[0].FromState != "requesting" {
		t.Errorf("FromState = %q, want %q", step2.StateChanges[0].FromState, "requesting")
	}

	if ctx.GetEntityState("client") != "authorized" {
		t.Errorf("State after step2 = %q, want %q", ctx.GetEntityState("client"), "authorized")
	}
}

func TestExecutor_StateMismatch(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle", Initial: true},
					{ID: "active"},
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
				Sets: []StateMutation{
					{Entity: "client", From: "active", To: "done"}, // Wrong from state
				},
			},
		},
	}

	exec := NewExecutor(p)
	ctx := exec.NewContext()

	_, err := exec.Step(ctx)
	if err == nil {
		t.Error("Expected state mismatch error")
	}

	if ctx.Error == nil {
		t.Error("Context.Error should be set")
	}
}

func TestExecutor_ConditionalFlow(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "always"},
			{From: "b", To: "a", Action: "conditional", Condition: "success"},
			{From: "a", To: "b", Action: "final"},
		},
	}

	// Test with condition met
	exec := NewExecutor(p)
	exec.ConditionEvaluator = func(ctx *ExecutionContext, flow *Flow) bool {
		return flow.Condition == "success"
	}

	ctx := exec.NewContext()
	trace, _ := exec.Run(ctx)

	if trace.StepCount() != 3 {
		t.Errorf("StepCount = %d, want 3", trace.StepCount())
	}

	if trace.SkippedCount() != 0 {
		t.Errorf("SkippedCount = %d, want 0", trace.SkippedCount())
	}

	// Test with condition not met
	exec.ConditionEvaluator = func(ctx *ExecutionContext, flow *Flow) bool {
		return false
	}

	ctx = exec.NewContext()
	trace, _ = exec.Run(ctx)

	if trace.SkippedCount() != 1 {
		t.Errorf("SkippedCount = %d, want 1", trace.SkippedCount())
	}

	// Find the skipped step
	for _, step := range trace.Steps {
		if step.Action == "conditional" {
			if !step.Skipped {
				t.Error("Conditional step should be skipped")
			}
			if step.ConditionMet == nil || *step.ConditionMet {
				t.Error("ConditionMet should be false")
			}
		}
	}
}

func TestExecutor_Reset(t *testing.T) {
	p := NewMinimalProtocol("test", "Test")
	exec := NewExecutor(p)
	ctx := exec.NewContext()

	// Run to completion
	if _, err := exec.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !ctx.Completed {
		t.Error("Should be completed")
	}

	// Reset
	exec.Reset(ctx)

	if ctx.Completed {
		t.Error("Should not be completed after reset")
	}

	if ctx.FlowIndex != 0 {
		t.Errorf("FlowIndex = %d, want 0", ctx.FlowIndex)
	}

	if len(ctx.Trace.Steps) != 0 {
		t.Errorf("Steps = %d, want 0", len(ctx.Trace.Steps))
	}
}

func TestExecutionContext_Progress(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "1"},
			{From: "b", To: "a", Action: "2"},
			{From: "a", To: "b", Action: "3"},
			{From: "b", To: "a", Action: "4"},
		},
	}

	exec := NewExecutor(p)
	ctx := exec.NewContext()

	if ctx.Progress() != 0 {
		t.Errorf("Progress at start = %v, want 0", ctx.Progress())
	}

	if _, err := exec.Step(ctx); err != nil {
		t.Fatalf("Step() error = %v", err)
	}
	if ctx.Progress() != 25 {
		t.Errorf("Progress after 1 step = %v, want 25", ctx.Progress())
	}

	if _, err := exec.Step(ctx); err != nil {
		t.Fatalf("Step() error = %v", err)
	}
	if ctx.Progress() != 50 {
		t.Errorf("Progress after 2 steps = %v, want 50", ctx.Progress())
	}

	if _, err := exec.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if ctx.Progress() != 100 {
		t.Errorf("Progress at end = %v, want 100", ctx.Progress())
	}
}

func TestExecutionContext_CurrentFlow(t *testing.T) {
	p := NewMinimalProtocol("test", "Test")
	exec := NewExecutor(p)
	ctx := exec.NewContext()

	flow := ctx.CurrentFlow()
	if flow == nil {
		t.Fatal("CurrentFlow should not be nil")
	}

	if flow.Action != "request" {
		t.Errorf("CurrentFlow.Action = %q, want %q", flow.Action, "request")
	}

	if _, err := exec.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if ctx.CurrentFlow() != nil {
		t.Error("CurrentFlow should be nil after completion")
	}
}

func TestExecutionTrace_Statistics(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "s1", Initial: true},
					{ID: "s2"},
					{ID: "s3"},
				},
			},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "a", Sets: []StateMutation{{Entity: "client", To: "s2"}}},
			{From: "server", To: "client", Action: "b", Sets: []StateMutation{{Entity: "client", To: "s3"}}},
			{From: "client", To: "server", Action: "c"},
		},
	}

	exec := NewExecutor(p)
	ctx := exec.NewContext()
	trace, _ := exec.Run(ctx)

	if trace.StepCount() != 3 {
		t.Errorf("StepCount = %d, want 3", trace.StepCount())
	}

	if trace.StateChangeCount() != 2 {
		t.Errorf("StateChangeCount = %d, want 2", trace.StateChangeCount())
	}

	if trace.Duration() <= 0 {
		t.Error("Duration should be positive")
	}

	if trace.FinalStates["client"] != "s3" {
		t.Errorf("FinalStates[client] = %q, want %q", trace.FinalStates["client"], "s3")
	}
}

func TestExecutionContext_SetEntityState(t *testing.T) {
	p := NewMinimalProtocol("test", "Test")
	exec := NewExecutor(p)
	ctx := exec.NewContext()

	ctx.SetEntityState("client", "custom_state")

	if ctx.GetEntityState("client") != "custom_state" {
		t.Errorf("GetEntityState = %q, want %q", ctx.GetEntityState("client"), "custom_state")
	}
}

func TestExecutor_EmptyProtocol(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "empty", Name: "Empty"},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{},
	}

	exec := NewExecutor(p)
	ctx := exec.NewContext()

	step, err := exec.Step(ctx)
	if err != nil {
		t.Fatalf("Step() error = %v", err)
	}

	if step != nil {
		t.Error("Step should return nil for empty protocol")
	}

	if !ctx.Completed {
		t.Error("Should be completed immediately")
	}

	if ctx.Progress() != 100 {
		t.Errorf("Progress = %v, want 100", ctx.Progress())
	}
}
