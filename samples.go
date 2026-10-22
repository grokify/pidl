package pidl

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	oscfs "github.com/grokify/oscompat/fs"
)

// Sample is code that sends or handles a step's message, such as a Go SDK
// call. It is either inline Code or a Source reference to a marked region
// of a tested file, resolved into a SampleLock.
type Sample struct {
	// Language names the code's language with a common highlighter name,
	// such as "go", "typescript", "python" or "csharp". Samples are also
	// representations a step's narration can be overridden for.
	Language string `json:"language"`

	// Label distinguishes samples in one language, such as "Agent" and
	// "Resource".
	Label string `json:"label,omitempty"`

	// Role is the entity ID whose developer writes this code.
	Role string `json:"role,omitempty"`

	// Code is inline code, or a source's code once resolved. A sample has
	// Code, Source, or both.
	Code string `json:"code,omitempty"`

	// Source refers to a marked region of a file in a repository.
	Source *SampleSource `json:"source,omitempty"`
}

// SampleSource refers to the code between "pidl:begin <region>" and
// "pidl:end <region>" markers, in comments, in a file of a repository.
type SampleSource struct {
	// Repo is the repository's module-style path, such as
	// "github.com/aauth-dev/aauth-go".
	Repo string `json:"repo"`

	// Ref is a tag or commit to read the file at. Empty reads the
	// checkout's working tree.
	Ref string `json:"ref,omitempty"`

	// Path is the file's path within the repository.
	Path string `json:"path"`

	// Region is the marked region's name.
	Region string `json:"region"`
}

// Key identifies the source in a SampleLock.
func (s SampleSource) Key() string {
	return s.Repo + "@" + s.Ref + ":" + s.Path + "#" + s.Region
}

// reservedSampleLanguages are representations that are not samples.
var reservedSampleLanguages = map[string]bool{RepresentationHTTP: true, RepresentationCurl: true}

var (
	sampleLanguagePattern = regexp.MustCompile(`^[a-z][a-z0-9+#-]*$`)
	regionNamePattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)
)

// SampleLock records the code each sample source resolved to, so documents
// generate the same output without the source repositories at hand.
type SampleLock struct {
	Version int               `json:"version"`
	Samples []SampleLockEntry `json:"samples"`
}

// SampleLockEntry is one resolved source.
type SampleLockEntry struct {
	SampleSource
	// SHA256 is the hex SHA-256 of Code.
	SHA256 string `json:"sha256"`
	Code   string `json:"code"`
}

// Lookup returns the entry for src, or nil.
func (l *SampleLock) Lookup(src SampleSource) *SampleLockEntry {
	if l == nil {
		return nil
	}
	for i := range l.Samples {
		if l.Samples[i].Key() == src.Key() {
			return &l.Samples[i]
		}
	}
	return nil
}

// ReadSampleLock reads a lock file.
func ReadSampleLock(path string) (*SampleLock, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- the path is chosen by the caller
	if err != nil {
		return nil, err
	}
	l, err := ParseSampleLock(data)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return l, nil
}

// ParseSampleLock parses a lock file's contents.
func ParseSampleLock(data []byte) (*SampleLock, error) {
	var l SampleLock
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// SampleLockSuffix ends the name of a document's sample lock file.
const SampleLockSuffix = ".samples.json"

// WriteFile writes the lock with entries sorted by source.
func (l *SampleLock) WriteFile(path string) error {
	sort.Slice(l.Samples, func(i, j int) bool { return l.Samples[i].Key() < l.Samples[j].Key() })
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// SampleLockPath returns the conventional lock path for a document:
// "aauth.json" uses "aauth.samples.json".
func SampleLockPath(docPath string) string {
	return strings.TrimSuffix(docPath, filepath.Ext(docPath)) + SampleLockSuffix
}

// SourceReader reads a file of a repository at a ref ("" for the working
// tree).
type SourceReader func(repo, ref, path string) ([]byte, error)

// LocalSourceReader reads repositories checked out under root, at
// root/<repo>. With a ref it reads the file at that ref through git.
func LocalSourceReader(root string) SourceReader {
	return func(repo, ref, path string) ([]byte, error) {
		dir := filepath.Join(root, filepath.FromSlash(repo))
		if ref == "" {
			return os.ReadFile(filepath.Join(dir, filepath.FromSlash(path))) // #nosec G304 -- paths come from the document's sample sources
		}
		cmd := exec.Command("git", "-C", dir, "show", ref+":"+path) // #nosec G204 -- arguments are passed without a shell
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("git show %s:%s in %s: %w: %s", ref, path, dir, err, strings.TrimSpace(stderr.String()))
		}
		return out, nil
	}
}

// ExtractRegion returns the lines between "pidl:begin <name>" and
// "pidl:end <name>", with marker lines of other regions removed, common
// indentation removed, and leading and trailing blank lines trimmed.
// Markers may sit in any comment syntax.
func ExtractRegion(content []byte, name string) (string, error) {
	begin, end := "pidl:begin "+name, "pidl:end "+name
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	start, stop := -1, -1
	for i, line := range lines {
		switch {
		case markerIs(line, begin):
			if start >= 0 {
				return "", fmt.Errorf("region %q begins twice", name)
			}
			start = i
		case markerIs(line, end):
			if start < 0 {
				return "", fmt.Errorf("region %q ends before it begins", name)
			}
			if stop < 0 {
				stop = i
			}
		}
	}
	if start < 0 {
		return "", fmt.Errorf("region %q not found", name)
	}
	if stop < 0 {
		return "", fmt.Errorf("region %q has no end marker", name)
	}

	var body []string
	for _, line := range lines[start+1 : stop] {
		if strings.Contains(line, "pidl:begin ") || strings.Contains(line, "pidl:end ") {
			continue
		}
		body = append(body, line)
	}
	for len(body) > 0 && strings.TrimSpace(body[0]) == "" {
		body = body[1:]
	}
	for len(body) > 0 && strings.TrimSpace(body[len(body)-1]) == "" {
		body = body[:len(body)-1]
	}
	if len(body) == 0 {
		return "", fmt.Errorf("region %q is empty", name)
	}
	return dedent(body), nil
}

// markerIs reports whether line holds marker as a whole word sequence.
func markerIs(line, marker string) bool {
	i := strings.Index(line, marker)
	if i < 0 {
		return false
	}
	rest := line[i+len(marker):]
	return rest == "" || rest[0] == ' ' || rest[0] == '\t' || strings.HasPrefix(rest, "*/") || strings.HasPrefix(rest, "-->")
}

// dedent removes the longest whitespace prefix shared by non-blank lines.
func dedent(lines []string) string {
	prefix := ""
	first := true
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		ws := l[:len(l)-len(strings.TrimLeft(l, " \t"))]
		if first {
			prefix, first = ws, false
			continue
		}
		for !strings.HasPrefix(ws, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = strings.TrimRight(strings.TrimPrefix(l, prefix), " \t")
	}
	return strings.Join(out, "\n")
}

// codeHash returns the hex SHA-256 of code.
func codeHash(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// SampleSources returns every distinct sample source in the document, in
// document order.
func (p *Protocol) SampleSources() []SampleSource {
	var out []SampleSource
	seen := map[string]bool{}
	add := func(samples []Sample) {
		for _, s := range samples {
			if s.Source != nil && !seen[s.Source.Key()] {
				seen[s.Source.Key()] = true
				out = append(out, *s.Source)
			}
		}
	}
	for _, sc := range p.Scenarios {
		add(sc.Samples)
	}
	for _, list := range p.flowLists() {
		for _, f := range list.flows {
			if f.Example != nil {
				add(f.Example.Samples)
			}
		}
	}
	return out
}

// ResolveSamples reads every sample source and returns a lock with its
// code and hash. It reports every source that cannot be read or extracted.
func (p *Protocol) ResolveSamples(read SourceReader) (*SampleLock, error) {
	lock := &SampleLock{Version: 1, Samples: []SampleLockEntry{}}
	var errs []error
	for _, src := range p.SampleSources() {
		data, err := read(src.Repo, src.Ref, src.Path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", src.Key(), err))
			continue
		}
		code, err := ExtractRegion(data, src.Region)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", src.Key(), err))
			continue
		}
		lock.Samples = append(lock.Samples, SampleLockEntry{SampleSource: src, SHA256: codeHash(code), Code: code})
	}
	return lock, errors.Join(errs...)
}

// CheckSamples compares the lock with the sources. It reports sources
// missing from the lock, locked code whose hash is wrong, sources whose
// region changed since it was resolved, and lock entries no sample uses.
func (p *Protocol) CheckSamples(lock *SampleLock, read SourceReader) []string {
	var problems []string
	used := map[string]bool{}
	for _, src := range p.SampleSources() {
		used[src.Key()] = true
		entry := lock.Lookup(src)
		if entry == nil {
			problems = append(problems, fmt.Sprintf("%s: not resolved; run pidl samples resolve", src.Key()))
			continue
		}
		if codeHash(entry.Code) != entry.SHA256 {
			problems = append(problems, fmt.Sprintf("%s: locked code does not match its sha256", src.Key()))
		}
		if read == nil {
			continue
		}
		data, err := read(src.Repo, src.Ref, src.Path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", src.Key(), err))
			continue
		}
		code, err := ExtractRegion(data, src.Region)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", src.Key(), err))
			continue
		}
		if codeHash(code) != entry.SHA256 {
			problems = append(problems, fmt.Sprintf("%s: source changed since it was resolved; run pidl samples resolve", src.Key()))
		}
	}
	if lock != nil {
		for _, e := range lock.Samples {
			if !used[e.Key()] {
				problems = append(problems, fmt.Sprintf("%s: locked but no sample uses it", e.Key()))
			}
		}
	}
	return problems
}

// WithSamples returns a copy of the protocol whose source samples carry
// their locked code, ready for generation. Sources missing from the lock
// keep empty code. The receiver is not modified.
func (p *Protocol) WithSamples(lock *SampleLock) *Protocol {
	fill := func(samples []Sample) []Sample {
		if len(samples) == 0 {
			return samples
		}
		out := make([]Sample, len(samples))
		for i, s := range samples {
			if s.Source != nil {
				if e := lock.Lookup(*s.Source); e != nil {
					s.Code = e.Code
				}
			}
			out[i] = s
		}
		return out
	}
	fillFlows := func(flows []Flow) []Flow {
		out := make([]Flow, len(flows))
		for i, f := range flows {
			if f.Example != nil && len(f.Example.Samples) > 0 {
				ex := *f.Example
				ex.Samples = fill(ex.Samples)
				f.Example = &ex
			}
			out[i] = f
		}
		return out
	}

	cp := *p
	cp.Flows = fillFlows(p.Flows)
	cp.Scenarios = make([]Scenario, len(p.Scenarios))
	for i, sc := range p.Scenarios {
		sc.Samples = fill(sc.Samples)
		sc.Flows = fillFlows(sc.Flows)
		cp.Scenarios[i] = sc
	}
	return &cp
}

// sampleLanguages returns the languages of a flow's example samples.
func sampleLanguages(f Flow) map[string]bool {
	langs := map[string]bool{}
	if f.Example != nil {
		for _, s := range f.Example.Samples {
			langs[s.Language] = true
		}
	}
	return langs
}

// validateSamples checks one list of samples.
func validateSamples(field string, samples []Sample, entityIDs map[string]bool) ValidationErrors {
	var errs ValidationErrors
	for i, s := range samples {
		sf := fmt.Sprintf("%s[%d]", field, i)
		switch {
		case s.Language == "":
			errs = append(errs, ValidationError{Field: sf + ".language", Message: "required"})
		case reservedSampleLanguages[s.Language]:
			errs = append(errs, ValidationError{Field: sf + ".language", Message: fmt.Sprintf("%q is generated from the example, not a sample", s.Language)})
		case !sampleLanguagePattern.MatchString(s.Language):
			errs = append(errs, ValidationError{Field: sf + ".language", Message: "must match pattern ^[a-z][a-z0-9+#-]*$"})
		}
		if s.Code == "" && s.Source == nil {
			errs = append(errs, ValidationError{Field: sf, Message: "set code or source"})
		}
		if s.Role != "" && !entityIDs[s.Role] {
			errs = append(errs, ValidationError{Field: sf + ".role", Message: fmt.Sprintf("unknown entity %q", s.Role)})
		}
		if src := s.Source; src != nil {
			if src.Repo == "" {
				errs = append(errs, ValidationError{Field: sf + ".source.repo", Message: "required"})
			}
			// A repository path uses forward slashes on every platform, so it is
			// checked the same way everywhere, not with filepath.IsAbs.
			if oscfs.ValidateSlashPath(src.Path) != nil {
				errs = append(errs, ValidationError{Field: sf + ".source.path", Message: "required, relative, and within the repository"})
			}
			if !regionNamePattern.MatchString(src.Region) {
				errs = append(errs, ValidationError{Field: sf + ".source.region", Message: "must match pattern ^[A-Za-z0-9][A-Za-z0-9_.-]*$"})
			}
		}
	}
	return errs
}
