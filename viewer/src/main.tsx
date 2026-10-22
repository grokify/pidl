import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { loadPage } from "./data";
import { fromOtherOrigin, PAGE_HOST } from "./host";
import "./page.css";
import "./styles.css";

// The standalone viewer page: `pidl generate -f viewer`, or the dev server.
const root = createRoot(document.getElementById("root")!);

loadPage()
  .then(({ walkthrough: w, base }) => {
    document.title = `${w.name} walkthrough`;
    root.render(
      <StrictMode>
        <App walkthrough={w} host={{ ...PAGE_HOST, source: fromOtherOrigin(base, window.location.href) }} />
      </StrictMode>,
    );
  })
  .catch((err: unknown) => {
    root.render(
      <div className="pidl-viewer">
        <p className="load-error">Could not load the walkthrough: {String(err)}</p>
      </div>,
    );
  });
