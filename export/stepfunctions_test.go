package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

// parallelProcessProtocol builds a process spec with a fork entity that
// declares the given parallel branches and join condition.
func parallelProcessProtocol(join pidl.JoinCondition, branches ...pidl.ParallelBranch) *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "parallel-etl",
			Name: "Parallel ETL",
			Kind: pidl.ProtocolKindProcess,
		},
		Entities: []pidl.Entity{
			{
				ID:       "fork",
				Name:     "Fork",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeDeterministic,
				Parallel: &pidl.ParallelConfig{
					Mode:          pidl.ParallelModeForkJoin,
					Branches:      branches,
					JoinCondition: join,
				},
			},
			{ID: "classify", Name: "Classify", Type: pidl.EntityTypeServer, StepType: pidl.StepTypeLLM, Description: "LLM branch"},
			{ID: "fetch", Name: "Fetch", Type: pidl.EntityTypeServer, StepType: pidl.StepTypeExternal, Description: "HTTP branch"},
			{ID: "join", Name: "Join", Type: pidl.EntityTypeServer, StepType: pidl.StepTypeDeterministic},
		},
		Flows: []pidl.Flow{
			{From: "fork", To: "classify", Action: "send"},
			{From: "fork", To: "fetch", Action: "send"},
			{From: "classify", To: "join", Action: "send"},
			{From: "fetch", To: "join", Action: "send"},
		},
	}
}

func TestStepFunctionsExporter_ParallelBlock(t *testing.T) {
	twoBranches := []pidl.ParallelBranch{
		{ID: "b1", EntityID: "classify"},
		{ID: "b2", EntityID: "fetch"},
	}

	tests := []struct {
		name            string
		protocol        *pidl.Protocol
		wantState       bool
		wantBranches    int
		wantResources   []string
		wantRaceComment bool
	}{
		{
			name:          "join all emits parallel state",
			protocol:      parallelProcessProtocol(pidl.JoinConditionAll, twoBranches...),
			wantState:     true,
			wantBranches:  2,
			wantResources: []string{"arn:aws:states:::bedrock:invokeModel", "arn:aws:states:::http:invoke"},
		},
		{
			name:            "join any annotates race",
			protocol:        parallelProcessProtocol(pidl.JoinConditionAny, twoBranches...),
			wantState:       true,
			wantBranches:    2,
			wantResources:   []string{"arn:aws:states:::bedrock:invokeModel", "arn:aws:states:::http:invoke"},
			wantRaceComment: true,
		},
		{
			name:          "unset join condition behaves like all",
			protocol:      parallelProcessProtocol("", twoBranches...),
			wantState:     true,
			wantBranches:  2,
			wantResources: []string{"arn:aws:states:::bedrock:invokeModel", "arn:aws:states:::http:invoke"},
		},
		{
			name: "branch referencing unknown entity is dropped",
			protocol: parallelProcessProtocol(pidl.JoinConditionAll,
				pidl.ParallelBranch{ID: "b1", EntityID: "classify"},
				pidl.ParallelBranch{ID: "b2", EntityID: "does-not-exist"},
			),
			wantState:     true,
			wantBranches:  1,
			wantResources: []string{"arn:aws:states:::bedrock:invokeModel"},
		},
		{
			name:      "single branch is not a parallel block",
			protocol:  parallelProcessProtocol(pidl.JoinConditionAll, pidl.ParallelBranch{ID: "b1", EntityID: "classify"}),
			wantState: false,
		},
		{
			name:      "no branches is not a parallel block",
			protocol:  parallelProcessProtocol(pidl.JoinConditionAll),
			wantState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := NewStepFunctionsExporter().Export(tt.protocol)
			if err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			var sm StateMachine
			if err := json.Unmarshal([]byte(out), &sm); err != nil {
				t.Fatalf("output is not a valid state machine: %v\n%s", err, out)
			}

			state, ok := sm.States["fork-parallel"]
			if ok != tt.wantState {
				t.Fatalf("parallel state present = %v, want %v; states: %v", ok, tt.wantState, stateNames(sm))
			}
			if !tt.wantState {
				return
			}

			if state.Type != "Parallel" {
				t.Errorf("Type = %q, want Parallel", state.Type)
			}
			if len(state.Branches) != tt.wantBranches {
				t.Fatalf("got %d branches, want %d", len(state.Branches), tt.wantBranches)
			}
			for i, branch := range state.Branches {
				task, ok := branch.States[branch.StartAt]
				if !ok {
					t.Fatalf("branch %d StartAt %q not in its States", i, branch.StartAt)
				}
				if task.Type != "Task" || !task.End {
					t.Errorf("branch %d state = %+v, want terminal Task", i, task)
				}
				if task.Resource != tt.wantResources[i] {
					t.Errorf("branch %d Resource = %q, want %q", i, task.Resource, tt.wantResources[i])
				}
			}

			hasRace := strings.Contains(state.Comment, "race")
			if hasRace != tt.wantRaceComment {
				t.Errorf("Comment = %q, race annotation = %v, want %v", state.Comment, hasRace, tt.wantRaceComment)
			}
		})
	}
}

func stateNames(sm StateMachine) []string {
	names := make([]string, 0, len(sm.States))
	for name := range sm.States {
		names = append(names, name)
	}
	return names
}
