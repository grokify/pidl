# Go Library Reference

## Installation

```bash
go get github.com/grokify/pidl
```

## Packages

| Package | Description |
|---------|-------------|
| `github.com/grokify/pidl` | Core types and parsing |
| `github.com/grokify/pidl/render` | Diagram rendering |
| `github.com/grokify/pidl/examples` | Built-in examples |
| `github.com/grokify/pidl/schema` | Embedded JSON Schema |

## Core Types

### Protocol

```go
type Protocol struct {
    ProtocolMeta ProtocolMeta `json:"protocol"`
    Entities     []Entity     `json:"entities"`
    Phases       []Phase      `json:"phases,omitempty"`
    Flows        []Flow       `json:"flows"`
}
```

### Entity

```go
type Entity struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Type        EntityType `json:"type"`
    Description string     `json:"description,omitempty"`
}
```

### Phase

```go
type Phase struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Parent      string `json:"parent,omitempty"`
}
```

### Flow

```go
type Flow struct {
    From         string        `json:"from"`
    To           string        `json:"to"`
    Action       string        `json:"action"`
    Label        string        `json:"label,omitempty"`
    Mode         FlowMode      `json:"mode,omitempty"`
    Phase        string        `json:"phase,omitempty"`
    Description  string        `json:"description,omitempty"`
    Sequence     int           `json:"sequence,omitempty"`
    Condition    string        `json:"condition,omitempty"`
    Note         string        `json:"note,omitempty"`
    Annotations  []Annotation  `json:"annotations,omitempty"`
    Alternatives []Alternative `json:"alternatives,omitempty"`
}
```

### Annotation

```go
type Annotation struct {
    Type    AnnotationType `json:"type"`
    Text    string         `json:"text"`
    Details string         `json:"details,omitempty"`
}
```

### Alternative

```go
type Alternative struct {
    Condition   string `json:"condition"`
    Flows       []Flow `json:"flows"`
    Description string `json:"description,omitempty"`
}
```

## Parsing

```go
// Parse from file
p, err := pidl.ParseFile("protocol.json")

// Parse from bytes
p, err := pidl.Parse(jsonBytes)

// Parse from reader
p, err := pidl.ParseReader(reader)

// Must parse (panics on error)
p := pidl.MustParse(jsonBytes)
```

## Validation

```go
// Validate and get errors
errs := p.Validate()
if errs.HasErrors() {
    for _, e := range errs {
        fmt.Printf("%s: %s\n", e.Field, e.Message)
    }
}

// Quick validity check
if p.IsValid() {
    // proceed
}
```

## Protocol Methods

```go
// Find entities/phases
entity := p.EntityByID("client")
phase := p.PhaseByID("auth")

// Get flows by phase
flows := p.FlowsByPhase("auth")

// Get all IDs
entityIDs := p.EntityIDs()
phaseIDs := p.PhaseIDs()

// Phase hierarchy
roots := p.RootPhases()
children := p.ChildPhases("auth")
depth := p.PhaseDepth("mfa")
```

## Flow Methods

```go
// Display helpers
label := flow.DisplayLabel()  // Label or Action
mode := flow.EffectiveMode()  // Mode or FlowModeRequest

// Feature checks
if flow.HasCondition() { ... }
if flow.HasNote() { ... }
if flow.HasAnnotations() { ... }
if flow.HasAlternatives() { ... }
```

## Rendering

```go
import "github.com/grokify/pidl/render"

// Create renderer
r := render.NewMermaid()
r := render.NewPlantUML()
r := render.NewD2()
r := render.NewDOT()

// Render to string
diagram, err := r.RenderString(p)

// Render to writer
err := r.Render(os.Stdout, p)

// Quick render by format
diagram, err := render.RenderString(render.FormatMermaid, p)
```

## Examples Package

```go
import "github.com/grokify/pidl/examples"

// List available examples
names := examples.List()

// Get example JSON
jsonBytes, err := examples.GetJSON("oauth2_authorization_code")

// Get parsed protocol
p, err := examples.GetProtocol("oauth2_authorization_code")

// Get all protocols
all, err := examples.All()
```

## Creating Protocols

```go
// Create minimal protocol
p := pidl.NewMinimalProtocol("my-protocol", "My Protocol")

// Write to file
err := pidl.WriteProtocolFile("output.json", p)
```

## Full Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/grokify/pidl"
    "github.com/grokify/pidl/render"
    "github.com/grokify/pidl/examples"
)

func main() {
    // Load built-in example
    p, err := examples.GetProtocol("oauth2_pkce")
    if err != nil {
        log.Fatal(err)
    }

    // Validate
    if errs := p.Validate(); errs.HasErrors() {
        log.Fatal(errs)
    }

    // Generate Mermaid diagram
    r := render.NewMermaid()
    r.Autonumber = true

    diagram, err := r.RenderString(p)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(diagram)
}
```
