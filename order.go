package pidl

import "sort"

// HasScenarios reports whether the document uses named scenarios rather
// than a single top-level flows list.
func (p *Protocol) HasScenarios() bool {
	return len(p.Scenarios) > 0
}

// ScenarioByID returns the scenario with the given ID, or nil. It also
// finds the fragments a collapsed projection carries in Fragments.
func (p *Protocol) ScenarioByID(id string) *Scenario {
	for i := range p.Scenarios {
		if p.Scenarios[i].ID == id {
			return &p.Scenarios[i]
		}
	}
	for i := range p.Fragments {
		if p.Fragments[i].ID == id {
			return &p.Fragments[i]
		}
	}
	return nil
}

// ScenarioIDs returns scenario IDs in document order. Fragments are
// included only when includeFragments is true.
func (p *Protocol) ScenarioIDs(includeFragments bool) []string {
	var ids []string
	for _, s := range p.Scenarios {
		if s.Fragment && !includeFragments {
			continue
		}
		ids = append(ids, s.ID)
	}
	return ids
}

// EntityIndex returns the position of entity id in Entities, or -1.
func (p *Protocol) EntityIndex(id string) int {
	for i, e := range p.Entities {
		if e.ID == id {
			return i
		}
	}
	return -1
}

// OrderEntities returns ids sorted by their position in Entities, with
// duplicates removed. IDs that are not entities sort last, in their
// original relative order.
func (p *Protocol) OrderEntities(ids []string) []string {
	seen := make(map[string]bool, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	rank := func(id string) int {
		if i := p.EntityIndex(id); i >= 0 {
			return i
		}
		return len(p.Entities)
	}
	sort.SliceStable(out, func(i, j int) bool { return rank(out[i]) < rank(out[j]) })
	return out
}

// IsInclude reports whether the flow is an include step.
func (f *Flow) IsInclude() bool {
	return f.Include != ""
}
