import type { Theme } from "../hooks/usePrefs";
import { scenarioGroups } from "../nav";
import type { Mode } from "../observed";
import type { Scenario, Walkthrough } from "../types";

interface Props {
  walkthrough: Walkthrough;
  // source is the host the walkthrough came from, when not this page's.
  source?: string;
  scenario: Scenario;
  onScenario: (id: string) => void;
  theme: Theme;
  onTheme: () => void;
  present: boolean;
  onPresent: () => void;
  // mode is set for an observed entry, which offers the Concept / Observed switch.
  mode?: Mode;
  onMode?: (mode: Mode) => void;
}

const THEME_LABEL: Record<Theme, string> = { auto: "Auto", light: "Light", dark: "Dark" };

export function TopBar({ walkthrough, source, scenario, onScenario, theme, onTheme, present, onPresent, mode, onMode }: Props) {
  const { main, fragments } = scenarioGroups(walkthrough);
  return (
    <header className="topbar">
      <div className="brand">
        <span className="brand-name">{walkthrough.name}</span>
        <span className="brand-sub">protocol walkthrough{source && ` · from ${source}`}</span>
      </div>
      <label className="scenario-select">
        <span>Scenario</span>
        <select value={scenario.id} onChange={(e) => onScenario(e.target.value)} aria-label="Scenario">
          <optgroup label="Scenarios">
            {main.map((s) => (
              <option key={s.id} value={s.id}>
                {s.position}. {s.name}
              </option>
            ))}
          </optgroup>
          {fragments.length > 0 && (
            <optgroup label="Shared steps">
              {fragments.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.position}. {s.name}
                </option>
              ))}
            </optgroup>
          )}
        </select>
      </label>
      {mode && onMode && (
        <div className="mode-switch" role="group" aria-label="Show the concept or the observed run">
          <button className={mode === "concept" ? "on" : ""} aria-pressed={mode === "concept"} onClick={() => onMode("concept")} title="Each step's example, as the document explains it">
            Concept
          </button>
          <button className={mode === "observed" ? "on" : ""} aria-pressed={mode === "observed"} onClick={() => onMode("observed")} title="The exchanges a real run produced">
            Observed
          </button>
        </div>
      )}
      <div className="view-controls">
        <button onClick={onTheme} title="Theme: auto, light or dark">
          Theme: {THEME_LABEL[theme]}
        </button>
        <button onClick={onPresent} className={present ? "on" : ""} title="Present on a 16:9 stage (P)">
          {present ? "Exit present" : "Present"}
        </button>
      </div>
    </header>
  );
}
