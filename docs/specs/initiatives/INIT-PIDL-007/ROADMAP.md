# ROADMAP — Walkthrough format and embeddable viewer

**Initiative:** `INIT-PIDL-007`
**Repository:** `github.com/grokify/pidl`

PIDL's processed output becomes an official format, `pidl-walkthrough/v1`, and the viewer becomes a script any page can load from a CDN. Commits implementing an item carry `Refs: RMI-PIDL-<NNN>`.

## Phase 1 — Walkthrough Format

**Theme:** A versioned, schema-described format that any tool can produce or read

- [ ] `RMI-PIDL-106` `format: "pidl-walkthrough/v1"` in every walkthrough, and the viewer's version check
- [ ] `RMI-PIDL-107` `schema/walkthrough.schema.json` generated from the Go types in a nested module, embedded, with a drift test
  - Depends on: `RMI-PIDL-106`
- [ ] `RMI-PIDL-108` TypeScript types generated from the schema, with a CI check
  - Depends on: `RMI-PIDL-107`
- [ ] `RMI-PIDL-109` Format reference: fields, the `.walkthrough.json` extension and the compatibility rule
  - Depends on: `RMI-PIDL-107`

## Phase 2 — Embeddable Viewer

**Theme:** One script tag shows a walkthrough on any page

- [ ] `RMI-PIDL-110` Loader: inline data, URL and `?src=`, base URL for relative paths, URL safety
  - Depends on: `RMI-PIDL-106`
- [ ] `RMI-PIDL-111` `PidlViewer.mount` and the library build (IIFE and ES module) with scoped CSS
  - Depends on: `RMI-PIDL-110`
- [ ] `RMI-PIDL-112` Auto-mount from the script tag, and hash navigation only for a sole auto-mounted viewer
  - Depends on: `RMI-PIDL-111`
- [ ] `RMI-PIDL-113` `@grokify/pidl-viewer` package and the embedding guide
  - Depends on: `RMI-PIDL-108`, `RMI-PIDL-112`
