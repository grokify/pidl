interface Props {
  step: number;
  total: number;
  onStep: (step: number) => void;
}

// StepControls are presentation-style controls: overview, back, the step
// position, forward, and a progress bar.
export function StepControls({ step, total, onStep }: Props) {
  const pct = total === 0 ? 0 : (step / total) * 100;
  return (
    <div className="step-controls">
      <div className="progress" aria-hidden="true">
        <div className="progress-fill" style={{ width: `${pct}%` }} />
      </div>
      <div className="step-buttons">
        <button className="ghost" onClick={() => onStep(0)} disabled={step === 0} title="Overview (Home)">
          Overview
        </button>
        <button className="nav" onClick={() => onStep(step - 1)} disabled={step === 0} title="Previous step (←)" aria-label="Previous step">
          ←
        </button>
        <span className="step-count" aria-live="polite">
          {step === 0 ? (
            <>
              <strong>Overview</strong> · {total} steps
            </>
          ) : (
            <>
              Step <strong>{step}</strong> / {total}
            </>
          )}
        </span>
        <button className="nav" onClick={() => onStep(step + 1)} disabled={step === total} title="Next step (→)" aria-label="Next step">
          →
        </button>
        <button className="ghost" onClick={() => onStep(total)} disabled={step === total} title="Last step (End)">
          End
        </button>
      </div>
    </div>
  );
}
