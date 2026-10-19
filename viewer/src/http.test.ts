import { describe, expect, it } from "vitest";
import { coveredComponents, splitURL, statusLine } from "./http";

describe("http helpers", () => {
  it("formats status lines", () => {
    expect(statusLine(202)).toBe("HTTP/1.1 202 Accepted");
    expect(statusLine(299)).toBe("HTTP/1.1 299");
  });
  it("splits URLs", () => {
    expect(splitURL("https://ps.example/pending/f7a3b9c?x=1")).toEqual({ target: "/pending/f7a3b9c?x=1", host: "ps.example" });
    expect(splitURL("not a url")).toEqual({ target: "not a url", host: "" });
  });
  it("reads covered components", () => {
    const covered = coveredComponents([
      { name: "Signature-Input", value: 'sig=("@method" "@authority" "@path" "content-digest" "Signature-Key");created=1' },
    ]);
    expect([...covered]).toEqual(["@method", "@authority", "@path", "content-digest", "signature-key"]);
    expect(coveredComponents([]).size).toBe(0);
  });
});
