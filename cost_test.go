package pidl

import (
	"reflect"
	"testing"
	"time"
)

func createCostTestProtocol() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "cost-test",
			Name: "Cost Test Pipeline",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "extract",
				Name:     "Extract",
				StepType: StepTypeExternal,
				Processing: &ProcessingConfig{
					Timeout: "30s",
					CostModel: &CostModel{
						Type:        CostModelTypeAPICall,
						APICallCost: 0.01,
						CostUnit:    "USD",
					},
				},
			},
			{
				ID:       "transform",
				Name:     "Transform",
				StepType: StepTypeLLM,
				Processing: &ProcessingConfig{
					CostModel: &CostModel{
						Type:                  CostModelTypeTokenBased,
						InputTokenCost:        0.003,
						OutputTokenCost:       0.015,
						EstimatedInputTokens:  1000,
						EstimatedOutputTokens: 500,
						CostUnit:              "USD",
					},
				},
			},
			{
				ID:       "load",
				Name:     "Load",
				StepType: StepTypeDeterministic,
				Processing: &ProcessingConfig{
					CostModel: &CostModel{
						Type:      CostModelTypeFixed,
						FixedCost: 0.001,
						CostUnit:  "USD",
					},
				},
			},
		},
		Flows: []Flow{
			{From: "extract", To: "transform", Action: "send"},
			{From: "transform", To: "load", Action: "send"},
		},
	}
}

func TestAnalyzeProcessCosts(t *testing.T) {
	p := createCostTestProtocol()
	analysis := AnalyzeProcessCosts(p)

	if analysis.ProtocolID != "cost-test" {
		t.Errorf("expected protocol ID 'cost-test', got '%s'", analysis.ProtocolID)
	}

	// Should have 3 step estimates
	if len(analysis.StepEstimates) != 3 {
		t.Errorf("expected 3 step estimates, got %d", len(analysis.StepEstimates))
	}

	// Total cost should be positive
	if analysis.TotalEstimate.ExpectedCost <= 0 {
		t.Error("expected positive total cost")
	}

	// Check cost unit
	if analysis.TotalEstimate.CostUnit != "USD" {
		t.Errorf("expected cost unit 'USD', got '%s'", analysis.TotalEstimate.CostUnit)
	}
}

func TestAnalyzeProcessCosts_ByType(t *testing.T) {
	p := createCostTestProtocol()
	analysis := AnalyzeProcessCosts(p)

	// Should have costs for different step types
	if _, ok := analysis.CostByType[StepTypeLLM]; !ok {
		t.Error("expected LLM cost in breakdown")
	}

	if _, ok := analysis.CostByType[StepTypeExternal]; !ok {
		t.Error("expected External cost in breakdown")
	}

	if _, ok := analysis.CostByType[StepTypeDeterministic]; !ok {
		t.Error("expected Deterministic cost in breakdown")
	}
}

func TestAnalyzeProcessCosts_StepDetails(t *testing.T) {
	p := createCostTestProtocol()
	analysis := AnalyzeProcessCosts(p)

	// Find LLM step estimate
	var llmEstimate *CostEstimate
	for i, est := range analysis.StepEstimates {
		if est.EntityID == "transform" {
			llmEstimate = &analysis.StepEstimates[i]
			break
		}
	}

	if llmEstimate == nil {
		t.Fatal("expected to find transform step estimate")
	}

	// LLM cost: 1000 input tokens * 0.003/1K + 500 output tokens * 0.015/1K
	// = 0.003 + 0.0075 = 0.0105
	expectedCost := 0.0105
	if llmEstimate.ExpectedCost < expectedCost*0.9 || llmEstimate.ExpectedCost > expectedCost*1.1 {
		t.Errorf("expected LLM cost around %f, got %f", expectedCost, llmEstimate.ExpectedCost)
	}
}

func TestCalculateExecutionCost_TokenBased(t *testing.T) {
	entity := &Entity{
		Processing: &ProcessingConfig{
			CostModel: &CostModel{
				Type:            CostModelTypeTokenBased,
				InputTokenCost:  0.003, // per 1K
				OutputTokenCost: 0.015, // per 1K
			},
		},
	}

	metrics := ExecutionMetrics{
		InputTokens:  2000, // 2K tokens
		OutputTokens: 1000, // 1K tokens
	}

	cost := CalculateExecutionCost(entity, metrics)

	// Expected: 2 * 0.003 + 1 * 0.015 = 0.006 + 0.015 = 0.021
	expected := 0.021
	if cost < expected*0.99 || cost > expected*1.01 {
		t.Errorf("expected cost %f, got %f", expected, cost)
	}
}

func TestCalculateExecutionCost_TimeBased(t *testing.T) {
	entity := &Entity{
		Processing: &ProcessingConfig{
			CostModel: &CostModel{
				Type:                 CostModelTypeTimeBased,
				ComputeCostPerSecond: 0.001, // $0.001 per second
			},
		},
	}

	metrics := ExecutionMetrics{
		Duration: 30 * time.Second,
	}

	cost := CalculateExecutionCost(entity, metrics)

	// Expected: 30 * 0.001 = 0.03
	expected := 0.03
	if cost < expected*0.99 || cost > expected*1.01 {
		t.Errorf("expected cost %f, got %f", expected, cost)
	}
}

func TestCalculateExecutionCost_Fixed(t *testing.T) {
	entity := &Entity{
		Processing: &ProcessingConfig{
			CostModel: &CostModel{
				Type:      CostModelTypeFixed,
				FixedCost: 0.05,
			},
		},
	}

	metrics := ExecutionMetrics{} // Metrics don't matter for fixed cost

	cost := CalculateExecutionCost(entity, metrics)

	if cost != 0.05 {
		t.Errorf("expected cost 0.05, got %f", cost)
	}
}

func TestCalculateExecutionCost_APICall(t *testing.T) {
	entity := &Entity{
		Processing: &ProcessingConfig{
			CostModel: &CostModel{
				Type:        CostModelTypeAPICall,
				APICallCost: 0.001,
			},
		},
	}

	metrics := ExecutionMetrics{
		APICallCount: 5,
	}

	cost := CalculateExecutionCost(entity, metrics)

	// Expected: 5 * 0.001 = 0.005
	if cost != 0.005 {
		t.Errorf("expected cost 0.005, got %f", cost)
	}
}

func TestCalculateExecutionCost_NoCostModel(t *testing.T) {
	entity := &Entity{
		Processing: nil,
	}

	metrics := ExecutionMetrics{}
	cost := CalculateExecutionCost(entity, metrics)

	if cost != 0.0 {
		t.Errorf("expected cost 0.0 for entity without cost model, got %f", cost)
	}
}

func TestEstimateDefaultCost_LLM(t *testing.T) {
	entity := &Entity{
		ID:       "llm-step",
		StepType: StepTypeLLM,
	}

	estimate := estimateDefaultCost(entity)

	if estimate.ExpectedCost <= 0 {
		t.Error("expected positive default cost for LLM step")
	}

	if estimate.MinCost >= estimate.ExpectedCost {
		t.Error("expected MinCost < ExpectedCost")
	}

	if estimate.MaxCost <= estimate.ExpectedCost {
		t.Error("expected MaxCost > ExpectedCost")
	}
}

func TestGetCostEfficiency(t *testing.T) {
	analysis := &ProcessCostAnalysis{
		CostByType: map[StepType]float64{
			StepTypeLLM:           0.01,
			StepTypeExternal:      0.001,
			StepTypeDeterministic: 0.0001,
		},
	}

	efficiency := GetCostEfficiency(analysis)

	if len(efficiency) != 3 {
		t.Errorf("expected 3 efficiency entries, got %d", len(efficiency))
	}

	if efficiency["llm"] != 0.01 {
		t.Errorf("expected LLM efficiency 0.01, got %f", efficiency["llm"])
	}
}

func TestLLMCostPresets(t *testing.T) {
	// Verify presets exist and have reasonable values
	presets := []string{"gpt-4", "gpt-3.5-turbo", "claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}

	for _, name := range presets {
		preset, ok := LLMCostPresets[name]
		if !ok {
			t.Errorf("expected preset for %s", name)
			continue
		}

		if preset.InputCost <= 0 {
			t.Errorf("expected positive input cost for %s", name)
		}

		if preset.OutputCost <= 0 {
			t.Errorf("expected positive output cost for %s", name)
		}
	}
}

func TestEstimateDefaultCost(t *testing.T) {
	llmPreset := LLMCostPresets["claude-3-sonnet"]
	llmExpected := 1.0*llmPreset.InputCost + 0.5*llmPreset.OutputCost

	tests := []struct {
		name         string
		entity       *Entity
		wantExpected float64
		wantMin      float64
		wantMax      float64
	}{
		{name: "llm", entity: &Entity{ID: "llm", StepType: StepTypeLLM}, wantExpected: llmExpected, wantMin: llmExpected * 0.2, wantMax: llmExpected * 5.0},
		{name: "external", entity: &Entity{ID: "ext", StepType: StepTypeExternal}, wantExpected: 0.001, wantMin: 0, wantMax: 0.01},
		{name: "human", entity: &Entity{ID: "hum", StepType: StepTypeHuman}, wantExpected: 0, wantMin: 0, wantMax: 0},
		{name: "tool", entity: &Entity{ID: "tool", StepType: StepTypeTool}, wantExpected: 0.0001, wantMin: 0, wantMax: 0.001},
		{name: "deterministic", entity: &Entity{ID: "det", StepType: StepTypeDeterministic}, wantExpected: 0.00001, wantMin: 0, wantMax: 0.0001},
		{name: "no step type falls back to deterministic", entity: &Entity{ID: "none"}, wantExpected: 0.00001, wantMin: 0, wantMax: 0.0001},
		{name: "cost model without values still uses defaults path", entity: &Entity{ID: "proc", StepType: StepTypeTool, Processing: &ProcessingConfig{}}, wantExpected: 0.0001, wantMin: 0, wantMax: 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// estimateEntityCost routes to estimateDefaultCost whenever no
			// CostModel is configured; assert both paths agree.
			direct := estimateDefaultCost(tt.entity)
			viaEntity := estimateEntityCost(tt.entity)
			if !reflect.DeepEqual(direct, viaEntity) {
				t.Errorf("estimateEntityCost() = %+v, estimateDefaultCost() = %+v; want equal", viaEntity, direct)
			}

			if direct.EntityID != tt.entity.ID {
				t.Errorf("EntityID = %q, want %q", direct.EntityID, tt.entity.ID)
			}
			if direct.CostUnit != "USD" {
				t.Errorf("CostUnit = %q, want USD", direct.CostUnit)
			}
			if direct.ExpectedCost != tt.wantExpected {
				t.Errorf("ExpectedCost = %v, want %v", direct.ExpectedCost, tt.wantExpected)
			}
			if direct.MinCost != tt.wantMin {
				t.Errorf("MinCost = %v, want %v", direct.MinCost, tt.wantMin)
			}
			if direct.MaxCost != tt.wantMax {
				t.Errorf("MaxCost = %v, want %v", direct.MaxCost, tt.wantMax)
			}
		})
	}
}
