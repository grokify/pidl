import type { Walkthrough } from "./types";

// loadWalkthrough reads the walkthrough embedded by `pidl generate -f viewer`,
// or, during development, fetches public/walkthrough.json.
export async function loadWalkthrough(doc: Document = document): Promise<Walkthrough> {
  const embedded = doc.getElementById("pidl-walkthrough")?.textContent?.trim();
  if (embedded) {
    return parseWalkthrough(embedded);
  }
  const res = await fetch("./walkthrough.json", { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`No embedded walkthrough and ./walkthrough.json returned ${res.status}`);
  }
  return parseWalkthrough(await res.text());
}

// parseWalkthrough parses walkthrough JSON and checks the fields the viewer
// relies on, so a wrong file fails with a clear message.
export function parseWalkthrough(text: string): Walkthrough {
  const data = JSON.parse(text) as Partial<Walkthrough>;
  if (!Array.isArray(data.roles) || !Array.isArray(data.scenarios)) {
    throw new Error("Not a PIDL walkthrough: expected roles and scenarios");
  }
  return data as Walkthrough;
}
