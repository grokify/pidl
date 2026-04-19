# Quick Start

## List Built-in Examples

```bash
pidl examples
```

Output:

```
Available examples:
  a2a_agent_delegation
  mcp_tool_invocation
  oauth2_authorization_code
  oauth2_pkce
  oidc_authentication
```

## Generate Diagrams

### PlantUML

```bash
pidl generate oauth2_authorization_code
```

### Mermaid

```bash
pidl generate -f mermaid oauth2_pkce
```

### D2

```bash
# Sequence diagram
pidl generate -f d2 oauth2_pkce

# Data flow diagram
pidl generate -f d2-flow oauth2_pkce

# Architecture diagram
pidl generate -f d2-arch oauth2_pkce
```

### Graphviz DOT

```bash
pidl generate -f dot mcp_tool_invocation
```

## Create a New Protocol

### From scratch

```bash
pidl init my-protocol.json
```

### From an example

```bash
pidl init -from oauth2_authorization_code my-oauth.json
```

## Validate Protocol Files

```bash
pidl validate my-protocol.json
pidl validate *.json
```

## View Example Content

```bash
pidl examples -json oauth2_authorization_code
```

## Use as Go Library

```go
package main

import (
    "fmt"
    "log"

    "github.com/grokify/pidl"
    "github.com/grokify/pidl/render"
)

func main() {
    // Parse a PIDL file
    p, err := pidl.ParseFile("protocol.json")
    if err != nil {
        log.Fatal(err)
    }

    // Validate
    if errs := p.Validate(); errs.HasErrors() {
        log.Fatal(errs)
    }

    // Generate Mermaid diagram
    diagram, err := render.RenderString(render.FormatMermaid, p)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(diagram)
}
```
