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

// WALKTHROUGH_FORMAT is the format this viewer reads. It reads any
// walkthrough of the same major version, since within a major version
// changes only add optional fields.
export const WALKTHROUGH_FORMAT = "pidl-walkthrough/v1";

const formatPattern = /^pidl-walkthrough\/v(\d+)(?:\.\d+)*$/;

// checkFormat accepts the viewer's major version, or no format at all
// (walkthroughs written before v1), and refuses anything else.
export function checkFormat(format: unknown): void {
  if (format === undefined) {
    return;
  }
  const want = formatPattern.exec(WALKTHROUGH_FORMAT)![1];
  const got = typeof format === "string" ? formatPattern.exec(format) : null;
  if (!got) {
    throw new Error(`Not a PIDL walkthrough: format is ${JSON.stringify(format)}`);
  }
  if (got[1] !== want) {
    throw new Error(`This viewer reads ${WALKTHROUGH_FORMAT}; the walkthrough is ${format as string}`);
  }
}

// parseWalkthrough parses walkthrough JSON and checks its format and the
// fields the viewer relies on, so a wrong file fails with a clear message.
export function parseWalkthrough(text: string): Walkthrough {
  const data = JSON.parse(text) as Partial<Walkthrough>;
  checkFormat(data.format);
  if (!Array.isArray(data.roles) || !Array.isArray(data.scenarios)) {
    throw new Error("Not a PIDL walkthrough: expected roles and scenarios");
  }
  // Walkthroughs written before v1 have no format; any v1 revision reads as v1.
  data.format = WALKTHROUGH_FORMAT;
  return data as Walkthrough;
}
