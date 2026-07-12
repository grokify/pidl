# PIDL Validate GitHub Action

A GitHub Action for validating PIDL (Protocol Interaction Description Language) files in your repository.

## Usage

Add this action to your workflow to validate PIDL files on every push or pull request:

```yaml
name: Validate PIDL

on:
  push:
    paths:
      - '**/*.pidl.json'
  pull_request:
    paths:
      - '**/*.pidl.json'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Validate PIDL files
        uses: grokify/pidl/integrations/github-action@main
        with:
          files: '**/*.pidl.json'
          security: 'true'
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `files` | Glob pattern for PIDL files to validate | No | `**/*.pidl.json` |
| `strict` | Enable strict validation mode | No | `false` |
| `security` | Run security analysis on protocols | No | `false` |
| `version` | PIDL CLI version to use | No | `latest` |

## Outputs

| Output | Description |
|--------|-------------|
| `valid` | Whether all files passed validation (`true`/`false`) |
| `file_count` | Number of files validated |
| `errors` | JSON array of validation errors |

## Examples

### Basic Validation

```yaml
- uses: grokify/pidl/integrations/github-action@main
```

### With Security Analysis

```yaml
- uses: grokify/pidl/integrations/github-action@main
  with:
    security: 'true'
```

### Specific Version

```yaml
- uses: grokify/pidl/integrations/github-action@main
  with:
    version: 'v0.8.0'
```

### Using Outputs

```yaml
- name: Validate PIDL files
  id: pidl
  uses: grokify/pidl/integrations/github-action@main

- name: Check results
  if: steps.pidl.outputs.valid == 'false'
  run: |
    echo "Validation failed for ${{ steps.pidl.outputs.file_count }} files"
    echo "${{ steps.pidl.outputs.errors }}" | jq .
```

### In a Matrix Build

```yaml
jobs:
  validate:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        protocol-dir: [protocols/auth, protocols/api, protocols/mcp]
    steps:
      - uses: actions/checkout@v4
      - uses: grokify/pidl/integrations/github-action@main
        with:
          files: '${{ matrix.protocol-dir }}/*.pidl.json'
```

## Validation Checks

The action validates:

- JSON syntax
- PIDL schema compliance
- Entity definitions
- Flow references
- Phase structure

With `security: true`:

- Token handling patterns
- Consent requirements
- Sensitive data flows
- Authentication flows

## License

MIT License - see [LICENSE](../../LICENSE)
