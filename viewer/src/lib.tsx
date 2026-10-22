// The embeddable viewer, published as @grokify/pidl-viewer and served by
// CDNs: PidlViewer.mount renders a walkthrough into any element.
import { StrictMode } from "react";
import { flushSync } from "react-dom";
import { createRoot, type Root } from "react-dom/client";
import App from "./App";
import { load, type LoadOptions, WALKTHROUGH_FORMAT } from "./data";
import { fromOtherOrigin } from "./host";
import { srcFromQuery } from "./urls";
import css from "./styles.css?inline";
import type { Walkthrough } from "./types";

export { WALKTHROUGH_FORMAT };
export type { LoadOptions };
export type * from "./walkthrough.gen";

/** Options for {@link mount}: the walkthrough as `data` or `src`, and where it keeps its position. */
export interface MountOptions extends LoadOptions {
  /**
   * Keep the scenario and step in the page's URL hash, so a link or reload
   * returns to the same step. Use it only when the viewer is the one
   * walkthrough on its page. Default false.
   */
  hash?: boolean;
}

/** A mounted viewer. */
export interface Viewer {
  /** The walkthrough shown. */
  walkthrough: Walkthrough;
  /** Removes the viewer from its element. */
  unmount(): void;
}

/**
 * Renders a walkthrough into `el`, inside a shadow root so the page's
 * styles and the viewer's do not affect each other. The element sizes the
 * viewer: 720px tall unless the page sets a height. Resolves once the
 * walkthrough has loaded and is shown; rejects, after showing the error
 * in `el`, when it cannot be loaded.
 */
export async function mount(el: Element, opts: MountOptions): Promise<Viewer> {
  const { shadow, root } = prepare(el);
  const unmount = () => {
    root.unmount();
    shadow.replaceChildren();
  };

  const pageUrl = window.location.href;
  try {
    const { walkthrough, base } = await load(opts, pageUrl);
    // Render now, so the viewer is on the page when mount resolves.
    flushSync(() => root.render(
      <StrictMode>
        <App
          walkthrough={walkthrough}
          host={{ hash: opts.hash ?? false, keys: "element", page: false, source: fromOtherOrigin(base, pageUrl) }}
        />
      </StrictMode>,
    ));
    return { walkthrough, unmount };
  } catch (err) {
    showError(root, err);
    throw err;
  }
}

// prepare gives el a fresh shadow root holding the viewer's styles and a
// React root.
function prepare(el: Element) {
  const shadow = el.shadowRoot ?? el.attachShadow({ mode: "open" });
  shadow.replaceChildren();
  const style = document.createElement("style");
  style.textContent = css;
  const container = document.createElement("div");
  container.style.height = "100%";
  shadow.append(style, container);
  return { shadow, root: createRoot(container) };
}

function showError(root: Root, err: unknown) {
  root.render(
    <div className="pidl-viewer">
      <p className="load-error">Could not load the walkthrough: {String(err)}</p>
    </div>,
  );
}

/** What a script tag's `data-pidl-*` attributes ask for. */
export interface AutoMount {
  target: string;
  options: MountOptions;
}

/**
 * Reads a script tag's attributes: `data-pidl-target`, a selector for the
 * element to mount on, and the walkthrough from, in order, the inline JSON
 * element whose id is `data-pidl-data`, the URL in `data-pidl-src`, or the
 * page's `?src=` parameter. The position goes in the URL hash only when
 * this is the one auto-mounted viewer on the page. Returns null when the
 * tag has no `data-pidl-target`.
 */
export function autoMountOptions(script: Element, doc: Document): AutoMount | null {
  const target = script.getAttribute("data-pidl-target");
  if (!target) {
    return null;
  }
  const hash = doc.querySelectorAll("script[data-pidl-target]").length === 1;
  const dataId = script.getAttribute("data-pidl-data");
  if (dataId) {
    const text = doc.getElementById(dataId)?.textContent?.trim();
    if (!text) {
      throw new Error(`data-pidl-data names #${dataId}, which holds no walkthrough`);
    }
    return { target, options: { data: text, hash } };
  }
  const src = script.getAttribute("data-pidl-src") ?? srcFromQuery(doc.location.href);
  if (!src) {
    throw new Error("Give the walkthrough with data-pidl-data, data-pidl-src or ?src=");
  }
  return { target, options: { src, hash } };
}

// A classic script tag mounts itself; an ES module import does not, since
// it has no script element of its own.
const script = typeof document === "undefined" ? null : document.currentScript;
if (script?.hasAttribute("data-pidl-target")) {
  const run = () => {
    const target = script.getAttribute("data-pidl-target")!;
    const el = document.querySelector(target);
    if (!el) {
      console.error(`pidl-viewer: data-pidl-target ${target} matches no element`);
      return;
    }
    let auto: AutoMount | null;
    try {
      auto = autoMountOptions(script, document);
    } catch (err) {
      showError(prepare(el).root, err);
      console.error("pidl-viewer:", err);
      return;
    }
    if (auto) {
      mount(el, auto.options).catch((err: unknown) => console.error("pidl-viewer:", err));
    }
  };
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", run, { once: true });
  } else {
    run();
  }
}
