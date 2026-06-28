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
| `simulate` | Run protocol simulation with state tracking |
| `diff` | Compare two protocols and show differences |
| `debug` | Interactive step-through protocol debugger |
| `analyze` | Security analysis with attack surface detection |
| `resolve` | Resolve protocol imports and inheritance |
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

## simulate

Run protocol simulation with state tracking.

```bash
pidl simulate [options] <file>
```

### Options

| Option | Description |
|--------|-------------|
| `-steps` | Maximum steps to execute (0 = all) |
| `-v` | Verbose output with each step |
| `-json` | Output trace as JSON |
| `--trace-format` | Output format: text, json, svg, mermaid |
| `--trace-output` | Write trace to file |
| `--show-states` | Show entity states in trace |
| `--show-timings` | Show timing information |

### Examples

```bash
# Run full simulation
pidl simulate protocol.json

# Verbose with step-by-step output
pidl simulate -v protocol.json

# Output trace as SVG
pidl simulate --trace-format=svg --trace-output=trace.svg protocol.json

# Run only 5 steps
pidl simulate -steps=5 protocol.json
```

## diff

Compare two protocols and show differences.

```bash
pidl diff [options] <base-file> <new-file>
```

### Options

| Option | Description |
|--------|-------------|
| `-f, --format` | Output format: text, json, markdown (default: text) |
| `-o, --output` | Output file (default: stdout) |
| `--ignore-metadata` | Ignore metadata differences |
| `-q, --quiet` | Summary only |

### Examples

```bash
# Compare two protocols
pidl diff base.json updated.json

# Output as markdown
pidl diff -f markdown base.json updated.json

# Summary only
pidl diff -q base.json updated.json
```

## debug

Interactive step-through protocol debugger.

```bash
pidl debug <file>
```

### Interactive Commands

| Command | Description |
|---------|-------------|
| `step`, `s` | Execute next flow |
| `continue`, `c` | Run until breakpoint or completion |
| `break <idx> [cond]` | Set breakpoint at flow index |
| `delete <idx>` | Remove breakpoint |
| `breakpoints` | List all breakpoints |
| `inspect` | Show current state |
| `inspect entity <id>` | Show entity details and state |
| `inspect flow <idx>` | Show flow details |
| `list` | List flows with position marker |
| `set <entity> <state>` | Set entity state |
| `reset` | Restart execution |
| `quit`, `q` | Exit debugger |

### Examples

```bash
# Start debugger
pidl debug protocol.json

# In debugger:
> break 3              # Set breakpoint at flow 3
> continue             # Run until breakpoint
> inspect              # Show current state
> step                 # Execute next flow
> quit                 # Exit
```

## analyze

Security analysis with attack surface detection.

```bash
pidl analyze [options] <file>
```

### Options

| Option | Description |
|--------|-------------|
| `-f, --format` | Output format: text, json, markdown (default: text) |
| `-o, --output` | Output file (default: stdout) |
| `--min-severity` | Minimum severity: critical, high, medium, low, info |
| `--category` | Filter by category (repeatable) |
| `--fail-on` | Exit code 1 if risks at this severity |
| `-q, --quiet` | Summary only |

### Built-in Rules

| Rule | Severity | Description |
|------|----------|-------------|
| SEC001 | High | Trust boundary violation |
| SEC002 | High | Missing encryption on confidential flow |
| SEC003 | Medium | Unbound bearer token |
| SEC004 | High | Missing authentication on external flow |
| SEC005 | Medium | JWT without defined audience |
| SEC006 | Medium | Token transmitted in redirect |
| SEC007 | Medium | Missing mTLS on sensitive flow |
| SEC008 | Low | Entity without defined trust level |
| SEC009 | Medium | Sensitive data in redirect parameters |
| SEC010 | Medium | Weak authentication method |

### Examples

```bash
# Run security analysis
pidl analyze protocol.json

# Output as markdown
pidl analyze -f markdown -o security-report.md protocol.json

# Fail if high-severity risks found (for CI)
pidl analyze --fail-on=high protocol.json

# Only show critical and high
pidl analyze --min-severity=high protocol.json
```

## resolve

Resolve protocol imports and inheritance.

```bash
pidl resolve <file>
```

### Examples

```bash
# Resolve and output merged protocol
pidl resolve protocol-with-imports.json

# Save resolved protocol
pidl resolve protocol.json > resolved.json
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
