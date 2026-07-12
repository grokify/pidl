package pidl

import (
	"fmt"
	"time"
)

// ParallelBlock represents a detected parallel execution block.
type ParallelBlock struct {
	// ID is the unique identifier for this block.
	ID string
	// ForkEntity is the entity that starts the parallel execution.
	ForkEntity string
	// JoinEntity is the entity that collects results (if any).
	JoinEntity string
	// Branches are the parallel branches.
	Branches []ParallelBranch
	// Mode is the parallel execution mode.
	Mode ParallelMode
	// JoinCondition specifies completion criteria.
	JoinCondition JoinCondition
}

// ExecutionGraph represents the execution order of steps.
type ExecutionGraph struct {
	// ProtocolID is the source protocol ID.
	ProtocolID string
	// Stages are ordered groups of steps that can execute in parallel.
	Stages []ExecutionStage
	// CriticalPath is the longest path through the graph.
	CriticalPath []string
	// EstimatedDuration is the estimated total execution time.
	EstimatedDuration time.Duration
}

// ExecutionStage represents a group of steps that can execute concurrently.
type ExecutionStage struct {
	// ID is the stage identifier.
	ID string
	// Steps are entity IDs that can execute in parallel at this stage.
	Steps []string
	// Dependencies are stage IDs that must complete before this stage.
	Dependencies []string
	// IsParallelBlock indicates this stage is an explicit parallel block.
	IsParallelBlock bool
	// ParallelConfig is the configuration for parallel execution.
	ParallelConfig *ParallelConfig
}

// AnalyzeExecutionGraph builds an execution graph from a process protocol.
func AnalyzeExecutionGraph(p *Protocol) *ExecutionGraph {
	graph := &ExecutionGraph{
		ProtocolID: p.ProtocolMeta.ID,
		Stages:     []ExecutionStage{},
	}

	// Build dependency graph from flows
	dependencies := make(map[string][]string) // entity -> entities it depends on
	dependents := make(map[string][]string)   // entity -> entities that depend on it
	entitySet := make(map[string]bool)

	for _, entity := range p.Entities {
		entitySet[entity.ID] = true
		dependencies[entity.ID] = []string{}
		dependents[entity.ID] = []string{}
	}

	for _, flow := range p.Flows {
		if _, ok := entitySet[flow.From]; !ok {
			continue
		}
		if _, ok := entitySet[flow.To]; !ok {
			continue
		}
		dependencies[flow.To] = append(dependencies[flow.To], flow.From)
		dependents[flow.From] = append(dependents[flow.From], flow.To)
	}

	// Topological sort into stages
	completed := make(map[string]bool)
	stageNum := 0

	for len(completed) < len(entitySet) {
		// Find all entities with satisfied dependencies
		ready := []string{}
		for id := range entitySet {
			if completed[id] {
				continue
			}

			// Check if all dependencies are satisfied
			allSatisfied := true
			for _, dep := range dependencies[id] {
				if !completed[dep] {
					allSatisfied = false
					break
				}
			}

			if allSatisfied {
				ready = append(ready, id)
			}
		}

		if len(ready) == 0 {
			// Cycle detected or error
			break
		}

		// Create stage with ready entities
		stage := ExecutionStage{
			ID:    fmt.Sprintf("stage-%d", stageNum),
			Steps: ready,
		}

		// Add dependencies on previous stages
		if stageNum > 0 {
			stage.Dependencies = []string{fmt.Sprintf("stage-%d", stageNum-1)}
		}

		// Check if any entity has parallel config
		for _, id := range ready {
			entity := findEntity(p, id)
			if entity != nil && entity.Parallel != nil {
				stage.IsParallelBlock = true
				stage.ParallelConfig = entity.Parallel
				break
			}
		}

		graph.Stages = append(graph.Stages, stage)

		// Mark as completed
		for _, id := range ready {
			completed[id] = true
		}

		stageNum++
	}

	// Calculate critical path
	graph.CriticalPath = calculateCriticalPath(p, dependencies)

	// Estimate duration
	graph.EstimatedDuration = estimateDuration(p, graph.CriticalPath)

	return graph
}

func findEntity(p *Protocol, id string) *Entity {
	for i := range p.Entities {
		if p.Entities[i].ID == id {
			return &p.Entities[i]
		}
	}
	return nil
}

func calculateCriticalPath(p *Protocol, dependencies map[string][]string) []string {
	// Simple longest path algorithm
	entityMap := make(map[string]*Entity)
	for i := range p.Entities {
		entityMap[p.Entities[i].ID] = &p.Entities[i]
	}

	// Calculate longest path to each node
	longestPath := make(map[string][]string)
	for id := range entityMap {
		longestPath[id] = []string{}
	}

	// Topological order processing
	visited := make(map[string]bool)
	var order []string

	var visit func(id string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		for _, dep := range dependencies[id] {
			visit(dep)
		}
		order = append(order, id)
	}

	for id := range entityMap {
		visit(id)
	}

	// Calculate longest paths
	for _, id := range order {
		maxPath := []string{}
		for _, dep := range dependencies[id] {
			if len(longestPath[dep]) > len(maxPath) {
				maxPath = longestPath[dep]
			}
		}
		longestPath[id] = append(append([]string{}, maxPath...), id)
	}

	// Find the longest of all paths
	var criticalPath []string
	for _, path := range longestPath {
		if len(path) > len(criticalPath) {
			criticalPath = path
		}
	}

	return criticalPath
}

func estimateDuration(p *Protocol, criticalPath []string) time.Duration {
	var total time.Duration

	for _, id := range criticalPath {
		entity := findEntity(p, id)
		if entity == nil {
			continue
		}

		// Use processing timeout if available
		if entity.Processing != nil && entity.Processing.Timeout != "" {
			if d, err := time.ParseDuration(entity.Processing.Timeout); err == nil {
				total += d
			}
		} else {
			// Default estimates based on step type
			switch entity.StepType {
			case StepTypeDeterministic:
				total += 100 * time.Millisecond
			case StepTypeLLM:
				total += 5 * time.Second
			case StepTypeHuman:
				total += time.Hour // Human steps are slow
			case StepTypeExternal:
				total += time.Second
			case StepTypeTool:
				total += 500 * time.Millisecond
			default:
				total += 100 * time.Millisecond
			}
		}
	}

	return total
}

// DetectParallelBlocks identifies explicit parallel execution blocks in the protocol.
func DetectParallelBlocks(p *Protocol) []ParallelBlock {
	var blocks []ParallelBlock

	for _, entity := range p.Entities {
		if entity.Parallel != nil {
			block := ParallelBlock{
				ID:            entity.ID + "-parallel",
				ForkEntity:    entity.ID,
				Branches:      entity.Parallel.Branches,
				Mode:          entity.Parallel.Mode,
				JoinCondition: entity.Parallel.JoinCondition,
			}

			// Try to find join entity (entity that depends on all branches)
			block.JoinEntity = findJoinEntity(p, entity.Parallel.Branches)

			blocks = append(blocks, block)
		}
	}

	return blocks
}

func findJoinEntity(p *Protocol, branches []ParallelBranch) string {
	if len(branches) == 0 {
		return ""
	}

	// Collect all branch entity IDs
	branchEntities := make(map[string]bool)
	for _, branch := range branches {
		if branch.EntityID != "" {
			branchEntities[branch.EntityID] = true
		}
	}

	if len(branchEntities) == 0 {
		return ""
	}

	// Find entity that receives flows from all branch entities
	for _, entity := range p.Entities {
		if branchEntities[entity.ID] {
			continue
		}

		incomingFrom := make(map[string]bool)
		for _, flow := range p.Flows {
			if flow.To == entity.ID && branchEntities[flow.From] {
				incomingFrom[flow.From] = true
			}
		}

		if len(incomingFrom) == len(branchEntities) {
			return entity.ID
		}
	}

	return ""
}

// CanExecuteInParallel determines if two entities can execute concurrently.
func CanExecuteInParallel(p *Protocol, entityA, entityB string) bool {
	// Check if there's any dependency between them
	for _, flow := range p.Flows {
		if (flow.From == entityA && flow.To == entityB) ||
			(flow.From == entityB && flow.To == entityA) {
			return false
		}
	}

	// Check transitive dependencies
	dependsOnA := getAllDependencies(p, entityA)
	dependsOnB := getAllDependencies(p, entityB)

	if dependsOnA[entityB] || dependsOnB[entityA] {
		return false
	}

	return true
}

func getAllDependencies(p *Protocol, entityID string) map[string]bool {
	deps := make(map[string]bool)
	visited := make(map[string]bool)

	var collect func(id string)
	collect = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true

		for _, flow := range p.Flows {
			if flow.To == id {
				deps[flow.From] = true
				collect(flow.From)
			}
		}
	}

	collect(entityID)
	return deps
}

// GetMaxParallelism returns the maximum number of entities that can execute concurrently.
func GetMaxParallelism(p *Protocol) int {
	graph := AnalyzeExecutionGraph(p)

	maxParallel := 0
	for _, stage := range graph.Stages {
		if len(stage.Steps) > maxParallel {
			maxParallel = len(stage.Steps)
		}
	}

	return maxParallel
}
