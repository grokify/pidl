# Examples

PIDL includes built-in examples for common protocol patterns.

## Available Examples

| Example | Protocol | Category |
|---------|----------|----------|
| `oauth2_authorization_code` | OAuth 2.0 Authorization Code Flow | auth |
| `oauth2_pkce` | OAuth 2.0 with PKCE | auth |
| `oidc_authentication` | OpenID Connect Authentication | auth |
| `mcp_tool_invocation` | MCP Tool Invocation | agent |
| `a2a_agent_delegation` | A2A Agent Delegation | agent |

## Using Examples

### CLI

```bash
# List examples
pidl examples

# Show example JSON
pidl examples -json oauth2_authorization_code

# Generate diagram from example
pidl generate oauth2_pkce

# Copy example to file
pidl init -from oauth2_authorization_code my-protocol.json
```

### Go Library

```go
import "github.com/grokify/pidl/examples"

// List all
names := examples.List()

// Get JSON
json, err := examples.GetJSON("oauth2_pkce")

// Get parsed protocol
p, err := examples.GetProtocol("oauth2_pkce")
```

## Example Categories

### OAuth 2.0

- [OAuth 2.0 Authorization Code](oauth2.md#authorization-code-flow)
- [OAuth 2.0 with PKCE](oauth2.md#pkce-flow)
- [OpenID Connect](oauth2.md#openid-connect)

### Agent Protocols

- [MCP Tool Invocation](mcp.md)
- [A2A Agent Delegation](a2a.md)
