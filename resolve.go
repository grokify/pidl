package pidl

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
)

// ResolveOptions configures protocol resolution behavior.
type ResolveOptions struct {
	// BasePath is the directory to resolve relative paths from.
	BasePath string

	// FS, when set, is the file system extends and imports paths are read
	// from, such as an embedded set of documents. Paths are then
	// slash-separated and relative to the file system's root, and may not
	// leave it.
	FS fs.FS

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

// locate returns the path to read for ref, relative to BasePath, and the
// key that identifies it in the resolution chain.
func (o ResolveOptions) locate(ref string) (p, key string, err error) {
	if o.FS != nil {
		p = path.Join(o.BasePath, filepath.ToSlash(ref))
		if !fs.ValidPath(p) {
			return "", "", fmt.Errorf("path %q leaves the file system", ref)
		}
		return p, "fs:" + p, nil
	}
	p = ref
	if !filepath.IsAbs(p) && o.BasePath != "" {
		p = filepath.Join(o.BasePath, p)
	}
	key, err = filepath.Abs(p)
	if err != nil {
		return "", "", fmt.Errorf("getting absolute path: %w", err)
	}
	return p, key, nil
}

// parse reads and parses the document at p.
func (o ResolveOptions) parse(p string) (*Protocol, error) {
	if o.FS == nil {
		return ParseFile(p)
	}
	data, err := fs.ReadFile(o.FS, p)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return Parse(data)
}

// dir returns the directory of p, the base of the paths p references.
func (o ResolveOptions) dir(p string) string {
	if o.FS != nil {
		return path.Dir(p)
	}
	return filepath.Dir(p)
}

// resolveExtends loads and resolves the base protocol.
func (p *Protocol) resolveExtends(opts ResolveOptions, chain []string, depth int) (*Protocol, error) {
	path, absPath, err := opts.locate(p.Extends.Path)
	if err != nil {
		return nil, err
	}

	for _, c := range chain {
		if c == absPath {
			return nil, fmt.Errorf("circular extends detected: %s", formatChain(append(chain, absPath)))
		}
	}

	base, err := opts.parse(path)
	if err != nil {
		return nil, fmt.Errorf("parsing base protocol %q: %w", path, err)
	}

	// Recursively resolve the base
	baseOpts := opts
	baseOpts.BasePath = opts.dir(path)
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
	if len(p.Extends.ExcludeScenarios) > 0 {
		resolved = excludeScenarios(resolved, p.Extends.ExcludeScenarios)
	}

	return resolved, nil
}

// resolveImport loads and resolves an imported protocol.
func (p *Protocol) resolveImport(imp ProtocolImport, opts ResolveOptions, chain []string, depth int) (*Protocol, error) {
	path, absPath, err := opts.locate(imp.Path)
	if err != nil {
		return nil, err
	}

	for _, c := range chain {
		if c == absPath {
			return nil, fmt.Errorf("circular import detected: %s", formatChain(append(chain, absPath)))
		}
	}

	imported, err := opts.parse(path)
	if err != nil {
		return nil, fmt.Errorf("parsing imported protocol %q: %w", path, err)
	}

	// Recursively resolve the import
	importOpts := opts
	importOpts.BasePath = opts.dir(path)
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
	if len(p.Groups) > 0 {
		result.Groups = make([]Group, len(p.Groups))
		copy(result.Groups, p.Groups)
	}
	result.Pronunciations = mergePronunciations(nil, p.Pronunciations)
	if len(p.Scenarios) > 0 {
		result.Scenarios = make([]Scenario, len(p.Scenarios))
		copy(result.Scenarios, p.Scenarios)
	}
	if len(p.Fragments) > 0 {
		result.Fragments = make([]Scenario, len(p.Fragments))
		copy(result.Fragments, p.Fragments)
	}

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
	result.Flows = append(append([]Flow{}, base.Flows...), derived.Flows...)

	// Merge scenarios: base order, a derived scenario with the same ID
	// replacing the base one in place, then new derived scenarios
	result.Scenarios = mergeScenarios(base.Scenarios, derived.Scenarios)

	// Merge groups the same way
	result.Groups = mergeGroups(base.Groups, derived.Groups)

	// Merge pronunciations: derived spoken forms win per term and language
	result.Pronunciations = mergePronunciations(base.Pronunciations, derived.Pronunciations)
	if result.ProtocolMeta.Language == "" {
		result.ProtocolMeta.Language = base.ProtocolMeta.Language
	}

	// Merge metadata
	result.Metadata = mergeMetadata(base.Metadata, derived.Metadata)

	return normalizeScenarios(result)
}

// mergeScenarios keeps base order, replaces base scenarios that derived
// redefines in their base position, and appends derived scenarios that
// are new.
func mergeScenarios(base, derived []Scenario) []Scenario {
	if len(base) == 0 && len(derived) == 0 {
		return nil
	}
	byID := make(map[string]Scenario, len(derived))
	for _, sc := range derived {
		byID[sc.ID] = sc
	}
	out := make([]Scenario, 0, len(base)+len(derived))
	used := make(map[string]bool, len(derived))
	for _, sc := range base {
		if d, ok := byID[sc.ID]; ok {
			out = append(out, d)
			used[sc.ID] = true
			continue
		}
		out = append(out, sc)
	}
	for _, sc := range derived {
		if !used[sc.ID] {
			out = append(out, sc)
		}
	}
	return out
}

// mergeGroups keeps base order, replaces base groups that derived
// redefines in their base position, and appends derived groups that are
// new.
func mergeGroups(base, derived []Group) []Group {
	if len(base) == 0 && len(derived) == 0 {
		return nil
	}
	byID := make(map[string]Group, len(derived))
	for _, g := range derived {
		byID[g.ID] = g
	}
	out := make([]Group, 0, len(base)+len(derived))
	used := make(map[string]bool, len(derived))
	for _, g := range base {
		if d, ok := byID[g.ID]; ok {
			out = append(out, d)
			used[g.ID] = true
			continue
		}
		out = append(out, g)
	}
	for _, g := range derived {
		if !used[g.ID] {
			out = append(out, g)
		}
	}
	return out
}

// mergePronunciations returns a new map with base's entries overlaid by
// over's, per term and language. It returns nil when both are empty.
func mergePronunciations(base, over map[string]map[string]string) map[string]map[string]string {
	if len(base) == 0 && len(over) == 0 {
		return nil
	}
	out := make(map[string]map[string]string, len(base)+len(over))
	for _, m := range []map[string]map[string]string{base, over} {
		for term, byLang := range m {
			if out[term] == nil {
				out[term] = make(map[string]string, len(byLang))
			}
			for lang, spoken := range byLang {
				out[term][lang] = spoken
			}
		}
	}
	return out
}

// normalizeScenarios turns top-level flows into the first scenario when
// composition leaves a protocol with both flows and scenarios. The new
// scenario takes the protocol's ID and name (with a "-flows" suffix when
// that ID is taken).
func normalizeScenarios(p *Protocol) *Protocol {
	if len(p.Flows) == 0 || len(p.Scenarios) == 0 {
		return p
	}
	id := p.ProtocolMeta.ID
	if id == "" {
		id = "flows"
	}
	if p.ScenarioByID(id) != nil {
		id += "-flows"
	}
	name := p.ProtocolMeta.Name
	if name == "" {
		name = id
	}
	p.Scenarios = append([]Scenario{{ID: id, Name: name, Flows: p.Flows}}, p.Scenarios...)
	p.Flows = nil
	return p
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

	// Import selected scenarios, with the fragments they include
	if len(imp.Scenarios) > 0 {
		scenarios := importScenarios(imported, imp)
		result.Scenarios = append(result.Scenarios, scenarios...)
		result.Groups = appendImportedGroups(result.Groups, imported, imp, scenarios)
		// The importing document's spoken forms win
		result.Pronunciations = mergePronunciations(imported.Pronunciations, result.Pronunciations)
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

	return normalizeScenarios(result)
}

// importScenarios returns the scenarios imp selects from imported, plus
// every fragment they include (transitively), in imported document order,
// with the alias prefix applied.
func importScenarios(imported *Protocol, imp ProtocolImport) []Scenario {
	want := make(map[string]bool)
	var mark func(id string)
	mark = func(id string) {
		if want[id] {
			return
		}
		sc := imported.ScenarioByID(id)
		if sc == nil {
			return
		}
		want[id] = true
		for _, f := range sc.Flows {
			if f.IsInclude() {
				mark(f.Include)
			}
		}
	}
	for _, id := range imp.Scenarios {
		mark(id)
	}

	prefix := func(id string) string {
		if id == "" || imp.Alias == "" {
			return id
		}
		return imp.Alias + id
	}
	var out []Scenario
	for _, sc := range imported.Scenarios {
		if !want[sc.ID] {
			continue
		}
		ns := sc
		ns.ID = prefix(sc.ID)
		ns.Group = prefix(sc.Group)
		ns.VariantOf = prefix(sc.VariantOf)
		ns.Participants = make([]string, len(sc.Participants))
		for i, id := range sc.Participants {
			ns.Participants[i] = prefix(id)
		}
		if len(sc.Participants) == 0 {
			ns.Participants = nil
		}
		ns.Flows = make([]Flow, len(sc.Flows))
		for i, f := range sc.Flows {
			nf := f
			nf.From, nf.To, nf.Include, nf.Phase = prefix(f.From), prefix(f.To), prefix(f.Include), prefix(f.Phase)
			ns.Flows[i] = nf
		}
		out = append(out, ns)
	}
	return out
}

// appendImportedGroups appends the imported groups that the imported
// scenarios use, with the alias prefix applied, in the imported file's
// group order. Groups already present are kept as they are.
func appendImportedGroups(groups []Group, imported *Protocol, imp ProtocolImport, scenarios []Scenario) []Group {
	used := map[string]bool{}
	for _, sc := range scenarios {
		if sc.Group != "" {
			used[sc.Group] = true
		}
	}
	have := make(map[string]bool, len(groups))
	for _, g := range groups {
		have[g.ID] = true
	}
	for _, g := range imported.Groups {
		ng := g
		ng.ID = imp.Alias + g.ID
		if used[ng.ID] && !have[ng.ID] {
			groups = append(groups, ng)
			have[ng.ID] = true
		}
	}
	return groups
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
	keep := func(flows []Flow) []Flow {
		filtered := make([]Flow, 0, len(flows))
		for _, f := range flows {
			if !excludeSet[f.From] && !excludeSet[f.To] {
				filtered = append(filtered, f)
			}
		}
		return filtered
	}
	result.Flows = keep(result.Flows)
	for i, sc := range result.Scenarios {
		result.Scenarios[i].Flows = keep(sc.Flows)
		if len(sc.Participants) > 0 {
			var parts []string
			for _, id := range sc.Participants {
				if !excludeSet[id] {
					parts = append(parts, id)
				}
			}
			result.Scenarios[i].Participants = parts
		}
	}

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
	for i, sc := range result.Scenarios {
		flows := make([]Flow, len(sc.Flows))
		copy(flows, sc.Flows)
		for j, f := range flows {
			if excludeSet[f.Phase] {
				flows[j].Phase = ""
			}
		}
		result.Scenarios[i].Flows = flows
	}

	return result
}

// excludeScenarios removes the scenarios with the given IDs.
func excludeScenarios(p *Protocol, ids []string) *Protocol {
	result := p.clone()
	excludeSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		excludeSet[id] = true
	}
	var kept []Scenario
	for _, sc := range result.Scenarios {
		if !excludeSet[sc.ID] {
			kept = append(kept, sc)
		}
	}
	result.Scenarios = kept
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
