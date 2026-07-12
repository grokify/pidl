package export

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

func createTestProcessProtocol() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:          "test-etl",
			Name:        "Test ETL Pipeline",
			Kind:        pidl.ProtocolKindProcess,
			Description: "A test ETL pipeline for export testing",
		},
		Entities: []pidl.Entity{
			{
				ID:       "extract",
				Name:     "Extract Data",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeExternal,
				Processing: &pidl.ProcessingConfig{
					Timeout: "30s",
				},
			},
			{
				ID:       "transform",
				Name:     "Transform Data",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeLLM,
				RetryStrategy: &pidl.RetryStrategy{
					MaxAttempts: 3,
				},
			},
			{
				ID:       "load",
				Name:     "Load Data",
				Type:     pidl.EntityTypeServer,
				StepType: pidl.StepTypeDeterministic,
			},
		},
		Flows: []pidl.Flow{
			{From: "extract", To: "transform", Action: "send"},
			{From: "transform", To: "load", Action: "send"},
		},
	}
}

func TestTemporalExporter_Export(t *testing.T) {
	p := createTestProcessProtocol()
	exporter := NewTemporalExporter()

	result, err := exporter.Export(p)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Check for expected content
	if !strings.Contains(result, "package workflow") {
		t.Error("expected package declaration")
	}

	if !strings.Contains(result, "TestEtlPipelineWorkflow") {
		t.Error("expected workflow function name")
	}

	if !strings.Contains(result, "ExtractDataActivity") {
		t.Error("expected ExtractDataActivity")
	}

	if !strings.Contains(result, "TransformDataActivity") {
		t.Error("expected TransformDataActivity")
	}

	if !strings.Contains(result, "LoadDataActivity") {
		t.Error("expected LoadDataActivity")
	}

	if !strings.Contains(result, "workflow.ExecuteActivity") {
		t.Error("expected workflow.ExecuteActivity call")
	}
}

func TestTemporalExporter_NonProcess(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "oauth",
			Name: "OAuth 2.0",
			Kind: pidl.ProtocolKindProtocol, // Not a process
		},
	}

	exporter := NewTemporalExporter()
	_, err := exporter.Export(p)

	if err == nil {
		t.Error("expected error for non-process protocol")
	}
}

func TestPrefectExporter_Export(t *testing.T) {
	p := createTestProcessProtocol()
	exporter := NewPrefectExporter()

	result, err := exporter.Export(p)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Check for expected content
	if !strings.Contains(result, "from prefect import flow, task") {
		t.Error("expected prefect imports")
	}

	if !strings.Contains(result, "test_etl_pipeline_flow") {
		t.Error("expected flow function name")
	}

	if !strings.Contains(result, "@task") {
		t.Error("expected @task decorator")
	}

	if !strings.Contains(result, "@flow") {
		t.Error("expected @flow decorator")
	}

	if !strings.Contains(result, "extract_data_task") {
		t.Error("expected extract_data_task")
	}

	if !strings.Contains(result, "async def") {
		t.Error("expected async functions")
	}
}

func TestPrefectExporter_SyncMode(t *testing.T) {
	p := createTestProcessProtocol()
	exporter := NewPrefectExporter()
	exporter.UseAsync = false

	result, err := exporter.Export(p)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if strings.Contains(result, "async def") {
		t.Error("expected no async functions in sync mode")
	}

	if strings.Contains(result, "await") {
		t.Error("expected no await in sync mode")
	}
}

func TestBPMNExporter_Export(t *testing.T) {
	p := createTestProcessProtocol()
	exporter := NewBPMNExporter()

	result, err := exporter.Export(p)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Check for expected XML content
	if !strings.Contains(result, "<?xml") {
		t.Error("expected XML declaration")
	}

	if !strings.Contains(result, "<definitions") {
		t.Error("expected definitions element")
	}

	if !strings.Contains(result, "<process") {
		t.Error("expected process element")
	}

	if !strings.Contains(result, "<startEvent") {
		t.Error("expected startEvent")
	}

	if !strings.Contains(result, "<endEvent") {
		t.Error("expected endEvent")
	}

	if !strings.Contains(result, "<sequenceFlow") {
		t.Error("expected sequenceFlow")
	}

	// Check for our tasks
	if !strings.Contains(result, "Extract Data") {
		t.Error("expected Extract Data task")
	}

	if !strings.Contains(result, "Transform Data") {
		t.Error("expected Transform Data task")
	}
}

func TestToGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "HelloWorld"},
		{"extract-data", "ExtractData"},
		{"load_data", "LoadData"},
		{"Test", "Test"},
	}

	for _, tc := range tests {
		result := toGoName(tc.input)
		if result != tc.expected {
			t.Errorf("toGoName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestToPythonName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello_world"},
		{"Extract-Data", "extract_data"},
		{"LoadData", "loaddata"},
		{"test", "test"},
	}

	for _, tc := range tests {
		result := toPythonName(tc.input)
		if result != tc.expected {
			t.Errorf("toPythonName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
