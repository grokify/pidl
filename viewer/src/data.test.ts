import { describe, expect, it } from "vitest";
import { checkFormat, parseWalkthrough, WALKTHROUGH_FORMAT } from "./data";

const minimal = { id: "x", name: "X", roles: [], scenarios: [] };

describe("format", () => {
  it("accepts v1 and later v1 revisions", () => {
    expect(() => checkFormat(WALKTHROUGH_FORMAT)).not.toThrow();
    expect(() => checkFormat("pidl-walkthrough/v1.2")).not.toThrow();
  });
  it("reads a walkthrough written before v1", () => {
    expect(parseWalkthrough(JSON.stringify(minimal)).id).toBe("x");
  });
  it("refuses another major version", () => {
    expect(() => parseWalkthrough(JSON.stringify({ format: "pidl-walkthrough/v2", ...minimal })))
      .toThrow("This viewer reads pidl-walkthrough/v1; the walkthrough is pidl-walkthrough/v2");
  });
  it("refuses another format", () => {
    expect(() => checkFormat("postman/v2.1")).toThrow('format is "postman/v2.1"');
    expect(() => checkFormat(1)).toThrow("format is 1");
  });
  it("still requires roles and scenarios", () => {
    expect(() => parseWalkthrough(JSON.stringify({ format: WALKTHROUGH_FORMAT, id: "x" })))
      .toThrow("expected roles and scenarios");
  });
});
