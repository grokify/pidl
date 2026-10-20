import { useCallback, useEffect, useState } from "react";
import { isTyping } from "../host";
import { clampStep, formatHash, parseHash, type Position } from "../nav";
import type { Walkthrough } from "../types";

// usePosition holds the current scenario and step and handles step keys
// sent to keyTarget. With hash, it keeps the position in the URL hash, so
// a link or reload returns to the same step.
export function usePosition(w: Walkthrough, hash: boolean, keyTarget: EventTarget | null) {
  const [pos, setPos] = useState<Position>(() => parseHash(hash ? window.location.hash : "", w));
  const scenario = w.scenarios.find((s) => s.id === pos.scenarioId) ?? w.scenarios[0];

  useEffect(() => {
    if (!hash) return;
    const target = formatHash(pos, w);
    if (window.location.hash !== target) {
      window.history.replaceState(null, "", target);
    }
  }, [hash, pos, w]);

  useEffect(() => {
    if (!hash) return;
    const onHash = () => setPos(parseHash(window.location.hash, w));
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, [hash, w]);

  const goTo = useCallback(
    (step: number) => setPos((p) => ({ ...p, step: clampStep(scenario, step) })),
    [scenario],
  );
  const selectScenario = useCallback((scenarioId: string) => setPos({ scenarioId, step: 0 }), []);
  // move steps from the latest position, so keys pressed in quick
  // succession each count.
  const move = useCallback(
    (delta: number) => setPos((p) => ({ ...p, step: clampStep(scenario, p.step + delta) })),
    [scenario],
  );

  useEffect(() => {
    if (!keyTarget) return;
    const onKey = (e: KeyboardEvent) => {
      if (isTyping(e)) return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      switch (e.key) {
        case "ArrowRight":
        case "ArrowDown":
        case "PageDown":
        case " ":
          move(1);
          break;
        case "ArrowLeft":
        case "ArrowUp":
        case "PageUp":
          move(-1);
          break;
        case "Home":
          goTo(0);
          break;
        case "End":
          goTo(scenario.steps.length);
          break;
        default:
          return;
      }
      e.preventDefault();
    };
    const listener = onKey as EventListener;
    keyTarget.addEventListener("keydown", listener);
    return () => keyTarget.removeEventListener("keydown", listener);
  }, [keyTarget, goTo, move, scenario.steps.length]);

  return { scenario, step: pos.step, goTo, selectScenario };
}
