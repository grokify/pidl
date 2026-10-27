import type { Example, Scenario, Step, Walkthrough } from "./types";

// Observed runs: a scenario entry with observed_run shows a run observed in
// a real system. Its steps keep their examples (the concept) and carry the
// observed exchange; exchanges that matched no step are in observed_extra.

/** Which form of an observed entry the viewer shows. */
export type Mode = "concept" | "observed";

/** The status of a step made from an exchange that matched no step. */
export const EXTRA = "extra";

/** The example kind of a message taken from an observed run. */
export const OBSERVED_KIND = "observed";

type ObservedMessage = NonNullable<Step["observed"]>;

export function isObserved(sc: Scenario): boolean {
  return sc.observed_run != null;
}

// prepareObserved puts each observed entry's extra exchanges among its
// steps, after the step each followed, and renumbers the steps, so the
// diagram and the step keys reach them. Extras whose participants the
// binding could not resolve cannot be drawn and are left out.
export function prepareObserved(w: Walkthrough): Walkthrough {
  if (!w.scenarios.some((s) => s.observed_extra?.length)) return w;
  return { ...w, scenarios: w.scenarios.map((s) => (s.observed_run && s.observed_extra?.length ? interleave(s) : s)) };
}

function interleave(sc: Scenario): Scenario {
  const after = new Map<number, Step[]>();
  (sc.observed_extra ?? []).forEach((e, i) => {
    if (!e.from || !e.to) return;
    const req = e.observed.request;
    const res = e.observed.response;
    const path = req ? pathOf(req.url) : "";
    const step: Step = {
      number: 0,
      id: `extra-${i + 1}`,
      from: e.from,
      to: e.to,
      label: [req ? `${req.method} ${path}` : "exchange", res ? `→ ${res.status}` : ""].filter(Boolean).join(" "),
      location: "",
      description: "Observed in this run, but not a step in the document: servers fetch metadata and keys, for example, between the steps the protocol describes.",
      observed: e.observed,
      observed_status: EXTRA,
    };
    after.set(e.after_step, [...(after.get(e.after_step) ?? []), step]);
  });
  const steps: Step[] = [...(after.get(0) ?? [])];
  for (const s of sc.steps) {
    steps.push(s);
    steps.push(...(after.get(s.number) ?? []));
  }
  return { ...sc, steps: steps.map((s, i) => ({ ...s, number: i + 1 })) };
}

function pathOf(url: string): string {
  try {
    const u = new URL(url);
    return u.pathname + u.search;
  } catch {
    return url;
  }
}

// viewScenario is the scenario as a mode shows it. In observed mode each
// step's example is the exchange the run produced, so the request,
// response and token panels show it as they show an example; a step with
// none has no example, and observedGap says why.
export function viewScenario(sc: Scenario, mode: Mode): Scenario {
  if (!isObserved(sc) || mode === "concept") return sc;
  return { ...sc, steps: sc.steps.map((s) => ({ ...s, example: observedExample(s) })) };
}

function observedExample(s: Step): Example | undefined {
  const o = s.observed;
  if (!o) return undefined;
  const source = ["Observed", o.time ? `at ${clock(o.time)}` : "", o.repeats ? `· ${o.repeats} exchanges` : ""]
    .filter(Boolean)
    .join(" ");
  return {
    source,
    kind: OBSERVED_KIND as Example["kind"],
    request: o.request,
    response: o.response,
    tokens: o.tokens,
    note: observedNote(o),
  };
}

// observedNote names what the capturing tool recorded about the exchange:
// its name for it and the outcome of its assertions.
export function observedNote(o: ObservedMessage): string | undefined {
  const parts: string[] = [];
  if (o.label) parts.push(`Captured as “${o.label}”.`);
  const assertions = (o.annotations ?? []).filter((a) => a.kind === "assertion");
  if (assertions.length > 0) {
    const failed = assertions.filter((a) => a.ok === false);
    parts.push(
      failed.length === 0
        ? `${assertions.length} assertions passed.`
        : `${failed.length} of ${assertions.length} assertions failed: ${failed.map((a) => a.text).join("; ")}.`,
    );
  }
  for (const a of o.annotations ?? []) {
    if (a.kind === "client-event") parts.push(a.text);
  }
  if (o.body_truncated) parts.push("A body was cut at the capture's size limit.");
  return parts.length > 0 ? parts.join(" ") : undefined;
}

// observedGap explains, for a step with no observed exchange, why: not
// seen, skipped, not matchable or not observable. In concept mode it
// explains an extra step, which has no example in the document.
export function observedGap(s: Step, mode: Mode): string | undefined {
  switch (s.observed_status) {
    case "notSeen":
      return mode === "observed" ? "Not seen in this run: no exchange matched this step." : undefined;
    case "skipped":
      return mode === "observed" ? "Skipped: an optional step this run did not perform." : undefined;
    case "unmatchable":
      return mode === "observed"
        ? "Not matched: the step has no rule to match by. Give it a match in the document or the binding."
        : undefined;
    case "unobservable":
      return mode === "observed"
        ? "Not observable: this step happens in the person's browser or out of band, where a server's capture cannot see it."
        : undefined;
    case EXTRA:
      return mode === "concept" ? "Observed only: this exchange is not a step in the document." : undefined;
  }
  return undefined;
}

// observedTiming returns when a step's exchange happened and the gap from
// the previous observed exchange, in milliseconds.
export function observedTiming(sc: Scenario, index: number): { time?: string; gapMs?: number } {
  const t = sc.steps[index]?.observed?.time;
  if (!t) return {};
  for (let j = index - 1; j >= 0; j--) {
    const p = sc.steps[j]?.observed?.time;
    if (p && p !== t) return { time: t, gapMs: Date.parse(t) - Date.parse(p) };
  }
  return { time: t };
}

// observedRunLabel summarizes an observed entry for its header.
export function observedRunLabel(sc: Scenario): string {
  const r = sc.observed_run;
  if (!r) return "";
  const parts = [
    `Observed ${r.instance} of ${r.instances}`,
    r.source,
    `${r.matched} matched${r.weak ? ` (${r.weak} weak)` : ""}`,
  ];
  if (r.not_seen) parts.push(`${r.not_seen} not seen`);
  if (r.extra) parts.push(`${r.extra} extra`);
  if (r.redacted) parts.push("values redacted");
  return parts.join(" · ");
}

// clock shows the time of day to the millisecond, in UTC.
export function clock(time: string): string {
  const d = new Date(time);
  if (Number.isNaN(d.getTime())) return time;
  return d.toISOString().slice(11, 23);
}

// gapText shows a gap between exchanges.
export function gapText(ms: number): string {
  if (ms < 1000) return `+${Math.round(ms)} ms`;
  if (ms < 60_000) return `+${(ms / 1000).toFixed(1)} s`;
  return `+${Math.round(ms / 60_000)} min`;
}
