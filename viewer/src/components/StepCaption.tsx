import type { Role, Scenario } from "../types";

interface Props {
  scenario: Scenario;
  roles: Role[];
  step: number;
}

// ScenarioIntro is the prose introduction above the diagram.
export function ScenarioIntro({ scenario }: { scenario: Scenario }) {
  return (
    <div className="intro">{scenario.description && <p>{scenario.description}</p>}</div>
  );
}

// StepNarrative explains the active step below the diagram. It has a fixed
// height so the diagram does not move as narratives change length.
export function StepNarrative({ scenario, roles, step }: Props) {
  const name = (id: string) => roles.find((r) => r.id === id)?.name ?? id;
  if (step === 0) {
    return (
      <div className="narrative">
        <p className="narrative-title">
          {scenario.steps.length} steps between {scenario.participants.map(name).join(", ")}
        </p>
        <p className="narrative-text muted">Press → or click a step to walk through the scenario one message at a time.</p>
      </div>
    );
  }
  const s = scenario.steps[step - 1];
  return (
    <div className="narrative">
      <p className="narrative-title">
        <span className="caption-num">{s.number}</span>
        {name(s.from)} <span className="caption-arrow">{s.response ? "⇢" : "→"}</span> {name(s.to)}
        <span className="narrative-label">{s.label}</span>
        {s.fragments && s.fragments.length > 0 && <span className="narrative-trail">{s.fragments.join(" › ")}</span>}
      </p>
      {s.description ? <p className="narrative-text">{s.description}</p> : <p className="narrative-text muted">No description for this step.</p>}
      {s.note && <p className="caption-note">{s.note}</p>}
      {s.findings && s.findings.length > 0 && (
        <div className="finding-chips">
          {s.findings.map((f) => (
            <span key={f.id + f.title} className={`finding-chip severity-${f.severity}`} title={f.description ? `${f.title}: ${f.description}` : f.title}>
              ⚠ <strong>{f.id}</strong> {f.title}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
