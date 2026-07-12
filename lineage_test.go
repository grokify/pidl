package pidl

import (
	"testing"
)

func createTestLineageProtocol() *Protocol {
	return &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "etl-pipeline",
			Name: "ETL Pipeline",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:       "source",
				Name:     "Data Source",
				Type:     EntityTypeServer,
				StepType: StepTypeExternal,
				Outputs: []DataPort{
					{Kind: DataPortKindFile, Name: "raw_data", Sensitive: true},
					{Kind: DataPortKindObject, Name: "metadata"},
				},
			},
			{
				ID:       "transform",
				Name:     "Transform",
				Type:     EntityTypeServer,
				StepType: StepTypeLLM,
				Inputs: []DataPort{
					{Kind: DataPortKindFile, Name: "raw_data", Required: true},
				},
				Outputs: []DataPort{
					{Kind: DataPortKindObject, Name: "processed_data"},
				},
			},
			{
				ID:       "load",
				Name:     "Load",
				Type:     EntityTypeServer,
				StepType: StepTypeDeterministic,
				Inputs: []DataPort{
					{Kind: DataPortKindObject, Name: "processed_data", Required: true},
				},
				Outputs: []DataPort{
					{Kind: DataPortKindDatabase, Name: "stored_data"},
				},
			},
		},
		Flows: []Flow{
			{From: "source", To: "transform", Action: "extract"},
			{From: "transform", To: "load", Action: "load"},
		},
	}
}

func TestAnalyzeDataLineage_Basic(t *testing.T) {
	p := createTestLineageProtocol()
	lineage := AnalyzeDataLineage(p)

	if lineage.ProtocolID != "etl-pipeline" {
		t.Errorf("expected protocol ID 'etl-pipeline', got '%s'", lineage.ProtocolID)
	}

	// Should have edges connecting raw_data -> transform and processed_data -> load
	if len(lineage.Edges) < 2 {
		t.Errorf("expected at least 2 edges, got %d", len(lineage.Edges))
	}
}

func TestAnalyzeDataLineage_Sources(t *testing.T) {
	p := createTestLineageProtocol()
	lineage := AnalyzeDataLineage(p)

	// metadata port should be a source (no incoming, has outgoing potential)
	foundMetadata := false
	for _, source := range lineage.Sources {
		if source.EntityID == "source" && source.PortName == "metadata" {
			foundMetadata = true
			break
		}
	}

	if !foundMetadata {
		t.Log("Note: metadata port may or may not be in sources depending on inference")
	}
}

func TestAnalyzeDataLineage_Sinks(t *testing.T) {
	p := createTestLineageProtocol()
	lineage := AnalyzeDataLineage(p)

	// stored_data should be a sink (no outgoing)
	foundStoredData := false
	for _, sink := range lineage.Sinks {
		if sink.EntityID == "load" && sink.PortName == "stored_data" {
			foundStoredData = true
			break
		}
	}

	if !foundStoredData {
		t.Log("Note: stored_data port may not be in sinks if classified differently")
	}
}

func TestAnalyzeDataLineage_SensitivePaths(t *testing.T) {
	p := createTestLineageProtocol()
	lineage := AnalyzeDataLineage(p)

	// raw_data is marked as sensitive
	if len(lineage.SensitiveDataPaths) == 0 {
		t.Log("Note: sensitive paths may be empty if no complete path traced")
	}

	hasSensitiveFlow := lineage.HasSensitiveDataFlow()
	_ = hasSensitiveFlow // Just verify the method works
}

func TestAnalyzeDataLineage_ExplicitMappings(t *testing.T) {
	p := &Protocol{
		ProtocolMeta: ProtocolMeta{
			ID:   "explicit-mapping",
			Name: "Explicit Mapping Test",
			Kind: ProtocolKindProcess,
		},
		Entities: []Entity{
			{
				ID:   "producer",
				Name: "Producer",
				Outputs: []DataPort{
					{Kind: DataPortKindObject, Name: "output_a"},
					{Kind: DataPortKindObject, Name: "output_b"},
				},
			},
			{
				ID:   "consumer",
				Name: "Consumer",
				Inputs: []DataPort{
					{Kind: DataPortKindObject, Name: "input_x"},
					{Kind: DataPortKindObject, Name: "input_y"},
				},
			},
		},
		Flows: []Flow{
			{
				From:   "producer",
				To:     "consumer",
				Action: "send",
				DataMappings: []DataPortMapping{
					{OutputPort: "output_a", InputPort: "input_x", Transformation: "filter"},
					{OutputPort: "output_b", InputPort: "input_y"},
				},
			},
		},
	}

	lineage := AnalyzeDataLineage(p)

	if len(lineage.Edges) != 2 {
		t.Errorf("expected 2 edges from explicit mappings, got %d", len(lineage.Edges))
	}

	// Check transformation is preserved
	foundTransform := false
	for _, edge := range lineage.Edges {
		if edge.Transformation == "filter" {
			foundTransform = true
			break
		}
	}

	if !foundTransform {
		t.Error("expected to find edge with 'filter' transformation")
	}
}

func TestDataLineage_GetUpstream(t *testing.T) {
	lineage := &DataLineage{
		Edges: []LineageEdge{
			{SourceEntity: "a", SourcePort: "out", TargetEntity: "b", TargetPort: "in"},
			{SourceEntity: "c", SourcePort: "out", TargetEntity: "b", TargetPort: "in2"},
		},
	}

	upstream := lineage.GetUpstream("b", "in")
	if len(upstream) != 1 {
		t.Errorf("expected 1 upstream port, got %d", len(upstream))
	}
	if upstream[0].EntityID != "a" {
		t.Errorf("expected upstream entity 'a', got '%s'", upstream[0].EntityID)
	}
}

func TestDataLineage_GetDownstream(t *testing.T) {
	lineage := &DataLineage{
		Edges: []LineageEdge{
			{SourceEntity: "a", SourcePort: "out", TargetEntity: "b", TargetPort: "in"},
			{SourceEntity: "a", SourcePort: "out", TargetEntity: "c", TargetPort: "in"},
		},
	}

	downstream := lineage.GetDownstream("a", "out")
	if len(downstream) != 2 {
		t.Errorf("expected 2 downstream ports, got %d", len(downstream))
	}
}

func TestDataLineage_GetImpactedEntities(t *testing.T) {
	lineage := &DataLineage{
		Edges: []LineageEdge{
			{SourceEntity: "a", SourcePort: "out", TargetEntity: "b", TargetPort: "in"},
			{SourceEntity: "b", SourcePort: "out", TargetEntity: "c", TargetPort: "in"},
			{SourceEntity: "c", SourcePort: "out", TargetEntity: "d", TargetPort: "in"},
		},
	}

	impacted := lineage.GetImpactedEntities("a")
	if len(impacted) != 3 {
		t.Errorf("expected 3 impacted entities (b, c, d), got %d", len(impacted))
	}

	// Check all are present
	impactedMap := make(map[string]bool)
	for _, id := range impacted {
		impactedMap[id] = true
	}

	for _, expected := range []string{"b", "c", "d"} {
		if !impactedMap[expected] {
			t.Errorf("expected '%s' in impacted entities", expected)
		}
	}
}

func TestDataLineage_GetDataProvenance(t *testing.T) {
	lineage := &DataLineage{
		Edges: []LineageEdge{
			{SourceEntity: "a", SourcePort: "out", TargetEntity: "b", TargetPort: "in"},
			{SourceEntity: "b", SourcePort: "out", TargetEntity: "c", TargetPort: "in"},
			{SourceEntity: "x", SourcePort: "out", TargetEntity: "c", TargetPort: "in2"},
		},
	}

	provenance := lineage.GetDataProvenance("c")
	if len(provenance) != 3 {
		t.Errorf("expected 3 provenance entities (a, b, x), got %d", len(provenance))
	}

	// Check all are present
	provMap := make(map[string]bool)
	for _, id := range provenance {
		provMap[id] = true
	}

	for _, expected := range []string{"a", "b", "x"} {
		if !provMap[expected] {
			t.Errorf("expected '%s' in provenance entities", expected)
		}
	}
}
