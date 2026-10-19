import { describe, expect, it } from "vitest";
import { clampStep, exchangeFor, formatHash, parseHash } from "./nav";
import type { Scenario, Step, Walkthrough } from "./types";

const step = (n: number, from: string, to: string, opts: Partial<Step> = {}): Step => ({
  number: n, from, to, label: `s${n}`, location: `flows[${n - 1}]`, ...opts,
});
const req = { source: "x", kind: "draft" as const, request: { method: "POST", url: "https://ps.example/token" } };
const res = { source: "x", kind: "draft" as const, response: { status: 200 } };

const scenario: Scenario = {
  position: 1, id: "three-party", name: "Three party", participants: ["agent", "ps"],
  steps: [
    step(1, "agent", "ps", { example: req }),                    // token request
    step(2, "ps", "agent", { response: true, example: res }),    // 202 interaction
    step(3, "agent", "person"),                                  // no example
    step(4, "agent", "ps", { example: req }),                    // poll
    step(5, "ps", "agent", { response: true, example: res }),    // auth token
  ],
};
const w: Walkthrough = { id: "x", name: "X", roles: [], scenarios: [scenario, { ...scenario, id: "other" }] };

describe("hash", () => {
  it("round-trips", () => {
    expect(parseHash(formatHash({ scenarioId: "other", step: 3 }), w)).toEqual({ scenarioId: "other", step: 3 });
  });
  it("falls back and clamps", () => {
    expect(parseHash("#nope/2", w)).toEqual({ scenarioId: "three-party", step: 0 });
    expect(parseHash("#other/99", w)).toEqual({ scenarioId: "other", step: 5 });
    expect(parseHash("", w)).toEqual({ scenarioId: "three-party", step: 0 });
    expect(parseHash("#other", w)).toEqual({ scenarioId: "other", step: 0 });
  });
  it("clamps steps", () => {
    expect(clampStep(scenario, -1)).toBe(0);
    expect(clampStep(scenario, 2.7)).toBe(2);
  });
});

describe("exchangeFor", () => {
  it("shows the request and reveals the response later", () => {
    expect(exchangeFor(scenario, 0)).toEqual({ request: 0, response: null, responseLater: 1 });
  });
  it("pairs a response with its request", () => {
    expect(exchangeFor(scenario, 1)).toEqual({ request: 0, response: 1, responseLater: null });
  });
  it("pairs a pending response with the poll, not the first request", () => {
    expect(exchangeFor(scenario, 4)).toEqual({ request: 3, response: 4, responseLater: null });
  });
  it("shows nothing for a step without an example", () => {
    expect(exchangeFor(scenario, 2)).toEqual({ request: null, response: null, responseLater: null });
  });
  it("shows a tokens-only response in the response pane", () => {
    const tokensOnly = { source: "x", kind: "derived" as const, tokens: [{ token: "agent_token" }] };
    const sc: Scenario = { ...scenario, steps: [step(1, "agent", "ap"), step(2, "ap", "agent", { response: true, example: tokensOnly })] };
    expect(exchangeFor(sc, 1)).toEqual({ request: null, response: 1, responseLater: null });
  });
  it("handles out of range", () => {
    expect(exchangeFor(scenario, 9)).toEqual({ request: null, response: null, responseLater: null });
  });
});
