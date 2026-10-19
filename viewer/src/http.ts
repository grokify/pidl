import type { Header } from "./types";

const REASONS: Record<number, string> = {
  200: "OK", 201: "Created", 202: "Accepted", 204: "No Content", 400: "Bad Request", 401: "Unauthorized",
  402: "Payment Required", 403: "Forbidden", 404: "Not Found", 409: "Conflict", 410: "Gone", 429: "Too Many Requests",
  500: "Internal Server Error", 503: "Service Unavailable",
};

export function statusLine(status: number): string {
  return `HTTP/1.1 ${status} ${REASONS[status] ?? ""}`.trimEnd();
}

// splitURL returns the request target (path and query) and the host.
export function splitURL(url: string): { target: string; host: string } {
  try {
    const u = new URL(url);
    return { target: u.pathname + u.search, host: u.host };
  } catch {
    return { target: url, host: "" };
  }
}

// coveredComponents reads the component list from a Signature-Input header,
// e.g. sig=("@method" "@authority" "@path" "signature-key");created=1.
export function coveredComponents(headers: Header[] = []): Set<string> {
  const input = headers.find((h) => h.name.toLowerCase() === "signature-input")?.value ?? "";
  const list = /\(([^)]*)\)/.exec(input)?.[1] ?? "";
  return new Set([...list.matchAll(/"([^"]+)"/g)].map((m) => m[1].toLowerCase()));
}
