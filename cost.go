package pidl

import (
	"time"
)

// CostEstimate represents an estimated cost for a step or process.
type CostEstimate struct {
	// EntityID is the entity this estimate is for (empty for total).
	EntityID string `json:"entity_id,omitempty"`
	// MinCost is the minimum expected cost.
	MinCost float64 `json:"min_cost"`
	// MaxCost is the maximum expected cost.
	MaxCost float64 `json:"max_cost"`
	// ExpectedCost is the average/expected cost.
	ExpectedCost float64 `json:"expected_cost"`
	// CostUnit is the currency or unit.
	CostUnit string `json:"cost_unit"`
	// Breakdown shows cost by category.
	Breakdown map[string]float64 `json:"breakdown,omitempty"`
}

// ProcessCostAnalysis provides complete cost analysis for a process.
type ProcessCostAnalysis struct {
	// ProtocolID is the source protocol.
	ProtocolID string `json:"protocol_id"`
	// TotalEstimate is the total estimated cost.
	TotalEstimate CostEstimate `json:"total_estimate"`
	// StepEstimates are per-step cost estimates.
	StepEstimates []CostEstimate `json:"step_estimates"`
	// CostByType shows costs grouped by step type.
	CostByType map[StepType]float64 `json:"cost_by_type"`
	// CriticalPathCost is the cost of the critical path.
	CriticalPathCost float64 `json:"critical_path_cost"`
	// ParallelSavings is estimated savings from parallelization.
	ParallelSavings float64 `json:"parallel_savings,omitempty"`
}

// LLMCostPresets provides common LLM pricing (per 1K tokens, in USD).
var LLMCostPresets = map[string]struct {
	InputCost  float64
	OutputCost float64
}{
	"gpt-4":           {InputCost: 0.03, OutputCost: 0.06},
	"gpt-4-turbo":     {InputCost: 0.01, OutputCost: 0.03},
	"gpt-3.5-turbo":   {InputCost: 0.0005, OutputCost: 0.0015},
	"claude-3-opus":   {InputCost: 0.015, OutputCost: 0.075},
	"claude-3-sonnet": {InputCost: 0.003, OutputCost: 0.015},
	"claude-3-haiku":  {InputCost: 0.00025, OutputCost: 0.00125},
}

// AnalyzeProcessCosts calculates cost estimates for a process protocol.
func AnalyzeProcessCosts(p *Protocol) *ProcessCostAnalysis {
	analysis := &ProcessCostAnalysis{
		ProtocolID:    p.ProtocolMeta.ID,
		StepEstimates: []CostEstimate{},
		CostByType:    make(map[StepType]float64),
	}

	totalMin := 0.0
	totalMax := 0.0
	totalExpected := 0.0
	defaultUnit := "USD"

	for _, entity := range p.Entities {
		estimate := estimateEntityCost(&entity)

		if estimate.CostUnit != "" {
			defaultUnit = estimate.CostUnit
		}

		analysis.StepEstimates = append(analysis.StepEstimates, estimate)

		totalMin += estimate.MinCost
		totalMax += estimate.MaxCost
		totalExpected += estimate.ExpectedCost

		// Aggregate by type
		if entity.StepType != "" {
			analysis.CostByType[entity.StepType] += estimate.ExpectedCost
		}
	}

	analysis.TotalEstimate = CostEstimate{
		MinCost:      totalMin,
		MaxCost:      totalMax,
		ExpectedCost: totalExpected,
		CostUnit:     defaultUnit,
		Breakdown:    make(map[string]float64),
	}

	// Add breakdown
	for stepType, cost := range analysis.CostByType {
		analysis.TotalEstimate.Breakdown[string(stepType)] = cost
	}

	// Calculate critical path cost
	graph := AnalyzeExecutionGraph(p)
	for _, entityID := range graph.CriticalPath {
		for _, est := range analysis.StepEstimates {
			if est.EntityID == entityID {
				analysis.CriticalPathCost += est.ExpectedCost
				break
			}
		}
	}

	// Estimate parallel savings
	if len(graph.Stages) > 0 {
		sequentialCost := totalExpected
		parallelCost := analysis.CriticalPathCost
		analysis.ParallelSavings = sequentialCost - parallelCost
	}

	return analysis
}

func estimateEntityCost(entity *Entity) CostEstimate {
	estimate := CostEstimate{
		EntityID: entity.ID,
		CostUnit: "USD",
	}

	if entity.Processing == nil || entity.Processing.CostModel == nil {
		// Use defaults based on step type
		return estimateDefaultCost(entity)
	}

	model := entity.Processing.CostModel
	estimate.CostUnit = model.CostUnit
	if estimate.CostUnit == "" {
		estimate.CostUnit = "USD"
	}

	switch model.Type {
	case CostModelTypeFixed:
		estimate.MinCost = model.FixedCost
		estimate.MaxCost = model.FixedCost
		estimate.ExpectedCost = model.FixedCost

	case CostModelTypeTokenBased:
		inputCost := float64(model.EstimatedInputTokens) * model.InputTokenCost / 1000
		outputCost := float64(model.EstimatedOutputTokens) * model.OutputTokenCost / 1000
		estimate.ExpectedCost = inputCost + outputCost
		// Variance of ±50% for token-based
		estimate.MinCost = estimate.ExpectedCost * 0.5
		estimate.MaxCost = estimate.ExpectedCost * 1.5
		estimate.Breakdown = map[string]float64{
			"input_tokens":  inputCost,
			"output_tokens": outputCost,
		}

	case CostModelTypeTimeBased:
		if entity.Processing.Timeout != "" {
			if duration, err := time.ParseDuration(entity.Processing.Timeout); err == nil {
				seconds := duration.Seconds()
				estimate.MaxCost = seconds * model.ComputeCostPerSecond
				estimate.MinCost = estimate.MaxCost * 0.1 // Assume 10% of timeout on average
				estimate.ExpectedCost = estimate.MaxCost * 0.5
			}
		}

	case CostModelTypeAPICall:
		estimate.MinCost = model.APICallCost
		estimate.MaxCost = model.APICallCost
		estimate.ExpectedCost = model.APICallCost

	case CostModelTypeHybrid:
		// Combine fixed + variable components
		estimate.ExpectedCost = model.FixedCost + model.VariableCost
		estimate.MinCost = model.FixedCost
		estimate.MaxCost = model.FixedCost + model.VariableCost*2
	}

	return estimate
}

func estimateDefaultCost(entity *Entity) CostEstimate {
	estimate := CostEstimate{
		EntityID: entity.ID,
		CostUnit: "USD",
	}

	// Default costs based on step type
	switch entity.StepType {
	case StepTypeLLM:
		// Assume Claude-3-Sonnet with ~1K input, 500 output tokens
		preset := LLMCostPresets["claude-3-sonnet"]
		inputCost := 1.0 * preset.InputCost   // 1K tokens
		outputCost := 0.5 * preset.OutputCost // 500 tokens
		estimate.ExpectedCost = inputCost + outputCost
		estimate.MinCost = estimate.ExpectedCost * 0.2
		estimate.MaxCost = estimate.ExpectedCost * 5.0

	case StepTypeExternal:
		// Typical API call cost
		estimate.ExpectedCost = 0.001 // $0.001 per call
		estimate.MinCost = 0.0
		estimate.MaxCost = 0.01

	case StepTypeHuman:
		// Human review - expensive in time, not direct cost
		estimate.ExpectedCost = 0.0 // No direct cost
		estimate.MinCost = 0.0
		estimate.MaxCost = 0.0

	case StepTypeTool:
		// Tool invocation - minimal cost
		estimate.ExpectedCost = 0.0001
		estimate.MinCost = 0.0
		estimate.MaxCost = 0.001

	default:
		// Deterministic - compute only
		estimate.ExpectedCost = 0.00001
		estimate.MinCost = 0.0
		estimate.MaxCost = 0.0001
	}

	return estimate
}

// CalculateExecutionCost calculates actual cost from execution metrics.
func CalculateExecutionCost(entity *Entity, metrics ExecutionMetrics) float64 {
	if entity.Processing == nil || entity.Processing.CostModel == nil {
		return 0.0
	}

	model := entity.Processing.CostModel

	switch model.Type {
	case CostModelTypeFixed:
		return model.FixedCost

	case CostModelTypeTokenBased:
		inputCost := float64(metrics.InputTokens) * model.InputTokenCost / 1000
		outputCost := float64(metrics.OutputTokens) * model.OutputTokenCost / 1000
		return inputCost + outputCost

	case CostModelTypeTimeBased:
		seconds := metrics.Duration.Seconds()
		return seconds * model.ComputeCostPerSecond

	case CostModelTypeAPICall:
		return model.APICallCost * float64(metrics.APICallCount)

	case CostModelTypeHybrid:
		cost := model.FixedCost
		cost += float64(metrics.InputTokens) * model.InputTokenCost / 1000
		cost += float64(metrics.OutputTokens) * model.OutputTokenCost / 1000
		cost += metrics.Duration.Seconds() * model.ComputeCostPerSecond
		cost += model.APICallCost * float64(metrics.APICallCount)
		return cost

	default:
		return 0.0
	}
}

// ExecutionMetrics captures actual execution data for cost calculation.
type ExecutionMetrics struct {
	// Duration is the actual execution time.
	Duration time.Duration `json:"duration"`
	// InputTokens is the actual input token count.
	InputTokens int `json:"input_tokens,omitempty"`
	// OutputTokens is the actual output token count.
	OutputTokens int `json:"output_tokens,omitempty"`
	// APICallCount is the number of API calls made.
	APICallCount int `json:"api_call_count,omitempty"`
}

// GetCostEfficiency returns cost per unit of output for comparison.
func GetCostEfficiency(analysis *ProcessCostAnalysis) map[string]float64 {
	efficiency := make(map[string]float64)

	for stepType, cost := range analysis.CostByType {
		if cost > 0 {
			efficiency[string(stepType)] = cost
		}
	}

	return efficiency
}
