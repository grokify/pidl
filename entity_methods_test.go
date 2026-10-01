package pidl

import (
	"reflect"
	"testing"
)

// processStepEntity returns an entity with a mix of required/sensitive ports
// and failure modes, used by the Entity accessor tests.
func processStepEntity() Entity {
	return Entity{
		ID:       "step",
		Name:     "Step",
		Type:     EntityTypeServer,
		StepType: StepTypeDeterministic,
		Inputs: []DataPort{
			{Kind: DataPortKindObject, Name: "in_required", Required: true},
			{Kind: DataPortKindObject, Name: "in_sensitive", Sensitive: true},
			{Kind: DataPortKindObject, Name: "in_plain"},
		},
		Outputs: []DataPort{
			{Kind: DataPortKindObject, Name: "out_sensitive", Sensitive: true},
			{Kind: DataPortKindObject, Name: "out_plain"},
		},
		FailureModes: []FailureMode{
			{ID: "timeout", Name: "Timeout"},
			{ID: "bad_input", Name: "Bad Input"},
		},
	}
}

func TestEntity_PortPredicates(t *testing.T) {
	tests := []struct {
		name       string
		entity     Entity
		hasInputs  bool
		hasOutputs bool
	}{
		{name: "no ports", entity: Entity{ID: "bare"}},
		{name: "inputs only", entity: Entity{Inputs: []DataPort{{Name: "a"}}}, hasInputs: true},
		{name: "outputs only", entity: Entity{Outputs: []DataPort{{Name: "b"}}}, hasOutputs: true},
		{name: "both", entity: processStepEntity(), hasInputs: true, hasOutputs: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entity.HasInputs(); got != tt.hasInputs {
				t.Errorf("HasInputs() = %v, want %v", got, tt.hasInputs)
			}
			if got := tt.entity.HasOutputs(); got != tt.hasOutputs {
				t.Errorf("HasOutputs() = %v, want %v", got, tt.hasOutputs)
			}
		})
	}
}

func TestEntity_PortFilters(t *testing.T) {
	tests := []struct {
		name             string
		entity           Entity
		requiredInputs   []string
		sensitiveInputs  []string
		sensitiveOutputs []string
	}{
		{name: "no ports", entity: Entity{ID: "bare"}},
		{
			name:             "mixed ports",
			entity:           processStepEntity(),
			requiredInputs:   []string{"in_required"},
			sensitiveInputs:  []string{"in_sensitive"},
			sensitiveOutputs: []string{"out_sensitive"},
		},
		{
			name: "all flags set",
			entity: Entity{
				Inputs:  []DataPort{{Name: "x", Required: true, Sensitive: true}, {Name: "y", Required: true, Sensitive: true}},
				Outputs: []DataPort{{Name: "z", Sensitive: true}},
			},
			requiredInputs:   []string{"x", "y"},
			sensitiveInputs:  []string{"x", "y"},
			sensitiveOutputs: []string{"z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := portNames(tt.entity.RequiredInputs()); !reflect.DeepEqual(got, tt.requiredInputs) {
				t.Errorf("RequiredInputs() = %v, want %v", got, tt.requiredInputs)
			}
			if got := portNames(tt.entity.SensitiveInputs()); !reflect.DeepEqual(got, tt.sensitiveInputs) {
				t.Errorf("SensitiveInputs() = %v, want %v", got, tt.sensitiveInputs)
			}
			if got := portNames(tt.entity.SensitiveOutputs()); !reflect.DeepEqual(got, tt.sensitiveOutputs) {
				t.Errorf("SensitiveOutputs() = %v, want %v", got, tt.sensitiveOutputs)
			}
		})
	}
}

// portNames extracts port names; nil for an empty slice so DeepEqual works
// against the nil expectation in the table.
func portNames(ports []DataPort) []string {
	if len(ports) == 0 {
		return nil
	}
	names := make([]string, len(ports))
	for i, p := range ports {
		names[i] = p.Name
	}
	return names
}

func TestEntity_LookupByName(t *testing.T) {
	entity := processStepEntity()

	// Each lookup is normalised to (identifier, found) so one runner covers
	// InputByName, OutputByName and FailureModeByID.
	byInput := func(k string) (string, bool) {
		if p := entity.InputByName(k); p != nil {
			return p.Name, true
		}
		return "", false
	}
	byOutput := func(k string) (string, bool) {
		if p := entity.OutputByName(k); p != nil {
			return p.Name, true
		}
		return "", false
	}
	byFailureMode := func(k string) (string, bool) {
		if fm := entity.FailureModeByID(k); fm != nil {
			return fm.ID, true
		}
		return "", false
	}

	tests := []struct {
		name    string
		lookup  func(string) (string, bool)
		key     string
		wantHit bool
	}{
		{name: "input found", lookup: byInput, key: "in_sensitive", wantHit: true},
		{name: "input missing", lookup: byInput, key: "nope"},
		{name: "input does not match outputs", lookup: byInput, key: "out_plain"},
		{name: "output found", lookup: byOutput, key: "out_plain", wantHit: true},
		{name: "output missing", lookup: byOutput, key: "in_plain"},
		{name: "failure mode found", lookup: byFailureMode, key: "bad_input", wantHit: true},
		{name: "failure mode missing", lookup: byFailureMode, key: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, hit := tt.lookup(tt.key)
			if hit != tt.wantHit {
				t.Fatalf("lookup(%q) hit = %v, want %v", tt.key, hit, tt.wantHit)
			}
			if hit && got != tt.key {
				t.Errorf("lookup(%q) returned %q", tt.key, got)
			}
		})
	}
}

func TestEntity_LookupReturnsPointerIntoEntity(t *testing.T) {
	entity := processStepEntity()

	if p := entity.InputByName("in_plain"); p != &entity.Inputs[2] {
		t.Error("InputByName should return a pointer into the entity's Inputs slice")
	}
	if p := entity.OutputByName("out_plain"); p != &entity.Outputs[1] {
		t.Error("OutputByName should return a pointer into the entity's Outputs slice")
	}
	if fm := entity.FailureModeByID("timeout"); fm != &entity.FailureModes[0] {
		t.Error("FailureModeByID should return a pointer into the entity's FailureModes slice")
	}
}

func TestEntity_StepTypePredicates(t *testing.T) {
	tests := []struct {
		name          string
		entity        Entity
		isProcessStep bool
		isDet         bool
		isLLM         bool
		isHuman       bool
		isExternal    bool
		isTool        bool
		isNonDet      bool
	}{
		{name: "no step type", entity: Entity{ID: "plain"}},
		{name: "deterministic", entity: Entity{StepType: StepTypeDeterministic}, isProcessStep: true, isDet: true},
		{name: "llm", entity: Entity{StepType: StepTypeLLM}, isProcessStep: true, isLLM: true, isNonDet: true},
		{name: "human", entity: Entity{StepType: StepTypeHuman}, isProcessStep: true, isHuman: true, isNonDet: true},
		{name: "external", entity: Entity{StepType: StepTypeExternal}, isProcessStep: true, isExternal: true},
		{name: "tool", entity: Entity{StepType: StepTypeTool}, isProcessStep: true, isTool: true},
		{
			name: "deterministic step marked non-deterministic",
			entity: Entity{
				StepType:   StepTypeDeterministic,
				Processing: &ProcessingConfig{Determinism: DeterminismNonDeterministic},
			},
			isProcessStep: true,
			isDet:         true,
			isNonDet:      true,
		},
		{
			name: "tool step marked deterministic",
			entity: Entity{
				StepType:   StepTypeTool,
				Processing: &ProcessingConfig{Determinism: DeterminismDeterministic},
			},
			isProcessStep: true,
			isTool:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checks := []struct {
				name string
				got  bool
				want bool
			}{
				{"IsProcessStep", tt.entity.IsProcessStep(), tt.isProcessStep},
				{"IsDeterministic", tt.entity.IsDeterministic(), tt.isDet},
				{"IsLLMStep", tt.entity.IsLLMStep(), tt.isLLM},
				{"IsHumanStep", tt.entity.IsHumanStep(), tt.isHuman},
				{"IsExternalStep", tt.entity.IsExternalStep(), tt.isExternal},
				{"IsToolStep", tt.entity.IsToolStep(), tt.isTool},
				{"IsNonDeterministic", tt.entity.IsNonDeterministic(), tt.isNonDet},
			}
			for _, c := range checks {
				if c.got != c.want {
					t.Errorf("%s() = %v, want %v", c.name, c.got, c.want)
				}
			}
		})
	}
}
