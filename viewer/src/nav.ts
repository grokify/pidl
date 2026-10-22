import type { Scenario, Walkthrough } from "./types";

// Position is where the viewer is: a scenario and a step. Step 0 is the
// overview, with every step shown and none active.
export interface Position {
  scenarioId: string;
  step: number;
}

// parseHash reads "#<scenario-id>/<step>", where <step> is a step number
// or a step ID. Unknown scenarios fall back to the first one; numbers are
// clamped to the scenario and unknown step IDs open the overview.
export function parseHash(hash: string, w: Walkthrough): Position {
  const first = w.scenarios[0]?.id ?? "";
  const m = /^#?([^/]+)(?:\/([^/]+))?$/.exec(decodeURIComponent(hash));
  const sc = m ? w.scenarios.find((s) => s.id === m[1]) : undefined;
  if (!sc || !m) {
    return { scenarioId: first, step: 0 };
  }
  const target = m[2] ?? "";
  if (/^\d+$/.test(target) || target === "") {
    return { scenarioId: sc.id, step: clampStep(sc, Number(target)) };
  }
  const i = sc.steps.findIndex((s) => s.id === target);
  return { scenarioId: sc.id, step: i + 1 };
}

// formatHash writes the step's ID when it has one, so links survive steps
// being added or moved, and its number otherwise.
export function formatHash(p: Position, w?: Walkthrough): string {
  const sc = w?.scenarios.find((s) => s.id === p.scenarioId);
  const id = p.step > 0 ? sc?.steps[p.step - 1]?.id : undefined;
  return `#${encodeURIComponent(p.scenarioId)}/${id ? encodeURIComponent(id) : p.step}`;
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

// exchangeFor pairs requests with responses by direction, whether or not
// the steps carry examples. When a request is active, its response is the
// next step going back between the same two roles; it is revealed only
// once the presenter steps to it. When a response is active, its request is
// the nearest earlier request going the other way, such as the poll a
// pending response answers. A step without an example still takes its
// place, so the pane can say what the message is.
export function exchangeFor(sc: Scenario, index: number): Exchange {
  const steps = sc.steps;
  const s = steps[index];
  if (!s) {
    return { request: null, response: null, responseLater: null };
  }
  if (!s.response) {
    if (s.example?.request && s.example.response) {
      return { request: index, response: index, responseLater: null };
    }
    // The answer may come several steps later, after other parties act
    // (a pending poll answered once the PS has federated with the AS). A new
    // request between the same two roles ends the search.
    let later: number | null = null;
    for (let j = index + 1; j < steps.length; j++) {
      const n = steps[j];
      if (n.response && n.from === s.to && n.to === s.from) {
        later = j;
        break;
      }
      if (!n.response && n.from === s.from && n.to === s.to) break;
    }
    return { request: index, response: null, responseLater: later };
  }
  for (let j = index - 1; j >= 0; j--) {
    const r = steps[j];
    if (!r.response && r.from === s.to && r.to === s.from) {
      return { request: j, response: index, responseLater: null };
    }
  }
  return { request: null, response: index, responseLater: null };
}

// scenarioOrder lists scenarios as the selector shows them: main scenarios,
// then fragments, as the walkthrough already orders them.
export function scenarioGroups(w: Walkthrough): { main: Scenario[]; fragments: Scenario[] } {
  return {
    main: w.scenarios.filter((s) => !s.fragment),
    fragments: w.scenarios.filter((s) => s.fragment),
  };
}
