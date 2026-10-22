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

// ScenarioIDs returns scenario IDs in output order (see OrderedScenarios).
// Fragments are included only when includeFragments is true.
func (p *Protocol) ScenarioIDs(includeFragments bool) []string {
	var ids []string
	for _, i := range p.OrderedScenarios() {
		if p.Scenarios[i].Fragment && !includeFragments {
			continue
		}
		ids = append(ids, p.Scenarios[i].ID)
	}
	return ids
}

// GroupByID returns the group with the given ID, or nil.
func (p *Protocol) GroupByID(id string) *Group {
	for i := range p.Groups {
		if p.Groups[i].ID == id {
			return &p.Groups[i]
		}
	}
	return nil
}

// OrderedScenarios returns indexes into Scenarios in output order. Main
// scenarios come first: grouped scenarios in group order, then scenarios
// without a known group, each in document order within its section, with
// every variant placed directly after its base in document order. Fragments
// follow in document order. A file without groups or variants keeps
// document order.
func (p *Protocol) OrderedScenarios() []int {
	index := make(map[string]int, len(p.Scenarios))
	for i, sc := range p.Scenarios {
		if _, dup := index[sc.ID]; !dup {
			index[sc.ID] = i
		}
	}
	// isBase reports whether scenario i is listed on its own rather than
	// under a base: it is not a variant, or its base is unusable.
	isBase := func(i int) bool {
		sc := p.Scenarios[i]
		if sc.VariantOf == "" {
			return true
		}
		b, ok := index[sc.VariantOf]
		return !ok || b == i || p.Scenarios[b].Fragment || p.Scenarios[b].VariantOf != ""
	}
	variants := map[int][]int{}
	for i, sc := range p.Scenarios {
		if !sc.Fragment && !isBase(i) {
			b := index[sc.VariantOf]
			variants[b] = append(variants[b], i)
		}
	}

	var bases []int
	placed := make(map[int]bool, len(p.Scenarios))
	addBases := func(match func(Scenario) bool) {
		for i, sc := range p.Scenarios {
			if !sc.Fragment && isBase(i) && !placed[i] && match(sc) {
				bases = append(bases, i)
				placed[i] = true
			}
		}
	}
	for _, g := range p.Groups {
		addBases(func(sc Scenario) bool { return sc.Group == g.ID })
	}
	addBases(func(Scenario) bool { return true })

	out := make([]int, 0, len(p.Scenarios))
	for _, b := range bases {
		out = append(out, b)
		out = append(out, variants[b]...)
	}
	for i, sc := range p.Scenarios {
		if sc.Fragment {
			out = append(out, i)
		}
	}
	return out
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
