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
