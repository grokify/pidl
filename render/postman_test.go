package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/examples"
)

// apiProtocol is the generic worked example, which has every feature the
// collection renders: browser steps it does not send, chained tokens, a
// challenge parameter, a deferred response and a signed request.
func apiProtocol(t *testing.T) *pidl.Protocol {
	t.Helper()
	p, err := examples.GetProtocol("oauth2_api_access")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// folder finds a scenario folder by name prefix, inside group folders.
func folder(t *testing.T, c *PostmanCollection, prefix string) PostmanItem {
	t.Helper()
	var find func(items []PostmanItem) *PostmanItem
	find = func(items []PostmanItem) *PostmanItem {
		for i, f := range items {
			if f.Request != nil {
				continue
			}
			if strings.HasPrefix(f.Name, prefix) {
				return &items[i]
			}
			if got := find(f.Item); got != nil {
				return got
			}
		}
		return nil
	}
	if f := find(c.Item); f != nil {
		return *f
	}
	t.Fatalf("no folder %q", prefix)
	return PostmanItem{}
}

func request(t *testing.T, f PostmanItem, idSuffix string) PostmanItem {
	t.Helper()
	for _, it := range f.Item {
		if strings.HasSuffix(it.ID, "/"+idSuffix) {
			return it
		}
	}
	t.Fatalf("no request %q in %s", idSuffix, f.Name)
	return PostmanItem{}
}

func header(it PostmanItem, key string) string {
	for _, h := range it.Request.Header {
		if h.Key == key {
			return h.Value
		}
	}
	return ""
}

func script(it PostmanItem) string {
	var lines []string
	for _, e := range it.Event {
		lines = append(lines, e.Script.Exec...)
	}
	return strings.Join(lines, "\n")
}

func TestPostmanCollection(t *testing.T) {
	c, err := NewPostman(PostmanOptions{Signer: SignerProxy}).Collection(apiProtocol(t))
	if err != nil {
		t.Fatal(err)
	}
	if c.Info.Schema != PostmanSchema || !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(c.Info.PostmanID) {
		t.Fatalf("info = %+v", c.Info)
	}
	if len(c.Item) != 2 || c.Item[0].Name != "User-delegated access" || c.Item[1].Name != "Service access" {
		t.Fatalf("top-level folders = %d, want the two groups in order", len(c.Item))
	}

	// Steps between other roles are listed, not sent.
	code := folder(t, c, "Authorization code")
	if !strings.Contains(code.Description, "not sent by this collection") || !strings.Contains(code.Description, "GET /authorize") {
		t.Errorf("authorization code description should list the browser steps:\n%s", code.Description)
	}
	if len(code.Item) != 2 {
		t.Errorf("authorization code sends %d requests, want the token request and the API call", len(code.Item))
	}
	token := request(t, code, "token-request")
	if token.Request.URL != "{{as_url}}/token" || !strings.Contains(token.Request.Body.Raw, "code_verifier=") {
		t.Errorf("token request = %+v", token.Request)
	}
	s := script(token)
	for _, want := range []string{`pm.collectionVariables.set("access_token", pm.response.json()["access_token"]);`, "access_token claims", `"scope","sub"`} {
		if !strings.Contains(s, want) {
			t.Errorf("token request script missing %q:\n%s", want, s)
		}
	}
	if got := header(request(t, code, "api-request"), "Authorization"); got != "Bearer {{access_token}}" {
		t.Errorf("API request Authorization = %q", got)
	}

	// A parameter of a challenge flows into the next request.
	challenge := folder(t, c, "Client credentials after")
	if s := script(request(t, challenge, "first-request")); !strings.Contains(s, `pm.collectionVariables.set("scope", ((pm.response.headers.get("WWW-Authenticate") || "").match(/(?:^|[;,\s])\s*scope="([^"]*)"/) || [])[1]);`) {
		t.Errorf("challenge does not capture the scope:\n%s", s)
	}
	if body := request(t, challenge, "token.request").Request.Body.Raw; body != "grant_type=client_credentials&scope={{scope}}" {
		t.Errorf("token request body = %q, want the captured scope", body)
	}

	// A deferred response is polled, or the poll is skipped.
	export := folder(t, c, "Asynchronous export")
	s = script(request(t, export, "export-request"))
	for _, want := range []string{"if (pm.response.code === 200) {", `pm.collectionVariables.set("location", pm.response.headers.get("Location"));`, "pm.test('status 202'"} {
		if !strings.Contains(s, want) {
			t.Errorf("deferred request script missing %q:\n%s", want, s)
		}
	}
	poll := request(t, export, "export-poll")
	if poll.Request.URL != "{{location}}" {
		t.Errorf("poll URL = %q", poll.Request.URL)
	}
	if s := script(poll); !strings.Contains(s, "if (pm.response.code === 202) {") || !strings.Contains(s, `setNextRequest("async-export/export-poll")`) {
		t.Errorf("poll script:\n%s", s)
	}
	if !strings.Contains(script(export), `"async-export.api_url"`) {
		t.Error("export folder lacks the per-scenario role override")
	}

	// A signed request goes through the signing proxy, without the
	// signature headers it cannot create.
	signed := request(t, folder(t, c, "Signed request"), "post-message")
	if signed.Request.URL != "{{signer_url}}/{{api_url}}/messages" {
		t.Errorf("signed request URL = %q", signed.Request.URL)
	}
	for _, h := range []string{"Signature", "Signature-Input", "Content-Digest", "Signature-Key"} {
		if header(signed, h) != "" {
			t.Errorf("signed request sends %s", h)
		}
	}
	if len(token.Response) != 0 {
		t.Error("a live collection should not carry saved examples")
	}
}

func TestPostmanMock(t *testing.T) {
	c, err := NewPostman(PostmanOptions{Mock: true}).Collection(apiProtocol(t))
	if err != nil {
		t.Fatal(err)
	}
	authorize := request(t, folder(t, c, "Signed request"), "post-message")
	if authorize.Request.URL != "{{api_url}}/messages" || header(authorize, "Signature") == "" {
		t.Errorf("mock request should be direct and keep its example headers: %+v", authorize.Request)
	}
	if header(authorize, "x-mock-response-name") != authorize.Name || len(authorize.Response) != 1 || authorize.Response[0].Code != 201 {
		t.Fatalf("mock example = %+v", authorize.Response)
	}
	if s := script(authorize); strings.Contains(s, "body has") {
		t.Errorf("mock checks should be status only:\n%s", s)
	}
	for _, v := range c.Variable {
		if v.Key == "signer_url" {
			t.Error("mock collection has a signer")
		}
	}
}

func TestPostmanJSON(t *testing.T) {
	out, err := NewPostman(PostmanOptions{Signer: SignerProxy}).RenderString(apiProtocol(t))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, `\u0026`) || strings.Contains(out, `\u003c`) {
		t.Error("JSON escapes HTML characters")
	}
	again, err := NewPostman(PostmanOptions{Signer: SignerProxy}).RenderString(apiProtocol(t))
	if err != nil {
		t.Fatal(err)
	}
	if out != again {
		t.Error("collection output is not deterministic")
	}
	var any map[string]any
	if err := json.Unmarshal([]byte(out), &any); err != nil {
		t.Fatal(err)
	}
}

func TestPostmanNoClient(t *testing.T) {
	p := apiProtocol(t)
	if _, err := NewPostman(PostmanOptions{Client: "nobody"}).Collection(p); err != nil {
		t.Fatalf("an unknown client yields empty folders, not an error: %v", err)
	}
	q := &pidl.Protocol{ProtocolMeta: pidl.ProtocolMeta{ID: "x", Name: "X"}, Entities: []pidl.Entity{{ID: "a", Name: "A", Type: pidl.EntityTypeResourceServer}}}
	if _, err := NewPostman(PostmanOptions{}).Collection(q); err == nil {
		t.Error("a protocol without an agent or client role should be an error")
	}
}

func TestPostmanEnvironment(t *testing.T) {
	p := apiProtocol(t)
	path := filepath.Join(t.TempDir(), "targets.json")
	targets := `{
  "local": {
    "signer": "http://127.0.0.1:8099",
    "roles": {"as": "https://as.local.test", "api": "https://api.local.test"},
    "scenarios": {"async-export": {"api": "https://exports.local.test"}}
  },
  "bad": {"roles": {"nobody": "https://x.test"}}
}`
	if err := os.WriteFile(path, []byte(targets), 0o600); err != nil {
		t.Fatal(err)
	}
	ts, err := ReadPostmanTargets(path)
	if err != nil {
		t.Fatal(err)
	}
	local := ts["local"]
	out, err := (&PostmanEnvironmentRenderer{Name: "local", Target: &local}).RenderString(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"key": "as_url"`, `"value": "https://as.local.test"`, `"key": "async-export.api_url"`, `"value": "https://exports.local.test"`, `"key": "signer_url"`, `"_postman_variable_scope": "environment"`} {
		if !strings.Contains(out, want) {
			t.Errorf("environment missing %s:\n%s", want, out)
		}
	}
	bad := ts["bad"]
	if _, err := (&PostmanEnvironmentRenderer{Name: "bad", Target: &bad}).RenderString(p); err == nil || !strings.Contains(err.Error(), `unknown role "nobody"`) {
		t.Errorf("bad target error = %v", err)
	}
	def, err := (&PostmanEnvironmentRenderer{}).RenderString(p)
	if err != nil || !strings.Contains(def, "https://as.example") {
		t.Errorf("default environment = %v\n%s", err, def)
	}
}

func TestPostmanHelpers(t *testing.T) {
	if got := replaceOrigin(`{"r":"https://a.example","s":"https://a.example.org","t":"https://a.example/x"}`, "https://a.example", "{{a_url}}"); got != `{"r":"{{a_url}}","s":"https://a.example.org","t":"{{a_url}}/x"}` {
		t.Errorf("replaceOrigin = %s", got)
	}
	if got := pointerExpr("/a~1b/0/c~0d"); got != `["a/b"][0]["c~d"]` {
		t.Errorf("pointerExpr = %s", got)
	}
	if nameUUID("x") != nameUUID("x") || nameUUID("x") == nameUUID("y") {
		t.Error("nameUUID is not deterministic and distinct")
	}
	if got := sanitizeVar("session-token"); got != "session_token" {
		t.Errorf("sanitizeVar = %q", got)
	}
	if got := nextRequest(""); !strings.Contains(got, "setNextRequest(null)") {
		t.Errorf("nextRequest(\"\") = %q", got)
	}
}

func TestHeaderParam(t *testing.T) {
	v := `requirement=auth-token; resource-token="eyJ.a\"b"; code="X"`
	if got := headerParam(v, "resource-token"); got != `eyJ.a\"b` {
		t.Errorf("resource-token = %q", got)
	}
	if got := headerParam(v, "code"); got != "X" {
		t.Errorf("code = %q", got)
	}
	if got := headerParam(v, "url"); got != "" {
		t.Errorf("url = %q", got)
	}
	challenge := `Bearer realm="api.example", error="insufficient_scope", scope="reports.monthly.read"`
	if got := headerParam(challenge, "realm"); got != "api.example" {
		t.Errorf("realm = %q", got)
	}
	if got := headerParam(challenge, "scope"); got != "reports.monthly.read" {
		t.Errorf("scope = %q", got)
	}
}

func TestProxySignatureKey(t *testing.T) {
	req := func(headers ...pidl.ExampleHeader) *pidl.ExampleRequest {
		return &pidl.ExampleRequest{Method: "GET", URL: "https://api.example/", Headers: headers}
	}
	signed := []pidl.ExampleHeader{{Name: "Signature-Input", Value: `sig=("@method")`}, {Name: "Signature", Value: "sig=:x:"}}
	tests := []struct {
		name  string
		req   *pidl.ExampleRequest
		token string
		want  string
	}{
		{"jwt with a captured token", req(append(signed, pidl.ExampleHeader{Name: "Signature-Key", Value: `sig=jwt;jwt="eyJ..."`})...), "user_token", `sig=jwt;jwt="{{user_token}}"`},
		{"jwt without one", req(append(signed, pidl.ExampleHeader{Name: "Signature-Key", Value: `sig=jwt;jwt="eyJ..."`})...), "", "sig=jwt"},
		{"another label and scheme", req(append(signed, pidl.ExampleHeader{Name: "Signature-Key", Value: `key1=hwk;kty="OKP";x="abc"`})...), "user_token", "key1=hwk"},
		{"no Signature-Key", req(signed...), "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proxySignatureKey(tt.req, tt.token); got != tt.want {
				t.Errorf("proxySignatureKey() = %q, want %q", got, tt.want)
			}
		})
	}
	if !isSigned(req(signed...)) || isSigned(req(pidl.ExampleHeader{Name: "Authorization", Value: "Bearer x"})) {
		t.Error("isSigned is wrong")
	}
}

func TestPostmanProxySendsUnsignedRequestsDirectly(t *testing.T) {
	c, err := NewPostman(PostmanOptions{Signer: SignerProxy}).Collection(apiProtocol(t))
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "X-Sandbox") {
		t.Error("the collection names a particular signer's headers")
	}
	for _, f := range c.Item {
		for _, it := range f.Item {
			if it.Request == nil || strings.HasPrefix(it.Request.URL, "{{signer_url}}/") {
				continue
			}
			for _, h := range it.Request.Header {
				if strings.EqualFold(h.Key, "Signature") || strings.EqualFold(h.Key, "Signature-Input") {
					t.Errorf("%s is signed but sent directly", it.ID)
				}
			}
		}
	}
}
