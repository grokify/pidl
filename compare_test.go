package pidl

import (
	"testing"
)

func TestCompare_NoChanges(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	new := NewMinimalProtocol("test-protocol", "Test Protocol")

	diff := Compare(base, new, DefaultDiffOptions())

	if diff.HasChanges() {
		t.Errorf("expected no changes, got %d", len(diff.Items))
	}

	if diff.Summary.TotalChanges != 0 {
		t.Errorf("expected 0 total changes, got %d", diff.Summary.TotalChanges)
	}
}

func TestCompare_EntityAdded(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	new := NewMinimalProtocol("test-protocol", "Test Protocol")
	new.Entities = append(new.Entities, Entity{
		ID:   "new_entity",
		Name: "New Entity",
		Type: EntityTypeServer,
	})

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	if diff.Summary.Added != 1 {
		t.Errorf("expected 1 addition, got %d", diff.Summary.Added)
	}

	if diff.Summary.ByCategory[DiffCategoryEntity] != 1 {
		t.Errorf("expected 1 entity change, got %d", diff.Summary.ByCategory[DiffCategoryEntity])
	}
}

func TestCompare_EntityRemoved(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	base.Entities = append(base.Entities, Entity{
		ID:   "old_entity",
		Name: "Old Entity",
		Type: EntityTypeServer,
	})
	new := NewMinimalProtocol("test-protocol", "Test Protocol")

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	if diff.Summary.Removed != 1 {
		t.Errorf("expected 1 removal, got %d", diff.Summary.Removed)
	}
}

func TestCompare_EntityModified(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	base.Entities = append(base.Entities, Entity{
		ID:   "entity",
		Name: "Entity V1",
		Type: EntityTypeServer,
	})
	new := NewMinimalProtocol("test-protocol", "Test Protocol")
	new.Entities = append(new.Entities, Entity{
		ID:   "entity",
		Name: "Entity V2",
		Type: EntityTypeServer,
	})

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	if diff.Summary.Modified != 1 {
		t.Errorf("expected 1 modification, got %d", diff.Summary.Modified)
	}
}

func TestCompare_FlowAdded(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	new := NewMinimalProtocol("test-protocol", "Test Protocol")
	new.Flows = append(new.Flows, Flow{
		From:   "client",
		To:     "server",
		Action: "request",
	})

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	if diff.Summary.Added != 1 {
		t.Errorf("expected 1 addition, got %d", diff.Summary.Added)
	}

	if diff.Summary.ByCategory[DiffCategoryFlow] != 1 {
		t.Errorf("expected 1 flow change, got %d", diff.Summary.ByCategory[DiffCategoryFlow])
	}
}

func TestCompare_FlowModified(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	base.Flows = []Flow{{
		From:   "client",
		To:     "server",
		Action: "request",
		Mode:   FlowModeRequest,
	}}

	new := NewMinimalProtocol("test-protocol", "Test Protocol")
	new.Flows = []Flow{{
		From:   "client",
		To:     "server",
		Action: "modified_request",
		Mode:   FlowModeRequest,
	}}

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	found := false
	for _, item := range diff.Items {
		if item.Path == "flows[0].action" && item.Type == DiffTypeModified {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected flow action modification")
	}
}

func TestCompare_PhaseChanges(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	base.Phases = []Phase{{
		ID:   "phase1",
		Name: "Phase 1",
	}}

	new := NewMinimalProtocol("test-protocol", "Test Protocol")
	new.Phases = []Phase{
		{ID: "phase1", Name: "Phase 1 Updated"},
		{ID: "phase2", Name: "Phase 2"},
	}

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	// Should have one modification (name change) and one addition
	if diff.Summary.Added != 1 {
		t.Errorf("expected 1 addition, got %d", diff.Summary.Added)
	}

	if diff.Summary.Modified != 1 {
		t.Errorf("expected 1 modification, got %d", diff.Summary.Modified)
	}
}

func TestCompare_IgnoreDescriptions(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	base.ProtocolMeta.Description = "Old description"

	new := NewMinimalProtocol("test-protocol", "Test Protocol")
	new.ProtocolMeta.Description = "New description"

	opts := DefaultDiffOptions()
	opts.IgnoreDescriptions = true

	diff := Compare(base, new, opts)

	if diff.HasChanges() {
		t.Error("expected no changes when ignoring descriptions")
	}
}

func TestCompare_IgnoreMetadata(t *testing.T) {
	base := NewMinimalProtocol("test-protocol", "Test Protocol")
	base.ProtocolMeta.Version = "1.0.0"

	new := NewMinimalProtocol("test-protocol", "Test Protocol")
	new.ProtocolMeta.Version = "2.0.0"

	opts := DefaultDiffOptions()
	opts.IgnoreMetadata = true

	diff := Compare(base, new, opts)

	if diff.HasChanges() {
		t.Error("expected no changes when ignoring metadata")
	}
}

func TestCompare_String(t *testing.T) {
	base := NewMinimalProtocol("base", "Base Protocol")
	new := NewMinimalProtocol("new", "New Protocol")
	new.Entities = append(new.Entities, Entity{
		ID:   "added",
		Name: "Added Entity",
		Type: EntityTypeServer,
	})

	diff := Compare(base, new, DefaultDiffOptions())
	output := diff.String()

	if output == "" {
		t.Error("expected non-empty string output")
	}

	if len(output) < 50 {
		t.Errorf("output seems too short: %s", output)
	}
}

func TestCompare_ToMarkdown(t *testing.T) {
	base := NewMinimalProtocol("base", "Base Protocol")
	new := NewMinimalProtocol("new", "New Protocol")
	new.Entities = append(new.Entities, Entity{
		ID:   "added",
		Name: "Added Entity",
		Type: EntityTypeServer,
	})

	diff := Compare(base, new, DefaultDiffOptions())
	output := diff.ToMarkdown()

	if output == "" {
		t.Error("expected non-empty markdown output")
	}

	// Should contain markdown headers
	if len(output) < 50 {
		t.Errorf("output seems too short: %s", output)
	}
}

func TestCompare_ToJSON(t *testing.T) {
	base := NewMinimalProtocol("base", "Base Protocol")
	new := NewMinimalProtocol("new", "New Protocol")
	new.Entities = append(new.Entities, Entity{
		ID:   "added",
		Name: "Added Entity",
		Type: EntityTypeServer,
	})

	diff := Compare(base, new, DefaultDiffOptions())
	jsonBytes, err := diff.ToJSON()

	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

func TestCompare_EntityStateChanges(t *testing.T) {
	base := NewMinimalProtocol("test", "Test")
	base.Entities = []Entity{{
		ID:   "client",
		Name: "Client",
		Type: EntityTypeClient,
		States: []EntityState{
			{ID: "idle", Name: "Idle", Initial: true},
		},
	}}

	new := NewMinimalProtocol("test", "Test")
	new.Entities = []Entity{{
		ID:   "client",
		Name: "Client",
		Type: EntityTypeClient,
		States: []EntityState{
			{ID: "idle", Name: "Idle", Initial: true},
			{ID: "active", Name: "Active"},
		},
	}}

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	found := false
	for _, item := range diff.Items {
		if item.Category == DiffCategoryEntity && item.Type == DiffTypeAdded {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected entity state addition")
	}
}

func TestCompare_TrustLevelChange(t *testing.T) {
	base := NewMinimalProtocol("test", "Test")
	base.Entities = []Entity{{
		ID:         "server",
		Name:       "Server",
		Type:       EntityTypeServer,
		TrustLevel: TrustLevelTrusted,
	}}

	new := NewMinimalProtocol("test", "Test")
	new.Entities = []Entity{{
		ID:         "server",
		Name:       "Server",
		Type:       EntityTypeServer,
		TrustLevel: TrustLevelSemiTrusted,
	}}

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	found := false
	for _, item := range diff.Items {
		if item.Path == "entities[server].trust_level" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected trust level change")
	}
}

func TestCompare_SecurityChange(t *testing.T) {
	base := NewMinimalProtocol("test", "Test")
	base.Flows = []Flow{{
		From:   "client",
		To:     "server",
		Action: "request",
		Security: &FlowSecurity{
			Requires: []SecurityRequirement{SecurityRequirementToken},
		},
	}}

	new := NewMinimalProtocol("test", "Test")
	new.Flows = []Flow{{
		From:   "client",
		To:     "server",
		Action: "request",
		Security: &FlowSecurity{
			Requires: []SecurityRequirement{SecurityRequirementToken, SecurityRequirementMTLS},
		},
	}}

	diff := Compare(base, new, DefaultDiffOptions())

	if !diff.HasChanges() {
		t.Error("expected changes")
	}

	found := false
	for _, item := range diff.Items {
		if item.Path == "flows[0].security" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected security change")
	}
}
