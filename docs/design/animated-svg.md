# Design: SVG Renderer with Animation Support

**Status:** Draft
**Author:** John Wang
**Created:** 2025-05-17
**Last Updated:** 2025-05-17

## Summary

Add SVG rendering capabilities to PIDL, including static sequence diagrams, animated flow visualizations with moving dots, and network boundary diagrams. This extends the existing `render/` package to output self-contained SVG files with optional CSS animations.

## Motivation

Current PIDL renderers output text-based diagram formats (Mermaid, PlantUML, D2, Dot) that require external tools to render. Direct SVG output provides:

1. **Self-contained output** - No external rendering tools needed
2. **Animation support** - CSS-based flow animations (moving dots) for visualizing data flow
3. **Network boundary views** - Architecture diagrams showing trust boundaries and network segments
4. **Embeddable** - Works in HTML documents, datasheets, presentations
5. **Themeable** - CSS custom properties for light/dark themes
6. **Print-friendly** - Static fallback for print media

### Use Cases

- Protocol datasheets with embedded diagrams
- Documentation with animated flow visualizations
- Architecture diagrams showing network boundaries
- LinkedIn/social media posts with animated protocol flows
- Interactive documentation

## Proposed Formats

Extend `render.Format` with three new SVG variants:

| Format | Output | Description |
|--------|--------|-------------|
| `svg` | Static SVG | Sequence diagram similar to Mermaid output |
| `svg-animated` | Animated SVG | Sequence diagram with CSS animations (moving dots) |
| `svg-network` | Network SVG | Architecture diagram with boundaries and components |

```go
const (
    FormatSVG         Format = "svg"
    FormatSVGAnimated Format = "svg-animated"
    FormatSVGNetwork  Format = "svg-network"
)
```

## Architecture

### Package Structure

```
render/
├── render.go              # existing - add new formats
├── mermaid.go             # existing
├── plantuml.go            # existing
├── d2.go                  # existing
├── dot.go                 # existing
├── svg.go                 # NEW: SVGRenderer implementation
├── svg_network.go         # NEW: SVGNetworkRenderer implementation
└── svg/                   # NEW: SVG support package
    ├── layout.go          # participant/message layout calculations
    ├── animation.go       # CSS animation generation
    ├── theme.go           # CSS custom properties / theming
    ├── boundary.go        # network boundary inference/layout
    ├── templates.go       # template loading and registration
    └── templates/         # embedded template resources
        ├── default/
        │   ├── template.svg
        │   ├── styles.css
        │   └── config.json
        ├── sketch/
        │   ├── template.svg
        │   ├── styles.css
        │   └── config.json
        ├── minimal/
        │   └── ...
        ├── blueprint/
        │   └── ...
        └── high-contrast/
            └── ...
```

### Renderer Interface

The new renderers implement the existing `Renderer` interface:

```go
type SVGRenderer struct {
    // Animated enables CSS animations (moving dots along message lines)
    Animated bool

    // Theme selects color scheme ("light", "dark", "auto")
    Theme string

    // Width sets the SVG width (default: auto-calculated)
    Width int

    // ParticipantSpacing sets horizontal spacing between participants
    ParticipantSpacing int

    // MessageSpacing sets vertical spacing between messages
    MessageSpacing int

    // ShowPhases renders phase boxes as background regions
    ShowPhases bool

    // AnimationDuration sets the flow animation duration (default: 2s)
    AnimationDuration time.Duration

    // AnimationStagger staggers animation start times per message
    AnimationStagger bool
}

type SVGNetworkRenderer struct {
    // Animated enables CSS animations
    Animated bool

    // Theme selects color scheme
    Theme string

    // Boundaries defines network boundary regions
    // If empty, auto-detected from entity metadata
    Boundaries []NetworkBoundary

    // Layout algorithm ("horizontal", "vertical", "auto")
    Layout string
}

type NetworkBoundary struct {
    ID       string
    Name     string
    Entities []string  // entity IDs within this boundary
    Style    string    // "trusted", "dmz", "external"
}
```

## SVG Structure

### Sequence Diagram Layout

```
┌─────────────────────────────────────────────────────────────┐
│ SVG ViewBox                                                 │
│                                                             │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐  │
│  │ Human   │    │  IdP    │    │  Agent  │    │ Server  │  │
│  └────┬────┘    └────┬────┘    └────┬────┘    └────┬────┘  │
│       │              │              │              │        │
│       │──────────────>              │              │        │
│       │   1. Auth    │●────────────>│              │  ←dots │
│       │              │              │──────────────>        │
│       │              │              │   2. Request │        │
│       │              │              │<─ ─ ─ ─ ─ ─ ─        │
│       │              │              │   3. Response│        │
│       ▼              ▼              ▼              ▼        │
└─────────────────────────────────────────────────────────────┘
```

### SVG Elements

```svg
<svg viewBox="0 0 800 400" xmlns="http://www.w3.org/2000/svg">
  <style>
    /* CSS custom properties for theming */
    :root { --color-primary: #1a365d; ... }

    /* Animation keyframes */
    @keyframes flow {
      0% { offset-distance: 0%; }
      100% { offset-distance: 100%; }
    }
    .flow-dot { animation: flow 2s linear infinite; }
  </style>

  <defs>
    <!-- Reusable elements: arrows, gradients -->
    <marker id="arrow" ...>
  </defs>

  <!-- Participants -->
  <g class="participants">
    <rect class="participant-box" x="50" y="10" width="100" height="30"/>
    <text class="participant-text">Human</text>
  </g>

  <!-- Lifelines -->
  <g class="lifelines">
    <line class="lifeline" x1="100" y1="40" x2="100" y2="380"/>
  </g>

  <!-- Messages with optional animation -->
  <g class="messages">
    <path id="msg1" class="message-line" d="M100,60 L250,60"/>
    <text class="message-text">1. Authenticate</text>

    <!-- Animated dot (only in svg-animated) -->
    <circle class="flow-dot" r="4" style="offset-path: path('M100,60 L250,60')"/>
  </g>
</svg>
```

### Network Boundary Layout

```
┌──────────────────────────────────────────────────────────────────┐
│ SVG ViewBox                                                      │
│                                                                  │
│  ┌─ ─ ─ ─ ─ ─ ─ ─┐   ┌─ ─ ─ ─ ─ ─ ─ ─┐   ┌─ ─ ─ ─ ─ ─ ─ ─┐    │
│  │ Client Network │   │      DMZ      │   │ Internal Network│    │
│  │                │   │               │   │                │    │
│  │  ┌─────────┐  │   │  ┌─────────┐  │   │  ┌─────────┐  │    │
│  │  │  Agent  │──┼───┼─>│Auth Srv │──┼───┼─>│ Resource│  │    │
│  │  └─────────┘  │   │  └─────────┘  │   │  └─────────┘  │    │
│  │       │       │   │       │       │   │               │    │
│  │       │       │   │       ▼       │   │               │    │
│  │       │       │   │  ┌─────────┐  │   │               │    │
│  │       └───────┼───┼─>│   IdP   │  │   │               │    │
│  │               │   │  └─────────┘  │   │               │    │
│  └─ ─ ─ ─ ─ ─ ─ ─┘   └─ ─ ─ ─ ─ ─ ─ ─┘   └─ ─ ─ ─ ─ ─ ─ ─┘    │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Animation System

### CSS Offset Path Animation

Use CSS `offset-path` to animate dots along message paths:

```css
@keyframes flow {
  0% { offset-distance: 0%; }
  100% { offset-distance: 100%; }
}

.flow-dot {
  offset-rotate: 0deg;
  animation: flow var(--animation-duration, 2s) linear infinite;
}

/* Staggered animation delays */
.flow-dot[data-step="1"] { animation-delay: 0s; }
.flow-dot[data-step="2"] { animation-delay: 0.5s; }
.flow-dot[data-step="3"] { animation-delay: 1s; }
```

### Animation Options

| Option | Description | Default |
|--------|-------------|---------|
| `--animation-duration` | Total animation cycle time | `2s` |
| `--dot-size` | Radius of animated dots | `4px` |
| `--dot-color` | Dot fill color | `var(--color-accent)` |
| `stagger` | Delay between message animations | `0.5s` |
| `direction` | `forward`, `reverse`, `alternate` | `forward` |

### Print Media

Animations are disabled in print:

```css
@media print {
  .flow-dot { display: none; }
  /* Or show as static dots at midpoint */
  .flow-dot {
    animation: none;
    offset-distance: 50%;
  }
}
```

## Theming

### CSS Custom Properties

```css
:root, .theme-light {
  --color-primary: #1a365d;
  --color-secondary: #2c5282;
  --color-accent: #3182ce;
  --color-bg: #ffffff;
  --color-text: #1a202c;
  --color-line: #cbd5e0;
  --color-line-response: #a0aec0;
  --color-boundary-trusted: rgba(39, 103, 73, 0.1);
  --color-boundary-dmz: rgba(49, 130, 206, 0.1);
  --color-boundary-external: rgba(197, 48, 48, 0.1);
}

.theme-dark {
  --color-primary: #63b3ed;
  --color-bg: #1a202c;
  --color-text: #e2e8f0;
  /* ... */
}
```

### Theme Detection

```css
@media (prefers-color-scheme: dark) {
  .theme-auto { /* dark theme values */ }
}
```

## Network Boundaries

Network boundaries group entities into logical regions (trust zones, network segments, deployment environments). The renderer supports both automatic inference and explicit configuration.

### Inference + Override Model

Boundaries follow a layered approach:

1. **Inference (default)**: If entities have `metadata.network`, group them automatically
2. **Override**: Explicit `metadata.networks` section takes precedence
3. **CLI override**: `--boundary` flags override everything

This allows sensible defaults while giving humans/LLMs full control when needed.

### PIDL Schema Extensions

```json
{
  "entities": [
    {
      "id": "agent",
      "name": "AI Agent",
      "metadata": {
        "network": "client",
        "type": "service"
      }
    },
    {
      "id": "auth_server",
      "name": "Authorization Server",
      "metadata": {
        "network": "dmz",
        "type": "service"
      }
    },
    {
      "id": "idp",
      "name": "Identity Provider",
      "metadata": {
        "network": "dmz",
        "type": "service"
      }
    },
    {
      "id": "resource",
      "name": "Resource API",
      "metadata": {
        "network": "internal",
        "type": "service"
      }
    }
  ],
  "metadata": {
    "networks": {
      "client": {
        "name": "Client Network",
        "style": "external",
        "description": "Untrusted client environment"
      },
      "dmz": {
        "name": "DMZ",
        "style": "dmz",
        "entities": ["auth_server", "idp"],
        "description": "Perimeter network with public-facing services"
      },
      "internal": {
        "name": "Internal Network",
        "style": "trusted",
        "entities": ["resource", "database"],
        "description": "Protected internal services"
      }
    },
    "network_layout": {
      "direction": "horizontal",
      "order": ["client", "dmz", "internal"]
    }
  }
}
```

### Network Configuration Options

| Field | Description | Default |
|-------|-------------|---------|
| `name` | Display name for the boundary | ID capitalized |
| `style` | Visual style: `trusted`, `dmz`, `external`, `cloud` | `external` |
| `entities` | Explicit entity list (overrides inference) | Inferred from entity metadata |
| `description` | Tooltip/hover text | None |
| `color` | Custom boundary color | Style default |

### Boundary Styles

| Style | Default Color | Border | Use Case |
|-------|---------------|--------|----------|
| `trusted` | Green tint | Solid | Internal/protected networks |
| `dmz` | Blue tint | Dashed | Perimeter/semi-trusted |
| `external` | Red tint | Dotted | Untrusted/public networks |
| `cloud` | Gray tint | Cloud shape | Cloud provider boundaries |

### Resolution Order

```
1. CLI flags (--boundary="dmz:auth,idp")
     ↓ overrides
2. metadata.networks[id].entities
     ↓ overrides
3. entity.metadata.network (inference)
     ↓ fallback
4. "default" boundary (ungrouped entities)
```

## Custom SVG Templates

Templates provide different visual styles for different contexts. Each template defines SVG structure, CSS styling, and layout parameters.

### Built-in Templates

| Template | Description | Use Case |
|----------|-------------|----------|
| `default` | Clean, professional styling | Documentation, datasheets |
| `sketch` | Hand-drawn, informal look | Whiteboards, brainstorming |
| `blueprint` | Technical with grid background | Engineering specs |
| `minimal` | Ultra-clean, reduced chrome | Presentations, slides |
| `corporate` | Customizable brand colors | Company documentation |
| `high-contrast` | WCAG AA compliant | Accessibility |
| `dark` | Dark background optimized | Dark mode UIs |

### Template Structure

Templates are Go `embed.FS` resources:

```
render/svg/templates/
├── default/
│   ├── template.svg      # SVG structure template
│   ├── styles.css        # CSS styling
│   └── config.json       # Layout parameters
├── sketch/
│   ├── template.svg
│   ├── styles.css
│   ├── config.json
│   └── fonts/            # Custom fonts (optional)
│       └── handwritten.woff2
├── blueprint/
│   └── ...
└── minimal/
    └── ...
```

### Template Configuration

`config.json` defines layout and rendering parameters:

```json
{
  "name": "sketch",
  "description": "Hand-drawn informal style",
  "layout": {
    "participant_spacing": 150,
    "message_spacing": 50,
    "padding": 20,
    "participant_box": {
      "width": 120,
      "height": 40,
      "corner_radius": 8
    }
  },
  "fonts": {
    "primary": "Architects Daughter, cursive",
    "mono": "Comic Mono, monospace"
  },
  "line_style": {
    "stroke_width": 2,
    "hand_drawn": true,
    "wobble_factor": 0.3
  },
  "animation": {
    "dot_shape": "circle",
    "dot_size": 6,
    "trail_enabled": true
  }
}
```

### Using Templates

**CLI:**
```bash
pidl render --format=svg --template=sketch protocol.json
pidl render --format=svg-animated --template=minimal protocol.json
```

**Go API:**
```go
renderer := render.NewSVG()
renderer.Template = "sketch"
renderer.RenderString(protocol)
```

### Custom Templates

Users can provide custom templates:

```bash
# From directory
pidl render --format=svg --template-dir=./my-template protocol.json

# Register in config
pidl config set templates.corporate=/path/to/corporate-template
pidl render --format=svg --template=corporate protocol.json
```

### Template Variables

Templates can use variables for customization:

```svg
<svg xmlns="http://www.w3.org/2000/svg">
  <style>
    :root {
      --brand-primary: {{.BrandPrimary | default "#1a365d"}};
      --brand-secondary: {{.BrandSecondary | default "#2c5282"}};
      --font-family: {{.FontFamily | default "Inter, sans-serif"}};
    }
  </style>
  <!-- ... -->
</svg>
```

```bash
pidl render --format=svg --template=corporate \
  --set brand-primary="#ff6600" \
  --set brand-secondary="#333333" \
  protocol.json
```

## Per-Message Animation Control

Animation can be configured globally and overridden per-flow (message). This enables highlighting critical paths, color-coding by outcome, or keeping setup flows static.

### Flow Animation Schema

Extend the PIDL flow schema with optional `animation` field:

```json
{
  "flows": [
    {
      "from": "agent",
      "to": "idp",
      "label": "1. Request ID-JAG",
      "animation": {
        "enabled": true,
        "duration": "1.5s",
        "dot_color": "#3182ce",
        "dot_size": 5
      }
    },
    {
      "from": "idp",
      "to": "agent",
      "label": "2. Return ID-JAG",
      "mode": "response",
      "animation": {
        "enabled": true,
        "dot_color": "#68d391",
        "delay": "0.5s"
      }
    },
    {
      "from": "agent",
      "to": "resource",
      "label": "3. Access API",
      "animation": {
        "enabled": true,
        "dot_color": "#3182ce"
      }
    },
    {
      "from": "resource",
      "to": "agent",
      "label": "4. 403 Forbidden",
      "mode": "response",
      "animation": {
        "enabled": true,
        "dot_color": "#fc8181",
        "pulse": true
      }
    },
    {
      "from": "human",
      "to": "agent",
      "label": "Setup: Delegate task",
      "animation": {
        "enabled": false
      }
    }
  ],
  "metadata": {
    "animation": {
      "default_enabled": true,
      "default_duration": "2s",
      "default_dot_color": "#3182ce",
      "stagger": "0.3s",
      "loop": true
    }
  }
}
```

### Animation Properties

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `enabled` | boolean | `true` | Enable/disable animation for this flow |
| `duration` | string | `"2s"` | Animation cycle duration |
| `delay` | string | `"0s"` | Start delay (for sequencing) |
| `dot_color` | string | theme accent | Dot fill color |
| `dot_size` | number | `4` | Dot radius in pixels |
| `dot_shape` | string | `"circle"` | Shape: `circle`, `square`, `diamond`, `arrow` |
| `pulse` | boolean | `false` | Add pulsing effect (for errors/warnings) |
| `trail` | boolean | `false` | Show fading trail behind dot |
| `direction` | string | `"forward"` | `forward`, `reverse`, `alternate` |
| `easing` | string | `"linear"` | CSS easing: `linear`, `ease-in-out`, etc. |

### Semantic Presets

For convenience, use semantic presets:

```json
{
  "flows": [
    {
      "from": "agent", "to": "resource",
      "label": "API Call",
      "animation": "request"
    },
    {
      "from": "resource", "to": "agent",
      "label": "Success",
      "animation": "success"
    },
    {
      "from": "resource", "to": "agent",
      "label": "Error",
      "animation": "error"
    },
    {
      "from": "human", "to": "agent",
      "label": "Background",
      "animation": "none"
    }
  ]
}
```

Preset definitions:

| Preset | Color | Effect | Use Case |
|--------|-------|--------|----------|
| `request` | Blue | Standard | Outgoing requests |
| `response` | Gray | Dashed path | Return values |
| `success` | Green | Standard | Successful operations |
| `error` | Red | Pulse | Errors, failures |
| `warning` | Orange | Pulse | Warnings |
| `highlight` | Yellow | Glow + larger dot | Critical path |
| `background` | Gray, 50% opacity | Subtle | Setup/teardown |
| `none` | N/A | Static arrow | No animation |

### Global Animation Settings

Set defaults in protocol metadata:

```json
{
  "metadata": {
    "animation": {
      "default_enabled": true,
      "default_duration": "2s",
      "stagger": "0.3s",
      "stagger_mode": "sequential",
      "loop": true,
      "pause_on_hover": true,
      "color_scheme": {
        "request": "#3182ce",
        "response": "#a0aec0",
        "success": "#68d391",
        "error": "#fc8181",
        "warning": "#f6ad55"
      }
    }
  }
}
```

### Stagger Modes

| Mode | Description |
|------|-------------|
| `sequential` | Each flow starts after previous ends |
| `overlap` | Flows start with fixed delay, may overlap |
| `parallel` | All flows animate simultaneously |
| `phase` | Flows within same phase animate together |

### Generated CSS Example

For a flow with custom animation:

```css
/* Flow: agent → idp (step 1) */
#flow-1-dot {
  fill: #3182ce;
  r: 5;
  offset-path: path('M150,80 L300,80');
  offset-rotate: 0deg;
  animation: flow-1 1.5s linear infinite;
  animation-delay: 0s;
}

@keyframes flow-1 {
  0% { offset-distance: 0%; }
  100% { offset-distance: 100%; }
}

/* Flow: resource → agent (step 4, error) */
#flow-4-dot {
  fill: #fc8181;
  r: 4;
  offset-path: path('M550,140 L150,140');
  animation: flow-4 2s linear infinite, pulse 0.5s ease-in-out infinite;
  animation-delay: 0.9s;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(1.3); }
}
```

## CLI Integration

```bash
# Static SVG sequence diagram
pidl render --format=svg protocol.json -o diagram.svg

# Animated SVG with flow dots
pidl render --format=svg-animated protocol.json -o animated.svg

# Network boundary diagram
pidl render --format=svg-network protocol.json -o network.svg

# With template
pidl render --format=svg --template=sketch protocol.json
pidl render --format=svg-animated --template=minimal protocol.json

# Custom template from directory
pidl render --format=svg --template-dir=./my-template protocol.json

# Template with brand customization
pidl render --format=svg --template=corporate \
  --set brand-primary="#ff6600" \
  --set font-family="Helvetica" \
  protocol.json

# Animation options
pidl render --format=svg-animated \
  --theme=dark \
  --animation-duration=3s \
  --stagger=0.5s \
  --stagger-mode=sequential \
  protocol.json

# Network boundary overrides
pidl render --format=svg-network \
  --boundary="dmz:auth_server,idp" \
  --boundary="internal:resource,database" \
  --layout=horizontal \
  protocol.json

# Combined options
pidl render --format=svg-animated \
  --template=minimal \
  --theme=auto \
  --animation-duration=2s \
  --stagger \
  protocol.json -o flow.svg
```

## Implementation Phases

### Phase 1: Static SVG Renderer ✅
- [x] Basic `SVGRenderer` struct implementing `Renderer` interface
- [x] Participant layout calculation
- [x] Lifeline rendering
- [x] Message arrows (solid for requests, dashed for responses)
- [x] Message labels with step numbers
- [x] Light/dark theme support via CSS custom properties
- [ ] Default template (`default/`)
- [x] Unit tests and golden file tests

### Phase 2: CSS Animation Support ✅
- [x] `offset-path` animation for flow dots
- [x] Global animation configuration (`metadata.animation`)
- [x] Per-message animation properties (`flow.animation`)
- [x] Semantic presets (`request`, `success`, `error`, `none`)
- [x] Stagger modes (`sequential`, `overlap`, `parallel`)
- [x] Print media fallback (static or hidden dots)
- [ ] Browser compatibility testing (Chrome, Firefox, Safari)

### Phase 3: Template System
- [ ] Template loading from embedded FS
- [ ] Template configuration parsing (`config.json`)
- [ ] Built-in templates: `sketch`, `minimal`, `blueprint`
- [ ] Custom template loading from directory
- [ ] Template variable substitution
- [ ] CLI `--template` and `--template-dir` flags
- [ ] Template validation

### Phase 4: Network Boundary Diagrams
- [ ] `SVGNetworkRenderer` implementation
- [ ] Boundary inference from `entity.metadata.network`
- [ ] Explicit boundary configuration (`metadata.networks`)
- [ ] Boundary region layout algorithm
- [ ] Entity placement within boundaries
- [ ] Connection routing between/across boundaries
- [ ] Boundary styles (`trusted`, `dmz`, `external`, `cloud`)
- [ ] CLI `--boundary` override flags

### Phase 5: Advanced Features
- [ ] Phase boxes/regions in sequence diagrams
- [ ] Alternative paths visualization
- [ ] High-contrast template for accessibility
- [ ] Interactive hover states (optional, CSS-only)
- [ ] Animation pulse/glow effects for errors
- [ ] Export to GIF/WebP documentation (via Puppeteer/ffmpeg)

## Testing Strategy

1. **Unit tests** - Layout calculations, SVG element generation
2. **Golden file tests** - Compare generated SVG against expected output
3. **Visual regression** - Screenshot comparison (optional, CI integration)
4. **Browser testing** - Verify animations in Chrome, Firefox, Safari

## Alternatives Considered

### 1. SMIL Animations (rejected)
- Deprecated in Chrome, inconsistent browser support
- CSS animations have better support and performance

### 2. JavaScript Animations (rejected)
- Requires runtime, not self-contained
- CSS `offset-path` provides sufficient capability

### 3. External Tools (partial)
- Keep as option: export SVG, use external tool for GIF conversion
- Puppeteer/Playwright can capture animated SVG as video/GIF

## Resolved Questions

1. **Should `svg-network` infer boundaries from entity metadata, or require explicit configuration?**

   **Resolution:** Both. Use inference by default from `entity.metadata.network`, but allow explicit override via `metadata.networks[id].entities` or CLI `--boundary` flags. See [Network Boundaries](#network-boundaries) section.

2. **Should we support custom SVG templates for different visual styles?**

   **Resolution:** Yes. Support built-in templates (`default`, `sketch`, `minimal`, etc.) and custom templates from directories. Templates include SVG structure, CSS styling, and layout config. See [Custom SVG Templates](#custom-svg-templates) section.

3. **Should animation be configurable per-message (e.g., some flows animate, others don't)?**

   **Resolution:** Yes. Each flow can have an `animation` field with properties like `enabled`, `dot_color`, `duration`, `pulse`. Also support semantic presets (`request`, `success`, `error`, `none`). See [Per-Message Animation Control](#per-message-animation-control) section.

## Open Questions

1. Should templates support JavaScript for advanced interactivity, or keep SVG/CSS only for maximum portability?
2. Should we provide a template authoring guide or template validation tool?
3. Should animation presets be customizable in user config, or only in protocol metadata?
4. Should we support exporting animation frames for GIF generation, or rely on external tools?

## References

- [CSS Motion Path](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_motion_path)
- [SVG Animation Best Practices](https://css-tricks.com/guide-svg-animations-smil/)
- [offset-path Browser Support](https://caniuse.com/css-motion-paths)
- [PIDL Specification](../SPECIFICATION.md)
- [Existing Renderers](../../render/)
