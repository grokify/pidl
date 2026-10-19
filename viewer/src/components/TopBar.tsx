import { scenarioGroups } from "../nav";
import type { Scenario, Walkthrough } from "../types";

interface Props {
  walkthrough: Walkthrough;
  scenario: Scenario;
  step: number;
  onScenario: (id: string) => void;
  onStep: (step: number) => void;
}

export function TopBar({ walkthrough, scenario, step, onScenario, onStep }: Props) {
  const { main, fragments } = scenarioGroups(walkthrough);
  const total = scenario.steps.length;
  return (
    <header className="topbar">
      <div className="brand">
        <span className="brand-name">{walkthrough.name}</span>
        <span className="brand-sub">protocol walkthrough</span>
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
      <nav className="stepper" aria-label="Steps">
        <button onClick={() => onStep(0)} disabled={step === 0} title="Overview (Home)">
          Overview
        </button>
        <button onClick={() => onStep(step - 1)} disabled={step === 0} title="Previous step (←)" aria-label="Previous step">
          ←
        </button>
        <span className="step-count" aria-live="polite">
          {step === 0 ? `${total} steps` : `Step ${step} of ${total}`}
        </span>
        <button onClick={() => onStep(step + 1)} disabled={step === total} title="Next step (→)" aria-label="Next step">
          →
        </button>
      </nav>
    </header>
  );
}
