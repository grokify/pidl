package pidl

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// BindingFormat names the binding format and its major version.
const BindingFormat = "pidl-binding/v1"

// Binding describes one deployment of a protocol, so observed exchanges
// can be matched to a document's steps: the labels a capturing tool uses
// for each participant and the hosts it answers at, paths that differ from
// the document's examples, URL rewrites a capture needs, and rules that
// replace a step's own for this deployment.
type Binding struct {
	// Format names the format and its major version: pidl-binding/v1.
	Format string `json:"format"`

	// Participants maps a document participant ID to how a capture names it.
	Participants map[string]BindingParticipant `json:"participants"`

	// LabelsByPath resolves a generic sender label, such as a trace's
	// "client", by the request path.
	LabelsByPath []BindingLabelRule `json:"labelsByPath,omitempty"`

	// Paths gives, per participant, the path of a step where the
	// deployment's path differs from the document's example. Keys are step
	// addresses, as in Steps.
	Paths map[string]map[string]string `json:"paths,omitempty"`

	// Rewrite rewrites every exchange's URL before matching, in order.
	Rewrite []BindingRewrite `json:"rewrite,omitempty"`

	// Unobservable lists participants a capture cannot see, in addition to
	// those whose role type is user.
	Unobservable []string `json:"unobservable,omitempty"`

	// Steps replaces a step's rule for this deployment. A key is
	// "<scenario>/<step-id>", or "*/<step-id>" for every scenario that has
	// the step: step IDs repeat across scenarios. A scenario's key wins over
	// the wildcard.
	Steps map[string]BindingStep `json:"steps,omitempty"`
}

// BindingParticipant is how a capture names one participant.
type BindingParticipant struct {
	// Labels are the capturing tool's names for the participant.
	Labels []string `json:"labels,omitempty"`
	// Hosts are the host names the participant answers at.
	Hosts []string `json:"hosts,omitempty"`
}

// BindingLabelRule resolves a sender label by the request path.
type BindingLabelRule struct {
	// Label is the sender label the rule applies to, such as "client".
	Label string `json:"label"`
	// Path is a regular expression the request path must match.
	Path string `json:"path"`
	// Participant is the participant the sender is taken to be.
	Participant string `json:"participant"`
}

// BindingRewrite rewrites a URL by a regular expression.
type BindingRewrite struct {
	// Match is a regular expression over the whole URL.
	Match string `json:"match"`
	// To is the replacement, with $1-style references to Match's groups.
	To string `json:"to"`
}

// BindingStep is a deployment's rule for one step.
type BindingStep struct {
	Match StepMatch `json:"match"`
}

// StepMatch is a rule that says which observed exchanges are a step. Every
// field is optional; a field left empty is not tested.
type StepMatch struct {
	// Method is the HTTP method.
	Method string `json:"method,omitempty"`

	// Path is a template, whose segments in braces match any one segment
	// ("/pending/{id}"), or a regular expression when it starts with "^".
	Path string `json:"path,omitempty"`

	// Status is the response status.
	Status int `json:"status,omitempty"`

	// Headers maps a header name, compared without regard to case, to a
	// value: exact, a prefix when it ends with "*", or a regular expression
	// when it starts with "^". Response headers are searched before request
	// headers.
	Headers map[string]string `json:"headers,omitempty"`

	// Body maps a JSON pointer into the request body to the value it must
	// have, or to true when the member must merely exist.
	Body map[string]json.RawMessage `json:"body,omitempty"`

	// Repeat lets the step take every consecutive exchange that matches it,
	// such as a client polling.
	Repeat bool `json:"repeat,omitempty"`

	// Optional lets a run omit the step without failing a check.
	Optional bool `json:"optional,omitempty"`
}

// Validate checks the rule's regular expressions, status and JSON pointers.
func (m StepMatch) Validate(field string) ValidationErrors {
	var errs ValidationErrors
	if strings.HasPrefix(m.Path, "^") {
		if _, err := regexp.Compile(m.Path); err != nil {
			errs = append(errs, ValidationError{Field: field + ".path", Message: fmt.Sprintf("invalid regular expression: %v", err)})
		}
	}
	if m.Status != 0 && (m.Status < 100 || m.Status > 599) {
		errs = append(errs, ValidationError{Field: field + ".status", Message: "must be an HTTP status from 100 to 599"})
	}
	for _, name := range sortedKeys(m.Headers) {
		if v := m.Headers[name]; strings.HasPrefix(v, "^") {
			if _, err := regexp.Compile(v); err != nil {
				errs = append(errs, ValidationError{Field: fmt.Sprintf("%s.headers[%s]", field, name), Message: fmt.Sprintf("invalid regular expression: %v", err)})
			}
		}
	}
	for _, ptr := range sortedKeys(m.Body) {
		if ptr != "" && !strings.HasPrefix(ptr, "/") {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("%s.body[%s]", field, ptr), Message: `a JSON pointer starts with "/"`})
		}
		if !json.Valid(m.Body[ptr]) {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("%s.body[%s]", field, ptr), Message: "must be a JSON value"})
		}
	}
	return errs
}

// ParseBinding parses a binding. It refuses another major version of the
// format.
func ParseBinding(data []byte) (*Binding, error) {
	var b Binding
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	if b.Format != BindingFormat {
		return nil, fmt.Errorf("binding format %q is not %s", b.Format, BindingFormat)
	}
	return &b, nil
}

// ReadBinding reads a binding file.
func ReadBinding(path string) (*Binding, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- the path is chosen by the caller
	if err != nil {
		return nil, err
	}
	b, err := ParseBinding(data)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return b, nil
}

// WriteFile writes the binding as indented JSON.
func (b *Binding) WriteFile(path string) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// StepRule returns the binding's rule for a step of a scenario: the
// scenario's own key if present, else the wildcard key.
func (b *Binding) StepRule(scenario, step string) (StepMatch, bool) {
	if s, ok := b.Steps[scenario+"/"+step]; ok {
		return s.Match, true
	}
	if s, ok := b.Steps["*/"+step]; ok {
		return s.Match, true
	}
	return StepMatch{}, false
}

// Validate checks the binding's structure: its format, regular
// expressions, step keys and the participants its rules name.
func (b *Binding) Validate() ValidationErrors {
	var errs ValidationErrors
	if b.Format != BindingFormat {
		errs = append(errs, ValidationError{Field: "format", Message: fmt.Sprintf("must be %s", BindingFormat)})
	}
	if len(b.Participants) == 0 {
		errs = append(errs, ValidationError{Field: "participants", Message: "required"})
	}
	for _, id := range sortedKeys(b.Participants) {
		p := b.Participants[id]
		for i, h := range p.Hosts {
			if h == "" || strings.ContainsAny(h, "/:") {
				errs = append(errs, ValidationError{Field: fmt.Sprintf("participants[%s].hosts[%d]", id, i), Message: "must be a host name without scheme, port or path"})
			}
		}
	}
	for i, r := range b.LabelsByPath {
		field := fmt.Sprintf("labelsByPath[%d]", i)
		if r.Label == "" {
			errs = append(errs, ValidationError{Field: field + ".label", Message: "required"})
		}
		if _, err := regexp.Compile(r.Path); err != nil || r.Path == "" {
			errs = append(errs, ValidationError{Field: field + ".path", Message: "must be a regular expression"})
		}
		if _, ok := b.Participants[r.Participant]; !ok {
			errs = append(errs, ValidationError{Field: field + ".participant", Message: fmt.Sprintf("%q is not in participants", r.Participant)})
		}
	}
	for _, id := range sortedKeys(b.Paths) {
		if _, ok := b.Participants[id]; !ok {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("paths[%s]", id), Message: "is not in participants"})
		}
		for _, key := range sortedKeys(b.Paths[id]) {
			if !validStepKey(key) {
				errs = append(errs, ValidationError{Field: fmt.Sprintf("paths[%s][%s]", id, key), Message: `must be "<scenario>/<step-id>" or "*/<step-id>"`})
			}
		}
	}
	for i, r := range b.Rewrite {
		if _, err := regexp.Compile(r.Match); err != nil || r.Match == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("rewrite[%d].match", i), Message: "must be a regular expression"})
		}
	}
	for _, key := range sortedKeys(b.Steps) {
		field := fmt.Sprintf("steps[%s]", key)
		if !validStepKey(key) {
			errs = append(errs, ValidationError{Field: field, Message: `must be "<scenario>/<step-id>" or "*/<step-id>"`})
		}
		errs = append(errs, b.Steps[key].Match.Validate(field+".match")...)
	}
	return errs
}

// validStepKey reports whether key is "<scenario>/<step-id>" or
// "*/<step-id>", with neither part empty.
func validStepKey(key string) bool {
	scenario, step, ok := strings.Cut(key, "/")
	return ok && scenario != "" && step != "" && !strings.Contains(step, "/")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
