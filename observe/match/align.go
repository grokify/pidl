package match

import "slices"

// Score weights. A match also earns matchBonus, so the alignment prefers
// matching an exchange to leaving it out.
const (
	WeightPath   = 3
	WeightStatus = 2
	WeightHeader = 2
	WeightBody   = 2
	WeightSender = 1
	matchBonus   = 1
)

// Status is a step's result.
type Status string

const (
	// StatusMatched means an exchange was matched to the step.
	StatusMatched Status = "matched"
	// StatusNotSeen means a required step had no matching exchange.
	StatusNotSeen Status = "notSeen"
	// StatusSkipped means an optional step had no matching exchange.
	StatusSkipped Status = "skipped"
	// StatusUnmatchable means the step has no rule to match by.
	StatusUnmatchable Status = "unmatchable"
	// StatusUnobservable means the step involves a participant a capture
	// cannot see.
	StatusUnobservable Status = "unobservable"
)

// Strength qualifies a match.
type Strength string

const (
	// Strong means a path, status, header or body test passed.
	Strong Strength = "strong"
	// Weak means only participants and method agreed.
	Weak Strength = "weak"
)

// Score returns the score of an exchange that passed a step's required
// tests, and whether the match is strong.
func Score(t Tests, s Step) (int, Strength) {
	score := 0
	if t.Path == Passed {
		score += WeightPath
	}
	if t.Status == Passed {
		score += WeightStatus
	}
	if t.Headers == Passed {
		score += WeightHeader * len(s.Rule.Headers)
	}
	if t.Body == Passed {
		score += WeightBody * len(s.Rule.Body)
	}
	strength := Weak
	if score > 0 {
		strength = Strong
	}
	if t.SenderKnown {
		score += WeightSender
	}
	return score, strength
}

// StepResult is a step's outcome in one scenario instance.
type StepResult struct {
	Step     Step
	Status   Status
	Strength Strength
	// Exchanges are the positions, in the observation, of the exchanges
	// matched to the step: one, several for a repeated step, and for a
	// response step its request step's.
	Exchanges []int
	// Score is the match's score, before the match bonus.
	Score int
}

// Extra is an exchange that matched no step.
type Extra struct {
	// Exchange is its position in the observation.
	Exchange int
	// AfterStep is the index of the scenario step it followed, or -1 when
	// it came before the first matched step.
	AfterStep int
}

// Instance is the outcome of aligning a run of exchanges to one scenario.
type Instance struct {
	Scenario string
	Steps    []StepResult
	Extra    []Extra
	// Score is the alignment's total score, match bonuses included.
	Score int
	// Explained is the number of exchanges matched to steps.
	Explained int
	// Counts of step outcomes. Strong and Weak count matched steps.
	Matched, Strong, Weak, NotSeen, Skipped, Unmatchable, Unobservable int
}

// Align assigns exchanges to the steps of one scenario instance. It keeps
// order, uses each exchange once and gives each step at most one exchange
// (more for a repeated step), maximizing the total score by dynamic
// programming over the two sequences; ties prefer the earlier exchange.
// Response steps take their request step's exchange. exchanges are the
// instance's exchanges, in order.
func Align(scenario string, steps []Step, exchanges []Resolved) Instance {
	var alignable []int
	for i, s := range steps {
		if (s.Kind == KindRequest || s.Kind == KindStandalone) && s.HasRule {
			alignable = append(alignable, i)
		}
	}
	n, m := len(exchanges), len(alignable)

	type pair struct {
		score    int
		strength Strength
		ok       bool
	}
	cand := make([][]pair, n)
	for i := range exchanges {
		cand[i] = make([]pair, m)
		for j, si := range alignable {
			t := Test(steps[si], exchanges[i])
			if t.Candidate() {
				score, strength := Score(t, steps[si])
				cand[i][j] = pair{score: score, strength: strength, ok: true}
			}
		}
	}

	const (
		moveNone = iota
		moveSkipExchange
		moveSkipStep
		moveMatch
	)
	best := make([][]int, n+1)
	move := make([][]int, n+1)
	for i := range best {
		best[i] = make([]int, m+1)
		move[i] = make([]int, m+1)
	}
	for i := 0; i <= n; i++ {
		for j := 0; j <= m; j++ {
			if i == 0 && j == 0 {
				continue
			}
			// Options in tie order: leaving the later exchange out first, so
			// that a step takes the earlier of two equal exchanges.
			bestVal, bestMove := -1, moveNone
			if i > 0 && best[i-1][j] > bestVal {
				bestVal, bestMove = best[i-1][j], moveSkipExchange
			}
			if i > 0 && j > 0 && cand[i-1][j-1].ok {
				if v := best[i-1][j-1] + matchBonus + cand[i-1][j-1].score; v > bestVal {
					bestVal, bestMove = v, moveMatch
				}
			}
			if j > 0 && best[i][j-1] > bestVal {
				bestVal, bestMove = best[i][j-1], moveSkipStep
			}
			best[i][j], move[i][j] = bestVal, bestMove
		}
	}

	assigned := map[int]int{} // alignable index → exchange index within exchanges
	for i, j := n, m; i > 0 || j > 0; {
		switch move[i][j] {
		case moveMatch:
			assigned[j-1] = i - 1
			i, j = i-1, j-1
		case moveSkipExchange:
			i--
		default:
			j--
		}
	}

	results := make([]StepResult, len(steps))
	used := make([]bool, n)
	for j, si := range alignable {
		x, ok := assigned[j]
		if !ok {
			continue
		}
		used[x] = true
		results[si] = StepResult{Status: StatusMatched, Strength: cand[x][j].strength, Score: cand[x][j].score, Exchanges: []int{exchanges[x].Index}}
	}
	// A repeated step takes every matching exchange between the previous
	// matched step and the next, passing over unrelated exchanges, so a
	// poll's earlier pending answers join its final one.
	for j, si := range alignable {
		x, ok := assigned[j]
		if !ok || !steps[si].Rule.Repeat {
			continue
		}
		low, high := 0, n
		for k := j - 1; k >= 0; k-- {
			if y, ok := assigned[k]; ok {
				low = y + 1
				break
			}
		}
		for k := j + 1; k < m; k++ {
			if y, ok := assigned[k]; ok {
				high = y
				break
			}
		}
		var taken []int
		for y := low; y < high; y++ {
			if y == x || (!used[y] && cand[y][j].ok) {
				used[y] = true
				taken = append(taken, exchanges[y].Index)
			}
		}
		results[si].Exchanges = taken
	}

	inst := Instance{Scenario: scenario, Score: best[n][m]}
	for i, s := range steps {
		r := results[i]
		r.Step = s
		switch {
		case s.Kind == KindUnobservable:
			r.Status = StatusUnobservable
		case s.Kind == KindResponse:
			req := results[s.PairedWith]
			if req.Status == StatusMatched {
				r.Status, r.Strength, r.Score = StatusMatched, req.Strength, req.Score
				r.Exchanges = slices.Clone(req.Exchanges)
			} else {
				r.Status = req.Status
				if r.Status == "" {
					r.Status = statusUnmatched(steps[s.PairedWith])
				}
			}
		case !s.HasRule:
			r.Status = StatusUnmatchable
		case r.Status != StatusMatched:
			r.Status = statusUnmatched(s)
		}
		results[i] = r
		inst.count(r)
	}
	inst.Steps = results

	// Extra exchanges, each placed after the last step matched before it,
	// and after that step's response when it has one.
	lastBefore := func(x int) int {
		after := -1
		for i, r := range results {
			if r.Status == StatusMatched && len(r.Exchanges) > 0 && r.Exchanges[0] < exchanges[x].Index {
				after = i
			}
		}
		return after
	}
	for x := range exchanges {
		if used[x] {
			inst.Explained++
			continue
		}
		inst.Extra = append(inst.Extra, Extra{Exchange: exchanges[x].Index, AfterStep: lastBefore(x)})
	}
	return inst
}

func statusUnmatched(s Step) Status {
	if s.Rule.Optional {
		return StatusSkipped
	}
	return StatusNotSeen
}

func (inst *Instance) count(r StepResult) {
	switch r.Status {
	case StatusMatched:
		inst.Matched++
		if r.Strength == Strong {
			inst.Strong++
		} else {
			inst.Weak++
		}
	case StatusNotSeen:
		inst.NotSeen++
	case StatusSkipped:
		inst.Skipped++
	case StatusUnmatchable:
		inst.Unmatchable++
	case StatusUnobservable:
		inst.Unobservable++
	}
}
