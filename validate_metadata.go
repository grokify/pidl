package pidl

import "fmt"

func (p *Protocol) validateTokenDefinitions() ValidationErrors {
	var errs ValidationErrors

	if p.Metadata == nil || len(p.Metadata.Tokens) == 0 {
		return errs
	}

	// Build entity lookup map for issuer/audience validation
	entityIDs := make(map[string]bool)
	for _, e := range p.Entities {
		entityIDs[e.ID] = true
	}

	seen := make(map[string]bool)
	for i, t := range p.Metadata.Tokens {
		field := fmt.Sprintf("metadata.tokens[%d]", i)

		if t.ID == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".id",
				Message: "required",
			})
		} else {
			if !idPattern.MatchString(t.ID) {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: "must match pattern ^[a-z][a-z0-9_]*$",
				})
			}
			if seen[t.ID] {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: fmt.Sprintf("duplicate token ID %q", t.ID),
				})
			}
			seen[t.ID] = true
		}

		// Validate issuer reference if provided
		if t.Issuer != "" && !entityIDs[t.Issuer] {
			errs = append(errs, ValidationError{
				Field:   field + ".issuer",
				Message: fmt.Sprintf("unknown entity %q", t.Issuer),
			})
		}

		// Validate audience reference if provided
		if t.Audience != "" && !entityIDs[t.Audience] {
			errs = append(errs, ValidationError{
				Field:   field + ".audience",
				Message: fmt.Sprintf("unknown entity %q", t.Audience),
			})
		}

		errs = append(errs, validateEntityRefs(field+".verified_by", t.VerifiedBy, entityIDs)...)

		claimNames := make(map[string]bool, len(t.Claims))
		for j, c := range t.Claims {
			cf := fmt.Sprintf("%s.claims[%d]", field, j)
			switch {
			case c.Name == "":
				errs = append(errs, ValidationError{Field: cf + ".name", Message: "required"})
			case claimNames[c.Name]:
				errs = append(errs, ValidationError{Field: cf + ".name", Message: fmt.Sprintf("duplicate claim %q", c.Name)})
			}
			claimNames[c.Name] = true
			errs = append(errs, validateEntityRefs(cf+".checked_by", c.CheckedBy, entityIDs)...)
		}
	}

	return errs
}

func (p *Protocol) validateFlowSecurity() ValidationErrors {
	var errs ValidationErrors

	// Build token lookup map
	tokenIDs := make(map[string]bool)
	if p.Metadata != nil {
		for _, t := range p.Metadata.Tokens {
			tokenIDs[t.ID] = true
		}
	}

	entityIDs := make(map[string]bool, len(p.Entities))
	for _, e := range p.Entities {
		entityIDs[e.ID] = true
	}

	for _, list := range p.flowLists() {
		errs = append(errs, validateFlowListSecurity(list, tokenIDs, entityIDs)...)
	}

	return errs
}

// validateEntityRefs checks that every ID in ids is a known entity, listed
// once.
func validateEntityRefs(field string, ids []string, entityIDs map[string]bool) ValidationErrors {
	var errs ValidationErrors
	seen := make(map[string]bool, len(ids))
	for i, id := range ids {
		f := fmt.Sprintf("%s[%d]", field, i)
		switch {
		case !entityIDs[id]:
			errs = append(errs, ValidationError{Field: f, Message: fmt.Sprintf("unknown entity %q", id)})
		case seen[id]:
			errs = append(errs, ValidationError{Field: f, Message: fmt.Sprintf("duplicate entity %q", id)})
		}
		seen[id] = true
	}
	return errs
}

// validateTokenRefs checks that every ID in ids is a known token definition.
func validateTokenRefs(field string, ids []string, tokenIDs map[string]bool) ValidationErrors {
	var errs ValidationErrors
	for i, id := range ids {
		if !tokenIDs[id] {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("%s[%d]", field, i),
				Message: fmt.Sprintf("unknown token definition %q", id),
			})
		}
	}
	return errs
}

func validateFlowListSecurity(list flowList, tokenIDs, entityIDs map[string]bool) ValidationErrors {
	var errs ValidationErrors

	for i, f := range list.flows {
		if f.Security == nil {
			continue
		}

		field := fmt.Sprintf("%s[%d].security", list.field, i)

		// Validate security requirements
		for j, req := range f.Security.Requires {
			if !IsValidSecurityRequirement(req) {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("%s.requires[%d]", field, j),
					Message: fmt.Sprintf("invalid security requirement %q", req),
				})
			}
		}

		// Validate token reference
		if f.Security.Token != "" && !tokenIDs[f.Security.Token] {
			errs = append(errs, ValidationError{
				Field:   field + ".token",
				Message: fmt.Sprintf("unknown token definition %q", f.Security.Token),
			})
		}

		if f.Security.SignedBy != "" && !entityIDs[f.Security.SignedBy] {
			errs = append(errs, ValidationError{
				Field:   field + ".signed_by",
				Message: fmt.Sprintf("unknown entity %q", f.Security.SignedBy),
			})
		}
		errs = append(errs, validateTokenRefs(field+".issues", f.Security.Issues, tokenIDs)...)
		errs = append(errs, validateTokenRefs(field+".verifies", f.Security.Verifies, tokenIDs)...)
	}

	return errs
}

func (p *Protocol) validateProtocolRoles() ValidationErrors {
	var errs ValidationErrors

	for i, e := range p.Entities {
		for j, r := range e.ProtocolRoles {
			field := fmt.Sprintf("entities[%d].protocol_roles[%d]", i, j)

			if r.Protocol == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".protocol",
					Message: "required",
				})
			} else if !IsValidProtocol(r.Protocol) {
				errs = append(errs, ValidationError{
					Field:   field + ".protocol",
					Message: fmt.Sprintf("unknown protocol %q", r.Protocol),
				})
			}

			if r.Role == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".role",
					Message: "required",
				})
			}
		}
	}

	return errs
}

func (p *Protocol) validateComponents() ValidationErrors {
	var errs ValidationErrors

	if p.Metadata == nil || len(p.Metadata.Components) == 0 {
		return errs
	}

	// Build entity lookup map
	entityIDs := make(map[string]bool)
	for _, e := range p.Entities {
		entityIDs[e.ID] = true
	}

	seen := make(map[string]bool)
	for i, c := range p.Metadata.Components {
		field := fmt.Sprintf("metadata.components[%d]", i)

		if c.ID == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".id",
				Message: "required",
			})
		} else {
			if !protocolIDPattern.MatchString(c.ID) {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: "must match pattern ^[a-z][a-z0-9_-]*$",
				})
			}
			if seen[c.ID] {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: fmt.Sprintf("duplicate component ID %q", c.ID),
				})
			}
			seen[c.ID] = true
		}

		if c.Name == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".name",
				Message: "required",
			})
		}

		if c.Type == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".type",
				Message: "required",
			})
		} else if !IsValidComponentType(c.Type) {
			errs = append(errs, ValidationError{
				Field:   field + ".type",
				Message: fmt.Sprintf("unknown component type %q", c.Type),
			})
		}

		// Validate entity references
		for j, eid := range c.Entities {
			if !entityIDs[eid] {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("%s.entities[%d]", field, j),
					Message: fmt.Sprintf("unknown entity %q", eid),
				})
			}
		}

		// Validate implements roles
		for j, r := range c.Implements {
			roleField := fmt.Sprintf("%s.implements[%d]", field, j)

			if r.Protocol == "" {
				errs = append(errs, ValidationError{
					Field:   roleField + ".protocol",
					Message: "required",
				})
			} else if !IsValidProtocol(r.Protocol) {
				errs = append(errs, ValidationError{
					Field:   roleField + ".protocol",
					Message: fmt.Sprintf("unknown protocol %q", r.Protocol),
				})
			}

			if r.Role == "" {
				errs = append(errs, ValidationError{
					Field:   roleField + ".role",
					Message: "required",
				})
			}
		}
	}

	return errs
}

func (p *Protocol) validateTrustRelations() ValidationErrors {
	var errs ValidationErrors

	if p.Metadata == nil || len(p.Metadata.TrustRelations) == 0 {
		return errs
	}

	// Build lookup maps for entities and components
	validIDs := make(map[string]bool)
	for _, e := range p.Entities {
		validIDs[e.ID] = true
	}
	if p.Metadata != nil {
		for _, c := range p.Metadata.Components {
			validIDs[c.ID] = true
		}
	}

	seen := make(map[string]bool)
	for i, tr := range p.Metadata.TrustRelations {
		field := fmt.Sprintf("metadata.trust_relations[%d]", i)

		// Validate optional ID uniqueness
		if tr.ID != "" {
			if seen[tr.ID] {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: fmt.Sprintf("duplicate trust relation ID %q", tr.ID),
				})
			}
			seen[tr.ID] = true
		}

		if tr.From == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".from",
				Message: "required",
			})
		} else if !validIDs[tr.From] {
			errs = append(errs, ValidationError{
				Field:   field + ".from",
				Message: fmt.Sprintf("unknown entity or component %q", tr.From),
			})
		}

		if tr.To == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".to",
				Message: "required",
			})
		} else if !validIDs[tr.To] {
			errs = append(errs, ValidationError{
				Field:   field + ".to",
				Message: fmt.Sprintf("unknown entity or component %q", tr.To),
			})
		}

		if tr.Type == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".type",
				Message: "required",
			})
		} else if !IsValidTrustRelationType(tr.Type) {
			errs = append(errs, ValidationError{
				Field:   field + ".type",
				Message: fmt.Sprintf("unknown trust relation type %q", tr.Type),
			})
		}

		// Validate credentials
		for j, cred := range tr.Credentials {
			if !IsValidCredential(cred) {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("%s.credentials[%d]", field, j),
					Message: fmt.Sprintf("unknown credential type %q", cred),
				})
			}
		}
	}

	return errs
}
