import { clock, gapText, observedTiming } from "../observed";
import type { Role, Scenario } from "../types";

interface Props {
  scenario: Scenario;
  roles: Role[];
  step: number;
  // observed adds when the step's exchange happened and how it matched.
  observed?: boolean;
}

const STATUS_TEXT: Record<string, string> = {
  notSeen: "Not seen in this run",
  skipped: "Skipped in this run",
  unmatchable: "No rule to match this step",
  unobservable: "Not observable in a server's capture",
  extra: "Not a step in the document",
};

// ScenarioIntro is the prose introduction above the diagram.
export function ScenarioIntro({ scenario }: { scenario: Scenario }) {
  return (
    <div className="intro">{scenario.description && <p>{scenario.description}</p>}</div>
  );
}

// StepNarrative explains the active step below the diagram. It has a fixed
// height so the diagram does not move as narratives change length.
export function StepNarrative({ scenario, roles, step, observed = false }: Props) {
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
      {observed && <ObservedLine scenario={scenario} index={step - 1} />}
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

// ObservedLine says when the step's exchange happened, the gap from the
// previous one, and how well it matched.
function ObservedLine({ scenario, index }: { scenario: Scenario; index: number }) {
  const s = scenario.steps[index];
  const { time, gapMs } = observedTiming(scenario, index);
  const status = s.observed_status;
  return (
    <p className="observed-line">
      {time && (
        <span className="observed-time">
          Observed at {clock(time)}
          {gapMs !== undefined && ` · ${gapText(gapMs)} after the previous exchange`}
        </span>
      )}
      {status === "matched" && s.observed_strength === "weak" && (
        <span className="observed-chip weak" title="Only the participants and method agreed; the path, status, headers and body did not, or were not tested">
          Weak match
        </span>
      )}
      {status && status !== "matched" && <span className={`observed-chip status-${status}`}>{STATUS_TEXT[status] ?? status}</span>}
    </p>
  );
}
