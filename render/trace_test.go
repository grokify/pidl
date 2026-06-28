package render

import (
	"strings"
	"testing"
	"time"

	"github.com/grokify/pidl"
)

func createTestProtocol() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test-protocol",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{ID: "client", Name: "Client", Type: pidl.EntityTypeClient},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request", Mode: pidl.FlowModeRequest},
			{From: "server", To: "client", Action: "response", Mode: pidl.FlowModeResponse},
		},
	}
}

func createTestTrace() *pidl.ExecutionTrace {
	now := time.Now()
	return &pidl.ExecutionTrace{
		ProtocolID:   "test-protocol",
		ProtocolName: "Test Protocol",
		StartTime:    now,
		EndTime:      now.Add(100 * time.Millisecond),
		Steps: []pidl.ExecutionStep{
			{
				StepNumber: 1,
				FlowIndex:  0,
				Timestamp:  now,
				From:       "client",
				To:         "server",
				Action:     "request",
				Mode:       pidl.FlowModeRequest,
			},
			{
				StepNumber: 2,
				FlowIndex:  1,
				Timestamp:  now.Add(50 * time.Millisecond),
				From:       "server",
				To:         "client",
				Action:     "response",
				Mode:       pidl.FlowModeResponse,
			},
		},
		InitialStates: map[string]string{"client": "idle"},
		FinalStates:   map[string]string{"client": "done"},
		Completed:     true,
	}
}

func createTraceWithStateChanges() *pidl.ExecutionTrace {
	trace := createTestTrace()
	trace.Steps[0].StateChanges = []pidl.StateChange{
		{Entity: "client", FromState: "idle", ToState: "waiting"},
	}
	trace.Steps[1].StateChanges = []pidl.StateChange{
		{Entity: "client", FromState: "waiting", ToState: "done"},
	}
	return trace
}

func createTraceWithSkippedStep() *pidl.ExecutionTrace {
	trace := createTestTrace()
	condFalse := false
	trace.Steps[1] = pidl.ExecutionStep{
		StepNumber:   2,
		FlowIndex:    1,
		From:         "server",
		To:           "client",
		Action:       "response",
		Mode:         pidl.FlowModeResponse,
		Condition:    "success",
		ConditionMet: &condFalse,
		Skipped:      true,
		SkipReason:   `condition "success" not met`,
	}
	trace.Completed = false
	return trace
}

func TestTraceRenderer_RenderText(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTestTrace()
	opts := DefaultTraceTextOptions()

	result := renderer.RenderText(trace, opts)

	if result == "" {
		t.Error("expected non-empty result")
	}

	if !strings.Contains(result, "Test Protocol") {
		t.Error("expected protocol name in output")
	}

	if !strings.Contains(result, "Step 1") {
		t.Error("expected step 1 in output")
	}

	if !strings.Contains(result, "Step 2") {
		t.Error("expected step 2 in output")
	}

	if !strings.Contains(result, "client") {
		t.Error("expected client in output")
	}

	if !strings.Contains(result, "server") {
		t.Error("expected server in output")
	}
}

func TestTraceRenderer_RenderTextCompact(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTestTrace()
	opts := TraceTextOptions{
		Compact: true,
	}

	result := renderer.RenderText(trace, opts)

	// Compact mode should be shorter
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}

	// Should still contain key information
	if !strings.Contains(result, "client") {
		t.Error("expected client in output")
	}
}

func TestTraceRenderer_RenderTextWithTimestamps(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTestTrace()
	opts := TraceTextOptions{
		ShowTimestamps: true,
	}

	result := renderer.RenderText(trace, opts)

	if !strings.Contains(result, "Started:") {
		t.Error("expected timestamp in output")
	}

	if !strings.Contains(result, "Duration:") {
		t.Error("expected duration in output")
	}
}

func TestTraceRenderer_RenderTextWithStateChanges(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTraceWithStateChanges()
	opts := TraceTextOptions{
		ShowStates: true,
	}

	result := renderer.RenderText(trace, opts)

	if !strings.Contains(result, "State Changes") {
		t.Error("expected state changes in output")
	}

	if !strings.Contains(result, "idle") {
		t.Error("expected initial state in output")
	}

	if !strings.Contains(result, "done") {
		t.Error("expected final state in output")
	}
}

func TestTraceRenderer_RenderTextWithSkippedStep(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTraceWithSkippedStep()
	opts := DefaultTraceTextOptions()

	result := renderer.RenderText(trace, opts)

	if !strings.Contains(result, "skipped") || !strings.Contains(result, "Skipped") {
		t.Error("expected skipped indication in output")
	}

	if !strings.Contains(result, "Steps skipped:  1") {
		t.Error("expected skipped count of 1")
	}
}

func TestTraceRenderer_RenderMermaid(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTestTrace()
	protocol := createTestProtocol()

	result := renderer.RenderMermaid(trace, protocol)

	if result == "" {
		t.Error("expected non-empty result")
	}

	if !strings.Contains(result, "sequenceDiagram") {
		t.Error("expected mermaid sequence diagram header")
	}

	if !strings.Contains(result, "participant client") {
		t.Error("expected client participant")
	}

	if !strings.Contains(result, "participant server") {
		t.Error("expected server participant")
	}

	if !strings.Contains(result, "request") {
		t.Error("expected request action")
	}
}

func TestTraceRenderer_RenderMermaidWithStateChanges(t *testing.T) {
	renderer := NewTraceRenderer()
	renderer.ShowStates = true
	trace := createTraceWithStateChanges()
	protocol := createTestProtocol()

	result := renderer.RenderMermaid(trace, protocol)

	if !strings.Contains(result, "State") {
		t.Error("expected state notes in output")
	}
}

func TestTraceRenderer_RenderMermaidWithSkippedStep(t *testing.T) {
	renderer := NewTraceRenderer()
	renderer.HighlightSkipped = true
	trace := createTraceWithSkippedStep()
	protocol := createTestProtocol()

	result := renderer.RenderMermaid(trace, protocol)

	if !strings.Contains(result, "SKIPPED") {
		t.Error("expected skipped indicator")
	}

	if !strings.Contains(result, "rect") {
		t.Error("expected rect block for skipped step")
	}
}

func TestTraceRenderer_RenderSVG(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTestTrace()
	protocol := createTestProtocol()

	result, err := renderer.RenderSVG(trace, protocol)
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty result")
	}

	if !strings.Contains(result, "<svg") {
		t.Error("expected SVG element")
	}

	if !strings.Contains(result, "Test Protocol") {
		t.Error("expected protocol name in SVG")
	}

	if !strings.Contains(result, "Client") {
		t.Error("expected client actor in SVG")
	}

	if !strings.Contains(result, "Server") {
		t.Error("expected server actor in SVG")
	}
}

func TestTraceRenderer_RenderSVGWithStateChanges(t *testing.T) {
	renderer := NewTraceRenderer()
	renderer.ShowStates = true
	trace := createTraceWithStateChanges()
	protocol := createTestProtocol()

	result, err := renderer.RenderSVG(trace, protocol)
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(result, "state-change") {
		t.Error("expected state change class in SVG")
	}
}

func TestTraceRenderer_RenderSVGWithSkippedStep(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTraceWithSkippedStep()
	protocol := createTestProtocol()

	result, err := renderer.RenderSVG(trace, protocol)
	if err != nil {
		t.Fatalf("RenderSVG failed: %v", err)
	}

	if !strings.Contains(result, "skipped") {
		t.Error("expected skipped class or text in SVG")
	}
}

func TestTraceRenderer_RenderJSON(t *testing.T) {
	renderer := NewTraceRenderer()
	trace := createTestTrace()

	result, err := renderer.RenderJSON(trace)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty JSON")
	}

	if !strings.Contains(string(result), "test-protocol") {
		t.Error("expected protocol ID in JSON")
	}

	if !strings.Contains(string(result), "steps") {
		t.Error("expected steps in JSON")
	}
}

func TestNewTraceRenderer(t *testing.T) {
	renderer := NewTraceRenderer()

	if renderer == nil {
		t.Error("expected non-nil renderer")
	}

	if !renderer.ShowStates {
		t.Error("expected ShowStates to be true by default")
	}

	if !renderer.HighlightSkipped {
		t.Error("expected HighlightSkipped to be true by default")
	}
}

func TestDefaultTraceTextOptions(t *testing.T) {
	opts := DefaultTraceTextOptions()

	if opts.ShowTimestamps {
		t.Error("expected ShowTimestamps to be false by default")
	}

	if !opts.ShowStates {
		t.Error("expected ShowStates to be true by default")
	}

	if opts.Compact {
		t.Error("expected Compact to be false by default")
	}
}
