# Tokens Demo - Tokens

| Token | Type | Issuer | Audience | Binding | Verified by | Claims checked | Issued in | Verified in |
|-------|------|--------|----------|---------|-------------|----------------|-----------|-------------|
| **Resource token** | jwt | Resource | Person Server | dpop | Person Server, Access Server | `aud` (Person Server); `agent_jkt` (Person Server, Access Server) | Federated authorization | Federated authorization, PS to AS |
| **Auth token** | jwt | Access Server | Resource | dpop | Resource | `iss` (Resource) | PS to AS | Federated authorization |
| **Session token** | — | — | — | — | — | — | — | — |

## Claims

| Token | Claim | Checked by | Description |
|-------|-------|------------|-------------|
| Resource token | `aud` | Person Server | — |
| Resource token | `agent_jkt` | Person Server, Access Server | Must match the key that signed the request \| thumbprint |
| Auth token | `iss` | Resource | — |
