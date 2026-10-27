package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

const (
	testTrace    = "../../observe/adapter/sandbox/testdata/petstore.trace.json"
	testDocument = "../../observe/match/testdata/aauth_protocol.json"
	testBinding  = "../../observe/match/testdata/petstore.binding.json"
)

// importTrace converts the Petstore sandbox trace to an observation file.
func importTrace(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "run.json")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"observe", "import", "--from", "sandbox", "-o", out, testTrace}, &stdout, &stderr); code != 0 {
		t.Fatalf("observe import exited %d: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "19 exchange(s)") {
		t.Errorf("import report = %q", stderr.String())
	}
	return out
}

func TestObserveImport(t *testing.T) {
	path := importTrace(t)
	o, err := pidl.ReadObservation(path)
	if err != nil {
		t.Fatal(err)
	}
	if errs := o.Validate(); len(errs) > 0 || len(o.Exchanges) != 19 {
		t.Errorf("imported observation: %d exchanges, %v", len(o.Exchanges), errs)
	}

	var stdout, stderr bytes.Buffer
	if code := run([]string{"observe", "import", "--from", "sandbox", testTrace}, &stdout, &stderr); code != 0 || !pidl.IsObservation(stdout.Bytes()) {
		t.Errorf("import to stdout exited %d", code)
	}
	for _, args := range [][]string{
		{"observe", "import", "--from", "pcap", testTrace},
		{"observe", "import", "--from", "sandbox"},
	} {
		stdout.Reset()
		stderr.Reset()
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Errorf("%v exited %d, want 2", args, code)
		}
	}
	if code := run([]string{"observe", "nope"}, &stdout, &stderr); code != 1 {
		t.Errorf("unknown observe command exited %d", code)
	}
}

func TestGenerateWithObservation(t *testing.T) {
	obs := importTrace(t)
	out := filepath.Join(t.TempDir(), "live.walkthrough.json")
	var stdout, stderr bytes.Buffer
	code := run([]string{"generate", "-f", "walkthrough-json", "--observation", obs, "--binding", testBinding, "-o", out, testDocument}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("generate exited %d: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "observed: 2 scenario instance(s) [agent-identity, ps-authorization-challenge]") {
		t.Errorf("report = %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), `warning: consent/poll: example path /pending/f7a3b9c`) {
		t.Errorf("no literal-identifier warning in %q", stderr.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var w pidl.Walkthrough
	if err := json.Unmarshal(data, &w); err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, sc := range w.Scenarios {
		if sc.ObservedRun != nil {
			ids = append(ids, sc.ID)
		}
	}
	if strings.Join(ids, ",") != "agent-identity.observed-1,ps-authorization-challenge.observed-2" {
		t.Errorf("observed entries = %v", ids)
	}
}

func TestGenerateViewerWithObservation(t *testing.T) {
	obs := importTrace(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"generate", "-f", "viewer", "--observation", obs, "--binding", testBinding, "--from", "2026-10-27T09:02:31Z", testDocument}, &stdout, &stderr); code != 0 {
		t.Fatalf("generate viewer exited %d: %s", code, stderr.String())
	}
	page := stdout.String()
	if !strings.Contains(page, `"observed_run"`) || !strings.Contains(page, "ps-authorization-challenge.observed-2") {
		t.Error("the viewer page does not embed the observed run")
	}
	if !strings.Contains(stderr.String(), "1 outside the time window") {
		t.Errorf("report = %q", stderr.String())
	}
}

func TestGenerateObservationErrors(t *testing.T) {
	obs := importTrace(t)
	dir := t.TempDir()
	notObservation := filepath.Join(dir, "trace.json")
	data, err := os.ReadFile(testTrace)
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := root.WriteFile("trace.json", data, 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		args []string
		code int
		want string
	}{
		{"other format", []string{"generate", "-f", "plantuml", "--observation", obs, testDocument}, 2, "works with -f walkthrough-json and -f viewer"},
		{"not an observation", []string{"generate", "-f", "walkthrough-json", "--observation", notObservation, testDocument}, 1, "pidl observe import"},
		{"two observations", []string{"generate", "-f", "walkthrough-json", "--observation", obs, "--observation", obs, testDocument}, 1, "one --observation at a time"},
		{"bad time", []string{"generate", "-f", "walkthrough-json", "--observation", obs, "--from", "yesterday", testDocument}, 1, "--from"},
		{"unknown scenario", []string{"generate", "-f", "walkthrough-json", "--observation", obs, "--scenarios", "nope", testDocument}, 1, `no scenario "nope"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr)
			if code != tt.code || !strings.Contains(stderr.String(), tt.want) {
				t.Errorf("exit %d, stderr %q; want %d containing %q", code, stderr.String(), tt.code, tt.want)
			}
		})
	}
}

func TestGenerateRawWarns(t *testing.T) {
	obs := importTrace(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"generate", "-f", "walkthrough-json", "--observation", obs, "--binding", testBinding, "--raw", testDocument}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--raw keeps raw tokens") || !strings.Contains(stdout.String(), `"redacted": false`) {
		t.Errorf("--raw did not warn or mark the run: %q", stderr.String())
	}
}
