package pidl

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// DiffType represents the type of difference.
type DiffType string

const (
	// DiffTypeAdded indicates an element was added.
	DiffTypeAdded DiffType = "added"
	// DiffTypeRemoved indicates an element was removed.
	DiffTypeRemoved DiffType = "removed"
	// DiffTypeModified indicates an element was modified.
	DiffTypeModified DiffType = "modified"
)

// DiffCategory represents the category of a diff item.
type DiffCategory string

const (
	DiffCategoryEntity   DiffCategory = "entity"
	DiffCategoryFlow     DiffCategory = "flow"
	DiffCategoryPhase    DiffCategory = "phase"
	DiffCategoryMetadata DiffCategory = "metadata"
)

// DiffItem represents a single difference between two protocols.
type DiffItem struct {
	// Type indicates whether the item was added, removed, or modified.
	Type DiffType `json:"type"`

	// Category classifies what was changed.
	Category DiffCategory `json:"category"`

	// Path identifies the location of the change (e.g., "entities[0]", "flows[2].action").
	Path string `json:"path"`

	// OldValue is the value before the change (nil for additions).
	OldValue interface{} `json:"old_value,omitempty"`

	// NewValue is the value after the change (nil for removals).
	NewValue interface{} `json:"new_value,omitempty"`

	// Summary is a human-readable description of the change.
	Summary string `json:"summary"`
}

// DiffSummary provides counts of changes by category.
type DiffSummary struct {
	// TotalChanges is the total number of differences.
	TotalChanges int `json:"total_changes"`

	// Added counts items that were added.
	Added int `json:"added"`

	// Removed counts items that were removed.
	Removed int `json:"removed"`

	// Modified counts items that were modified.
	Modified int `json:"modified"`

	// ByCategory breaks down changes by category.
	ByCategory map[DiffCategory]int `json:"by_category"`
}

// ProtocolDiff contains the complete comparison result.
type ProtocolDiff struct {
	// BaseProtocolID is the ID of the base protocol.
	BaseProtocolID string `json:"base_protocol_id"`

	// NewProtocolID is the ID of the new protocol.
	NewProtocolID string `json:"new_protocol_id"`

	// Items contains all differences found.
	Items []DiffItem `json:"items"`

	// Summary provides aggregate statistics.
	Summary DiffSummary `json:"summary"`
}

// DiffOptions controls comparison behavior.
type DiffOptions struct {
	// IgnoreMetadata skips metadata comparison.
	IgnoreMetadata bool

	// IgnoreDescriptions skips description field comparisons.
	IgnoreDescriptions bool

	// DeepFlowCompare enables detailed flow field comparison.
	DeepFlowCompare bool
}

// DefaultDiffOptions returns default comparison options.
func DefaultDiffOptions() DiffOptions {
	return DiffOptions{
		IgnoreMetadata:     false,
		IgnoreDescriptions: false,
		DeepFlowCompare:    true,
	}
}

// Compare compares two protocols and returns the differences.
func Compare(base, new *Protocol, opts DiffOptions) *ProtocolDiff {
	diff := &ProtocolDiff{
		BaseProtocolID: base.ProtocolMeta.ID,
		NewProtocolID:  new.ProtocolMeta.ID,
		Items:          make([]DiffItem, 0),
		Summary: DiffSummary{
			ByCategory: make(map[DiffCategory]int),
		},
	}

	// Compare protocol metadata
	if !opts.IgnoreMetadata {
		compareProtocolMeta(diff, &base.ProtocolMeta, &new.ProtocolMeta, opts)
	}

	// Compare entities
	compareEntities(diff, base.Entities, new.Entities, opts)

	// Compare phases
	comparePhases(diff, base.Phases, new.Phases, opts)

	// Compare flows
	compareFlows(diff, base.Flows, new.Flows, opts)

	// Compare metadata (networks, tokens, components, trust relations)
	if !opts.IgnoreMetadata {
		compareMetadata(diff, base.Metadata, new.Metadata, opts)
	}

	// Calculate summary
	diff.calculateSummary()

	return diff
}

// calculateSummary computes aggregate statistics from the items.
func (d *ProtocolDiff) calculateSummary() {
	d.Summary.TotalChanges = len(d.Items)
	d.Summary.Added = 0
	d.Summary.Removed = 0
	d.Summary.Modified = 0
	d.Summary.ByCategory = make(map[DiffCategory]int)

	for _, item := range d.Items {
		switch item.Type {
		case DiffTypeAdded:
			d.Summary.Added++
		case DiffTypeRemoved:
			d.Summary.Removed++
		case DiffTypeModified:
			d.Summary.Modified++
		}
		d.Summary.ByCategory[item.Category]++
	}
}

// HasChanges returns true if there are any differences.
func (d *ProtocolDiff) HasChanges() bool {
	return len(d.Items) > 0
}

// String returns a human-readable text representation of the diff.
func (d *ProtocolDiff) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Protocol Diff: %s → %s\n", d.BaseProtocolID, d.NewProtocolID))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	if !d.HasChanges() {
		sb.WriteString("No differences found.\n")
		return sb.String()
	}

	// Group items by category
	byCategory := make(map[DiffCategory][]DiffItem)
	for _, item := range d.Items {
		byCategory[item.Category] = append(byCategory[item.Category], item)
	}

	// Order categories
	categories := []DiffCategory{DiffCategoryEntity, DiffCategoryPhase, DiffCategoryFlow, DiffCategoryMetadata}

	for _, cat := range categories {
		items := byCategory[cat]
		if len(items) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("%s Changes (%d)\n", strings.Title(string(cat)), len(items)))
		sb.WriteString(strings.Repeat("-", 40) + "\n")

		for _, item := range items {
			symbol := symbolForDiffType(item.Type)
			sb.WriteString(fmt.Sprintf("  %s %s: %s\n", symbol, item.Path, item.Summary))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Summary: %d changes (+%d/-%d/~%d)\n",
		d.Summary.TotalChanges, d.Summary.Added, d.Summary.Removed, d.Summary.Modified))

	return sb.String()
}

// ToMarkdown returns a markdown representation of the diff.
func (d *ProtocolDiff) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Protocol Diff: %s → %s\n\n", d.BaseProtocolID, d.NewProtocolID))

	if !d.HasChanges() {
		sb.WriteString("No differences found.\n")
		return sb.String()
	}

	// Group items by category
	byCategory := make(map[DiffCategory][]DiffItem)
	for _, item := range d.Items {
		byCategory[item.Category] = append(byCategory[item.Category], item)
	}

	categories := []DiffCategory{DiffCategoryEntity, DiffCategoryPhase, DiffCategoryFlow, DiffCategoryMetadata}

	for _, cat := range categories {
		items := byCategory[cat]
		if len(items) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s Changes\n\n", strings.Title(string(cat))))

		for _, item := range items {
			symbol := symbolForDiffType(item.Type)
			sb.WriteString(fmt.Sprintf("- %s **%s**: %s\n", symbol, item.Path, item.Summary))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Count |\n|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Total Changes | %d |\n", d.Summary.TotalChanges))
	sb.WriteString(fmt.Sprintf("| Added | %d |\n", d.Summary.Added))
	sb.WriteString(fmt.Sprintf("| Removed | %d |\n", d.Summary.Removed))
	sb.WriteString(fmt.Sprintf("| Modified | %d |\n", d.Summary.Modified))

	return sb.String()
}

// ToJSON returns the diff as JSON bytes.
func (d *ProtocolDiff) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

func symbolForDiffType(t DiffType) string {
	switch t {
	case DiffTypeAdded:
		return "+"
	case DiffTypeRemoved:
		return "-"
	case DiffTypeModified:
		return "~"
	default:
		return "?"
	}
}

func compareProtocolMeta(diff *ProtocolDiff, base, new *ProtocolMeta, opts DiffOptions) {
	if base.Name != new.Name {
		diff.addItem(DiffTypeModified, DiffCategoryMetadata, "protocol.name",
			base.Name, new.Name, fmt.Sprintf("name changed from %q to %q", base.Name, new.Name))
	}

	if base.Version != new.Version {
		diff.addItem(DiffTypeModified, DiffCategoryMetadata, "protocol.version",
			base.Version, new.Version, fmt.Sprintf("version changed from %q to %q", base.Version, new.Version))
	}

	if !opts.IgnoreDescriptions && base.Description != new.Description {
		diff.addItem(DiffTypeModified, DiffCategoryMetadata, "protocol.description",
			base.Description, new.Description, "description changed")
	}

	if base.Category != new.Category {
		diff.addItem(DiffTypeModified, DiffCategoryMetadata, "protocol.category",
			base.Category, new.Category, fmt.Sprintf("category changed from %q to %q", base.Category, new.Category))
	}
}

func compareEntities(diff *ProtocolDiff, base, new []Entity, opts DiffOptions) {
	baseMap := make(map[string]*Entity)
	for i := range base {
		baseMap[base[i].ID] = &base[i]
	}

	newMap := make(map[string]*Entity)
	for i := range new {
		newMap[new[i].ID] = &new[i]
	}

	// Find removed entities
	for id, e := range baseMap {
		if _, exists := newMap[id]; !exists {
			diff.addItem(DiffTypeRemoved, DiffCategoryEntity, fmt.Sprintf("entities[%s]", id),
				e, nil, fmt.Sprintf("entity %q removed", e.Name))
		}
	}

	// Find added entities
	for id, e := range newMap {
		if _, exists := baseMap[id]; !exists {
			diff.addItem(DiffTypeAdded, DiffCategoryEntity, fmt.Sprintf("entities[%s]", id),
				nil, e, fmt.Sprintf("entity %q added", e.Name))
		}
	}

	// Compare existing entities
	for id, baseEntity := range baseMap {
		newEntity, exists := newMap[id]
		if !exists {
			continue
		}

		compareEntity(diff, baseEntity, newEntity, id, opts)
	}
}

func compareEntity(diff *ProtocolDiff, base, new *Entity, id string, opts DiffOptions) {
	if base.Name != new.Name {
		diff.addItem(DiffTypeModified, DiffCategoryEntity, fmt.Sprintf("entities[%s].name", id),
			base.Name, new.Name, fmt.Sprintf("name changed from %q to %q", base.Name, new.Name))
	}

	if base.Type != new.Type {
		diff.addItem(DiffTypeModified, DiffCategoryEntity, fmt.Sprintf("entities[%s].type", id),
			base.Type, new.Type, fmt.Sprintf("type changed from %q to %q", base.Type, new.Type))
	}

	if base.TrustLevel != new.TrustLevel {
		diff.addItem(DiffTypeModified, DiffCategoryEntity, fmt.Sprintf("entities[%s].trust_level", id),
			base.TrustLevel, new.TrustLevel, fmt.Sprintf("trust level changed from %q to %q", base.TrustLevel, new.TrustLevel))
	}

	if !opts.IgnoreDescriptions && base.Description != new.Description {
		diff.addItem(DiffTypeModified, DiffCategoryEntity, fmt.Sprintf("entities[%s].description", id),
			base.Description, new.Description, "description changed")
	}

	// Compare states
	compareEntityStates(diff, base.States, new.States, id)

	// Compare protocol roles
	if !reflect.DeepEqual(base.ProtocolRoles, new.ProtocolRoles) {
		diff.addItem(DiffTypeModified, DiffCategoryEntity, fmt.Sprintf("entities[%s].protocol_roles", id),
			base.ProtocolRoles, new.ProtocolRoles, "protocol roles changed")
	}
}

func compareEntityStates(diff *ProtocolDiff, base, new []EntityState, entityID string) {
	baseMap := make(map[string]*EntityState)
	for i := range base {
		baseMap[base[i].ID] = &base[i]
	}

	newMap := make(map[string]*EntityState)
	for i := range new {
		newMap[new[i].ID] = &new[i]
	}

	for id := range baseMap {
		if _, exists := newMap[id]; !exists {
			diff.addItem(DiffTypeRemoved, DiffCategoryEntity, fmt.Sprintf("entities[%s].states[%s]", entityID, id),
				baseMap[id], nil, fmt.Sprintf("state %q removed", id))
		}
	}

	for id := range newMap {
		if _, exists := baseMap[id]; !exists {
			diff.addItem(DiffTypeAdded, DiffCategoryEntity, fmt.Sprintf("entities[%s].states[%s]", entityID, id),
				nil, newMap[id], fmt.Sprintf("state %q added", id))
		}
	}
}

func comparePhases(diff *ProtocolDiff, base, new []Phase, opts DiffOptions) {
	baseMap := make(map[string]*Phase)
	for i := range base {
		baseMap[base[i].ID] = &base[i]
	}

	newMap := make(map[string]*Phase)
	for i := range new {
		newMap[new[i].ID] = &new[i]
	}

	// Find removed phases
	for id, p := range baseMap {
		if _, exists := newMap[id]; !exists {
			diff.addItem(DiffTypeRemoved, DiffCategoryPhase, fmt.Sprintf("phases[%s]", id),
				p, nil, fmt.Sprintf("phase %q removed", p.Name))
		}
	}

	// Find added phases
	for id, p := range newMap {
		if _, exists := baseMap[id]; !exists {
			diff.addItem(DiffTypeAdded, DiffCategoryPhase, fmt.Sprintf("phases[%s]", id),
				nil, p, fmt.Sprintf("phase %q added", p.Name))
		}
	}

	// Compare existing phases
	for id, basePhase := range baseMap {
		newPhase, exists := newMap[id]
		if !exists {
			continue
		}

		if basePhase.Name != newPhase.Name {
			diff.addItem(DiffTypeModified, DiffCategoryPhase, fmt.Sprintf("phases[%s].name", id),
				basePhase.Name, newPhase.Name, fmt.Sprintf("name changed from %q to %q", basePhase.Name, newPhase.Name))
		}

		if basePhase.Parent != newPhase.Parent {
			diff.addItem(DiffTypeModified, DiffCategoryPhase, fmt.Sprintf("phases[%s].parent", id),
				basePhase.Parent, newPhase.Parent, fmt.Sprintf("parent changed from %q to %q", basePhase.Parent, newPhase.Parent))
		}

		if !opts.IgnoreDescriptions && basePhase.Description != newPhase.Description {
			diff.addItem(DiffTypeModified, DiffCategoryPhase, fmt.Sprintf("phases[%s].description", id),
				basePhase.Description, newPhase.Description, "description changed")
		}
	}
}

func compareFlows(diff *ProtocolDiff, base, new []Flow, opts DiffOptions) {
	// Compare flows by index (flows are ordered)
	maxLen := len(base)
	if len(new) > maxLen {
		maxLen = len(new)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(base) {
			// Flow was added
			f := &new[i]
			diff.addItem(DiffTypeAdded, DiffCategoryFlow, fmt.Sprintf("flows[%d]", i),
				nil, f, fmt.Sprintf("flow added: %s→%s %s", f.From, f.To, f.Action))
			continue
		}

		if i >= len(new) {
			// Flow was removed
			f := &base[i]
			diff.addItem(DiffTypeRemoved, DiffCategoryFlow, fmt.Sprintf("flows[%d]", i),
				f, nil, fmt.Sprintf("flow removed: %s→%s %s", f.From, f.To, f.Action))
			continue
		}

		// Compare existing flow
		compareFlow(diff, &base[i], &new[i], i, opts)
	}
}

func compareFlow(diff *ProtocolDiff, base, new *Flow, idx int, opts DiffOptions) {
	path := fmt.Sprintf("flows[%d]", idx)

	// Quick check: if flows are deeply equal, skip
	if !opts.DeepFlowCompare {
		if base.From == new.From && base.To == new.To &&
			base.Action == new.Action && base.Mode == new.Mode {
			return
		}
	}

	if base.From != new.From {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".from",
			base.From, new.From, fmt.Sprintf("from changed from %q to %q", base.From, new.From))
	}

	if base.To != new.To {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".to",
			base.To, new.To, fmt.Sprintf("to changed from %q to %q", base.To, new.To))
	}

	if base.Action != new.Action {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".action",
			base.Action, new.Action, fmt.Sprintf("action changed from %q to %q", base.Action, new.Action))
	}

	if base.Label != new.Label {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".label",
			base.Label, new.Label, fmt.Sprintf("label changed from %q to %q", base.Label, new.Label))
	}

	if base.Mode != new.Mode {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".mode",
			base.Mode, new.Mode, fmt.Sprintf("mode changed from %q to %q", base.Mode, new.Mode))
	}

	if base.Phase != new.Phase {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".phase",
			base.Phase, new.Phase, fmt.Sprintf("phase changed from %q to %q", base.Phase, new.Phase))
	}

	if base.Condition != new.Condition {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".condition",
			base.Condition, new.Condition, fmt.Sprintf("condition changed from %q to %q", base.Condition, new.Condition))
	}

	if !opts.IgnoreDescriptions && base.Description != new.Description {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".description",
			base.Description, new.Description, "description changed")
	}

	// Compare state mutations
	if !reflect.DeepEqual(base.Sets, new.Sets) {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".sets",
			base.Sets, new.Sets, "state mutations changed")
	}

	// Compare security
	if !reflect.DeepEqual(base.Security, new.Security) {
		diff.addItem(DiffTypeModified, DiffCategoryFlow, path+".security",
			base.Security, new.Security, "security requirements changed")
	}
}

func compareMetadata(diff *ProtocolDiff, base, new *ProtocolMetadata, _ DiffOptions) {
	// Handle nil cases
	if base == nil && new == nil {
		return
	}

	if base == nil {
		diff.addItem(DiffTypeAdded, DiffCategoryMetadata, "metadata",
			nil, new, "metadata added")
		return
	}

	if new == nil {
		diff.addItem(DiffTypeRemoved, DiffCategoryMetadata, "metadata",
			base, nil, "metadata removed")
		return
	}

	// Compare tokens
	if !reflect.DeepEqual(base.Tokens, new.Tokens) {
		diff.addItem(DiffTypeModified, DiffCategoryMetadata, "metadata.tokens",
			base.Tokens, new.Tokens, "token definitions changed")
	}

	// Compare components
	if !reflect.DeepEqual(base.Components, new.Components) {
		diff.addItem(DiffTypeModified, DiffCategoryMetadata, "metadata.components",
			base.Components, new.Components, "deployment components changed")
	}

	// Compare trust relations
	if !reflect.DeepEqual(base.TrustRelations, new.TrustRelations) {
		diff.addItem(DiffTypeModified, DiffCategoryMetadata, "metadata.trust_relations",
			base.TrustRelations, new.TrustRelations, "trust relations changed")
	}

	// Compare networks
	if !reflect.DeepEqual(base.Networks, new.Networks) {
		diff.addItem(DiffTypeModified, DiffCategoryMetadata, "metadata.networks",
			base.Networks, new.Networks, "network configurations changed")
	}
}

func (d *ProtocolDiff) addItem(diffType DiffType, category DiffCategory, path string, oldVal, newVal interface{}, summary string) {
	d.Items = append(d.Items, DiffItem{
		Type:     diffType,
		Category: category,
		Path:     path,
		OldValue: oldVal,
		NewValue: newVal,
		Summary:  summary,
	})
}
