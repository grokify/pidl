package pidl

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExecutionContext holds the runtime state of a protocol execution.
type ExecutionContext struct {
	// Protocol being executed.
	Protocol *Protocol

	// EntityStates maps entity ID to current state ID.
	EntityStates map[string]string

	// FlowIndex is the current position in the protocol's flows.
	FlowIndex int

	// Trace records execution history.
	Trace *ExecutionTrace

	// EventQueue holds pending events for event-driven execution.
	EventQueue []ExecutionEvent

	// Completed indicates execution has finished.
	Completed bool

	// Error holds any execution error.
	Error error
}

// ExecutionTrace records the complete execution history.
type ExecutionTrace struct {
	// ProtocolID identifies the executed protocol.
	ProtocolID string `json:"protocol_id"`

	// ProtocolName is the human-readable protocol name.
	ProtocolName string `json:"protocol_name"`

	// StartTime when execution began.
	StartTime time.Time `json:"start_time"`

	// EndTime when execution completed.
	EndTime time.Time `json:"end_time,omitempty"`

	// Steps records each executed flow.
	Steps []ExecutionStep `json:"steps"`

	// InitialStates captures entity states at start.
	InitialStates map[string]string `json:"initial_states,omitempty"`

	// FinalStates captures entity states at end.
	FinalStates map[string]string `json:"final_states,omitempty"`

	// Completed indicates if execution ran to completion.
	Completed bool `json:"completed"`

	// Error message if execution failed.
	Error string `json:"error,omitempty"`
}

// ExecutionStep records a single flow execution.
type ExecutionStep struct {
	// StepNumber is the 1-based step counter.
	StepNumber int `json:"step_number"`

	// FlowIndex is the index in Protocol.Flows.
	FlowIndex int `json:"flow_index"`

	// Timestamp when this step executed.
	Timestamp time.Time `json:"timestamp"`

	// From is the source entity ID.
	From string `json:"from"`

	// To is the target entity ID.
	To string `json:"to"`

	// Action is the flow action.
	Action string `json:"action"`

	// Label is the flow display label.
	Label string `json:"label,omitempty"`

	// Mode is the flow mode.
	Mode FlowMode `json:"mode,omitempty"`

	// Phase is the phase ID.
	Phase string `json:"phase,omitempty"`

	// Condition is the flow's condition (if any).
	Condition string `json:"condition,omitempty"`

	// ConditionMet indicates if condition was satisfied (for conditional flows).
	ConditionMet *bool `json:"condition_met,omitempty"`

	// StateChanges records state mutations from this step.
	StateChanges []StateChange `json:"state_changes,omitempty"`

	// Skipped indicates this flow was skipped (condition not met).
	Skipped bool `json:"skipped,omitempty"`

	// SkipReason explains why flow was skipped.
	SkipReason string `json:"skip_reason,omitempty"`
}

// StateChange records a single state mutation.
type StateChange struct {
	// Entity is the entity ID.
	Entity string `json:"entity"`

	// FromState is the prior state (empty if not tracked).
	FromState string `json:"from_state,omitempty"`

	// ToState is the new state.
	ToState string `json:"to_state"`
}

// ExecutionEvent represents a pending event in the queue.
type ExecutionEvent struct {
	// Type categorizes the event.
	Type string `json:"type"`

	// FlowIndex references a specific flow (if applicable).
	FlowIndex int `json:"flow_index,omitempty"`

	// EntityID references a specific entity (if applicable).
	EntityID string `json:"entity_id,omitempty"`

	// Data holds event-specific payload.
	Data map[string]interface{} `json:"data,omitempty"`
}

// Executor runs protocol simulations.
type Executor struct {
	// Protocol to execute.
	Protocol *Protocol

	// ConditionEvaluator is called to evaluate flow conditions.
	// If nil, all conditions are assumed to be met.
	ConditionEvaluator func(ctx *ExecutionContext, flow *Flow) bool
}

// NewExecutor creates an Executor for the given protocol.
func NewExecutor(p *Protocol) *Executor {
	return &Executor{
		Protocol: p,
	}
}

// NewContext creates a fresh execution context with initial states.
func (e *Executor) NewContext() *ExecutionContext {
	ctx := &ExecutionContext{
		Protocol:     e.Protocol,
		EntityStates: make(map[string]string),
		FlowIndex:    0,
		Trace: &ExecutionTrace{
			ProtocolID:    e.Protocol.ProtocolMeta.ID,
			ProtocolName:  e.Protocol.ProtocolMeta.Name,
			StartTime:     time.Now(),
			Steps:         make([]ExecutionStep, 0),
			InitialStates: make(map[string]string),
		},
		EventQueue: make([]ExecutionEvent, 0),
	}

	// Initialize entity states to their initial states
	for _, entity := range e.Protocol.Entities {
		if entity.HasStates() {
			for _, state := range entity.States {
				if state.Initial {
					ctx.EntityStates[entity.ID] = state.ID
					ctx.Trace.InitialStates[entity.ID] = state.ID
					break
				}
			}
		}
	}

	return ctx
}

// Step executes the next flow and returns the step details.
// Returns nil when execution is complete.
func (e *Executor) Step(ctx *ExecutionContext) (*ExecutionStep, error) {
	if ctx.Completed {
		return nil, nil
	}

	if ctx.FlowIndex >= len(e.Protocol.Flows) {
		ctx.Completed = true
		ctx.Trace.EndTime = time.Now()
		ctx.Trace.Completed = true
		ctx.Trace.FinalStates = copyStates(ctx.EntityStates)
		return nil, nil
	}

	flow := &e.Protocol.Flows[ctx.FlowIndex]
	step := &ExecutionStep{
		StepNumber: len(ctx.Trace.Steps) + 1,
		FlowIndex:  ctx.FlowIndex,
		Timestamp:  time.Now(),
		From:       flow.From,
		To:         flow.To,
		Action:     flow.Action,
		Label:      flow.Label,
		Mode:       flow.Mode,
		Phase:      flow.Phase,
		Condition:  flow.Condition,
	}

	// Evaluate condition if present
	if flow.Condition != "" {
		conditionMet := true
		if e.ConditionEvaluator != nil {
			conditionMet = e.ConditionEvaluator(ctx, flow)
		}
		step.ConditionMet = &conditionMet

		if !conditionMet {
			step.Skipped = true
			step.SkipReason = fmt.Sprintf("condition %q not met", flow.Condition)
			ctx.Trace.Steps = append(ctx.Trace.Steps, *step)
			ctx.FlowIndex++
			return step, nil
		}
	}

	// Apply state mutations
	for _, mutation := range flow.Sets {
		change := StateChange{
			Entity:    mutation.Entity,
			FromState: ctx.EntityStates[mutation.Entity],
			ToState:   mutation.To,
		}

		// Validate from state if specified
		if mutation.From != "" && ctx.EntityStates[mutation.Entity] != mutation.From {
			err := fmt.Errorf("state mismatch for %s: expected %q, got %q",
				mutation.Entity, mutation.From, ctx.EntityStates[mutation.Entity])
			ctx.Error = err
			ctx.Trace.Error = err.Error()
			return nil, err
		}

		ctx.EntityStates[mutation.Entity] = mutation.To
		step.StateChanges = append(step.StateChanges, change)
	}

	ctx.Trace.Steps = append(ctx.Trace.Steps, *step)
	ctx.FlowIndex++

	return step, nil
}

// Run executes all flows to completion.
func (e *Executor) Run(ctx *ExecutionContext) (*ExecutionTrace, error) {
	for !ctx.Completed {
		_, err := e.Step(ctx)
		if err != nil {
			return ctx.Trace, err
		}
	}
	return ctx.Trace, nil
}

// RunN executes up to n steps.
func (e *Executor) RunN(ctx *ExecutionContext, n int) (*ExecutionTrace, error) {
	for i := 0; i < n && !ctx.Completed; i++ {
		_, err := e.Step(ctx)
		if err != nil {
			return ctx.Trace, err
		}
	}
	return ctx.Trace, nil
}

// Reset restarts execution from the beginning.
func (e *Executor) Reset(ctx *ExecutionContext) {
	ctx.FlowIndex = 0
	ctx.Completed = false
	ctx.Error = nil
	ctx.EntityStates = make(map[string]string)
	ctx.Trace = &ExecutionTrace{
		ProtocolID:    e.Protocol.ProtocolMeta.ID,
		ProtocolName:  e.Protocol.ProtocolMeta.Name,
		StartTime:     time.Now(),
		Steps:         make([]ExecutionStep, 0),
		InitialStates: make(map[string]string),
	}
	ctx.EventQueue = make([]ExecutionEvent, 0)

	// Re-initialize entity states
	for _, entity := range e.Protocol.Entities {
		if entity.HasStates() {
			for _, state := range entity.States {
				if state.Initial {
					ctx.EntityStates[entity.ID] = state.ID
					ctx.Trace.InitialStates[entity.ID] = state.ID
					break
				}
			}
		}
	}
}

// GetEntityState returns the current state of an entity.
func (ctx *ExecutionContext) GetEntityState(entityID string) string {
	return ctx.EntityStates[entityID]
}

// SetEntityState manually sets an entity's state.
func (ctx *ExecutionContext) SetEntityState(entityID, stateID string) {
	ctx.EntityStates[entityID] = stateID
}

// CurrentFlow returns the next flow to be executed, or nil if complete.
func (ctx *ExecutionContext) CurrentFlow() *Flow {
	if ctx.Completed || ctx.FlowIndex >= len(ctx.Protocol.Flows) {
		return nil
	}
	return &ctx.Protocol.Flows[ctx.FlowIndex]
}

// Progress returns execution progress as a percentage (0-100).
func (ctx *ExecutionContext) Progress() float64 {
	if len(ctx.Protocol.Flows) == 0 {
		return 100
	}
	return float64(ctx.FlowIndex) / float64(len(ctx.Protocol.Flows)) * 100
}

// Duration returns the execution duration so far.
func (t *ExecutionTrace) Duration() time.Duration {
	if t.EndTime.IsZero() {
		return time.Since(t.StartTime)
	}
	return t.EndTime.Sub(t.StartTime)
}

// StepCount returns the number of executed steps.
func (t *ExecutionTrace) StepCount() int {
	return len(t.Steps)
}

// SkippedCount returns the number of skipped steps.
func (t *ExecutionTrace) SkippedCount() int {
	count := 0
	for _, step := range t.Steps {
		if step.Skipped {
			count++
		}
	}
	return count
}

// StateChangeCount returns the total number of state changes.
func (t *ExecutionTrace) StateChangeCount() int {
	count := 0
	for _, step := range t.Steps {
		count += len(step.StateChanges)
	}
	return count
}

// ToJSON returns the trace as JSON bytes.
func (t *ExecutionTrace) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

// copyStates creates a copy of the state map.
func copyStates(states map[string]string) map[string]string {
	if states == nil {
		return nil
	}
	result := make(map[string]string, len(states))
	for k, v := range states {
		result[k] = v
	}
	return result
}
