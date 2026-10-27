package pidl

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const observationJSON = `{
  "format": "pidl-observation/v1",
  "source": {"tool": "aauth-sandbox", "requestFidelity": "wire",
             "from": "2026-01-01T00:00:00Z", "to": "2026-01-01T00:00:03Z"},
  "exchanges": [
    {
      "time": "2026-01-01T00:00:00.12Z",
      "from": "agent",
      "to": "ap",
      "request": {"method": "POST", "url": "https://ap.example.test/agent-token",
                  "headers": [["Content-Type", "application/json"]], "body": {"a": 1}},
      "response": {"status": 200, "headers": [["Content-Type", "application/json"]], "body": {"agent_token": "eyJ…"}},
      "label": "1. POST agent token",
      "annotations": [{"kind": "assertion", "text": "status 200", "ok": true}]
    }
  ]
}`

func TestParseObservation(t *testing.T) {
	o, err := ParseObservation([]byte(observationJSON))
	if err != nil {
		t.Fatal(err)
	}
	if errs := o.Validate(); len(errs) > 0 {
		t.Fatalf("valid observation reported: %v", errs)
	}
	if o.Source.RequestFidelity != RequestFidelityWire || len(o.Exchanges) != 1 {
		t.Fatalf("unexpected observation: %+v", o)
	}
	e := o.Exchanges[0]
	if e.From != "agent" || e.To != "ap" || e.Response.Status != 200 || e.Label != "1. POST agent token" {
		t.Errorf("unexpected exchange: %+v", e)
	}
	if e.Annotations[0].OK == nil || !*e.Annotations[0].OK {
		t.Errorf("annotation ok = %v", e.Annotations[0].OK)
	}
	if want := time.Date(2026, 1, 1, 0, 0, 0, 120e6, time.UTC); !e.Time.Equal(want) {
		t.Errorf("time = %v, want %v", e.Time, want)
	}
}

func TestParseObservationRefusesOtherVersion(t *testing.T) {
	_, err := ParseObservation([]byte(`{"format": "pidl-observation/v2", "source": {"tool": "x"}, "exchanges": []}`))
	if err == nil || !strings.Contains(err.Error(), "pidl-observation/v2") {
		t.Fatalf("err = %v, want a message naming the version found", err)
	}
}

func TestIsObservation(t *testing.T) {
	if !IsObservation([]byte(observationJSON)) {
		t.Error("observation not recognized")
	}
	if IsObservation([]byte(`{"format": "pidl-samples/v1"}`)) || IsObservation([]byte(`not json`)) {
		t.Error("non-observation recognized")
	}
}

func TestObservationRoundTrip(t *testing.T) {
	o, err := ParseObservation([]byte(observationJSON))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "run.observation.json")
	if err := o.WriteFile(path); err != nil {
		t.Fatal(err)
	}
	back, err := ReadObservation(path)
	if err != nil {
		t.Fatal(err)
	}
	a, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(back)
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Errorf("round trip changed the observation:\n%s\n%s", a, b)
	}
}

func TestObservationValidate(t *testing.T) {
	o := &Observation{
		Format: "pidl-observation/v1",
		Source: ObservationSource{RequestFidelity: "proxy"},
		Exchanges: []Exchange{{
			Request:     ObservedRequest{URL: "/relative", Headers: ObservedHeaders{{"", "x"}}, Body: json.RawMessage(`{`)},
			Response:    ObservedResponse{Status: 99},
			Annotations: []ExchangeAnnotation{{Text: "no kind"}},
		}},
	}
	from := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	to := from.Add(-time.Hour)
	o.Source.From, o.Source.To = &from, &to
	got := map[string]bool{}
	for _, e := range o.Validate() {
		got[e.Field] = true
	}
	for _, f := range []string{
		"source.tool", "source.requestFidelity", "source.to",
		"exchanges[0].request.method", "exchanges[0].request.url", "exchanges[0].response.status",
		"exchanges[0].request.headers[0]", "exchanges[0].request.body", "exchanges[0].annotations[0].kind",
	} {
		if !got[f] {
			t.Errorf("no error for %s; got %v", f, got)
		}
	}
}

func TestObservedHeadersGet(t *testing.T) {
	h := ObservedHeaders{{"Aauth-Requirement", "requirement=person-token"}, {"AAuth-Requirement", "second"}}
	if v, ok := h.Get("AAuth-Requirement"); !ok || v != "requirement=person-token" {
		t.Errorf("Get = %q, %v; want the first value, matched without regard to case", v, ok)
	}
	if _, ok := h.Get("Location"); ok {
		t.Error("Get found a header that is not there")
	}
}

func TestNewObservedBody(t *testing.T) {
	tests := []struct {
		name, mediaType, data string
		limit                 int
		want                  string
		truncated             bool
	}{
		{"json", "application/json; charset=utf-8", `{ "a" : [1, 2] }`, 0, `{"a":[1,2]}`, false},
		{"json suffix", "application/problem+json", `{"title":"x"}`, 0, `{"title":"x"}`, false},
		{"invalid json is text", "application/json", `{"a":`, 0, `"{\"a\":"`, false},
		{"text", "text/plain", "hello", 0, `"hello"`, false},
		{"no media type, utf-8", "", "plain", 0, `"plain"`, false},
		{"binary", "application/octet-stream", "\x00\x01\xff", 0, `{"base64":"AAH/","mediaType":"application/octet-stream"}`, false},
		{"truncated json is text", "application/json", `{"a":"bcdef"}`, 6, `"{\"a\":\""`, true},
		{"empty", "application/json", "", 0, ``, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, truncated, err := NewObservedBody([]byte(tt.data), tt.mediaType, tt.limit)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want || truncated != tt.truncated {
				t.Errorf("NewObservedBody = %s, %v; want %s, %v", got, truncated, tt.want, tt.truncated)
			}
			if len(got) > 0 && !json.Valid(got) {
				t.Errorf("result is not JSON: %s", got)
			}
		})
	}
}

func TestNewObservedBodyDefaultLimit(t *testing.T) {
	big := strings.Repeat("x", DefaultBodyLimit+10)
	got, truncated, err := NewObservedBody([]byte(big), "text/plain", 0)
	if err != nil {
		t.Fatal(err)
	}
	var s string
	if err := json.Unmarshal(got, &s); err != nil {
		t.Fatal(err)
	}
	if !truncated || len(s) != DefaultBodyLimit {
		t.Errorf("truncated = %v, length = %d; want true, %d", truncated, len(s), DefaultBodyLimit)
	}
}
