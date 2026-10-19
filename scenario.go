package pidl

import (
	"fmt"
	"slices"
	"strings"
)

// ScenarioParticipants returns the entities shown for scenario id, in
// entities order. It is the scenario's Participants when set; otherwise
// the entities its flows use plus the participants of the fragments it
// includes. It returns nil for an unknown scenario.
func (p *Protocol) ScenarioParticipants(id string) []string {
	sc := p.ScenarioByID(id)
	if sc == nil {
		return nil
	}
	if len(sc.Participants) > 0 {
		return p.OrderEntities(sc.Participants)
	}
	return p.OrderEntities(p.usedEntities(sc, map[string]bool{}))
}

// usedEntities lists the entities a scenario's flows reference, plus the
// participants of included fragments. visited guards against include
// cycles, which validation reports separately.
func (p *Protocol) usedEntities(sc *Scenario, visited map[string]bool) []string {
	if visited[sc.ID] {
		return nil
	}
	visited[sc.ID] = true

	var ids []string
	for _, f := range sc.Flows {
		if f.IsInclude() {
			frag := p.ScenarioByID(f.Include)
			if frag == nil {
				continue
			}
			if len(frag.Participants) > 0 {
				ids = append(ids, frag.Participants...)
			} else {
				ids = append(ids, p.usedEntities(frag, visited)...)
			}
			continue
		}
		ids = append(ids, flowEntities(f)...)
	}
	return ids
}

// flowEntities returns the entities a message flow and its alternative
// paths reference.
func flowEntities(f Flow) []string {
	var ids []string
	if f.From != "" {
		ids = append(ids, f.From)
	}
	if f.To != "" {
		ids = append(ids, f.To)
	}
	for _, alt := range f.Alternatives {
		for _, af := range alt.Flows {
			ids = append(ids, flowEntities(af)...)
		}
	}
	return ids
}

// ProjectOptions controls how a scenario is projected into a single-flow
// Protocol.
type ProjectOptions struct {
	// Expand replaces include steps with the included fragment's flows,
	// recursively. When false, include steps are kept so renderers can
	// draw them as a labeled box.
	Expand bool
}

// Scenario returns a single-flow Protocol for scenario id, so code that
// works on Protocol.Flows (renderers, the executor, the debugger) can
// handle one scenario at a time.
//
// The result keeps document order: entities are filtered to the
// scenario's participants in entities order, and phases to those the
// scenario uses (with their ancestors) in phases order. Metadata is
// shared. The protocol name becomes "<protocol> — <scenario>".
//
// For a document without scenarios, Scenario("") returns a copy of p.
func (p *Protocol) Scenario(id string, opts ProjectOptions) (*Protocol, error) {
	if !p.HasScenarios() {
		if id != "" {
			return nil, fmt.Errorf("protocol %q has no scenarios", p.ProtocolMeta.ID)
		}
		return p.clone(), nil
	}
	if id == "" {
		return nil, fmt.Errorf("protocol %q has scenarios; choose one of: %s",
			p.ProtocolMeta.ID, strings.Join(p.ScenarioIDs(true), ", "))
	}
	sc := p.ScenarioByID(id)
	if sc == nil {
		return nil, fmt.Errorf("unknown scenario %q; choose one of: %s",
			id, strings.Join(p.ScenarioIDs(true), ", "))
	}

	flows := sc.Flows
	if opts.Expand {
		var err error
		flows, err = p.expandFlows(sc.Flows, []string{sc.ID})
		if err != nil {
			return nil, err
		}
	}

	result := p.clone()
	result.Scenarios = nil
	result.Flows = make([]Flow, len(flows))
	copy(result.Flows, flows)
	result.Entities = p.entitiesFor(p.ScenarioParticipants(id))
	result.Phases = p.phasesFor(flows)
	result.ProtocolMeta.Name = p.ProtocolMeta.Name + " — " + sc.Name
	if sc.Description != "" {
		result.ProtocolMeta.Description = sc.Description
	}
	return result, nil
}

// expandFlows inlines include steps. Inlined flows inherit the include
// step's phase and condition when they have none. path holds the
// scenarios being expanded, to report cycles.
func (p *Protocol) expandFlows(flows []Flow, path []string) ([]Flow, error) {
	out := make([]Flow, 0, len(flows))
	for _, f := range flows {
		if !f.IsInclude() {
			out = append(out, f)
			continue
		}
		if slices.Contains(path, f.Include) {
			return nil, fmt.Errorf("include cycle: %s -> %s", strings.Join(path, " -> "), f.Include)
		}
		frag := p.ScenarioByID(f.Include)
		if frag == nil {
			return nil, fmt.Errorf("unknown scenario %q in include", f.Include)
		}
		inner, err := p.expandFlows(frag.Flows, append(slices.Clone(path), f.Include))
		if err != nil {
			return nil, err
		}
		for _, in := range inner {
			if in.Phase == "" {
				in.Phase = f.Phase
			}
			if in.Condition == "" {
				in.Condition = f.Condition
			}
			out = append(out, in)
		}
	}
	return out, nil
}

// entitiesFor returns the entities with the given IDs, in entities order.
func (p *Protocol) entitiesFor(ids []string) []Entity {
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	var out []Entity
	for _, e := range p.Entities {
		if want[e.ID] {
			out = append(out, e)
		}
	}
	return out
}

// phasesFor returns the phases flows reference, plus their ancestors, in
// phases order.
func (p *Protocol) phasesFor(flows []Flow) []Phase {
	want := make(map[string]bool)
	var mark func(id string)
	mark = func(id string) {
		if id == "" || want[id] {
			return
		}
		want[id] = true
		if ph := p.PhaseByID(id); ph != nil {
			mark(ph.Parent)
		}
	}
	for _, f := range flows {
		mark(f.Phase)
		for _, alt := range f.Alternatives {
			for _, af := range alt.Flows {
				mark(af.Phase)
			}
		}
	}
	var out []Phase
	for _, ph := range p.Phases {
		if want[ph.ID] {
			out = append(out, ph)
		}
	}
	return out
}
