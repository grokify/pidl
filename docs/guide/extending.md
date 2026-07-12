# Extending PIDL

This guide explains how to extend PIDL with custom renderers, validation rules, and exporters.

## Custom Renderers

PIDL uses a `Renderer` interface that you can implement to create custom diagram output formats.

### The Renderer Interface

```go
package render

import (
    "io"
    "github.com/grokify/pidl"
)

type Renderer interface {
    // Render writes the diagram to the writer.
    Render(w io.Writer, p *pidl.Protocol) error

    // RenderString returns the diagram as a string.
    RenderString(p *pidl.Protocol) (string, error)

    // Format returns the output format name.
    Format() Format
}
```

### Implementing a Custom Renderer

Here's a complete example of a custom Markdown renderer:

```go
package render

import (
    "fmt"
    "io"
    "strings"

    "github.com/grokify/pidl"
)

// MyMarkdownRenderer generates a custom Markdown format.
type MyMarkdownRenderer struct {
    // Title includes the protocol name as header.
    Title bool
    // ShowDescriptions includes flow descriptions.
    ShowDescriptions bool
}

// NewMyMarkdown creates a new renderer with defaults.
func NewMyMarkdown() *MyMarkdownRenderer {
    return &MyMarkdownRenderer{
        Title:            true,
        ShowDescriptions: true,
    }
}

// Format returns the format identifier.
func (r *MyMarkdownRenderer) Format() Format {
    return Format("my-markdown")
}

// Render writes the Markdown to the writer.
func (r *MyMarkdownRenderer) Render(w io.Writer, p *pidl.Protocol) error {
    s, err := r.RenderString(p)
    if err != nil {
        return err
    }
    _, err = w.Write([]byte(s))
    return err
}

// RenderString returns the Markdown as a string.
func (r *MyMarkdownRenderer) RenderString(p *pidl.Protocol) (string, error) {
    var sb strings.Builder

    // Write title
    if r.Title {
        sb.WriteString(fmt.Sprintf("# %s\n\n", p.ProtocolMeta.Name))
    }

    // Write entities
    sb.WriteString("## Entities\n\n")
    for _, e := range p.Entities {
        sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", e.Name, e.Type))
    }
    sb.WriteString("\n")

    // Write flows
    sb.WriteString("## Flows\n\n")
    for i, f := range p.Flows {
        sb.WriteString(fmt.Sprintf("%d. %s → %s: %s\n",
            i+1, f.From, f.To, f.Label))
        if r.ShowDescriptions && f.Description != "" {
            sb.WriteString(fmt.Sprintf("   - %s\n", f.Description))
        }
    }

    return sb.String(), nil
}
```

### Registering a Custom Renderer

To integrate your renderer with the `render.New()` factory:

```go
// In your application code
func init() {
    // Register with the format parser
    render.RegisterFormat("my-markdown", func() render.Renderer {
        return NewMyMarkdown()
    })
}
```

Or use it directly:

```go
renderer := NewMyMarkdown()
output, err := renderer.RenderString(protocol)
```

### Best Practices for Renderers

1. **Support configuration options** - Use struct fields for customization
2. **Handle both io.Writer and string output** - Implement both methods
3. **Check protocol kind** - Process specs may need different handling
4. **Handle empty protocols gracefully** - Don't panic on missing data
5. **Consider step types** - For process specs, use `StepType` for visual differentiation

## Custom Validation Rules

PIDL's security analysis system uses rules that can be extended.

### The SecurityRule Structure

```go
type SecurityRule struct {
    // ID is a unique identifier (e.g., "SEC001")
    ID string

    // Name is a human-readable name
    Name string

    // Description explains what the rule checks
    Description string

    // Category groups related rules
    Category string

    // Severity indicates the risk level
    Severity Severity

    // Check is the function that evaluates the rule
    Check func(p *pidl.Protocol, opts *AnalysisOptions) []Risk
}
```

### Implementing a Custom Rule

```go
package analyze

import "github.com/grokify/pidl"

// MyCustomRule checks for a specific security pattern.
var MyCustomRule = SecurityRule{
    ID:          "CUSTOM001",
    Name:        "Unencrypted Sensitive Data",
    Description: "Sensitive data should use encryption",
    Category:    CategoryDataProtection,
    Severity:    SeverityHigh,
    Check: func(p *pidl.Protocol, opts *AnalysisOptions) []Risk {
        var risks []Risk

        for _, flow := range p.Flows {
            // Check if flow has sensitive data
            if flow.Annotation != nil && flow.Annotation.DataClassification == "confidential" {
                // Check if encryption is required
                if !hasSecurityRequirement(flow, "encryption") {
                    risks = append(risks, Risk{
                        RuleID:      "CUSTOM001",
                        Severity:    SeverityHigh,
                        Category:    CategoryDataProtection,
                        Location:    fmt.Sprintf("flow %s→%s", flow.From, flow.To),
                        Message:     "Confidential data transmitted without encryption",
                        Remediation: "Add security requirement: encryption",
                    })
                }
            }
        }

        return risks
    },
}

func hasSecurityRequirement(flow pidl.Flow, req string) bool {
    if flow.Security == nil {
        return false
    }
    for _, r := range flow.Security.Requires {
        if r == req {
            return true
        }
    }
    return false
}
```

### Registering Custom Rules

```go
// Add your rule to the analyzer
func init() {
    analyze.RegisterRule(MyCustomRule)
}

// Or create a custom analyzer
analyzer := analyze.NewAnalyzer()
analyzer.AddRule(MyCustomRule)

results := analyzer.Analyze(protocol)
```

### Rule Categories

| Category | Description |
|----------|-------------|
| `CategoryAuthentication` | Identity verification issues |
| `CategoryAuthorization` | Access control issues |
| `CategoryTokenSecurity` | Token handling issues |
| `CategoryDataProtection` | Data security issues |
| `CategoryNetworkSecurity` | Network boundary issues |
| `CategoryProcessSecurity` | Process spec issues |

### Severity Levels

| Severity | When to Use |
|----------|-------------|
| `SeverityCritical` | Immediate security threat |
| `SeverityHigh` | Significant security risk |
| `SeverityMedium` | Moderate security concern |
| `SeverityLow` | Minor issue or best practice |
| `SeverityInfo` | Informational finding |

## Custom Exporters

Workflow exporters convert PIDL process specs to executable formats.

### The Exporter Pattern

```go
package export

import "github.com/grokify/pidl"

type MyExporter struct {
    // Configuration options
    OutputFormat string
}

func NewMyExporter() *MyExporter {
    return &MyExporter{
        OutputFormat: "default",
    }
}

func (e *MyExporter) Export(p *pidl.Protocol) (string, error) {
    // Validate it's a process spec
    if p.ProtocolMeta.Kind != pidl.ProtocolKindProcess {
        return "", fmt.Errorf("only process specs supported")
    }

    // Build your output format
    var sb strings.Builder

    for _, entity := range p.Entities {
        if entity.StepType == "" {
            continue
        }

        // Generate code for this step
        sb.WriteString(e.generateStep(entity))
    }

    return sb.String(), nil
}

func (e *MyExporter) generateStep(entity pidl.Entity) string {
    // Step-type-specific code generation
    switch entity.StepType {
    case pidl.StepTypeDeterministic:
        return fmt.Sprintf("// Deterministic: %s\n", entity.Name)
    case pidl.StepTypeLLM:
        return fmt.Sprintf("// LLM: %s\n", entity.Name)
    case pidl.StepTypeHuman:
        return fmt.Sprintf("// Human: %s\n", entity.Name)
    default:
        return fmt.Sprintf("// Step: %s\n", entity.Name)
    }
}
```

### Using Parallel Block Detection

```go
import "github.com/grokify/pidl"

func (e *MyExporter) Export(p *pidl.Protocol) (string, error) {
    // Detect parallel execution blocks
    blocks := pidl.DetectParallelBlocks(p)

    for _, block := range blocks {
        fmt.Printf("Parallel block: %s\n", block.ID)
        fmt.Printf("  Join condition: %s\n", block.JoinCondition)

        for _, branch := range block.Branches {
            fmt.Printf("  Branch: %s\n", branch.EntityID)
        }
    }

    // Use execution graph for ordering
    graph := pidl.AnalyzeExecutionGraph(p)
    for _, stage := range graph.Stages {
        fmt.Printf("Stage %s: %v\n", stage.ID, stage.Steps)
    }

    return "", nil
}
```

### Using Cost Information

```go
import "github.com/grokify/pidl"

func (e *MyExporter) Export(p *pidl.Protocol) (string, error) {
    // Analyze costs
    analysis := pidl.AnalyzeProcessCosts(p)

    fmt.Printf("Total estimated cost: $%.4f\n",
        analysis.TotalEstimate.ExpectedCost)

    for stepType, cost := range analysis.CostByType {
        fmt.Printf("  %s: $%.4f\n", stepType, cost)
    }

    return "", nil
}
```

## Schema Validation Extension

You can extend schema validation with custom validators.

### Custom Port Validator

```go
import "github.com/grokify/pidl"

type CustomValidator struct {
    *pidl.PortSchemaValidator
    customRules []func(port pidl.DataPort, data interface{}) error
}

func NewCustomValidator() *CustomValidator {
    return &CustomValidator{
        PortSchemaValidator: pidl.NewPortSchemaValidator(),
    }
}

func (v *CustomValidator) AddRule(rule func(pidl.DataPort, interface{}) error) {
    v.customRules = append(v.customRules, rule)
}

func (v *CustomValidator) ValidatePortData(port pidl.DataPort, data interface{}) (*pidl.SchemaValidationResult, error) {
    // Run base validation
    result, err := v.PortSchemaValidator.ValidatePortData(port, data)
    if err != nil {
        return nil, err
    }

    // Run custom rules
    for _, rule := range v.customRules {
        if err := rule(port, data); err != nil {
            result.Valid = false
            result.Errors = append(result.Errors, pidl.SchemaValidationError{
                Path:    port.Name,
                Message: err.Error(),
            })
        }
    }

    return result, nil
}
```

## Testing Extensions

### Testing Renderers

```go
func TestMyRenderer(t *testing.T) {
    p := &pidl.Protocol{
        ProtocolMeta: pidl.ProtocolMeta{
            ID:   "test",
            Name: "Test Protocol",
        },
        Entities: []pidl.Entity{
            {ID: "a", Name: "Service A"},
            {ID: "b", Name: "Service B"},
        },
        Flows: []pidl.Flow{
            {From: "a", To: "b", Label: "Request"},
        },
    }

    r := NewMyMarkdown()
    output, err := r.RenderString(p)
    if err != nil {
        t.Fatalf("Render failed: %v", err)
    }

    if !strings.Contains(output, "# Test Protocol") {
        t.Error("expected title in output")
    }
}
```

### Testing Rules

```go
func TestMyCustomRule(t *testing.T) {
    p := &pidl.Protocol{
        // ... protocol with security issue
    }

    opts := analyze.DefaultAnalysisOptions()
    risks := MyCustomRule.Check(p, opts)

    if len(risks) == 0 {
        t.Error("expected rule to find risk")
    }
}
```

## See Also

- [Process Specifications](process.md) - Process spec fundamentals
- [Workflow Exports](workflow-exports.md) - Built-in exporters
- [Go Library Reference](../reference/library.md) - Full API documentation
