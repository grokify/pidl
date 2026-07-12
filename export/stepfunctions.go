package export

import (
	"encoding/json"
	"fmt"

	"github.com/grokify/pidl"
)

// StepFunctionsExporter exports PIDL process specs to AWS Step Functions (Amazon States Language).
type StepFunctionsExporter struct {
	// StateMachineName overrides the generated state machine name.
	StateMachineName string
	// Comment adds a comment to the state machine definition.
	Comment string
	// TimeoutSeconds sets the default timeout for the state machine.
	TimeoutSeconds int
}

// NewStepFunctionsExporter creates a new Step Functions exporter with defaults.
func NewStepFunctionsExporter() *StepFunctionsExporter {
	return &StepFunctionsExporter{
		TimeoutSeconds: 3600,
	}
}

// StateMachine represents an AWS Step Functions state machine definition.
type StateMachine struct {
	Comment        string           `json:"Comment,omitempty"`
	StartAt        string           `json:"StartAt"`
	States         map[string]State `json:"States"`
	TimeoutSeconds int              `json:"TimeoutSeconds,omitempty"`
}

// State represents a state in the state machine.
type State struct {
	Type       string  `json:"Type"`
	Comment    string  `json:"Comment,omitempty"`
	Resource   string  `json:"Resource,omitempty"`
	Next       string  `json:"Next,omitempty"`
	End        bool    `json:"End,omitempty"`
	Retry      []Retry `json:"Retry,omitempty"`
	Catch      []Catch `json:"Catch,omitempty"`
	ResultPath string  `json:"ResultPath,omitempty"`

	// Task-specific fields
	TimeoutSeconds   int `json:"TimeoutSeconds,omitempty"`
	HeartbeatSeconds int `json:"HeartbeatSeconds,omitempty"`

	// Choice state fields
	Choices []Choice `json:"Choices,omitempty"`
	Default string   `json:"Default,omitempty"`

	// Parallel state fields
	Branches []Branch `json:"Branches,omitempty"`

	// Wait state fields
	Seconds     int    `json:"Seconds,omitempty"`
	SecondsPath string `json:"SecondsPath,omitempty"`

	// Pass state fields
	Result interface{} `json:"Result,omitempty"`

	// Map state fields
	ItemsPath      string        `json:"ItemsPath,omitempty"`
	MaxConcurrency int           `json:"MaxConcurrency,omitempty"`
	Iterator       *StateMachine `json:"Iterator,omitempty"`
}

// Retry defines retry behavior for a state.
type Retry struct {
	ErrorEquals     []string `json:"ErrorEquals"`
	IntervalSeconds int      `json:"IntervalSeconds,omitempty"`
	MaxAttempts     int      `json:"MaxAttempts,omitempty"`
	BackoffRate     float64  `json:"BackoffRate,omitempty"`
}

// Catch defines error handling for a state.
type Catch struct {
	ErrorEquals []string `json:"ErrorEquals"`
	Next        string   `json:"Next"`
	ResultPath  string   `json:"ResultPath,omitempty"`
}

// Choice defines a choice rule for Choice states.
type Choice struct {
	Variable      string `json:"Variable,omitempty"`
	StringEquals  string `json:"StringEquals,omitempty"`
	NumericEquals int    `json:"NumericEquals,omitempty"`
	BooleanEquals *bool  `json:"BooleanEquals,omitempty"`
	Next          string `json:"Next"`
}

// Branch defines a parallel branch.
type Branch struct {
	StartAt string           `json:"StartAt"`
	States  map[string]State `json:"States"`
}

// Export converts a PIDL protocol to AWS Step Functions state machine JSON.
func (e *StepFunctionsExporter) Export(p *pidl.Protocol) (string, error) {
	if p.ProtocolMeta.Kind != pidl.ProtocolKindProcess {
		return "", fmt.Errorf("step functions export only supports process specifications")
	}

	sm := e.buildStateMachine(p)

	data, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal state machine: %w", err)
	}

	return string(data), nil
}

func (e *StepFunctionsExporter) buildStateMachine(p *pidl.Protocol) *StateMachine {
	comment := e.Comment
	if comment == "" {
		comment = fmt.Sprintf("Generated from PIDL: %s", p.ProtocolMeta.Name)
	}

	sm := &StateMachine{
		Comment:        comment,
		States:         make(map[string]State),
		TimeoutSeconds: e.TimeoutSeconds,
	}

	// Filter entities that are process steps
	var steps []pidl.Entity
	for _, entity := range p.Entities {
		if entity.StepType != "" {
			steps = append(steps, entity)
		}
	}

	if len(steps) == 0 {
		return sm
	}

	// Set start state
	sm.StartAt = toStateName(steps[0].ID)

	// Build execution order from flows
	flowGraph := buildFlowGraph(p)

	// Create states for each step
	for i, step := range steps {
		stateName := toStateName(step.ID)
		state := e.buildState(step, flowGraph, i == len(steps)-1)
		sm.States[stateName] = state
	}

	// Handle parallel blocks
	parallelBlocks := pidl.DetectParallelBlocks(p)
	for i := range parallelBlocks {
		e.handleParallelBlock(sm, &parallelBlocks[i], steps)
	}

	return sm
}

func (e *StepFunctionsExporter) buildState(step pidl.Entity, flowGraph map[string][]string, isLast bool) State {
	state := State{
		Type:    "Task",
		Comment: step.Description,
	}

	// Set resource ARN based on step type
	state.Resource = e.stepTypeToResource(step.StepType)

	// Find next state from flow graph
	if !isLast {
		if nexts, ok := flowGraph[step.ID]; ok && len(nexts) > 0 {
			state.Next = toStateName(nexts[0])
		}
	}

	// Set End if this is the last state
	if isLast || len(flowGraph[step.ID]) == 0 {
		state.End = true
		state.Next = ""
	}

	// Add retry policy if configured
	if step.RetryStrategy != nil {
		state.Retry = []Retry{
			{
				ErrorEquals:     []string{"States.ALL"},
				IntervalSeconds: 1,
				MaxAttempts:     step.RetryStrategy.MaxAttempts,
				BackoffRate:     2.0,
			},
		}
	}

	// Add timeout if configured
	if step.Processing != nil && step.Processing.Timeout != "" {
		// Parse timeout and set (simplified - assumes seconds)
		state.TimeoutSeconds = 300 // Default 5 minutes
	}

	// Special handling for human steps - use Activity or callback
	if step.StepType == pidl.StepTypeHuman {
		state.Resource = "arn:aws:states:::activity:HumanApproval"
		state.HeartbeatSeconds = 3600
	}

	return state
}

func (e *StepFunctionsExporter) stepTypeToResource(stepType pidl.StepType) string {
	switch stepType {
	case pidl.StepTypeDeterministic:
		// Lambda function for deterministic processing
		return "arn:aws:states:::lambda:invoke"
	case pidl.StepTypeLLM:
		// Bedrock for LLM processing
		return "arn:aws:states:::bedrock:invokeModel"
	case pidl.StepTypeHuman:
		// Activity for human-in-the-loop
		return "arn:aws:states:::activity:HumanReview"
	case pidl.StepTypeExternal:
		// HTTP endpoint for external calls
		return "arn:aws:states:::http:invoke"
	case pidl.StepTypeTool:
		// Lambda for tool invocations
		return "arn:aws:states:::lambda:invoke"
	default:
		return "arn:aws:states:::lambda:invoke"
	}
}

func (e *StepFunctionsExporter) handleParallelBlock(sm *StateMachine, block *pidl.ParallelBlock, steps []pidl.Entity) {
	if block == nil || len(block.Branches) < 2 {
		return
	}

	// Create a parallel state
	parallelStateName := toStateName(block.ID)
	parallelState := State{
		Type:     "Parallel",
		Comment:  fmt.Sprintf("Parallel execution: %s", block.ID),
		Branches: make([]Branch, 0, len(block.Branches)),
	}

	for _, branch := range block.Branches {
		branchStates := make(map[string]State)
		branchStateName := toStateName(branch.EntityID)

		// Find the entity for this branch
		for _, step := range steps {
			if step.ID == branch.EntityID {
				branchState := State{
					Type:     "Task",
					Resource: e.stepTypeToResource(step.StepType),
					Comment:  step.Description,
					End:      true,
				}
				branchStates[branchStateName] = branchState
				break
			}
		}

		if len(branchStates) > 0 {
			parallelState.Branches = append(parallelState.Branches, Branch{
				StartAt: branchStateName,
				States:  branchStates,
			})
		}
	}

	// Handle join condition
	switch block.JoinCondition {
	case pidl.JoinConditionAll:
		// Default parallel behavior - wait for all
	case pidl.JoinConditionAny:
		// Would need to use a different pattern (not directly supported)
		parallelState.Comment += " (race - first completes wins)"
	}

	sm.States[parallelStateName] = parallelState
}

func buildFlowGraph(p *pidl.Protocol) map[string][]string {
	graph := make(map[string][]string)
	for _, flow := range p.Flows {
		graph[flow.From] = append(graph[flow.From], flow.To)
	}
	return graph
}

func toStateName(id string) string {
	// Convert to valid Step Functions state name
	// State names must be 1-80 characters, and can contain letters, numbers,
	// hyphens, underscores, and periods
	result := ""
	for _, c := range id {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			result += string(c)
		} else if c == ' ' {
			result += "_"
		}
	}
	if len(result) > 80 {
		result = result[:80]
	}
	return result
}
