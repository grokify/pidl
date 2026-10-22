package pidl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sdkSource = "func TestSign(t *testing.T) {\n\t// pidl:begin sign\n\treq.Sign(key)\n\t// pidl:end sign\n\t// pidl:begin approve\n\tserver.Approve(id)\n\t// pidl:end approve\n}\n"

func sdkSet() *SampleSet {
	src := func(region string) *SampleSource {
		return &SampleSource{Repo: "example.com/sdk", Path: "walkthrough_test.go", Region: region}
	}
	return &SampleSet{
		Format: SampleSetFormat, Protocol: "aauth", Name: "example-sdk",
		Samples: []SampleSetEntry{
			{Step: "agent-identity/sign", Sample: Sample{Language: "go", Label: "Agent", Role: "agent", Source: src("sign")}},
			{Step: "interaction/approve", Sample: Sample{Language: "go", Label: "Server", Source: src("approve")}},
			{Step: "agent-identity", Sample: Sample{Language: "go", Label: "Setup", Code: "key := sdk.NewKey()"}},
		},
	}
}

func sdkReader() SourceReader {
	return fakeReader(map[string]string{"example.com/sdk@:walkthrough_test.go": sdkSource})
}

func TestIsSampleSet(t *testing.T) {
	if !IsSampleSet([]byte(`{"format": "pidl-samples/v1", "samples": []}`)) {
		t.Error("a sample set was not recognized")
	}
	if IsSampleSet([]byte(`{"version": 1, "samples": []}`)) {
		t.Error("a sample lock was taken for a sample set")
	}
	if _, err := ParseSampleSet([]byte(`{"format": "pidl-samples/v2", "samples": []}`)); err == nil {
		t.Error("another major version was accepted")
	}
}

func TestSampleSetResolveCheck(t *testing.T) {
	set := sdkSet()
	if errs := set.Validate(); len(errs) > 0 {
		t.Fatalf("Validate() = %v", errs)
	}
	if problems := set.Check(nil); len(problems) != 2 {
		t.Fatalf("Check before resolving = %v, want 2 unresolved sources", problems)
	}
	if err := set.Resolve(sdkReader()); err != nil {
		t.Fatal(err)
	}
	if got := set.Samples[0].Code; got != "req.Sign(key)" {
		t.Fatalf("resolved code = %q", got)
	}
	if problems := set.Check(sdkReader()); len(problems) > 0 {
		t.Fatalf("Check after resolving = %v", problems)
	}

	changed := fakeReader(map[string]string{"example.com/sdk@:walkthrough_test.go": strings.Replace(sdkSource, "req.Sign(key)", "req.Sign(otherKey)", 1)})
	if problems := set.Check(changed); len(problems) != 1 || !strings.Contains(problems[0], "source changed") {
		t.Fatalf("Check after the source changed = %v", problems)
	}
	set.Samples[1].Code = "tampered"
	if problems := set.Check(nil); len(problems) != 1 || !strings.Contains(problems[0], "does not match its sha256") {
		t.Fatalf("Check after the code was edited = %v", problems)
	}

	bad := &SampleSet{Format: SampleSetFormat, Samples: []SampleSetEntry{
		{Sample: Sample{Language: "curl", Code: "x"}},
		{Step: "s/a", Sample: Sample{Language: "go", Source: &SampleSource{Repo: "example.com/sdk", Path: "/etc/passwd", Region: "r"}}},
	}}
	if errs := bad.Validate(); len(errs) != 3 {
		t.Fatalf("Validate() = %v, want a missing step, a reserved language and a bad path", errs)
	}
	if err := (&SampleSet{Format: SampleSetFormat, Samples: []SampleSetEntry{
		{Step: "s/a", Sample: Sample{Language: "go", Source: &SampleSource{Repo: "example.com/sdk", Path: "missing.go", Region: "r"}}},
	}}).Resolve(sdkReader()); err == nil {
		t.Error("resolving a missing file succeeded")
	}
}

func TestWithSampleSet(t *testing.T) {
	p := structuredProtocol()
	set := sdkSet()
	if err := set.Resolve(sdkReader()); err != nil {
		t.Fatal(err)
	}
	got, err := p.WithSampleSet(set)
	if err != nil {
		t.Fatal(err)
	}
	if p.Scenarios[0].Flows[0].Example != nil || len(p.Scenarios[0].Samples) != 0 {
		t.Fatal("WithSampleSet modified the receiver")
	}

	ex := got.Scenarios[0].Flows[0].Example
	if ex == nil || len(ex.Samples) != 1 || ex.Samples[0].Code != "req.Sign(key)" {
		t.Fatalf("agent-identity/sign example = %+v", ex)
	}
	if ex.Source != "example-sdk" || ex.Kind != ExampleKindDerived {
		t.Errorf("created example cites %q (%s), want the set's name", ex.Source, ex.Kind)
	}
	if s := got.Scenarios[0].Samples; len(s) != 1 || s[0].Label != "Setup" {
		t.Errorf("scenario samples = %+v", s)
	}

	// A sample on a fragment step appears wherever the fragment is
	// included.
	w := got.Walkthrough(nil)
	var approve *WalkthroughStep
	for _, sc := range w.Scenarios {
		if sc.ID != "three-party" {
			continue
		}
		for i := range sc.Steps {
			if sc.Steps[i].ID == "consent.approve" {
				approve = &sc.Steps[i]
			}
		}
	}
	if approve == nil || approve.Example == nil || len(approve.Example.Samples) != 1 || approve.Example.Samples[0].Code != "server.Approve(id)" {
		t.Fatalf("three-party step consent.approve = %+v", approve)
	}

	// A step's own samples come first.
	p.Scenarios[0].Flows[0].Example = &FlowExample{Source: "spec", Kind: ExampleKindDraft, Samples: []Sample{{Language: "python", Code: "sign()"}}}
	got, err = p.WithSampleSet(set)
	if err != nil {
		t.Fatal(err)
	}
	if s := got.Scenarios[0].Flows[0].Example.Samples; len(s) != 2 || s[0].Language != "python" || s[1].Language != "go" {
		t.Errorf("samples = %+v, want the step's own first", s)
	}
	if got.Scenarios[0].Flows[0].Example.Source != "spec" {
		t.Error("an existing example lost its source")
	}
}

func TestWithSampleSetErrors(t *testing.T) {
	p := structuredProtocol()
	tests := []struct {
		name string
		set  *SampleSet
		want string
	}{
		{"another protocol", &SampleSet{Format: SampleSetFormat, Protocol: "other"}, `for protocol "other"`},
		{"unknown scenario", &SampleSet{Format: SampleSetFormat, Samples: []SampleSetEntry{
			{Step: "nope/sign", Sample: Sample{Language: "go", Code: "x"}}}}, `no scenario "nope"`},
		{"unknown step", &SampleSet{Format: SampleSetFormat, Samples: []SampleSetEntry{
			{Step: "agent-identity/nope", Sample: Sample{Language: "go", Code: "x"}}}}, `no step "nope"`},
		{"unresolved source", &SampleSet{Format: SampleSetFormat, Samples: []SampleSetEntry{
			{Step: "agent-identity/sign", Sample: Sample{Language: "go", Source: &SampleSource{Repo: "r", Path: "p", Region: "x"}}}}}, "not resolved"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := p.WithSampleSet(tt.set); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("WithSampleSet() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestWithSampleSetFlowList(t *testing.T) {
	p := NewMinimalProtocol("demo", "Demo")
	p.Flows[0].ID = "call"
	set := &SampleSet{Format: SampleSetFormat, Samples: []SampleSetEntry{{Step: "call", Sample: Sample{Language: "go", Code: "c.Call()"}}}}
	got, err := p.WithSampleSet(set)
	if err != nil {
		t.Fatal(err)
	}
	if ex := got.Flows[0].Example; ex == nil || len(ex.Samples) != 1 || ex.Source != "code samples" {
		t.Fatalf("example = %+v", ex)
	}
}

func TestSampleSetFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sdk.pidl-samples.json")
	set := sdkSet()
	if err := set.WriteFile(path); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSampleSet(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Samples) != 3 || got.Samples[0].Step != "agent-identity/sign" || got.Samples[2].Code == "" {
		t.Fatalf("read back %+v", got)
	}
	if err := os.WriteFile(path, []byte(`{"format": "x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSampleSet(path); err == nil {
		t.Error("a file with another format was read")
	}
}

func TestCheckoutSourceReader(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "walkthrough_test.go"), []byte(sdkSource), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := CheckoutSourceReader(dir)("any.example/repo", "", "walkthrough_test.go")
	if err != nil || string(data) != sdkSource {
		t.Fatalf("read %q, %v", data, err)
	}
}
