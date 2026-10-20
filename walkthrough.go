package pidl

import (
	"fmt"
	"regexp"
	"slices"
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
}

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
		if f.Security != nil {
			step.SignedBy = f.Security.SignedBy
			step.Issues = f.Security.Issues
			step.Verifies = f.Security.Verifies
		}
		steps = append(steps, step)
	}
	return steps
}
