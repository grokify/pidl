package pidl

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestWireBody(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"empty", "", ""},
		{"compact keeps key order", "{\n  \"z\": 1,\n  \"a\": [1, 2]\n}", `{"z":1,"a":[1,2]}`},
		{"string body is sent as is", `"grant_type=code&code=abc"`, "grant_type=code&code=abc"},
		{"unicode kept", `{"name": "café"}`, `{"name":"café"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WireBody(json.RawMessage(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Fatalf("WireBody = %q, want %q", got, tt.want)
			}
		})
	}
	if _, err := WireBody(json.RawMessage(`{"a":`)); err == nil {
		t.Error("WireBody accepted invalid JSON")
	}
}

func TestCurl(t *testing.T) {
	r := &ExampleRequest{
		Method: "POST",
		URL:    "https://ps.example/token",
		Headers: []ExampleHeader{
			{Name: "Content-Type", Value: "application/json"},
			{Name: "Signature-Input", Value: `sig=("@method");created=1`},
			{Name: "Signature", Value: "sig=:abc:"},
			{Name: "X-Note", Value: "it's"},
		},
		Body: json.RawMessage(`{ "resource_token": "eyJ..." }`),
	}
	got, err := r.Curl()
	if err != nil {
		t.Fatal(err)
	}
	want := `# The signature headers are replayed from the example; curl cannot create HTTP message signatures.
curl -X POST 'https://ps.example/token' \
  -H 'Content-Type: application/json' \
  -H 'Signature-Input: sig=("@method");created=1' \
  -H 'Signature: sig=:abc:' \
  -H 'X-Note: it'\''s' \
  --data-raw '{"resource_token":"eyJ..."}'`
	if got != want {
		t.Fatalf("Curl() =\n%s\nwant\n%s", got, want)
	}

	get, err := (&ExampleRequest{Method: "GET", URL: "https://resource.example/data"}).Curl()
	if err != nil {
		t.Fatal(err)
	}
	if get != "curl 'https://resource.example/data'" {
		t.Errorf("GET Curl() = %q", get)
	}
	if s, err := (*ExampleRequest)(nil).Curl(); s != "" || err != nil {
		t.Errorf("nil Curl() = %q, %v", s, err)
	}
}

// TestCurlBodyMatchesDigest checks that every AAuth example whose request
// carries a Content-Digest produces a curl body with that digest.
func TestCurlBodyMatchesDigest(t *testing.T) {
	p, err := ParseFile("examples/aauth_protocol.json")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, list := range p.flowLists() {
		for _, f := range list.flows {
			if f.Example == nil || f.Example.Request == nil {
				continue
			}
			for _, h := range f.Example.Request.Headers {
				if !strings.EqualFold(h.Name, "Content-Digest") {
					continue
				}
				body, err := WireBody(f.Example.Request.Body)
				if err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(body)
				if want := "sha-256=:" + base64.StdEncoding.EncodeToString(sum[:]) + ":"; h.Value != want {
					t.Errorf("%s: Content-Digest %s, wire body hashes to %s", f.Action, h.Value, want)
				}
				checked++
			}
		}
	}
	if checked == 0 {
		t.Fatal("no examples with Content-Digest found")
	}
}

func TestWalkthroughCurl(t *testing.T) {
	p, err := ParseFile("examples/aauth_protocol.json")
	if err != nil {
		t.Fatal(err)
	}
	w := p.Walkthrough(nil)
	found := false
	for _, sc := range w.Scenarios {
		for _, st := range sc.Steps {
			if st.Example != nil && st.Example.Request != nil {
				if !strings.HasPrefix(strings.TrimPrefix(st.Curl, "# The signature headers are replayed from the example; curl cannot create HTTP message signatures.\n"), "curl ") {
					t.Fatalf("%s step %d curl = %q", sc.ID, st.Number, st.Curl)
				}
				found = true
			}
			if st.Example == nil && st.Curl != "" {
				t.Errorf("%s step %d has curl without an example", sc.ID, st.Number)
			}
		}
	}
	if !found {
		t.Fatal("no request examples in the walkthrough")
	}
}
