package pidl

import (
	"reflect"
	"testing"
)

func createTestProcessSpec() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-process",
			Name: "Test Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "extract",
				Name:     "Data Extractor",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Outputs: []DataPort{
					{Kind: DataPortKindFile, Name: "raw_data", Required: false},
				},
			},
			{
				ID:       "transform",
				Name:     "Data Transformer",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindFile, Name: "raw_data", Required: true},
				},
				Outputs: []DataPort{
					{Kind: DataPortKindFile, Name: "transformed_data", Required: false},
				},
			},
			{
				ID:       "load",
				Name:     "Data Loader",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindFile, Name: "transformed_data", Required: true},
				},
				Outputs: []DataPort{
					{Kind: DataPortKindObject, Name: "load_result", Required: false},
				},
			},
		},
		Flows: []Flow{
			{From: "extract", To: "transform", Action: "send", Label: "raw_data"},
			{From: "transform", To: "load", Action: "send", Label: "transformed_data"},
		},
	}
}

func createParallelProcessSpec() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "parallel-process",
			Name: "Parallel Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "source",
				Name:     "Source",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Outputs: []DataPort{
					{Kind: DataPortKindObject, Name: "data"},
				},
			},
			{
				ID:       "branch_a",
				Name:     "Branch A",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindObject, Name: "data", Required: true},
				},
				Outputs: []DataPort{
					{Kind: DataPortKindObject, Name: "result_a"},
				},
			},
			{
				ID:       "branch_b",
				Name:     "Branch B",
				Type:     EntityTypeServer,
				StepType: StepTypeLLM,
				Inputs: []DataPort{
					{Kind: DataPortKindObject, Name: "data", Required: true},
				},
				Outputs: []DataPort{
					{Kind: DataPortKindObject, Name: "result_b"},
				},
			},
			{
				ID:       "merge",
				Name:     "Merge",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindObject, Name: "result_a", Required: true},
					{Kind: DataPortKindObject, Name: "result_b", Required: true},
				},
			},
		},
		Flows: []Flow{
			{From: "source", To: "branch_a", Action: "process"},
			{From: "source", To: "branch_b", Action: "process"},
			{From: "branch_a", To: "merge", Action: "send", Label: "result_a"},
			{From: "branch_b", To: "merge", Action: "send", Label: "result_b"},
		},
	}
}

func TestNewProcessExecutor(t *testing.T) {
	p := createTestProcessSpec()
	pe, err := NewProcessExecutor(p)
	if err != nil {
		t.Fatalf("NewProcessExecutor failed: %v", err)
	}

	if pe.Protocol != p {
		t.Error("expected protocol to be set")
	}

	// Check dependency graph
	if len(pe.Dependencies["transform"]) != 1 || pe.Dependencies["transform"][0] != "extract" {
		t.Errorf("expected transform to depend on extract, got %v", pe.Dependencies["transform"])
	}

	if len(pe.Dependencies["load"]) != 1 || pe.Dependencies["load"][0] != "transform" {
		t.Errorf("expected load to depend on transform, got %v", pe.Dependencies["load"])
	}
}

func TestNewProcessExecutor_NonProcessSpec(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "regular-protocol",
			Name: "Regular Protocol",
			Kind: ProtocolKindProtocol,
		},
	}

	_, err := NewProcessExecutor(p)
	if err == nil {
		t.Error("expected error for non-process spec")
	}
}

func TestTopologicalSort(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)

	order := pe.TopologicalSort()

	// Extract should come before transform, transform before load
	extractIdx, transformIdx, loadIdx := -1, -1, -1
	for i, id := range order {
		switch id {
		case "extract":
			extractIdx = i
		case "transform":
			transformIdx = i
		case "load":
			loadIdx = i
		}
	}

	if extractIdx >= transformIdx {
		t.Error("extract should come before transform")
	}
	if transformIdx >= loadIdx {
		t.Error("transform should come before load")
	}
}

func TestTopologicalSort_Parallel(t *testing.T) {
	p := createParallelProcessSpec()
	pe, _ := NewProcessExecutor(p)

	order := pe.TopologicalSort()

	// Source should come first, merge should come last
	if order[0] != "source" {
		t.Errorf("expected source first, got %s", order[0])
	}

	// Merge should be last
	if order[len(order)-1] != "merge" {
		t.Errorf("expected merge last, got %s", order[len(order)-1])
	}
}

func TestProcessExecutionContext_InitialState(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	// All steps should start as pending or ready
	for id, status := range ctx.StepStatus {
		if status != StepStatusPending && status != StepStatusReady && status != StepStatusBlocked {
			t.Errorf("step %s has unexpected initial status: %s", id, status)
		}
	}

	// Extract should be ready (no dependencies)
	if ctx.StepStatus["extract"] == StepStatusBlocked {
		t.Error("extract should not be blocked initially")
	}

	// Transform should be blocked (needs raw_data from extract)
	if ctx.StepStatus["transform"] != StepStatusBlocked {
		t.Errorf("transform should be blocked, got %s", ctx.StepStatus["transform"])
	}
}

func TestGetBlockedSteps(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	blocked := pe.GetBlockedSteps(ctx)

	// Transform should be blocked
	if _, ok := blocked["transform"]; !ok {
		t.Error("expected transform to be blocked")
	}

	// Load should be blocked
	if _, ok := blocked["load"]; !ok {
		t.Error("expected load to be blocked")
	}

	// Extract should NOT be blocked
	if _, ok := blocked["extract"]; ok {
		t.Error("extract should not be blocked")
	}
}

func TestGetReadySteps(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	ready := pe.GetReadySteps(ctx)

	// Only extract should be ready initially
	found := false
	for _, id := range ready {
		if id == "extract" {
			found = true
		}
		if id == "transform" || id == "load" {
			t.Errorf("step %s should not be ready initially", id)
		}
	}

	if !found {
		t.Error("extract should be ready")
	}
}

func TestCompleteStep(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	// Complete extract
	err := pe.CompleteStep(ctx, "extract")
	if err != nil {
		t.Fatalf("CompleteStep failed: %v", err)
	}

	// Extract should be completed
	if ctx.StepStatus["extract"] != StepStatusCompleted {
		t.Errorf("expected extract to be completed, got %s", ctx.StepStatus["extract"])
	}

	// raw_data output should be produced
	if !ctx.ProducedOutputs["extract.raw_data"] {
		t.Error("expected raw_data to be produced")
	}

	// Transform should now be ready
	ready := pe.GetReadySteps(ctx)
	found := false
	for _, id := range ready {
		if id == "transform" {
			found = true
			break
		}
	}
	if !found {
		t.Error("transform should be ready after extract completes")
	}
}

func TestMarkInputAvailable(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	// Initially transform is blocked
	if ctx.StepStatus["transform"] != StepStatusBlocked {
		t.Fatalf("expected transform to be blocked initially")
	}

	// Complete extract to make raw_data available
	if err := pe.CompleteStep(ctx, "extract"); err != nil {
		t.Fatalf("CompleteStep failed: %v", err)
	}

	// Now transform should be ready
	if ctx.StepStatus["transform"] == StepStatusBlocked {
		t.Error("transform should not be blocked after extract completes")
	}
}

func TestStartStep(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	// Start extract (should work - it's ready)
	err := pe.StartStep(ctx, "extract")
	if err != nil {
		t.Fatalf("StartStep failed: %v", err)
	}

	if ctx.StepStatus["extract"] != StepStatusInProgress {
		t.Errorf("expected in_progress, got %s", ctx.StepStatus["extract"])
	}

	// Try to start transform (should fail - it's blocked)
	err = pe.StartStep(ctx, "transform")
	if err == nil {
		t.Error("expected error when starting blocked step")
	}
}

func TestFailStep(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	err := pe.FailStep(ctx, "extract")
	if err != nil {
		t.Fatalf("FailStep failed: %v", err)
	}

	if ctx.StepStatus["extract"] != StepStatusFailed {
		t.Errorf("expected failed, got %s", ctx.StepStatus["extract"])
	}
}

func TestAnalyzeDependencies(t *testing.T) {
	p := createParallelProcessSpec()
	pe, _ := NewProcessExecutor(p)

	analyses := pe.AnalyzeDependencies()

	if len(analyses) != 4 {
		t.Errorf("expected 4 analyses, got %d", len(analyses))
	}

	// Find merge analysis
	var mergeAnalysis *DependencyAnalysis
	for i := range analyses {
		if analyses[i].EntityID == "merge" {
			mergeAnalysis = &analyses[i]
			break
		}
	}

	if mergeAnalysis == nil {
		t.Fatal("merge analysis not found")
	}

	// Merge should have 2 direct dependencies
	if len(mergeAnalysis.DirectDependencies) != 2 {
		t.Errorf("expected 2 direct dependencies, got %d", len(mergeAnalysis.DirectDependencies))
	}

	// Merge should have 3 transitive dependencies (source, branch_a, branch_b)
	if len(mergeAnalysis.TransitiveDependencies) != 3 {
		t.Errorf("expected 3 transitive dependencies, got %d: %v",
			len(mergeAnalysis.TransitiveDependencies), mergeAnalysis.TransitiveDependencies)
	}
}

func TestGetExecutionReadiness(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	readiness := pe.GetExecutionReadiness(ctx)

	if readiness.TotalSteps != 3 {
		t.Errorf("expected 3 total steps, got %d", readiness.TotalSteps)
	}

	if readiness.BlockedSteps != 2 {
		t.Errorf("expected 2 blocked steps, got %d", readiness.BlockedSteps)
	}

	// Extract should be in next steps
	found := false
	for _, id := range readiness.NextSteps {
		if id == "extract" {
			found = true
			break
		}
	}
	if !found {
		t.Error("extract should be in next steps")
	}
}

func TestExecutionFlow(t *testing.T) {
	p := createTestProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	// Execute extract
	_ = pe.StartStep(ctx, "extract")
	_ = pe.CompleteStep(ctx, "extract")

	// Transform should now be ready
	readiness := pe.GetExecutionReadiness(ctx)
	if readiness.CompletedSteps != 1 {
		t.Errorf("expected 1 completed step, got %d", readiness.CompletedSteps)
	}

	found := false
	for _, id := range readiness.NextSteps {
		if id == "transform" {
			found = true
			break
		}
	}
	if !found {
		t.Error("transform should be ready after extract completes")
	}

	// Execute transform
	_ = pe.StartStep(ctx, "transform")
	_ = pe.CompleteStep(ctx, "transform")

	// Load should now be ready
	readiness = pe.GetExecutionReadiness(ctx)
	if readiness.CompletedSteps != 2 {
		t.Errorf("expected 2 completed steps, got %d", readiness.CompletedSteps)
	}

	// Execute load
	_ = pe.StartStep(ctx, "load")
	_ = pe.CompleteStep(ctx, "load")

	// All should be complete
	readiness = pe.GetExecutionReadiness(ctx)
	if readiness.CompletedSteps != 3 {
		t.Errorf("expected 3 completed steps, got %d", readiness.CompletedSteps)
	}
	if readiness.BlockedSteps != 0 {
		t.Errorf("expected 0 blocked steps, got %d", readiness.BlockedSteps)
	}
}

func TestParallelExecution(t *testing.T) {
	p := createParallelProcessSpec()
	pe, _ := NewProcessExecutor(p)
	ctx := pe.NewProcessContext()

	// Complete source
	_ = pe.CompleteStep(ctx, "source")

	// Both branches should be ready
	readiness := pe.GetExecutionReadiness(ctx)
	if readiness.ReadySteps < 2 {
		t.Errorf("expected at least 2 ready steps after source, got %d: %v",
			readiness.ReadySteps, readiness.NextSteps)
	}

	// Complete both branches
	_ = pe.CompleteStep(ctx, "branch_a")
	_ = pe.CompleteStep(ctx, "branch_b")

	// Merge should now be ready
	readiness = pe.GetExecutionReadiness(ctx)
	found := false
	for _, id := range readiness.NextSteps {
		if id == "merge" {
			found = true
			break
		}
	}
	if !found {
		t.Error("merge should be ready after both branches complete")
	}
}

// externalInputProcessSpec returns a two-step process where the second step
// needs both the first step's output and an externally supplied "config"
// input that no step produces.
func externalInputProcessSpec() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "ext-input", Name: "External Input", Kind: ProtocolKindProcess},
		Entities: []Entity{
			{
				ID: "extract", Name: "Extract", Type: EntityTypeServer, StepType: StepTypeDeterministic,
				Outputs: []DataPort{{Kind: DataPortKindFile, Name: "raw_data"}},
			},
			{
				ID: "transform", Name: "Transform", Type: EntityTypeServer, StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindFile, Name: "raw_data", Required: true},
					{Kind: DataPortKindObject, Name: "config", Required: true},
				},
			},
		},
		// The label names the raw_data port. It must link only raw_data to
		// raw_data; "config" has no producer and stays externally supplied.
		Flows: []Flow{{From: "extract", To: "transform", Action: "send", Label: "raw_data"}},
	}
}

func TestProcessExecutor_MarkInputAvailable(t *testing.T) {
	tests := []struct {
		name             string
		completeFirst    bool
		entityID         string
		inputName        string
		wantStatus       StepExecutionStatus
		wantBlockedOn    []string
		wantAvailableKey string
	}{
		{
			name:             "external input alone keeps step blocked on predecessor",
			entityID:         "transform",
			inputName:        "config",
			wantStatus:       StepStatusBlocked,
			wantBlockedOn:    []string{"raw_data", "step:extract"},
			wantAvailableKey: "transform.config",
		},
		{
			name:             "external input after predecessor completes makes step ready",
			completeFirst:    true,
			entityID:         "transform",
			inputName:        "config",
			wantStatus:       StepStatusReady,
			wantAvailableKey: "transform.config",
		},
		{
			name:             "unknown port is recorded but step stays blocked",
			completeFirst:    true,
			entityID:         "transform",
			inputName:        "does_not_exist",
			wantStatus:       StepStatusBlocked,
			wantBlockedOn:    []string{"config"},
			wantAvailableKey: "transform.does_not_exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe, err := NewProcessExecutor(externalInputProcessSpec())
			if err != nil {
				t.Fatalf("NewProcessExecutor failed: %v", err)
			}
			ctx := pe.NewProcessContext()
			if tt.completeFirst {
				if err := pe.CompleteStep(ctx, "extract"); err != nil {
					t.Fatalf("CompleteStep failed: %v", err)
				}
			}

			pe.MarkInputAvailable(ctx, tt.entityID, tt.inputName)

			if !ctx.AvailableInputs[tt.wantAvailableKey] {
				t.Errorf("AvailableInputs[%q] = false, want true", tt.wantAvailableKey)
			}
			if got := ctx.StepStatus[tt.entityID]; got != tt.wantStatus {
				t.Errorf("StepStatus[%q] = %q, want %q", tt.entityID, got, tt.wantStatus)
			}
			if got := ctx.BlockedSteps[tt.entityID]; !reflect.DeepEqual(got, tt.wantBlockedOn) {
				t.Errorf("BlockedSteps[%q] = %v, want %v", tt.entityID, got, tt.wantBlockedOn)
			}
		})
	}
}

func TestPortConnections(t *testing.T) {
	from := &Entity{ID: "a", Outputs: []DataPort{{Name: "x"}, {Name: "y"}}}
	to := &Entity{ID: "b", Inputs: []DataPort{{Name: "x"}, {Name: "z"}}}
	single := &Entity{ID: "s", Inputs: []DataPort{{Name: "only"}}}
	oneOut := &Entity{ID: "o", Outputs: []DataPort{{Name: "out"}}}

	tests := []struct {
		name string
		flow Flow
		from *Entity
		to   *Entity
		want []portConnection
	}{
		{name: "same name only", flow: Flow{}, from: from, to: to, want: []portConnection{{"x", "x"}}},
		{name: "label naming an output is not a wildcard", flow: Flow{Label: "y"}, from: from, to: to, want: []portConnection{{"x", "x"}}},
		{name: "label naming an output links to the sole input", flow: Flow{Label: "y"}, from: from, to: single, want: []portConnection{{"y", "only"}}},
		{name: "label naming an input links from the sole output", flow: Flow{Label: "z"}, from: oneOut, to: to, want: []portConnection{{"out", "z"}}},
		{name: "label naming an input with several outputs links nothing", flow: Flow{Label: "z"}, from: from, to: to, want: []portConnection{{"x", "x"}}},
		{name: "single pair fallback", flow: Flow{Label: "step one"}, from: oneOut, to: single, want: []portConnection{{"out", "only"}}},
		{name: "explicit mappings win", flow: Flow{Label: "y", DataMappings: []DataPortMapping{{OutputPort: "y", InputPort: "z"}}}, from: from, to: to, want: []portConnection{{"y", "z"}}},
		{name: "no ports", flow: Flow{Label: "x"}, from: &Entity{ID: "e"}, to: &Entity{ID: "f"}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := portConnections(tt.flow, tt.from, tt.to)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("portConnections() = %v, want %v", got, tt.want)
			}
		})
	}
}
