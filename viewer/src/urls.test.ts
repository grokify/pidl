import { describe, expect, it } from "vitest";
import { allowedSrc, resolve, safeHref, srcFromQuery } from "./urls";

const page = "https://docs.example.com/guide/viewer.html";

describe("allowedSrc", () => {
  it("allows https anywhere", () => {
    expect(allowedSrc("https://raw.example.org/aauth.walkthrough.json", page)).toBe(true);
  });
  it("allows the page's own origin, relative or absolute", () => {
    expect(allowedSrc("../data/aauth.walkthrough.json", page)).toBe(true);
    expect(allowedSrc("./x.json", "http://localhost:5173/")).toBe(true);
  });
  it("refuses http elsewhere and other schemes", () => {
    expect(allowedSrc("http://other.example/x.json", page)).toBe(false);
    expect(allowedSrc("javascript:alert(1)", page)).toBe(false);
    expect(allowedSrc("data:application/json,{}", page)).toBe(false);
    expect(allowedSrc("file:///etc/passwd", "file:///tmp/viewer.html")).toBe(false);
  });
  it("refuses credentials in the URL", () => {
    expect(allowedSrc("https://user:pw@example.org/x.json", page)).toBe(false);
  });
});

describe("srcFromQuery", () => {
  it("returns the resolved URL", () => {
    expect(srcFromQuery(`${page}?src=../w/aauth.walkthrough.json`)).toBe("https://docs.example.com/w/aauth.walkthrough.json");
  });
  it("returns null without src", () => {
    expect(srcFromQuery(page)).toBeNull();
    expect(srcFromQuery(`${page}?src=`)).toBeNull();
  });
  it("throws for a refused URL", () => {
    expect(() => srcFromQuery(`${page}?src=${encodeURIComponent("http://evil.example/x.json")}`)).toThrow("https:");
  });
});

describe("links", () => {
  const base = "https://cdn.example.org/walkthroughs/aauth.walkthrough.json";
  it("resolves against the walkthrough's URL", () => {
    expect(resolve("audio/clip.mp3", base)).toBe("https://cdn.example.org/walkthroughs/audio/clip.mp3");
  });
  it("keeps http and https links only", () => {
    expect(safeHref("https://github.com/aauth-dev/aauth-go", base)).toBe("https://github.com/aauth-dev/aauth-go");
    expect(safeHref("docs/x.html", base)).toBe("https://cdn.example.org/walkthroughs/docs/x.html");
    expect(safeHref("javascript:alert(1)", base)).toBeUndefined();
    expect(safeHref("data:text/html,<b>x</b>", base)).toBeUndefined();
  });
});
