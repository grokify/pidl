package pidl

// IsAnimationEnabled returns whether animation is enabled for this flow.
func (f *FlowAnimation) IsAnimationEnabled() bool {
	if f == nil {
		return true // default enabled
	}
	if f.Preset == AnimationPresetNone {
		return false
	}
	if f.Enabled != nil {
		return *f.Enabled
	}
	return true
}

// EffectiveDotColor returns the dot color, applying preset defaults.
func (f *FlowAnimation) EffectiveDotColor(defaultColor string) string {
	if f == nil || f.DotColor == "" {
		if f != nil {
			switch f.Preset {
			case AnimationPresetSuccess:
				return "#68d391"
			case AnimationPresetError:
				return "#fc8181"
			case AnimationPresetWarning:
				return "#f6ad55"
			case AnimationPresetHighlight:
				return "#faf089"
			case AnimationPresetResponse:
				return "#a0aec0"
			}
		}
		return defaultColor
	}
	return f.DotColor
}

// EffectiveDotSize returns the dot size, applying preset defaults.
func (f *FlowAnimation) EffectiveDotSize(defaultSize int) int {
	if f == nil || f.DotSize == 0 {
		if f != nil && f.Preset == AnimationPresetHighlight {
			return 6
		}
		return defaultSize
	}
	return f.DotSize
}

// ShouldPulse returns whether the dot should pulse.
func (f *FlowAnimation) ShouldPulse() bool {
	if f == nil {
		return false
	}
	if f.Pulse {
		return true
	}
	return f.Preset == AnimationPresetError || f.Preset == AnimationPresetWarning
}

// ShouldGlow returns whether the dot should have a glow effect.
func (f *FlowAnimation) ShouldGlow() bool {
	if f == nil {
		return false
	}
	return f.Preset == AnimationPresetHighlight
}

// DisplayLabel returns the label for display, falling back to Action if Label is empty.
func (f Flow) DisplayLabel() string {
	if f.Label != "" {
		return f.Label
	}
	return f.Action
}

// EffectiveMode returns the flow mode, defaulting to FlowModeRequest if empty.
func (f Flow) EffectiveMode() FlowMode {
	if f.Mode == "" {
		return FlowModeRequest
	}
	return f.Mode
}

// HasCondition returns true if the flow has a condition.
func (f Flow) HasCondition() bool {
	return f.Condition != ""
}

// HasAlternatives returns true if the flow has alternative paths.
func (f Flow) HasAlternatives() bool {
	return len(f.Alternatives) > 0
}

// HasAnnotations returns true if the flow has annotations.
func (f Flow) HasAnnotations() bool {
	return len(f.Annotations) > 0
}

// HasNote returns true if the flow has a note.
func (f Flow) HasNote() bool {
	return f.Note != ""
}

// HasStateMutations returns true if the flow has any state mutations.
func (f Flow) HasStateMutations() bool {
	return len(f.Sets) > 0
}

// HasSecurity returns true if the flow has security requirements.
func (f Flow) HasSecurity() bool {
	return f.Security != nil && (len(f.Security.Requires) > 0 || f.Security.Token != "" || f.Security.Confidential)
}

// RequiresEncryption returns true if the flow requires encryption.
func (f Flow) RequiresEncryption() bool {
	if f.Security == nil {
		return false
	}
	for _, req := range f.Security.Requires {
		if req == SecurityRequirementEncryption {
			return true
		}
	}
	return false
}

// RequiresToken returns true if the flow requires a token.
func (f Flow) RequiresToken() bool {
	if f.Security == nil {
		return false
	}
	if f.Security.Token != "" {
		return true
	}
	for _, req := range f.Security.Requires {
		if req == SecurityRequirementToken {
			return true
		}
	}
	return false
}
