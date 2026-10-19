package pidl

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
