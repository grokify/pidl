package pidl

import (
	"fmt"
	"sort"
)

// ProcessExecutionContext extends ExecutionContext with process-specific tracking.
type ProcessExecutionContext struct {
	*ExecutionContext

	// AvailableInputs tracks which data ports have data available.
	// Key is "entityID.portName", value is true if available.
	AvailableInputs map[string]bool

	// ProducedOutputs tracks which outputs have been produced.
	// Key is "entityID.portName", value is true if produced.
	ProducedOutputs map[string]bool

	// BlockedSteps tracks which steps are blocked and why.
	BlockedSteps map[string][]string // entityID -> list of missing inputs

	// ExecutionOrder is the topologically sorted execution order.
	ExecutionOrder []string

	// StepStatus tracks the status of each step.
	StepStatus map[string]StepExecutionStatus
}

// StepExecutionStatus represents the execution status of a process step.
type StepExecutionStatus string

const (
	StepStatusPending    StepExecutionStatus = "pending"
	StepStatusReady      StepExecutionStatus = "ready"
	StepStatusBlocked    StepExecutionStatus = "blocked"
	StepStatusInProgress StepExecutionStatus = "in_progress"
	StepStatusCompleted  StepExecutionStatus = "completed"
	StepStatusFailed     StepExecutionStatus = "failed"
	StepStatusSkipped    StepExecutionStatus = "skipped"
)

// ProcessExecutor provides process-aware execution with data flow tracking.
type ProcessExecutor struct {
	*Executor

	// DataFlowGraph maps entity outputs to consuming entity inputs.
	// Key is "sourceEntityID.outputName", value is list of "targetEntityID.inputName".
	DataFlowGraph map[string][]string

	// Dependencies maps entity ID to its required predecessor entity IDs.
	Dependencies map[string][]string

	// Dependents maps entity ID to entities that depend on it.
	Dependents map[string][]string
}

// NewProcessExecutor creates a ProcessExecutor for process spec simulation.
func NewProcessExecutor(p *Protocol) (*ProcessExecutor, error) {
	if !p.IsProcessSpec() {
		return nil, fmt.Errorf("protocol is not a process spec")
	}

	pe := &ProcessExecutor{
		Executor:      NewExecutor(p),
		DataFlowGraph: make(map[string][]string),
		Dependencies:  make(map[string][]string),
		Dependents:    make(map[string][]string),
	}

	// Build dependency graph from flows
	pe.buildDependencyGraph()

	return pe, nil
}

// buildDependencyGraph analyzes flows to build the dependency graph.
func (pe *ProcessExecutor) buildDependencyGraph() {
	for _, flow := range pe.Protocol.Flows {
		// Add direct dependency: 'to' depends on 'from'
		pe.Dependencies[flow.To] = appendUnique(pe.Dependencies[flow.To], flow.From)
		pe.Dependents[flow.From] = appendUnique(pe.Dependents[flow.From], flow.To)

		// Build data flow graph based on port connections
		fromEntity := pe.Protocol.EntityByID(flow.From)
		toEntity := pe.Protocol.EntityByID(flow.To)

		if fromEntity != nil && toEntity != nil {
			// Match outputs to inputs based on flow label or port names
			for _, output := range fromEntity.Outputs {
				outputKey := flow.From + "." + output.Name

				// Check if this output matches any input on the target
				for _, input := range toEntity.Inputs {
					// Match by name, flow label, or if only one output/input pair exists
					matched := false
					if output.Name == input.Name {
						matched = true
					} else if flow.Label != "" && (flow.Label == output.Name || flow.Label == input.Name) {
						matched = true
					} else if len(fromEntity.Outputs) == 1 && len(toEntity.Inputs) == 1 {
						// If there's only one output and one input, they connect
						matched = true
					}

					if matched {
						inputKey := flow.To + "." + input.Name
						pe.DataFlowGraph[outputKey] = appendUnique(pe.DataFlowGraph[outputKey], inputKey)
					}
				}
			}
		}
	}
}

// NewProcessContext creates a fresh process execution context.
func (pe *ProcessExecutor) NewProcessContext() *ProcessExecutionContext {
	ctx := &ProcessExecutionContext{
		ExecutionContext: pe.Executor.NewContext(),
		AvailableInputs:  make(map[string]bool),
		ProducedOutputs:  make(map[string]bool),
		BlockedSteps:     make(map[string][]string),
		StepStatus:       make(map[string]StepExecutionStatus),
	}

	// Calculate initial execution order
	ctx.ExecutionOrder = pe.TopologicalSort()

	// Initialize step statuses
	for _, entity := range pe.Protocol.Entities {
		if entity.IsProcessStep() {
			ctx.StepStatus[entity.ID] = StepStatusPending
		}
	}

	// Mark initially ready steps (no dependencies or all dependencies available)
	pe.updateReadySteps(ctx)

	return ctx
}

// TopologicalSort returns entities in topological execution order.
// Steps with no dependencies come first.
func (pe *ProcessExecutor) TopologicalSort() []string {
	// Kahn's algorithm for topological sort
	inDegree := make(map[string]int)
	entities := make(map[string]bool)

	// Initialize in-degrees
	for _, entity := range pe.Protocol.Entities {
		if entity.IsProcessStep() {
			entities[entity.ID] = true
			inDegree[entity.ID] = 0
		}
	}

	// Calculate in-degrees from dependencies
	for entityID := range entities {
		for _, dep := range pe.Dependencies[entityID] {
			if entities[dep] {
				inDegree[entityID]++
			}
		}
	}

	// Start with nodes that have no dependencies
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// Sort queue for deterministic output
	sort.Strings(queue)

	var result []string
	for len(queue) > 0 {
		// Take first node
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// Reduce in-degree of dependents
		dependents := pe.Dependents[node]
		sort.Strings(dependents)
		for _, dependent := range dependents {
			if !entities[dependent] {
				continue
			}
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				sort.Strings(queue)
			}
		}
	}

	return result
}

// GetBlockedSteps returns steps that cannot execute due to missing inputs.
func (pe *ProcessExecutor) GetBlockedSteps(ctx *ProcessExecutionContext) map[string][]string {
	blocked := make(map[string][]string)

	for _, entity := range pe.Protocol.Entities {
		if !entity.IsProcessStep() {
			continue
		}

		// Check required inputs
		var missingInputs []string
		for _, input := range entity.RequiredInputs() {
			inputKey := entity.ID + "." + input.Name
			if !ctx.AvailableInputs[inputKey] {
				missingInputs = append(missingInputs, input.Name)
			}
		}

		// Check dependencies (preceding steps must be completed)
		for _, depID := range pe.Dependencies[entity.ID] {
			if ctx.StepStatus[depID] != StepStatusCompleted {
				missingInputs = append(missingInputs, fmt.Sprintf("step:%s", depID))
			}
		}

		if len(missingInputs) > 0 {
			blocked[entity.ID] = missingInputs
			ctx.StepStatus[entity.ID] = StepStatusBlocked
		}
	}

	ctx.BlockedSteps = blocked
	return blocked
}

// GetReadySteps returns steps that have all required inputs available.
func (pe *ProcessExecutor) GetReadySteps(ctx *ProcessExecutionContext) []string {
	var ready []string

	for _, entity := range pe.Protocol.Entities {
		if !entity.IsProcessStep() {
			continue
		}

		// Skip completed, failed, or in-progress steps
		status := ctx.StepStatus[entity.ID]
		if status == StepStatusCompleted || status == StepStatusFailed ||
			status == StepStatusInProgress || status == StepStatusSkipped {
			continue
		}

		// Check if all required inputs are available
		allAvailable := true
		for _, input := range entity.RequiredInputs() {
			inputKey := entity.ID + "." + input.Name
			if !ctx.AvailableInputs[inputKey] {
				allAvailable = false
				break
			}
		}

		// Check if all dependencies are completed
		if allAvailable {
			for _, depID := range pe.Dependencies[entity.ID] {
				if ctx.StepStatus[depID] != StepStatusCompleted {
					allAvailable = false
					break
				}
			}
		}

		if allAvailable {
			ready = append(ready, entity.ID)
			ctx.StepStatus[entity.ID] = StepStatusReady
		}
	}

	sort.Strings(ready)
	return ready
}

// updateReadySteps updates the status of steps that are now ready to execute.
func (pe *ProcessExecutor) updateReadySteps(ctx *ProcessExecutionContext) {
	// First, check what's ready (this will update blocked->ready if inputs became available)
	pe.GetReadySteps(ctx)
	// Then mark remaining incomplete steps as blocked
	pe.GetBlockedSteps(ctx)
}

// MarkInputAvailable marks a data port as having data available.
func (pe *ProcessExecutor) MarkInputAvailable(ctx *ProcessExecutionContext, entityID, inputName string) {
	key := entityID + "." + inputName
	ctx.AvailableInputs[key] = true
	pe.updateReadySteps(ctx)
}

// MarkOutputProduced marks an output as produced and propagates to connected inputs.
func (pe *ProcessExecutor) MarkOutputProduced(ctx *ProcessExecutionContext, entityID, outputName string) {
	outputKey := entityID + "." + outputName
	ctx.ProducedOutputs[outputKey] = true

	// Propagate to connected inputs
	for _, inputKey := range pe.DataFlowGraph[outputKey] {
		ctx.AvailableInputs[inputKey] = true
	}

	pe.updateReadySteps(ctx)
}

// CompleteStep marks a step as completed and produces its outputs.
func (pe *ProcessExecutor) CompleteStep(ctx *ProcessExecutionContext, entityID string) error {
	entity := pe.Protocol.EntityByID(entityID)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	if !entity.IsProcessStep() {
		return fmt.Errorf("entity is not a process step: %s", entityID)
	}

	ctx.StepStatus[entityID] = StepStatusCompleted

	// Produce all outputs
	for _, output := range entity.Outputs {
		pe.MarkOutputProduced(ctx, entityID, output.Name)
	}

	return nil
}

// StartStep marks a step as in progress.
func (pe *ProcessExecutor) StartStep(ctx *ProcessExecutionContext, entityID string) error {
	entity := pe.Protocol.EntityByID(entityID)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	if ctx.StepStatus[entityID] != StepStatusReady && ctx.StepStatus[entityID] != StepStatusPending {
		return fmt.Errorf("step %s is not ready to start (status: %s)", entityID, ctx.StepStatus[entityID])
	}

	ctx.StepStatus[entityID] = StepStatusInProgress
	return nil
}

// FailStep marks a step as failed.
func (pe *ProcessExecutor) FailStep(ctx *ProcessExecutionContext, entityID string) error {
	entity := pe.Protocol.EntityByID(entityID)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	ctx.StepStatus[entityID] = StepStatusFailed
	return nil
}

// DependencyAnalysis returns a detailed analysis of step dependencies.
type DependencyAnalysis struct {
	// EntityID is the step being analyzed.
	EntityID string `json:"entity_id"`

	// Name is the entity name.
	Name string `json:"name"`

	// StepType is the step type.
	StepType StepType `json:"step_type"`

	// DirectDependencies are steps that must complete before this one.
	DirectDependencies []string `json:"direct_dependencies"`

	// TransitiveDependencies are all upstream steps (direct and indirect).
	TransitiveDependencies []string `json:"transitive_dependencies"`

	// DirectDependents are steps that depend on this one.
	DirectDependents []string `json:"direct_dependents"`

	// TransitiveDependents are all downstream steps (direct and indirect).
	TransitiveDependents []string `json:"transitive_dependents"`

	// RequiredInputs lists required input ports.
	RequiredInputs []string `json:"required_inputs"`

	// ProducedOutputs lists output ports.
	ProducedOutputs []string `json:"produced_outputs"`

	// CriticalPath indicates if this step is on the critical path.
	CriticalPath bool `json:"critical_path"`
}

// AnalyzeDependencies returns dependency analysis for all process steps.
func (pe *ProcessExecutor) AnalyzeDependencies() []DependencyAnalysis {
	var analyses []DependencyAnalysis

	for _, entity := range pe.Protocol.Entities {
		if !entity.IsProcessStep() {
			continue
		}

		analysis := DependencyAnalysis{
			EntityID:           entity.ID,
			Name:               entity.Name,
			StepType:           entity.StepType,
			DirectDependencies: pe.Dependencies[entity.ID],
			DirectDependents:   pe.Dependents[entity.ID],
		}

		// Calculate transitive dependencies
		analysis.TransitiveDependencies = pe.getTransitiveDependencies(entity.ID)

		// Calculate transitive dependents
		analysis.TransitiveDependents = pe.getTransitiveDependents(entity.ID)

		// List required inputs
		for _, input := range entity.RequiredInputs() {
			analysis.RequiredInputs = append(analysis.RequiredInputs, input.Name)
		}

		// List outputs
		for _, output := range entity.Outputs {
			analysis.ProducedOutputs = append(analysis.ProducedOutputs, output.Name)
		}

		// Check if on critical path (has both upstream and downstream dependencies)
		if len(analysis.TransitiveDependencies) > 0 && len(analysis.TransitiveDependents) > 0 {
			analysis.CriticalPath = true
		}

		analyses = append(analyses, analysis)
	}

	return analyses
}

// getTransitiveDependencies returns all upstream dependencies (direct and indirect).
func (pe *ProcessExecutor) getTransitiveDependencies(entityID string) []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(id string)
	visit = func(id string) {
		for _, dep := range pe.Dependencies[id] {
			if !visited[dep] {
				visited[dep] = true
				result = append(result, dep)
				visit(dep)
			}
		}
	}

	visit(entityID)
	sort.Strings(result)
	return result
}

// getTransitiveDependents returns all downstream dependents (direct and indirect).
func (pe *ProcessExecutor) getTransitiveDependents(entityID string) []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(id string)
	visit = func(id string) {
		for _, dep := range pe.Dependents[id] {
			if !visited[dep] {
				visited[dep] = true
				result = append(result, dep)
				visit(dep)
			}
		}
	}

	visit(entityID)
	sort.Strings(result)
	return result
}

// ExecutionReadiness summarizes the execution readiness of the process.
type ExecutionReadiness struct {
	// TotalSteps is the total number of process steps.
	TotalSteps int `json:"total_steps"`

	// ReadySteps is the number of steps ready to execute.
	ReadySteps int `json:"ready_steps"`

	// BlockedSteps is the number of blocked steps.
	BlockedSteps int `json:"blocked_steps"`

	// CompletedSteps is the number of completed steps.
	CompletedSteps int `json:"completed_steps"`

	// InProgressSteps is the number of steps currently executing.
	InProgressSteps int `json:"in_progress_steps"`

	// FailedSteps is the number of failed steps.
	FailedSteps int `json:"failed_steps"`

	// NextSteps lists the entity IDs of steps ready to execute.
	NextSteps []string `json:"next_steps"`

	// BlockedReasons maps blocked step IDs to their blocking reasons.
	BlockedReasons map[string][]string `json:"blocked_reasons"`
}

// GetExecutionReadiness returns a summary of execution readiness.
func (pe *ProcessExecutor) GetExecutionReadiness(ctx *ProcessExecutionContext) ExecutionReadiness {
	// Update blocked/ready status
	pe.updateReadySteps(ctx)

	readiness := ExecutionReadiness{
		NextSteps:      make([]string, 0),
		BlockedReasons: make(map[string][]string),
	}

	for _, entity := range pe.Protocol.Entities {
		if !entity.IsProcessStep() {
			continue
		}

		readiness.TotalSteps++

		switch ctx.StepStatus[entity.ID] {
		case StepStatusReady:
			readiness.ReadySteps++
			readiness.NextSteps = append(readiness.NextSteps, entity.ID)
		case StepStatusBlocked:
			readiness.BlockedSteps++
			if reasons, ok := ctx.BlockedSteps[entity.ID]; ok {
				readiness.BlockedReasons[entity.ID] = reasons
			}
		case StepStatusCompleted:
			readiness.CompletedSteps++
		case StepStatusInProgress:
			readiness.InProgressSteps++
		case StepStatusFailed:
			readiness.FailedSteps++
		case StepStatusPending:
			// Pending steps that aren't blocked are ready
			if _, blocked := ctx.BlockedSteps[entity.ID]; !blocked {
				readiness.ReadySteps++
				readiness.NextSteps = append(readiness.NextSteps, entity.ID)
			}
		}
	}

	sort.Strings(readiness.NextSteps)
	return readiness
}

// Helper function to append unique strings
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
