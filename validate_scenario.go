package pidl

import (
	"fmt"
	"slices"
	"strings"
)

// validateScenarios checks scenario IDs, participants and include steps.
// Per-flow checks run in validateFlows for every scenario's flows.
func (p *Protocol) validateScenarios() ValidationErrors {
	if !p.HasScenarios() {
		return nil
	}

	var errs ValidationErrors

	if len(p.Flows) > 0 {
		errs = append(errs, ValidationError{
			Field:   "flows",
			Message: "must be empty when scenarios are present; move these flows into a scenario",
		})
	}

	entityIDs := make(map[string]bool, len(p.Entities))
	for _, e := range p.Entities {
		entityIDs[e.ID] = true
	}

	seen := make(map[string]bool, len(p.Scenarios))
	mainCount := 0
	for i, sc := range p.Scenarios {
		field := fmt.Sprintf("scenarios[%d]", i)

		switch {
		case sc.ID == "":
			errs = append(errs, ValidationError{Field: field + ".id", Message: "required"})
		case !protocolIDPattern.MatchString(sc.ID):
			errs = append(errs, ValidationError{Field: field + ".id", Message: "must match pattern ^[a-z][a-z0-9_-]*$"})
		case seen[sc.ID]:
			errs = append(errs, ValidationError{Field: field + ".id", Message: fmt.Sprintf("duplicate scenario ID %q", sc.ID)})
		}
		seen[sc.ID] = true

		if sc.Name == "" {
			errs = append(errs, ValidationError{Field: field + ".name", Message: "required"})
		}
		if len(sc.Flows) == 0 {
			errs = append(errs, ValidationError{Field: field + ".flows", Message: "must have at least 1 flow"})
		}
		if !sc.Fragment {
			mainCount++
		}

		errs = append(errs, p.validateParticipants(field, &p.Scenarios[i], entityIDs)...)
		errs = append(errs, p.validateIncludeTargets(field, &sc)...)
	}

	if mainCount == 0 {
		errs = append(errs, ValidationError{
			Field:   "scenarios",
			Message: "must have at least 1 scenario that is not a fragment",
		})
	}

	errs = append(errs, p.validateIncludeCycles()...)
	errs = append(errs, p.validateGroups()...)
	errs = append(errs, p.validateVariants()...)
	errs = append(errs, p.validateStepIDs()...)

	return errs
}

// validateGroups checks group IDs and names and that each scenario's group
// exists. Fragments are not listed in navigation, so they have no group.
func (p *Protocol) validateGroups() ValidationErrors {
	var errs ValidationErrors
	seen := make(map[string]bool, len(p.Groups))
	for i, g := range p.Groups {
		field := fmt.Sprintf("groups[%d]", i)
		switch {
		case g.ID == "":
			errs = append(errs, ValidationError{Field: field + ".id", Message: "required"})
		case !protocolIDPattern.MatchString(g.ID):
			errs = append(errs, ValidationError{Field: field + ".id", Message: "must match pattern ^[a-z][a-z0-9_-]*$"})
		case seen[g.ID]:
			errs = append(errs, ValidationError{Field: field + ".id", Message: fmt.Sprintf("duplicate group ID %q", g.ID)})
		}
		seen[g.ID] = true
		if g.Name == "" {
			errs = append(errs, ValidationError{Field: field + ".name", Message: "required"})
		}
	}
	for i, sc := range p.Scenarios {
		if sc.Group == "" {
			continue
		}
		field := fmt.Sprintf("scenarios[%d].group", i)
		switch {
		case sc.Fragment:
			errs = append(errs, ValidationError{Field: field, Message: "fragments are not listed in navigation and cannot have a group"})
		case !seen[sc.Group]:
			errs = append(errs, ValidationError{Field: field, Message: fmt.Sprintf("unknown group %q", sc.Group)})
		}
	}
	return errs
}

// validateVariants checks that a variant names a main scenario that is not
// itself a variant, sits in the same group, and has a label.
func (p *Protocol) validateVariants() ValidationErrors {
	var errs ValidationErrors
	for i, sc := range p.Scenarios {
		field := fmt.Sprintf("scenarios[%d]", i)
		if sc.VariantOf == "" {
			if sc.VariantLabel != "" {
				errs = append(errs, ValidationError{Field: field + ".variant_label", Message: "requires variant_of"})
			}
			continue
		}
		if sc.VariantLabel == "" {
			errs = append(errs, ValidationError{Field: field + ".variant_label", Message: "required with variant_of"})
		}
		base := p.ScenarioByID(sc.VariantOf)
		vf := field + ".variant_of"
		switch {
		case sc.Fragment:
			errs = append(errs, ValidationError{Field: vf, Message: "a fragment cannot be a variant"})
		case sc.VariantOf == sc.ID:
			errs = append(errs, ValidationError{Field: vf, Message: "a scenario cannot be a variant of itself"})
		case base == nil:
			errs = append(errs, ValidationError{Field: vf, Message: fmt.Sprintf("unknown scenario %q", sc.VariantOf)})
		case base.Fragment:
			errs = append(errs, ValidationError{Field: vf, Message: fmt.Sprintf("scenario %q is a fragment", sc.VariantOf)})
		case base.VariantOf != "":
			errs = append(errs, ValidationError{Field: vf, Message: fmt.Sprintf("scenario %q is itself a variant; name its base %q", sc.VariantOf, base.VariantOf)})
		case base.Group != sc.Group:
			errs = append(errs, ValidationError{Field: field + ".group", Message: fmt.Sprintf("must match the group of its base %q (%q)", base.ID, base.Group)})
		}
	}
	return errs
}

// validateStepIDs reports step IDs that repeat once fragments are expanded,
// which happens when a scenario includes the same fragment twice without
// distinct include step IDs.
func (p *Protocol) validateStepIDs() ValidationErrors {
	var errs ValidationErrors
	for i, sc := range p.Scenarios {
		if sc.Fragment {
			continue
		}
		seen := map[string]bool{}
		reported := map[string]bool{}
		for _, id := range p.expandedStepIDs(sc.Flows, "", []string{sc.ID}) {
			if id == "" {
				continue
			}
			if seen[id] && !reported[id] {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("scenarios[%d].flows", i),
					Message: fmt.Sprintf("step ID %q repeats once fragments are expanded; give each include step of the same fragment its own id", id),
				})
				reported[id] = true
			}
			seen[id] = true
		}
	}
	return errs
}

// validateParticipants checks that a declared participants list names
// known entities, once each, in entities order, and covers every entity
// the scenario uses.
func (p *Protocol) validateParticipants(field string, sc *Scenario, entityIDs map[string]bool) ValidationErrors {
	if len(sc.Participants) == 0 {
		return nil
	}

	var errs ValidationErrors
	listed := make(map[string]bool, len(sc.Participants))
	for j, id := range sc.Participants {
		pf := fmt.Sprintf("%s.participants[%d]", field, j)
		switch {
		case !entityIDs[id]:
			errs = append(errs, ValidationError{Field: pf, Message: fmt.Sprintf("unknown entity %q", id)})
		case listed[id]:
			errs = append(errs, ValidationError{Field: pf, Message: fmt.Sprintf("duplicate participant %q", id)})
		}
		listed[id] = true
	}

	if want := p.OrderEntities(sc.Participants); !slices.Equal(want, sc.Participants) && len(errs) == 0 {
		errs = append(errs, ValidationError{
			Field:   field + ".participants",
			Message: fmt.Sprintf("must follow entities order: [%s]", strings.Join(want, ", ")),
		})
	}

	var missing []string
	for _, id := range p.OrderEntities(p.usedEntities(sc, map[string]bool{})) {
		if entityIDs[id] && !listed[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		errs = append(errs, ValidationError{
			Field:   field + ".participants",
			Message: fmt.Sprintf("missing entities used by the scenario: [%s]", strings.Join(missing, ", ")),
		})
	}

	return errs
}

// validateIncludeTargets checks that each include step names an existing
// fragment scenario.
func (p *Protocol) validateIncludeTargets(field string, sc *Scenario) ValidationErrors {
	var errs ValidationErrors
	for j, f := range sc.Flows {
		if !f.IsInclude() {
			continue
		}
		ff := fmt.Sprintf("%s.flows[%d].include", field, j)
		target := p.ScenarioByID(f.Include)
		switch {
		case target == nil:
			errs = append(errs, ValidationError{Field: ff, Message: fmt.Sprintf("unknown scenario %q", f.Include)})
		case !target.Fragment:
			errs = append(errs, ValidationError{Field: ff, Message: fmt.Sprintf("scenario %q is not a fragment", f.Include)})
		}
	}
	return errs
}

// validateIncludeCycles reports each include cycle once, with its path.
func (p *Protocol) validateIncludeCycles() ValidationErrors {
	const (
		unvisited = iota
		active
		done
	)
	state := make(map[string]int, len(p.Scenarios))
	var errs ValidationErrors
	var path []string

	var visit func(id string)
	visit = func(id string) {
		sc := p.ScenarioByID(id)
		if sc == nil {
			return
		}
		state[id] = active
		path = append(path, id)
		for _, f := range sc.Flows {
			if !f.IsInclude() {
				continue
			}
			switch state[f.Include] {
			case active:
				start := slices.Index(path, f.Include)
				cycle := append(append([]string{}, path[start:]...), f.Include)
				errs = append(errs, ValidationError{
					Field:   "scenarios",
					Message: "include cycle: " + strings.Join(cycle, " -> "),
				})
			case unvisited:
				visit(f.Include)
			}
		}
		path = path[:len(path)-1]
		state[id] = done
	}

	for _, sc := range p.Scenarios {
		if state[sc.ID] == unvisited {
			visit(sc.ID)
		}
	}
	return errs
}

// validateIncludeStep rejects message fields on an include step.
func validateIncludeStep(field string, f Flow) ValidationErrors {
	var bad []string
	if f.From != "" {
		bad = append(bad, "from")
	}
	if f.To != "" {
		bad = append(bad, "to")
	}
	if f.Action != "" {
		bad = append(bad, "action")
	}
	if f.Mode != "" {
		bad = append(bad, "mode")
	}
	if len(f.Sets) > 0 {
		bad = append(bad, "sets")
	}
	if f.Security != nil {
		bad = append(bad, "security")
	}
	if len(f.Alternatives) > 0 {
		bad = append(bad, "alternatives")
	}
	if len(f.DataMappings) > 0 {
		bad = append(bad, "data_mappings")
	}
	if f.Example != nil {
		bad = append(bad, "example")
	}
	if len(bad) == 0 {
		return nil
	}
	return ValidationErrors{{
		Field:   field,
		Message: fmt.Sprintf("include step cannot set %s", strings.Join(bad, ", ")),
	}}
}
