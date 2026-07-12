# PIDL for VS Code

Visual Studio Code extension for PIDL (Protocol Interaction Description Language) files.

## Features

- **Syntax Highlighting**: Full syntax highlighting for `.pidl.json` files
- **Validation**: Real-time validation with error highlighting
- **Diagram Preview**: Live Mermaid diagram preview
- **Export**: Export to Mermaid or SVG formats
- **Security Analysis**: Run security analysis on protocols

## Installation

### From VSIX

1. Download the latest `.vsix` file from releases
2. Open VS Code
3. Run `Extensions: Install from VSIX...` command
4. Select the downloaded file

### From Source

```bash
cd integrations/vscode-pidl
npm install
npm run compile
npm run package
```

## Requirements

- VS Code 1.74.0 or later
- PIDL CLI installed and in PATH (`go install github.com/grokify/pidl/cmd/pidl@latest`)

## Usage

### Creating PIDL Files

Create a new file with `.pidl.json` extension:

```json
{
    "id": "my-protocol",
    "name": "My Protocol",
    "entities": [
        {"id": "user", "name": "User", "type": "user"},
        {"id": "server", "name": "Server", "type": "server"}
    ],
    "flows": [
        {"from": "user", "to": "server", "action": "request"}
    ]
}
```

### Commands

Access commands via Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`):

| Command | Description |
|---------|-------------|
| `PIDL: Validate Current File` | Validate the current PIDL file |
| `PIDL: Preview Diagram` | Open diagram preview in side panel |
| `PIDL: Export as Mermaid` | Export as Mermaid diagram |
| `PIDL: Export as SVG` | Export as SVG image |
| `PIDL: Run Security Analysis` | Run security analysis |

### Context Menu

Right-click in a PIDL file to access:

- Preview Diagram
- Validate

### Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `pidl.cliPath` | `pidl` | Path to PIDL CLI |
| `pidl.validateOnSave` | `true` | Validate on save |
| `pidl.showPreviewOnOpen` | `false` | Show preview when opening PIDL files |

## Syntax Highlighting

The extension provides highlighting for:

- **Keywords**: `id`, `name`, `type`, `kind`, `version`
- **Entity Types**: `user`, `client`, `agent`, `server`, etc.
- **Step Types**: `deterministic`, `llm`, `human`, `external`, etc.
- **Protocol Kinds**: `aauth`, `mcp`, `a2a`, `process`
- **Field Names**: `entities`, `flows`, `phases`, `from`, `to`, etc.
- **Values**: strings, numbers, booleans

## Development

```bash
# Install dependencies
npm install

# Compile TypeScript
npm run compile

# Watch for changes
npm run watch

# Run linter
npm run lint

# Package extension
npm run package
```

## License

MIT License - see [LICENSE](../../LICENSE)
