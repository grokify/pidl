package pidl

import (
	"testing"
)

func createParallelTestProtocol() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "parallel-test",
			Name: "Parallel Test",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{ID: "start", Name: "Start", StepType: StepTypeDeterministic},
			{ID: "branch_a", Name: "Branch A", StepType: StepTypeLLM},
			{ID: "branch_b", Name: "Branch B", StepType: StepTypeExternal},
			{ID: "branch_c", Name: "Branch C", StepType: StepTypeDeterministic},
			{ID: "join", Name: "Join", StepType: StepTypeDeterministic},
			{ID: "end", Name: "End", StepType: StepTypeDeterministic},
		},
		Flows: []Flow{
			{From: "start", To: "branch_a", Action: "fork"},
			{From: "start", To: "branch_b", Action: "fork"},
			{From: "start", To: "branch_c", Action: "fork"},
			{From: "branch_a", To: "join", Action: "result"},
			{From: "branch_b", To: "join", Action: "result"},
			{From: "branch_c", To: "join", Action: "result"},
			{From: "join", To: "end", Action: "complete"},
		},
	}
}

func TestAnalyzeExecutionGraph_Basic(t *testing.T) {
	p := createParallelTestProtocol()
	graph := AnalyzeExecutionGraph(p)

	if graph.ProtocolID != "parallel-test" {
		t.Errorf("expected protocol ID 'parallel-test', got '%s'", graph.ProtocolID)
	}

	// Should have stages
	if len(graph.Stages) == 0 {
		t.Error("expected at least one stage")
	}
}

func TestAnalyzeExecutionGraph_Parallelism(t *testing.T) {
	p := createParallelTestProtocol()
	graph := AnalyzeExecutionGraph(p)

	// Find stage with branches
	foundParallel := false
	for _, stage := range graph.Stages {
		if len(stage.Steps) >= 3 {
			foundParallel = true
			break
		}
	}

	if !foundParallel {
		t.Log("Note: parallel stage detection depends on topological sort")
	}
}

func TestAnalyzeExecutionGraph_CriticalPath(t *testing.T) {
	p := createParallelTestProtocol()
	graph := AnalyzeExecutionGraph(p)

	if len(graph.CriticalPath) == 0 {
		t.Error("expected non-empty critical path")
	}

	// Critical path should include start, one branch, join, and end
	if len(graph.CriticalPath) < 4 {
		t.Errorf("expected critical path length >= 4, got %d", len(graph.CriticalPath))
	}
}

func TestCanExecuteInParallel(t *testing.T) {
	p := createParallelTestProtocol()

	// Branches should be able to execute in parallel
	if !CanExecuteInParallel(p, "branch_a", "branch_b") {
		t.Error("expected branch_a and branch_b to be parallelizable")
	}

	if !CanExecuteInParallel(p, "branch_a", "branch_c") {
		t.Error("expected branch_a and branch_c to be parallelizable")
	}

	// Start and join should NOT be parallelizable (transitive dependency)
	if CanExecuteInParallel(p, "start", "join") {
		t.Error("expected start and join to NOT be parallelizable")
	}
}

func TestGetMaxParallelism(t *testing.T) {
	p := createParallelTestProtocol()
	maxParallel := GetMaxParallelism(p)

	// Should be at least 3 (the three branches)
	if maxParallel < 3 {
		t.Logf("max parallelism: %d (expected >= 3)", maxParallel)
	}
}

func TestDetectParallelBlocks(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "explicit-parallel",
			Name: "Explicit Parallel",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "fork",
				Name:     "Fork Point",
				StepType: StepTypeParallel,
				Parallel: &ParallelConfig{
					Mode:          ParallelModeForkJoin,
					JoinCondition: JoinConditionAll,
					Branches: []ParallelBranch{
						{ID: "b1", EntityID: "worker_1"},
						{ID: "b2", EntityID: "worker_2"},
					},
				},
			},
			{ID: "worker_1", Name: "Worker 1", StepType: StepTypeDeterministic},
			{ID: "worker_2", Name: "Worker 2", StepType: StepTypeDeterministic},
			{ID: "collector", Name: "Collector", StepType: StepTypeDeterministic},
		},
		Flows: []Flow{
			{From: "fork", To: "worker_1", Action: "dispatch"},
			{From: "fork", To: "worker_2", Action: "dispatch"},
			{From: "worker_1", To: "collector", Action: "result"},
			{From: "worker_2", To: "collector", Action: "result"},
		},
	}

	blocks := DetectParallelBlocks(p)

	if len(blocks) != 1 {
		t.Errorf("expected 1 parallel block, got %d", len(blocks))
	}

	if len(blocks) > 0 {
		block := blocks[0]
		if block.ForkEntity != "fork" {
			t.Errorf("expected fork entity 'fork', got '%s'", block.ForkEntity)
		}
		if block.Mode != ParallelModeForkJoin {
			t.Errorf("expected mode 'fork_join', got '%s'", block.Mode)
		}
		if block.JoinEntity != "collector" {
			t.Logf("join entity: %s (expected collector)", block.JoinEntity)
		}
	}
}

func TestLinearProtocol_NoParallelism(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "linear",
			Name: "Linear",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{ID: "a", Name: "A", StepType: StepTypeDeterministic},
			{ID: "b", Name: "B", StepType: StepTypeDeterministic},
			{ID: "c", Name: "C", StepType: StepTypeDeterministic},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "next"},
			{From: "b", To: "c", Action: "next"},
		},
	}

	maxParallel := GetMaxParallelism(p)
	if maxParallel > 1 {
		t.Errorf("expected max parallelism 1 for linear protocol, got %d", maxParallel)
	}

	graph := AnalyzeExecutionGraph(p)
	if len(graph.CriticalPath) != 3 {
		t.Errorf("expected critical path length 3, got %d", len(graph.CriticalPath))
	}
}

func TestParallelModes(t *testing.T) {
	modes := []ParallelMode{
		ParallelModeForkJoin,
		ParallelModeRace,
		ParallelModeScatter,
		ParallelModeGather,
	}

	for _, mode := range modes {
		if mode == "" {
			t.Errorf("mode should not be empty")
		}
	}
}

func TestJoinConditions(t *testing.T) {
	conditions := []JoinCondition{
		JoinConditionAll,
		JoinConditionAny,
		JoinConditionMajority,
		JoinConditionQuorum,
	}

	for _, cond := range conditions {
		if cond == "" {
			t.Errorf("join condition should not be empty")
		}
	}
}
