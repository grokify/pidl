package svg

import "strings"

// Theme represents a color theme for SVG diagrams.
type Theme string

const (
	// ThemeLight is the light color theme.
	ThemeLight Theme = "light"
	// ThemeDark is the dark color theme.
	ThemeDark Theme = "dark"
	// ThemeAuto uses system preference via prefers-color-scheme.
	ThemeAuto Theme = "auto"
)

// GenerateCSS generates CSS styles for the SVG diagram.
func GenerateCSS(theme Theme) string {
	var sb strings.Builder

	// CSS custom properties for theming
	sb.WriteString(`
    /* Light theme (default) */
    :root, .theme-light {
      --color-primary: #1a365d;
      --color-secondary: #2c5282;
      --color-accent: #3182ce;
      --color-bg: #ffffff;
      --color-text: #1a202c;
      --color-text-muted: #718096;
      --color-line: #3182ce;
      --color-line-response: #a0aec0;
      --color-lifeline: #cbd5e0;
      --color-participant-bg: #1a365d;
      --color-participant-text: #ffffff;
      --color-dot: #3182ce;
    }

    /* Dark theme */
    .theme-dark {
      --color-primary: #63b3ed;
      --color-secondary: #90cdf4;
      --color-accent: #4299e1;
      --color-bg: #1a202c;
      --color-text: #e2e8f0;
      --color-text-muted: #a0aec0;
      --color-line: #63b3ed;
      --color-line-response: #4a5568;
      --color-lifeline: #4a5568;
      --color-participant-bg: #4299e1;
      --color-participant-text: #ffffff;
      --color-dot: #63b3ed;
    }
`)

	// Auto theme using prefers-color-scheme
	if theme == ThemeAuto {
		sb.WriteString(`
    @media (prefers-color-scheme: dark) {
      :root {
        --color-primary: #63b3ed;
        --color-secondary: #90cdf4;
        --color-accent: #4299e1;
        --color-bg: #1a202c;
        --color-text: #e2e8f0;
        --color-text-muted: #a0aec0;
        --color-line: #63b3ed;
        --color-line-response: #4a5568;
        --color-lifeline: #4a5568;
        --color-participant-bg: #4299e1;
        --color-participant-text: #ffffff;
        --color-dot: #63b3ed;
      }
    }
`)
	}

	// Element styles
	sb.WriteString(`
    /* Element styles */
    .participant-box {
      fill: var(--color-participant-bg);
    }

    .participant-text {
      fill: var(--color-participant-text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 12px;
      font-weight: 600;
      text-anchor: middle;
      dominant-baseline: central;
    }

    .lifeline {
      stroke: var(--color-lifeline);
      stroke-width: 1;
      stroke-dasharray: 4,4;
    }

    .message-line {
      fill: none;
      stroke: var(--color-line);
      stroke-width: 1.5;
    }

    .message-line-response {
      fill: none;
      stroke: var(--color-line-response);
      stroke-width: 1.5;
      stroke-dasharray: 6,3;
    }

    .message-arrow {
      fill: var(--color-line);
    }

    .message-arrow-response {
      fill: var(--color-line-response);
    }

    .message-text {
      fill: var(--color-text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 11px;
    }

    .step-circle {
      fill: var(--color-accent);
    }

    .step-number {
      fill: white;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 9px;
      font-weight: bold;
      text-anchor: middle;
      dominant-baseline: central;
    }

    /* Phase box styles */
    .phase-box {
      stroke: var(--color-lifeline);
      stroke-width: 1;
      stroke-dasharray: 4, 2;
    }

    .phase-label {
      fill: var(--color-text-muted);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 10px;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    /* Alternative badge styles */
    .alt-badge-bg {
      fill: var(--color-accent);
      opacity: 0.9;
    }

    .alt-badge-text {
      fill: white;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 8px;
      font-weight: 600;
      text-anchor: middle;
      dominant-baseline: central;
    }

    .alt-badge {
      cursor: pointer;
      transition: transform 0.2s ease;
    }

    .alt-badge:hover {
      transform: scale(1.1);
    }

    /* Note styles */
    .message-note {
      fill: var(--color-text-muted);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      font-size: 9px;
      font-style: italic;
    }
`)

	// Interactive hover states
	sb.WriteString(`
    /* Interactive hover states */
    .participant-box {
      transition: filter 0.2s ease, transform 0.2s ease;
      cursor: pointer;
    }

    .participant-box:hover {
      filter: brightness(1.15);
      transform: translateY(-1px);
    }

    .message-line, .message-line-response {
      transition: stroke-width 0.2s ease, filter 0.2s ease;
      cursor: pointer;
    }

    .message-line:hover, .message-line-response:hover {
      stroke-width: 2.5;
      filter: drop-shadow(0 0 2px var(--color-line));
    }

    .message-text {
      transition: font-weight 0.2s ease;
      cursor: pointer;
    }

    .message-text:hover {
      font-weight: 600;
    }

    .flow-dot {
      transition: r 0.15s ease;
    }

    .flow-dot:hover {
      r: 6;
    }

    .step-circle {
      transition: transform 0.2s ease;
      cursor: pointer;
    }

    .step-circle:hover {
      transform: scale(1.2);
    }
`)

	// Print styles
	sb.WriteString(`
    /* Print styles */
    @media print {
      .flow-dot {
        display: none;
      }
      /* Disable hover effects in print */
      .participant-box:hover,
      .message-line:hover,
      .message-line-response:hover {
        filter: none;
        transform: none;
      }
    }
`)

	return sb.String()
}

// ThemeClass returns the CSS class for the given theme.
func ThemeClass(theme Theme) string {
	switch theme {
	case ThemeDark:
		return "theme-dark"
	case ThemeLight:
		return "theme-light"
	default:
		return ""
	}
}
