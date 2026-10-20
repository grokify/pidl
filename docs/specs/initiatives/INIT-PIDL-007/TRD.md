# TRD — Walkthrough format and embeddable viewer

**Initiative:** `INIT-PIDL-007`
**Repository:** `github.com/grokify/pidl`

## Format

```json
{
  "format": "pidl-walkthrough/v1",
  "id": "aauth-protocol",
  "name": "AAuth",
  "roles": [ ... ],
  "scenarios": [ ... ]
}
```

`pidl.WalkthroughFormat` holds the identifier, and `Protocol.Walkthrough` sets it, so every producer writes it. Readers compare the major version only: the text after `/v`, up to the first `.`.

## Schema Generation

The root module has no dependencies and keeps none. The generator lives in a nested module, `schema/gen` (`go.mod` with `replace github.com/grokify/pidl => ../..`, which stays inside the repository), using `github.com/invopop/jsonschema`:

- `go run ./schema/gen` (run from `schema/gen`) reflects `pidl.Walkthrough`, sets `$id` to `https://github.com/grokify/pidl/schema/walkthrough.schema.json` and the title, and writes `schema/walkthrough.schema.json`.
- `format` is a `const` of the identifier, so the schema documents the version it describes.
- Fields without `omitempty` are required, matching what producers always write.
- A test in `schema/gen` regenerates the schema in memory and fails when it differs from the committed file; CI runs it.
- `schemakit lint --property-case snake_case` passes on the generated schema.

The root module embeds the file as `schema.WalkthroughSchemaJSON`. A root test checks that every built-in example's walkthrough has exactly the top-level fields the schema lists and that `format` matches the schema's `const`, without a validator dependency.

## TypeScript Types

`json-schema-to-typescript` (dev dependency) turns the schema into `viewer/src/walkthrough.gen.ts` (`npm run types`). `viewer/src/types.ts` re-exports the generated types under the names the viewer uses today, so components do not change. The viewer CI job regenerates the file and fails on a diff.

## Embeddable Build

A second Vite configuration, `vite.lib.config.ts`, builds `src/lib.tsx` in library mode:

- formats `iife` (name `PidlViewer`) and `es`; React bundled, so the page needs nothing else;
- CSS injected by the bundle into a `<style>` element once per page;
- every selector scoped under `.pidl-viewer`, the class the root element gets on mount.

```ts
export interface MountOptions {
  data?: Walkthrough | string;   // parsed or JSON text
  src?: string;                  // URL of a .walkthrough.json
  base?: string;                 // base for relative paths (default: src)
}
export function mount(el: Element, opts: MountOptions): Promise<{ unmount(): void }>;
```

Auto-mount runs when the bundle's own `<script>` (`document.currentScript`) has `data-pidl-target`.

## Loading and Safety

`loadWalkthrough` takes a source (inline element, URL) and returns the parsed walkthrough plus its base URL. It checks `format`, then the required fields. `?src=` accepts `https:` URLs, or same-origin URLs on any scheme so local development works. Fetches use `credentials: "omit"`. The viewer uses React text rendering only; any link taken from data passes `safeHref`, which allows `http:` and `https:` after resolving against the base.

## Hash Navigation

The viewer keeps navigation state in the page's hash today. When embedded, it uses the hash only when it is the sole viewer on the page and was auto-mounted; otherwise it keeps state in memory, so it does not fight the host page's routing.

## Testing

- Go: format constant, schema drift, examples against the schema's top-level fields.
- Viewer: loading from inline data, URL and `?src=`; version refusal; `safeHref`; relative base resolution; mount and unmount; two mounts on one page.
