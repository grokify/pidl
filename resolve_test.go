package pidl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProtocolResolve_NoImports(t *testing.T) {
	p := NewMinimalProtocol("test", "Test")

	resolved, err := p.Resolve(DefaultResolveOptions())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if !resolved.IsResolved() {
		t.Error("Resolved protocol should have IsResolved() = true")
	}

	if len(resolved.Entities) != len(p.Entities) {
		t.Errorf("Entities count = %d, want %d", len(resolved.Entities), len(p.Entities))
	}
}

func TestProtocolResolve_Extends(t *testing.T) {
	// Create temp directory with base and derived protocols
	dir := t.TempDir()

	// Create base protocol
	base := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "base", Name: "Base"},
		Entities: []Entity{
			{ID: "a", Name: "Entity A", Type: EntityTypeClient},
			{ID: "b", Name: "Entity B", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "phase1", Name: "Phase 1"},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "request"},
		},
	}
	baseFile := filepath.Join(dir, "base.json")
	if err := base.WriteFile(baseFile); err != nil {
		t.Fatal(err)
	}

	// Create derived protocol that extends base
	derived := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "derived", Name: "Derived"},
		Extends: &ProtocolExtends{
			Path: "base.json",
		},
		Entities: []Entity{
			{ID: "c", Name: "Entity C", Type: EntityTypeAgent},
		},
		Flows: []Flow{
			{From: "c", To: "a", Action: "delegate"},
		},
	}
	derivedFile := filepath.Join(dir, "derived.json")
	if err := derived.WriteFile(derivedFile); err != nil {
		t.Fatal(err)
	}

	// Parse and resolve
	p, err := ParseFile(derivedFile)
	if err != nil {
		t.Fatal(err)
	}

	opts := DefaultResolveOptions()
	opts.BasePath = dir
	resolved, err := p.Resolve(opts)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// Should have entities from both base and derived
	if len(resolved.Entities) != 3 {
		t.Errorf("Entities count = %d, want 3", len(resolved.Entities))
	}

	// Should have phases from base
	if len(resolved.Phases) != 1 {
		t.Errorf("Phases count = %d, want 1", len(resolved.Phases))
	}

	// Should have flows from both
	if len(resolved.Flows) != 2 {
		t.Errorf("Flows count = %d, want 2", len(resolved.Flows))
	}

	// Check that extends is cleared after resolution
	if resolved.Extends != nil {
		t.Error("Extends should be nil after resolution")
	}
}

func TestProtocolResolve_Import(t *testing.T) {
	dir := t.TempDir()

	// Create protocol to import
	imported := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "imported", Name: "Imported"},
		Entities: []Entity{
			{ID: "x", Name: "Entity X", Type: EntityTypeClient},
			{ID: "y", Name: "Entity Y", Type: EntityTypeServer},
		},
		Phases: []Phase{
			{ID: "import_phase", Name: "Import Phase"},
		},
		Flows: []Flow{
			{From: "x", To: "y", Action: "call", Phase: "import_phase"},
		},
	}
	importedFile := filepath.Join(dir, "imported.json")
	if err := imported.WriteFile(importedFile); err != nil {
		t.Fatal(err)
	}

	// Create main protocol that imports
	main := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "main", Name: "Main"},
		Imports: []ProtocolImport{
			{
				Path:         "imported.json",
				Entities:     []string{"x"},
				Phases:       []string{"import_phase"},
				IncludeFlows: false,
			},
		},
		Entities: []Entity{
			{ID: "m", Name: "Main Entity", Type: EntityTypeAgent},
		},
		Flows: []Flow{
			{From: "m", To: "x", Action: "use"},
		},
	}
	mainFile := filepath.Join(dir, "main.json")
	if err := main.WriteFile(mainFile); err != nil {
		t.Fatal(err)
	}

	p, err := ParseFile(mainFile)
	if err != nil {
		t.Fatal(err)
	}

	opts := DefaultResolveOptions()
	opts.BasePath = dir
	resolved, err := p.Resolve(opts)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// Should have main entity + imported entity x (not y)
	if len(resolved.Entities) != 2 {
		t.Errorf("Entities count = %d, want 2", len(resolved.Entities))
	}

	// Should have imported phase
	if len(resolved.Phases) != 1 {
		t.Errorf("Phases count = %d, want 1", len(resolved.Phases))
	}

	// Should only have main flow (no imported flows since IncludeFlows=false)
	if len(resolved.Flows) != 1 {
		t.Errorf("Flows count = %d, want 1", len(resolved.Flows))
	}
}

func TestProtocolResolve_ImportWithAlias(t *testing.T) {
	dir := t.TempDir()

	imported := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "imported", Name: "Imported"},
		Entities: []Entity{
			{ID: "client", Name: "Client", Type: EntityTypeClient},
		},
		Flows: []Flow{},
	}
	importedFile := filepath.Join(dir, "imported.json")
	if err := imported.WriteFile(importedFile); err != nil {
		t.Fatal(err)
	}

	main := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "main", Name: "Main"},
		Imports: []ProtocolImport{
			{
				Path:  "imported.json",
				Alias: "oauth_",
			},
		},
		Entities: []Entity{
			{ID: "client", Name: "Main Client", Type: EntityTypeClient},
			{ID: "server", Name: "Server", Type: EntityTypeServer},
		},
		Flows: []Flow{
			{From: "client", To: "server", Action: "request"},
		},
	}
	mainFile := filepath.Join(dir, "main.json")
	if err := main.WriteFile(mainFile); err != nil {
		t.Fatal(err)
	}

	p, err := ParseFile(mainFile)
	if err != nil {
		t.Fatal(err)
	}

	opts := DefaultResolveOptions()
	opts.BasePath = dir
	resolved, err := p.Resolve(opts)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// Should have both clients (main and aliased imported)
	if len(resolved.Entities) != 3 {
		t.Errorf("Entities count = %d, want 3", len(resolved.Entities))
	}

	// Check for aliased entity
	found := false
	for _, e := range resolved.Entities {
		if e.ID == "oauth_client" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected aliased entity oauth_client not found")
	}
}

func TestProtocolResolve_CircularExtends(t *testing.T) {
	dir := t.TempDir()

	// Create two protocols that extend each other
	p1 := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "p1", Name: "P1"},
		Extends:      &ProtocolExtends{Path: "p2.json"},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{{From: "a", To: "b", Action: "x"}},
	}
	p1File := filepath.Join(dir, "p1.json")
	if err := p1.WriteFile(p1File); err != nil {
		t.Fatal(err)
	}

	p2 := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "p2", Name: "P2"},
		Extends:      &ProtocolExtends{Path: "p1.json"},
		Entities: []Entity{
			{ID: "c", Name: "C", Type: EntityTypeAgent},
			{ID: "d", Name: "D", Type: EntityTypeServer},
		},
		Flows: []Flow{{From: "c", To: "d", Action: "y"}},
	}
	p2File := filepath.Join(dir, "p2.json")
	if err := p2.WriteFile(p2File); err != nil {
		t.Fatal(err)
	}

	p, err := ParseFile(p1File)
	if err != nil {
		t.Fatal(err)
	}

	opts := DefaultResolveOptions()
	opts.BasePath = dir
	_, err = p.Resolve(opts)
	if err == nil {
		t.Error("Expected circular extends error, got nil")
	}
}

func TestProtocolResolve_CircularImport(t *testing.T) {
	dir := t.TempDir()

	p1 := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "p1", Name: "P1"},
		Imports: []ProtocolImport{
			{Path: "p2.json"},
		},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
		},
		Flows: []Flow{{From: "a", To: "b", Action: "x"}},
	}
	p1File := filepath.Join(dir, "p1.json")
	if err := p1.WriteFile(p1File); err != nil {
		t.Fatal(err)
	}

	p2 := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "p2", Name: "P2"},
		Imports: []ProtocolImport{
			{Path: "p1.json"},
		},
		Entities: []Entity{
			{ID: "c", Name: "C", Type: EntityTypeAgent},
			{ID: "d", Name: "D", Type: EntityTypeServer},
		},
		Flows: []Flow{{From: "c", To: "d", Action: "y"}},
	}
	p2File := filepath.Join(dir, "p2.json")
	if err := p2.WriteFile(p2File); err != nil {
		t.Fatal(err)
	}

	p, err := ParseFile(p1File)
	if err != nil {
		t.Fatal(err)
	}

	opts := DefaultResolveOptions()
	opts.BasePath = dir
	_, err = p.Resolve(opts)
	if err == nil {
		t.Error("Expected circular import error, got nil")
	}
}

func TestProtocolResolve_MaxDepth(t *testing.T) {
	dir := t.TempDir()

	// Create a chain of 15 files extending each other
	for i := 14; i >= 0; i-- {
		var extends *ProtocolExtends
		if i < 14 {
			extends = &ProtocolExtends{Path: filepath.Base(filepath.Join(dir, "p"+string(rune('a'+i+1))+".json"))}
		}

		p := &Protocol{
			ProtocolMeta: ProtocolMeta{
				ID:   "p" + string(rune('a'+i)),
				Name: "P" + string(rune('A'+i)),
			},
			Extends: extends,
			Entities: []Entity{
				{ID: "e" + string(rune('a'+i)), Name: "E", Type: EntityTypeClient},
				{ID: "s" + string(rune('a'+i)), Name: "S", Type: EntityTypeServer},
			},
			Flows: []Flow{
				{From: "e" + string(rune('a'+i)), To: "s" + string(rune('a'+i)), Action: "x"},
			},
		}
		pFile := filepath.Join(dir, "p"+string(rune('a'+i))+".json")
		if err := p.WriteFile(pFile); err != nil {
			t.Fatal(err)
		}
	}

	p, err := ParseFile(filepath.Join(dir, "pa.json"))
	if err != nil {
		t.Fatal(err)
	}

	opts := DefaultResolveOptions()
	opts.BasePath = dir
	opts.MaxDepth = 5 // Less than the chain depth

	_, err = p.Resolve(opts)
	if err == nil {
		t.Error("Expected max depth error, got nil")
	}
}

func TestProtocolResolve_ExcludeEntities(t *testing.T) {
	dir := t.TempDir()

	base := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "base", Name: "Base"},
		Entities: []Entity{
			{ID: "a", Name: "A", Type: EntityTypeClient},
			{ID: "b", Name: "B", Type: EntityTypeServer},
			{ID: "c", Name: "C", Type: EntityTypeAgent},
		},
		Flows: []Flow{
			{From: "a", To: "b", Action: "x"},
			{From: "b", To: "c", Action: "y"},
		},
	}
	baseFile := filepath.Join(dir, "base.json")
	if err := base.WriteFile(baseFile); err != nil {
		t.Fatal(err)
	}

	derived := &Protocol{
		ProtocolMeta: ProtocolMeta{ID: "derived", Name: "Derived"},
		Extends: &ProtocolExtends{
			Path:            "base.json",
			ExcludeEntities: []string{"c"},
		},
		Entities: []Entity{},
		Flows:    []Flow{},
	}
	derivedFile := filepath.Join(dir, "derived.json")
	if err := derived.WriteFile(derivedFile); err != nil {
		t.Fatal(err)
	}

	p, err := ParseFile(derivedFile)
	if err != nil {
		t.Fatal(err)
	}

	opts := DefaultResolveOptions()
	opts.BasePath = dir
	resolved, err := p.Resolve(opts)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// Should have 2 entities (c excluded)
	if len(resolved.Entities) != 2 {
		t.Errorf("Entities count = %d, want 2", len(resolved.Entities))
	}

	// Should have 1 flow (flow involving c excluded)
	if len(resolved.Flows) != 1 {
		t.Errorf("Flows count = %d, want 1", len(resolved.Flows))
	}
}

func TestProtocolNeedsResolution(t *testing.T) {
	tests := []struct {
		name string
		p    *Protocol
		want bool
	}{
		{
			name: "no imports or extends",
			p:    NewMinimalProtocol("test", "Test"),
			want: false,
		},
		{
			name: "has extends",
			p: &Protocol{
				Extends: &ProtocolExtends{Path: "base.json"},
			},
			want: true,
		},
		{
			name: "has imports",
			p: &Protocol{
				Imports: []ProtocolImport{{Path: "other.json"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.NeedsResolution(); got != tt.want {
				t.Errorf("NeedsResolution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProtocolResolve_RealStdlib(t *testing.T) {
	// Test with actual stdlib files
	examplesDir := "examples"
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("examples directory not found")
	}

	composedFile := filepath.Join(examplesDir, "oauth_mcp_composed.json")
	if _, err := os.Stat(composedFile); os.IsNotExist(err) {
		t.Skip("oauth_mcp_composed.json not found")
	}

	p, err := ParseFile(composedFile)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if !p.NeedsResolution() {
		t.Fatal("Expected composed protocol to need resolution")
	}

	opts := DefaultResolveOptions()
	opts.BasePath = examplesDir
	resolved, err := p.Resolve(opts)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if !resolved.IsResolved() {
		t.Error("Resolved protocol should have IsResolved() = true")
	}

	// Check that we have entities from oauth_entities.json and mcp_entities.json
	entityIDs := make(map[string]bool)
	for _, e := range resolved.Entities {
		entityIDs[e.ID] = true
	}

	// From oauth_entities.json
	if !entityIDs["authorization_server"] {
		t.Error("Missing entity authorization_server from oauth base")
	}
	if !entityIDs["client"] {
		t.Error("Missing entity client from oauth base")
	}

	// From mcp_entities.json (imported)
	if !entityIDs["mcp_client"] {
		t.Error("Missing entity mcp_client from mcp import")
	}
	if !entityIDs["mcp_server"] {
		t.Error("Missing entity mcp_server from mcp import")
	}

	// From derived
	if !entityIDs["gateway"] {
		t.Error("Missing entity gateway from derived")
	}

	// Should be valid after resolution
	if !resolved.IsValid() {
		errs := resolved.Validate()
		t.Errorf("Resolved protocol is not valid: %v", errs)
	}
}
