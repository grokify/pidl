package pidl

import (
	"fmt"
	"time"
)

// LatencyEstimate represents an estimated latency for a step.
type LatencyEstimate struct {
	// EntityID is the entity this estimate is for (empty for total).
	EntityID string `json:"entity_id,omitempty"`
	// P50 is the 50th percentile latency.
	P50 time.Duration `json:"p50"`
	// P95 is the 95th percentile latency.
	P95 time.Duration `json:"p95"`
	// P99 is the 99th percentile latency.
	P99 time.Duration `json:"p99"`
	// Max is the maximum expected latency.
	Max time.Duration `json:"max"`
	// Expected is the expected typical latency.
	Expected time.Duration `json:"expected"`
	// VarianceClass indicates the latency variability.
	VarianceClass LatencyVarianceClass `json:"variance_class"`
	// OnCriticalPath indicates if this step is on the critical path.
	OnCriticalPath bool `json:"on_critical_path"`
}

// ProcessLatencyAnalysis provides complete latency analysis for a process.
type ProcessLatencyAnalysis struct {
	// ProtocolID is the source protocol.
	ProtocolID string `json:"protocol_id"`
	// TotalLatency is the estimated end-to-end latency.
	TotalLatency LatencyEstimate `json:"total_latency"`
	// StepLatencies are per-step latency estimates.
	StepLatencies []LatencyEstimate `json:"step_latencies"`
	// CriticalPathLatency is the latency of the critical path.
	CriticalPathLatency LatencyEstimate `json:"critical_path_latency"`
	// CriticalPath is the list of entity IDs on the critical path.
	CriticalPath []string `json:"critical_path"`
	// ParallelSavings is the latency saved by parallel execution.
	ParallelSavings time.Duration `json:"parallel_savings"`
	// BudgetViolations lists steps exceeding their latency budget.
	BudgetViolations []LatencyBudgetViolation `json:"budget_violations,omitempty"`
	// LatencyByType shows latencies grouped by step type.
	LatencyByType map[StepType]time.Duration `json:"latency_by_type"`
}

// LatencyBudgetViolation describes a step exceeding its latency budget.
type LatencyBudgetViolation struct {
	// EntityID is the violating entity.
	EntityID string `json:"entity_id"`
	// Metric is the violated metric (p50, p95, p99, max).
	Metric string `json:"metric"`
	// Budget is the budgeted latency.
	Budget time.Duration `json:"budget"`
	// Estimated is the estimated latency.
	Estimated time.Duration `json:"estimated"`
	// Severity indicates the violation severity.
	Severity string `json:"severity"`
}

// StepTypeLatencyDefaults provides default latency estimates by step type.
var StepTypeLatencyDefaults = map[StepType]struct {
	P50      time.Duration
	P95      time.Duration
	P99      time.Duration
	Max      time.Duration
	Variance LatencyVarianceClass
}{
	StepTypeDeterministic: {
		P50:      50 * time.Millisecond,
		P95:      100 * time.Millisecond,
		P99:      200 * time.Millisecond,
		Max:      500 * time.Millisecond,
		Variance: LatencyVarianceLow,
	},
	StepTypeLLM: {
		P50:      2 * time.Second,
		P95:      8 * time.Second,
		P99:      15 * time.Second,
		Max:      30 * time.Second,
		Variance: LatencyVarianceHigh,
	},
	StepTypeHuman: {
		P50:      5 * time.Minute,
		P95:      30 * time.Minute,
		P99:      2 * time.Hour,
		Max:      24 * time.Hour,
		Variance: LatencyVarianceHigh,
	},
	StepTypeExternal: {
		P50:      200 * time.Millisecond,
		P95:      1 * time.Second,
		P99:      3 * time.Second,
		Max:      10 * time.Second,
		Variance: LatencyVarianceMedium,
	},
	StepTypeTool: {
		P50:      100 * time.Millisecond,
		P95:      500 * time.Millisecond,
		P99:      1 * time.Second,
		Max:      5 * time.Second,
		Variance: LatencyVarianceMedium,
	},
}

// AnalyzeProcessLatency calculates latency estimates for a process protocol.
func AnalyzeProcessLatency(p *Protocol) *ProcessLatencyAnalysis {
	analysis := &ProcessLatencyAnalysis{
		ProtocolID:    p.ProtocolMeta.ID,
		StepLatencies: []LatencyEstimate{},
		LatencyByType: make(map[StepType]time.Duration),
	}

	// Build execution graph
	graph := AnalyzeExecutionGraph(p)
	criticalPathSet := make(map[string]bool)
	for _, id := range graph.CriticalPath {
		criticalPathSet[id] = true
	}
	analysis.CriticalPath = graph.CriticalPath

	// Calculate per-step latencies
	var totalP50, totalP95, totalP99, totalMax time.Duration

	for _, entity := range p.Entities {
		estimate := estimateEntityLatency(&entity)
		estimate.OnCriticalPath = criticalPathSet[entity.ID]

		analysis.StepLatencies = append(analysis.StepLatencies, estimate)

		// Aggregate sequential (worst case for total)
		totalP50 += estimate.P50
		totalP95 += estimate.P95
		totalP99 += estimate.P99
		totalMax += estimate.Max

		// Aggregate by type
		if entity.StepType != "" {
			analysis.LatencyByType[entity.StepType] += estimate.Expected
		}

		// Check for budget violations
		violations := checkLatencyBudget(&entity, &estimate)
		analysis.BudgetViolations = append(analysis.BudgetViolations, violations...)
	}

	// Calculate critical path latency
	var cpP50, cpP95, cpP99, cpMax time.Duration
	var cpVariance LatencyVarianceClass = LatencyVarianceLow

	for _, entityID := range graph.CriticalPath {
		for _, est := range analysis.StepLatencies {
			if est.EntityID == entityID {
				cpP50 += est.P50
				cpP95 += est.P95
				cpP99 += est.P99
				cpMax += est.Max

				// Take highest variance
				if est.VarianceClass == LatencyVarianceHigh {
					cpVariance = LatencyVarianceHigh
				} else if est.VarianceClass == LatencyVarianceMedium && cpVariance != LatencyVarianceHigh {
					cpVariance = LatencyVarianceMedium
				}
				break
			}
		}
	}

	analysis.CriticalPathLatency = LatencyEstimate{
		P50:           cpP50,
		P95:           cpP95,
		P99:           cpP99,
		Max:           cpMax,
		Expected:      cpP50, // Use P50 as expected
		VarianceClass: cpVariance,
	}

	// Total latency (use critical path for realistic estimate)
	analysis.TotalLatency = LatencyEstimate{
		P50:           cpP50,
		P95:           cpP95,
		P99:           cpP99,
		Max:           cpMax,
		Expected:      cpP50,
		VarianceClass: cpVariance,
	}

	// Calculate parallel savings
	sequentialP50 := totalP50
	parallelP50 := cpP50
	if sequentialP50 > parallelP50 {
		analysis.ParallelSavings = sequentialP50 - parallelP50
	}

	return analysis
}

func estimateEntityLatency(entity *Entity) LatencyEstimate {
	estimate := LatencyEstimate{
		EntityID: entity.ID,
	}

	// Use explicit latency budget if provided
	if entity.Processing != nil && entity.Processing.LatencyBudget != nil {
		budget := entity.Processing.LatencyBudget

		if d, err := time.ParseDuration(budget.P50); err == nil {
			estimate.P50 = d
		}
		if d, err := time.ParseDuration(budget.P95); err == nil {
			estimate.P95 = d
		}
		if d, err := time.ParseDuration(budget.P99); err == nil {
			estimate.P99 = d
		}
		if d, err := time.ParseDuration(budget.Max); err == nil {
			estimate.Max = d
		}
		if d, err := time.ParseDuration(budget.ExpectedLatency); err == nil {
			estimate.Expected = d
		}

		estimate.VarianceClass = budget.VarianceClass

		// Fill in missing values from defaults or interpolation
		fillMissingLatencyValues(&estimate, entity.StepType)

		return estimate
	}

	// Use timeout as hint if no latency budget
	if entity.Processing != nil && entity.Processing.Timeout != "" {
		if timeout, err := time.ParseDuration(entity.Processing.Timeout); err == nil {
			estimate.Max = timeout
			estimate.P99 = timeout * 8 / 10 // 80% of timeout
			estimate.P95 = timeout * 6 / 10 // 60% of timeout
			estimate.P50 = timeout * 3 / 10 // 30% of timeout
			estimate.Expected = estimate.P50
			estimate.VarianceClass = LatencyVarianceMedium
			return estimate
		}
	}

	// Use defaults based on step type
	return estimateDefaultLatency(entity)
}

func estimateDefaultLatency(entity *Entity) LatencyEstimate {
	estimate := LatencyEstimate{
		EntityID: entity.ID,
	}

	defaults, ok := StepTypeLatencyDefaults[entity.StepType]
	if !ok {
		// Unknown step type - use deterministic defaults
		defaults = StepTypeLatencyDefaults[StepTypeDeterministic]
	}

	estimate.P50 = defaults.P50
	estimate.P95 = defaults.P95
	estimate.P99 = defaults.P99
	estimate.Max = defaults.Max
	estimate.Expected = defaults.P50
	estimate.VarianceClass = defaults.Variance

	return estimate
}

func fillMissingLatencyValues(estimate *LatencyEstimate, stepType StepType) {
	defaults, ok := StepTypeLatencyDefaults[stepType]
	if !ok {
		defaults = StepTypeLatencyDefaults[StepTypeDeterministic]
	}

	// Fill missing values with interpolation or defaults
	if estimate.P50 == 0 {
		if estimate.Expected > 0 {
			estimate.P50 = estimate.Expected
		} else if estimate.P95 > 0 {
			estimate.P50 = estimate.P95 / 2
		} else {
			estimate.P50 = defaults.P50
		}
	}

	if estimate.Expected == 0 {
		estimate.Expected = estimate.P50
	}

	if estimate.P95 == 0 {
		if estimate.P99 > 0 {
			estimate.P95 = estimate.P99 * 8 / 10
		} else {
			estimate.P95 = estimate.P50 * 3
		}
	}

	if estimate.P99 == 0 {
		estimate.P99 = estimate.P95 * 15 / 10
	}

	if estimate.Max == 0 {
		estimate.Max = estimate.P99 * 2
	}

	if estimate.VarianceClass == "" {
		estimate.VarianceClass = defaults.Variance
	}
}

func checkLatencyBudget(entity *Entity, estimate *LatencyEstimate) []LatencyBudgetViolation {
	var violations []LatencyBudgetViolation

	if entity.Processing == nil || entity.Processing.LatencyBudget == nil {
		return violations
	}

	budget := entity.Processing.LatencyBudget

	// Check P50
	if budget.P50 != "" {
		if budgetP50, err := time.ParseDuration(budget.P50); err == nil {
			if estimate.P50 > budgetP50 {
				violations = append(violations, LatencyBudgetViolation{
					EntityID:  entity.ID,
					Metric:    "p50",
					Budget:    budgetP50,
					Estimated: estimate.P50,
					Severity:  "warning",
				})
			}
		}
	}

	// Check P95
	if budget.P95 != "" {
		if budgetP95, err := time.ParseDuration(budget.P95); err == nil {
			if estimate.P95 > budgetP95 {
				violations = append(violations, LatencyBudgetViolation{
					EntityID:  entity.ID,
					Metric:    "p95",
					Budget:    budgetP95,
					Estimated: estimate.P95,
					Severity:  "high",
				})
			}
		}
	}

	// Check P99
	if budget.P99 != "" {
		if budgetP99, err := time.ParseDuration(budget.P99); err == nil {
			if estimate.P99 > budgetP99 {
				violations = append(violations, LatencyBudgetViolation{
					EntityID:  entity.ID,
					Metric:    "p99",
					Budget:    budgetP99,
					Estimated: estimate.P99,
					Severity:  "high",
				})
			}
		}
	}

	// Check Max
	if budget.Max != "" {
		if budgetMax, err := time.ParseDuration(budget.Max); err == nil {
			if estimate.Max > budgetMax {
				violations = append(violations, LatencyBudgetViolation{
					EntityID:  entity.ID,
					Metric:    "max",
					Budget:    budgetMax,
					Estimated: estimate.Max,
					Severity:  "critical",
				})
			}
		}
	}

	return violations
}

// GetLatencyBreakdown returns latency contribution by step type.
func GetLatencyBreakdown(analysis *ProcessLatencyAnalysis) map[string]time.Duration {
	breakdown := make(map[string]time.Duration)

	for stepType, latency := range analysis.LatencyByType {
		breakdown[string(stepType)] = latency
	}

	return breakdown
}

// EstimateLatencyPercentile returns a specific percentile latency estimate.
func EstimateLatencyPercentile(analysis *ProcessLatencyAnalysis, percentile int) time.Duration {
	switch percentile {
	case 50:
		return analysis.TotalLatency.P50
	case 95:
		return analysis.TotalLatency.P95
	case 99:
		return analysis.TotalLatency.P99
	default:
		// Interpolate between known percentiles
		if percentile < 50 {
			// Below P50 - use linear interpolation from 0
			return analysis.TotalLatency.P50 * time.Duration(percentile) / 50
		} else if percentile < 95 {
			// Between P50 and P95
			ratio := time.Duration(percentile-50) * (analysis.TotalLatency.P95 - analysis.TotalLatency.P50) / 45
			return analysis.TotalLatency.P50 + ratio
		} else if percentile < 99 {
			// Between P95 and P99
			ratio := time.Duration(percentile-95) * (analysis.TotalLatency.P99 - analysis.TotalLatency.P95) / 4
			return analysis.TotalLatency.P95 + ratio
		}
		// Above P99
		return analysis.TotalLatency.Max
	}
}

// CalculateActualLatency calculates latency from execution metrics.
func CalculateActualLatency(entity *Entity, duration time.Duration) *LatencyMeasurement {
	measurement := &LatencyMeasurement{
		EntityID: entity.ID,
		Duration: duration,
	}

	if entity.Processing == nil || entity.Processing.LatencyBudget == nil {
		return measurement
	}

	budget := entity.Processing.LatencyBudget

	// Check against budget
	if budget.P50 != "" {
		if budgetP50, err := time.ParseDuration(budget.P50); err == nil {
			if duration > budgetP50 {
				measurement.ExceedsP50 = true
			}
		}
	}
	if budget.P95 != "" {
		if budgetP95, err := time.ParseDuration(budget.P95); err == nil {
			if duration > budgetP95 {
				measurement.ExceedsP95 = true
			}
		}
	}
	if budget.P99 != "" {
		if budgetP99, err := time.ParseDuration(budget.P99); err == nil {
			if duration > budgetP99 {
				measurement.ExceedsP99 = true
			}
		}
	}
	if budget.Max != "" {
		if budgetMax, err := time.ParseDuration(budget.Max); err == nil {
			if duration > budgetMax {
				measurement.ExceedsMax = true
			}
		}
	}

	return measurement
}

// LatencyMeasurement records actual execution latency.
type LatencyMeasurement struct {
	// EntityID is the entity that was measured.
	EntityID string `json:"entity_id"`
	// Duration is the actual execution duration.
	Duration time.Duration `json:"duration"`
	// ExceedsP50 indicates if duration exceeded P50 budget.
	ExceedsP50 bool `json:"exceeds_p50,omitempty"`
	// ExceedsP95 indicates if duration exceeded P95 budget.
	ExceedsP95 bool `json:"exceeds_p95,omitempty"`
	// ExceedsP99 indicates if duration exceeded P99 budget.
	ExceedsP99 bool `json:"exceeds_p99,omitempty"`
	// ExceedsMax indicates if duration exceeded Max budget.
	ExceedsMax bool `json:"exceeds_max,omitempty"`
}

// FormatLatencyReport generates a human-readable latency report.
func FormatLatencyReport(analysis *ProcessLatencyAnalysis) string {
	var report string

	report += fmt.Sprintf("Latency Analysis: %s\n", analysis.ProtocolID)
	report += fmt.Sprintf("================================\n\n")

	report += fmt.Sprintf("Total Latency:\n")
	report += fmt.Sprintf("  P50: %s\n", analysis.TotalLatency.P50)
	report += fmt.Sprintf("  P95: %s\n", analysis.TotalLatency.P95)
	report += fmt.Sprintf("  P99: %s\n", analysis.TotalLatency.P99)
	report += fmt.Sprintf("  Max: %s\n\n", analysis.TotalLatency.Max)

	report += fmt.Sprintf("Critical Path: %v\n", analysis.CriticalPath)
	report += fmt.Sprintf("Critical Path Latency: %s (P50)\n\n", analysis.CriticalPathLatency.P50)

	if analysis.ParallelSavings > 0 {
		report += fmt.Sprintf("Parallel Savings: %s\n\n", analysis.ParallelSavings)
	}

	if len(analysis.BudgetViolations) > 0 {
		report += fmt.Sprintf("Budget Violations:\n")
		for _, v := range analysis.BudgetViolations {
			report += fmt.Sprintf("  - %s: %s exceeded (budget: %s, estimated: %s) [%s]\n",
				v.EntityID, v.Metric, v.Budget, v.Estimated, v.Severity)
		}
		report += "\n"
	}

	report += fmt.Sprintf("Step Latencies:\n")
	for _, est := range analysis.StepLatencies {
		cpMarker := ""
		if est.OnCriticalPath {
			cpMarker = " [CRITICAL PATH]"
		}
		report += fmt.Sprintf("  %s: P50=%s, P95=%s, P99=%s%s\n",
			est.EntityID, est.P50, est.P95, est.P99, cpMarker)
	}

	return report
}
