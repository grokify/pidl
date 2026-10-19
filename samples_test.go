package pidl

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const goSource = `package walkthrough

func TestSign(t *testing.T) {
	// pidl:begin agent.sign
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	// pidl:begin agent.attach
	if err := aauth.SignRequest(req, key); err != nil {
		t.Fatal(err)
	}
	// pidl:end agent.attach
	// pidl:end agent.sign
}
`

func TestExtractRegion(t *testing.T) {
	got, err := ExtractRegion([]byte(goSource), "agent.sign")
	if err != nil {
		t.Fatal(err)
	}
	want := `req, err := http.NewRequest("GET", url, nil)
if err != nil {
	t.Fatal(err)
}
if err := aauth.SignRequest(req, key); err != nil {
	t.Fatal(err)
}`
	if got != want {
		t.Fatalf("ExtractRegion =\n%s\nwant\n%s", got, want)
	}

	inner, err := ExtractRegion([]byte(goSource), "agent.attach")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(inner, "if err := aauth.SignRequest") {
		t.Errorf("nested region = %q", inner)
	}

	other := "# pidl:begin py\r\n    x = 1\r\n\r\n    y = 2\r\n# pidl:end py\r\n<!-- pidl:begin html -->\n<p/>\n<!-- pidl:end html -->\n/* pidl:begin c */\nint x;\n/* pidl:end c */\n"
	for name, want := range map[string]string{"py": "x = 1\n\ny = 2", "html": "<p/>", "c": "int x;"} {
		got, err := ExtractRegion([]byte(other), name)
		if err != nil || got != want {
			t.Errorf("ExtractRegion(%q) = %q, %v; want %q", name, got, err, want)
		}
	}

	errCases := map[string]string{
		"missing":   "// pidl:begin a\nx\n// pidl:end a\n",
		"no end":    "// pidl:begin missing\nx\n",
		"twice":     "// pidl:begin missing\n// pidl:begin missing\nx\n// pidl:end missing\n",
		"early end": "// pidl:end missing\n// pidl:begin missing\nx\n",
		"empty":     "// pidl:begin missing\n\n// pidl:end missing\n",
	}
	for name, src := range errCases {
		if _, err := ExtractRegion([]byte(src), "missing"); err == nil {
			t.Errorf("%s: ExtractRegion succeeded", name)
		}
	}
	// A region name that is a prefix of another does not match it.
	if _, err := ExtractRegion([]byte("// pidl:begin agent.signed\nx\n// pidl:end agent.signed\n"), "agent.sign"); err == nil {
		t.Error("prefix region name matched")
	}
}

// sampledProtocol returns a protocol with an inline sample, a source
// sample on a step and a scenario-level source sample.
func sampledProtocol() *Protocol {
	p := narratedProtocol()
	src := &SampleSource{Repo: "github.com/aauth-dev/aauth-go", Path: "examples/walkthrough/agent_test.go", Region: "agent.sign"}
	p.Scenarios[0].Flows[0].Example = &FlowExample{
		Source: "draft-11", Kind: ExampleKindDerived,
		Request: &ExampleRequest{Method: "GET", URL: "https://resource.example/"},
		Samples: []Sample{
			{Language: "go", Label: "Agent", Role: "agent", Source: src},
			{Language: "python", Code: "print('hi')"},
		},
	}
	p.Scenarios[0].Flows[0].Representations["go"] = RepresentationNarration{Narration: Narration{"en-US": "In Go, the SDK signs the request."}}
	p.Scenarios[0].Samples = []Sample{{Language: "go", Label: "Setup", Source: &SampleSource{
		Repo: "github.com/aauth-dev/aauth-go", Path: "examples/walkthrough/agent_test.go", Region: "agent.attach"}}}
	return p
}

func fakeReader(files map[string]string) SourceReader {
	return func(repo, ref, path string) ([]byte, error) {
		s, ok := files[repo+"@"+ref+":"+path]
		if !ok {
			return nil, os.ErrNotExist
		}
		return []byte(s), nil
	}
}

func TestSamplesValid(t *testing.T) {
	if errs := sampledProtocol().Validate(); len(errs) > 0 {
		t.Fatalf("Validate() = %v", errs)
	}
}

func TestResolveCheckApplySamples(t *testing.T) {
	p := sampledProtocol()
	files := map[string]string{"github.com/aauth-dev/aauth-go@:examples/walkthrough/agent_test.go": goSource}
	lock, err := p.ResolveSamples(fakeReader(files))
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Samples) != 2 {
		t.Fatalf("lock has %d entries, want 2", len(lock.Samples))
	}
	if problems := p.CheckSamples(lock, fakeReader(files)); len(problems) > 0 {
		t.Fatalf("CheckSamples = %v", problems)
	}

	applied := p.WithSamples(lock)
	code := applied.Scenarios[0].Flows[0].Example.Samples[0].Code
	if !strings.HasPrefix(code, "req, err := http.NewRequest") {
		t.Fatalf("applied code = %q", code)
	}
	if p.Scenarios[0].Flows[0].Example.Samples[0].Code != "" {
		t.Fatal("WithSamples modified the receiver")
	}
	if applied.Scenarios[0].Samples[0].Code == "" {
		t.Fatal("scenario sample not applied")
	}
	w := applied.Walkthrough(nil)
	if w.Scenarios[1].Samples[0].Code == "" { // agent-identity is second in output order
		t.Fatalf("walkthrough scenario samples = %+v", w.Scenarios[1].Samples)
	}

	// The source changes after resolving.
	files["github.com/aauth-dev/aauth-go@:examples/walkthrough/agent_test.go"] = strings.Replace(goSource, `"GET"`, `"POST"`, 1)
	problems := p.CheckSamples(lock, fakeReader(files))
	if len(problems) != 1 || !strings.Contains(problems[0], "source changed") {
		t.Fatalf("after change, CheckSamples = %v", problems)
	}

	// The lock is tampered with, has an unused entry, and misses a source.
	lock.Lookup(*p.Scenarios[0].Flows[0].Example.Samples[0].Source).Code += "\n// edited"
	lock.Samples = append(lock.Samples, SampleLockEntry{SampleSource: SampleSource{Repo: "x", Path: "y", Region: "z"}, SHA256: codeHash("q"), Code: "q"})
	p.Scenarios[0].Samples[0].Source.Region = "agent.other"
	text := strings.Join(p.CheckSamples(lock, nil), "\n")
	for _, want := range []string{"does not match its sha256", "no sample uses it", "not resolved"} {
		if !strings.Contains(text, want) {
			t.Errorf("CheckSamples missing %q:\n%s", want, text)
		}
	}

	// A missing file is reported by resolve.
	if _, err := p.ResolveSamples(fakeReader(nil)); err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("ResolveSamples with no files = %v", err)
	}
}

func TestValidateSamplesErrors(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(s *Sample)
		wantField string
		wantMsg   string
	}{
		{"language required", func(s *Sample) { s.Language = "" }, ".language", "required"},
		{"curl is generated", func(s *Sample) { s.Language = "curl" }, ".language", "generated"},
		{"bad language", func(s *Sample) { s.Language = "Go" }, ".language", "must match"},
		{"both code and source", func(s *Sample) { s.Code = "x" }, "", "exactly one"},
		{"neither", func(s *Sample) { s.Source = nil }, "", "exactly one"},
		{"unknown role", func(s *Sample) { s.Role = "nobody" }, ".role", "unknown entity"},
		{"absolute path", func(s *Sample) { s.Source.Path = "/etc/passwd" }, ".source.path", "within the repository"},
		{"escaping path", func(s *Sample) { s.Source.Path = "../x.go" }, ".source.path", "within the repository"},
		{"repo required", func(s *Sample) { s.Source.Repo = "" }, ".source.repo", "required"},
		{"bad region", func(s *Sample) { s.Source.Region = "a b" }, ".source.region", "must match"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := sampledProtocol()
			s := p.Scenarios[0].Flows[0].Example.Samples[0]
			cp := *s.Source
			s.Source = &cp
			tt.mutate(&s)
			p.Scenarios[0].Flows[0].Example.Samples[0] = s
			want := "scenarios[0].flows[0].example.samples[0]" + tt.wantField
			for _, e := range p.Validate() {
				if e.Field == want && strings.Contains(e.Message, tt.wantMsg) {
					return
				}
			}
			t.Fatalf("Validate() = %v, want %s: %s", p.Validate(), want, tt.wantMsg)
		})
	}

	p := sampledProtocol()
	p.Scenarios[0].Flows[0].Representations["ruby"] = RepresentationNarration{Narration: Narration{"en-US": "x"}}
	found := false
	for _, e := range p.Validate() {
		if strings.Contains(e.Field, `representations["ruby"]`) {
			found = true
		}
	}
	if !found {
		t.Error("a representation with no matching sample language was accepted")
	}
}

func TestLocalSourceReader(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	repo := filepath.Join(root, "example.com", "sdk")
	if err := os.MkdirAll(repo, 0o750); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append( // #nosec G204 -- fixed git subcommands in a temporary test repository
			[]string{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(s string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(repo, "a.go"), []byte(s), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	run("init", "-q")
	write("// pidl:begin r\nv1\n// pidl:end r\n")
	run("add", "a.go")
	run("commit", "-q", "-m", "v1")
	run("tag", "v1")
	write("// pidl:begin r\nv2\n// pidl:end r\n")

	read := LocalSourceReader(root)
	for ref, want := range map[string]string{"": "v2", "v1": "v1"} {
		data, err := read("example.com/sdk", ref, "a.go")
		if err != nil {
			t.Fatalf("ref %q: %v", ref, err)
		}
		if got, err := ExtractRegion(data, "r"); err != nil || got != want {
			t.Errorf("ref %q: region = %q, %v; want %q", ref, got, err, want)
		}
	}
	if _, err := read("example.com/sdk", "nope", "a.go"); err == nil {
		t.Error("unknown ref succeeded")
	}
}

func TestSampleLockFile(t *testing.T) {
	if got := SampleLockPath("dir/aauth.json"); got != "dir/aauth.samples.json" {
		t.Errorf("SampleLockPath = %q", got)
	}
	path := filepath.Join(t.TempDir(), "x.samples.json")
	lock := &SampleLock{Version: 1, Samples: []SampleLockEntry{
		{SampleSource: SampleSource{Repo: "b", Path: "p", Region: "r"}, SHA256: codeHash("2"), Code: "2"},
		{SampleSource: SampleSource{Repo: "a", Path: "p", Region: "r"}, SHA256: codeHash("1"), Code: "1"},
	}}
	if err := lock.WriteFile(path); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSampleLock(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Samples[0].Repo != "a" || got.Lookup(SampleSource{Repo: "b", Path: "p", Region: "r"}).Code != "2" {
		t.Fatalf("lock round trip = %+v", got)
	}
}
