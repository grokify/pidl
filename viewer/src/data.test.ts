import { describe, expect, it } from "vitest";
import { checkFormat, load, parseWalkthrough, WALKTHROUGH_FORMAT } from "./data";

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

describe("load", () => {
  const page = "https://docs.example.com/guide/";
  const body = JSON.stringify({ format: WALKTHROUGH_FORMAT, ...minimal });
  const fetchOk = (calls: { url: string; init?: RequestInit }[]) =>
    (async (url: string, init?: RequestInit) => {
      calls.push({ url, init });
      return new Response(body, { status: 200 });
    }) as unknown as typeof fetch;

  it("reads inline data, parsed or as text, with the page as base", async () => {
    expect((await load({ data: minimal }, page)).base).toBe(page);
    expect((await load({ data: body }, page)).walkthrough.format).toBe(WALKTHROUGH_FORMAT);
  });
  it("fetches src without credentials and uses it as base", async () => {
    const calls: { url: string; init?: RequestInit }[] = [];
    const got = await load({ src: "../w/api.walkthrough.json" }, page, fetchOk(calls));
    expect(calls[0].url).toBe("https://docs.example.com/w/api.walkthrough.json");
    expect(calls[0].init?.credentials).toBe("omit");
    expect(got.base).toBe("https://docs.example.com/w/api.walkthrough.json");
  });
  it("reports a failed fetch", async () => {
    const notFound = (async () => new Response("", { status: 404 })) as unknown as typeof fetch;
    await expect(load({ src: "https://x.example/w.json" }, page, notFound)).rejects.toThrow("https://x.example/w.json returned 404");
  });
  it("needs data or src", async () => {
    await expect(load({}, page)).rejects.toThrow("give data or src");
  });
  it("refuses a value that is not an object", () => {
    expect(() => parseWalkthrough("[]")).toThrow("expected a JSON object");
  });
});
