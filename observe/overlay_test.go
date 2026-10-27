package observe

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

var update = flag.Bool("update", false, "rewrite golden files")

const (
	aauthDocument  = "match/testdata/aauth_protocol.json"
	petstoreRun    = "adapter/sandbox/testdata/petstore.observation.json"
	petstoreBind   = "match/testdata/petstore.binding.json"
	observedGolden = "testdata/petstore.observed.json"
)

func petstore(t *testing.T) (*pidl.Protocol, pidl.Walkthrough, *pidl.Observation, *pidl.Binding) {
	t.Helper()
	p, err := pidl.ParseFile(aauthDocument)
	if err != nil {
		t.Fatal(err)
	}
	o, err := pidl.ReadObservation(petstoreRun)
	if err != nil {
		t.Fatal(err)
	}
	b, err := pidl.ReadBinding(petstoreBind)
	if err != nil {
		t.Fatal(err)
	}
	return p, p.Walkthrough(nil), o, b
}

func TestApplyPetstore(t *testing.T) {
	p, w, o, b := petstore(t)
	before := len(w.Scenarios)
	report, err := Apply(p, &w, o, b, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Instances) != 2 || len(w.Scenarios) != before+2 {
		t.Fatalf("instances = %d, scenarios %d → %d", len(report.Instances), before, len(w.Scenarios))
	}
	var observed []pidl.WalkthroughScenario
	for i, sc := range w.Scenarios {
		if sc.Position != i+1 {
			t.Errorf("%s at position %d, index %d", sc.ID, sc.Position, i)
		}
		if sc.ObservedRun == nil {
			continue
		}
		if prev := w.Scenarios[i-1].ID; prev != sc.ObservedRun.Scenario {
			t.Errorf("%s follows %s, want the scenario it is an instance of", sc.ID, prev)
		}
		observed = append(observed, sc)
	}
	browse, order := observed[0], observed[1]
	if browse.ID != "agent-identity.observed-1" || order.ID != "ps-authorization-challenge.observed-2" {
		t.Errorf("entry IDs = %s, %s", browse.ID, order.ID)
	}
	if order.Name != "Observed 2 of 2: PS Authorization, Challenged on the API Call" || order.VariantOf != "" || order.Fragment {
		t.Errorf("order entry = %q variant_of %q fragment %v", order.Name, order.VariantOf, order.Fragment)
	}
	run := order.ObservedRun
	if run.Source != "aauth-sandbox" || run.RequestFidelity != pidl.RequestFidelityWire || run.Instance != 2 || run.Instances != 2 ||
		run.RunnerUp != "ps-authorization" || run.Matched != 14 || run.Weak != 2 || run.Unobservable != 2 ||
		run.Explained != 7 || run.Extra != 7 || !run.Redacted || run.From == nil || run.To == nil || run.To.Before(*run.From) {
		t.Errorf("run summary = %+v", run)
	}
	for _, st := range order.Steps {
		switch st.ObservedStatus {
		case pidl.ObservedMatched:
			if st.Observed == nil || st.ObservedStrength == "" {
				t.Errorf("%s matched without an observed exchange or strength", st.ID)
			}
			if st.Example == nil {
				t.Errorf("%s lost its example", st.ID)
			}
		case pidl.ObservedUnobservable:
			if st.Observed != nil {
				t.Errorf("%s is unobservable but has an observed exchange", st.ID)
			}
		default:
			t.Errorf("%s = %s", st.ID, st.ObservedStatus)
		}
	}
	// A request step whose response is its own step shows only the request.
	if st := order.Steps[2]; st.ID != "request" || st.Observed.Request == nil || st.Observed.Response != nil {
		t.Errorf("request step observed = %+v", st.Observed)
	}
	if st := order.Steps[3]; st.Observed.Request != nil || st.Observed.Response == nil || st.Observed.Response.Status != 401 {
		t.Errorf("response step observed = %+v", st.Observed)
	}
	var after []int
	for _, e := range order.ObservedExtra {
		after = append(after, e.AfterStep)
	}
	if want := []int{4, 4, 6, 6, 8, 8, 16}; !equalInts(after, want) {
		t.Errorf("extra after steps %v, want %v: after each response step", after, want)
	}
	// The first extra is the person server fetching the agent provider's
	// metadata before it issues the person token.
	if e := order.ObservedExtra[0]; e.From != "ps" || e.To != "ap" || e.Observed.Request == nil || e.Observed.Response == nil {
		t.Errorf("first extra = %+v", e)
	}
}

func TestApplyPetstoreGolden(t *testing.T) {
	p, w, o, b := petstore(t)
	if _, err := Apply(p, &w, o, b, Options{}); err != nil {
		t.Fatal(err)
	}
	var observed []pidl.WalkthroughScenario
	for _, sc := range w.Scenarios {
		if sc.ObservedRun != nil {
			observed = append(observed, sc)
		}
	}
	got, err := json.MarshalIndent(observed, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	if *update {
		if err := os.MkdirAll(filepath.Dir(observedGolden), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(observedGolden, got, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(observedGolden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("observed entries differ from %s; run go test -update to rewrite it after checking the change", observedGolden)
	}
}

func TestApplyNoInstances(t *testing.T) {
	p, w, _, b := petstore(t)
	before := len(w.Scenarios)
	empty := &pidl.Observation{Format: pidl.ObservationFormat, Source: pidl.ObservationSource{Tool: "test"}}
	report, err := Apply(p, &w, empty, b, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Instances) != 0 || len(w.Scenarios) != before {
		t.Errorf("an empty observation changed the walkthrough: %d instances, %d scenarios", len(report.Instances), len(w.Scenarios))
	}
	if _, err := Apply(p, &w, empty, b, Options{Scenarios: []string{"nope"}}); err == nil {
		t.Error("an unknown scenario was accepted")
	}
}

// jwt builds a compact JWT with the given claims. Its signature is a
// placeholder: decoding never verifies one.
func jwt(t *testing.T, claims map[string]any) string {
	t.Helper()
	enc := func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return enc(map[string]string{"alg": "EdDSA", "typ": "aa-agent+jwt"}) + "." + enc(claims) + ".c2lnbmF0dXJl"
}

//nolint:gosec // G101: Test fixture with fake credentials
const tokenDocument = `{
  "protocol": {"id": "tok", "name": "Tokens", "version": "1.0"},
  "entities": [{"id": "agent", "name": "Agent", "type": "agent"}, {"id": "ap", "name": "AP", "type": "server"}],
  "metadata": {"tokens": [
    {"id": "agent_token", "name": "Agent token", "type": "jwt", "carried_in": [
      {"in": "header", "name": "Signature-Key", "param": "jwt", "typ": "aa-agent+jwt"},
      {"in": "body", "path": "/agent_token"}
    ]},
    {"id": "person_token", "name": "Person token", "type": "jwt", "carried_in": [
      {"in": "header", "name": "Signature-Key", "param": "jwt", "typ": "aa-person+jwt"}
    ]},
    {"id": "api_key", "name": "API key", "type": "opaque", "carried_in": [{"in": "header", "name": "Authorization"}]}
  ]},
  "scenarios": [{"id": "enroll", "name": "Enroll", "flows": [
    {"id": "enroll", "from": "agent", "to": "ap", "action": "enroll", "label": "Enroll",
     "example": {"source": "test", "kind": "derived", "request": {"method": "POST", "url": "https://ap.example/enroll"}, "response": {"status": 200}}}
  ]}]
}`

func tokenRun(t *testing.T) (*pidl.Protocol, *pidl.Observation, string, string) {
	t.Helper()
	p, err := pidl.Parse([]byte(tokenDocument))
	if err != nil {
		t.Fatal(err)
	}
	if errs := p.Validate(); len(errs) > 0 {
		t.Fatalf("token document does not validate: %v", errs)
	}
	request := jwt(t, map[string]any{"sub": "aauth:assistant@ap.example", "cnf": map[string]string{"jkt": "abc"}})
	issued := jwt(t, map[string]any{"sub": "aauth:assistant@ap.example", "iss": "https://ap.example"})
	body, _, err := pidl.NewObservedBody([]byte(`{"agent_token":"`+issued+`","expires_in":300}`), "application/json", 0)
	if err != nil {
		t.Fatal(err)
	}
	o := &pidl.Observation{Format: pidl.ObservationFormat, Source: pidl.ObservationSource{Tool: "test"}, Exchanges: []pidl.Exchange{{
		From: "agent", To: "ap",
		Request: pidl.ObservedRequest{Method: "POST", URL: "https://ap.example/enroll", Headers: pidl.ObservedHeaders{
			{"Signature-Key", `sig=jwt;jwt="` + request + `"`},
			{"Signature", "sig=:AAAAAAAAAAAAAAAAAAAAAA==:"},
			{"Authorization", "Bearer opaque-key-123"},
		}},
		Response: pidl.ObservedResponse{Status: 200, Headers: pidl.ObservedHeaders{{"Content-Type", "application/json"}}, Body: body},
	}}}
	return p, o, request, issued
}

func TestApplyDecodesAndMasksTokens(t *testing.T) {
	p, o, request, issued := tokenRun(t)
	w := p.Walkthrough(nil)
	if _, err := Apply(p, &w, o, nil, Options{}); err != nil {
		t.Fatal(err)
	}
	step := w.Scenarios[1].Steps[0]
	m := step.Observed
	if m == nil || m.Request == nil || m.Response == nil {
		t.Fatalf("observed = %+v, want request and response on a step with no response step", m)
	}
	var got []string
	for _, tok := range m.Tokens {
		got = append(got, tok.Token+": "+tok.Label)
		if tok.Token == "agent_token" && !strings.Contains(string(tok.Payload), `"sub":"aauth:assistant@ap.example"`) {
			t.Errorf("agent_token claims = %s", tok.Payload)
		}
	}
	want := []string{
		"agent_token: Signature-Key header (jwt) in the request; decoded, not verified",
		"api_key: Authorization header in the request; decoded, not verified",
		"agent_token: /agent_token in the body in the response; decoded, not verified",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("tokens:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	headers := map[string]string{}
	for _, h := range m.Request.Headers {
		headers[h.Name] = h.Value
	}
	if headers["Signature-Key"] != `sig=jwt;jwt="[redacted:jwt]"` || headers["Signature"] != "sig=:[redacted]:" || headers["Authorization"] != "Bearer [redacted]" {
		t.Errorf("masked headers = %v", headers)
	}
	out, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{request, issued, "opaque-key-123", "AAAAAAAAAAAAAAAAAAAAAA"} {
		if strings.Contains(string(out), raw) {
			t.Errorf("a raw value survived redaction: %.24s…", raw)
		}
	}
	if !json.Valid(m.Response.Body) || !strings.Contains(string(m.Response.Body), `"expires_in":300`) {
		t.Errorf("masked body = %s", m.Response.Body)
	}
	if !w.Scenarios[1].ObservedRun.Redacted {
		t.Error("the run does not say it was redacted")
	}
}

func TestApplyRaw(t *testing.T) {
	p, o, request, issued := tokenRun(t)
	w := p.Walkthrough(nil)
	if _, err := Apply(p, &w, o, nil, Options{Raw: true}); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{request, issued, "opaque-key-123"} {
		if !strings.Contains(string(out), raw) {
			t.Errorf("--raw dropped a value: %.24s…", raw)
		}
	}
	if w.Scenarios[1].ObservedRun.Redacted {
		t.Error("a raw run says it was redacted")
	}
}

func TestParam(t *testing.T) {
	tests := []struct {
		value, name, want string
		ok                bool
	}{
		{`sig=jwt;jwt="eyJ.a.b"`, "jwt", "eyJ.a.b", true},
		{`requirement=auth-token; resource-token="eyJ.r.s"`, "resource-token", "eyJ.r.s", true},
		{`requirement=person-token`, "requirement", "person-token", true},
		{`sig=jwt`, "jwt", "", false},
		{`a="x\"y"`, "a", `x\"y`, true},
	}
	for _, tt := range tests {
		got, ok := param(tt.value, tt.name)
		if got != tt.want || ok != tt.ok {
			t.Errorf("param(%q, %q) = %q, %v; want %q, %v", tt.value, tt.name, got, ok, tt.want, tt.ok)
		}
	}
}

func TestDecodeJWT(t *testing.T) {
	tok := jwt(t, map[string]any{"a": 1})
	if h, p, ok := decodeJWT(tok); !ok || !strings.Contains(string(h), "EdDSA") || string(p) != `{"a":1}` {
		t.Errorf("decodeJWT = %s %s %v", h, p, ok)
	}
	for _, bad := range []string{"eyJ…", "a.b", "e30.!!!.x", "opaque"} {
		if _, _, ok := decodeJWT(bad); ok {
			t.Errorf("decodeJWT(%q) succeeded", bad)
		}
	}
}

func TestMaskHeader(t *testing.T) {
	for _, tt := range []struct{ name, value, want string }{
		{"Signature", "sig=:abc:", "sig=:[redacted]:"},
		{"authorization", "Basic dXNlcjpwYXNz", "Basic [redacted]"},
		{"Authorization", "token", "[redacted]"},
		{"Set-Cookie", "session=abc; Path=/", "[redacted]"},
		{"Content-Type", "application/json", "application/json"},
	} {
		if got := maskHeader(tt.name, tt.value); got != tt.want {
			t.Errorf("maskHeader(%s, %q) = %q, want %q", tt.name, tt.value, got, tt.want)
		}
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
