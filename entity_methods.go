package pidl

// HasStates returns true if the entity has any states defined.
func (e Entity) HasStates() bool {
	return len(e.States) > 0
}

// StateByID returns the state with the given ID, or nil if not found.
func (e Entity) StateByID(id string) *EntityState {
	for i := range e.States {
		if e.States[i].ID == id {
			return &e.States[i]
		}
	}
	return nil
}

// InitialState returns the initial state, or nil if none is marked as initial.
func (e Entity) InitialState() *EntityState {
	for i := range e.States {
		if e.States[i].Initial {
			return &e.States[i]
		}
	}
	return nil
}

// StateIDs returns a slice of all state IDs for this entity.
func (e Entity) StateIDs() []string {
	ids := make([]string, len(e.States))
	for i, s := range e.States {
		ids[i] = s.ID
	}
	return ids
}

// HasProtocolRoles returns true if the entity has protocol roles defined.
func (e Entity) HasProtocolRoles() bool {
	return len(e.ProtocolRoles) > 0
}

// RolesByProtocol returns all roles for a specific protocol.
func (e Entity) RolesByProtocol(protocol string) []ProtocolRole {
	var roles []ProtocolRole
	for _, r := range e.ProtocolRoles {
		if r.Protocol == protocol {
			roles = append(roles, r)
		}
	}
	return roles
}

// HasRole checks if the entity has a specific protocol role.
func (e Entity) HasRole(protocol, role string) bool {
	for _, r := range e.ProtocolRoles {
		if r.Protocol == protocol && r.Role == role {
			return true
		}
	}
	return false
}

// Process Spec helper methods

// IsProcessStep returns true if this entity has process step semantics.
func (e Entity) IsProcessStep() bool {
	return e.StepType != ""
}

// IsDeterministic returns true if this is a deterministic step.
func (e Entity) IsDeterministic() bool {
	return e.StepType == StepTypeDeterministic
}

// IsLLMStep returns true if this is an LLM-powered step.
func (e Entity) IsLLMStep() bool {
	return e.StepType == StepTypeLLM
}

// IsHumanStep returns true if this is a human-in-the-loop step.
func (e Entity) IsHumanStep() bool {
	return e.StepType == StepTypeHuman
}

// IsExternalStep returns true if this is an external service step.
func (e Entity) IsExternalStep() bool {
	return e.StepType == StepTypeExternal
}

// IsToolStep returns true if this is a tool invocation step.
func (e Entity) IsToolStep() bool {
	return e.StepType == StepTypeTool
}

// HasInputs returns true if this entity has inputs defined.
func (e Entity) HasInputs() bool {
	return len(e.Inputs) > 0
}

// HasOutputs returns true if this entity has outputs defined.
func (e Entity) HasOutputs() bool {
	return len(e.Outputs) > 0
}

// RequiredInputs returns inputs that are marked as required.
func (e Entity) RequiredInputs() []DataPort {
	var required []DataPort
	for _, p := range e.Inputs {
		if p.Required {
			required = append(required, p)
		}
	}
	return required
}

// SensitiveInputs returns inputs that are marked as sensitive.
func (e Entity) SensitiveInputs() []DataPort {
	var sensitive []DataPort
	for _, p := range e.Inputs {
		if p.Sensitive {
			sensitive = append(sensitive, p)
		}
	}
	return sensitive
}

// SensitiveOutputs returns outputs that are marked as sensitive.
func (e Entity) SensitiveOutputs() []DataPort {
	var sensitive []DataPort
	for _, p := range e.Outputs {
		if p.Sensitive {
			sensitive = append(sensitive, p)
		}
	}
	return sensitive
}

// InputByName returns the input with the given name, or nil if not found.
func (e Entity) InputByName(name string) *DataPort {
	for i := range e.Inputs {
		if e.Inputs[i].Name == name {
			return &e.Inputs[i]
		}
	}
	return nil
}

// OutputByName returns the output with the given name, or nil if not found.
func (e Entity) OutputByName(name string) *DataPort {
	for i := range e.Outputs {
		if e.Outputs[i].Name == name {
			return &e.Outputs[i]
		}
	}
	return nil
}

// FailureModeByID returns the failure mode with the given ID, or nil if not found.
func (e Entity) FailureModeByID(id string) *FailureMode {
	for i := range e.FailureModes {
		if e.FailureModes[i].ID == id {
			return &e.FailureModes[i]
		}
	}
	return nil
}

// IsNonDeterministic returns true if this step may produce different outputs for the same inputs.
func (e Entity) IsNonDeterministic() bool {
	if e.Processing != nil && e.Processing.Determinism == DeterminismNonDeterministic {
		return true
	}
	// LLM and Human steps are inherently non-deterministic
	return e.StepType == StepTypeLLM || e.StepType == StepTypeHuman
}
