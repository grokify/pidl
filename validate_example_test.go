package pidl

import (
	"encoding/json"
	"strings"
	"testing"
)

func exampleProtocol() *Protocol {
	p := verificationProtocol()
	p.Scenarios[2].Flows[1].Example = &FlowExample{
		Source: "draft-11 §7.2.1",
		Kind:   ExampleKindDraft,
		Request: &ExampleRequest{
			Method:  "POST",
			URL:     "https://ps.example/auth_token",
			Headers: []ExampleHeader{{Name: "Content-Type", Value: "application/json"}, {Name: "Signature-Key", Value: `sig=jwt;jwt="eyJ..."`}},
			Body:    json.RawMessage(`{"resource_token":"eyJ...","presented_token":"eyJ..."}`),
		},
		Response: &ExampleResponse{Status: 200, Body: json.RawMessage(`{"auth_token":"eyJ...","expires_in":3600}`)},
		Tokens: []ExampleToken{{
			Token: "resource_token", Label: "request body",
			Header:  json.RawMessage(`{"typ":"aa-resource+jwt","alg":"EdDSA"}`),
			Payload: json.RawMessage(`{"iss":"https://resource.example","aud":"https://ps.example","agent_jkt":"NzbL..."}`),
		}},
	}
	return p
}

func TestValidateExamples(t *testing.T) {
	if errs := exampleProtocol().Validate(); len(errs) > 0 {
		t.Fatalf("Validate() = %v, want no errors", errs)
	}

	const f = "scenarios[2].flows[1].example"
	tests := []struct {
		name      string
		mutate    func(ex *FlowExample)
		wantField string
		wantMsg   string
	}{
		{"missing source", func(ex *FlowExample) { ex.Source = "" }, f + ".source", "required"},
		{"missing kind", func(ex *FlowExample) { ex.Kind = "" }, f + ".kind", "required"},
		{"bad kind", func(ex *FlowExample) { ex.Kind = "mock" }, f + ".kind", `invalid example kind "mock"`},
		{"empty example", func(ex *FlowExample) { ex.Request, ex.Response, ex.Tokens = nil, nil, nil }, f, "needs a request"},
		{"missing method", func(ex *FlowExample) { ex.Request.Method = "" }, f + ".request.method", "required"},
		{"missing url", func(ex *FlowExample) { ex.Request.URL = "" }, f + ".request.url", "required"},
		{"unnamed header", func(ex *FlowExample) { ex.Request.Headers[1].Name = "" }, f + ".request.headers[1].name", "required"},
		{"bad body", func(ex *FlowExample) { ex.Request.Body = json.RawMessage(`{`) }, f + ".request.body", "valid JSON"},
		{"bad status", func(ex *FlowExample) { ex.Response.Status = 42 }, f + ".response.status", "invalid HTTP status 42"},
		{"unknown token", func(ex *FlowExample) { ex.Tokens[0].Token = "nope" }, f + ".tokens[0].token", `unknown token definition "nope"`},
		{"bad payload", func(ex *FlowExample) { ex.Tokens[0].Payload = json.RawMessage(`[`) }, f + ".tokens[0].payload", "valid JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := exampleProtocol()
			tt.mutate(p.Scenarios[2].Flows[1].Example)
			for _, e := range p.Validate() {
				if e.Field == tt.wantField && strings.Contains(e.Message, tt.wantMsg) {
					return
				}
			}
			t.Fatalf("Validate() = %v, want %s: ...%s...", p.Validate(), tt.wantField, tt.wantMsg)
		})
	}
}

func TestExampleOnIncludeStepRejected(t *testing.T) {
	p := exampleProtocol()
	p.Scenarios[2].Flows[2].Example = &FlowExample{Source: "x", Kind: ExampleKindDerived, Response: &ExampleResponse{Status: 200}}
	for _, e := range p.Validate() {
		if e.Field == "scenarios[2].flows[2]" && strings.Contains(e.Message, "example") {
			return
		}
	}
	t.Fatal("include step with an example was not rejected")
}

func TestExampleRoundTripKeepsOrder(t *testing.T) {
	data, err := exampleProtocol().ToJSONCompact()
	if err != nil {
		t.Fatalf("ToJSONCompact: %v", err)
	}
	for _, want := range []string{
		`"body":{"resource_token":"eyJ...","presented_token":"eyJ..."}`,
		`"payload":{"iss":"https://resource.example","aud":"https://ps.example","agent_jkt":"NzbL..."}`,
		`"headers":[{"name":"Content-Type","value":"application/json"},{"name":"Signature-Key"`,
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("JSON lost order or content; missing %s", want)
		}
	}
	back, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if errs := back.Validate(); len(errs) > 0 {
		t.Fatalf("round trip does not validate: %v", errs)
	}
}

func TestCompareReportsExampleChange(t *testing.T) {
	base, next := exampleProtocol(), exampleProtocol()
	next.Scenarios[2].Flows[1].Example.Response.Status = 202
	for _, it := range Compare(base, next, DefaultDiffOptions()).Items {
		if it.Path == "scenarios[three-party].flows[1].example" {
			return
		}
	}
	t.Fatal("example change not reported")
}
