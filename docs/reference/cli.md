# CLI Reference

```
pidl <command> [options] [arguments]
```

## Commands

| Command | Description |
|---------|-------------|
| `validate` | Validate PIDL JSON files |
| `generate` | Generate diagrams from PIDL files |
| `scenarios` | List the scenarios in a PIDL file |
| `samples` | Resolve and check code samples from source repositories |
| `roles` | List protocol roles |
| `components` | List deployment components |
| `trust` | List trust relationships |
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
| `-o` | Output file, or directory with `--all-scenarios` | stdout |
| `--scenario` | Scenario to render (required for files with scenarios, except whole-file formats) | |
| `--expand` | Show included fragments step by step instead of as boxes | false |
| `--all-scenarios` | Render every main scenario to `NN-<id><ext>` in the `-o` directory | false |
| `--include-fragments` | With `--all-scenarios`, also render fragments after the main scenarios | false |
| `--resolve` | Resolve `imports` and `extends` before generating | false |
| `--samples` | Samples for `walkthrough-json` and `viewer`: the document's lock file or a [sample set](../guide/samples.md#sample-sets). Can be repeated | `<file without .json>.samples.json`, when present |
| `--client` | Postman: the role whose requests are sent | the first `agent` or `client` entity |
| `--mock` | Postman: save example responses for a mock server and check only status codes | false |
| `--signer` | Postman: `proxy` sends requests with HTTP message signatures through a [signing proxy](../guide/postman.md#signing-proxies) | |
| `--targets` | Postman environments: a JSON file naming each target's role URLs and signer | |
| `--target` | Postman environments: the target in `--targets` to write | |

### Formats

| Format | Description |
|--------|-------------|
| `plantuml` | PlantUML sequence diagram |
| `mermaid` | Mermaid sequence diagram |
| `mermaid-state` | Mermaid state diagram (requires entity states) |
| `mermaid-component` | Mermaid deployment component diagram |
| `mermaid-trust` | Mermaid trust relationship diagram |
| `markdown-matrix` | Markdown protocol role matrix |
| `dot` | Graphviz DOT data flow |
| `d2` | D2 sequence diagram |
| `d2-flow` | D2 data flow diagram |
| `d2-arch` | D2 architecture diagram |
| `svg`, `svg-animated` | SVG sequence diagram, optionally with animated flow dots |
| `svg-network` | Network boundary diagram |
| `svg-component`, `svg-trust` | SVG deployment component and trust relationship diagrams |
| `infographic` | Compact SVG for LinkedIn and datasheets |
| `scenario-matrix` | Markdown table of roles by scenario (whole file) |
| `token-table` | Markdown table of tokens, checkers and claims (whole file) |
| `outline` | Markdown outline of roles, scenarios and tokens (whole file) |
| `outline-json` | The outline as JSON (whole file) |
| `walkthrough-json` | A [walkthrough](walkthrough-format.md) (`pidl-walkthrough/v1`): expanded scenarios with examples and findings, for viewers (whole file) |
| `viewer` | Self-contained HTML walkthrough viewer (whole file) |
| `postman` | [Postman collection](../guide/postman.md): a folder per scenario, chained requests and checks (whole file) |
| `postman-environment` | Postman environment for one target in a `--targets` file |

### Examples

```bash
# One scenario of a multi-scenario file
pidl generate -f mermaid --scenario client-credentials oauth2_api_access

# Every scenario, numbered in file order
pidl generate -f svg --all-scenarios -o diagrams/ oauth2_api_access

# Step-by-step walkthrough viewer in one HTML file
pidl generate -f viewer -o api.html oauth2_api_access

# Postman collection that sends signed requests through a signing proxy
pidl generate -f postman --signer proxy -o api.postman_collection.json oauth2_api_access

# Viewer with the code samples of two implementations
pidl generate -f viewer --samples sdk-go.pidl-samples.json --samples sdk-ts.pidl-samples.json -o api.html api.json

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

## scenarios

List the scenarios in a PIDL file: main scenarios in file order, then fragments. The `#` column matches the file numbers written by `generate --all-scenarios --include-fragments`.

```bash
pidl scenarios [options] <file>
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `-f` | Output format: table, json | `table` |

### Examples

```bash
pidl scenarios oauth2_api_access
pidl scenarios -f json oauth2_api_access
```

See the [Scenarios guide](../guide/scenarios.md).

## samples

Resolve and check [code samples](../guide/samples.md) whose code comes from marked regions of files in source repositories.

```bash
pidl samples <resolve|check> [options] <file>
```

`<file>` is a document whose steps carry samples, or a [sample set](../guide/samples.md#sample-sets) (`pidl-samples/v1`). `resolve` reads every sample source and records the code and its SHA-256: for a document in a lock file beside it, for a sample set in the set itself. `check` fails when a source is not resolved, its region is missing, or its code changed since it was resolved; run it in CI.

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--root` | Directory holding repository checkouts at `<root>/<repo>` | `$GOPATH/src`, or `~/go/src` |
| `--lock` | A document's lock file | `<file without .json>.samples.json` |
| `--repo-dir` | Read every source from this checkout instead of `<root>/<repo>`, such as the repository a sample set lives in | |
| `--protocol` | With a sample set, also check that every step address names a step of this document | |

### Examples

```bash
# Write a document's lock file
pidl samples resolve api.json

# Resolve an implementation's sample set from its own checkout
pidl samples resolve --repo-dir . sdk-go.pidl-samples.json

# In the implementation's CI: fail when a sample fell behind its source
# or names a step the protocol no longer has
pidl samples check --repo-dir . --protocol api.json sdk-go.pidl-samples.json
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
| `--scenario` | Analyze one scenario (default: the whole file) |

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
| SEC017 | Low | Role that no flow uses |
| SEC018 | Medium | Token checked before any earlier step issues it |
| SEC019 | Low | Request signed by someone other than its sender or the token holder |
| SEC020 | Low | Token checked by an entity not in its `verified_by` |
| SEC021 | Info | Fragment that no scenario includes |

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
