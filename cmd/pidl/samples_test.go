package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const samplesDoc = `{
  "protocol": {"id": "demo", "name": "Demo"},
  "entities": [
    {"id": "agent", "name": "Agent", "type": "agent"},
    {"id": "resource", "name": "Resource", "type": "resource_server"}
  ],
  "scenarios": [
    {"id": "sign", "name": "Signed request", "flows": [
      {"id": "request", "from": "agent", "to": "resource", "action": "signed_request",
       "example": {"source": "spec", "kind": "derived",
         "request": {"method": "GET", "url": "https://resource.example/"},
         "samples": [{"language": "go", "role": "agent",
           "source": {"repo": "example.com/sdk", "path": "sign_test.go", "region": "sign"}}]}}
    ]}
  ]
}`

func TestRunSamples(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "example.com", "sdk", "sign_test.go")
	if err := os.MkdirAll(filepath.Dir(src), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("func x() {\n\t// pidl:begin sign\n\tsdk.Sign(req)\n\t// pidl:end sign\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	doc := filepath.Join(dir, "demo.json")
	if err := os.WriteFile(doc, []byte(samplesDoc), 0o600); err != nil {
		t.Fatal(err)
	}

	call := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		code := run(args, &out, &errOut)
		return code, out.String(), errOut.String()
	}

	if code, _, errOut := call("samples", "check", "--root", root, doc); code != 1 || !strings.Contains(errOut, "demo.samples.json") {
		t.Fatalf("check before resolve = %d, %q", code, errOut)
	}
	if code, out, errOut := call("samples", "resolve", "--root", root, doc); code != 0 || !strings.Contains(out, "Resolved 1 sample sources") {
		t.Fatalf("resolve = %d, %q, %q", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "demo.samples.json")); err != nil {
		t.Fatal(err)
	}
	if code, out, errOut := call("samples", "check", "--root", root, doc); code != 0 {
		t.Fatalf("check = %d, %q, %q", code, out, errOut)
	}
	code, out, errOut := call("generate", "-f", "walkthrough-json", doc)
	if code != 0 || !strings.Contains(out, `"code": "sdk.Sign(req)"`) || !strings.Contains(out, `"curl": "curl 'https://resource.example/'"`) {
		t.Fatalf("generate = %d\n%s\n%s", code, out, errOut)
	}

	if err := os.WriteFile(src, []byte("// pidl:begin sign\nsdk.Sign(req, opts)\n// pidl:end sign\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut := call("samples", "check", "--root", root, doc); code != 1 || !strings.Contains(errOut, "source changed") {
		t.Fatalf("check after change = %d, %q", code, errOut)
	}
	if code, _, errOut := call("generate", "-f", "walkthrough-json", "--samples", filepath.Join(dir, "nope.json"), doc); code != 1 || !strings.Contains(errOut, "Error reading samples") {
		t.Fatalf("generate with missing explicit lock = %d, %q", code, errOut)
	}
	if code, _, _ := call("samples", "frobnicate", doc); code != 1 {
		t.Fatal("unknown samples command succeeded")
	}
}
