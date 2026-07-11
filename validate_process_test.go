package pidl

import (
	"strings"
	"testing"
)

func TestValidateProcess_NonProcessSpec(t *testing.T) {
	// Protocol specs should pass process validation (returns no errors)
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-protocol",
			Name: "Test Protocol",
			Kind: ProtocolKindProtocol,
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.ValidateProcess()
	if errs.HasErrors() {
		t.Errorf("expected no process validation errors for protocol spec, got: %v", errs)
	}
}

func TestValidateProcess_ValidProcessSpec(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-process",
			Name: "Test Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindFile, Name: "input.md", Required: true},
				},
				Outputs: []DataPort{
					{Kind: DataPortKindObject, Name: "parsed"},
				},
				Processing: &ProcessingConfig{
					Engine:      "markdown-parser",
					Determinism: DeterminismDeterministic,
				},
			},
			{
				ID:       "step2",
				Name:     "Step 2",
				Type:     EntityTypeServer,
				StepType: StepTypeLLM,
				Inputs: []DataPort{
					{Kind: DataPortKindObject, Name: "parsed", Required: true},
				},
				Outputs: []DataPort{
					{Kind: DataPortKindFile, Name: "output.md"},
				},
				Processing: &ProcessingConfig{
					Engine:      "llm",
					Determinism: DeterminismNonDeterministic,
					ModelPolicy: "quality-gated",
				},
			},
		},
		Flows: []Flow{
			{From: "step1", To: "step2", Action: "process"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("expected no validation errors, got: %v", errs)
	}
}

func TestValidateProcess_InvalidStepType(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-process",
			Name: "Test Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				Type:     EntityTypeServer,
				StepType: StepType("invalid_type"),
			},
			{
				ID:   "step2",
				Name: "Step 2",
				Type: EntityTypeServer,
			},
		},
		Flows: []Flow{
			{From: "step1", To: "step2", Action: "process"},
		},
	}

	errs := p.Validate()
	if !errs.HasErrors() {
		t.Error("expected validation error for invalid step type")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Field, "step_type") && strings.Contains(err.Message, "invalid step type") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about invalid step type, got: %v", errs)
	}
}

func TestValidateProcess_InvalidDataPortKind(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-process",
			Name: "Test Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKind("invalid_kind"), Name: "input"},
				},
			},
			{
				ID:   "step2",
				Name: "Step 2",
				Type: EntityTypeServer,
			},
		},
		Flows: []Flow{
			{From: "step1", To: "step2", Action: "process"},
		},
	}

	errs := p.Validate()
	if !errs.HasErrors() {
		t.Error("expected validation error for invalid data port kind")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Field, "inputs") && strings.Contains(err.Message, "invalid data port kind") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about invalid data port kind, got: %v", errs)
	}
}

func TestValidateProcess_MissingDataPortName(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-process",
			Name: "Test Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindFile, Name: ""}, // Missing name
				},
			},
			{
				ID:   "step2",
				Name: "Step 2",
				Type: EntityTypeServer,
			},
		},
		Flows: []Flow{
			{From: "step1", To: "step2", Action: "process"},
		},
	}

	errs := p.Validate()
	if !errs.HasErrors() {
		t.Error("expected validation error for missing data port name")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Field, "name") && err.Message == "required" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about required name, got: %v", errs)
	}
}

func TestValidateProcess_DuplicateInputNames(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-process",
			Name: "Test Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindFile, Name: "input"},
					{Kind: DataPortKindFile, Name: "input"}, // Duplicate
				},
			},
			{
				ID:   "step2",
				Name: "Step 2",
				Type: EntityTypeServer,
			},
		},
		Flows: []Flow{
			{From: "step1", To: "step2", Action: "process"},
		},
	}

	errs := p.Validate()
	if !errs.HasErrors() {
		t.Error("expected validation error for duplicate input names")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Message, "duplicate input name") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about duplicate input name, got: %v", errs)
	}
}

func TestValidateProcess_RetryOnUnknownFailureMode(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-process",
			Name: "Test Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				Type:     EntityTypeServer,
				StepType: StepTypeExternal,
				FailureModes: []FailureMode{
					{ID: "timeout", Name: "Timeout"},
				},
				RetryStrategy: &RetryStrategy{
					MaxAttempts: 3,
					RetryOn:     []string{"timeout", "unknown_error"}, // unknown_error doesn't exist
				},
			},
			{
				ID:   "step2",
				Name: "Step 2",
				Type: EntityTypeServer,
			},
		},
		Flows: []Flow{
			{From: "step1", To: "step2", Action: "process"},
		},
	}

	errs := p.Validate()
	if !errs.HasErrors() {
		t.Error("expected validation error for unknown failure mode in retry_on")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Field, "retry_on") && strings.Contains(err.Message, "unknown failure mode") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about unknown failure mode, got: %v", errs)
	}
}

func TestValidateProcess_DeterminismMismatch(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test-process",
			Name: "Test Process",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "step1",
				Name:     "Step 1",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Processing: &ProcessingConfig{
					Determinism: DeterminismNonDeterministic, // Contradicts step_type
				},
			},
			{
				ID:   "step2",
				Name: "Step 2",
				Type: EntityTypeServer,
			},
		},
		Flows: []Flow{
			{From: "step1", To: "step2", Action: "process"},
		},
	}

	errs := p.Validate()
	if !errs.HasErrors() {
		t.Error("expected validation error for determinism mismatch")
	}

	found := false
	for _, err := range errs {
		if strings.Contains(err.Field, "determinism") && strings.Contains(err.Message, "contradicts") {
			found = true
			break
		}
	}
	// This is a warning, so it should be present
	if !found {
		t.Logf("Warning about determinism mismatch expected but got: %v", errs)
	}
}

func TestValidateProcess_AllStepTypes(t *testing.T) {
	stepTypes := []StepType{
		StepTypeDeterministic,
		StepTypeLLM,
		StepTypeHuman,
		StepTypeExternal,
		StepTypeTool,
	}

	for _, st := range stepTypes {
		t.Run(string(st), func(t *testing.T) {
			p := &Protocol{
				ProtocolMeta: ProtocolMeta{
					ID:   "test-process",
					Name: "Test Process",
					Kind: ProtocolKindProcess,
				},
				Entities: []Entity{
					{
						ID:       "step1",
						Name:     "Step 1",
						Type:     EntityTypeServer,
						StepType: st,
					},
					{
						ID:   "step2",
						Name: "Step 2",
						Type: EntityTypeServer,
					},
				},
				Flows: []Flow{
					{From: "step1", To: "step2", Action: "process"},
				},
			}

			errs := p.ValidateProcess()
			// Should not have step type errors
			for _, err := range errs {
				if strings.Contains(err.Field, "step_type") {
					t.Errorf("unexpected step type error for %s: %v", st, err)
				}
			}
		})
	}
}

func TestValidateProcess_AllDataPortKinds(t *testing.T) {
	kinds := []DataPortKind{
		DataPortKindFile,
		DataPortKindObject,
		DataPortKindAPI,
		DataPortKindDatabase,
		DataPortKindQueue,
		DataPortKindStream,
	}

	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			p := &Protocol{
				ProtocolMeta: ProtocolMeta{
					ID:   "test-process",
					Name: "Test Process",
					Kind: ProtocolKindProcess,
				},
				Entities: []Entity{
					{
						ID:       "step1",
						Name:     "Step 1",
						Type:     EntityTypeServer,
						StepType: StepTypeDeterministic,
						Inputs: []DataPort{
							{Kind: kind, Name: "input"},
						},
						Outputs: []DataPort{
							{Kind: kind, Name: "output"},
						},
					},
					{
						ID:   "step2",
						Name: "Step 2",
						Type: EntityTypeServer,
					},
				},
				Flows: []Flow{
					{From: "step1", To: "step2", Action: "process"},
				},
			}

			errs := p.ValidateProcess()
			// Should not have data port kind errors
			for _, err := range errs {
				if strings.Contains(err.Message, "invalid data port kind") {
					t.Errorf("unexpected data port kind error for %s: %v", kind, err)
				}
			}
		})
	}
}

// Test helper methods
func TestEntity_IsProcessStep(t *testing.T) {
	tests := []struct {
		name     string
		entity   Entity
		expected bool
	}{
		{
			name:     "with step type",
			entity:   Entity{StepType: StepTypeLLM},
			expected: true,
		},
		{
			name:     "without step type",
			entity:   Entity{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entity.IsProcessStep(); got != tt.expected {
				t.Errorf("IsProcessStep() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEntity_StepTypeChecks(t *testing.T) {
	tests := []struct {
		stepType        StepType
		isDeterministic bool
		isLLM           bool
		isHuman         bool
		isExternal      bool
		isTool          bool
	}{
		{StepTypeDeterministic, true, false, false, false, false},
		{StepTypeLLM, false, true, false, false, false},
		{StepTypeHuman, false, false, true, false, false},
		{StepTypeExternal, false, false, false, true, false},
		{StepTypeTool, false, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.stepType), func(t *testing.T) {
			e := Entity{StepType: tt.stepType}

			if got := e.IsDeterministic(); got != tt.isDeterministic {
				t.Errorf("IsDeterministic() = %v, want %v", got, tt.isDeterministic)
			}
			if got := e.IsLLMStep(); got != tt.isLLM {
				t.Errorf("IsLLMStep() = %v, want %v", got, tt.isLLM)
			}
			if got := e.IsHumanStep(); got != tt.isHuman {
				t.Errorf("IsHumanStep() = %v, want %v", got, tt.isHuman)
			}
			if got := e.IsExternalStep(); got != tt.isExternal {
				t.Errorf("IsExternalStep() = %v, want %v", got, tt.isExternal)
			}
			if got := e.IsToolStep(); got != tt.isTool {
				t.Errorf("IsToolStep() = %v, want %v", got, tt.isTool)
			}
		})
	}
}

func TestEntity_IsNonDeterministic(t *testing.T) {
	tests := []struct {
		name     string
		entity   Entity
		expected bool
	}{
		{
			name: "LLM step is non-deterministic",
			entity: Entity{
				StepType: StepTypeLLM,
			},
			expected: true,
		},
		{
			name: "Human step is non-deterministic",
			entity: Entity{
				StepType: StepTypeHuman,
			},
			expected: true,
		},
		{
			name: "Deterministic step is deterministic",
			entity: Entity{
				StepType: StepTypeDeterministic,
			},
			expected: false,
		},
		{
			name: "Processing config overrides",
			entity: Entity{
				StepType: StepTypeDeterministic,
				Processing: &ProcessingConfig{
					Determinism: DeterminismNonDeterministic,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entity.IsNonDeterministic(); got != tt.expected {
				t.Errorf("IsNonDeterministic() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProtocol_IsProcessSpec(t *testing.T) {
	tests := []struct {
		name     string
		kind     ProtocolKind
		expected bool
	}{
		{"process spec", ProtocolKindProcess, true},
		{"protocol spec", ProtocolKindProtocol, false},
		{"empty kind", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Protocol{
				ProtocolMeta: ProtocolMeta{
					ID:   "test",
					Name: "Test",
					Kind: tt.kind,
				},
			}
			if got := p.IsProcessSpec(); got != tt.expected {
				t.Errorf("IsProcessSpec() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProtocol_Kind(t *testing.T) {
	tests := []struct {
		name     string
		kind     ProtocolKind
		expected ProtocolKind
	}{
		{"process spec", ProtocolKindProcess, ProtocolKindProcess},
		{"protocol spec", ProtocolKindProtocol, ProtocolKindProtocol},
		{"empty defaults to protocol", "", ProtocolKindProtocol},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Protocol{
				ProtocolMeta: ProtocolMeta{
					ID:   "test",
					Name: "Test",
					Kind: tt.kind,
				},
			}
			if got := p.Kind(); got != tt.expected {
				t.Errorf("Kind() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProtocol_ProcessSteps(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{ID: "step1", StepType: StepTypeDeterministic},
			{ID: "step2", StepType: StepTypeLLM},
			{ID: "step3"}, // Not a process step
		},
	}

	steps := p.ProcessSteps()
	if len(steps) != 2 {
		t.Errorf("ProcessSteps() returned %d steps, want 2", len(steps))
	}
}

func TestProtocol_LLMSteps(t *testing.T) {
	p := &Protocol{
		Entities: []Entity{
			{ID: "step1", StepType: StepTypeDeterministic},
			{ID: "step2", StepType: StepTypeLLM},
			{ID: "step3", StepType: StepTypeLLM},
		},
	}

	steps := p.LLMSteps()
	if len(steps) != 2 {
		t.Errorf("LLMSteps() returned %d steps, want 2", len(steps))
	}
}
