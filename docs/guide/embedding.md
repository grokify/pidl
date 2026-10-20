# Embedding the Viewer

The walkthrough viewer comes in two forms that read the same [walkthrough format](../reference/walkthrough-format.md):

| Form | Use it for |
|------|------------|
| `pidl generate -f viewer` | one self-contained HTML file, for offline use, email or opening from disk ([Walkthrough Viewer](viewer.md)) |
| `@grokify/pidl-viewer` | a script any page loads from a CDN, to show one or more walkthroughs inside docs, blogs or apps |

## Publish a Walkthrough

Generate the walkthrough and host it next to your pages, or anywhere that serves it over HTTPS:

```bash
pidl generate -f walkthrough-json -o aauth.walkthrough.json aauth_protocol
```

Generate it in CI from the PIDL source rather than committing it, so it never falls behind the protocol.

## One Script Tag

```html
<div id="walkthrough"></div>
<script src="https://cdn.jsdelivr.net/npm/@grokify/pidl-viewer@0.9.0/dist-lib/pidl-viewer.js"
        data-pidl-target="#walkthrough"
        data-pidl-src="aauth.walkthrough.json"></script>
```

The script mounts the viewer on the element named by `data-pidl-target` once the page has loaded. It reads the walkthrough from the first of these that is present:

| Attribute | Walkthrough |
|-----------|-------------|
| `data-pidl-data="<id>"` | the JSON inside `<script type="application/json" id="<id>">` on the page |
| `data-pidl-src="<url>"` | the URL, relative to the page |
| *(neither)* | the page's `?src=` query parameter |

The same file is on unpkg at `https://unpkg.com/@grokify/pidl-viewer@0.9.0/dist-lib/pidl-viewer.js`. Always name an exact version, so a new release cannot change your pages.

### Inline Data

```html
<div id="walkthrough"></div>
<script type="application/json" id="aauth-data">
  { "format": "pidl-walkthrough/v1", "id": "aauth-protocol", ... }
</script>
<script src="https://cdn.jsdelivr.net/npm/@grokify/pidl-viewer@0.9.0/dist-lib/pidl-viewer.js"
        data-pidl-target="#walkthrough"
        data-pidl-data="aauth-data"></script>
```

### One Page for Any Walkthrough

A page with a script tag that names no walkthrough shows the one in its `?src=` parameter, so one page can show any walkthrough:

```text
https://example.com/viewer.html?src=https://example.org/aauth.walkthrough.json
```

`?src=` loads only `https:` URLs, or URLs on the page's own origin. The top bar names the host a walkthrough came from when it is not the page's own, so readers can tell whose content they are reading.

## From JavaScript

`PidlViewer.mount` renders a walkthrough into an element and returns a promise for the viewer:

```html
<div id="a"></div>
<div id="b"></div>
<script src="https://cdn.jsdelivr.net/npm/@grokify/pidl-viewer@0.9.0/dist-lib/pidl-viewer.js"></script>
<script>
  PidlViewer.mount(document.getElementById("a"), { src: "aauth.walkthrough.json" });
  PidlViewer.mount(document.getElementById("b"), { data: myWalkthrough });
</script>
```

As an ES module, with types, from npm or a CDN:

```js
import { mount } from "https://cdn.jsdelivr.net/npm/@grokify/pidl-viewer@0.9.0/dist-lib/pidl-viewer.mjs";

const viewer = await mount(document.querySelector("#walkthrough"), { src: "aauth.walkthrough.json" });
// later: viewer.unmount();
```

| Option | Meaning |
|--------|---------|
| `data` | a walkthrough, parsed or as JSON text |
| `src` | the URL of a walkthrough, relative to the page |
| `base` | the URL relative paths inside the walkthrough resolve against (default: `src`) |
| `hash` | keep the scenario and step in the page's URL hash (default: false) |

`mount` resolves once the walkthrough is shown, and rejects, after showing the error in the element, when it cannot load. The package also exports the walkthrough types and `WALKTHROUGH_FORMAT`, and ships the schema as `@grokify/pidl-viewer/walkthrough.schema.json`.

## On the Page

- **Size.** The viewer fills its element, which is 720px tall unless your page sets a height, for example `#walkthrough { height: 80vh; }`. It is laid out for widths of about 1000px and up.
- **Styles.** The viewer renders inside a shadow root, so your page's styles do not change it and its styles do not leak into your page.
- **Keys.** An embedded viewer takes the step keys (→, ←, Home, End, P) only while it has focus; click it first. The standalone page takes them anywhere.
- **Position.** An auto-mounted viewer that is the only one on its page keeps the scenario and step in the URL hash, so links open at a step. Otherwise each viewer keeps its position to itself, unless you pass `hash: true`.
- **Theme and presenting.** Theme and Present work as on the [standalone page](viewer.md); Present fills the browser window.

## Hosting the Walkthrough

- **CORS.** A walkthrough on another origin must be served with `Access-Control-Allow-Origin`. GitHub Pages and `raw.githubusercontent.com` send it; for your own server, allow at least the origins of the pages that embed it.
- **Credentials.** The viewer fetches walkthroughs without cookies or other credentials, so a walkthrough must be readable without signing in.
- **Content.** The viewer treats a walkthrough as data: it shows its text as text, never as HTML, and opens only `http:` and `https:` links from it.
