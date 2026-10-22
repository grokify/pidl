# PLAN — Postman collections from PIDL

**Initiative:** `INIT-PIDL-006`

1. **Collection structure:** folders, requests, base URL variables and documentation; mock mode can ship with this phase since it needs no signer.
2. **Chaining and checks:** captured variables, response checks and polling.
3. **Signer and environments:** the `aauth-sandbox` signer profile and environment files, once the sandbox's signing helper exists (`RMI-AAUTHGO-025`).
4. **AAuth collection:** generated into `aauth-dev/aauth-postman` with Newman CI against the sandbox image (`RMI-AAUTHGO-027`).
