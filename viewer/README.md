# @grokify/pidl-viewer

A step-by-step viewer for [PIDL](https://github.com/grokify/pidl) protocol walkthroughs: pick a scenario, step through its messages on a sequence diagram, and read each request and response as HTTP, cURL or code.

It reads walkthroughs (`pidl-walkthrough/v1`), which `pidl generate -f walkthrough-json` produces from a PIDL file.

## Use It from a CDN

```html
<div id="walkthrough"></div>
<script src="https://cdn.jsdelivr.net/npm/@grokify/pidl-viewer@0.10.0/dist-lib/pidl-viewer.js"
        data-pidl-target="#walkthrough"
        data-pidl-src="aauth.walkthrough.json"></script>
```

## Use It from JavaScript

```js
import { mount } from "@grokify/pidl-viewer";

const viewer = await mount(document.querySelector("#walkthrough"), { src: "aauth.walkthrough.json" });
```

The viewer renders inside a shadow root, so page and viewer styles stay apart, and several viewers can share a page.

See [Embedding the Viewer](https://github.com/grokify/pidl/blob/main/docs/guides/embedding.md) for every option, and the [walkthrough format](https://github.com/grokify/pidl/blob/main/docs/reference/walkthrough-format.md) reference.

## Observed Runs

A walkthrough generated with `pidl generate --observation` has an entry for each run observed in a real system. On such an entry the viewer shows a **Concept / Observed** switch. Observed, the default, shows each step's observed exchange in the request, response and token panels, the run's summary in the header, when each exchange happened and the gap from the one before, and a mark on weak matches. Exchanges the document does not describe appear between the steps they fell between, dotted; steps the run did not perform, or that a server's capture cannot see, are dimmed and say why. Concept shows each step's example, as for any scenario.

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

## Release

The package is released with PIDL and shares its version: the `v0.10.0` tag publishes `@grokify/pidl-viewer@0.10.0`.

1. Set `version` in `package.json` to the release version, and run `npm install` to update `package-lock.json`.
2. Push, and wait for CI to pass.
3. Push the `vX.Y.Z` tag. The Viewer Release workflow checks that the tag matches `package.json`, runs the tests, builds `dist-lib/` and publishes to npm with provenance.
4. Check that the version is on the CDN: `curl -sI https://cdn.jsdelivr.net/npm/@grokify/pidl-viewer@X.Y.Z/dist-lib/pidl-viewer.js` answers `200`.

The workflow authenticates with npm trusted publishing, so the repository stores no npm token. npm sets up a trusted publisher only for a package that exists, so the first version is published by hand, by a maintainer of the `@grokify` scope:

```bash
npm login
npm publish --access public   # runs the tests and build:lib first
```

Then, in the package's settings on npmjs.com, add a trusted publisher: GitHub Actions, repository `grokify/pidl`, workflow `viewer-release.yaml`. The workflow skips a version that is already on npm, so the tag for a version published by hand does not fail.

## License

MIT
