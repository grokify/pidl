# CLI Reference

```
pidl <command> [options] [arguments]
```

## Commands

| Command | Description |
|---------|-------------|
| `validate` | Validate PIDL JSON files |
| `generate` | Generate diagrams from PIDL files |
| `examples` | List or show built-in examples |
| `init` | Create a new PIDL file from template |
| `version` | Print version information |
| `help` | Show help message |

## validate

Validate PIDL JSON files for correctness.

```bash
pidl validate [options] <file> [file...]
```

### Options

| Option | Description |
|--------|-------------|
| `-q` | Quiet mode (only show errors) |

### Examples

```bash
# Validate single file
pidl validate protocol.json

# Validate multiple files
pidl validate *.json

# Quiet mode
pidl validate -q protocol.json
```

## generate

Generate diagrams from PIDL files.

```bash
pidl generate [options] <file>
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `-f` | Output format | `plantuml` |
| `-o` | Output file | stdout |

### Formats

| Format | Description |
|--------|-------------|
| `plantuml` | PlantUML sequence diagram |
| `mermaid` | Mermaid sequence diagram |
| `dot` | Graphviz DOT data flow |
| `d2` | D2 sequence diagram |
| `d2-flow` | D2 data flow diagram |
| `d2-arch` | D2 architecture diagram |

### Examples

```bash
# PlantUML to stdout
pidl generate protocol.json

# Mermaid to file
pidl generate -f mermaid -o diagram.mmd protocol.json

# D2 sequence diagram
pidl generate -f d2 protocol.json

# D2 architecture diagram
pidl generate -f d2-arch protocol.json

# From built-in example
pidl generate oauth2_authorization_code
```

## examples

List or show built-in protocol examples.

```bash
pidl examples [options] [name]
```

### Options

| Option | Description |
|--------|-------------|
| `-json` | Show example JSON content |

### Examples

```bash
# List all examples
pidl examples

# Show example JSON
pidl examples -json oauth2_authorization_code
```

## init

Create a new PIDL file from template or example.

```bash
pidl init [options] <filename>
```

### Options

| Option | Description |
|--------|-------------|
| `-name` | Protocol name |
| `-from` | Initialize from example |

### Examples

```bash
# Create minimal protocol
pidl init my-protocol.json

# Create with name
pidl init -name "My Protocol" my-protocol.json

# Copy from example
pidl init -from oauth2_authorization_code my-oauth.json
```

## version

Print version information.

```bash
pidl version
```

## help

Show help message.

```bash
pidl help
pidl help <command>
```
