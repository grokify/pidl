import { describe, expect, it } from "vitest";
import { fromOtherOrigin, isTyping } from "./host";

describe("fromOtherOrigin", () => {
  const page = "https://docs.example.com/guide/";
  it("names another origin's host", () => {
    expect(fromOtherOrigin("https://raw.example.org/a.walkthrough.json", page)).toBe("raw.example.org");
  });
  it("is empty for the page's own origin", () => {
    expect(fromOtherOrigin("../w/a.walkthrough.json", page)).toBeUndefined();
    expect(fromOtherOrigin(page, page)).toBeUndefined();
  });
});

describe("isTyping", () => {
  const event = (tagName: string) =>
    ({ composedPath: () => [{ tagName }], target: null }) as unknown as KeyboardEvent;
  it("sees form controls, also inside a shadow root", () => {
    expect(isTyping(event("SELECT"))).toBe(true);
    expect(isTyping(event("INPUT"))).toBe(true);
    expect(isTyping(event("DIV"))).toBe(false);
  });
});
