package pidl

import (
	"strings"
	"testing"
)

func TestValidateValidProtocol(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:       "test-protocol",
			Name:     "Test Protocol",
			Category: CategoryAuth,
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "auth", Name: "Authorization"},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request", Phase: "auth"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors", errs)
	}

	if !p.IsValid() {
		t.Error("IsValid() = false, want true")
	}
}

func TestValidateMissingProtocolID(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x"},
		},
	}

	errs := p.Validate()
	if !errs.HasErrors() {
		t.Error("Validate() should error on missing protocol.id")
	}

	found := false
	for _, e := range errs {
		if e.Field == "protocol.id" && strings.Contains(e.Message, "required") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should have protocol.id required error")
	}
}

func TestValidateInvalidProtocolID(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "Invalid-ID", // uppercase not allowed
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x"},
		},
	}

	errs := p.Validate()
	if !errs.HasErrors() {
		t.Error("Validate() should error on invalid protocol.id")
	}
}

func TestValidateTooFewEntities(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
		},
		Flows: []Flow{
			{From: "a", To: "a", Action: "x"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "entities" && strings.Contains(e.Message, "at least 2") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should require at least 2 entities")
	}
}

func TestValidateDuplicateEntityID(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client 1", Type: EntityTypeClient},
			{ID: "client", Name: "Client 2", Type: EntityTypeClient}, // duplicate
		},
		Flows: []Flow{
			{From: "client", To: "client", Action: "x"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "duplicate") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect duplicate entity IDs")
	}
}

func TestValidateInvalidEntityType(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: "invalid_type"},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid entity type") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect invalid entity type")
	}
}

func TestValidateUnknownEntityInFlow(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "unknown", Action: "x"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown entity") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown entity in flow")
	}
}

func TestValidateUnknownPhaseInFlow(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x", Phase: "unknown_phase"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown phase") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown phase in flow")
	}
}

func TestValidateNoFlows(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "flows" && strings.Contains(e.Message, "at least 1") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should require at least 1 flow")
	}
}

func TestValidateInvalidFlowMode(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x", Mode: "invalid_mode"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid flow mode") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect invalid flow mode")
	}
}

func TestValidationErrorsString(t *testing.T) {
	errs := ValidationErrors{
		{Field: "field1", Message: "error1"},
		{Field: "field2", Message: "error2"},
	}

	s := errs.Error()
	if !strings.Contains(s, "2 validation errors") {
		t.Errorf("Error() = %q, should contain count", s)
	}
	if !strings.Contains(s, "field1: error1") {
		t.Errorf("Error() = %q, should contain first error", s)
	}
}

func TestValidationErrorsSingle(t *testing.T) {
	errs := ValidationErrors{
		{Field: "field", Message: "message"},
	}

	s := errs.Error()
	if s != "field: message" {
		t.Errorf("Error() = %q, want %q", s, "field: message")
	}
}

func TestValidationErrorsEmpty(t *testing.T) {
	errs := ValidationErrors{}
	if errs.Error() != "" {
		t.Errorf("Error() = %q, want empty", errs.Error())
	}
	if errs.HasErrors() {
		t.Error("HasErrors() = true, want false")
	}
}

func TestValidateNestedPhases(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "parent", Name: "Parent Phase"},
			{ID: "child", Name: "Child Phase", Parent: "parent"},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x", Phase: "child"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid nested phases", errs)
	}
}

func TestValidateInvalidPhaseParent(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "child", Name: "Child Phase", Parent: "nonexistent"},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown parent phase") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown parent phase")
	}
}

func TestValidateSelfParentPhase(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "self", Name: "Self Parent", Parent: "self"},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "cannot be its own parent") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect self-referential parent")
	}
}

func TestValidateCircularPhaseHierarchy(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "phase1", Name: "Phase 1", Parent: "phase2"},
			{ID: "phase2", Name: "Phase 2", Parent: "phase1"},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "circular reference") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect circular phase hierarchy")
	}
}

func TestValidateInvalidAnnotationType(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "a",
				To:     "b",
				Action: "x",
				Annotations: []Annotation{
					{Type: "invalid_type", Text: "test"},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid annotation type") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect invalid annotation type")
	}
}

func TestValidateMissingAnnotationText(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "a",
				To:     "b",
				Action: "x",
				Annotations: []Annotation{
					{Type: AnnotationTypeSecurity, Text: ""},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, "annotations") && strings.Contains(e.Message, "required") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect missing annotation text")
	}
}

func TestValidateAlternativeFlows(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "a",
				To:     "b",
				Action: "x",
				Alternatives: []Alternative{
					{
						Condition: "error",
						Flows: []Flow{
							{From: "b", To: "a", Action: "error_response"},
						},
					},
				},
			},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid alternatives", errs)
	}
}

func TestValidateAlternativeMissingCondition(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "a",
				To:     "b",
				Action: "x",
				Alternatives: []Alternative{
					{
						Condition: "", // missing condition
						Flows: []Flow{
							{From: "b", To: "a", Action: "y"},
						},
					},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, "alternatives") && strings.Contains(e.Message, "required") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect missing alternative condition")
	}
}

func TestValidateAlternativeEmptyFlows(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "a",
				To:     "b",
				Action: "x",
				Alternatives: []Alternative{
					{
						Condition: "error",
						Flows:     []Flow{}, // empty flows
					},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, "flows") && strings.Contains(e.Message, "at least 1") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect empty alternative flows")
	}
}

func TestValidateAlternativeUnknownEntity(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "a",
				To:     "b",
				Action: "x",
				Alternatives: []Alternative{
					{
						Condition: "error",
						Flows: []Flow{
							{From: "unknown", To: "a", Action: "y"},
						},
					},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown entity") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown entity in alternative flows")
	}
}

func TestFlowHelpers(t *testing.T) {
	f := Flow{
		From:      "a",
		To:        "b",
		Action:    "test",
		Condition: "when_valid",
		Note:      "This is a note",
		Annotations: []Annotation{
			{Type: AnnotationTypeSecurity, Text: "check auth"},
		},
		Alternatives: []Alternative{
			{Condition: "error", Flows: []Flow{{From: "b", To: "a", Action: "err"}}},
		},
	}

	if !f.HasCondition() {
		t.Error("HasCondition() = false, want true")
	}
	if !f.HasNote() {
		t.Error("HasNote() = false, want true")
	}
	if !f.HasAnnotations() {
		t.Error("HasAnnotations() = false, want true")
	}
	if !f.HasAlternatives() {
		t.Error("HasAlternatives() = false, want true")
	}

	// Test empty flow
	empty := Flow{}
	if empty.HasCondition() {
		t.Error("HasCondition() = true for empty, want false")
	}
	if empty.HasNote() {
		t.Error("HasNote() = true for empty, want false")
	}
	if empty.HasAnnotations() {
		t.Error("HasAnnotations() = true for empty, want false")
	}
	if empty.HasAlternatives() {
		t.Error("HasAlternatives() = true for empty, want false")
	}
}

func TestIsValidAnnotationType(t *testing.T) {
	validTypes := []AnnotationType{
		AnnotationTypeSecurity,
		AnnotationTypePerformance,
		AnnotationTypeDeprecated,
		AnnotationTypeInfo,
		AnnotationTypeWarning,
		AnnotationTypeError,
	}

	for _, at := range validTypes {
		if !IsValidAnnotationType(at) {
			t.Errorf("IsValidAnnotationType(%q) = false, want true", at)
		}
	}

	if IsValidAnnotationType("invalid") {
		t.Error("IsValidAnnotationType(invalid) = true, want false")
	}
}

func TestValidateEntityStates(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle", Name: "Idle", Initial: true},
					{ID: "active", Name: "Active"},
					{ID: "error", Final: true},
				},
			},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid entity states", errs)
	}
}

func TestValidateEntityStatesDuplicateID(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle"},
					{ID: "idle"}, // duplicate
				},
			},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "duplicate state ID") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect duplicate state IDs")
	}
}

func TestValidateEntityStatesInvalidID(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "Invalid-State"}, // invalid pattern
				},
			},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "must match pattern") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect invalid state ID pattern")
	}
}

func TestValidateEntityStatesMultipleInitial(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle", Initial: true},
					{ID: "ready", Initial: true}, // second initial
				},
			},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "initial states") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect multiple initial states")
	}
}

func TestValidateStateMutationsValid(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle", Initial: true},
					{ID: "active"},
				},
			},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "login",
				Sets: []StateMutation{
					{Entity: "client", From: "idle", To: "active"},
				},
			},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid state mutations", errs)
	}
}

func TestValidateStateMutationsUnknownEntity(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Sets: []StateMutation{
					{Entity: "unknown", To: "active"},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown entity") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown entity in state mutation")
	}
}

func TestValidateStateMutationsEntityNoStates(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient}, // no states
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Sets: []StateMutation{
					{Entity: "client", To: "active"},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "has no states defined") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect entity with no states in mutation")
	}
}

func TestValidateStateMutationsUnknownToState(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle"},
				},
			},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Sets: []StateMutation{
					{Entity: "client", To: "unknown_state"},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, ".to") && strings.Contains(e.Message, "unknown state") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown 'to' state")
	}
}

func TestValidateStateMutationsUnknownFromState(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "client",
				Name: "Client",
				Type: EntityTypeClient,
				States: []EntityState{
					{ID: "idle"},
					{ID: "active"},
				},
			},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Sets: []StateMutation{
					{Entity: "client", From: "unknown_state", To: "active"},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, ".from") && strings.Contains(e.Message, "unknown state") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown 'from' state")
	}
}

func TestPhaseHierarchyHelpers(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "test", Name: "Test"},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "root1", Name: "Root 1"},
			{ID: "root2", Name: "Root 2"},
			{ID: "child1", Name: "Child 1", Parent: "root1"},
			{ID: "grandchild", Name: "Grandchild", Parent: "child1"},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x"},
		},
	}

	// Test RootPhases
	roots := p.RootPhases()
	if len(roots) != 2 {
		t.Errorf("RootPhases() = %d phases, want 2", len(roots))
	}

	// Test ChildPhases
	children := p.ChildPhases("root1")
	if len(children) != 1 {
		t.Errorf("ChildPhases(root1) = %d phases, want 1", len(children))
	}
	if children[0].ID != "child1" {
		t.Errorf("ChildPhases(root1)[0].ID = %q, want %q", children[0].ID, "child1")
	}

	// Test PhaseDepth
	if depth := p.PhaseDepth("root1"); depth != 0 {
		t.Errorf("PhaseDepth(root1) = %d, want 0", depth)
	}
	if depth := p.PhaseDepth("child1"); depth != 1 {
		t.Errorf("PhaseDepth(child1) = %d, want 1", depth)
	}
	if depth := p.PhaseDepth("grandchild"); depth != 2 {
		t.Errorf("PhaseDepth(grandchild) = %d, want 2", depth)
	}
}

func TestValidateTrustLevel(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient, TrustLevel: TrustLevelUntrusted},
			{ID: "server", Name: "Server", Type: EntityTypeServer, TrustLevel: TrustLevelTrusted},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid trust levels", errs)
	}
}

func TestValidateInvalidTrustLevel(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient, TrustLevel: "invalid_level"},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid trust level") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect invalid trust level")
	}
}

func TestValidateTokenDefinitions(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "auth_server", Name: "Auth Server", Type: EntityTypeAuthorizationServer},
			{ID: "resource_server", Name: "Resource Server", Type: EntityTypeResourceServer},
		},
		Metadata: &ProtocolMetadata{
			Tokens: []TokenDefinition{
				{
					ID:       "access_token",
					Name:     "Access Token",
					Type:     "jwt",
					Issuer:   "auth_server",
					Audience: "resource_server",
				},
			},
		},
		Flows: []Flow{
			{From: "auth_server", To: "resource_server", Action: "issue_token"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid token definitions", errs)
	}
}

func TestValidateTokenDefinitionsDuplicateID(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "auth_server", Name: "Auth Server", Type: EntityTypeAuthorizationServer},
			{ID: "resource_server", Name: "Resource Server", Type: EntityTypeResourceServer},
		},
		Metadata: &ProtocolMetadata{
			Tokens: []TokenDefinition{
				{ID: "access_token", Name: "Access Token 1"},
				{ID: "access_token", Name: "Access Token 2"}, // duplicate
			},
		},
		Flows: []Flow{
			{From: "auth_server", To: "resource_server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "duplicate token ID") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect duplicate token ID")
	}
}

func TestValidateTokenDefinitionsInvalidIssuer(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "auth_server", Name: "Auth Server", Type: EntityTypeAuthorizationServer},
			{ID: "resource_server", Name: "Resource Server", Type: EntityTypeResourceServer},
		},
		Metadata: &ProtocolMetadata{
			Tokens: []TokenDefinition{
				{ID: "access_token", Issuer: "unknown_entity"},
			},
		},
		Flows: []Flow{
			{From: "auth_server", To: "resource_server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, ".issuer") && strings.Contains(e.Message, "unknown entity") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown issuer entity")
	}
}

func TestValidateFlowSecurity(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			Tokens: []TokenDefinition{
				{ID: "access_token"},
			},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Security: &FlowSecurity{
					Requires:     []SecurityRequirement{SecurityRequirementToken, SecurityRequirementEncryption},
					Token:        "access_token",
					Confidential: true,
				},
			},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid flow security", errs)
	}
}

func TestValidateFlowSecurityInvalidRequirement(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Security: &FlowSecurity{
					Requires: []SecurityRequirement{"invalid_requirement"},
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "invalid security requirement") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect invalid security requirement")
	}
}

func TestValidateFlowSecurityUnknownToken(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Security: &FlowSecurity{
					Token: "unknown_token",
				},
			},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown token definition") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown token reference")
	}
}

func TestValidateProtocolRoles(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "idp",
				Name: "Identity Provider",
				Type: EntityTypeAuthorizationServer,
				ProtocolRoles: []ProtocolRole{
					{Protocol: ProtocolOAuth, Role: "authorization_server"},
					{Protocol: ProtocolSCIM, Role: "service_provider"},
				},
			},
			{ID: "client", Name: "Client", Type: EntityTypeClient},
		},
		Flows: []Flow{
			{From: "client", To: "idp", Action: "request"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid protocol roles", errs)
	}
}

func TestValidateProtocolRolesInvalidProtocol(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "entity",
				Name: "Entity",
				Type: EntityTypeServer,
				ProtocolRoles: []ProtocolRole{
					{Protocol: "invalid_protocol", Role: "some_role"},
				},
			},
			{ID: "client", Name: "Client", Type: EntityTypeClient},
		},
		Flows: []Flow{
			{From: "client", To: "entity", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown protocol") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown protocol")
	}
}

func TestValidateProtocolRolesMissingRole(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{
				ID:   "entity",
				Name: "Entity",
				Type: EntityTypeServer,
				ProtocolRoles: []ProtocolRole{
					{Protocol: ProtocolOAuth, Role: ""}, // missing role
				},
			},
			{ID: "client", Name: "Client", Type: EntityTypeClient},
		},
		Flows: []Flow{
			{From: "client", To: "entity", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, ".role") && strings.Contains(e.Message, "required") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect missing role")
	}
}

func TestValidateComponents(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "auth_server", Name: "Auth Server", Type: EntityTypeAuthorizationServer},
			{ID: "token_endpoint", Name: "Token Endpoint", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{
					ID:       "idp",
					Name:     "Identity Provider",
					Type:     ComponentTypeIdP,
					Entities: []string{"auth_server", "token_endpoint"},
					Implements: []ProtocolRole{
						{Protocol: ProtocolOAuth, Role: "authorization_server"},
					},
					Examples: []string{"Okta", "Auth0"},
				},
			},
		},
		Flows: []Flow{
			{From: "auth_server", To: "token_endpoint", Action: "request"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid components", errs)
	}
}

func TestValidateComponentsDuplicateID(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "server1", Name: "Server 1", Type: EntityTypeServer},
			{ID: "server2", Name: "Server 2", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{ID: "comp", Name: "Component 1", Type: ComponentTypeIdP},
				{ID: "comp", Name: "Component 2", Type: ComponentTypeGateway}, // duplicate
			},
		},
		Flows: []Flow{
			{From: "server1", To: "server2", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "duplicate component ID") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect duplicate component ID")
	}
}

func TestValidateComponentsInvalidType(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "server1", Name: "Server 1", Type: EntityTypeServer},
			{ID: "server2", Name: "Server 2", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{ID: "comp", Name: "Component", Type: "invalid_type"},
			},
		},
		Flows: []Flow{
			{From: "server1", To: "server2", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown component type") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown component type")
	}
}

func TestValidateComponentsUnknownEntity(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "server1", Name: "Server 1", Type: EntityTypeServer},
			{ID: "server2", Name: "Server 2", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{
					ID:       "comp",
					Name:     "Component",
					Type:     ComponentTypeIdP,
					Entities: []string{"unknown_entity"},
				},
			},
		},
		Flows: []Flow{
			{From: "server1", To: "server2", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown entity") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown entity in component")
	}
}

func TestValidateTrustRelations(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "idp", Name: "IdP", Type: EntityTypeIdentityProvider},
		},
		Metadata: &ProtocolMetadata{
			TrustRelations: []TrustRelationship{
				{
					ID:          "tr1",
					From:        "client",
					To:          "idp",
					Type:        TrustTypeTrusts,
					Credentials: []string{CredentialAccessToken},
				},
			},
		},
		Flows: []Flow{
			{From: "client", To: "idp", Action: "request"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for valid trust relations", errs)
	}
}

func TestValidateTrustRelationsDuplicateID(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			TrustRelations: []TrustRelationship{
				{ID: "tr1", From: "client", To: "server", Type: TrustTypeTrusts},
				{ID: "tr1", From: "server", To: "client", Type: TrustTypeAuthenticates}, // duplicate
			},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "duplicate trust relation ID") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect duplicate trust relation ID")
	}
}

func TestValidateTrustRelationsUnknownEntity(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			TrustRelations: []TrustRelationship{
				{From: "client", To: "unknown", Type: TrustTypeTrusts},
			},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown entity or component") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown entity in trust relation")
	}
}

func TestValidateTrustRelationsInvalidType(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			TrustRelations: []TrustRelationship{
				{From: "client", To: "server", Type: "invalid_type"},
			},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown trust relation type") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown trust relation type")
	}
}

func TestValidateTrustRelationsInvalidCredential(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			TrustRelations: []TrustRelationship{
				{
					From:        "client",
					To:          "server",
					Type:        TrustTypeTrusts,
					Credentials: []string{"invalid_credential"},
				},
			},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "unknown credential type") {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should detect unknown credential type")
	}
}

func TestValidateTrustRelationsToComponent(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "test",
			Name: "Test",
		},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Metadata: &ProtocolMetadata{
			Components: []DeploymentComponent{
				{ID: "idp", Name: "IdP", Type: ComponentTypeIdP},
			},
			TrustRelations: []TrustRelationship{
				{From: "client", To: "idp", Type: TrustTypeTrusts}, // 'to' is a component
			},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}

	errs := p.Validate()
	if errs.HasErrors() {
		t.Errorf("Validate() = %v, want no errors for trust relation to component", errs)
	}
}
