// Package observe shows runs observed in real systems in walkthroughs.
//
// Apply matches an observation to a walkthrough's scenarios and writes each
// scenario instance found as its own scenario entry. The entry keeps every
// step's explanation and example and adds the exchange that was observed,
// with its tokens decoded but not verified, the step's status, and the
// exchanges that matched no step, each after the step it followed. Raw
// token, signature and credential values are masked unless asked for.
package observe

import (
	"fmt"
	"time"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/observe/match"
)

// Options controls matching and output.
type Options struct {
	// Scenarios names the scenarios to look for, in order; empty means
	// every scenario that is not a fragment.
	Scenarios []string
	// From and To limit matching to exchanges in a time range.
	From, To *time.Time
	// Raw keeps raw token, signature and credential values; by default
	// they are masked and decoded claims are kept.
	Raw bool
}

// Apply matches o to w's scenarios and adds an entry for each scenario
// instance found, right after the scenario it is an instance of. p supplies
// the token definitions; b may be nil.
func Apply(p *pidl.Protocol, w *pidl.Walkthrough, o *pidl.Observation, b *pidl.Binding, opts Options) (*match.Report, error) {
	report, err := match.Match(w, o, b, match.Options{Scenarios: opts.Scenarios, From: opts.From, To: opts.To})
	if err != nil {
		return nil, err
	}
	resolver, err := match.NewResolver(b)
	if err != nil {
		return nil, err
	}
	tokens := newTokenFinder(p)
	var masker *masker
	if !opts.Raw {
		masker = newMasker()
	}

	entries := map[string][]pidl.WalkthroughScenario{}
	for n, f := range report.Instances {
		base := scenarioByID(w, f.Scenario)
		if base == nil {
			return nil, fmt.Errorf("no scenario %q", f.Scenario)
		}
		entries[f.Scenario] = append(entries[f.Scenario],
			entry(*base, f, n+1, len(report.Instances), o, resolver, tokens, masker))
	}
	if len(entries) == 0 {
		return report, nil
	}
	var out []pidl.WalkthroughScenario
	for _, sc := range w.Scenarios {
		out = append(out, sc)
		out = append(out, entries[sc.ID]...)
	}
	for i := range out {
		out[i].Position = i + 1
	}
	w.Scenarios = out
	return report, nil
}

func scenarioByID(w *pidl.Walkthrough, id string) *pidl.WalkthroughScenario {
	for i := range w.Scenarios {
		if w.Scenarios[i].ID == id {
			return &w.Scenarios[i]
		}
	}
	return nil
}

// entry builds the scenario entry for one instance.
func entry(base pidl.WalkthroughScenario, f match.Found, n, of int, o *pidl.Observation, resolver *match.Resolver, tokens *tokenFinder, masker *masker) pidl.WalkthroughScenario {
	sc := base
	sc.ID = fmt.Sprintf("%s.observed-%d", base.ID, n)
	sc.Name = "Observed: " + base.Name
	if of > 1 {
		sc.Name = fmt.Sprintf("Observed %d of %d: %s", n, of, base.Name)
	}
	sc.Description = fmt.Sprintf("%s, as a real run performed it: the exchanges observed in a capture, matched to its steps. Tokens are decoded, not verified.", base.Name)
	sc.Fragment, sc.VariantOf, sc.VariantLabel = false, "", ""
	sc.Steps = append([]pidl.WalkthroughStep(nil), base.Steps...)

	run := &pidl.ObservedRun{
		Scenario:        base.ID,
		Source:          o.Source.Tool,
		RequestFidelity: o.Source.RequestFidelity,
		Instance:        n,
		Instances:       of,
		RunnerUp:        f.RunnerUp,
		Matched:         f.Matched,
		Weak:            f.Weak,
		NotSeen:         f.NotSeen,
		Skipped:         f.Skipped,
		Unmatchable:     f.Unmatchable,
		Unobservable:    f.Unobservable,
		Explained:       f.Explained,
		Extra:           len(f.Extra),
		Redacted:        masker != nil,
	}

	paired := map[int]bool{}
	for _, r := range f.Steps {
		if r.Step.Kind == match.KindResponse {
			paired[r.Step.PairedWith] = true
		}
	}
	for i, r := range f.Steps {
		ws := &sc.Steps[i]
		ws.ObservedStatus = string(r.Status)
		if r.Status != match.StatusMatched || len(r.Exchanges) == 0 {
			continue
		}
		ws.ObservedStrength = string(r.Strength)
		// A repeated step shows its last exchange, which for a poll is the
		// answer that ended it.
		x := &o.Exchanges[r.Exchanges[len(r.Exchanges)-1]]
		var msg pidl.ObservedMessage
		switch r.Step.Kind {
		case match.KindResponse:
			msg = message(x, false, true, tokens)
		default:
			msg = message(x, true, !paired[i], tokens)
		}
		if len(r.Exchanges) > 1 {
			msg.Repeats = len(r.Exchanges)
		}
		masker.mask(&msg, tokens.values(x))
		ws.Observed = &msg
		widen(run, x.Time)
	}
	for _, e := range f.Extra {
		x := &o.Exchanges[e.Exchange]
		res := resolver.Resolve(e.Exchange, x)
		msg := message(x, true, true, tokens)
		masker.mask(&msg, tokens.values(x))
		sc.ObservedExtra = append(sc.ObservedExtra, pidl.ObservedExtra{AfterStep: e.AfterStep + 1, From: res.From, To: res.To, Observed: msg})
		widen(run, x.Time)
	}
	sc.ObservedRun = run
	return sc
}

// widen extends the run's time range to t.
func widen(run *pidl.ObservedRun, t *time.Time) {
	if t == nil {
		return
	}
	if run.From == nil || t.Before(*run.From) {
		run.From = t
	}
	if run.To == nil || t.After(*run.To) {
		run.To = t
	}
}

// message shows an exchange as a step's example is shown: its request,
// its response, or both, with the tokens each carries.
func message(x *pidl.Exchange, withRequest, withResponse bool, tokens *tokenFinder) pidl.ObservedMessage {
	m := pidl.ObservedMessage{
		Time: x.Time, From: x.From, To: x.To, RequestID: x.RequestID,
		Label: x.Label, Annotations: x.Annotations, BodyTruncated: x.BodyTruncated,
	}
	if withRequest {
		m.Request = &pidl.ExampleRequest{Method: x.Request.Method, URL: x.Request.URL, Headers: exampleHeaders(x.Request.Headers), Body: x.Request.Body}
		m.Tokens = append(m.Tokens, tokens.find(x.Request.Headers, x.Request.Body, "request")...)
	}
	if withResponse {
		m.Response = &pidl.ExampleResponse{Status: x.Response.Status, Headers: exampleHeaders(x.Response.Headers), Body: x.Response.Body}
		m.Tokens = append(m.Tokens, tokens.find(x.Response.Headers, x.Response.Body, "response")...)
	}
	return m
}

func exampleHeaders(h pidl.ObservedHeaders) []pidl.ExampleHeader {
	if len(h) == 0 {
		return nil
	}
	out := make([]pidl.ExampleHeader, len(h))
	for i, kv := range h {
		out[i] = pidl.ExampleHeader{Name: kv[0], Value: kv[1]}
	}
	return out
}
