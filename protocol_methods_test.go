package pidl

import (
	"reflect"
	"testing"
)

func TestProtocol_DeterministicSteps(t *testing.T) {
	tests := []struct {
		name     string
		entities []Entity
		wantIDs  []string
	}{
		{name: "no entities"},
		{
			name:     "no deterministic steps",
			entities: []Entity{{ID: "a", StepType: StepTypeLLM}, {ID: "b"}},
		},
		{
			name: "mixed step types",
			entities: []Entity{
				{ID: "extract", StepType: StepTypeDeterministic},
				{ID: "classify", StepType: StepTypeLLM},
				{ID: "load", StepType: StepTypeDeterministic},
				{ID: "review", StepType: StepTypeHuman},
			},
			wantIDs: []string{"extract", "load"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Protocol{Entities: tt.entities}
			steps := p.DeterministicSteps()
			var got []string
			for _, s := range steps {
				got = append(got, s.ID)
			}
			if !reflect.DeepEqual(got, tt.wantIDs) {
				t.Errorf("DeterministicSteps() ids = %v, want %v", got, tt.wantIDs)
			}
		})
	}
}

func TestProtocol_PhaseIDs(t *testing.T) {
	tests := []struct {
		name   string
		phases []Phase
		want   []string
	}{
		{name: "no phases", want: []string{}},
		{name: "single phase", phases: []Phase{{ID: "setup"}}, want: []string{"setup"}},
		{
			name:   "ordered phases",
			phases: []Phase{{ID: "setup"}, {ID: "auth", Parent: "setup"}, {ID: "teardown"}},
			want:   []string{"setup", "auth", "teardown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Protocol{Phases: tt.phases}
			if got := p.PhaseIDs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PhaseIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
