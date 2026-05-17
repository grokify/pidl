// Package svg provides SVG rendering utilities for PIDL protocols.
package svg

import "fmt"

// Layout contains calculated positions for SVG diagram elements.
type Layout struct {
	// Width is the total SVG width.
	Width int
	// Height is the total SVG height.
	Height int
	// Participants contains position data for each participant.
	Participants []ParticipantLayout
	// Messages contains position data for each message.
	Messages []MessageLayout
}

// ParticipantLayout contains position data for a participant.
type ParticipantLayout struct {
	// ID is the entity ID.
	ID string
	// Name is the display name.
	Name string
	// Index is the participant order (0-based).
	Index int
	// BoxX is the top-left X of the participant box.
	BoxX int
	// BoxY is the top-left Y of the participant box.
	BoxY int
	// BoxWidth is the participant box width.
	BoxWidth int
	// BoxHeight is the participant box height.
	BoxHeight int
	// CenterX is the X coordinate of the lifeline.
	CenterX int
	// LifelineStartY is where the lifeline starts.
	LifelineStartY int
	// LifelineEndY is where the lifeline ends.
	LifelineEndY int
}

// MessageLayout contains position data for a message arrow.
type MessageLayout struct {
	// Step is the message number (1-based).
	Step int
	// Label is the message text.
	Label string
	// FromX is the start X coordinate.
	FromX int
	// ToX is the end X coordinate.
	ToX int
	// Y is the Y coordinate of the message line.
	Y int
	// IsDashed indicates if the line should be dashed (for responses).
	IsDashed bool
	// IsReverse indicates if the arrow points left (to < from).
	IsReverse bool
	// PathD is the SVG path data for the message line.
	PathD string
}

// LayoutConfig contains configuration for layout calculations.
type LayoutConfig struct {
	// Padding around the entire diagram.
	Padding int
	// ParticipantBoxWidth is the width of participant boxes.
	ParticipantBoxWidth int
	// ParticipantBoxHeight is the height of participant boxes.
	ParticipantBoxHeight int
	// ParticipantSpacing is the horizontal space between participants.
	ParticipantSpacing int
	// MessageSpacing is the vertical space between messages.
	MessageSpacing int
	// MessageStartY is where messages begin (below participant boxes).
	MessageStartY int
	// LifelineEndPadding is padding below the last message.
	LifelineEndPadding int
}

// DefaultLayoutConfig returns the default layout configuration.
func DefaultLayoutConfig() LayoutConfig {
	return LayoutConfig{
		Padding:              20,
		ParticipantBoxWidth:  100,
		ParticipantBoxHeight: 36,
		ParticipantSpacing:   140,
		MessageSpacing:       45,
		MessageStartY:        70,
		LifelineEndPadding:   30,
	}
}

// CalculateLayout computes positions for all diagram elements.
func CalculateLayout(participantCount, messageCount int, config LayoutConfig) Layout {
	layout := Layout{
		Participants: make([]ParticipantLayout, participantCount),
		Messages:     make([]MessageLayout, messageCount),
	}

	// Calculate participant positions
	for i := 0; i < participantCount; i++ {
		centerX := config.Padding + config.ParticipantBoxWidth/2 + i*config.ParticipantSpacing
		boxX := centerX - config.ParticipantBoxWidth/2

		layout.Participants[i] = ParticipantLayout{
			Index:          i,
			BoxX:           boxX,
			BoxY:           config.Padding,
			BoxWidth:       config.ParticipantBoxWidth,
			BoxHeight:      config.ParticipantBoxHeight,
			CenterX:        centerX,
			LifelineStartY: config.Padding + config.ParticipantBoxHeight,
		}
	}

	// Calculate message Y positions
	for i := 0; i < messageCount; i++ {
		layout.Messages[i] = MessageLayout{
			Step: i + 1,
			Y:    config.MessageStartY + i*config.MessageSpacing,
		}
	}

	// Calculate lifeline end Y (after all messages)
	lifelineEndY := config.MessageStartY + messageCount*config.MessageSpacing + config.LifelineEndPadding
	for i := range layout.Participants {
		layout.Participants[i].LifelineEndY = lifelineEndY
	}

	// Calculate total dimensions
	if participantCount > 0 {
		lastParticipant := layout.Participants[participantCount-1]
		layout.Width = lastParticipant.BoxX + lastParticipant.BoxWidth + config.Padding
	} else {
		layout.Width = config.Padding * 2
	}
	layout.Height = lifelineEndY + config.Padding

	return layout
}

// ParticipantCenterX returns the center X coordinate for a participant by index.
func (l *Layout) ParticipantCenterX(index int) int {
	if index >= 0 && index < len(l.Participants) {
		return l.Participants[index].CenterX
	}
	return 0
}

// SetMessageEndpoints sets the from/to X coordinates for a message.
func (l *Layout) SetMessageEndpoints(msgIndex, fromParticipantIndex, toParticipantIndex int) {
	if msgIndex < 0 || msgIndex >= len(l.Messages) {
		return
	}

	fromX := l.ParticipantCenterX(fromParticipantIndex)
	toX := l.ParticipantCenterX(toParticipantIndex)

	l.Messages[msgIndex].FromX = fromX
	l.Messages[msgIndex].ToX = toX
	l.Messages[msgIndex].IsReverse = toX < fromX

	// Generate SVG path data
	l.Messages[msgIndex].PathD = l.generatePathD(fromX, toX, l.Messages[msgIndex].Y)
}

// generatePathD creates the SVG path data for a message line.
func (l *Layout) generatePathD(fromX, toX, y int) string {
	// Simple straight line path
	return fmt.Sprintf("M%d,%d L%d,%d", fromX, y, toX, y)
}
