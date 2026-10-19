package pidl

// Outline is an ordered summary of a protocol for building guides and
// presentations: roles in entities order, then scenarios in output order
// (main scenarios, then fragments) with numbered steps, then tokens in
// declared order. Every list keeps document order, so material built from
// an outline matches the diagrams generated from the same file.
type Outline struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Roles       []OutlineRole     `json:"roles"`
	Groups      []Group           `json:"groups,omitempty"`
	Scenarios   []OutlineScenario `json:"scenarios"`
	Tokens      []OutlineToken    `json:"tokens,omitempty"`
}

// OutlineRole is one entity in an outline.
type OutlineRole struct {
	Position    int    `json:"position"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

// OutlineScenario is one scenario in an outline.
type OutlineScenario struct {
	Position     int           `json:"position"`
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	Fragment     bool          `json:"fragment,omitempty"`
	Group        string        `json:"group,omitempty"`
	VariantOf    string        `json:"variant_of,omitempty"`
	VariantLabel string        `json:"variant_label,omitempty"`
	Participants []string      `json:"participants"`
	Steps        []OutlineStep `json:"steps"`
}

// OutlineStep is one numbered step of a scenario: a message, or an include
// of a fragment.
type OutlineStep struct {
	Number   int    `json:"number"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Label    string `json:"label"`
	Response bool   `json:"response,omitempty"`
	Phase    string `json:"phase,omitempty"`
	// Include is the fragment ID for an include step, and IncludePosition
	// its position among the outline's scenarios.
	Include         string `json:"include,omitempty"`
	IncludePosition int    `json:"include_position,omitempty"`
}

// OutlineToken is one token definition in an outline.
type OutlineToken struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Issuer     string   `json:"issuer,omitempty"`
	Audience   string   `json:"audience,omitempty"`
	VerifiedBy []string `json:"verified_by,omitempty"`
}

// Outline returns an ordered summary of the protocol. A document without
// scenarios appears as a single scenario named after the protocol.
func (p *Protocol) Outline() Outline {
	o := Outline{
		ID:          p.ProtocolMeta.ID,
		Name:        p.ProtocolMeta.Name,
		Description: p.ProtocolMeta.Description,
		Roles:       []OutlineRole{},
		Groups:      p.Groups,
		Scenarios:   []OutlineScenario{},
	}
	for i, e := range p.Entities {
		o.Roles = append(o.Roles, OutlineRole{
			Position: i + 1, ID: e.ID, Name: e.Name, Type: string(e.Type), Description: e.Description,
		})
	}

	if !p.HasScenarios() {
		o.Scenarios = append(o.Scenarios, OutlineScenario{
			Position:     1,
			ID:           p.ProtocolMeta.ID,
			Name:         p.ProtocolMeta.Name,
			Participants: p.OrderEntities(flowsEntities(p.Flows)),
			Steps:        outlineSteps(p.Flows, nil),
		})
	} else {
		summaries := p.ScenarioSummaries()
		positions := make(map[string]int, len(summaries))
		for _, s := range summaries {
			positions[s.ID] = s.Position
		}
		for _, s := range summaries {
			sc := p.ScenarioByID(s.ID)
			o.Scenarios = append(o.Scenarios, OutlineScenario{
				Position:     s.Position,
				ID:           s.ID,
				Name:         s.Name,
				Description:  sc.Description,
				Fragment:     s.Fragment,
				Group:        s.Group,
				VariantOf:    s.VariantOf,
				VariantLabel: s.VariantLabel,
				Participants: s.Participants,
				Steps:        outlineSteps(sc.Flows, func(id string) (string, int) { return p.scenarioName(id), positions[id] }),
			})
		}
	}

	if p.Metadata != nil {
		for _, t := range p.Metadata.Tokens {
			name := t.Name
			if name == "" {
				name = t.ID
			}
			o.Tokens = append(o.Tokens, OutlineToken{
				ID: t.ID, Name: name, Issuer: t.Issuer, Audience: t.Audience,
				VerifiedBy: p.OrderEntities(t.VerifiedBy),
			})
		}
	}
	return o
}

// outlineSteps numbers flows as outline steps. lookup returns an included
// fragment's name and position; it is nil for documents without scenarios.
func outlineSteps(flows []Flow, lookup func(id string) (string, int)) []OutlineStep {
	steps := make([]OutlineStep, 0, len(flows))
	for i, f := range flows {
		step := OutlineStep{Number: i + 1, Phase: f.Phase}
		if f.IsInclude() {
			name, pos := f.Include, 0
			if lookup != nil {
				name, pos = lookup(f.Include)
			}
			step.Include = f.Include
			step.IncludePosition = pos
			step.Label = name
			if f.Label != "" {
				step.Label = f.Label
			}
		} else {
			step.From, step.To = f.From, f.To
			step.Label = f.DisplayLabel()
			step.Response = f.EffectiveMode() == FlowModeResponse
		}
		steps = append(steps, step)
	}
	return steps
}

// flowsEntities lists the entities a flow list references.
func flowsEntities(flows []Flow) []string {
	var ids []string
	for _, f := range flows {
		ids = append(ids, flowEntities(f)...)
	}
	return ids
}

// scenarioName returns a scenario's name, or its ID when unnamed or unknown.
func (p *Protocol) scenarioName(id string) string {
	if sc := p.ScenarioByID(id); sc != nil && sc.Name != "" {
		return sc.Name
	}
	return id
}
