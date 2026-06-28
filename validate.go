package pidl

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error with context.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d validation errors:\n", len(e))
	for _, err := range e {
		sb.WriteString("  - ")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

var (
	// idPattern matches valid entity and phase IDs: lowercase, starts with letter, alphanumeric + underscore.
	idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

	// protocolIDPattern matches valid protocol IDs: lowercase, starts with letter, alphanumeric + underscore + hyphen.
	protocolIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
)

// Validate checks the Protocol for errors and returns all found issues.
func (p *Protocol) Validate() ValidationErrors {
	var errs ValidationErrors

	// Validate protocol metadata
	errs = append(errs, p.validateProtocolMeta()...)

	// Validate token definitions
	errs = append(errs, p.validateTokenDefinitions()...)

	// Validate entities
	errs = append(errs, p.validateEntities()...)

	// Validate entity states
	errs = append(errs, p.validateEntityStates()...)

	// Validate phases
	errs = append(errs, p.validatePhases()...)

	// Validate flows
	errs = append(errs, p.validateFlows()...)

	// Validate state mutations in flows
	errs = append(errs, p.validateStateMutations()...)

	// Validate flow security
	errs = append(errs, p.validateFlowSecurity()...)

	// Validate protocol roles on entities
	errs = append(errs, p.validateProtocolRoles()...)

	// Validate deployment components
	errs = append(errs, p.validateComponents()...)

	// Validate trust relationships
	errs = append(errs, p.validateTrustRelations()...)

	return errs
}

// IsValid returns true if the Protocol passes validation.
func (p *Protocol) IsValid() bool {
	return !p.Validate().HasErrors()
}

func (p *Protocol) validateProtocolMeta() ValidationErrors {
	var errs ValidationErrors

	if p.ProtocolMeta.ID == "" {
		errs = append(errs, ValidationError{
			Field:   "protocol.id",
			Message: "required",
		})
	} else if !protocolIDPattern.MatchString(p.ProtocolMeta.ID) {
		errs = append(errs, ValidationError{
			Field:   "protocol.id",
			Message: "must match pattern ^[a-z][a-z0-9_-]*$",
		})
	}

	if p.ProtocolMeta.Name == "" {
		errs = append(errs, ValidationError{
			Field:   "protocol.name",
			Message: "required",
		})
	}

	if p.ProtocolMeta.Category != "" {
		if !isValidCategory(p.ProtocolMeta.Category) {
			errs = append(errs, ValidationError{
				Field:   "protocol.category",
				Message: fmt.Sprintf("invalid category %q", p.ProtocolMeta.Category),
			})
		}
	}

	for i, ref := range p.ProtocolMeta.References {
		if ref.Name == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("protocol.references[%d].name", i),
				Message: "required",
			})
		}
		if ref.URL == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("protocol.references[%d].url", i),
				Message: "required",
			})
		}
	}

	return errs
}

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
		} else if !isValidEntityType(e.Type) {
			errs = append(errs, ValidationError{
				Field:   field + ".type",
				Message: fmt.Sprintf("invalid entity type %q", e.Type),
			})
		}

		if e.TrustLevel != "" && !isValidTrustLevel(e.TrustLevel) {
			errs = append(errs, ValidationError{
				Field:   field + ".trust_level",
				Message: fmt.Sprintf("invalid trust level %q", e.TrustLevel),
			})
		}
	}

	return errs
}

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

func (p *Protocol) validateFlows() ValidationErrors {
	var errs ValidationErrors

	if len(p.Flows) < 1 {
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

	for i, f := range p.Flows {
		field := fmt.Sprintf("flows[%d]", i)

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

		if f.Mode != "" && !isValidFlowMode(f.Mode) {
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
			} else if !isValidAnnotationType(ann.Type) {
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

func (p *Protocol) validateStateMutations() ValidationErrors {
	var errs ValidationErrors

	// Build entity lookup map
	entityMap := make(map[string]*Entity)
	for i := range p.Entities {
		entityMap[p.Entities[i].ID] = &p.Entities[i]
	}

	for i, f := range p.Flows {
		for j, m := range f.Sets {
			field := fmt.Sprintf("flows[%d].sets[%d]", i, j)

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

func isValidCategory(c Category) bool {
	switch c {
	case CategoryAuth, CategoryAgent, CategoryMessaging, CategoryProvisioning, CategoryOther:
		return true
	}
	return false
}

func isValidEntityType(t EntityType) bool {
	switch t {
	case EntityTypeClient, EntityTypeAuthorizationServer, EntityTypeResourceServer,
		EntityTypeUser, EntityTypeBrowser, EntityTypeAgent, EntityTypeToolServer,
		EntityTypeTool, EntityTypeDelegatedAgent, EntityTypeIdentityProvider,
		EntityTypeServiceProvider, EntityTypeServer, EntityTypeOther:
		return true
	}
	return false
}

func isValidFlowMode(m FlowMode) bool {
	switch m {
	case FlowModeRequest, FlowModeResponse, FlowModeRedirect, FlowModeCallback,
		FlowModeInteractive, FlowModeEvent, FlowModeToolCall, FlowModeToolResult:
		return true
	}
	return false
}

func isValidAnnotationType(t AnnotationType) bool {
	switch t {
	case AnnotationTypeSecurity, AnnotationTypePerformance, AnnotationTypeDeprecated,
		AnnotationTypeInfo, AnnotationTypeWarning, AnnotationTypeError:
		return true
	}
	return false
}

func isValidTrustLevel(t TrustLevel) bool {
	switch t {
	case TrustLevelTrusted, TrustLevelSemiTrusted, TrustLevelUntrusted, TrustLevelAuthoritative:
		return true
	}
	return false
}

func isValidSecurityRequirement(r SecurityRequirement) bool {
	switch r {
	case SecurityRequirementToken, SecurityRequirementSignature, SecurityRequirementEncryption,
		SecurityRequirementMTLS, SecurityRequirementMAC:
		return true
	}
	return false
}

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

	for i, f := range p.Flows {
		if f.Security == nil {
			continue
		}

		field := fmt.Sprintf("flows[%d].security", i)

		// Validate security requirements
		for j, req := range f.Security.Requires {
			if !isValidSecurityRequirement(req) {
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
