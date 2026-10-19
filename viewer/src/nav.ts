import type { Scenario, Step, Walkthrough } from "./types";

// Position is where the viewer is: a scenario and a step. Step 0 is the
// overview, with every step shown and none active.
export interface Position {
  scenarioId: string;
  step: number;
}

// parseHash reads "#<scenario-id>/<step>". Unknown scenarios fall back to
// the first one; steps are clamped to the scenario.
export function parseHash(hash: string, w: Walkthrough): Position {
  const first = w.scenarios[0]?.id ?? "";
  const m = /^#?([^/]+)(?:\/(\d+))?$/.exec(decodeURIComponent(hash));
  const sc = m ? w.scenarios.find((s) => s.id === m[1]) : undefined;
  if (!sc) {
    return { scenarioId: first, step: 0 };
  }
  return { scenarioId: sc.id, step: clampStep(sc, m && m[2] ? Number(m[2]) : 0) };
}

export function formatHash(p: Position): string {
  return `#${encodeURIComponent(p.scenarioId)}/${p.step}`;
}

export function clampStep(sc: Scenario, step: number): number {
  return Math.max(0, Math.min(sc.steps.length, Math.trunc(step)));
}

// Exchange is the request and response shown for the active step, as step
// indexes into the scenario (or null when there is none). responseLater is
// true when the response exists but the presenter has not reached it yet.
export interface Exchange {
  request: number | null;
  response: number | null;
  responseLater: number | null;
}

// A step's example fills the request pane when the step is a request and
// the response pane when it is a response. An example with only decoded
// tokens (such as a token handed over outside HTTP) follows the same rule.
const tokensOnly = (s: Step) => Boolean(s.example && !s.example.request && !s.example.response && s.example.tokens?.length);
const hasRequest = (s: Step) => Boolean(s.example?.request) || (!s.response && tokensOnly(s));
const hasResponse = (s: Step) => Boolean(s.example?.response) || (Boolean(s.response) && tokensOnly(s));

// exchangeFor pairs requests with responses. When a request is active, its
// response is the next step going the other way between the same two roles;
// it is revealed only once the presenter steps to it. When a response is
// active, its request is the nearest earlier request going the other way,
// such as the poll a pending response answers.
export function exchangeFor(sc: Scenario, index: number): Exchange {
  const steps = sc.steps;
  const s = steps[index];
  if (!s) {
    return { request: null, response: null, responseLater: null };
  }
  if (hasRequest(s)) {
    if (hasResponse(s)) {
      return { request: index, response: index, responseLater: null };
    }
    const next = steps[index + 1];
    const later = next && next.response && next.from === s.to && next.to === s.from && hasResponse(next) ? index + 1 : null;
    return { request: index, response: null, responseLater: later };
  }
  if (hasResponse(s)) {
    for (let j = index - 1; j >= 0; j--) {
      const r = steps[j];
      if (!r.response && r.from === s.to && r.to === s.from && hasRequest(r)) {
        return { request: j, response: index, responseLater: null };
      }
    }
    return { request: null, response: index, responseLater: null };
  }
  return { request: null, response: null, responseLater: null };
}

// scenarioOrder lists scenarios as the selector shows them: main scenarios,
// then fragments, as the walkthrough already orders them.
export function scenarioGroups(w: Walkthrough): { main: Scenario[]; fragments: Scenario[] } {
  return {
    main: w.scenarios.filter((s) => !s.fragment),
    fragments: w.scenarios.filter((s) => s.fragment),
  };
}
