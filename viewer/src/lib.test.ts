import { describe, expect, it } from "vitest";
import { autoMountOptions } from "./lib";

const script = (attrs: Record<string, string>) =>
  ({ getAttribute: (n: string) => attrs[n] ?? null }) as unknown as Element;

const page = (url: string, scripts = 1, inline: Record<string, string> = {}) =>
  ({
    location: { href: url },
    querySelectorAll: () => ({ length: scripts }),
    getElementById: (id: string) => (id in inline ? { textContent: inline[id] } : null),
  }) as unknown as Document;

describe("autoMountOptions", () => {
  it("ignores a tag without a target", () => {
    expect(autoMountOptions(script({}), page("https://d.example/"))).toBeNull();
  });
  it("prefers inline data, then data-pidl-src, then ?src=", () => {
    const doc = page("https://d.example/v.html?src=q.json", 1, { w: '{"id":"x"}' });
    expect(autoMountOptions(script({ "data-pidl-target": "#v", "data-pidl-data": "w", "data-pidl-src": "a.json" }), doc)?.options)
      .toEqual({ data: '{"id":"x"}', hash: true });
    expect(autoMountOptions(script({ "data-pidl-target": "#v", "data-pidl-src": "a.json" }), doc)?.options)
      .toEqual({ src: "a.json", hash: true });
    expect(autoMountOptions(script({ "data-pidl-target": "#v" }), doc)?.options)
      .toEqual({ src: "https://d.example/q.json", hash: true });
  });
  it("uses the hash only for the one auto-mounted viewer", () => {
    const tag = script({ "data-pidl-target": "#v", "data-pidl-src": "a.json" });
    expect(autoMountOptions(tag, page("https://d.example/", 2))?.options.hash).toBe(false);
  });
  it("explains a missing walkthrough", () => {
    expect(() => autoMountOptions(script({ "data-pidl-target": "#v" }), page("https://d.example/"))).toThrow("data-pidl-src");
    expect(() => autoMountOptions(script({ "data-pidl-target": "#v", "data-pidl-data": "w" }), page("https://d.example/"))).toThrow("#w");
  });
  it("refuses a ?src= from another origin over http", () => {
    expect(() => autoMountOptions(script({ "data-pidl-target": "#v" }), page("https://d.example/?src=http://e.example/x.json"))).toThrow("https:");
  });
});
