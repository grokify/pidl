// Package match matches observed exchanges to the steps of a walkthrough's
// scenarios.
//
// A document describes one HTTP exchange as two steps: the request, whose
// example has the request, and the response that follows with its
// participants reversed. The matcher aligns exchanges to request steps and
// standalone steps, and fills each response step from its request step's
// exchange. Steps involving a participant a capture cannot see, such as a
// person in a browser, are unobservable.
//
// Each step has a rule: the defaults drawn from its example (method, path
// and status), then the document's own match, then the binding's, each
// replacing the fields it sets. Participants and method must agree for an
// exchange to be a candidate; path, status and headers raise the score.
package match

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/grokify/pidl"
)

// Kind classifies a step for matching.
type Kind int

const (
	// KindRequest is a step whose example has a request.
	KindRequest Kind = iota
	// KindResponse is the response half of an earlier request step: the
	// nearest one going the other way between the same participants.
	KindResponse
	// KindStandalone is any other observable step; it matches only by an
	// explicit rule.
	KindStandalone
	// KindUnobservable is a step involving a participant a capture cannot see.
	KindUnobservable
)

func (k Kind) String() string {
	switch k {
	case KindRequest:
		return "request"
	case KindResponse:
		return "response"
	case KindStandalone:
		return "standalone"
	case KindUnobservable:
		return "unobservable"
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// RoleTypeUser is the role type of a person, whose steps a server's capture
// cannot see.
const RoleTypeUser = "user"

// Step is one step of a scenario, prepared for matching.
type Step struct {
	// Index is the step's position in the scenario's Steps.
	Index int
	// Step is the walkthrough step.
	Step *pidl.WalkthroughStep
	// Kind classifies the step.
	Kind Kind
	// Rule is the step's rule; meaningful when HasRule.
	Rule pidl.StepMatch
	// HasRule reports that the step can be matched: a request step, or a
	// standalone step with a document or binding rule.
	HasRule bool
	// RuleSources names where the rule's fields came from, in the order
	// they were applied: "example", "document", "binding <key>", "binding
	// path <key>".
	RuleSources []string
	// PairedWith is the index of the request step a response step belongs
	// to, and -1 for other kinds.
	PairedWith int
}

// Warning is a note for the author of a document or binding. Scenario and
// Step name where the author fixes it: the scenario or fragment that
// defines the step, and the step's own ID.
type Warning struct {
	Scenario string
	Step     string
	Message  string
}

func (w Warning) String() string {
	return fmt.Sprintf("%s/%s: %s", w.Scenario, w.Step, w.Message)
}

// Prepare classifies a scenario's steps and resolves each one's rule. b may
// be nil.
func Prepare(w *pidl.Walkthrough, sc *pidl.WalkthroughScenario, b *pidl.Binding) ([]Step, []Warning) {
	unobservable := map[string]bool{}
	for _, r := range w.Roles {
		if r.Type == RoleTypeUser {
			unobservable[r.ID] = true
		}
	}
	if b != nil {
		for _, id := range b.Unobservable {
			unobservable[id] = true
		}
	}

	steps := make([]Step, len(sc.Steps))
	var warnings []Warning
	for i := range sc.Steps {
		ws := &sc.Steps[i]
		s := Step{Index: i, Step: ws, PairedWith: -1}
		switch {
		case unobservable[ws.From] || unobservable[ws.To]:
			s.Kind = KindUnobservable
		case ws.Response && pairable(steps, sc.Steps, i) >= 0:
			s.Kind, s.PairedWith = KindResponse, pairable(steps, sc.Steps, i)
		case ws.Example != nil && ws.Example.Request != nil:
			s.Kind = KindRequest
		default:
			s.Kind = KindStandalone
		}
		if s.Kind == KindRequest || s.Kind == KindStandalone {
			warnings = append(warnings, resolveRule(&s, sc.ID, b)...)
		}
		steps[i] = s
	}
	return steps, warnings
}

// pairable returns the request a response at index i answers: the nearest
// earlier request going the other way between the same two participants,
// as when a person server talks to an access server between the agent's
// request and its answer. It returns -1 when there is none, or when that
// request cannot be matched.
func pairable(steps []Step, ws []pidl.WalkthroughStep, i int) int {
	for j := i - 1; j >= 0; j-- {
		if !ws[j].Response && ws[j].From == ws[i].To && ws[j].To == ws[i].From {
			if steps[j].Kind == KindRequest || (steps[j].Kind == KindStandalone && steps[j].HasRule) {
				return j
			}
			return -1
		}
	}
	return -1
}

// resolveRule builds a request or standalone step's rule: the example's
// defaults, then the document's match, then the binding's step rule and
// path, each replacing the fields it sets.
func resolveRule(s *Step, scenario string, b *pidl.Binding) []Warning {
	ws := s.Step
	pathFromExample := false
	if ex := ws.Example; ex != nil && ex.Request != nil {
		s.Rule.Method = ex.Request.Method
		if u, err := url.Parse(ex.Request.URL); err == nil {
			s.Rule.Path = u.Path
			pathFromExample = u.Path != ""
		}
		if ex.Response != nil {
			s.Rule.Status = ex.Response.Status
		}
		s.HasRule = true
		s.RuleSources = append(s.RuleSources, "example")
	}
	if ws.Match != nil {
		pathFromExample = pathFromExample && ws.Match.Path == ""
		overlay(&s.Rule, *ws.Match)
		s.HasRule = true
		s.RuleSources = append(s.RuleSources, "document")
	}
	if b != nil {
		if m, key, ok := b.StepRule(scenario, ws.ID, ws.DefinedIn, ws.FlowID); ok {
			pathFromExample = pathFromExample && m.Path == ""
			overlay(&s.Rule, m)
			s.HasRule = true
			s.RuleSources = append(s.RuleSources, "binding "+key)
		}
		if p, key, ok := b.StepPath(ws.To, scenario, ws.ID, ws.DefinedIn, ws.FlowID); ok {
			s.Rule.Path = p
			pathFromExample = false
			s.RuleSources = append(s.RuleSources, "binding path "+key)
		}
	}
	if pathFromExample {
		if seg := literalSegment(s.Rule.Path); seg != "" {
			tmpl := strings.Replace(s.Rule.Path, "/"+seg, "/{id}", 1)
			where, id := ws.DefinedIn, ws.FlowID
			if where == "" || id == "" {
				where, id = scenario, ws.ID
			}
			return []Warning{{Scenario: where, Step: id, Message: fmt.Sprintf(
				"example path %s has a literal identifier %q that a real run will not repeat; give the step a match with a template such as %s", s.Rule.Path, seg, tmpl)}}
		}
	}
	return nil
}

// overlay replaces the fields of dst that src sets.
func overlay(dst *pidl.StepMatch, src pidl.StepMatch) {
	if src.Method != "" {
		dst.Method = src.Method
	}
	if src.Path != "" {
		dst.Path = src.Path
	}
	if src.Status != 0 {
		dst.Status = src.Status
	}
	if len(src.Headers) > 0 {
		dst.Headers = src.Headers
	}
	if len(src.Body) > 0 {
		dst.Body = src.Body
	}
	dst.Repeat = dst.Repeat || src.Repeat
	dst.Optional = dst.Optional || src.Optional
}

// literalSegment returns the first path segment that looks like an
// identifier: six or more characters with at least two letters and two
// digits, so "f7a3b9c" counts and "oauth2" does not.
func literalSegment(path string) string {
	for _, seg := range strings.Split(path, "/") {
		if len(seg) < 6 || strings.HasPrefix(seg, "{") {
			continue
		}
		letters, digits := 0, 0
		for _, r := range seg {
			switch {
			case unicode.IsLetter(r):
				letters++
			case unicode.IsDigit(r):
				digits++
			}
		}
		if letters >= 2 && digits >= 2 {
			return seg
		}
	}
	return ""
}

// Resolved is an exchange with its URL rewritten and its participants
// resolved through a binding.
type Resolved struct {
	// Index is the exchange's position in the observation.
	Index int
	// Exchange is the observed exchange.
	Exchange *pidl.Exchange
	// URL is the request URL after the binding's rewrites; Host and Path
	// are its parts.
	URL, Host, Path string
	// From and To are participant IDs; empty when the binding cannot
	// resolve them.
	From, To string
}

// Resolver applies a binding's rewrites and label rules to exchanges.
type Resolver struct {
	binding  *pidl.Binding
	rewrites []*regexp.Regexp
	byPath   []*regexp.Regexp
}

// NewResolver compiles a binding's regular expressions. b may be nil, in
// which case exchanges resolve by their labels alone.
func NewResolver(b *pidl.Binding) (*Resolver, error) {
	r := &Resolver{binding: b}
	if b == nil {
		return r, nil
	}
	for i, rw := range b.Rewrite {
		re, err := regexp.Compile(rw.Match)
		if err != nil {
			return nil, fmt.Errorf("binding rewrite[%d]: %w", i, err)
		}
		r.rewrites = append(r.rewrites, re)
	}
	for i, rule := range b.LabelsByPath {
		re, err := regexp.Compile(rule.Path)
		if err != nil {
			return nil, fmt.Errorf("binding labelsByPath[%d]: %w", i, err)
		}
		r.byPath = append(r.byPath, re)
	}
	return r, nil
}

// Resolve rewrites an exchange's URL and resolves its participants.
func (r *Resolver) Resolve(i int, e *pidl.Exchange) Resolved {
	out := Resolved{Index: i, Exchange: e, URL: e.Request.URL}
	for j, re := range r.rewrites {
		out.URL = re.ReplaceAllString(out.URL, r.binding.Rewrite[j].To)
	}
	if u, err := url.Parse(out.URL); err == nil {
		out.Host, out.Path = u.Hostname(), u.Path
	}
	out.To = r.participant(e.To, out.Host)
	out.From = r.participant(e.From, "")
	if out.From == "" && r.binding != nil {
		for j, rule := range r.binding.LabelsByPath {
			if rule.Label == e.From && r.byPath[j].MatchString(out.Path) {
				out.From = rule.Participant
				break
			}
		}
	}
	return out
}

// participant finds the participant a label names, then the one that
// answers at host. Without a binding, a label is taken as a participant ID.
func (r *Resolver) participant(label, host string) string {
	if r.binding == nil {
		return label
	}
	for _, id := range sortedIDs(r.binding.Participants) {
		if label != "" && slices.Contains(r.binding.Participants[id].Labels, label) {
			return id
		}
	}
	if host == "" {
		return ""
	}
	for _, id := range sortedIDs(r.binding.Participants) {
		if slices.Contains(r.binding.Participants[id].Hosts, host) {
			return id
		}
	}
	return ""
}

func sortedIDs(m map[string]pidl.BindingParticipant) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// Outcome is the result of one test of a rule.
type Outcome int

const (
	// NotTested means the rule does not name the test.
	NotTested Outcome = iota
	// Passed means the exchange passed the test.
	Passed
	// Failed means the exchange failed the test.
	Failed
)

// Tests holds the outcome of testing an exchange against a step's rule.
type Tests struct {
	// Participants and Method are the required tests.
	Participants, Method Outcome
	// SenderKnown reports that the exchange's sender resolved and agreed,
	// not merely that it was unknown.
	SenderKnown bool
	// Path, Status, Headers and Body are the scored tests.
	Path, Status, Headers, Body Outcome
}

// Candidate reports whether the exchange passed the required tests.
func (t Tests) Candidate() bool {
	return t.Participants == Passed && t.Method != Failed
}

// Test tests a resolved exchange against a step's rule. The receiver must
// be known and agree; an unknown sender agrees with any.
func Test(s Step, r Resolved) Tests {
	var t Tests
	switch {
	case r.To == "" || r.To != s.Step.To || (r.From != "" && r.From != s.Step.From):
		t.Participants = Failed
	default:
		t.Participants = Passed
		t.SenderKnown = r.From != ""
	}
	if s.Rule.Method != "" {
		t.Method = outcome(strings.EqualFold(s.Rule.Method, r.Exchange.Request.Method))
	}
	if s.Rule.Path != "" {
		t.Path = outcome(PathMatches(s.Rule.Path, r.Path))
	}
	if s.Rule.Status != 0 {
		t.Status = outcome(s.Rule.Status == r.Exchange.Response.Status)
	}
	if len(s.Rule.Headers) > 0 {
		t.Headers = Passed
		for name, want := range s.Rule.Headers {
			v, ok := r.Exchange.Response.Headers.Get(name)
			if !ok {
				v, ok = r.Exchange.Request.Headers.Get(name)
			}
			if !ok || !ValueMatches(want, v) {
				t.Headers = Failed
				break
			}
		}
	}
	if len(s.Rule.Body) > 0 {
		t.Body = outcome(BodyMatches(s.Rule.Body, r.Exchange.Request.Body))
	}
	return t
}

func outcome(ok bool) Outcome {
	if ok {
		return Passed
	}
	return Failed
}

// PathMatches reports whether path matches a rule's path: a regular
// expression when the rule starts with "^", otherwise a template whose
// segments in braces match any one non-empty segment. A trailing slash is
// ignored.
func PathMatches(rule, path string) bool {
	if strings.HasPrefix(rule, "^") {
		re, err := regexp.Compile(rule)
		return err == nil && re.MatchString(path)
	}
	want := strings.Split(strings.TrimSuffix(rule, "/"), "/")
	got := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(want) != len(got) {
		return false
	}
	for i := range want {
		if strings.HasPrefix(want[i], "{") && strings.HasSuffix(want[i], "}") {
			if got[i] == "" {
				return false
			}
			continue
		}
		if want[i] != got[i] {
			return false
		}
	}
	return true
}

// ValueMatches reports whether a header value matches a rule's value:
// exact, a prefix when the rule ends with "*", or a regular expression when
// it starts with "^".
func ValueMatches(rule, value string) bool {
	switch {
	case strings.HasPrefix(rule, "^"):
		re, err := regexp.Compile(rule)
		return err == nil && re.MatchString(value)
	case strings.HasSuffix(rule, "*"):
		return strings.HasPrefix(value, strings.TrimSuffix(rule, "*"))
	}
	return rule == value
}

// BodyMatches reports whether a request body satisfies every predicate: a
// JSON pointer to the value the member must have, or to true when it must
// merely exist. Numbers compare by value. A body that is not JSON
// satisfies no predicate.
func BodyMatches(predicates map[string]json.RawMessage, body json.RawMessage) bool {
	doc, ok := decodeJSON(body)
	if !ok {
		return false
	}
	for ptr, want := range predicates {
		got, found := Pointer(doc, ptr)
		if !found {
			return false
		}
		if string(bytes.TrimSpace(want)) == "true" {
			continue
		}
		w, ok := decodeJSON(want)
		if !ok || !jsonEqual(got, w) {
			return false
		}
	}
	return true
}

// Pointer resolves a JSON pointer (RFC 6901) in a decoded document.
func Pointer(doc any, ptr string) (any, bool) {
	if ptr == "" {
		return doc, true
	}
	if !strings.HasPrefix(ptr, "/") {
		return nil, false
	}
	cur := doc
	for _, tok := range strings.Split(ptr[1:], "/") {
		tok = strings.ReplaceAll(strings.ReplaceAll(tok, "~1", "/"), "~0", "~")
		switch v := cur.(type) {
		case map[string]any:
			next, ok := v[tok]
			if !ok {
				return nil, false
			}
			cur = next
		case []any:
			i, err := strconv.Atoi(tok)
			if err != nil || i < 0 || i >= len(v) || (len(tok) > 1 && tok[0] == '0') {
				return nil, false
			}
			cur = v[i]
		default:
			return nil, false
		}
	}
	return cur, true
}

func decodeJSON(b json.RawMessage) (any, bool) {
	if len(bytes.TrimSpace(b)) == 0 {
		return nil, false
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, false
	}
	return v, true
}

// jsonEqual compares decoded JSON values, numbers by value.
func jsonEqual(a, b any) bool {
	switch x := a.(type) {
	case json.Number:
		y, ok := b.(json.Number)
		if !ok {
			return false
		}
		if x == y {
			return true
		}
		fx, errx := x.Float64()
		fy, erry := y.Float64()
		return errx == nil && erry == nil && fx == fy
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for k, v := range x {
			if w, ok := y[k]; !ok || !jsonEqual(v, w) {
				return false
			}
		}
		return true
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !jsonEqual(x[i], y[i]) {
				return false
			}
		}
		return true
	}
	return a == b
}
