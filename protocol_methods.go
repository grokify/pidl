package pidl

// EntityByID returns the entity with the given ID, or nil if not found.
func (p *Protocol) EntityByID(id string) *Entity {
	for i := range p.Entities {
		if p.Entities[i].ID == id {
			return &p.Entities[i]
		}
	}
	return nil
}

// PhaseByID returns the phase with the given ID, or nil if not found.
func (p *Protocol) PhaseByID(id string) *Phase {
	for i := range p.Phases {
		if p.Phases[i].ID == id {
			return &p.Phases[i]
		}
	}
	return nil
}

// FlowsByPhase returns all flows belonging to the given phase.
func (p *Protocol) FlowsByPhase(phaseID string) []Flow {
	var flows []Flow
	for _, f := range p.Flows {
		if f.Phase == phaseID {
			flows = append(flows, f)
		}
	}
	return flows
}

// EntityIDs returns a slice of all entity IDs.
func (p *Protocol) EntityIDs() []string {
	ids := make([]string, len(p.Entities))
	for i, e := range p.Entities {
		ids[i] = e.ID
	}
	return ids
}

// PhaseIDs returns a slice of all phase IDs.
func (p *Protocol) PhaseIDs() []string {
	ids := make([]string, len(p.Phases))
	for i, ph := range p.Phases {
		ids[i] = ph.ID
	}
	return ids
}

// RootPhases returns phases that have no parent (top-level phases).
func (p *Protocol) RootPhases() []Phase {
	var roots []Phase
	for _, ph := range p.Phases {
		if ph.Parent == "" {
			roots = append(roots, ph)
		}
	}
	return roots
}

// ChildPhases returns phases that have the given parent ID.
func (p *Protocol) ChildPhases(parentID string) []Phase {
	var children []Phase
	for _, ph := range p.Phases {
		if ph.Parent == parentID {
			children = append(children, ph)
		}
	}
	return children
}

// PhaseDepth returns the nesting depth of a phase (0 for root phases).
func (p *Protocol) PhaseDepth(phaseID string) int {
	depth := 0
	current := p.PhaseByID(phaseID)
	for current != nil && current.Parent != "" {
		depth++
		current = p.PhaseByID(current.Parent)
	}
	return depth
}

// EntitiesWithStates returns all entities that have states defined.
func (p *Protocol) EntitiesWithStates() []Entity {
	var entities []Entity
	for _, e := range p.Entities {
		if e.HasStates() {
			entities = append(entities, e)
		}
	}
	return entities
}

// StateTransitions extracts all state transitions from the protocol's flows.
func (p *Protocol) StateTransitions() []StateTransition {
	var transitions []StateTransition
	for _, f := range p.Flows {
		for _, m := range f.Sets {
			transitions = append(transitions, StateTransition{
				EntityID:   m.Entity,
				FromState:  m.From,
				ToState:    m.To,
				FlowAction: f.Action,
				FlowLabel:  f.DisplayLabel(),
			})
		}
	}
	return transitions
}

// StateTransitionsForEntity returns state transitions for a specific entity.
func (p *Protocol) StateTransitionsForEntity(entityID string) []StateTransition {
	var transitions []StateTransition
	for _, t := range p.StateTransitions() {
		if t.EntityID == entityID {
			transitions = append(transitions, t)
		}
	}
	return transitions
}

// TokenByID returns the token definition with the given ID, or nil if not found.
func (p *Protocol) TokenByID(id string) *TokenDefinition {
	if p.Metadata == nil {
		return nil
	}
	for i := range p.Metadata.Tokens {
		if p.Metadata.Tokens[i].ID == id {
			return &p.Metadata.Tokens[i]
		}
	}
	return nil
}

// EntitiesWithProtocolRoles returns all entities that have protocol roles defined.
func (p *Protocol) EntitiesWithProtocolRoles() []Entity {
	var entities []Entity
	for _, e := range p.Entities {
		if e.HasProtocolRoles() {
			entities = append(entities, e)
		}
	}
	return entities
}

// EntitiesByProtocol returns all entities that implement a role for a specific protocol.
func (p *Protocol) EntitiesByProtocol(protocol string) []Entity {
	var entities []Entity
	for _, e := range p.Entities {
		if len(e.RolesByProtocol(protocol)) > 0 {
			entities = append(entities, e)
		}
	}
	return entities
}

// EntitiesByRole returns all entities that implement a specific protocol role.
func (p *Protocol) EntitiesByRole(protocol, role string) []Entity {
	var entities []Entity
	for _, e := range p.Entities {
		if e.HasRole(protocol, role) {
			entities = append(entities, e)
		}
	}
	return entities
}

// ComponentByID returns the component with the given ID, or nil if not found.
func (p *Protocol) ComponentByID(id string) *DeploymentComponent {
	if p.Metadata == nil {
		return nil
	}
	for i := range p.Metadata.Components {
		if p.Metadata.Components[i].ID == id {
			return &p.Metadata.Components[i]
		}
	}
	return nil
}

// ComponentsByType returns all components of a given type.
// Returns an empty slice if no components match or Metadata is nil.
func (p *Protocol) ComponentsByType(componentType string) []DeploymentComponent {
	if p.Metadata == nil {
		return []DeploymentComponent{}
	}
	var components []DeploymentComponent
	for _, c := range p.Metadata.Components {
		if c.Type == componentType {
			components = append(components, c)
		}
	}
	if components == nil {
		return []DeploymentComponent{}
	}
	return components
}

// EntitiesInComponent returns all entities that belong to a component.
// Returns an empty slice if the component is not found.
func (p *Protocol) EntitiesInComponent(componentID string) []Entity {
	c := p.ComponentByID(componentID)
	if c == nil {
		return []Entity{}
	}
	var entities []Entity
	for _, eid := range c.Entities {
		if e := p.EntityByID(eid); e != nil {
			entities = append(entities, *e)
		}
	}
	if entities == nil {
		return []Entity{}
	}
	return entities
}

// ComponentForEntity returns the component that contains an entity, or nil if not found.
func (p *Protocol) ComponentForEntity(entityID string) *DeploymentComponent {
	if p.Metadata == nil {
		return nil
	}
	for i := range p.Metadata.Components {
		for _, eid := range p.Metadata.Components[i].Entities {
			if eid == entityID {
				return &p.Metadata.Components[i]
			}
		}
	}
	return nil
}

// TrustRelationByID returns the trust relation with the given ID, or nil if not found.
func (p *Protocol) TrustRelationByID(id string) *TrustRelationship {
	if p.Metadata == nil {
		return nil
	}
	for i := range p.Metadata.TrustRelations {
		if p.Metadata.TrustRelations[i].ID == id {
			return &p.Metadata.TrustRelations[i]
		}
	}
	return nil
}

// TrustRelationsFrom returns all trust relations originating from an entity or component.
// Returns an empty slice if no relations match or Metadata is nil.
func (p *Protocol) TrustRelationsFrom(id string) []TrustRelationship {
	if p.Metadata == nil {
		return []TrustRelationship{}
	}
	var relations []TrustRelationship
	for _, r := range p.Metadata.TrustRelations {
		if r.From == id {
			relations = append(relations, r)
		}
	}
	if relations == nil {
		return []TrustRelationship{}
	}
	return relations
}

// TrustRelationsTo returns all trust relations targeting an entity or component.
// Returns an empty slice if no relations match or Metadata is nil.
func (p *Protocol) TrustRelationsTo(id string) []TrustRelationship {
	if p.Metadata == nil {
		return []TrustRelationship{}
	}
	var relations []TrustRelationship
	for _, r := range p.Metadata.TrustRelations {
		if r.To == id {
			relations = append(relations, r)
		}
	}
	if relations == nil {
		return []TrustRelationship{}
	}
	return relations
}

// TrustRelationsByType returns all trust relations of a given type.
// Returns an empty slice if no relations match or Metadata is nil.
func (p *Protocol) TrustRelationsByType(relType string) []TrustRelationship {
	if p.Metadata == nil {
		return []TrustRelationship{}
	}
	var relations []TrustRelationship
	for _, r := range p.Metadata.TrustRelations {
		if r.Type == relType {
			relations = append(relations, r)
		}
	}
	if relations == nil {
		return []TrustRelationship{}
	}
	return relations
}

// AllProtocols returns a unique list of all protocols referenced in entity roles.
// Returns an empty slice if no protocols are found.
func (p *Protocol) AllProtocols() []string {
	seen := make(map[string]bool)
	var protocols []string
	for _, e := range p.Entities {
		for _, r := range e.ProtocolRoles {
			if !seen[r.Protocol] {
				seen[r.Protocol] = true
				protocols = append(protocols, r.Protocol)
			}
		}
	}
	if protocols == nil {
		return []string{}
	}
	return protocols
}

// AllComponentTypes returns a unique list of all component types.
// Returns an empty slice if no components exist or Metadata is nil.
func (p *Protocol) AllComponentTypes() []string {
	if p.Metadata == nil {
		return []string{}
	}
	seen := make(map[string]bool)
	var types []string
	for _, c := range p.Metadata.Components {
		if !seen[c.Type] {
			seen[c.Type] = true
			types = append(types, c.Type)
		}
	}
	if types == nil {
		return []string{}
	}
	return types
}

// AllTrustRelationTypes returns a unique list of all trust relation types.
// Returns an empty slice if no trust relations exist or Metadata is nil.
func (p *Protocol) AllTrustRelationTypes() []string {
	if p.Metadata == nil {
		return []string{}
	}
	seen := make(map[string]bool)
	var types []string
	for _, r := range p.Metadata.TrustRelations {
		if !seen[r.Type] {
			seen[r.Type] = true
			types = append(types, r.Type)
		}
	}
	if types == nil {
		return []string{}
	}
	return types
}
