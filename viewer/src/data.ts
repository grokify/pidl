import type { Walkthrough } from "./types";
import { srcFromQuery } from "./urls";

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

// parseWalkthrough checks a walkthrough's format and the fields the viewer
// relies on, so a wrong file fails with a clear message. It takes JSON text
// or an already parsed value.
export function parseWalkthrough(input: unknown): Walkthrough {
  const data = (typeof input === "string" ? JSON.parse(input) : input) as Partial<Walkthrough> | null;
  if (data === null || typeof data !== "object" || Array.isArray(data)) {
    throw new Error("Not a PIDL walkthrough: expected a JSON object");
  }
  checkFormat(data.format);
  if (!Array.isArray(data.roles) || !Array.isArray(data.scenarios)) {
    throw new Error("Not a PIDL walkthrough: expected roles and scenarios");
  }
  // Walkthroughs written before v1 have no format; any v1 revision reads as v1.
  return { ...data, format: WALKTHROUGH_FORMAT } as Walkthrough;
}

// Loaded is a walkthrough and the URL its relative paths resolve against.
export interface Loaded {
  walkthrough: Walkthrough;
  base: string;
}

export interface LoadOptions {
  // data is a walkthrough, parsed or as JSON text.
  data?: unknown;
  // src is the URL of a .walkthrough.json, resolved against the page.
  src?: string;
  // base resolves relative paths in the walkthrough (default: src, or
  // the page for inline data).
  base?: string;
}

// load reads a walkthrough from inline data or a URL. Fetches send no
// cookies or credentials, so a URL cannot act as the reader.
export async function load(opts: LoadOptions, pageUrl: string, fetchFn: typeof fetch = fetch): Promise<Loaded> {
  if (opts.data !== undefined) {
    return { walkthrough: parseWalkthrough(opts.data), base: new URL(opts.base ?? pageUrl, pageUrl).href };
  }
  if (!opts.src) {
    throw new Error("No walkthrough: give data or src");
  }
  const url = new URL(opts.src, pageUrl).href;
  const res = await fetchFn(url, { credentials: "omit", cache: "no-cache" });
  if (!res.ok) {
    throw new Error(`${url} returned ${res.status}`);
  }
  return { walkthrough: parseWalkthrough(await res.text()), base: new URL(opts.base ?? url, pageUrl).href };
}

// loadPage reads the walkthrough for the standalone viewer page: the one
// embedded by `pidl generate -f viewer`, else the page's ?src= URL, else,
// during development, ./walkthrough.json.
export async function loadPage(doc: Document = document, fetchFn: typeof fetch = fetch): Promise<Loaded> {
  const pageUrl = doc.location?.href ?? doc.baseURI;
  const embedded = doc.getElementById("pidl-walkthrough")?.textContent?.trim();
  if (embedded) {
    return load({ data: embedded }, pageUrl, fetchFn);
  }
  return load({ src: srcFromQuery(pageUrl) ?? "./walkthrough.json" }, pageUrl, fetchFn);
}
