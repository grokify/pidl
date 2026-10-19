# ROADMAP — Postman collections from PIDL

**Initiative:** `INIT-PIDL-006`
**Repository:** `github.com/grokify/pidl`

Commits implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — Collection Structure

**Theme:** A runnable outline of every scenario

- [ ] `RMI-PIDL-096` `generate -f postman`: folders by group and scenario, requests for the client role, base URL variables and documentation
- [ ] `RMI-PIDL-097` Mock mode: example responses as saved Postman examples
  - Depends on: `RMI-PIDL-096`
- [ ] `RMI-PIDL-098` Validation against the Postman v2.1 collection schema and golden collections
  - Depends on: `RMI-PIDL-096`

## Phase 2 — Chaining and Checks

**Theme:** Each response feeds the next request

- [ ] `RMI-PIDL-099` Captured variables for tokens, locations and codes, matched by value across examples
  - Depends on: `RMI-PIDL-096`
- [ ] `RMI-PIDL-100` Response checks from examples: status, headers, body members and token claims
  - Depends on: `RMI-PIDL-096`
- [ ] `RMI-PIDL-101` Deferred responses: polling with a retry limit
  - Depends on: `RMI-PIDL-099`

## Phase 3 — Signer and Environments

**Theme:** Signed protocols without cryptography in Postman

- [ ] `RMI-PIDL-102` Signer profiles, starting with the AAuth sandbox's signing helper
  - Depends on: `RMI-PIDL-099`, `RMI-AAUTHGO-025`
- [ ] `RMI-PIDL-103` `generate -f postman-environment` for named targets, and headless approval steps
  - Depends on: `RMI-PIDL-102`

## Phase 4 — AAuth Collection

**Theme:** Publish and test

- [ ] `RMI-PIDL-104` AAuth collection and environments generated into `aauth-dev/aauth-postman`, with Newman CI against the sandbox image
  - Depends on: `RMI-PIDL-100`, `RMI-PIDL-101`, `RMI-PIDL-103`, `RMI-AAUTHGO-027`
- [ ] `RMI-PIDL-105` Postman guide and release notes
  - Depends on: `RMI-PIDL-104`
