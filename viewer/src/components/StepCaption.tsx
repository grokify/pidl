import type { Role, Scenario } from "../types";

interface Props {
  scenario: Scenario;
  roles: Role[];
  step: number;
}

// StepCaption describes the active step under the diagram, or the
// scenario itself on the overview.
export function StepCaption({ scenario, roles, step }: Props) {
  const name = (id: string) => roles.find((r) => r.id === id)?.name ?? id;
  if (step === 0) {
    return (
      <div className="caption">
        <p className="caption-title">{scenario.name}</p>
        {scenario.description && <p>{scenario.description}</p>}
        <p className="caption-hint">Use → or click a step to walk through it.</p>
      </div>
    );
  }
  const s = scenario.steps[step - 1];
  return (
    <div className="caption">
      <p className="caption-title">
        <span className="caption-num">{s.number}</span>
        {name(s.from)} <span className="caption-arrow">{s.response ? "⇢" : "→"}</span> {name(s.to)}: {s.label}
      </p>
      {s.fragments && s.fragments.length > 0 && <p className="caption-trail">Inside: {s.fragments.join(" › ")}</p>}
      {s.description && <p>{s.description}</p>}
      {s.note && <p className="caption-note">{s.note}</p>}
      {s.findings?.map((f) => (
        <p key={f.id + f.title} className={`finding severity-${f.severity}`}>
          <strong>{f.id}</strong> {f.title}
          {f.description && <span>: {f.description}</span>}
        </p>
      ))}
    </div>
  );
}
