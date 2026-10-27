package pidl

import (
	"fmt"
	"regexp"
	"slices"
	"time"
)

// WalkthroughFormat identifies the walkthrough format and its major
// version. Within a major version, changes only add optional fields.
const WalkthroughFormat = "pidl-walkthrough/v1"

// Walkthrough is a protocol prepared for step-by-step presentation: roles
// in entities order and, for each scenario in output order, its steps with
// fragments expanded in place. Viewers render it without needing PIDL's
// include, ordering or analysis rules.
type Walkthrough struct {
	// Format names the format and its major version, "pidl-walkthrough/v1".
	Format      string `json:"format"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Language is the source language of the narration.
	Language       string                       `json:"language"`
	Pronunciations map[string]map[string]string `json:"pronunciations,omitempty"`
	Roles          []OutlineRole                `json:"roles"`
	Groups         []Group                      `json:"groups,omitempty"`
	Scenarios      []WalkthroughScenario        `json:"scenarios"`
	Tokens         []WalkthroughToken           `json:"tokens,omitempty"`
	// Findings holds findings that do not locate a single flow, such as an
	// unused role.
	Findings []Finding `json:"findings,omitempty"`
}

// WalkthroughScenario is one scenario with its expanded steps.
type WalkthroughScenario struct {
	Position     int               `json:"position"`
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Fragment     bool              `json:"fragment,omitempty"`
	Group        string            `json:"group,omitempty"`
	VariantOf    string            `json:"variant_of,omitempty"`
	VariantLabel string            `json:"variant_label,omitempty"`
	Narration    Narration         `json:"narration,omitempty"`
	Briefing     *Briefing         `json:"briefing,omitempty"`
	Samples      []Sample          `json:"samples,omitempty"`
	Participants []string          `json:"participants"`
	Steps        []WalkthroughStep `json:"steps"`

	// ObservedRun is set on a scenario entry that shows a run observed in a
	// real system: one instance of the scenario found in a capture.
	ObservedRun *ObservedRun `json:"observed_run,omitempty"`
	// ObservedExtra are the run's exchanges that matched no step, each
	// placed after the step it followed.
	ObservedExtra []ObservedExtra `json:"observed_extra,omitempty"`
}

// ObservedRun summarizes one instance of a scenario found in a capture.
type ObservedRun struct {
	// Scenario is the ID of the scenario this entry is an instance of.
	Scenario string `json:"scenario"`
	// Source names the tool that captured the run; RequestFidelity says
	// whether its requests are the wire messages or a client's.
	Source          string          `json:"source"`
	RequestFidelity RequestFidelity `json:"request_fidelity,omitempty"`
	// Instance is this instance's number, from 1, of Instances found.
	Instance  int `json:"instance"`
	Instances int `json:"instances"`
	// From and To bound the instance's exchanges in time.
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
	// RunnerUp is the next-best scenario at this point of the capture.
	RunnerUp string `json:"runner_up,omitempty"`
	// Counts of step outcomes; Matched counts every matched step and Weak
	// the matched steps that are weak. Explained counts exchanges matched
	// to steps and Extra those that matched none.
	Matched      int `json:"matched"`
	Weak         int `json:"weak"`
	NotSeen      int `json:"not_seen"`
	Skipped      int `json:"skipped"`
	Unmatchable  int `json:"unmatchable"`
	Unobservable int `json:"unobservable"`
	Explained    int `json:"explained"`
	Extra        int `json:"extra"`
	// Redacted reports that raw token, signature and credential values
	// were masked.
	Redacted bool `json:"redacted"`
}

// ObservedMessage is an exchange observed in a real system, shown as a
// step's example is: request, response and decoded tokens.
type ObservedMessage struct {
	Time *time.Time `json:"time,omitempty"`
	// From and To are the capturing tool's labels for sender and receiver.
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
	// RequestID is the value of a request ID header, when the capture has one.
	RequestID string           `json:"request_id,omitempty"`
	Request   *ExampleRequest  `json:"request,omitempty"`
	Response  *ExampleResponse `json:"response,omitempty"`
	// Tokens are the tokens found in the message, decoded and not verified.
	Tokens []ExampleToken `json:"tokens,omitempty"`
	// Label and Annotations are evidence the capturing tool recorded.
	Label       string               `json:"label,omitempty"`
	Annotations []ExchangeAnnotation `json:"annotations,omitempty"`
	// Repeats is how many exchanges a repeated step took, when more than one.
	Repeats int `json:"repeats,omitempty"`
	// BodyTruncated reports that a body was cut at the capture's size limit.
	BodyTruncated bool `json:"body_truncated,omitempty"`
}

// ObservedExtra is an observed exchange that matched no step.
type ObservedExtra struct {
	// AfterStep is the number of the step it followed, or 0 when it came
	// before the first matched step.
	AfterStep int `json:"after_step"`
	// From and To are the participants, when the binding resolved them.
	From     string          `json:"from,omitempty"`
	To       string          `json:"to,omitempty"`
	Observed ObservedMessage `json:"observed"`
}

// Observed status and strength values of a walkthrough step.
const (
	ObservedMatched      = "matched"
	ObservedNotSeen      = "notSeen"
	ObservedSkipped      = "skipped"
	ObservedUnmatchable  = "unmatchable"
	ObservedUnobservable = "unobservable"
	ObservedStrong       = "strong"
	ObservedWeak         = "weak"
)

// WalkthroughStep is one message of an expanded scenario.
type WalkthroughStep struct {
	// Number is the 1-based position within the expanded scenario.
	Number int `json:"number"`
	// ID is the step's flow ID, qualified as "<include>.<id>" for a step
	// inside an included fragment; empty when the flow has no ID.
	ID       string `json:"id,omitempty"`
	From     string `json:"from"`
	To       string `json:"to"`
	Label    string `json:"label"`
	Response bool   `json:"response,omitempty"`
	Phase    string `json:"phase,omitempty"`
	// Fragments names the fragments this step sits inside, outermost first.
	Fragments   []string  `json:"fragments,omitempty"`
	Description string    `json:"description,omitempty"`
	Narration   Narration `json:"narration,omitempty"`
	// Representations holds narration overrides by representation.
	Representations map[string]RepresentationNarration `json:"representations,omitempty"`
	Note            string                             `json:"note,omitempty"`
	SignedBy        string                             `json:"signed_by,omitempty"`
	Issues          []string                           `json:"issues,omitempty"`
	Verifies        []string                           `json:"verifies,omitempty"`
	Example         *FlowExample                       `json:"example,omitempty"`

	// Match is the step's own rule for matching observed exchanges, from
	// the document. DefinedIn is the scenario or fragment that defines the
	// step and FlowID its unqualified ID, which a binding's step keys name.
	// They serve generation in process and are not part of the format.
	Match     *StepMatch `json:"-"`
	DefinedIn string     `json:"-"`
	FlowID    string     `json:"-"`

	// Observed is the exchange a real run produced for this step, in a
	// scenario entry with an ObservedRun. ObservedStatus is matched,
	// notSeen, skipped, unmatchable or unobservable; ObservedStrength is
	// strong or weak for a matched step.
	Observed         *ObservedMessage `json:"observed,omitempty"`
	ObservedStatus   string           `json:"observed_status,omitempty"`
	ObservedStrength string           `json:"observed_strength,omitempty"`
	// Curl is a curl command equivalent to the example request.
	Curl     string    `json:"curl,omitempty"`
	Location string    `json:"location"`
	Findings []Finding `json:"findings,omitempty"`
}

// WalkthroughToken is a token definition with the claims its checkers use.
type WalkthroughToken struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       string       `json:"type,omitempty"`
	Issuer     string       `json:"issuer,omitempty"`
	Audience   string       `json:"audience,omitempty"`
	VerifiedBy []string     `json:"verified_by,omitempty"`
	Claims     []TokenClaim `json:"claims,omitempty"`
}

// Finding is an analysis result attached to a step. Location uses the
// written-flow form, such as "scenarios[1].flows[0] (consent)" or
// "flows[3]", optionally followed by a field path.
type Finding struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
}

// flowPathPattern extracts the flow path from a location string.
var flowPathPattern = regexp.MustCompile(`^(?:scenarios\[\d+\]\.)?flows\[\d+\]`)

// Walkthrough builds the walkthrough. findings, typically from the
// analyzer, are attached to the steps whose written flow they locate; a
// fragment flow's findings appear in every scenario that includes it.
func (p *Protocol) Walkthrough(findings []Finding) Walkthrough {
	byPath := map[string][]Finding{}
	var general []Finding
	for _, f := range findings {
		if path := flowPathPattern.FindString(f.Location); path != "" {
			byPath[path] = append(byPath[path], f)
		} else {
			general = append(general, f)
		}
	}

	o := p.Outline()
	w := Walkthrough{Format: WalkthroughFormat, ID: o.ID, Name: o.Name, Description: o.Description, Roles: o.Roles, Groups: p.Groups,
		Language: p.SourceLanguage(), Pronunciations: p.Pronunciations, Scenarios: []WalkthroughScenario{}, Findings: general}

	if !p.HasScenarios() {
		steps := p.walkthroughSteps(p.Flows, func(i int) string { return fmt.Sprintf("flows[%d]", i) }, "", nil, nil, byPath)
		w.Scenarios = append(w.Scenarios, WalkthroughScenario{
			Position: 1, ID: o.ID, Name: o.Name,
			Participants: p.OrderEntities(flowsEntities(p.Flows)),
			Steps:        steps,
		})
	} else {
		index := make(map[string]int, len(p.Scenarios))
		for i, sc := range p.Scenarios {
			index[sc.ID] = i
		}
		for _, s := range p.ScenarioSummaries() {
			i := index[s.ID]
			sc := p.Scenarios[i]
			w.Scenarios = append(w.Scenarios, WalkthroughScenario{
				Position: s.Position, ID: s.ID, Name: s.Name, Description: sc.Description,
				Fragment: s.Fragment, Group: s.Group, VariantOf: s.VariantOf, VariantLabel: s.VariantLabel,
				Narration: sc.Narration, Briefing: sc.Briefing, Samples: sc.Samples,
				Participants: s.Participants,
				Steps:        p.walkthroughSteps(sc.Flows, scenarioLocation(i, sc.ID), "", nil, []string{sc.ID}, byPath),
			})
		}
	}
	for i := range w.Scenarios {
		for j := range w.Scenarios[i].Steps {
			w.Scenarios[i].Steps[j].Number = j + 1
		}
	}

	if p.Metadata != nil {
		for _, t := range p.Metadata.Tokens {
			name := t.Name
			if name == "" {
				name = t.ID
			}
			claims := make([]TokenClaim, len(t.Claims))
			for k, c := range t.Claims {
				c.CheckedBy = p.OrderEntities(c.CheckedBy)
				claims[k] = c
			}
			if len(claims) == 0 {
				claims = nil
			}
			w.Tokens = append(w.Tokens, WalkthroughToken{
				ID: t.ID, Name: name, Type: t.Type, Issuer: t.Issuer, Audience: t.Audience,
				VerifiedBy: p.OrderEntities(t.VerifiedBy), Claims: claims,
			})
		}
	}
	return w
}

// scenarioLocation returns the written location for flow j of scenario i.
func scenarioLocation(i int, id string) func(int) string {
	return func(j int) string { return fmt.Sprintf("scenarios[%d].flows[%d] (%s)", i, j, id) }
}

// walkthroughSteps expands flows in order. prefix qualifies step IDs inside
// included fragments; trail holds the names of the enclosing fragments;
// seen holds scenario IDs being expanded, to stop at an include cycle
// (which validation reports).
func (p *Protocol) walkthroughSteps(flows []Flow, location func(int) string, prefix string, trail, seen []string, byPath map[string][]Finding) []WalkthroughStep {
	var steps []WalkthroughStep
	for j, f := range flows {
		if f.IsInclude() {
			if slices.Contains(seen, f.Include) {
				continue
			}
			idx := -1
			for k := range p.Scenarios {
				if p.Scenarios[k].ID == f.Include {
					idx = k
				}
			}
			if idx < 0 {
				continue
			}
			frag := p.Scenarios[idx]
			name := frag.Name
			if f.Label != "" {
				name = f.Label
			}
			inner := p.walkthroughSteps(frag.Flows, scenarioLocation(idx, frag.ID), includePrefix(prefix, f),
				append(slices.Clone(trail), name), append(slices.Clone(seen), frag.ID), byPath)
			for k := range inner {
				if inner[k].Phase == "" {
					inner[k].Phase = f.Phase
				}
			}
			steps = append(steps, inner...)
			continue
		}
		loc := location(j)
		step := WalkthroughStep{
			ID:   qualifyStepID(prefix, f.ID),
			From: f.From, To: f.To, Label: f.DisplayLabel(),
			Response:        f.EffectiveMode() == FlowModeResponse,
			Phase:           f.Phase,
			Fragments:       slices.Clone(trail),
			Description:     f.Description,
			Narration:       f.Narration,
			Representations: f.Representations,
			Note:            f.Note,
			Example:         f.Example,
			Match:           f.Match,
			FlowID:          f.ID,
			Location:        loc,
			Findings:        byPath[flowPathPattern.FindString(loc)],
		}
		if f.Example != nil && f.Example.Request != nil {
			cmd, err := f.Example.Request.Curl()
			if err != nil {
				step.Findings = append(step.Findings, Finding{
					ID: "EXAMPLE", Severity: "medium", Title: "Example request body is not valid JSON",
					Description: err.Error(), Location: loc + ".example.request.body",
				})
			}
			step.Curl = cmd
		}
		if len(seen) > 0 {
			step.DefinedIn = seen[len(seen)-1]
		}
		if f.Security != nil {
			step.SignedBy = f.Security.SignedBy
			step.Issues = f.Security.Issues
			step.Verifies = f.Security.Verifies
		}
		steps = append(steps, step)
	}
	return steps
}
