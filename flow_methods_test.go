package pidl

import "testing"

func boolPtr(b bool) *bool { return &b }

func TestFlowAnimation_IsAnimationEnabled(t *testing.T) {
	tests := []struct {
		name string
		anim *FlowAnimation
		want bool
	}{
		{name: "nil receiver defaults to enabled", anim: nil, want: true},
		{name: "empty config defaults to enabled", anim: &FlowAnimation{}, want: true},
		{name: "preset none disables", anim: &FlowAnimation{Preset: AnimationPresetNone}, want: false},
		{name: "preset none wins over explicit enabled", anim: &FlowAnimation{Preset: AnimationPresetNone, Enabled: boolPtr(true)}, want: false},
		{name: "explicit enabled true", anim: &FlowAnimation{Enabled: boolPtr(true)}, want: true},
		{name: "explicit enabled false", anim: &FlowAnimation{Enabled: boolPtr(false)}, want: false},
		{name: "explicit enabled false with preset", anim: &FlowAnimation{Preset: AnimationPresetSuccess, Enabled: boolPtr(false)}, want: false},
		{name: "preset without enabled", anim: &FlowAnimation{Preset: AnimationPresetError}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.anim.IsAnimationEnabled(); got != tt.want {
				t.Errorf("IsAnimationEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlowAnimation_EffectiveDotColor(t *testing.T) {
	const fallback = "#000000"

	tests := []struct {
		name string
		anim *FlowAnimation
		want string
	}{
		{name: "nil receiver uses default", anim: nil, want: fallback},
		{name: "empty config uses default", anim: &FlowAnimation{}, want: fallback},
		{name: "request preset uses default", anim: &FlowAnimation{Preset: AnimationPresetRequest}, want: fallback},
		{name: "none preset uses default", anim: &FlowAnimation{Preset: AnimationPresetNone}, want: fallback},
		{name: "success preset", anim: &FlowAnimation{Preset: AnimationPresetSuccess}, want: "#68d391"},
		{name: "error preset", anim: &FlowAnimation{Preset: AnimationPresetError}, want: "#fc8181"},
		{name: "warning preset", anim: &FlowAnimation{Preset: AnimationPresetWarning}, want: "#f6ad55"},
		{name: "highlight preset", anim: &FlowAnimation{Preset: AnimationPresetHighlight}, want: "#faf089"},
		{name: "response preset", anim: &FlowAnimation{Preset: AnimationPresetResponse}, want: "#a0aec0"},
		{name: "explicit color wins over preset", anim: &FlowAnimation{Preset: AnimationPresetError, DotColor: "#123456"}, want: "#123456"},
		{name: "explicit color without preset", anim: &FlowAnimation{DotColor: "#abcdef"}, want: "#abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.anim.EffectiveDotColor(fallback); got != tt.want {
				t.Errorf("EffectiveDotColor(%q) = %q, want %q", fallback, got, tt.want)
			}
		})
	}
}

func TestFlowAnimation_EffectiveDotSize(t *testing.T) {
	const fallback = 4

	tests := []struct {
		name string
		anim *FlowAnimation
		want int
	}{
		{name: "nil receiver uses default", anim: nil, want: fallback},
		{name: "empty config uses default", anim: &FlowAnimation{}, want: fallback},
		{name: "non-highlight preset uses default", anim: &FlowAnimation{Preset: AnimationPresetError}, want: fallback},
		{name: "highlight preset enlarges", anim: &FlowAnimation{Preset: AnimationPresetHighlight}, want: 6},
		{name: "explicit size wins over highlight", anim: &FlowAnimation{Preset: AnimationPresetHighlight, DotSize: 9}, want: 9},
		{name: "explicit size without preset", anim: &FlowAnimation{DotSize: 2}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.anim.EffectiveDotSize(fallback); got != tt.want {
				t.Errorf("EffectiveDotSize(%d) = %d, want %d", fallback, got, tt.want)
			}
		})
	}
}

func TestFlowAnimation_PulseAndGlow(t *testing.T) {
	tests := []struct {
		name      string
		anim      *FlowAnimation
		wantPulse bool
		wantGlow  bool
	}{
		{name: "nil receiver", anim: nil},
		{name: "empty config", anim: &FlowAnimation{}},
		{name: "explicit pulse", anim: &FlowAnimation{Pulse: true}, wantPulse: true},
		{name: "explicit pulse with success preset", anim: &FlowAnimation{Preset: AnimationPresetSuccess, Pulse: true}, wantPulse: true},
		{name: "error preset pulses", anim: &FlowAnimation{Preset: AnimationPresetError}, wantPulse: true},
		{name: "warning preset pulses", anim: &FlowAnimation{Preset: AnimationPresetWarning}, wantPulse: true},
		{name: "highlight preset glows", anim: &FlowAnimation{Preset: AnimationPresetHighlight}, wantGlow: true},
		{name: "request preset neither", anim: &FlowAnimation{Preset: AnimationPresetRequest}},
		{name: "response preset neither", anim: &FlowAnimation{Preset: AnimationPresetResponse}},
		{name: "none preset neither", anim: &FlowAnimation{Preset: AnimationPresetNone}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.anim.ShouldPulse(); got != tt.wantPulse {
				t.Errorf("ShouldPulse() = %v, want %v", got, tt.wantPulse)
			}
			if got := tt.anim.ShouldGlow(); got != tt.wantGlow {
				t.Errorf("ShouldGlow() = %v, want %v", got, tt.wantGlow)
			}
		})
	}
}
