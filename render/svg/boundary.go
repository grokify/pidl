package svg

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grokify/pidl"
)

// BoundaryStyle represents the visual style for a network boundary.
type BoundaryStyle string

const (
	// BoundaryStyleTrusted is for internal/protected networks (green).
	BoundaryStyleTrusted BoundaryStyle = "trusted"
	// BoundaryStyleDMZ is for perimeter/semi-trusted networks (blue).
	BoundaryStyleDMZ BoundaryStyle = "dmz"
	// BoundaryStyleExternal is for untrusted/public networks (red).
	BoundaryStyleExternal BoundaryStyle = "external"
	// BoundaryStyleCloud is for cloud provider boundaries (gray).
	BoundaryStyleCloud BoundaryStyle = "cloud"
)

// Boundary represents a resolved network boundary with its entities.
type Boundary struct {
	// ID is the boundary identifier.
	ID string
	// Name is the display name.
	Name string
	// Style is the visual style.
	Style BoundaryStyle
	// Entities are the entity IDs within this boundary.
	Entities []string
	// Description is tooltip text.
	Description string
	// Color overrides the default color.
	Color string
}

// BoundaryLayout contains position data for a boundary region.
type BoundaryLayout struct {
	// ID is the boundary identifier.
	ID string
	// Name is the display name.
	Name string
	// Style is the visual style.
	Style BoundaryStyle
	// X is the left edge.
	X int
	// Y is the top edge.
	Y int
	// Width of the boundary region.
	Width int
	// Height of the boundary region.
	Height int
	// Entities contains layout for entities within this boundary.
	Entities []EntityLayout
}

// EntityLayout contains position data for an entity within a boundary.
type EntityLayout struct {
	// ID is the entity ID.
	ID string
	// Name is the display name.
	Name string
	// X is the center X coordinate.
	X int
	// Y is the center Y coordinate.
	Y int
	// Width of the entity box.
	Width int
	// Height of the entity box.
	Height int
}

// ConnectionLayout contains position data for a connection between entities.
type ConnectionLayout struct {
	// FromID is the source entity ID.
	FromID string
	// ToID is the target entity ID.
	ToID string
	// Label is the connection label (aggregated from flows).
	Label string
	// FromX, FromY is the start point.
	FromX, FromY int
	// ToX, ToY is the end point.
	ToX, ToY int
	// CrossesBoundary indicates if the connection crosses a boundary.
	CrossesBoundary bool
}

// NetworkLayout contains the complete layout for a network diagram.
type NetworkLayout struct {
	// Width is the total diagram width.
	Width int
	// Height is the total diagram height.
	Height int
	// Boundaries contains layout for each boundary.
	Boundaries []BoundaryLayout
	// Connections contains layout for connections.
	Connections []ConnectionLayout
}

// NetworkLayoutConfig contains configuration for network layout calculations.
type NetworkLayoutConfig struct {
	// Padding around the entire diagram.
	Padding int
	// BoundaryPadding inside boundary regions.
	BoundaryPadding int
	// BoundarySpacing between boundaries.
	BoundarySpacing int
	// EntityWidth is the width of entity boxes.
	EntityWidth int
	// EntityHeight is the height of entity boxes.
	EntityHeight int
	// EntitySpacing between entities within a boundary.
	EntitySpacing int
	// Direction is horizontal or vertical layout.
	Direction string
}

// DefaultNetworkLayoutConfig returns the default network layout configuration.
func DefaultNetworkLayoutConfig() NetworkLayoutConfig {
	return NetworkLayoutConfig{
		Padding:         30,
		BoundaryPadding: 20,
		BoundarySpacing: 40,
		EntityWidth:     100,
		EntityHeight:    50,
		EntitySpacing:   30,
		Direction:       "horizontal",
	}
}

// ResolveBoundaries resolves network boundaries from a protocol.
// It applies the resolution order: CLI overrides > metadata.networks > entity.metadata.network > default.
func ResolveBoundaries(p *pidl.Protocol, cliOverrides map[string][]string) []Boundary {
	// Map entity ID to boundary ID
	entityToBoundary := make(map[string]string)

	// 1. Start with entity metadata inference
	for _, e := range p.Entities {
		if e.Metadata != nil && e.Metadata.Network != "" {
			entityToBoundary[e.ID] = e.Metadata.Network
		}
	}

	// 2. Apply metadata.networks explicit entities (overrides inference)
	if p.Metadata != nil && p.Metadata.Networks != nil {
		for boundaryID, config := range p.Metadata.Networks {
			for _, entityID := range config.Entities {
				entityToBoundary[entityID] = boundaryID
			}
		}
	}

	// 3. Apply CLI overrides (highest priority)
	for boundaryID, entities := range cliOverrides {
		for _, entityID := range entities {
			entityToBoundary[entityID] = boundaryID
		}
	}

	// Group entities by boundary
	boundaryEntities := make(map[string][]string)
	for entityID, boundaryID := range entityToBoundary {
		boundaryEntities[boundaryID] = append(boundaryEntities[boundaryID], entityID)
	}

	// Add ungrouped entities to "default" boundary
	for _, e := range p.Entities {
		if _, ok := entityToBoundary[e.ID]; !ok {
			boundaryEntities["default"] = append(boundaryEntities["default"], e.ID)
		}
	}

	// Build boundary list
	var boundaries []Boundary
	boundaryOrder := getBoundaryOrder(p, boundaryEntities)

	for _, boundaryID := range boundaryOrder {
		entities := boundaryEntities[boundaryID]
		if len(entities) == 0 {
			continue
		}

		// Sort entities for consistent ordering
		sort.Strings(entities)

		boundary := Boundary{
			ID:       boundaryID,
			Name:     getBoundaryName(p, boundaryID),
			Style:    getBoundaryStyle(p, boundaryID),
			Entities: entities,
		}

		// Apply config from metadata
		if p.Metadata != nil && p.Metadata.Networks != nil {
			if config, ok := p.Metadata.Networks[boundaryID]; ok {
				if config.Description != "" {
					boundary.Description = config.Description
				}
				if config.Color != "" {
					boundary.Color = config.Color
				}
			}
		}

		boundaries = append(boundaries, boundary)
	}

	return boundaries
}

// getBoundaryOrder determines the order of boundaries.
func getBoundaryOrder(p *pidl.Protocol, boundaryEntities map[string][]string) []string {
	// If explicit order is specified in metadata, use it
	if p.Metadata != nil && p.Metadata.NetworkLayout != nil && len(p.Metadata.NetworkLayout.Order) > 0 {
		order := p.Metadata.NetworkLayout.Order
		// Add any boundaries not in the explicit order
		orderSet := make(map[string]bool)
		for _, id := range order {
			orderSet[id] = true
		}
		for id := range boundaryEntities {
			if !orderSet[id] {
				order = append(order, id)
			}
		}
		return order
	}

	// Default order: external, dmz, trusted, cloud, default, then alphabetical
	styleOrder := map[string]int{
		"external": 0,
		"dmz":      1,
		"trusted":  2,
		"cloud":    3,
		"default":  4,
	}

	var ids []string
	for id := range boundaryEntities {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		styleI := string(getBoundaryStyle(p, ids[i]))
		styleJ := string(getBoundaryStyle(p, ids[j]))
		orderI, okI := styleOrder[styleI]
		orderJ, okJ := styleOrder[styleJ]
		if okI && okJ && orderI != orderJ {
			return orderI < orderJ
		}
		return ids[i] < ids[j]
	})

	return ids
}

// getBoundaryName returns the display name for a boundary.
func getBoundaryName(p *pidl.Protocol, boundaryID string) string {
	if p.Metadata != nil && p.Metadata.Networks != nil {
		if config, ok := p.Metadata.Networks[boundaryID]; ok && config.Name != "" {
			return config.Name
		}
	}
	// Default: capitalize the ID
	return strings.Title(strings.ReplaceAll(boundaryID, "_", " ")) //nolint:staticcheck
}

// getBoundaryStyle returns the visual style for a boundary.
func getBoundaryStyle(p *pidl.Protocol, boundaryID string) BoundaryStyle {
	if p.Metadata != nil && p.Metadata.Networks != nil {
		if config, ok := p.Metadata.Networks[boundaryID]; ok && config.Style != "" {
			return BoundaryStyle(config.Style)
		}
	}
	// Default styles based on ID
	switch boundaryID {
	case "internal", "trusted", "private":
		return BoundaryStyleTrusted
	case "dmz", "perimeter":
		return BoundaryStyleDMZ
	case "external", "public", "client":
		return BoundaryStyleExternal
	case "cloud", "aws", "azure", "gcp":
		return BoundaryStyleCloud
	default:
		return BoundaryStyleExternal
	}
}

// CalculateNetworkLayout computes positions for all network diagram elements.
func CalculateNetworkLayout(boundaries []Boundary, connections []ConnectionLayout, config NetworkLayoutConfig) NetworkLayout {
	// Create entity ID to position map
	entityPositions := make(map[string]struct{ X, Y int })

	var layout NetworkLayout
	if config.Direction == "vertical" {
		layout = calculateVerticalLayout(boundaries, config, entityPositions)
	} else {
		layout = calculateHorizontalLayout(boundaries, config, entityPositions)
	}

	// Copy connections and update endpoints based on entity positions
	layout.Connections = make([]ConnectionLayout, len(connections))
	copy(layout.Connections, connections)

	for i := range layout.Connections {
		conn := &layout.Connections[i]
		if pos, ok := entityPositions[conn.FromID]; ok {
			conn.FromX = pos.X
			conn.FromY = pos.Y
		}
		if pos, ok := entityPositions[conn.ToID]; ok {
			conn.ToX = pos.X
			conn.ToY = pos.Y
		}
	}

	return layout
}

func calculateHorizontalLayout(boundaries []Boundary, config NetworkLayoutConfig, entityPositions map[string]struct{ X, Y int }) NetworkLayout {
	layout := NetworkLayout{
		Boundaries: make([]BoundaryLayout, len(boundaries)),
	}

	currentX := config.Padding

	for i, boundary := range boundaries {
		// Calculate boundary dimensions based on entities
		numEntities := len(boundary.Entities)
		if numEntities == 0 {
			numEntities = 1
		}

		boundaryWidth := config.BoundaryPadding*2 + config.EntityWidth
		boundaryHeight := config.BoundaryPadding*2 + numEntities*config.EntityHeight + (numEntities-1)*config.EntitySpacing

		bl := BoundaryLayout{
			ID:       boundary.ID,
			Name:     boundary.Name,
			Style:    boundary.Style,
			X:        currentX,
			Y:        config.Padding,
			Width:    boundaryWidth,
			Height:   boundaryHeight,
			Entities: make([]EntityLayout, len(boundary.Entities)),
		}

		// Position entities within boundary
		entityX := currentX + config.BoundaryPadding + config.EntityWidth/2
		entityY := config.Padding + config.BoundaryPadding + config.EntityHeight/2

		for j, entityID := range boundary.Entities {
			bl.Entities[j] = EntityLayout{
				ID:     entityID,
				X:      entityX,
				Y:      entityY,
				Width:  config.EntityWidth,
				Height: config.EntityHeight,
			}
			entityPositions[entityID] = struct{ X, Y int }{entityX, entityY}
			entityY += config.EntityHeight + config.EntitySpacing
		}

		layout.Boundaries[i] = bl
		currentX += boundaryWidth + config.BoundarySpacing
	}

	// Calculate total dimensions
	if len(layout.Boundaries) > 0 {
		lastBoundary := layout.Boundaries[len(layout.Boundaries)-1]
		layout.Width = lastBoundary.X + lastBoundary.Width + config.Padding

		maxHeight := 0
		for _, bl := range layout.Boundaries {
			if bl.Y+bl.Height > maxHeight {
				maxHeight = bl.Y + bl.Height
			}
		}
		layout.Height = maxHeight + config.Padding
	}

	return layout
}

func calculateVerticalLayout(boundaries []Boundary, config NetworkLayoutConfig, entityPositions map[string]struct{ X, Y int }) NetworkLayout {
	layout := NetworkLayout{
		Boundaries: make([]BoundaryLayout, len(boundaries)),
	}

	currentY := config.Padding

	for i, boundary := range boundaries {
		numEntities := len(boundary.Entities)
		if numEntities == 0 {
			numEntities = 1
		}

		boundaryWidth := config.BoundaryPadding*2 + numEntities*config.EntityWidth + (numEntities-1)*config.EntitySpacing
		boundaryHeight := config.BoundaryPadding*2 + config.EntityHeight

		bl := BoundaryLayout{
			ID:       boundary.ID,
			Name:     boundary.Name,
			Style:    boundary.Style,
			X:        config.Padding,
			Y:        currentY,
			Width:    boundaryWidth,
			Height:   boundaryHeight,
			Entities: make([]EntityLayout, len(boundary.Entities)),
		}

		entityX := config.Padding + config.BoundaryPadding + config.EntityWidth/2
		entityY := currentY + config.BoundaryPadding + config.EntityHeight/2

		for j, entityID := range boundary.Entities {
			bl.Entities[j] = EntityLayout{
				ID:     entityID,
				X:      entityX,
				Y:      entityY,
				Width:  config.EntityWidth,
				Height: config.EntityHeight,
			}
			entityPositions[entityID] = struct{ X, Y int }{entityX, entityY}
			entityX += config.EntityWidth + config.EntitySpacing
		}

		layout.Boundaries[i] = bl
		currentY += boundaryHeight + config.BoundarySpacing
	}

	// Calculate total dimensions
	if len(layout.Boundaries) > 0 {
		lastBoundary := layout.Boundaries[len(layout.Boundaries)-1]
		layout.Height = lastBoundary.Y + lastBoundary.Height + config.Padding

		maxWidth := 0
		for _, bl := range layout.Boundaries {
			if bl.X+bl.Width > maxWidth {
				maxWidth = bl.X + bl.Width
			}
		}
		layout.Width = maxWidth + config.Padding
	}

	return layout
}

// BoundaryStyleColors returns the default colors for a boundary style.
func BoundaryStyleColors(style BoundaryStyle) (fill, stroke string) {
	switch style {
	case BoundaryStyleTrusted:
		return "rgba(39, 103, 73, 0.1)", "#276749"
	case BoundaryStyleDMZ:
		return "rgba(49, 130, 206, 0.1)", "#3182ce"
	case BoundaryStyleExternal:
		return "rgba(197, 48, 48, 0.1)", "#c53030"
	case BoundaryStyleCloud:
		return "rgba(113, 128, 150, 0.1)", "#718096"
	default:
		return "rgba(113, 128, 150, 0.1)", "#718096"
	}
}

// BoundaryStrokeDashArray returns the stroke dash array for a boundary style.
func BoundaryStrokeDashArray(style BoundaryStyle) string {
	switch style {
	case BoundaryStyleTrusted:
		return "" // solid
	case BoundaryStyleDMZ:
		return "8,4" // dashed
	case BoundaryStyleExternal:
		return "4,4" // dotted
	case BoundaryStyleCloud:
		return "" // solid (cloud shape handles it)
	default:
		return "4,4"
	}
}

// GenerateNetworkCSS generates CSS for network diagrams.
func GenerateNetworkCSS(theme Theme) string {
	var sb strings.Builder

	// Start with base theme
	sb.WriteString(GenerateCSS(theme))

	// Add network-specific styles
	sb.WriteString(`
    /* Network boundary styles */
    .boundary {
      stroke-width: 2;
    }

    .boundary-label {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 12px;
      font-weight: 600;
      fill: var(--color-text);
    }

    .entity-box {
      fill: var(--color-participant-bg);
      stroke: var(--color-line);
      stroke-width: 1;
    }

    .entity-text {
      fill: var(--color-participant-text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 11px;
      font-weight: 500;
      text-anchor: middle;
      dominant-baseline: central;
    }

    .connection {
      fill: none;
      stroke: var(--color-line);
      stroke-width: 1.5;
    }

    .connection-cross-boundary {
      stroke: var(--color-accent);
      stroke-width: 2;
    }

    .connection-label {
      fill: var(--color-text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 10px;
    }

    /* Boundary style colors */
    .boundary-trusted {
      fill: rgba(39, 103, 73, 0.1);
      stroke: #276749;
    }

    .boundary-dmz {
      fill: rgba(49, 130, 206, 0.1);
      stroke: #3182ce;
      stroke-dasharray: 8,4;
    }

    .boundary-external {
      fill: rgba(197, 48, 48, 0.1);
      stroke: #c53030;
      stroke-dasharray: 4,4;
    }

    .boundary-cloud {
      fill: rgba(113, 128, 150, 0.1);
      stroke: #718096;
    }
`)

	return sb.String()
}

// AggregateConnections aggregates flows between entities into connections.
func AggregateConnections(p *pidl.Protocol, entityToBoundary map[string]string) []ConnectionLayout {
	// Group flows by from-to pair
	type pair struct{ from, to string }
	flowsByPair := make(map[pair][]string)

	for _, f := range p.Flows {
		key := pair{f.From, f.To}
		flowsByPair[key] = append(flowsByPair[key], f.DisplayLabel())
	}

	var connections []ConnectionLayout
	for key, labels := range flowsByPair {
		// Determine if crosses boundary
		fromBoundary := entityToBoundary[key.from]
		toBoundary := entityToBoundary[key.to]
		crossesBoundary := fromBoundary != toBoundary

		conn := ConnectionLayout{
			FromID:          key.from,
			ToID:            key.to,
			Label:           aggregateLabels(labels),
			CrossesBoundary: crossesBoundary,
		}
		connections = append(connections, conn)
	}

	return connections
}

func aggregateLabels(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	if len(labels) == 1 {
		return labels[0]
	}
	if len(labels) <= 3 {
		return strings.Join(labels, ", ")
	}
	return fmt.Sprintf("%s (+%d more)", labels[0], len(labels)-1)
}
