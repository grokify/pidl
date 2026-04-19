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
