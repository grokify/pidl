# MkDocs PIDL Plugin

A MkDocs plugin for rendering PIDL (Protocol Interaction Description Language) diagrams in your documentation.

## Installation

```bash
pip install mkdocs-pidl
```

Or install from source:

```bash
cd integrations/mkdocs-pidl
pip install -e .
```

## Requirements

- Python 3.8+
- MkDocs 1.0+
- PIDL CLI (`pidl` command must be available in PATH)

## Configuration

Add the plugin to your `mkdocs.yml`:

```yaml
plugins:
  - search
  - pidl
```

### Options

```yaml
plugins:
  - pidl:
      pidl_path: pidl           # Path to pidl CLI (default: "pidl")
      default_format: mermaid   # Default output: mermaid, svg, infographic
      default_theme: bold       # Default infographic theme
      default_direction: horizontal  # Layout direction for infographics
      cache_enabled: true       # Cache rendered diagrams
```

## Usage

### Inline PIDL

Write PIDL JSON directly in your markdown:

~~~markdown
```pidl
{
    "id": "oauth2-auth-code",
    "name": "OAuth 2.0 Authorization Code",
    "entities": [
        {"id": "user", "name": "User", "type": "user"},
        {"id": "client", "name": "Client", "type": "client"},
        {"id": "auth_server", "name": "Auth Server", "type": "authorization_server"}
    ],
    "flows": [
        {"from": "user", "to": "client", "action": "initiate"},
        {"from": "client", "to": "auth_server", "action": "authorize"}
    ]
}
```
~~~

### With Options

Specify format, theme, and direction:

~~~markdown
```pidl format=infographic theme=dark direction=vertical
{
    "id": "my-protocol",
    ...
}
```
~~~

### External Files

Reference external PIDL files:

~~~markdown
```pidl file=protocols/oauth2.pidl.json format=mermaid
```
~~~

The file path is relative to your `docs/` directory.

## Output Formats

### Mermaid (default)

Renders as a Mermaid sequence diagram. Requires the MkDocs Material theme or Mermaid extension.

### SVG

Renders as an inline SVG image.

### Infographic

Renders as a stylized infographic with animated flow dots.

Available themes:

- `bold` - High contrast, saturated colors
- `minimal` - Clean, subtle design
- `dark` - Dark background
- `tech` - Tech/engineering aesthetic
- `corporate` - Professional blues and grays
- `accessible` - Colorblind-friendly (CVD safe)

## Development

```bash
# Install in development mode
pip install -e ".[dev]"

# Run tests
pytest

# Type checking
mypy mkdocs_pidl
```

## License

MIT License - see [LICENSE](../../LICENSE)
