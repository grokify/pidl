import { useCallback, useEffect, useState } from "react";
import { clampStep, formatHash, parseHash, type Position } from "../nav";
import type { Walkthrough } from "../types";

// usePosition keeps the current scenario and step in the URL hash, so a
// link or reload returns to the same step, and handles step keys.
export function usePosition(w: Walkthrough) {
  const [pos, setPos] = useState<Position>(() => parseHash(window.location.hash, w));
  const scenario = w.scenarios.find((s) => s.id === pos.scenarioId) ?? w.scenarios[0];

  useEffect(() => {
    const target = formatHash(pos);
    if (window.location.hash !== target) {
      window.history.replaceState(null, "", target);
    }
  }, [pos]);

  useEffect(() => {
    const onHash = () => setPos(parseHash(window.location.hash, w));
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, [w]);

  const goTo = useCallback(
    (step: number) => setPos((p) => ({ ...p, step: clampStep(scenario, step) })),
    [scenario],
  );
  const selectScenario = useCallback((scenarioId: string) => setPos({ scenarioId, step: 0 }), []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === "SELECT" || t.tagName === "INPUT" || t.tagName === "TEXTAREA")) return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      switch (e.key) {
        case "ArrowRight":
        case "ArrowDown":
        case "PageDown":
        case " ":
          goTo(pos.step + 1);
          break;
        case "ArrowLeft":
        case "ArrowUp":
        case "PageUp":
          goTo(pos.step - 1);
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
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [goTo, pos.step, scenario.steps.length]);

  return { scenario, step: pos.step, goTo, selectScenario };
}
