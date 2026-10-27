import { describe, expect, it } from "vitest";
import petstoreObserved from "../../observe/testdata/petstore.observed.json?raw";
import { exchangeFor } from "./nav";
import {
  EXTRA,
  OBSERVED_KIND,
  clock,
  gapText,
  isObserved,
  observedGap,
  observedNote,
  observedRunLabel,
  observedTiming,
  prepareObserved,
  viewScenario,
} from "./observed";
import type { Scenario, Step, Walkthrough } from "./types";

// petstore.observed.json holds the observed entries pidl writes for the
// Petstore run behind the aauth-go sandbox (the Go side's golden file).
const entries = JSON.parse(petstoreObserved) as Scenario[];
const walkthrough = { format: "pidl-walkthrough/v1", id: "aauth", name: "AAuth", language: "en-US", roles: [], scenarios: entries } as unknown as Walkthrough;
const prepared = prepareObserved(walkthrough);
const order = prepared.scenarios.find((s) => s.id === "ps-authorization-challenge.observed-2")!;

describe("prepareObserved", () => {
  it("puts each extra exchange after the step it followed, and renumbers", () => {
    const original = entries.find((s) => s.id === order.id)!;
    expect(order.steps).toHaveLength(original.steps.length + original.observed_extra!.length);
    expect(order.steps.map((s) => s.number)).toEqual(order.steps.map((_, i) => i + 1));
    // The first extra follows the person token's request and response (steps 3 and 4 of the scenario).
    const firstExtra = order.steps.findIndex((s) => s.observed_status === EXTRA);
    expect(order.steps[firstExtra - 1].id).toBe("person-token-required");
    expect(order.steps[firstExtra]).toMatchObject({ from: "ps", to: "ap", label: "GET /.well-known/aauth-agent.json → 200" });
    // The last extra follows the last step.
    expect(order.steps.at(-1)!.observed_status).toBe(EXTRA);
  });

  it("leaves scenarios without extras, and walkthroughs without observed runs, as they are", () => {
    const plain = { ...walkthrough, scenarios: [{ id: "s", name: "S", position: 1, participants: [], steps: [] }] } as unknown as Walkthrough;
    expect(prepareObserved(plain)).toBe(plain);
  });

  it("leaves out extras it cannot draw", () => {
    const sc = {
      id: "s.observed-1", name: "S", position: 1, participants: ["a", "b"],
      steps: [{ number: 1, id: "x", from: "a", to: "b", label: "X", location: "" }],
      observed_run: { scenario: "s", source: "t", instance: 1, instances: 1 },
      observed_extra: [
        { after_step: 1, observed: { request: { method: "GET", url: "https://h.test/p" } } },
        { after_step: 0, from: "b", to: "a", observed: { request: { method: "GET", url: "https://h.test/q" } } },
      ],
    } as unknown as Scenario;
    const out = prepareObserved({ ...walkthrough, scenarios: [sc] }).scenarios[0];
    expect(out.steps.map((s) => s.label)).toEqual(["GET /q", "X"]);
  });
});

describe("viewScenario", () => {
  it("shows the observed exchange as each step's example in observed mode", () => {
    const view = viewScenario(order, "observed");
    const request = view.steps.find((s) => s.id === "request")!;
    expect(request.example).toMatchObject({ kind: OBSERVED_KIND, request: { method: "POST" } });
    expect(request.example!.source).toMatch(/^Observed at \d{2}:\d{2}:\d{2}\.\d{3}$/);
    expect(view.steps.find((s) => s.id === "consent.direct-user")!.example).toBeUndefined();
  });

  it("keeps the concept in concept mode, and plain scenarios in either mode", () => {
    expect(viewScenario(order, "concept")).toBe(order);
    const plain = { id: "s", name: "S", position: 1, participants: [], steps: [] } as unknown as Scenario;
    expect(viewScenario(plain, "observed")).toBe(plain);
    expect(isObserved(plain)).toBe(false);
  });

  it("pairs an observed request with its response step as the concept does", () => {
    const view = viewScenario(order, "observed");
    const i = view.steps.findIndex((s) => s.id === "request");
    expect(exchangeFor(view, i)).toEqual({ request: i, response: null, responseLater: i + 1 });
    const extra = view.steps.findIndex((s) => s.observed_status === EXTRA);
    expect(exchangeFor(view, extra)).toEqual({ request: extra, response: extra, responseLater: null });
  });
});

describe("observed details", () => {
  const step = (status: string) => ({ number: 1, from: "a", to: "b", label: "X", location: "", observed_status: status }) as Step;

  it("explains steps without an observed exchange", () => {
    expect(observedGap(step("notSeen"), "observed")).toMatch(/^Not seen in this run/);
    expect(observedGap(step("unmatchable"), "observed")).toMatch(/no rule to match by/);
    expect(observedGap(step("unobservable"), "observed")).toMatch(/person's browser/);
    expect(observedGap(step("skipped"), "observed")).toMatch(/^Skipped/);
    expect(observedGap(step(EXTRA), "concept")).toMatch(/not a step in the document/);
    expect(observedGap(step("notSeen"), "concept")).toBeUndefined();
    expect(observedGap(step("matched"), "observed")).toBeUndefined();
  });

  it("times each exchange from the previous one", () => {
    const i = order.steps.findIndex((s) => s.id === "person-token.request");
    const { time, gapMs } = observedTiming(order, i);
    expect(time).toBeDefined();
    expect(gapMs).toBeGreaterThan(0);
    expect(observedTiming(order, 0).gapMs).toBeUndefined();
  });

  it("summarizes the run", () => {
    expect(observedRunLabel(order)).toBe("Observed 2 of 2 · aauth-sandbox · 14 matched (2 weak) · 7 extra · values redacted");
  });

  it("notes the capturing tool's name and assertions", () => {
    expect(
      observedNote({
        label: "8. POST /token",
        annotations: [
          { kind: "assertion", text: "status 202", ok: true },
          { kind: "assertion", text: "header Location", ok: false },
        ],
      }),
    ).toBe("Captured as “8. POST /token”. 1 of 2 assertions failed: header Location.");
    expect(observedNote({})).toBeUndefined();
  });

  it("formats times and gaps", () => {
    expect(clock("2026-10-27T09:02:31.555652175Z")).toBe("09:02:31.555");
    expect(gapText(26)).toBe("+26 ms");
    expect(gapText(2500)).toBe("+2.5 s");
    expect(gapText(180_000)).toBe("+3 min");
  });
});
