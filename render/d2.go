package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// D2Style represents different D2 diagram styles.
type D2Style string

const (
	// D2StyleSequence renders as a sequence diagram.
	D2StyleSequence D2Style = "sequence"
	// D2StyleFlow renders as a data flow diagram.
	D2StyleFlow D2Style = "flow"
	// D2StyleArch renders as an architecture diagram with grouped entities.
	D2StyleArch D2Style = "arch"
)

// D2Renderer renders PIDL protocols as D2 diagrams.
type D2Renderer struct {
	// SequenceRenderOptions contains common sequence diagram options.
	SequenceRenderOptions

	// Style determines the diagram style (sequence, flow, or arch).
	Style D2Style

	// ShowDescriptions includes flow descriptions as tooltips.
	ShowDescriptions bool

	// Direction sets the diagram direction (down, right, left, up).
	Direction string

	// ShowStepTypes includes visual styling for process step types.
	ShowStepTypes bool
}

// NewD2 creates a new D2 renderer with default options (sequence diagram).
func NewD2() *D2Renderer {
	return &D2Renderer{
		SequenceRenderOptions: DefaultSequenceRenderOptions(),
		Style:                 D2StyleSequence,
		ShowDescriptions:      false,
		Direction:             "right",
		ShowStepTypes:         true,
	}
}

// NewD2Flow creates a new D2 renderer for data flow diagrams.
func NewD2Flow() *D2Renderer {
	return &D2Renderer{
		SequenceRenderOptions: DefaultSequenceRenderOptions(),
		Style:                 D2StyleFlow,
		ShowDescriptions:      false,
		Direction:             "right",
		ShowStepTypes:         true,
	}
}

// NewD2Arch creates a new D2 renderer for architecture diagrams.
func NewD2Arch() *D2Renderer {
	return &D2Renderer{
		SequenceRenderOptions: DefaultSequenceRenderOptions(),
		Style:                 D2StyleArch,
		ShowDescriptions:      false,
		Direction:             "right",
		ShowStepTypes:         true,
	}
}

// Format returns the output format.
func (r *D2Renderer) Format() Format {
	switch r.Style {
	case D2StyleFlow:
		return FormatD2Flow
	case D2StyleArch:
		return FormatD2Arch
	default:
		return FormatD2
	}
}

// Render writes the D2 diagram to the writer.
func (r *D2Renderer) Render(w io.Writer, p *pidl.Protocol) error {
	_, err := w.Write([]byte(r.render(p)))
	return err
}

// RenderString returns the D2 diagram as a string.
func (r *D2Renderer) RenderString(p *pidl.Protocol) (string, error) {
	return r.render(p), nil
}

func (r *D2Renderer) render(p *pidl.Protocol) string {
	switch r.Style {
	case D2StyleFlow:
		return r.renderFlow(p)
	case D2StyleArch:
		return r.renderArch(p)
	default:
		return r.renderSequence(p)
	}
}

// renderSequence renders a D2 sequence diagram.
func (r *D2Renderer) renderSequence(p *pidl.Protocol) string {
	var sb strings.Builder

	// Title as a label
	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "title: %s {\n  shape: text\n  near: top-center\n  style.font-size: 24\n}\n\n", p.ProtocolMeta.Name)
	}

	// Declare the sequence diagram shape
	sb.WriteString("sequence: {\n")
	sb.WriteString("  shape: sequence_diagram\n\n")

	// Declare actors
	for _, e := range p.Entities {
		fmt.Fprintf(&sb, "  %s: %s\n", r.sanitizeID(e.ID), e.Name)
	}

	sb.WriteString("\n")

	// Track sequence number for ordering
	seq := 1

	// Track current phase for grouping and nesting
	currentPhase := ""
	phaseStack := []string{}

	for _, f := range p.Flows {
		// Handle phase changes with nesting support
		if f.Phase != currentPhase {
			// Close previous phase groups
			for range phaseStack {
				sb.WriteString("  }\n\n")
			}
			phaseStack = nil

			// Open new phase groups (including parent hierarchy)
			if f.Phase != "" {
				phase := p.PhaseByID(f.Phase)
				if phase != nil {
					phaseStack = r.openPhaseGroups(&sb, p, phase)
				}
			}
			currentPhase = f.Phase
		}

		// Render the flow
		indent := "  "
		for range phaseStack {
			indent += "  "
		}

		seq = r.renderSequenceFlow(&sb, p, f, indent, seq)
	}

	// Close remaining phase groups
	for range phaseStack {
		sb.WriteString("  }\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}

// openPhaseGroups opens D2 groups for a phase and its parent hierarchy, returns the stack.
func (r *D2Renderer) openPhaseGroups(sb *strings.Builder, p *pidl.Protocol, phase *pidl.Phase) []string {
	// Build the hierarchy from root to current phase
	var hierarchy []*pidl.Phase
	current := phase
	for current != nil {
		hierarchy = append([]*pidl.Phase{current}, hierarchy...)
		if current.Parent == "" {
			break
		}
		current = p.PhaseByID(current.Parent)
	}

	// Open groups from root to leaf
	var stack []string
	for i, ph := range hierarchy {
		indent := "  "
		for j := 0; j < i; j++ {
			indent += "  "
		}
		fmt.Fprintf(sb, "%s%s: %s {\n", indent, r.sanitizeID(ph.ID), ph.Name)
		stack = append(stack, ph.ID)
	}

	return stack
}

// renderSequenceFlow renders a single flow in a D2 sequence diagram.
func (r *D2Renderer) renderSequenceFlow(sb *strings.Builder, _ *pidl.Protocol, f pidl.Flow, indent string, seq int) int {
	from := r.sanitizeID(f.From)
	to := r.sanitizeID(f.To)
	label := f.DisplayLabel()

	// Add condition to label if present
	if r.ShowConditions && f.HasCondition() {
		label = fmt.Sprintf("[%s] %s", f.Condition, label)
	}

	// Add mode annotation
	if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
		label = fmt.Sprintf("%s (%s)", label, ann)
	}

	// D2 sequence diagram message syntax
	arrow := r.modeToArrow(f.EffectiveMode())
	fmt.Fprintf(sb, "%sseq%d: %s %s %s: %s", indent, seq, from, arrow, to, label)

	// Add note as tooltip if present
	if r.ShowNotes && f.HasNote() {
		fmt.Fprintf(sb, " {\n%s  tooltip: %s\n%s}", indent, f.Note, indent)
	}

	sb.WriteString("\n")
	seq++

	// Render annotations as separate note messages
	if r.ShowAnnotations && f.HasAnnotations() {
		for _, ann := range f.Annotations {
			prefix := r.annotationPrefix(ann.Type)
			fmt.Fprintf(sb, "%snote%d: %s -> %s: %s%s\n", indent, seq, to, to, prefix, ann.Text)
			seq++
		}
	}

	// Render alternatives as additional flows
	if r.ShowAlternatives && f.HasAlternatives() {
		for _, alt := range f.Alternatives {
			fmt.Fprintf(sb, "%salt%d: [%s] {\n", indent, seq, alt.Condition)
			altIndent := indent + "  "
			for _, altFlow := range alt.Flows {
				seq = r.renderSequenceFlow(sb, nil, altFlow, altIndent, seq)
			}
			fmt.Fprintf(sb, "%s}\n", indent)
			seq++
		}
	}

	return seq
}

// annotationPrefix returns a visual prefix for annotation types.
func (r *D2Renderer) annotationPrefix(t pidl.AnnotationType) string {
	switch t {
	case pidl.AnnotationTypeSecurity:
		return "⚠️ SECURITY: "
	case pidl.AnnotationTypePerformance:
		return "⏱️ PERF: "
	case pidl.AnnotationTypeDeprecated:
		return "🚫 DEPRECATED: "
	case pidl.AnnotationTypeWarning:
		return "⚠️ WARNING: "
	case pidl.AnnotationTypeError:
		return "❌ ERROR: "
	default:
		return ""
	}
}

// renderFlow renders a D2 data flow diagram.
func (r *D2Renderer) renderFlow(p *pidl.Protocol) string {
	var sb strings.Builder

	// Direction
	if r.Direction != "" {
		fmt.Fprintf(&sb, "direction: %s\n\n", r.Direction)
	}

	// Title
	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "title: %s {\n  shape: text\n  near: top-center\n  style.font-size: 24\n}\n\n", p.ProtocolMeta.Name)
	}

	// Render data ports for process specs
	isProcessSpec := p.IsProcessSpec()
	if r.ShowDataPorts && isProcessSpec {
		r.renderDataPorts(&sb, p)
	}

	// Declare entities with shapes based on type
	for _, e := range p.Entities {
		id := r.sanitizeID(e.ID)
		shape := r.entityTypeToShape(e.Type)
		fmt.Fprintf(&sb, "%s: %s {\n  shape: %s\n", id, e.Name, shape)
		if e.Description != "" && r.ShowDescriptions {
			fmt.Fprintf(&sb, "  tooltip: %s\n", e.Description)
		}
		// Add step type styling for process specs
		if r.ShowStepTypes && isProcessSpec && e.IsProcessStep() {
			fill, stroke := r.stepTypeColors(e.StepType)
			fmt.Fprintf(&sb, "  style.fill: \"%s\"\n", fill)
			fmt.Fprintf(&sb, "  style.stroke: \"%s\"\n", stroke)
		}
		sb.WriteString("}\n")
	}

	sb.WriteString("\n")

	// Render data port connections for process specs
	if r.ShowDataPorts && isProcessSpec {
		r.renderDataPortConnections(&sb, p)
	}

	// Render flows as connections
	for i, f := range p.Flows {
		from := r.sanitizeID(f.From)
		to := r.sanitizeID(f.To)
		label := f.DisplayLabel()

		// Add mode annotation
		if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
			label = fmt.Sprintf("%s (%s)", label, ann)
		}

		// Connection with label
		arrow := r.modeToD2Arrow(f.EffectiveMode())
		fmt.Fprintf(&sb, "%s %s %s: %d. %s\n", from, arrow, to, i+1, label)
	}

	return sb.String()
}

// renderDataPorts renders data port nodes for a process spec.
func (r *D2Renderer) renderDataPorts(sb *strings.Builder, p *pidl.Protocol) {
	// Collect unique data ports
	portsSeen := make(map[string]bool)

	sb.WriteString("# Data Ports\n")

	for _, e := range p.Entities {
		if !e.IsProcessStep() {
			continue
		}

		// Render inputs
		for _, port := range e.Inputs {
			portID := r.dataPortID(port)
			if portsSeen[portID] {
				continue
			}
			portsSeen[portID] = true

			shape := r.dataPortKindToShape(port.Kind)
			fill, stroke := r.dataPortKindToColor(port.Kind)
			icon := r.dataPortIcon(port.Kind)

			label := fmt.Sprintf("%s %s", icon, port.Name)
			fmt.Fprintf(sb, "%s: %s {\n", portID, label)
			fmt.Fprintf(sb, "  shape: %s\n", shape)
			fmt.Fprintf(sb, "  style.fill: \"%s\"\n", fill)
			fmt.Fprintf(sb, "  style.stroke: \"%s\"\n", stroke)
			if port.Description != "" {
				fmt.Fprintf(sb, "  tooltip: %s\n", port.Description)
			}
			sb.WriteString("}\n")
		}

		// Render outputs
		for _, port := range e.Outputs {
			portID := r.dataPortID(port)
			if portsSeen[portID] {
				continue
			}
			portsSeen[portID] = true

			shape := r.dataPortKindToShape(port.Kind)
			fill, stroke := r.dataPortKindToColor(port.Kind)
			icon := r.dataPortIcon(port.Kind)

			label := fmt.Sprintf("%s %s", icon, port.Name)
			fmt.Fprintf(sb, "%s: %s {\n", portID, label)
			fmt.Fprintf(sb, "  shape: %s\n", shape)
			fmt.Fprintf(sb, "  style.fill: \"%s\"\n", fill)
			fmt.Fprintf(sb, "  style.stroke: \"%s\"\n", stroke)
			if port.Description != "" {
				fmt.Fprintf(sb, "  tooltip: %s\n", port.Description)
			}
			sb.WriteString("}\n")
		}
	}

	sb.WriteString("\n")
}

// renderDataPortConnections renders connections between data ports and steps.
func (r *D2Renderer) renderDataPortConnections(sb *strings.Builder, p *pidl.Protocol) {
	sb.WriteString("# Data Port Connections\n")

	for _, e := range p.Entities {
		if !e.IsProcessStep() {
			continue
		}

		stepID := r.sanitizeID(e.ID)

		// Connect inputs to step
		for _, port := range e.Inputs {
			portID := r.dataPortID(port)
			style := "<-"
			if port.Required {
				style = "<--" // solid line for required
			}
			fmt.Fprintf(sb, "%s %s %s\n", portID, "->", stepID)
			_ = style // reserved for future use
		}

		// Connect step to outputs
		for _, port := range e.Outputs {
			portID := r.dataPortID(port)
			fmt.Fprintf(sb, "%s -> %s\n", stepID, portID)
		}
	}

	sb.WriteString("\n")
}

// dataPortID generates a sanitized ID for a data port.
func (r *D2Renderer) dataPortID(port pidl.DataPort) string {
	return r.sanitizeID("port_" + port.Name)
}

// renderArch renders a D2 architecture diagram with phase groupings.
func (r *D2Renderer) renderArch(p *pidl.Protocol) string {
	var sb strings.Builder

	// Direction
	if r.Direction != "" {
		fmt.Fprintf(&sb, "direction: %s\n\n", r.Direction)
	}

	// Title
	if r.Title && p.ProtocolMeta.Name != "" {
		fmt.Fprintf(&sb, "title: %s {\n  shape: text\n  near: top-center\n  style.font-size: 24\n}\n\n", p.ProtocolMeta.Name)
	}

	isProcessSpec := p.IsProcessSpec()

	// Render data ports grouped by kind for process specs
	if r.ShowDataPorts && isProcessSpec {
		r.renderDataPortsGrouped(&sb, p)
	}

	// Group entities by type for architecture view
	entityGroups := make(map[string][]pidl.Entity)
	for _, e := range p.Entities {
		group := r.entityTypeToGroup(e.Type)
		entityGroups[group] = append(entityGroups[group], e)
	}

	// Render grouped entities
	for group, entities := range entityGroups {
		if group != "" {
			fmt.Fprintf(&sb, "%s: %s {\n", r.sanitizeID(group), group)
			for _, e := range entities {
				id := r.sanitizeID(e.ID)
				shape := r.entityTypeToShape(e.Type)
				fmt.Fprintf(&sb, "  %s: %s {\n    shape: %s\n", id, e.Name, shape)
				// Add step type styling for process specs
				if r.ShowStepTypes && isProcessSpec && e.IsProcessStep() {
					fill, stroke := r.stepTypeColors(e.StepType)
					fmt.Fprintf(&sb, "    style.fill: \"%s\"\n", fill)
					fmt.Fprintf(&sb, "    style.stroke: \"%s\"\n", stroke)
				}
				sb.WriteString("  }\n")
			}
			sb.WriteString("}\n\n")
		} else {
			// Ungrouped entities at top level
			for _, e := range entities {
				id := r.sanitizeID(e.ID)
				shape := r.entityTypeToShape(e.Type)
				fmt.Fprintf(&sb, "%s: %s {\n  shape: %s\n", id, e.Name, shape)
				// Add step type styling for process specs
				if r.ShowStepTypes && isProcessSpec && e.IsProcessStep() {
					fill, stroke := r.stepTypeColors(e.StepType)
					fmt.Fprintf(&sb, "  style.fill: \"%s\"\n", fill)
					fmt.Fprintf(&sb, "  style.stroke: \"%s\"\n", stroke)
				}
				sb.WriteString("}\n")
			}
			sb.WriteString("\n")
		}
	}

	// Render data port connections for process specs
	if r.ShowDataPorts && isProcessSpec {
		r.renderDataPortConnectionsArch(&sb, p)
	}

	// Render flows as connections
	for i, f := range p.Flows {
		from := r.qualifiedID(p, f.From)
		to := r.qualifiedID(p, f.To)
		label := f.DisplayLabel()

		if ann := r.modeAnnotation(f.EffectiveMode()); ann != "" {
			label = fmt.Sprintf("%s (%s)", label, ann)
		}

		arrow := r.modeToD2Arrow(f.EffectiveMode())
		fmt.Fprintf(&sb, "%s %s %s: %d. %s\n", from, arrow, to, i+1, label)
	}

	return sb.String()
}

// renderDataPortsGrouped renders data ports grouped by kind for architecture diagrams.
func (r *D2Renderer) renderDataPortsGrouped(sb *strings.Builder, p *pidl.Protocol) {
	// Group ports by kind
	portsByKind := make(map[pidl.DataPortKind][]pidl.DataPort)
	portsSeen := make(map[string]bool)

	for _, e := range p.Entities {
		if !e.IsProcessStep() {
			continue
		}
		for _, port := range e.Inputs {
			if !portsSeen[port.Name] {
				portsSeen[port.Name] = true
				portsByKind[port.Kind] = append(portsByKind[port.Kind], port)
			}
		}
		for _, port := range e.Outputs {
			if !portsSeen[port.Name] {
				portsSeen[port.Name] = true
				portsByKind[port.Kind] = append(portsByKind[port.Kind], port)
			}
		}
	}

	// Render each kind group
	kindOrder := []pidl.DataPortKind{
		pidl.DataPortKindFile,
		pidl.DataPortKindDatabase,
		pidl.DataPortKindAPI,
		pidl.DataPortKindQueue,
		pidl.DataPortKindStream,
		pidl.DataPortKindObject,
	}

	for _, kind := range kindOrder {
		ports := portsByKind[kind]
		if len(ports) == 0 {
			continue
		}

		groupName := r.dataPortKindGroupName(kind)
		groupID := r.sanitizeID(groupName)
		fill, stroke := r.dataPortKindToColor(kind)

		fmt.Fprintf(sb, "%s: %s {\n", groupID, groupName)
		fmt.Fprintf(sb, "  style.fill: \"%s\"\n", fill)
		fmt.Fprintf(sb, "  style.stroke: \"%s\"\n", stroke)

		for _, port := range ports {
			portID := r.sanitizeID(port.Name)
			shape := r.dataPortKindToShape(kind)
			icon := r.dataPortIcon(kind)
			label := fmt.Sprintf("%s %s", icon, port.Name)

			fmt.Fprintf(sb, "  %s: %s {\n", portID, label)
			fmt.Fprintf(sb, "    shape: %s\n", shape)
			if port.Description != "" {
				fmt.Fprintf(sb, "    tooltip: %s\n", port.Description)
			}
			sb.WriteString("  }\n")
		}

		sb.WriteString("}\n\n")
	}
}

// renderDataPortConnectionsArch renders data port connections for architecture diagrams.
func (r *D2Renderer) renderDataPortConnectionsArch(sb *strings.Builder, p *pidl.Protocol) {
	sb.WriteString("# Data Port Connections\n")

	for _, e := range p.Entities {
		if !e.IsProcessStep() {
			continue
		}

		stepID := r.qualifiedID(p, e.ID)

		// Connect inputs to step
		for _, port := range e.Inputs {
			portID := r.qualifiedDataPortID(port)
			fmt.Fprintf(sb, "%s -> %s\n", portID, stepID)
		}

		// Connect step to outputs
		for _, port := range e.Outputs {
			portID := r.qualifiedDataPortID(port)
			fmt.Fprintf(sb, "%s -> %s\n", stepID, portID)
		}
	}

	sb.WriteString("\n")
}

// dataPortKindGroupName returns a display name for a data port kind group.
func (r *D2Renderer) dataPortKindGroupName(kind pidl.DataPortKind) string {
	switch kind {
	case pidl.DataPortKindFile:
		return "Files"
	case pidl.DataPortKindObject:
		return "Objects"
	case pidl.DataPortKindAPI:
		return "APIs"
	case pidl.DataPortKindDatabase:
		return "Databases"
	case pidl.DataPortKindQueue:
		return "Queues"
	case pidl.DataPortKindStream:
		return "Streams"
	default:
		return "Data"
	}
}

// qualifiedDataPortID generates a qualified ID for a data port in architecture diagrams.
func (r *D2Renderer) qualifiedDataPortID(port pidl.DataPort) string {
	groupName := r.dataPortKindGroupName(port.Kind)
	return fmt.Sprintf("%s.%s", r.sanitizeID(groupName), r.sanitizeID(port.Name))
}

func (r *D2Renderer) sanitizeID(id string) string {
	// D2 IDs: replace hyphens with underscores, ensure valid identifier
	result := strings.ReplaceAll(id, "-", "_")
	return result
}

func (r *D2Renderer) modeToArrow(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult, pidl.FlowModeEvent:
		return "<-"
	default:
		return "->"
	}
}

func (r *D2Renderer) modeToD2Arrow(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeResponse, pidl.FlowModeToolResult:
		return "<--"
	case pidl.FlowModeEvent:
		return "<-"
	default:
		return "->"
	}
}

func (r *D2Renderer) modeAnnotation(mode pidl.FlowMode) string {
	switch mode {
	case pidl.FlowModeRedirect:
		return "redirect"
	case pidl.FlowModeCallback:
		return "callback"
	case pidl.FlowModeToolCall:
		return "tool"
	case pidl.FlowModeToolResult:
		return "result"
	default:
		return ""
	}
}

func (r *D2Renderer) entityTypeToShape(t pidl.EntityType) string {
	switch t {
	case pidl.EntityTypeUser:
		return "person"
	case pidl.EntityTypeBrowser:
		return "rectangle"
	case pidl.EntityTypeClient:
		return "rectangle"
	case pidl.EntityTypeServer, pidl.EntityTypeResourceServer, pidl.EntityTypeAuthorizationServer:
		return "cylinder"
	case pidl.EntityTypeAgent, pidl.EntityTypeDelegatedAgent:
		return "hexagon"
	case pidl.EntityTypeToolServer, pidl.EntityTypeTool:
		return "package"
	case pidl.EntityTypeIdentityProvider, pidl.EntityTypeServiceProvider:
		return "cloud"
	default:
		return "rectangle"
	}
}

func (r *D2Renderer) entityTypeToGroup(t pidl.EntityType) string {
	switch t {
	case pidl.EntityTypeUser, pidl.EntityTypeBrowser:
		return "Users"
	case pidl.EntityTypeClient:
		return "Clients"
	case pidl.EntityTypeServer, pidl.EntityTypeResourceServer, pidl.EntityTypeAuthorizationServer:
		return "Servers"
	case pidl.EntityTypeAgent, pidl.EntityTypeDelegatedAgent:
		return "Agents"
	case pidl.EntityTypeToolServer, pidl.EntityTypeTool:
		return "Tools"
	case pidl.EntityTypeIdentityProvider, pidl.EntityTypeServiceProvider:
		return "Providers"
	default:
		return ""
	}
}

func (r *D2Renderer) qualifiedID(p *pidl.Protocol, entityID string) string {
	entity := p.EntityByID(entityID)
	if entity == nil {
		return r.sanitizeID(entityID)
	}

	group := r.entityTypeToGroup(entity.Type)
	if group != "" {
		return fmt.Sprintf("%s.%s", r.sanitizeID(group), r.sanitizeID(entityID))
	}
	return r.sanitizeID(entityID)
}

// stepTypeColors returns fill and stroke colors for a step type.
func (r *D2Renderer) stepTypeColors(st pidl.StepType) (fill, stroke string) {
	switch st {
	case pidl.StepTypeDeterministic:
		return "#E3F2FD", "#1976D2"
	case pidl.StepTypeLLM:
		return "#F3E5F5", "#7B1FA2"
	case pidl.StepTypeHuman:
		return "#E8F5E9", "#388E3C"
	case pidl.StepTypeExternal:
		return "#FFF3E0", "#F57C00"
	case pidl.StepTypeTool:
		return "#ECEFF1", "#607D8B"
	default:
		return "#FFFFFF", "#000000"
	}
}

// dataPortKindToShape returns the D2 shape for a data port kind.
func (r *D2Renderer) dataPortKindToShape(kind pidl.DataPortKind) string {
	switch kind {
	case pidl.DataPortKindFile:
		return "page"
	case pidl.DataPortKindObject:
		return "rectangle"
	case pidl.DataPortKindAPI:
		return "cloud"
	case pidl.DataPortKindDatabase:
		return "cylinder"
	case pidl.DataPortKindQueue:
		return "queue"
	case pidl.DataPortKindStream:
		return "parallelogram"
	default:
		return "rectangle"
	}
}

// dataPortKindToColor returns fill and stroke colors for a data port kind.
func (r *D2Renderer) dataPortKindToColor(kind pidl.DataPortKind) (fill, stroke string) {
	switch kind {
	case pidl.DataPortKindFile:
		return "#FFF8E1", "#FFA000" // amber
	case pidl.DataPortKindObject:
		return "#E3F2FD", "#1976D2" // blue
	case pidl.DataPortKindAPI:
		return "#E8F5E9", "#388E3C" // green
	case pidl.DataPortKindDatabase:
		return "#FCE4EC", "#C2185B" // pink
	case pidl.DataPortKindQueue:
		return "#F3E5F5", "#7B1FA2" // purple
	case pidl.DataPortKindStream:
		return "#E0F7FA", "#00838F" // cyan
	default:
		return "#FAFAFA", "#9E9E9E" // gray
	}
}

// dataPortIcon returns an emoji icon for a data port kind.
func (r *D2Renderer) dataPortIcon(kind pidl.DataPortKind) string {
	switch kind {
	case pidl.DataPortKindFile:
		return "📄"
	case pidl.DataPortKindObject:
		return "📦"
	case pidl.DataPortKindAPI:
		return "🌐"
	case pidl.DataPortKindDatabase:
		return "🗄️"
	case pidl.DataPortKindQueue:
		return "📬"
	case pidl.DataPortKindStream:
		return "🌊"
	default:
		return "📋"
	}
}
