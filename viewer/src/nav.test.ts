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
const w: Walkthrough = { format: "pidl-walkthrough/v1", language: "en-US", id: "x", name: "X", roles: [], scenarios: [scenario, { ...scenario, id: "other" }] };

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
  it("reads and writes step IDs", () => {
    const withIds: Walkthrough = {
      ...w,
      scenarios: [{ ...scenario, steps: scenario.steps.map((s, i) => (i === 3 ? { ...s, id: "consent.poll" } : s)) }],
    };
    expect(parseHash("#three-party/consent.poll", withIds)).toEqual({ scenarioId: "three-party", step: 4 });
    expect(parseHash("#three-party/missing", withIds)).toEqual({ scenarioId: "three-party", step: 0 });
    expect(formatHash({ scenarioId: "three-party", step: 4 }, withIds)).toBe("#three-party/consent.poll");
    expect(formatHash({ scenarioId: "three-party", step: 2 }, withIds)).toBe("#three-party/2");
    expect(formatHash({ scenarioId: "three-party", step: 0 }, withIds)).toBe("#three-party/0");
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
  it("keeps a request without an example in its pane", () => {
    expect(exchangeFor(scenario, 2)).toEqual({ request: 2, response: null, responseLater: null });
  });
  it("pairs a response with a request that has no example", () => {
    const tokensOnly = { source: "x", kind: "derived" as const, tokens: [{ token: "agent_token" }] };
    const sc: Scenario = { ...scenario, steps: [step(1, "agent", "ap"), step(2, "ap", "agent", { response: true, example: tokensOnly })] };
    expect(exchangeFor(sc, 0)).toEqual({ request: 0, response: null, responseLater: 1 });
    expect(exchangeFor(sc, 1)).toEqual({ request: 0, response: 1, responseLater: null });
  });
  it("leaves the request empty when a response has none in the scenario", () => {
    const sc: Scenario = { ...scenario, steps: [step(1, "ps", "agent", { response: true, example: res })] };
    expect(exchangeFor(sc, 0)).toEqual({ request: null, response: 0, responseLater: null });
  });
  it("finds a response that arrives after other parties act", () => {
    const sc: Scenario = {
      ...scenario,
      steps: [
        step(1, "agent", "ps", { example: req }), // poll
        step(2, "ps", "as", { example: req }), // federate
        step(3, "as", "ps", { response: true, example: res }),
        step(4, "ps", "agent", { response: true, example: res }), // answers the poll
      ],
    };
    expect(exchangeFor(sc, 0)).toEqual({ request: 0, response: null, responseLater: 3 });
    expect(exchangeFor(sc, 3)).toEqual({ request: 0, response: 3, responseLater: null });
  });
  it("stops at a new request between the same roles", () => {
    const sc: Scenario = { ...scenario, steps: [step(1, "agent", "ps"), step(2, "agent", "ps"), step(3, "ps", "agent", { response: true })] };
    expect(exchangeFor(sc, 0).responseLater).toBeNull();
    expect(exchangeFor(sc, 1).responseLater).toBe(2);
  });
  it("handles out of range", () => {
    expect(exchangeFor(scenario, 9)).toEqual({ request: null, response: null, responseLater: null });
  });
});
