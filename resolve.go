package pidl

import (
	"fmt"
	"path/filepath"
)

// ResolveOptions configures protocol resolution behavior.
type ResolveOptions struct {
	// BasePath is the directory to resolve relative paths from.
	BasePath string

	// MaxDepth limits import nesting depth (default: 10).
	MaxDepth int
}

// DefaultResolveOptions returns default resolution options.
func DefaultResolveOptions() ResolveOptions {
	return ResolveOptions{
		MaxDepth: 10,
	}
}

// Resolve resolves all imports and extends in the protocol, returning a new
// fully-resolved Protocol. The original is not modified.
func (p *Protocol) Resolve(opts ResolveOptions) (*Protocol, error) {
	return p.resolveWithChain(opts, nil, 0)
}

// IsResolved returns true if the protocol has been resolved.
func (p *Protocol) IsResolved() bool {
	return p.resolved
}

// NeedsResolution returns true if the protocol has imports or extends that need resolution.
func (p *Protocol) NeedsResolution() bool {
	return p.Extends != nil || len(p.Imports) > 0
}

// resolveWithChain performs resolution while tracking the import chain for cycle detection.
func (p *Protocol) resolveWithChain(opts ResolveOptions, chain []string, depth int) (*Protocol, error) {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = 10
	}

	if depth > opts.MaxDepth {
		return nil, fmt.Errorf("import depth exceeds maximum of %d", opts.MaxDepth)
	}

	// Start with a copy of the current protocol
	result := p.clone()

	// Handle extends first (base protocol)
	if p.Extends != nil {
		base, err := p.resolveExtends(opts, chain, depth)
		if err != nil {
			return nil, fmt.Errorf("resolving extends: %w", err)
		}
		result = mergeBase(base, result)
	}

	// Handle imports
	for i, imp := range p.Imports {
		imported, err := p.resolveImport(imp, opts, chain, depth)
		if err != nil {
			return nil, fmt.Errorf("resolving import[%d] %q: %w", i, imp.Path, err)
		}
		result = mergeImport(result, imported, imp)
	}

	result.resolved = true
	result.Extends = nil
	result.Imports = nil

	return result, nil
}

// resolveExtends loads and resolves the base protocol.
func (p *Protocol) resolveExtends(opts ResolveOptions, chain []string, depth int) (*Protocol, error) {
	path := p.Extends.Path
	if !filepath.IsAbs(path) && opts.BasePath != "" {
		path = filepath.Join(opts.BasePath, path)
	}

	// Check for circular reference
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("getting absolute path: %w", err)
	}

	for _, c := range chain {
		if c == absPath {
			return nil, fmt.Errorf("circular extends detected: %s", formatChain(append(chain, absPath)))
		}
	}

	base, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parsing base protocol %q: %w", path, err)
	}

	// Recursively resolve the base
	baseOpts := opts
	baseOpts.BasePath = filepath.Dir(path)
	resolved, err := base.resolveWithChain(baseOpts, append(chain, absPath), depth+1)
	if err != nil {
		return nil, err
	}

	// Apply exclusions
	if len(p.Extends.ExcludeEntities) > 0 {
		resolved = excludeEntities(resolved, p.Extends.ExcludeEntities)
	}
	if len(p.Extends.ExcludePhases) > 0 {
		resolved = excludePhases(resolved, p.Extends.ExcludePhases)
	}
	if len(p.Extends.ExcludeFlows) > 0 {
		resolved = excludeFlows(resolved, p.Extends.ExcludeFlows)
	}

	return resolved, nil
}

// resolveImport loads and resolves an imported protocol.
func (p *Protocol) resolveImport(imp ProtocolImport, opts ResolveOptions, chain []string, depth int) (*Protocol, error) {
	path := imp.Path
	if !filepath.IsAbs(path) && opts.BasePath != "" {
		path = filepath.Join(opts.BasePath, path)
	}

	// Check for circular reference
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("getting absolute path: %w", err)
	}

	for _, c := range chain {
		if c == absPath {
			return nil, fmt.Errorf("circular import detected: %s", formatChain(append(chain, absPath)))
		}
	}

	imported, err := ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parsing imported protocol %q: %w", path, err)
	}

	// Recursively resolve the import
	importOpts := opts
	importOpts.BasePath = filepath.Dir(path)
	resolved, err := imported.resolveWithChain(importOpts, append(chain, absPath), depth+1)
	if err != nil {
		return nil, err
	}

	return resolved, nil
}

// clone creates a shallow copy of the protocol.
func (p *Protocol) clone() *Protocol {
	result := &Protocol{
		ProtocolMeta: p.ProtocolMeta,
		Extends:      p.Extends,
		Imports:      p.Imports,
		Entities:     make([]Entity, len(p.Entities)),
		Phases:       make([]Phase, len(p.Phases)),
		Flows:        make([]Flow, len(p.Flows)),
		resolved:     p.resolved,
	}
	copy(result.Entities, p.Entities)
	copy(result.Phases, p.Phases)
	copy(result.Flows, p.Flows)

	if p.Metadata != nil {
		result.Metadata = &ProtocolMetadata{
			Networks:      p.Metadata.Networks,
			NetworkLayout: p.Metadata.NetworkLayout,
		}
		if len(p.Metadata.Tokens) > 0 {
			result.Metadata.Tokens = make([]TokenDefinition, len(p.Metadata.Tokens))
			copy(result.Metadata.Tokens, p.Metadata.Tokens)
		}
		if len(p.Metadata.Components) > 0 {
			result.Metadata.Components = make([]DeploymentComponent, len(p.Metadata.Components))
			copy(result.Metadata.Components, p.Metadata.Components)
		}
		if len(p.Metadata.TrustRelations) > 0 {
			result.Metadata.TrustRelations = make([]TrustRelationship, len(p.Metadata.TrustRelations))
			copy(result.Metadata.TrustRelations, p.Metadata.TrustRelations)
		}
	}

	return result
}

// mergeBase merges a base protocol into a derived protocol.
// Derived protocol values take precedence.
func mergeBase(base, derived *Protocol) *Protocol {
	result := derived.clone()

	// Merge entities: base first, then derived (derived can override by ID)
	derivedEntityIDs := make(map[string]bool)
	for _, e := range derived.Entities {
		derivedEntityIDs[e.ID] = true
	}

	baseEntities := make([]Entity, 0, len(base.Entities))
	for _, e := range base.Entities {
		if !derivedEntityIDs[e.ID] {
			baseEntities = append(baseEntities, e)
		}
	}
	result.Entities = append(baseEntities, derived.Entities...)

	// Merge phases: base first, then derived (derived can override by ID)
	derivedPhaseIDs := make(map[string]bool)
	for _, ph := range derived.Phases {
		derivedPhaseIDs[ph.ID] = true
	}

	basePhases := make([]Phase, 0, len(base.Phases))
	for _, ph := range base.Phases {
		if !derivedPhaseIDs[ph.ID] {
			basePhases = append(basePhases, ph)
		}
	}
	result.Phases = append(basePhases, derived.Phases...)

	// Merge flows: base flows first, then derived flows
	result.Flows = append(base.Flows, derived.Flows...)

	// Merge metadata
	result.Metadata = mergeMetadata(base.Metadata, derived.Metadata)

	return result
}

// mergeImport merges an imported protocol into the target.
func mergeImport(target, imported *Protocol, imp ProtocolImport) *Protocol {
	result := target.clone()

	// Determine which entities to import
	entitiesToImport := make(map[string]bool)
	if len(imp.Entities) > 0 {
		for _, id := range imp.Entities {
			entitiesToImport[id] = true
		}
	} else {
		// Import all entities if not specified
		for _, e := range imported.Entities {
			entitiesToImport[e.ID] = true
		}
	}

	// Import entities with optional alias
	existingEntityIDs := make(map[string]bool)
	for _, e := range result.Entities {
		existingEntityIDs[e.ID] = true
	}

	for _, e := range imported.Entities {
		if !entitiesToImport[e.ID] {
			continue
		}
		newEntity := e
		if imp.Alias != "" {
			newEntity.ID = imp.Alias + e.ID
		}
		if !existingEntityIDs[newEntity.ID] {
			result.Entities = append(result.Entities, newEntity)
			existingEntityIDs[newEntity.ID] = true
		}
	}

	// Import phases if specified
	if len(imp.Phases) > 0 {
		existingPhaseIDs := make(map[string]bool)
		for _, ph := range result.Phases {
			existingPhaseIDs[ph.ID] = true
		}

		phasesToImport := make(map[string]bool)
		for _, id := range imp.Phases {
			phasesToImport[id] = true
		}

		for _, ph := range imported.Phases {
			if !phasesToImport[ph.ID] {
				continue
			}
			newPhase := ph
			if imp.Alias != "" {
				newPhase.ID = imp.Alias + ph.ID
				if newPhase.Parent != "" {
					newPhase.Parent = imp.Alias + newPhase.Parent
				}
			}
			if !existingPhaseIDs[newPhase.ID] {
				result.Phases = append(result.Phases, newPhase)
				existingPhaseIDs[newPhase.ID] = true
			}
		}
	}

	// Import flows if requested
	if imp.IncludeFlows {
		for _, f := range imported.Flows {
			// Only import flows between imported entities
			fromImported := entitiesToImport[f.From]
			toImported := entitiesToImport[f.To]
			if !fromImported || !toImported {
				continue
			}

			newFlow := f
			if imp.Alias != "" {
				newFlow.From = imp.Alias + f.From
				newFlow.To = imp.Alias + f.To
				if newFlow.Phase != "" {
					newFlow.Phase = imp.Alias + newFlow.Phase
				}
			}
			result.Flows = append(result.Flows, newFlow)
		}
	}

	return result
}

// mergeMetadata merges metadata from base and derived protocols.
func mergeMetadata(base, derived *ProtocolMetadata) *ProtocolMetadata {
	if base == nil && derived == nil {
		return nil
	}
	if base == nil {
		return derived
	}
	if derived == nil {
		return base
	}

	result := &ProtocolMetadata{
		NetworkLayout: derived.NetworkLayout,
	}
	if result.NetworkLayout == nil {
		result.NetworkLayout = base.NetworkLayout
	}

	// Merge networks
	if base.Networks != nil || derived.Networks != nil {
		result.Networks = make(map[string]*NetworkConfig)
		for k, v := range base.Networks {
			result.Networks[k] = v
		}
		for k, v := range derived.Networks {
			result.Networks[k] = v // Derived overrides base
		}
	}

	// Merge tokens (derived first, then base for non-duplicates)
	tokenIDs := make(map[string]bool)
	for _, t := range derived.Tokens {
		result.Tokens = append(result.Tokens, t)
		tokenIDs[t.ID] = true
	}
	for _, t := range base.Tokens {
		if !tokenIDs[t.ID] {
			result.Tokens = append(result.Tokens, t)
		}
	}

	// Merge components (derived first, then base for non-duplicates)
	compIDs := make(map[string]bool)
	for _, c := range derived.Components {
		result.Components = append(result.Components, c)
		compIDs[c.ID] = true
	}
	for _, c := range base.Components {
		if !compIDs[c.ID] {
			result.Components = append(result.Components, c)
		}
	}

	// Merge trust relations (all from both)
	result.TrustRelations = append(base.TrustRelations, derived.TrustRelations...)

	return result
}

// excludeEntities removes specified entities and flows referencing them.
func excludeEntities(p *Protocol, ids []string) *Protocol {
	result := p.clone()
	excludeSet := make(map[string]bool)
	for _, id := range ids {
		excludeSet[id] = true
	}

	// Filter entities
	filtered := make([]Entity, 0, len(result.Entities))
	for _, e := range result.Entities {
		if !excludeSet[e.ID] {
			filtered = append(filtered, e)
		}
	}
	result.Entities = filtered

	// Filter flows referencing excluded entities
	filteredFlows := make([]Flow, 0, len(result.Flows))
	for _, f := range result.Flows {
		if !excludeSet[f.From] && !excludeSet[f.To] {
			filteredFlows = append(filteredFlows, f)
		}
	}
	result.Flows = filteredFlows

	return result
}

// excludePhases removes specified phases and clears phase references in flows.
func excludePhases(p *Protocol, ids []string) *Protocol {
	result := p.clone()
	excludeSet := make(map[string]bool)
	for _, id := range ids {
		excludeSet[id] = true
	}

	// Filter phases
	filtered := make([]Phase, 0, len(result.Phases))
	for _, ph := range result.Phases {
		if !excludeSet[ph.ID] {
			filtered = append(filtered, ph)
		}
	}
	result.Phases = filtered

	// Clear phase references in flows
	for i, f := range result.Flows {
		if excludeSet[f.Phase] {
			result.Flows[i].Phase = ""
		}
	}

	return result
}

// excludeFlows removes flows at specified indices.
func excludeFlows(p *Protocol, indices []int) *Protocol {
	result := p.clone()
	excludeSet := make(map[int]bool)
	for _, idx := range indices {
		excludeSet[idx] = true
	}

	filtered := make([]Flow, 0, len(result.Flows))
	for i, f := range result.Flows {
		if !excludeSet[i] {
			filtered = append(filtered, f)
		}
	}
	result.Flows = filtered

	return result
}

// formatChain formats an import chain for error messages.
func formatChain(chain []string) string {
	if len(chain) == 0 {
		return ""
	}
	result := chain[0]
	for i := 1; i < len(chain); i++ {
		result += " -> " + filepath.Base(chain[i])
	}
	return result
}
