package svg_test

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/render/svg"
)

func TestDefaultAnimationConfig(t *testing.T) {
	config := svg.DefaultAnimationConfig()

	if !config.DefaultEnabled {
		t.Error("DefaultEnabled should be true")
	}
	if config.DefaultDuration != "2s" {
		t.Errorf("DefaultDuration = %q, want %q", config.DefaultDuration, "2s")
	}
	if config.DefaultDotSize != 4 {
		t.Errorf("DefaultDotSize = %d, want 4", config.DefaultDotSize)
	}
	if config.StaggerMode != svg.StaggerOverlap {
		t.Errorf("StaggerMode = %v, want StaggerOverlap", config.StaggerMode)
	}
}

func TestResolveFlowAnimation_NilAnimation(t *testing.T) {
	config := svg.DefaultAnimationConfig()
	style := svg.ResolveFlowAnimation(nil, 0, config)

	if !style.Enabled {
		t.Error("Enabled should be true for nil animation")
	}
	if style.Duration != "2s" {
		t.Errorf("Duration = %q, want %q", style.Duration, "2s")
	}
	if style.DotSize != 4 {
		t.Errorf("DotSize = %d, want 4", style.DotSize)
	}
}

func TestResolveFlowAnimation_Presets(t *testing.T) {
	config := svg.DefaultAnimationConfig()

	tests := []struct {
		name      string
		preset    pidl.AnimationPreset
		wantColor string
		wantPulse bool
	}{
		{"success", pidl.AnimationPresetSuccess, "#68d391", false},
		{"error", pidl.AnimationPresetError, "#fc8181", true},
		{"warning", pidl.AnimationPresetWarning, "#f6ad55", true},
		{"highlight", pidl.AnimationPresetHighlight, "#faf089", false},
		{"response", pidl.AnimationPresetResponse, "#a0aec0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anim := &pidl.FlowAnimation{Preset: tt.preset}
			style := svg.ResolveFlowAnimation(anim, 0, config)

			if style.DotColor != tt.wantColor {
				t.Errorf("DotColor = %q, want %q", style.DotColor, tt.wantColor)
			}
			if style.Pulse != tt.wantPulse {
				t.Errorf("Pulse = %v, want %v", style.Pulse, tt.wantPulse)
			}
		})
	}
}

func TestResolveFlowAnimation_NonePreset(t *testing.T) {
	config := svg.DefaultAnimationConfig()
	anim := &pidl.FlowAnimation{Preset: pidl.AnimationPresetNone}
	style := svg.ResolveFlowAnimation(anim, 0, config)

	if style.Enabled {
		t.Error("Enabled should be false for none preset")
	}
}

func TestResolveFlowAnimation_CustomSettings(t *testing.T) {
	config := svg.DefaultAnimationConfig()
	anim := &pidl.FlowAnimation{
		Duration: "3s",
		Delay:    "1s",
		DotColor: "#ff0000",
		DotSize:  8,
		Easing:   "ease-in-out",
	}
	style := svg.ResolveFlowAnimation(anim, 0, config)

	if style.Duration != "3s" {
		t.Errorf("Duration = %q, want %q", style.Duration, "3s")
	}
	if style.Delay != "1s" {
		t.Errorf("Delay = %q, want %q", style.Delay, "1s")
	}
	if style.DotColor != "#ff0000" {
		t.Errorf("DotColor = %q, want %q", style.DotColor, "#ff0000")
	}
	if style.DotSize != 8 {
		t.Errorf("DotSize = %d, want 8", style.DotSize)
	}
	if style.Easing != "ease-in-out" {
		t.Errorf("Easing = %q, want %q", style.Easing, "ease-in-out")
	}
}

func TestResolveFlowAnimation_ExplicitEnabled(t *testing.T) {
	config := svg.DefaultAnimationConfig()

	// Explicitly disabled
	enabled := false
	anim := &pidl.FlowAnimation{Enabled: &enabled}
	style := svg.ResolveFlowAnimation(anim, 0, config)

	if style.Enabled {
		t.Error("Enabled should be false when explicitly set to false")
	}

	// Explicitly enabled
	enabled = true
	anim = &pidl.FlowAnimation{Enabled: &enabled}
	style = svg.ResolveFlowAnimation(anim, 0, config)

	if !style.Enabled {
		t.Error("Enabled should be true when explicitly set to true")
	}
}

func TestGenerateAnimationCSS(t *testing.T) {
	styles := []svg.FlowAnimationStyle{
		{Enabled: true, Duration: "2s", DotColor: "#3182ce", DotSize: 4, Easing: "linear", Delay: "0s"},
		{Enabled: true, Duration: "2s", DotColor: "#68d391", DotSize: 4, Easing: "linear", Delay: "0.3s", Pulse: true},
		{Enabled: false},
	}

	css := svg.GenerateAnimationCSS(styles, nil)

	// Check keyframes
	if !strings.Contains(css, "@keyframes flow") {
		t.Error("CSS should contain flow keyframes")
	}
	if !strings.Contains(css, "@keyframes pulse") {
		t.Error("CSS should contain pulse keyframes")
	}

	// Check enabled flow styles
	if !strings.Contains(css, ".flow-dot-0") {
		t.Error("CSS should contain style for flow-dot-0")
	}
	if !strings.Contains(css, ".flow-dot-1") {
		t.Error("CSS should contain style for flow-dot-1")
	}

	// Check disabled flow is not included
	if strings.Contains(css, ".flow-dot-2") {
		t.Error("CSS should not contain style for disabled flow-dot-2")
	}

	// Check pulse animation is included for flow 1
	if !strings.Contains(css, "pulse 0.5s") {
		t.Error("CSS should contain pulse animation")
	}
}

func TestStaggerModes(t *testing.T) {
	tests := []struct {
		mode  svg.StaggerMode
		index int
		want  string
	}{
		{svg.StaggerParallel, 0, "0s"},
		{svg.StaggerParallel, 5, "0s"},
		{svg.StaggerOverlap, 0, "calc(0 * 0.3s)"},
		{svg.StaggerOverlap, 2, "calc(2 * 0.3s)"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			config := svg.DefaultAnimationConfig()
			config.StaggerMode = tt.mode
			style := svg.ResolveFlowAnimation(nil, tt.index, config)

			if style.Delay != tt.want {
				t.Errorf("Delay = %q, want %q", style.Delay, tt.want)
			}
		})
	}
}
