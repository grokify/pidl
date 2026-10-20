// Host describes where a viewer runs: alone on its page, or embedded in
// someone else's page next to other content and possibly other viewers.
export interface Host {
  // hash keeps the scenario and step in the page's URL hash, so a link or
  // reload returns to the same step.
  hash: boolean;
  // keys is "page" to take step keys anywhere on the page, or "element"
  // to take them only while the viewer has focus.
  keys: "page" | "element";
  // page also applies the theme to the page itself.
  page: boolean;
  // source is the walkthrough's URL when it came from another origin,
  // shown so readers can tell whose content they are looking at.
  source?: string;
}

export const PAGE_HOST: Host = { hash: true, keys: "page", page: true };

// fromOtherOrigin returns the host name of url when it differs from the
// page's origin, and undefined otherwise.
export function fromOtherOrigin(url: string, pageUrl: string): string | undefined {
  try {
    const u = new URL(url, pageUrl);
    return u.origin === new URL(pageUrl).origin ? undefined : u.host;
  } catch {
    return undefined;
  }
}

// isTyping reports whether a key event comes from a form control, which
// keeps its own keys. It looks through shadow roots to the real target.
export function isTyping(e: KeyboardEvent): boolean {
  const t = (e.composedPath()[0] ?? e.target) as HTMLElement | null;
  return !!t && (t.tagName === "SELECT" || t.tagName === "INPUT" || t.tagName === "TEXTAREA");
}
