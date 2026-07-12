package export

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

// exporter is a local interface for testing export functions.
type exporter interface {
	Export(*pidl.Protocol) (string, error)
}

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

// assertExportContains is a test helper that exports using the given exporter
// and verifies the result contains all expected substrings.
func assertExportContains(t *testing.T, e exporter, expectations []struct{ substr, errMsg string }) {
	t.Helper()
	p := createTestProcessProtocol()

	result, err := e.Export(p)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	for _, exp := range expectations {
		if !strings.Contains(result, exp.substr) {
			t.Error(exp.errMsg)
		}
	}
}

func TestBPMNExporter_Export(t *testing.T) {
	assertExportContains(t, NewBPMNExporter(), []struct{ substr, errMsg string }{
		{"<?xml", "expected XML declaration"},
		{"<definitions", "expected definitions element"},
		{"<process", "expected process element"},
		{"<startEvent", "expected startEvent"},
		{"<endEvent", "expected endEvent"},
		{"<sequenceFlow", "expected sequenceFlow"},
		{"Extract Data", "expected Extract Data task"},
		{"Transform Data", "expected Transform Data task"},
	})
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

func TestStepFunctionsExporter_Export(t *testing.T) {
	assertExportContains(t, NewStepFunctionsExporter(), []struct{ substr, errMsg string }{
		{`"StartAt"`, "expected StartAt field in state machine"},
		{`"States"`, "expected States field in state machine"},
		{`"Comment"`, "expected Comment field in state machine"},
		{`"Type": "Task"`, "expected Task state type"},
		{"lambda:invoke", "expected lambda:invoke for deterministic step"},
		{"bedrock:invokeModel", "expected bedrock:invokeModel for LLM step"},
		{"http:invoke", "expected http:invoke for external step"},
		{`"Retry"`, "expected Retry field for step with retry strategy"},
	})
}

func TestStepFunctionsExporter_NonProcess(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "non-process",
			Name: "Non-Process Protocol",
			Kind: pidl.ProtocolKindProtocol,
		},
	}

	exporter := NewStepFunctionsExporter()
	_, err := exporter.Export(p)
	if err == nil {
		t.Error("expected error for non-process protocol")
	}
}

func TestStepFunctionsExporter_CustomName(t *testing.T) {
	p := createTestProcessProtocol()
	exporter := NewStepFunctionsExporter()
	exporter.StateMachineName = "CustomStateMachine"
	exporter.Comment = "Custom comment"

	result, err := exporter.Export(p)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if !strings.Contains(result, "Custom comment") {
		t.Error("expected custom comment in output")
	}
}

func TestStepFunctionsExporter_HumanStep(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "human-review",
			Name: "Human Review Process",
			Kind: pidl.ProtocolKindProcess,
		},
		Entities: []pidl.Entity{
			{
				ID:       "review",
				Name:     "Human Review",
				Type:     pidl.EntityTypeUser,
				StepType: pidl.StepTypeHuman,
			},
		},
	}

	exporter := NewStepFunctionsExporter()
	result, err := exporter.Export(p)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Check for activity resource for human step
	if !strings.Contains(result, "activity:") {
		t.Error("expected activity resource for human step")
	}

	// Check for heartbeat
	if !strings.Contains(result, `"HeartbeatSeconds"`) {
		t.Error("expected HeartbeatSeconds for human step")
	}
}

func TestToStateName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"extract", "extract"},
		{"extract-data", "extract-data"},
		{"extract_data", "extract_data"},
		{"Extract Data", "Extract_Data"},
		{"step.name", "step.name"},
	}

	for _, tc := range tests {
		result := toStateName(tc.input)
		if result != tc.expected {
			t.Errorf("toStateName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
