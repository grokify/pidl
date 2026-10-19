import { describe, expect, it } from "vitest";
import { MARGIN_X, ROW_H, layoutScenario, rowY } from "./layout";
import type { Role, Scenario, Step } from "./types";

const roles: Role[] = ["person", "agent", "resource", "ps"].map((id, i) => ({ position: i + 1, id, name: id.toUpperCase() }));
const step = (n: number, from: string, to: string, fragments?: string[]): Step => ({
  number: n, from, to, label: `s${n}`, location: `flows[${n}]`, fragments,
});

const sc: Scenario = {
  position: 1, id: "s", name: "S",
  participants: ["ps", "agent", "person"], // out of role order on purpose
  steps: [
    step(1, "agent", "ps"),
    step(2, "ps", "person", ["Consent"]),
    step(3, "person", "ps", ["Consent", "Login"]),
    step(4, "agent", "ps", ["Consent"]),
    step(5, "ps", "agent"),
    step(6, "agent", "agent", ["Other"]),
  ],
};

describe("layoutScenario", () => {
  const l = layoutScenario(sc, roles, 900);

  it("orders columns by role order and spreads them", () => {
    expect(l.columns.map((c) => c.id)).toEqual(["person", "agent", "ps"]);
    expect(l.columns[0].x).toBe(MARGIN_X);
    expect(l.columns[2].x).toBe(900 - MARGIN_X);
    expect(l.columns[0].name).toBe("PERSON");
  });

  it("places one row per step", () => {
    expect(l.rows).toHaveLength(6);
    expect(l.rows[2].y - l.rows[1].y).toBe(ROW_H);
    expect(l.rows[0].x1).toBe(l.columns[1].x);
    expect(l.rows[0].x2).toBe(l.columns[2].x);
    expect(l.rows[5].self).toBe(true);
    expect(l.height).toBeGreaterThan(rowY(5));
  });

  it("groups consecutive fragment steps into nested bands", () => {
    const summary = l.bands.map((b) => [b.name, b.depth, b.first, b.last]);
    expect(summary).toEqual([
      ["Consent", 0, 1, 3],
      ["Other", 0, 5, 5],
      ["Login", 1, 2, 2],
    ]);
    const [outer, , inner] = l.bands;
    expect(inner.x).toBeGreaterThan(outer.x);
    expect(inner.y).toBeGreaterThan(outer.y);
  });

  it("centers a single participant", () => {
    const one = layoutScenario({ ...sc, participants: ["agent"], steps: [step(1, "agent", "agent")] }, roles, 600);
    expect(one.columns[0].x).toBe(300);
  });
});
