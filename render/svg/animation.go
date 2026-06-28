package svg

import (
	"fmt"
	"strings"

	"github.com/grokify/pidl"
)

// AnimationConfig contains global animation settings.
type AnimationConfig struct {
	// DefaultEnabled is whether animation is enabled by default.
	DefaultEnabled bool
	// DefaultDuration is the default animation cycle duration.
	DefaultDuration string
	// DefaultDotSize is the default dot radius.
	DefaultDotSize int
	// DefaultDotColor is the default dot color.
	DefaultDotColor string
	// Stagger is the delay between message animations.
	Stagger string
	// StaggerMode controls how animations are staggered.
	StaggerMode StaggerMode
}

// StaggerMode controls how animations are staggered.
type StaggerMode string

const (
	// StaggerSequential starts each flow after the previous ends.
	StaggerSequential StaggerMode = "sequential"
	// StaggerOverlap starts flows with fixed delay, may overlap.
	StaggerOverlap StaggerMode = "overlap"
	// StaggerParallel animates all flows simultaneously.
	StaggerParallel StaggerMode = "parallel"
)

// DefaultAnimationConfig returns the default animation configuration.
func DefaultAnimationConfig() AnimationConfig {
	return AnimationConfig{
		DefaultEnabled:  true,
		DefaultDuration: "2s",
		DefaultDotSize:  4,
		DefaultDotColor: "var(--color-dot)",
		Stagger:         "0.3s",
		StaggerMode:     StaggerOverlap,
	}
}

// FlowAnimationStyle contains resolved animation properties for a flow.
type FlowAnimationStyle struct {
	// Enabled is whether animation is enabled.
	Enabled bool
	// Duration is the animation cycle duration.
	Duration string
	// Delay is the animation start delay.
	Delay string
	// DotColor is the dot fill color.
	DotColor string
	// DotSize is the dot radius.
	DotSize int
	// Pulse adds a pulsing effect.
	Pulse bool
	// Glow adds a glow effect for highlights.
	Glow bool
	// Easing is the CSS easing function.
	Easing string
}

// ResolveFlowAnimation resolves the animation style for a flow.
func ResolveFlowAnimation(flow *pidl.FlowAnimation, index int, config AnimationConfig) FlowAnimationStyle {
	style := FlowAnimationStyle{
		Enabled:  config.DefaultEnabled,
		Duration: config.DefaultDuration,
		DotColor: config.DefaultDotColor,
		DotSize:  config.DefaultDotSize,
		Easing:   "linear",
	}

	// Calculate default stagger delay
	style.Delay = calculateStaggerDelay(index, config)

	if flow == nil {
		return style
	}

	// Apply preset or enabled flag
	if !flow.IsAnimationEnabled() {
		style.Enabled = false
		return style
	}

	// Override with flow-specific settings
	if flow.Duration != "" {
		style.Duration = flow.Duration
	}
	if flow.Delay != "" {
		style.Delay = flow.Delay
	}
	style.DotColor = flow.EffectiveDotColor(config.DefaultDotColor)
	style.DotSize = flow.EffectiveDotSize(config.DefaultDotSize)
	style.Pulse = flow.ShouldPulse()
	style.Glow = flow.ShouldGlow()
	if flow.Easing != "" {
		style.Easing = flow.Easing
	}

	return style
}

// calculateStaggerDelay calculates the stagger delay for a message index.
func calculateStaggerDelay(index int, config AnimationConfig) string {
	switch config.StaggerMode {
	case StaggerParallel:
		return "0s"
	case StaggerSequential:
		// Each message starts after the previous ends
		return fmt.Sprintf("calc(%d * %s)", index, config.DefaultDuration)
	default: // StaggerOverlap
		return fmt.Sprintf("calc(%d * %s)", index, config.Stagger)
	}
}

// GenerateAnimationCSS generates CSS for flow animations.
func GenerateAnimationCSS(flows []FlowAnimationStyle, pathData []string) string {
	var sb strings.Builder

	// Base keyframes
	sb.WriteString(`
    /* Animation keyframes */
    @keyframes flow {
      0% { offset-distance: 0%; }
      100% { offset-distance: 100%; }
    }

    @keyframes pulse {
      0%, 100% { opacity: 1; transform: scale(1); }
      50% { opacity: 0.7; transform: scale(1.3); }
    }

    @keyframes glow {
      0%, 100% { filter: drop-shadow(0 0 2px currentColor); }
      50% { filter: drop-shadow(0 0 8px currentColor) drop-shadow(0 0 12px currentColor); }
    }

    .flow-dot {
      offset-rotate: 0deg;
    }
`)

	// Per-flow animation styles
	for i, style := range flows {
		if !style.Enabled {
			continue
		}

		animations := fmt.Sprintf("flow %s %s infinite", style.Duration, style.Easing)
		if style.Pulse {
			animations += ", pulse 0.5s ease-in-out infinite"
		}
		if style.Glow {
			animations += ", glow 1s ease-in-out infinite"
		}

		extraStyles := ""
		if style.Glow {
			extraStyles = "\n      filter: drop-shadow(0 0 4px currentColor);"
		}

		sb.WriteString(fmt.Sprintf(`
    .flow-dot-%d {
      fill: %s;
      r: %d;
      animation: %s;
      animation-delay: %s;%s
    }
`, i, style.DotColor, style.DotSize, animations, style.Delay, extraStyles))
	}

	return sb.String()
}

// GenerateLegacyAnimationCSS generates backwards-compatible animation CSS.
// This is used when flows don't have per-message animation config.
func GenerateLegacyAnimationCSS(messageCount int) string {
	var sb strings.Builder

	sb.WriteString(`
    /* Animation keyframes */
    @keyframes flow {
      0% { offset-distance: 0%; }
      100% { offset-distance: 100%; }
    }

    .flow-dot {
      fill: var(--color-dot);
      offset-rotate: 0deg;
      animation: flow 2s linear infinite;
    }
`)

	// Staggered delays for each message
	for i := 0; i < messageCount; i++ {
		delay := float64(i) * 0.3
		sb.WriteString(fmt.Sprintf("    .flow-dot-%d { animation-delay: %.1fs; }\n", i, delay))
	}

	return sb.String()
}
