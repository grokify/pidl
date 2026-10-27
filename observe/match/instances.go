package match

import (
	"fmt"
	"slices"
	"time"

	"github.com/grokify/pidl"
)

// Options controls how a capture is split into scenario instances.
type Options struct {
	// Scenarios names the scenarios to look for, in order. When empty,
	// every scenario that is not a fragment is a candidate at each point.
	Scenarios []string
	// From and To limit matching to exchanges in a time range. Exchanges
	// without a time are kept.
	From, To *time.Time
}

// Rank orders candidate instances: exchanges explained minus required
// steps not seen, then strong matches, then total score.
type Rank struct {
	Net, Strong, Score int
}

// Less reports whether r ranks below o.
func (r Rank) Less(o Rank) bool {
	if r.Net != o.Net {
		return r.Net < o.Net
	}
	if r.Strong != o.Strong {
		return r.Strong < o.Strong
	}
	return r.Score < o.Score
}

// Found is one scenario instance found in a capture.
type Found struct {
	Instance
	// Rank is the instance's rank among the candidates at its point.
	Rank Rank
	// RunnerUp is the next-best candidate scenario at that point, with its
	// rank; empty when there was none.
	RunnerUp     string
	RunnerUpRank Rank
}

// Report is the outcome of matching a capture to a walkthrough.
type Report struct {
	// Instances are the scenario instances found, in order.
	Instances []Found
	// Outside are the positions of exchanges in the window that fall in no
	// instance.
	Outside []int
	// LeftOut is the number of exchanges outside the time window.
	LeftOut int
	// Warnings are notes for the document's or binding's author, each once,
	// however many scenarios include the step.
	Warnings []Warning
}

// Match splits an observation into the scenario instances it contains and
// aligns each. A capture usually holds more than one journey: an instance
// starts at the first exchange that matches a candidate's first matchable
// step, and runs until the exchange before the next one that matches the
// first matchable step of any candidate, so a new journey begins where a
// first step recurs. At each point the best-ranked candidate is taken.
// b may be nil.
func Match(w *pidl.Walkthrough, o *pidl.Observation, b *pidl.Binding, opts Options) (*Report, error) {
	resolver, err := NewResolver(b)
	if err != nil {
		return nil, err
	}
	report := &Report{}
	var exchanges []Resolved
	for i := range o.Exchanges {
		e := &o.Exchanges[i]
		if e.Time != nil && ((opts.From != nil && e.Time.Before(*opts.From)) || (opts.To != nil && e.Time.After(*opts.To))) {
			report.LeftOut++
			continue
		}
		exchanges = append(exchanges, resolver.Resolve(i, e))
	}

	byID := map[string]*pidl.WalkthroughScenario{}
	var auto []string
	for i := range w.Scenarios {
		sc := &w.Scenarios[i]
		byID[sc.ID] = sc
		if !sc.Fragment {
			auto = append(auto, sc.ID)
		}
	}
	candidates := auto
	if len(opts.Scenarios) > 0 {
		for _, id := range opts.Scenarios {
			if byID[id] == nil {
				return nil, fmt.Errorf("no scenario %q", id)
			}
		}
		candidates = opts.Scenarios
	}

	prepared := map[string][]Step{}
	prepare := func(id string) []Step {
		if steps, ok := prepared[id]; ok {
			return steps
		}
		steps, warnings := Prepare(w, byID[id], b)
		prepared[id] = steps
		for _, wn := range warnings {
			if !slices.Contains(report.Warnings, wn) {
				report.Warnings = append(report.Warnings, wn)
			}
		}
		return steps
	}
	firstStep := func(id string) (Step, bool) {
		for _, s := range prepare(id) {
			if (s.Kind == KindRequest || s.Kind == KindStandalone) && s.HasRule {
				return s, true
			}
		}
		return Step{}, false
	}
	startsJourney := func(r Resolved) bool {
		for _, id := range candidates {
			if s, ok := firstStep(id); ok && Test(s, r).Candidate() {
				return true
			}
		}
		return false
	}

	type option struct {
		id         string
		start, end int
		inst       Instance
		rank       Rank
	}
	named := 0
	pos := 0
	for pos < len(exchanges) {
		ids := candidates
		if len(opts.Scenarios) > 0 {
			if named >= len(opts.Scenarios) {
				break
			}
			ids = opts.Scenarios[named : named+1]
		}
		var options []option
		for _, id := range ids {
			first, ok := firstStep(id)
			if !ok {
				continue
			}
			start := -1
			for k := pos; k < len(exchanges); k++ {
				if Test(first, exchanges[k]).Candidate() {
					start = k
					break
				}
			}
			if start < 0 {
				continue
			}
			end := len(exchanges)
			for k := start + 1; k < len(exchanges); k++ {
				if startsJourney(exchanges[k]) {
					end = k
					break
				}
			}
			steps := prepare(id)
			inst := Align(id, steps, exchanges[start:end])
			options = append(options, option{id: id, start: start, end: end, inst: inst, rank: rank(inst)})
		}
		if len(options) == 0 {
			break
		}
		slices.SortStableFunc(options, func(a, b option) int {
			switch {
			case b.rank.Less(a.rank):
				return -1
			case a.rank.Less(b.rank):
				return 1
			}
			return a.start - b.start
		})
		best := options[0]
		// An instance must explain more exchanges than it leaves required
		// steps unseen. Otherwise its first exchange, such as a health check
		// before a run, falls in no instance and the search moves past it.
		if best.rank.Net <= 0 {
			for k := pos; k <= best.start; k++ {
				report.Outside = append(report.Outside, exchanges[k].Index)
			}
			pos = best.start + 1
			continue
		}
		found := Found{Instance: best.inst, Rank: best.rank}
		if len(options) > 1 {
			found.RunnerUp, found.RunnerUpRank = options[1].id, options[1].rank
		}
		for k := pos; k < best.start; k++ {
			report.Outside = append(report.Outside, exchanges[k].Index)
		}
		report.Instances = append(report.Instances, found)
		pos = best.end
		named++
	}
	for k := pos; k < len(exchanges); k++ {
		report.Outside = append(report.Outside, exchanges[k].Index)
	}
	return report, nil
}

// rank scores an instance by the exchanges it explains, less the required
// request and standalone steps it did not see; response halves are not
// counted twice.
func rank(inst Instance) Rank {
	missing := 0
	for _, r := range inst.Steps {
		if r.Status == StatusNotSeen && r.Step.Kind != KindResponse {
			missing++
		}
	}
	return Rank{Net: inst.Explained - missing, Strong: inst.Strong, Score: inst.Score}
}
