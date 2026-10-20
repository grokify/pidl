# @grokify/pidl-viewer

A step-by-step viewer for [PIDL](https://github.com/grokify/pidl) protocol walkthroughs: pick a scenario, step through its messages on a sequence diagram, and read each request and response as HTTP, cURL or code.

It reads walkthroughs (`pidl-walkthrough/v1`), which `pidl generate -f walkthrough-json` produces from a PIDL file.

## Use It from a CDN

```html
<div id="walkthrough"></div>
<script src="https://cdn.jsdelivr.net/npm/@grokify/pidl-viewer@0.9.0/dist-lib/pidl-viewer.js"
        data-pidl-target="#walkthrough"
        data-pidl-src="aauth.walkthrough.json"></script>
```

## Use It from JavaScript

```js
import { mount } from "@grokify/pidl-viewer";

const viewer = await mount(document.querySelector("#walkthrough"), { src: "aauth.walkthrough.json" });
```

The viewer renders inside a shadow root, so page and viewer styles stay apart, and several viewers can share a page.

See [Embedding the Viewer](https://github.com/grokify/pidl/blob/main/docs/guide/embedding.md) for every option, and the [walkthrough format](https://github.com/grokify/pidl/blob/main/docs/reference/walkthrough-format.md) reference.

## Develop

```bash
npm ci
npm run data         # public/walkthrough.json from the AAuth example
npm run dev          # the standalone viewer at http://localhost:5173
npm test
npm run types        # src/walkthrough.gen.ts from the walkthrough schema
npm run build:lib    # dist-lib/: the embeddable viewer, types and schema
npm run build:embed  # render/viewer/viewer.html for pidl generate -f viewer
```

## License

MIT
