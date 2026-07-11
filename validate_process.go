package pidl

import "fmt"

// ValidateProcess performs process-specific validation.
// These validations only apply when protocol.kind is "process".
func (p *Protocol) ValidateProcess() ValidationErrors {
	var errs ValidationErrors

	if !p.IsProcessSpec() {
		return errs // Skip for non-process specs
	}

	errs = append(errs, p.validateStepTypes()...)
	errs = append(errs, p.validateDataPorts()...)
	errs = append(errs, p.validateProcessingConfigs()...)
	errs = append(errs, p.validateFailureModes()...)
	errs = append(errs, p.validateRetryStrategies()...)

	return errs
}

// validStepTypes contains all valid step type values.
var validStepTypes = map[StepType]bool{
	StepTypeDeterministic: true,
	StepTypeLLM:           true,
	StepTypeHuman:         true,
	StepTypeExternal:      true,
	StepTypeTool:          true,
}

// validDataPortKinds contains all valid data port kind values.
var validDataPortKinds = map[DataPortKind]bool{
	DataPortKindFile:     true,
	DataPortKindObject:   true,
	DataPortKindAPI:      true,
	DataPortKindDatabase: true,
	DataPortKindQueue:    true,
	DataPortKindStream:   true,
}

// validDeterminism contains all valid determinism values.
var validDeterminism = map[Determinism]bool{
	DeterminismDeterministic:    true,
	DeterminismNonDeterministic: true,
}

// validateStepTypes ensures step types are valid (VP001).
func (p *Protocol) validateStepTypes() ValidationErrors {
	var errs ValidationErrors

	for _, e := range p.Entities {
		if e.StepType != "" && !validStepTypes[e.StepType] {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("entities.%s.step_type", e.ID),
				Message: fmt.Sprintf("invalid step type %q; must be one of: deterministic, llm, human, external, tool", e.StepType),
			})
		}
	}

	return errs
}

// validateDataPorts ensures data port definitions are valid (VP002).
func (p *Protocol) validateDataPorts() ValidationErrors {
	var errs ValidationErrors

	for _, e := range p.Entities {
		// Validate inputs
		for i, port := range e.Inputs {
			field := fmt.Sprintf("entities.%s.inputs[%d]", e.ID, i)

			if port.Kind == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".kind",
					Message: "required",
				})
			} else if !validDataPortKinds[port.Kind] {
				errs = append(errs, ValidationError{
					Field:   field + ".kind",
					Message: fmt.Sprintf("invalid data port kind %q; must be one of: file, object, api, database, queue, stream", port.Kind),
				})
			}

			if port.Name == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".name",
					Message: "required",
				})
			}
		}

		// Validate outputs
		for i, port := range e.Outputs {
			field := fmt.Sprintf("entities.%s.outputs[%d]", e.ID, i)

			if port.Kind == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".kind",
					Message: "required",
				})
			} else if !validDataPortKinds[port.Kind] {
				errs = append(errs, ValidationError{
					Field:   field + ".kind",
					Message: fmt.Sprintf("invalid data port kind %q; must be one of: file, object, api, database, queue, stream", port.Kind),
				})
			}

			if port.Name == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".name",
					Message: "required",
				})
			}
		}

		// Check for duplicate input names
		inputNames := make(map[string]bool)
		for i, port := range e.Inputs {
			if port.Name != "" {
				if inputNames[port.Name] {
					errs = append(errs, ValidationError{
						Field:   fmt.Sprintf("entities.%s.inputs[%d].name", e.ID, i),
						Message: fmt.Sprintf("duplicate input name %q", port.Name),
					})
				}
				inputNames[port.Name] = true
			}
		}

		// Check for duplicate output names
		outputNames := make(map[string]bool)
		for i, port := range e.Outputs {
			if port.Name != "" {
				if outputNames[port.Name] {
					errs = append(errs, ValidationError{
						Field:   fmt.Sprintf("entities.%s.outputs[%d].name", e.ID, i),
						Message: fmt.Sprintf("duplicate output name %q", port.Name),
					})
				}
				outputNames[port.Name] = true
			}
		}
	}

	return errs
}

// validateProcessingConfigs validates processing configurations (VP003).
func (p *Protocol) validateProcessingConfigs() ValidationErrors {
	var errs ValidationErrors

	for _, e := range p.Entities {
		if e.Processing == nil {
			continue
		}

		field := fmt.Sprintf("entities.%s.processing", e.ID)

		// Validate determinism value
		if e.Processing.Determinism != "" && !validDeterminism[e.Processing.Determinism] {
			errs = append(errs, ValidationError{
				Field:   field + ".determinism",
				Message: fmt.Sprintf("invalid determinism %q; must be one of: deterministic, non_deterministic", e.Processing.Determinism),
			})
		}

		// LLM steps should have model_policy
		if e.StepType == StepTypeLLM && e.Processing.ModelPolicy == "" {
			errs = append(errs, ValidationError{
				Field:   field + ".model_policy",
				Message: "recommended for llm step types",
			})
		}

		// Warn if determinism contradicts step type
		if e.StepType == StepTypeDeterministic && e.Processing.Determinism == DeterminismNonDeterministic {
			errs = append(errs, ValidationError{
				Field:   field + ".determinism",
				Message: "step_type is 'deterministic' but processing.determinism is 'non_deterministic'",
			})
		}
	}

	return errs
}

// validateRetryStrategies validates retry strategy configurations (VP004).
func (p *Protocol) validateRetryStrategies() ValidationErrors {
	var errs ValidationErrors

	for _, e := range p.Entities {
		if e.RetryStrategy == nil {
			continue
		}

		field := fmt.Sprintf("entities.%s.retry_strategy", e.ID)

		// Validate max_attempts
		if e.RetryStrategy.MaxAttempts < 0 {
			errs = append(errs, ValidationError{
				Field:   field + ".max_attempts",
				Message: "must be non-negative",
			})
		}

		// Validate backoff_multiplier
		if e.RetryStrategy.BackoffMultiplier < 0 {
			errs = append(errs, ValidationError{
				Field:   field + ".backoff_multiplier",
				Message: "must be non-negative",
			})
		}

		// Validate retry_on references existing failure modes
		for i, retryOn := range e.RetryStrategy.RetryOn {
			found := false
			for _, fm := range e.FailureModes {
				if fm.ID == retryOn {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("%s.retry_on[%d]", field, i),
					Message: fmt.Sprintf("references unknown failure mode %q", retryOn),
				})
			}
		}
	}

	return errs
}

// validateFailureModes validates failure mode definitions (VP005).
func (p *Protocol) validateFailureModes() ValidationErrors {
	var errs ValidationErrors

	for _, e := range p.Entities {
		// Check for duplicate failure mode IDs
		fmIDs := make(map[string]bool)
		for i, fm := range e.FailureModes {
			field := fmt.Sprintf("entities.%s.failure_modes[%d]", e.ID, i)

			if fm.ID == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: "required",
				})
			} else if fmIDs[fm.ID] {
				errs = append(errs, ValidationError{
					Field:   field + ".id",
					Message: fmt.Sprintf("duplicate failure mode id %q", fm.ID),
				})
			} else {
				fmIDs[fm.ID] = true
			}

			if fm.Name == "" {
				errs = append(errs, ValidationError{
					Field:   field + ".name",
					Message: "required",
				})
			}
		}
	}

	return errs
}
