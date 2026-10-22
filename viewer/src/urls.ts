// URL rules for walkthroughs loaded from elsewhere. A walkthrough is
// data from any source, so the viewer loads and links only what these
// allow.

// allowedSrc reports whether the viewer may load a walkthrough from url,
// given in a page's ?src= parameter: https anywhere, or any scheme on the
// page's own origin, so local development over http works.
export function allowedSrc(url: string, pageUrl: string): boolean {
  let u: URL;
  let page: URL;
  try {
    page = new URL(pageUrl);
    u = new URL(url, page);
  } catch {
    return false;
  }
  if (u.username || u.password) {
    return false;
  }
  return u.protocol === "https:" || (u.origin === page.origin && u.origin !== "null");
}

// srcFromQuery returns the walkthrough URL in a page's ?src= parameter,
// resolved against the page, or null when there is none. It throws when
// the URL is not allowed.
export function srcFromQuery(pageUrl: string): string | null {
  const page = new URL(pageUrl);
  const src = page.searchParams.get("src");
  if (!src) {
    return null;
  }
  if (!allowedSrc(src, pageUrl)) {
    throw new Error(`Walkthroughs load only from https: URLs or this page's origin, not ${src}`);
  }
  return new URL(src, page).href;
}

// resolve returns ref resolved against base, the walkthrough's own URL, so
// paths inside a walkthrough work wherever it is hosted.
export function resolve(ref: string, base: string): string {
  return new URL(ref, base).href;
}

// safeHref returns ref resolved against base when it is an http: or https:
// link, and undefined for anything else, such as javascript: or data:.
export function safeHref(ref: string, base: string): string | undefined {
  try {
    const u = new URL(ref, base);
    return u.protocol === "https:" || u.protocol === "http:" ? u.href : undefined;
  } catch {
    return undefined;
  }
}
