# TRD — Postman collections from PIDL

**Initiative:** `INIT-PIDL-006`
**Repository:** `github.com/grokify/pidl`

## Renderer

`render/postman.go` builds Postman Collection v2.1 types (info, item folders and requests, event scripts, variables, examples) from the walkthrough model, so ordering, fragment expansion and examples follow the same rules as the viewer. `encoding/json` writes them with stable key order.

## Variables

- One base URL variable per role (`<role>_url`), derived from the scheme and host of the role's example URLs.
- One variable per chained value. A value is chained when a later request's example contains a string that an earlier response's example carries (a token, a `Location`, a code). The response's test script reads it from the same place (a JSON pointer into the body, or a header) and later requests use `{{name}}`. Names come from token definition IDs, or from the JSON member or header name.

## Scripts

Test scripts are generated, not templated by hand: `pm.response.to.have.status`, header presence, body member presence via JSON pointers, and token checks that split the JWT, base64url-decode the payload and assert claim names. Polling uses `pm.execution.setNextRequest` with a counter variable and a retry limit.

## Signer profiles

A signer profile names how a helper receives the target and presented token. The first profile, `aauth-sandbox`, prefixes the URL with the helper's address and sets `X-Sandbox-Present` and `X-Sandbox-Token` from the example's `Signature-Key` token.

## Testing

Golden collections for the AAuth example (signer and mock modes); schema validation against the v2.1 schema; a Newman run against the sandbox in the aauth-postman repository's CI.
