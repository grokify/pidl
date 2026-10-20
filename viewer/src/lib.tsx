// The embeddable viewer, published as @grokify/pidl-viewer and served by
// CDNs: PidlViewer.mount renders a walkthrough into any element.
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { load, type LoadOptions, WALKTHROUGH_FORMAT } from "./data";
import { fromOtherOrigin } from "./host";
import css from "./styles.css?inline";
import type { Walkthrough } from "./types";

export { WALKTHROUGH_FORMAT };
export type { Walkthrough };

export interface MountOptions extends LoadOptions {
  // hash keeps the scenario and step in the page's URL hash. Use it only
  // when the viewer is the one walkthrough on its page.
  hash?: boolean;
}

export interface Viewer {
  walkthrough: Walkthrough;
  unmount(): void;
}

// mount renders a walkthrough into el, inside a shadow root so the page's
// styles and the viewer's do not affect each other. The element sizes the
// viewer: 720px tall unless the page sets a height. It resolves once the
// walkthrough has loaded, and rejects, after showing the error in el, when
// it cannot be loaded.
export async function mount(el: Element, opts: MountOptions): Promise<Viewer> {
  const shadow = el.shadowRoot ?? el.attachShadow({ mode: "open" });
  shadow.replaceChildren();
  const style = document.createElement("style");
  style.textContent = css;
  const container = document.createElement("div");
  container.style.height = "100%";
  shadow.append(style, container);
  const root = createRoot(container);
  const unmount = () => {
    root.unmount();
    shadow.replaceChildren();
  };

  const pageUrl = window.location.href;
  try {
    const { walkthrough, base } = await load(opts, pageUrl);
    root.render(
      <StrictMode>
        <App
          walkthrough={walkthrough}
          host={{ hash: opts.hash ?? false, keys: "element", page: false, source: fromOtherOrigin(base, pageUrl) }}
        />
      </StrictMode>,
    );
    return { walkthrough, unmount };
  } catch (err) {
    root.render(
      <div className="pidl-viewer">
        <p className="load-error">Could not load the walkthrough: {String(err)}</p>
      </div>,
    );
    throw err;
  }
}
