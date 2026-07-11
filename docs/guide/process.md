# Process Specifications

PIDL supports two profiles: **protocol** (default) and **process**. This guide explains when to use each and how to model process workflows.

## Protocol vs Process

| Aspect | Protocol | Process |
|--------|----------|---------|
| **Focus** | System interactions | Data transformations |
| **Entities** | Services, actors, systems | Processing steps |
| **Flows** | Request/response patterns | Data flow between steps |
| **Primary concern** | "Who talks to whom" | "How data transforms" |
| **Output** | Sequence diagrams | Pipeline diagrams |

### When to Use Protocol

Use the **protocol** profile when modeling:

- Authentication/authorization flows (OAuth, OIDC, SAML)
- API choreography between services
- Agent protocols (MCP, A2A)
- Multi-party interactions with clear request/response patterns
- Trust relationships between systems

Example: OAuth 2.0 authorization code flow, MCP tool invocation.

### When to Use Process

Use the **process** profile when modeling:

- Data pipelines (ETL, data processing)
- AI/ML workflows with LLM steps
- Document processing workflows
- CI/CD pipelines
- Approval workflows with human review

Example: Document extraction → LLM transformation → human review → database load.

## Defining a Process Spec

Set `kind: "process"` in the protocol metadata:

```json
{
  "protocol": {
    "id": "document-pipeline",
    "name": "Document Processing Pipeline",
    "kind": "process"
  }
}
```

## Step Types

Each entity in a process spec should have a `step_type`:

```json
{
  "entities": [
    {
      "id": "extract",
      "name": "Extract Data",
      "type": "server",
      "step_type": "deterministic"
    },
    {
      "id": "summarize",
      "name": "LLM Summarization",
      "type": "server",
      "step_type": "llm"
    }
  ]
}
```

### Step Type Reference

| Step Type | When to Use | Rendering |
|-----------|-------------|-----------|
| `deterministic` | Predictable transformations, parsing, validation | ⚙️ Blue rectangle |
| `llm` | AI/ML processing, text generation, embeddings | 🧠 Purple hexagon |
| `human` | Manual review, approval, data entry | 👤 Green circle |
| `external` | API calls, third-party services | ☁️ Orange cloud |
| `tool` | Function calls, utility operations | 🔧 Gray diamond |

## Data Ports

Define explicit inputs and outputs for each step:

```json
{
  "id": "transform",
  "name": "Transform Step",
  "step_type": "llm",
  "inputs": [
    {
      "name": "raw_document",
      "kind": "file",
      "required": true,
      "description": "PDF or text document to process"
    }
  ],
  "outputs": [
    {
      "name": "summary",
      "kind": "object",
      "description": "Structured summary of the document"
    },
    {
      "name": "embeddings",
      "kind": "object",
      "sensitive": false,
      "description": "Vector embeddings for search"
    }
  ]
}
```

### Port Kinds

| Kind | Icon | Use For |
|------|------|---------|
| `file` | 📄 | Documents, images, binary data |
| `object` | 📦 | JSON objects, structs, in-memory data |
| `api` | 🌐 | API request/response payloads |
| `database` | 🗄️ | Database records, query results |
| `queue` | 📬 | Message queue items |
| `stream` | 🌊 | Streaming data, real-time feeds |

## Processing Configuration

Configure execution characteristics:

```json
{
  "id": "llm_step",
  "step_type": "llm",
  "processing": {
    "engine": "claude-3-opus",
    "deterministic": false,
    "timeout": "PT60S",
    "cache_key": "document_hash"
  }
}
```

## Failure Handling

Document failure modes and retry strategies:

```json
{
  "id": "api_call",
  "step_type": "external",
  "failure_modes": [
    {
      "id": "timeout",
      "name": "Request Timeout",
      "severity": "medium",
      "recovery": "Retry with exponential backoff"
    }
  ],
  "retry_strategy": {
    "max_attempts": 3,
    "initial_delay": "PT1S",
    "backoff_multiplier": 2.0
  }
}
```

## Security Considerations

Process specs trigger additional security analysis rules:

| Rule | Severity | Issue |
|------|----------|-------|
| SEC011 | Medium | LLM step without downstream validation |
| SEC012 | High | Sensitive data flows to LLM step |
| SEC013 | Medium | Non-deterministic step in critical path |
| SEC014 | Low | External step without failure modes |
| SEC015 | Medium | Human step without timeout |

### Best Practices

1. **Validate LLM outputs** - Add a validation step after LLM processing
2. **Mark sensitive data** - Use `sensitive: true` on ports containing PII
3. **Define failure modes** - Document how external steps can fail
4. **Set timeouts** - Human steps should have timeouts to prevent blocking
5. **Use deterministic steps** - For critical path operations that must be reproducible

## Rendering Process Specs

### Sequence Diagram

```bash
pidl generate -f svg process.json
```

Steps are styled by type with icons and colors.

### Infographic

```bash
# LinkedIn-friendly infographic
pidl generate -f infographic --size=linkedin-square --title="My Pipeline" process.json

# Datasheet tile
pidl generate -f infographic --size=datasheet-tile --theme=minimal process.json
```

Infographics render with:

- Custom shapes per step type (hexagon for LLM, circle for human, etc.)
- Animated dots showing data flow
- Step type icons and colors
- Compact layout for social media

### D2 Architecture

```bash
pidl generate -f d2-arch --show-ports process.json
```

Data ports are rendered as connected nodes grouped by kind.

## Example: Document Review Pipeline

```json
{
  "protocol": {
    "id": "doc-review",
    "name": "Document Review Pipeline",
    "kind": "process"
  },
  "entities": [
    {
      "id": "ingest",
      "name": "Document Ingestion",
      "step_type": "deterministic",
      "outputs": [{"name": "document", "kind": "file"}]
    },
    {
      "id": "extract",
      "name": "Content Extraction",
      "step_type": "deterministic",
      "inputs": [{"name": "document", "kind": "file", "required": true}],
      "outputs": [{"name": "content", "kind": "object"}]
    },
    {
      "id": "analyze",
      "name": "LLM Analysis",
      "step_type": "llm",
      "inputs": [{"name": "content", "kind": "object", "required": true}],
      "outputs": [{"name": "analysis", "kind": "object"}],
      "processing": {"engine": "claude-3-opus", "timeout": "PT60S"}
    },
    {
      "id": "review",
      "name": "Human Review",
      "step_type": "human",
      "inputs": [{"name": "analysis", "kind": "object", "required": true}],
      "outputs": [{"name": "approved", "kind": "object"}]
    },
    {
      "id": "store",
      "name": "Store Results",
      "step_type": "external",
      "inputs": [{"name": "approved", "kind": "object", "required": true}],
      "failure_modes": [{"id": "db_error", "severity": "high"}],
      "retry_strategy": {"max_attempts": 3}
    }
  ],
  "flows": [
    {"from": "ingest", "to": "extract", "action": "parse"},
    {"from": "extract", "to": "analyze", "action": "analyze"},
    {"from": "analyze", "to": "review", "action": "review"},
    {"from": "review", "to": "store", "action": "save"}
  ]
}
```

## Go Library Usage

```go
import "github.com/grokify/pidl"

// Check if protocol is a process spec
if p.IsProcessSpec() {
    // Get process steps
    steps := p.ProcessSteps()

    // Filter by step type
    llmSteps := p.LLMSteps()
    humanSteps := p.HumanSteps()

    // Check individual entities
    for _, e := range p.Entities {
        if e.IsLLMStep() {
            // Handle LLM step
        }

        // Get required inputs
        required := e.RequiredInputs()

        // Check for sensitive data
        sensitiveInputs := e.SensitiveInputs()
        sensitiveOutputs := e.SensitiveOutputs()
    }
}
```

## See Also

- [SPECIFICATION.md](../SPECIFICATION.md#process-profile) - Full process profile specification
- [Process Spec Examples](../examples/index.md) - Built-in process spec examples
- [Security Analysis](../reference/cli.md#analyze) - Process-aware security rules
