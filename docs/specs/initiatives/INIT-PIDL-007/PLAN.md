# PLAN — Walkthrough format and embeddable viewer

**Initiative:** `INIT-PIDL-007`
**Repository:** `github.com/grokify/pidl`

## Entry Gate

The viewer's representations and code samples (`feat/viewer-representations`) delivered.

## Build Order

1. **Format:** the identifier, the generated schema and the generated TypeScript types, then the format reference page.
2. **Embeddable viewer:** the loader with sources, version check and URL safety; mount and the library build; auto-mount; then the package and the embedding guide.

The format comes first so the CDN script, the first reader outside this repository, starts on a versioned contract.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Viewer CSS clashes with a host page | Scope every rule under `.pidl-viewer` |
| `?src=` used to show misleading content under a trusted docs domain | `https:` or same-origin only; no HTML rendering; the page shows the source URL |
| Schema and Go types drift | Generated schema with a drift test; generated TypeScript types with a CI diff |
| A dependency in the root module | Generator in a nested module |

## Follow-On

A hosted app (`INIT-PIDL-010`): upload or link a walkthrough or PIDL file, render it server-side, share a link. It reuses the format and the embeddable viewer.
