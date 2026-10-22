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

func aauthProtocol(t *testing.T) *pidl.Protocol {
	t.Helper()
	p, err := examples.GetProtocol("aauth_protocol")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func folder(t *testing.T, c *PostmanCollection, prefix string) PostmanItem {
	t.Helper()
	for _, f := range c.Item {
		if strings.HasPrefix(f.Name, prefix) {
			return f
		}
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

func TestPostmanSigned(t *testing.T) {
	c, err := NewPostman(PostmanOptions{Signer: SignerProxy}).Collection(aauthProtocol(t))
	if err != nil {
		t.Fatal(err)
	}
	if c.Info.Schema != PostmanSchema || !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(c.Info.PostmanID) {
		t.Fatalf("info = %+v", c.Info)
	}
	if len(c.Item) != 5 {
		t.Fatalf("folders = %d, want the five access modes", len(c.Item))
	}

	fed := folder(t, c, "Federated")
	if !strings.Contains(fed.Description, "not sent by this collection") || !strings.Contains(fed.Description, "Access Server") {
		t.Errorf("federated description should list the PS-to-AS steps:\n%s", fed.Description)
	}
	if !strings.Contains(script(fed), `"federated-authorization.resource_url"`) {
		t.Error("federated folder lacks the per-scenario role override")
	}

	attest := request(t, fed, "agent-token.attest")
	if attest.Request.URL != "{{signer_url}}/{{ap_url}}/agent-token" || header(attest, "Signature-Key") != "sig=jkt-jwt" || header(attest, "Signature") != "" {
		t.Errorf("attest = %+v", attest.Request)
	}
	if !strings.Contains(script(attest), `pm.collectionVariables.set("agent_token", pm.response.json()["agent_token"]);`) {
		t.Errorf("attest does not capture the agent token:\n%s", script(attest))
	}

	person := request(t, fed, "person-token.request")
	if !strings.Contains(person.Request.Body.Raw, `"resource": "{{resource_url}}"`) {
		t.Errorf("person request body should name the resource by variable:\n%s", person.Request.Body.Raw)
	}

	authorize := request(t, fed, "authorize")
	if got := header(authorize, "Signature-Key"); got != `sig=jwt;jwt="{{person_token}}"` {
		t.Errorf("authorize presents %q", got)
	}

	token := request(t, fed, "auth-token-request")
	if !strings.Contains(token.Request.Body.Raw, `"resource_token": "{{resource_token}}"`) || !strings.Contains(token.Request.Body.Raw, `"presented_token": "{{person_token}}"`) {
		t.Errorf("token request body does not chain tokens:\n%s", token.Request.Body.Raw)
	}
	s := script(token)
	for _, want := range []string{"if (pm.response.code === 200) {", `setNextRequest("federated-authorization/api-request")`, `pm.collectionVariables.set("location", pm.response.headers.get("Location"));`, "pm.test('status 202'"} {
		if !strings.Contains(s, want) {
			t.Errorf("deferred token request script missing %q:\n%s", want, s)
		}
	}

	poll := request(t, fed, "consent.poll")
	if poll.Request.URL != "{{signer_url}}/{{location}}" {
		t.Errorf("poll URL = %q", poll.Request.URL)
	}
	s = script(poll)
	for _, want := range []string{"if (pm.response.code === 202) {", `setNextRequest("federated-authorization/consent.poll")`, `pm.collectionVariables.set("auth_token"`, "auth_token claims"} {
		if !strings.Contains(s, want) {
			t.Errorf("poll script missing %q:\n%s", want, s)
		}
	}

	api := request(t, fed, "api-request")
	if got := header(api, "Signature-Key"); got != `sig=jwt;jwt="{{auth_token}}"` {
		t.Errorf("api request presents %q", got)
	}

	rm := folder(t, c, "Resource-managed")
	session := request(t, rm, "session-request")
	if header(session, "Authorization") != "AAuth {{aauth_access}}" {
		t.Errorf("session request Authorization = %q", header(session, "Authorization"))
	}
	if len(attest.Response) != 0 {
		t.Error("signed collection should not carry saved examples")
	}
}

func TestPostmanMock(t *testing.T) {
	c, err := NewPostman(PostmanOptions{Mock: true}).Collection(aauthProtocol(t))
	if err != nil {
		t.Fatal(err)
	}
	ps := folder(t, c, "PS authorization")
	authorize := request(t, ps, "authorize")
	if authorize.Request.URL != "{{resource_url}}/authorize" || header(authorize, "Signature-Key") == "" {
		t.Errorf("mock request should be direct and keep its example headers: %+v", authorize.Request)
	}
	if header(authorize, "x-mock-response-name") != authorize.Name || len(authorize.Response) != 1 || authorize.Response[0].Code != 200 {
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
	out, err := NewPostman(PostmanOptions{Signer: SignerProxy}).RenderString(aauthProtocol(t))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, `\u0026`) || strings.Contains(out, `\u003c`) {
		t.Error("JSON escapes HTML characters")
	}
	again, err := NewPostman(PostmanOptions{Signer: SignerProxy}).RenderString(aauthProtocol(t))
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
	p := aauthProtocol(t)
	if _, err := NewPostman(PostmanOptions{Client: "nobody"}).Collection(p); err != nil {
		t.Fatalf("an unknown client yields empty folders, not an error: %v", err)
	}
	q := &pidl.Protocol{ProtocolMeta: pidl.ProtocolMeta{ID: "x", Name: "X"}, Entities: []pidl.Entity{{ID: "a", Name: "A", Type: pidl.EntityTypeResourceServer}}}
	if _, err := NewPostman(PostmanOptions{}).Collection(q); err == nil {
		t.Error("a protocol without an agent or client role should be an error")
	}
}

func TestPostmanEnvironment(t *testing.T) {
	p := aauthProtocol(t)
	path := filepath.Join(t.TempDir(), "targets.json")
	targets := `{
  "sandbox": {
    "signer": "http://127.0.0.1:8099",
    "roles": {"ap": "https://ap.aauth.test", "ps": "https://ps.aauth.test", "resource": "https://resource.aauth.test", "as": "https://as.aauth.test"},
    "scenarios": {"federated-authorization": {"resource": "https://docs.aauth.test"}}
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
	sandbox := ts["sandbox"]
	out, err := (&PostmanEnvironmentRenderer{Name: "sandbox", Target: &sandbox}).RenderString(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"key": "ps_url"`, `"value": "https://ps.aauth.test"`, `"key": "federated-authorization.resource_url"`, `"value": "https://docs.aauth.test"`, `"key": "signer_url"`, `"_postman_variable_scope": "environment"`} {
		if !strings.Contains(out, want) {
			t.Errorf("environment missing %s:\n%s", want, out)
		}
	}
	bad := ts["bad"]
	if _, err := (&PostmanEnvironmentRenderer{Name: "bad", Target: &bad}).RenderString(p); err == nil || !strings.Contains(err.Error(), `unknown role "nobody"`) {
		t.Errorf("bad target error = %v", err)
	}
	def, err := (&PostmanEnvironmentRenderer{}).RenderString(p)
	if err != nil || !strings.Contains(def, "https://ps.example") {
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
	if got := sanitizeVar("aauth-access"); got != "aauth_access" {
		t.Errorf("sanitizeVar = %q", got)
	}
	if got := nextRequest(""); !strings.Contains(got, "setNextRequest(null)") {
		t.Errorf("nextRequest(\"\") = %q", got)
	}
}

func TestPostmanPetstoreChallenges(t *testing.T) {
	p, err := examples.GetProtocol("aauth_petstore")
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewPostman(PostmanOptions{Signer: SignerProxy}).Collection(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Item) != 2 {
		t.Fatalf("folders = %d, want browse and order", len(c.Item))
	}
	order := folder(t, c, "Order a pet")

	// The resource token arrives as a parameter of AAuth-Requirement.
	asPerson := request(t, order, "order-as-person")
	if got := header(asPerson, "Signature-Key"); got != `sig=jwt;jwt="{{person_token}}"` {
		t.Errorf("order as person presents %q", got)
	}
	if s := script(asPerson); !strings.Contains(s, `pm.collectionVariables.set("resource_token", ((pm.response.headers.get("AAuth-Requirement") || "").match(/(?:^|;)\s*resource-token="([^"]*)"/) || [])[1]);`) {
		t.Errorf("challenge does not capture the resource token:\n%s", s)
	}
	token := request(t, order, "auth-token-request")
	if !strings.Contains(token.Request.Body.Raw, `"resource_token": "{{resource_token}}"`) {
		t.Errorf("token request does not use the captured resource token:\n%s", token.Request.Body.Raw)
	}
	withAuth := request(t, order, "order-with-auth")
	if got := header(withAuth, "Signature-Key"); got != `sig=jwt;jwt="{{auth_token}}"` {
		t.Errorf("order with auth presents %q", got)
	}
	// The person token is requested for the Petstore, by role URL.
	person := request(t, order, "person-token.request")
	if !strings.Contains(person.Request.Body.Raw, `"resource": "{{resource_url}}"`) {
		t.Errorf("person token request body:\n%s", person.Request.Body.Raw)
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
	c, err := NewPostman(PostmanOptions{Signer: SignerProxy}).Collection(aauthProtocol(t))
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
