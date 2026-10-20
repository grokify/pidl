# PRD — Walkthrough format and embeddable viewer

**Initiative:** `INIT-PIDL-007`
**Repository:** `github.com/grokify/pidl`

## Problem

PIDL has two formats today, but only one is documented. People write PIDL; `pidl generate -f walkthrough-json` turns it into a walkthrough, with `extends` resolved, fragments expanded, cURL commands rendered, code samples resolved and findings attached. The viewer reads only the walkthrough. Yet the walkthrough has no version, no schema, and its TypeScript types are maintained by hand, so the Go and TypeScript sides can drift, and nobody outside this repository can rely on it.

The viewer itself ships only one way: `pidl generate -f viewer` writes a self-contained HTML file with the walkthrough embedded. That suits offline use, but a docs site that wants to show a walkthrough has to publish a separate 250 KB page per protocol, regenerate it whenever the viewer changes, and link out instead of embedding.

## Goals

- **An official walkthrough format** (`pidl-walkthrough/v1`): versioned, described by a JSON Schema generated from the Go types, with TypeScript types generated from that schema, and a written compatibility rule. Any tool can produce or read it.
- **An embeddable viewer** published to npm and served by CDNs: one script tag shows a walkthrough on any page, loaded from inline JSON, a URL attribute or a `?src=` query parameter.
- **One viewer, three deliveries:** the self-contained HTML file (offline, local files), the CDN script (docs sites), and later a hosted app. All read the same format.

## Non-Goals

- Reading PIDL source in the browser. Resolution, rendering and sample tooling stay in Go; the browser reads walkthroughs only.
- The hosted (cloud) app. It is a later initiative that builds on the format and the embeddable viewer.
- Publishing to npm from CI. The first releases are published by a maintainer.

## Requirements

1. **Format identifier.** Every walkthrough carries `"format": "pidl-walkthrough/v1"` as its first field. The viewer refuses a different major version with a message naming the version it found; a walkthrough with no `format` (written before v1) is read as v1.
2. **Generated schema.** `schema/walkthrough.schema.json` is generated from `pidl.Walkthrough` and embedded as `schema.WalkthroughSchemaJSON`. Property names are snake_case, matching PIDL. A check fails when the committed schema differs from the generated one.
3. **Generated TypeScript types.** `viewer/src/walkthrough.gen.ts` is generated from the schema; the viewer imports its types from there. CI fails when it is out of date.
4. **File naming.** Walkthroughs use the `.walkthrough.json` extension; the guides and examples use it.
5. **Compatibility rule.** Within v1, changes only add optional fields; readers ignore fields they do not know. Removing or renaming a field, changing its type or meaning, or making a field required is v2.
6. **Embeddable build.** `npm run build:lib` writes `dist/pidl-viewer.js` (IIFE, global `PidlViewer`) and `dist/pidl-viewer.mjs` (ES module), each with its CSS included and scoped to the viewer's root element.
7. **Mounting.** `PidlViewer.mount(element, { data | src })` renders a walkthrough into an element and returns a handle with `unmount()`. Several viewers can share a page.
8. **Auto-mount.** A script tag with `data-pidl-target` mounts on load, reading in this order: an inline `<script type="application/json">` named by `data-pidl-data`, the URL in `data-pidl-src`, then the page's `?src=` parameter.
9. **Relative URLs.** Paths inside a walkthrough loaded from a URL resolve against that URL, not the page.
10. **Untrusted input.** A walkthrough is data. The viewer renders no HTML from it, opens only `http:` and `https:` links, and `?src=` loads only `https:` URLs, or same-origin ones.
11. **Package.** `@grokify/pidl-viewer` on npm with the two builds, the generated types and the schema, versioned with pidl.
12. **Guide.** A page on embedding the viewer, with CDN snippets for inline data, a URL and `?src=`, and the CORS requirement.

## Success Criteria

- `pidl generate -f walkthrough-json` output validates against the embedded schema for every built-in example.
- A plain HTML page with one script tag from jsDelivr shows the AAuth walkthrough from a GitHub Pages URL.
- The self-contained viewer HTML and the CDN script render the same walkthrough identically.
