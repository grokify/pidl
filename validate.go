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
		if !IsValidCategory(p.ProtocolMeta.Category) {
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
