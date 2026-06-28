package pidl

import "fmt"

func (p *Protocol) validateEntities() ValidationErrors {
	var errs ValidationErrors

	if len(p.Entities) < 2 {
		errs = append(errs, ValidationError{
			Field:   "entities",
			Message: "must have at least 2 entities",
		})
	}

	seen := make(map[string]bool)
	for i, e := range p.Entities {
		field := fmt.Sprintf("entities[%d]", i)

		if e.ID == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".id",
				Message: "required",
			})
		} else {
			if !idPattern.MatchString(e.ID) {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: "must match pattern ^[a-z][a-z0-9_]*$",
				})
			}
			if seen[e.ID] {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: fmt.Sprintf("duplicate ID %q", e.ID),
				})
			}
			seen[e.ID] = true
		}

		if e.Name == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".name",
				Message: "required",
			})
		}

		if e.Type == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".type",
				Message: "required",
			})
		} else if !IsValidEntityType(e.Type) {
			errs = append(errs, ValidationError{
				Field:   field + ".type",
				Message: fmt.Sprintf("invalid entity type %q", e.Type),
			})
		}

		if e.TrustLevel != "" && !IsValidTrustLevel(e.TrustLevel) {
			errs = append(errs, ValidationError{
				Field:   field + ".trust_level",
				Message: fmt.Sprintf("invalid trust level %q", e.TrustLevel),
			})
		}
	}

	return errs
}

func (p *Protocol) validateEntityStates() ValidationErrors {
	var errs ValidationErrors

	for i, e := range p.Entities {
		if len(e.States) == 0 {
			continue
		}

		field := fmt.Sprintf("entities[%d]", i)
		seen := make(map[string]bool)
		initialCount := 0

		for j, s := range e.States {
			stateField := fmt.Sprintf("%s.states[%d]", field, j)

			if s.ID == "" {
				errs = append(errs, ValidationError{
					Field:   stateField + ".id",
					Message: "required",
				})
			} else {
				if !idPattern.MatchString(s.ID) {
					errs = append(errs, ValidationError{
						Field:   stateField + ".id",
						Message: "must match pattern ^[a-z][a-z0-9_]*$",
					})
				}
				if seen[s.ID] {
					errs = append(errs, ValidationError{
						Field:   stateField + ".id",
						Message: fmt.Sprintf("duplicate state ID %q within entity %q", s.ID, e.ID),
					})
				}
				seen[s.ID] = true
			}

			if s.Initial {
				initialCount++
			}
		}

		if initialCount > 1 {
			errs = append(errs, ValidationError{
				Field:   field + ".states",
				Message: fmt.Sprintf("entity %q has %d initial states, must have at most 1", e.ID, initialCount),
			})
		}
	}

	return errs
}
