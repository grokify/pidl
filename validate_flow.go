package pidl

import "fmt"

func (p *Protocol) validatePhases() ValidationErrors {
	var errs ValidationErrors

	seen := make(map[string]bool)
	for i, ph := range p.Phases {
		field := fmt.Sprintf("phases[%d]", i)

		if ph.ID == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".id",
				Message: "required",
			})
		} else {
			if !idPattern.MatchString(ph.ID) {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: "must match pattern ^[a-z][a-z0-9_]*$",
				})
			}
			if seen[ph.ID] {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: fmt.Sprintf("duplicate ID %q", ph.ID),
				})
			}
			seen[ph.ID] = true
		}

		if ph.Name == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".name",
				Message: "required",
			})
		}
	}

	// Validate parent references after collecting all IDs
	for i, ph := range p.Phases {
		if ph.Parent != "" {
			field := fmt.Sprintf("phases[%d]", i)
			if !seen[ph.Parent] {
				errs = append(errs, ValidationError{
					Field:   field + ".parent",
					Message: fmt.Sprintf("unknown parent phase %q", ph.Parent),
				})
			}
			if ph.Parent == ph.ID {
				errs = append(errs, ValidationError{
					Field:   field + ".parent",
					Message: "phase cannot be its own parent",
				})
			}
		}
	}

	// Check for circular references in phase hierarchy
	errs = append(errs, p.validatePhaseHierarchy()...)

	return errs
}

func (p *Protocol) validatePhaseHierarchy() ValidationErrors {
	var errs ValidationErrors

	for _, ph := range p.Phases {
		if ph.Parent == "" {
			continue
		}
		// Walk up the hierarchy to detect cycles
		visited := make(map[string]bool)
		current := &ph
		for current != nil && current.Parent != "" {
			if visited[current.ID] {
				errs = append(errs, ValidationError{
					Field:   "phases",
					Message: fmt.Sprintf("circular reference in phase hierarchy involving %q", ph.ID),
				})
				break
			}
			visited[current.ID] = true
			current = p.PhaseByID(current.Parent)
		}
	}

	return errs
}

// flowList is one ordered list of flows with the field path used in
// validation messages.
type flowList struct {
	field     string // e.g. "flows" or "scenarios[2].flows"
	flows     []Flow
	scenarios bool // true when the list belongs to a scenario
}

// flowLists returns every flow list in the document: the top-level flows
// for a legacy document, or each scenario's flows, in document order.
func (p *Protocol) flowLists() []flowList {
	if !p.HasScenarios() {
		return []flowList{{field: "flows", flows: p.Flows}}
	}
	lists := make([]flowList, 0, len(p.Scenarios))
	for i, sc := range p.Scenarios {
		lists = append(lists, flowList{
			field:     fmt.Sprintf("scenarios[%d].flows", i),
			flows:     sc.Flows,
			scenarios: true,
		})
	}
	return lists
}

func (p *Protocol) validateFlows() ValidationErrors {
	var errs ValidationErrors

	if !p.HasScenarios() && len(p.Flows) < 1 {
		errs = append(errs, ValidationError{
			Field:   "flows",
			Message: "must have at least 1 flow",
		})
	}

	entityIDs := make(map[string]bool)
	for _, e := range p.Entities {
		entityIDs[e.ID] = true
	}

	phaseIDs := make(map[string]bool)
	for _, ph := range p.Phases {
		phaseIDs[ph.ID] = true
	}

	for _, list := range p.flowLists() {
		errs = append(errs, validateFlowList(list, entityIDs, phaseIDs)...)
	}

	return errs
}

func validateFlowList(list flowList, entityIDs, phaseIDs map[string]bool) ValidationErrors {
	var errs ValidationErrors

	ids := make(map[string]bool, len(list.flows))
	for i, f := range list.flows {
		field := fmt.Sprintf("%s[%d]", list.field, i)

		if f.ID != "" {
			switch {
			case !protocolIDPattern.MatchString(f.ID):
				errs = append(errs, ValidationError{Field: field + ".id", Message: "must match pattern ^[a-z][a-z0-9_-]*$"})
			case ids[f.ID]:
				errs = append(errs, ValidationError{Field: field + ".id", Message: fmt.Sprintf("duplicate flow ID %q", f.ID)})
			}
			ids[f.ID] = true
		}

		if f.IsInclude() {
			if !list.scenarios {
				errs = append(errs, ValidationError{
					Field:   field + ".include",
					Message: "include steps are only allowed in scenarios",
				})
			}
			errs = append(errs, validateIncludeStep(field, f)...)
			if f.Phase != "" && !phaseIDs[f.Phase] {
				errs = append(errs, ValidationError{
					Field:   field + ".phase",
					Message: fmt.Sprintf("unknown phase %q", f.Phase),
				})
			}
			continue
		}

		if f.From == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".from",
				Message: "required",
			})
		} else if !entityIDs[f.From] {
			errs = append(errs, ValidationError{
				Field:   field + ".from",
				Message: fmt.Sprintf("unknown entity %q", f.From),
			})
		}

		if f.To == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".to",
				Message: "required",
			})
		} else if !entityIDs[f.To] {
			errs = append(errs, ValidationError{
				Field:   field + ".to",
				Message: fmt.Sprintf("unknown entity %q", f.To),
			})
		}

		if f.Action == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".action",
				Message: "required",
			})
		}

		if f.Mode != "" && !IsValidFlowMode(f.Mode) {
			errs = append(errs, ValidationError{
				Field:   field + ".mode",
				Message: fmt.Sprintf("invalid flow mode %q", f.Mode),
			})
		}

		if f.Phase != "" && !phaseIDs[f.Phase] {
			errs = append(errs, ValidationError{
				Field:   field + ".phase",
				Message: fmt.Sprintf("unknown phase %q", f.Phase),
			})
		}

		if f.Sequence < 0 {
			errs = append(errs, ValidationError{
				Field:   field + ".sequence",
				Message: "must be non-negative",
			})
		}

		// Validate annotations
		for j, ann := range f.Annotations {
			annField := fmt.Sprintf("%s.annotations[%d]", field, j)
			if ann.Type == "" {
				errs = append(errs, ValidationError{
					Field:   annField + ".type",
					Message: "required",
				})
			} else if !IsValidAnnotationType(ann.Type) {
				errs = append(errs, ValidationError{
					Field:   annField + ".type",
					Message: fmt.Sprintf("invalid annotation type %q", ann.Type),
				})
			}
			if ann.Text == "" {
				errs = append(errs, ValidationError{
					Field:   annField + ".text",
					Message: "required",
				})
			}
		}

		// Validate alternatives
		for j, alt := range f.Alternatives {
			altField := fmt.Sprintf("%s.alternatives[%d]", field, j)
			if alt.Condition == "" {
				errs = append(errs, ValidationError{
					Field:   altField + ".condition",
					Message: "required",
				})
			}
			if len(alt.Flows) == 0 {
				errs = append(errs, ValidationError{
					Field:   altField + ".flows",
					Message: "must have at least 1 flow",
				})
			}
			// Validate entity references in alternative flows
			for k, altFlow := range alt.Flows {
				altFlowField := fmt.Sprintf("%s.flows[%d]", altField, k)
				if altFlow.From != "" && !entityIDs[altFlow.From] {
					errs = append(errs, ValidationError{
						Field:   altFlowField + ".from",
						Message: fmt.Sprintf("unknown entity %q", altFlow.From),
					})
				}
				if altFlow.To != "" && !entityIDs[altFlow.To] {
					errs = append(errs, ValidationError{
						Field:   altFlowField + ".to",
						Message: fmt.Sprintf("unknown entity %q", altFlow.To),
					})
				}
				if altFlow.Phase != "" && !phaseIDs[altFlow.Phase] {
					errs = append(errs, ValidationError{
						Field:   altFlowField + ".phase",
						Message: fmt.Sprintf("unknown phase %q", altFlow.Phase),
					})
				}
			}
		}
	}

	return errs
}

func (p *Protocol) validateStateMutations() ValidationErrors {
	var errs ValidationErrors

	// Build entity lookup map
	entityMap := make(map[string]*Entity)
	for i := range p.Entities {
		entityMap[p.Entities[i].ID] = &p.Entities[i]
	}

	for _, list := range p.flowLists() {
		errs = append(errs, validateFlowListStateMutations(list, entityMap)...)
	}

	return errs
}

func validateFlowListStateMutations(list flowList, entityMap map[string]*Entity) ValidationErrors {
	var errs ValidationErrors

	for i, f := range list.flows {
		for j, m := range f.Sets {
			field := fmt.Sprintf("%s[%d].sets[%d]", list.field, i, j)

			if m.Entity == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".entity",
					Message: "required",
				})
				continue
			}

			if m.To == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".to",
					Message: "required",
				})
				continue
			}

			entity := entityMap[m.Entity]
			if entity == nil {
				errs = append(errs, ValidationError{
					Field:   field + ".entity",
					Message: fmt.Sprintf("unknown entity %q", m.Entity),
				})
				continue
			}

			if !entity.HasStates() {
				errs = append(errs, ValidationError{
					Field:   field + ".entity",
					Message: fmt.Sprintf("entity %q has no states defined", m.Entity),
				})
				continue
			}

			if entity.StateByID(m.To) == nil {
				errs = append(errs, ValidationError{
					Field:   field + ".to",
					Message: fmt.Sprintf("unknown state %q for entity %q", m.To, m.Entity),
				})
			}

			if m.From != "" && entity.StateByID(m.From) == nil {
				errs = append(errs, ValidationError{
					Field:   field + ".from",
					Message: fmt.Sprintf("unknown state %q for entity %q", m.From, m.Entity),
				})
			}
		}
	}

	return errs
}
