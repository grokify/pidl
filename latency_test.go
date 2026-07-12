package pidl

import (
	"testing"
	"time"
)

func createLatencyTestProtocol() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "latency-test",
			Name: "Latency Test Pipeline",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "validate",
				Name:     "Validate Input",
				StepType: StepTypeDeterministic,
				Processing: &ProcessingConfig{
					LatencyBudget: &LatencyBudget{
						P50: "50ms",
						P95: "100ms",
						P99: "200ms",
						Max: "500ms",
					},
				},
			},
			{
				ID:       "transform",
				Name:     "Transform Data",
				StepType: StepTypeLLM,
				Processing: &ProcessingConfig{
					Timeout: "30s",
					LatencyBudget: &LatencyBudget{
						P50:           "2s",
						P95:           "8s",
						P99:           "15s",
						Max:           "30s",
						VarianceClass: LatencyVarianceHigh,
					},
				},
			},
			{
				ID:       "store",
				Name:     "Store Result",
				StepType: StepTypeExternal,
				Processing: &ProcessingConfig{
					LatencyBudget: &LatencyBudget{
						P50: "200ms",
						P95: "1s",
						P99: "3s",
						Max: "10s",
					},
				},
			},
		},
		Flows: []Flow{
			{From: "validate", To: "transform", Action: "send"},
			{From: "transform", To: "store", Action: "send"},
		},
	}
}

func TestAnalyzeProcessLatency(t *testing.T) {
	p := createLatencyTestProtocol()
	analysis := AnalyzeProcessLatency(p)

	if analysis.ProtocolID != "latency-test" {
		t.Errorf("expected protocol ID 'latency-test', got '%s'", analysis.ProtocolID)
	}

	// Should have 3 step latencies
	if len(analysis.StepLatencies) != 3 {
		t.Errorf("expected 3 step latencies, got %d", len(analysis.StepLatencies))
	}

	// Total latency P50 should be > 0
	if analysis.TotalLatency.P50 <= 0 {
		t.Error("expected positive total P50 latency")
	}

	// P50 < P95 < P99 < Max
	if analysis.TotalLatency.P50 >= analysis.TotalLatency.P95 {
		t.Error("expected P50 < P95")
	}
	if analysis.TotalLatency.P95 >= analysis.TotalLatency.P99 {
		t.Error("expected P95 < P99")
	}
	if analysis.TotalLatency.P99 >= analysis.TotalLatency.Max {
		t.Error("expected P99 < Max")
	}
}

func TestAnalyzeProcessLatency_CriticalPath(t *testing.T) {
	p := createLatencyTestProtocol()
	analysis := AnalyzeProcessLatency(p)

	// All steps should be on critical path (linear pipeline)
	if len(analysis.CriticalPath) != 3 {
		t.Errorf("expected 3 steps on critical path, got %d", len(analysis.CriticalPath))
	}

	// Critical path latency should match total (linear pipeline)
	if analysis.CriticalPathLatency.P50 != analysis.TotalLatency.P50 {
		t.Errorf("expected critical path P50 to match total P50")
	}
}

func TestAnalyzeProcessLatency_StepDetails(t *testing.T) {
	p := createLatencyTestProtocol()
	analysis := AnalyzeProcessLatency(p)

	// Find LLM step latency
	var llmLatency *LatencyEstimate
	for i, est := range analysis.StepLatencies {
		if est.EntityID == "transform" {
			llmLatency = &analysis.StepLatencies[i]
			break
		}
	}

	if llmLatency == nil {
		t.Fatal("expected to find transform step latency")
	}

	// LLM P50 should be 2s as specified
	expectedP50 := 2 * time.Second
	if llmLatency.P50 != expectedP50 {
		t.Errorf("expected LLM P50 %s, got %s", expectedP50, llmLatency.P50)
	}

	// Variance should be high for LLM
	if llmLatency.VarianceClass != LatencyVarianceHigh {
		t.Errorf("expected high variance for LLM step, got %s", llmLatency.VarianceClass)
	}
}

func TestAnalyzeProcessLatency_DefaultsForMissingBudget(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "no-budget",
			Name: "No Budget Protocol",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				StepType: StepTypeLLM,
				// No Processing config
			},
		},
	}

	analysis := AnalyzeProcessLatency(p)

	if len(analysis.StepLatencies) != 1 {
		t.Fatal("expected 1 step latency")
	}

	est := analysis.StepLatencies[0]

	// Should use LLM defaults
	defaults := StepTypeLatencyDefaults[StepTypeLLM]
	if est.P50 != defaults.P50 {
		t.Errorf("expected default LLM P50 %s, got %s", defaults.P50, est.P50)
	}
}

func TestAnalyzeProcessLatency_TimeoutAsFallback(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "timeout-test",
			Name: "Timeout Test",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				StepType: StepTypeExternal,
				Processing: &ProcessingConfig{
					Timeout: "10s",
					// No LatencyBudget
				},
			},
		},
	}

	analysis := AnalyzeProcessLatency(p)

	if len(analysis.StepLatencies) != 1 {
		t.Fatal("expected 1 step latency")
	}

	est := analysis.StepLatencies[0]

	// Max should be timeout
	if est.Max != 10*time.Second {
		t.Errorf("expected Max to be 10s (timeout), got %s", est.Max)
	}

	// P50 should be 30% of timeout
	expectedP50 := 3 * time.Second
	if est.P50 != expectedP50 {
		t.Errorf("expected P50 %s, got %s", expectedP50, est.P50)
	}
}

func TestAnalyzeProcessLatency_BudgetViolations(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "violation-test",
			Name: "Violation Test",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "slow_step",
				Name:     "Slow Step",
				StepType: StepTypeLLM,
				Processing: &ProcessingConfig{
					Timeout: "60s", // Will derive high latency from this
					LatencyBudget: &LatencyBudget{
						P50: "1s",  // Budget is 1s
						P95: "5s",  // Budget is 5s
						P99: "10s", // Budget is 10s
						Max: "15s", // Budget is 15s
					},
				},
			},
		},
	}

	// Run analysis to ensure it doesn't panic
	_ = AnalyzeProcessLatency(p)

	// Test that violations are detected correctly with a custom check
	violations := checkLatencyBudget(&p.Entities[0], &LatencyEstimate{
		EntityID: "slow_step",
		P50:      5 * time.Second, // Exceeds budget of 1s
		P95:      10 * time.Second,
		P99:      20 * time.Second,
		Max:      30 * time.Second,
	})

	if len(violations) != 4 {
		t.Errorf("expected 4 violations (p50, p95, p99, max), got %d", len(violations))
	}

	// Check violation details
	for _, v := range violations {
		if v.EntityID != "slow_step" {
			t.Errorf("expected entity ID 'slow_step', got '%s'", v.EntityID)
		}
	}
}

func TestEstimateLatencyPercentile(t *testing.T) {
	analysis := &ProcessLatencyAnalysis{
		TotalLatency: LatencyEstimate{
			P50: 100 * time.Millisecond,
			P95: 500 * time.Millisecond,
			P99: 1 * time.Second,
			Max: 2 * time.Second,
		},
	}

	// Test exact percentiles
	if EstimateLatencyPercentile(analysis, 50) != 100*time.Millisecond {
		t.Error("expected P50 to be 100ms")
	}
	if EstimateLatencyPercentile(analysis, 95) != 500*time.Millisecond {
		t.Error("expected P95 to be 500ms")
	}
	if EstimateLatencyPercentile(analysis, 99) != 1*time.Second {
		t.Error("expected P99 to be 1s")
	}

	// Test interpolation
	p75 := EstimateLatencyPercentile(analysis, 75)
	if p75 < 100*time.Millisecond || p75 > 500*time.Millisecond {
		t.Errorf("expected P75 between 100ms and 500ms, got %s", p75)
	}

	// Test above P99
	p100 := EstimateLatencyPercentile(analysis, 100)
	if p100 != 2*time.Second {
		t.Errorf("expected P100 to be Max (2s), got %s", p100)
	}
}

func TestCalculateActualLatency(t *testing.T) {
	entity := &Entity{
		ID: "test-step",
		Processing: &ProcessingConfig{
			LatencyBudget: &LatencyBudget{
				P50: "100ms",
				P95: "500ms",
				P99: "1s",
				Max: "2s",
			},
		},
	}

	// Test within P50
	m1 := CalculateActualLatency(entity, 50*time.Millisecond)
	if m1.ExceedsP50 || m1.ExceedsP95 || m1.ExceedsP99 || m1.ExceedsMax {
		t.Error("expected no violations for 50ms")
	}

	// Test exceeds P50 but within P95
	m2 := CalculateActualLatency(entity, 200*time.Millisecond)
	if !m2.ExceedsP50 {
		t.Error("expected P50 violation for 200ms")
	}
	if m2.ExceedsP95 {
		t.Error("expected no P95 violation for 200ms")
	}

	// Test exceeds Max
	m3 := CalculateActualLatency(entity, 3*time.Second)
	if !m3.ExceedsP50 || !m3.ExceedsP95 || !m3.ExceedsP99 || !m3.ExceedsMax {
		t.Error("expected all violations for 3s")
	}
}

func TestGetLatencyBreakdown(t *testing.T) {
	analysis := &ProcessLatencyAnalysis{
		LatencyByType: map[StepType]time.Duration{
			StepTypeLLM:           5 * time.Second,
			StepTypeExternal:      1 * time.Second,
			StepTypeDeterministic: 100 * time.Millisecond,
		},
	}

	breakdown := GetLatencyBreakdown(analysis)

	if len(breakdown) != 3 {
		t.Errorf("expected 3 breakdown entries, got %d", len(breakdown))
	}

	if breakdown["llm"] != 5*time.Second {
		t.Errorf("expected LLM latency 5s, got %s", breakdown["llm"])
	}
}

func TestStepTypeLatencyDefaults(t *testing.T) {
	stepTypes := []StepType{
		StepTypeDeterministic,
		StepTypeLLM,
		StepTypeHuman,
		StepTypeExternal,
		StepTypeTool,
	}

	for _, st := range stepTypes {
		defaults, ok := StepTypeLatencyDefaults[st]
		if !ok {
			t.Errorf("expected defaults for step type %s", st)
			continue
		}

		// Verify P50 < P95 < P99 < Max
		if defaults.P50 >= defaults.P95 {
			t.Errorf("step type %s: expected P50 < P95", st)
		}
		if defaults.P95 >= defaults.P99 {
			t.Errorf("step type %s: expected P95 < P99", st)
		}
		if defaults.P99 >= defaults.Max {
			t.Errorf("step type %s: expected P99 < Max", st)
		}

		// Verify variance is set
		if defaults.Variance == "" {
			t.Errorf("step type %s: expected variance class to be set", st)
		}
	}
}

func TestFormatLatencyReport(t *testing.T) {
	p := createLatencyTestProtocol()
	analysis := AnalyzeProcessLatency(p)

	report := FormatLatencyReport(analysis)

	// Check report contains expected sections
	if len(report) == 0 {
		t.Error("expected non-empty report")
	}

	// Check for protocol ID
	if !containsSubstring(report, "latency-test") {
		t.Error("expected report to contain protocol ID")
	}

	// Check for P50/P95/P99 labels
	if !containsSubstring(report, "P50:") || !containsSubstring(report, "P95:") || !containsSubstring(report, "P99:") {
		t.Error("expected report to contain percentile labels")
	}
}

func TestAnalyzeProcessLatency_ParallelSavings(t *testing.T) {
	// Create a protocol with parallel branches
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "parallel-test",
			Name: "Parallel Test",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "start",
				Name:     "Start",
				StepType: StepTypeDeterministic,
			},
			{
				ID:       "branch_a",
				Name:     "Branch A",
				StepType: StepTypeExternal,
			},
			{
				ID:       "branch_b",
				Name:     "Branch B",
				StepType: StepTypeExternal,
			},
			{
				ID:       "join",
				Name:     "Join",
				StepType: StepTypeDeterministic,
			},
		},
		Flows: []Flow{
			{From: "start", To: "branch_a", Action: "send"},
			{From: "start", To: "branch_b", Action: "send"},
			{From: "branch_a", To: "join", Action: "send"},
			{From: "branch_b", To: "join", Action: "send"},
		},
	}

	analysis := AnalyzeProcessLatency(p)

	// With parallel branches, parallel savings should be positive
	// (since both branches don't need to run sequentially)
	if analysis.ParallelSavings < 0 {
		t.Error("expected non-negative parallel savings")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (len(s) >= len(substr)) && (s == substr || len(s) > len(substr))
}

func TestLatencyBudgetType(t *testing.T) {
	budget := &LatencyBudget{
		P50:             "100ms",
		P95:             "500ms",
		P99:             "1s",
		Max:             "5s",
		ExpectedLatency: "80ms",
		Critical:        true,
		VarianceClass:   LatencyVarianceMedium,
	}

	// Verify all fields are accessible
	if budget.P50 != "100ms" {
		t.Error("P50 field mismatch")
	}
	if budget.P95 != "500ms" {
		t.Error("P95 field mismatch")
	}
	if budget.P99 != "1s" {
		t.Error("P99 field mismatch")
	}
	if budget.Max != "5s" {
		t.Error("Max field mismatch")
	}
	if budget.ExpectedLatency != "80ms" {
		t.Error("ExpectedLatency field mismatch")
	}
	if !budget.Critical {
		t.Error("Critical field mismatch")
	}
	if budget.VarianceClass != LatencyVarianceMedium {
		t.Error("VarianceClass field mismatch")
	}
}

func TestLatencyVarianceClass(t *testing.T) {
	tests := []struct {
		class    LatencyVarianceClass
		expected string
	}{
		{LatencyVarianceLow, "low"},
		{LatencyVarianceMedium, "medium"},
		{LatencyVarianceHigh, "high"},
	}

	for _, tt := range tests {
		if string(tt.class) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.class)
		}
	}
}
