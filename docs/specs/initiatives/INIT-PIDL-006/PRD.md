# PRD — Postman collections from PIDL

**Initiative:** `INIT-PIDL-006`
**Repository:** `github.com/grokify/pidl`

## Problem

A PIDL walkthrough shows a protocol's messages; developers also want to send them. Postman is where many API developers try and test HTTP flows, and a collection that walks a protocol's scenarios step by step, with values passed from one response to the next request, turns the walkthrough into something runnable. Writing such a collection by hand repeats everything the PIDL file already says (order, URLs, headers, bodies, tokens) and drifts as the protocol changes.

## Goals

- `pidl generate -f postman` writes a Postman collection (format v2.1) from a PIDL file: one folder per scenario, in output order, with one request per step the client role sends.
- Requests chain: values a response delivers (tokens, `Location` URLs) are captured into variables the later requests use.
- Each request checks its response against the example: status, required headers, body members, and the claims of tokens it carries.
- Signed protocols run through a signing helper, such as the AAuth sandbox's (`INIT-AAUTHGO-002`), so Postman needs no cryptography.
- A mock mode turns example responses into saved Postman examples for a Postman mock server, for exploring without running anything.

## Non-Goals

- Signing inside Postman scripts.
- Playing server roles in Postman. Server-to-server steps are described in folder documentation, not sent.
- Publishing to Postman workspaces from PIDL. Publication belongs to the repository that holds a protocol's collection.

## Requirements

1. **Structure.** A collection per PIDL file with the protocol's name and description; a folder per group, then per scenario, in output order; fragments expanded in place. Folder descriptions carry the scenario description and narration, and list the steps between other roles that the collection does not send.
2. **Client role.** `--client <role>` (default: the first entity of type `agent` or `client`) selects the role whose requests become Postman requests.
3. **Requests.** Method, URL, headers and body from each step's example request, with each role's base URL replaced by a variable (`{{ps_url}}`). Signature headers are left out when a signer is configured; they are added by the signer.
4. **Chaining.** For tokens a step issues (`security.issues`) and values in example responses that later requests reuse (matched by value: tokens, `Location`, pending URLs, codes), the response's test script stores them in collection variables, and later requests refer to the variables.
5. **Checks.** Test scripts assert the status code; that headers named in the example are present; that body members in the example are present; and, for tokens in the example, that the token decodes and carries the example's claim names. Values that change per run are not compared.
6. **Deferred responses.** A `202` with `Location` is followed by a polling request that repeats until the final status or a retry limit, using Postman's request flow controls.
7. **Signer.** `--signer <url>` sends requests through a signing helper by prefixing the target URL; the token a step presents (from the example's `Signature-Key` token) is passed in the helper's presentation headers, named by `--signer-profile aauth-sandbox`.
8. **Person steps.** Steps where a person acts in a browser are described, not sent; `--approve-url` adds a request to a headless approval endpoint for automated runs.
9. **Mock mode.** `--mock` saves each step's example response as a Postman example on its request and points base URLs at `{{mock_url}}`; no signer, no checks beyond status.
10. **Environments.** `generate -f postman-environment` writes environment files for named targets from a small targets file (role base URLs, signer URL), for example sandbox, Docker, minikube and mock.
11. **Validation.** The generated collection validates against Postman's published v2.1 collection schema.

## Success Criteria

- The AAuth collection runs green in Newman against the AAuth sandbox for all five access modes.
- The mock-mode collection, imported into Postman with a mock server, steps through every scenario.
- Changing an example in the PIDL file changes only the affected requests in the generated collection.

## Decisions

- **Generated, never hand-edited.** The PIDL file is the source of truth; protocol repositories commit the generated collection.
- **Signing stays out of Postman.** A signing helper is simpler and works the same in Postman, Newman and curl.
- **The AAuth collection lives in `aauth-dev/aauth-postman`,** with its environments, workshop guide and a Newman CI job against the sandbox image.
