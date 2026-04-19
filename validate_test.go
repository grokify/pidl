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
