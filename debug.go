package pidl

import (
	"fmt"
	"strings"
)

// Breakpoint represents a debug breakpoint.
type Breakpoint struct {
	// FlowIndex is the flow index where the breakpoint is set.
	FlowIndex int

	// Condition is an optional condition expression (evaluated by ConditionEvaluator).
	Condition string

	// Enabled indicates if the breakpoint is active.
	Enabled bool

	// HitCount tracks how many times this breakpoint was hit.
	HitCount int
}

// DebugState represents the current state of a debug session.
type DebugState struct {
	// FlowIndex is the current position in the protocol.
	FlowIndex int

	// StepsExecuted is the number of steps executed so far.
	StepsExecuted int

	// CurrentFlow is the next flow to be executed (nil if complete).
	CurrentFlow *Flow

	// EntityStates maps entity IDs to their current states.
	EntityStates map[string]string

	// IsCompleted indicates if execution has finished.
	IsCompleted bool

	// AtBreakpoint indicates if execution stopped at a breakpoint.
	AtBreakpoint bool

	// Error holds any execution error.
	Error error
}

// DebugSession provides interactive debugging of protocol execution.
type DebugSession struct {
	// Executor runs the protocol.
	Executor *Executor

	// Context holds the execution state.
	Context *ExecutionContext

	// Breakpoints maps flow indices to breakpoints.
	Breakpoints map[int]*Breakpoint

	// Watchpoints maps entity IDs to watched state patterns.
	Watchpoints map[string]string

	// lastBreakpointIndex tracks the last hit breakpoint to avoid stopping twice.
	lastBreakpointIndex int
}

// NewDebugSession creates a new debug session for a protocol.
func NewDebugSession(p *Protocol) *DebugSession {
	executor := NewExecutor(p)
	return &DebugSession{
		Executor:            executor,
		Context:             executor.NewContext(),
		Breakpoints:         make(map[int]*Breakpoint),
		Watchpoints:         make(map[string]string),
		lastBreakpointIndex: -1,
	}
}

// Step executes the next flow and returns the step details.
// Returns nil when execution is complete.
func (d *DebugSession) Step() (*ExecutionStep, error) {
	if d.Context.Completed {
		return nil, nil
	}

	d.lastBreakpointIndex = -1
	return d.Executor.Step(d.Context)
}

// Continue runs until a breakpoint is hit or execution completes.
// Returns the step where it stopped (breakpoint or final step).
func (d *DebugSession) Continue() (*ExecutionStep, error) {
	var lastStep *ExecutionStep

	for !d.Context.Completed {
		// Check if we're at a breakpoint (but not the one we just hit)
		if d.Context.FlowIndex != d.lastBreakpointIndex {
			if bp, ok := d.Breakpoints[d.Context.FlowIndex]; ok && bp.Enabled {
				// Check condition if present
				if bp.Condition == "" || d.evaluateCondition(bp.Condition) {
					bp.HitCount++
					d.lastBreakpointIndex = d.Context.FlowIndex
					return lastStep, nil
				}
			}
		}

		step, err := d.Executor.Step(d.Context)
		if err != nil {
			return lastStep, err
		}
		lastStep = step
	}

	return lastStep, nil
}

// SetBreakpoint sets a breakpoint at the specified flow index.
func (d *DebugSession) SetBreakpoint(flowIndex int, condition string) error {
	if flowIndex < 0 || flowIndex >= len(d.Executor.Protocol.Flows) {
		return fmt.Errorf("flow index %d out of range (0-%d)", flowIndex, len(d.Executor.Protocol.Flows)-1)
	}

	d.Breakpoints[flowIndex] = &Breakpoint{
		FlowIndex: flowIndex,
		Condition: condition,
		Enabled:   true,
		HitCount:  0,
	}

	return nil
}

// RemoveBreakpoint removes a breakpoint at the specified flow index.
func (d *DebugSession) RemoveBreakpoint(flowIndex int) error {
	if _, ok := d.Breakpoints[flowIndex]; !ok {
		return fmt.Errorf("no breakpoint at flow index %d", flowIndex)
	}

	delete(d.Breakpoints, flowIndex)
	return nil
}

// EnableBreakpoint enables or disables a breakpoint.
func (d *DebugSession) EnableBreakpoint(flowIndex int, enabled bool) error {
	bp, ok := d.Breakpoints[flowIndex]
	if !ok {
		return fmt.Errorf("no breakpoint at flow index %d", flowIndex)
	}

	bp.Enabled = enabled
	return nil
}

// ListBreakpoints returns all breakpoints.
func (d *DebugSession) ListBreakpoints() []*Breakpoint {
	breakpoints := make([]*Breakpoint, 0, len(d.Breakpoints))
	for _, bp := range d.Breakpoints {
		breakpoints = append(breakpoints, bp)
	}
	return breakpoints
}

// Inspect returns the current debug state.
func (d *DebugSession) Inspect() *DebugState {
	state := &DebugState{
		FlowIndex:     d.Context.FlowIndex,
		StepsExecuted: len(d.Context.Trace.Steps),
		EntityStates:  copyStates(d.Context.EntityStates),
		IsCompleted:   d.Context.Completed,
		Error:         d.Context.Error,
	}

	if !d.Context.Completed && d.Context.FlowIndex < len(d.Executor.Protocol.Flows) {
		flow := &d.Executor.Protocol.Flows[d.Context.FlowIndex]
		state.CurrentFlow = flow
	}

	// Check if at a breakpoint
	if _, ok := d.Breakpoints[d.Context.FlowIndex]; ok {
		state.AtBreakpoint = true
	}

	return state
}

// InspectEntity returns the entity and its current state.
func (d *DebugSession) InspectEntity(id string) (*Entity, string, error) {
	entity := d.Executor.Protocol.EntityByID(id)
	if entity == nil {
		return nil, "", fmt.Errorf("entity %q not found", id)
	}

	state := d.Context.EntityStates[id]
	return entity, state, nil
}

// InspectFlow returns a flow by index.
func (d *DebugSession) InspectFlow(index int) (*Flow, error) {
	if index < 0 || index >= len(d.Executor.Protocol.Flows) {
		return nil, fmt.Errorf("flow index %d out of range (0-%d)", index, len(d.Executor.Protocol.Flows)-1)
	}
	return &d.Executor.Protocol.Flows[index], nil
}

// SetEntityState manually sets an entity's state.
func (d *DebugSession) SetEntityState(entityID, stateID string) error {
	entity := d.Executor.Protocol.EntityByID(entityID)
	if entity == nil {
		return fmt.Errorf("entity %q not found", entityID)
	}

	// Validate state exists if entity has states defined
	if entity.HasStates() {
		found := false
		for _, s := range entity.States {
			if s.ID == stateID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("state %q not found for entity %q", stateID, entityID)
		}
	}

	d.Context.SetEntityState(entityID, stateID)
	return nil
}

// Reset restarts the debug session from the beginning.
func (d *DebugSession) Reset() {
	d.Executor.Reset(d.Context)
	d.lastBreakpointIndex = -1
	// Keep breakpoints and watchpoints
}

// Protocol returns the protocol being debugged.
func (d *DebugSession) Protocol() *Protocol {
	return d.Executor.Protocol
}

// Trace returns the current execution trace.
func (d *DebugSession) Trace() *ExecutionTrace {
	return d.Context.Trace
}

// evaluateCondition evaluates a simple condition expression.
// Supports: entity.state == "value", entity.state != "value"
func (d *DebugSession) evaluateCondition(condition string) bool {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return true
	}

	// Parse simple equality: entity.state == "value"
	if strings.Contains(condition, "==") {
		parts := strings.SplitN(condition, "==", 2)
		if len(parts) == 2 {
			lhs := strings.TrimSpace(parts[0])
			rhs := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			return d.evaluateLHS(lhs) == rhs
		}
	}

	// Parse inequality: entity.state != "value"
	if strings.Contains(condition, "!=") {
		parts := strings.SplitN(condition, "!=", 2)
		if len(parts) == 2 {
			lhs := strings.TrimSpace(parts[0])
			rhs := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			return d.evaluateLHS(lhs) != rhs
		}
	}

	// Default: treat as truthy if not empty
	return true
}

// evaluateLHS evaluates the left-hand side of a condition.
// Supports: entity_id.state, entity_id
func (d *DebugSession) evaluateLHS(expr string) string {
	expr = strings.TrimSpace(expr)

	// Check for entity.state pattern
	if strings.HasSuffix(expr, ".state") {
		entityID := strings.TrimSuffix(expr, ".state")
		return d.Context.EntityStates[entityID]
	}

	// Try as entity ID
	return d.Context.EntityStates[expr]
}

// DebugStateString returns a human-readable representation of the debug state.
func (d *DebugState) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Flow Index: %d\n", d.FlowIndex))
	sb.WriteString(fmt.Sprintf("Steps Executed: %d\n", d.StepsExecuted))
	sb.WriteString(fmt.Sprintf("Completed: %v\n", d.IsCompleted))
	sb.WriteString(fmt.Sprintf("At Breakpoint: %v\n", d.AtBreakpoint))

	if d.CurrentFlow != nil {
		sb.WriteString(fmt.Sprintf("\nNext Flow:\n"))
		sb.WriteString(fmt.Sprintf("  %s -> %s: %s\n", d.CurrentFlow.From, d.CurrentFlow.To, d.CurrentFlow.DisplayLabel()))
		if d.CurrentFlow.Condition != "" {
			sb.WriteString(fmt.Sprintf("  Condition: %s\n", d.CurrentFlow.Condition))
		}
	}

	if len(d.EntityStates) > 0 {
		sb.WriteString(fmt.Sprintf("\nEntity States:\n"))
		for entity, state := range d.EntityStates {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", entity, state))
		}
	}

	if d.Error != nil {
		sb.WriteString(fmt.Sprintf("\nError: %v\n", d.Error))
	}

	return sb.String()
}

// FlowListItem represents a flow for listing purposes.
type FlowListItem struct {
	Index         int
	Flow          *Flow
	IsCurrent     bool
	HasBreakpoint bool
	IsExecuted    bool
}

// ListFlows returns all flows with their status.
func (d *DebugSession) ListFlows() []FlowListItem {
	items := make([]FlowListItem, len(d.Executor.Protocol.Flows))

	for i := range d.Executor.Protocol.Flows {
		flow := &d.Executor.Protocol.Flows[i]
		_, hasBP := d.Breakpoints[i]

		items[i] = FlowListItem{
			Index:         i,
			Flow:          flow,
			IsCurrent:     i == d.Context.FlowIndex,
			HasBreakpoint: hasBP,
			IsExecuted:    i < d.Context.FlowIndex,
		}
	}

	return items
}

// FormatFlowList returns a formatted string of all flows.
func (d *DebugSession) FormatFlowList() string {
	var sb strings.Builder

	flows := d.ListFlows()
	for _, item := range flows {
		// Position marker
		marker := "   "
		if item.IsCurrent {
			marker = "=> "
		}

		// Breakpoint indicator
		bp := " "
		if item.HasBreakpoint {
			bp = "*"
		}

		// Execution status
		status := " "
		if item.IsExecuted {
			status = "+"
		}

		// Flow info
		label := item.Flow.Action
		if item.Flow.Label != "" {
			label = item.Flow.Label
		}

		sb.WriteString(fmt.Sprintf("%s%s%s %3d: %s -> %s: %s\n",
			marker, bp, status, item.Index, item.Flow.From, item.Flow.To, label))
	}

	return sb.String()
}
