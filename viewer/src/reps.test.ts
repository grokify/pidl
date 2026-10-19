import { describe, expect, it } from "vitest";
import { hasRep, repName, representations } from "./reps";
import type { Step, Walkthrough } from "./types";

const req = { source: "x", kind: "draft" as const, request: { method: "GET", url: "https://r.example/" } };
const res = { source: "x", kind: "draft" as const, response: { status: 200 } };
const step = (n: number, opts: Partial<Step>): Step => ({ number: n, from: "a", to: "b", label: "s", location: "flows[0]", ...opts });

const signed = step(1, {
  curl: "curl 'https://r.example/'",
  example: { ...req, samples: [{ language: "go", code: "c.Do(req)" }, { language: "python", source: { repo: "r", path: "p", region: "x" } }] },
});
const ok = step(2, { response: true, example: { ...res, samples: [{ language: "typescript", code: "await r.json()" }] } });
const w: Walkthrough = { id: "x", name: "X", roles: [], scenarios: [{ position: 1, id: "s", name: "S", participants: [], steps: [signed, ok] }] };

describe("representations", () => {
  it("lists HTTP, cURL, then sample languages in order of appearance", () => {
    expect(representations(w)).toEqual(["http", "curl", "go", "python", "typescript"]);
  });
  it("omits cURL when no request has one", () => {
    const noCurl = { ...w, scenarios: [{ ...w.scenarios[0], steps: [ok] }] };
    expect(representations(noCurl)).toEqual(["http", "typescript"]);
  });
  it("knows which panel can show which representation", () => {
    expect(hasRep(signed, "request", "http")).toBe(true);
    expect(hasRep(signed, "request", "curl")).toBe(true);
    expect(hasRep(signed, "request", "go")).toBe(true);
    expect(hasRep(signed, "request", "python")).toBe(false); // unresolved source has no code
    expect(hasRep(ok, "response", "curl")).toBe(false);
    expect(hasRep(ok, "response", "http")).toBe(true);
    expect(hasRep(step(3, {}), "request", "http")).toBe(false);
  });
  it("names representations", () => {
    expect(repName("csharp")).toBe("C#");
    expect(repName("curl")).toBe("cURL");
    expect(repName("elixir")).toBe("Elixir");
  });
});
