import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { loadWalkthrough } from "./data";
import "./styles.css";

const root = createRoot(document.getElementById("root")!);

loadWalkthrough()
  .then((w) => {
    document.title = `${w.name} walkthrough`;
    root.render(
      <StrictMode>
        <App walkthrough={w} />
      </StrictMode>,
    );
  })
  .catch((err: unknown) => {
    root.render(<p className="load-error">Could not load the walkthrough: {String(err)}</p>);
  });
