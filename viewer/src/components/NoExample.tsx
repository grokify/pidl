import type { Role, Step } from "../types";

// NoExample stands in for a message the walkthrough has no example for. It
// still names the step, so a response is never shown without its request.
export function NoExample({ step, roles }: { step: Step; roles: Role[] }) {
  const name = (id: string) => roles.find((r) => r.id === id)?.name ?? id;
  const isPerson = (id: string) => roles.find((r) => r.id === id)?.type === "user";
  const why = isPerson(step.from) || isPerson(step.to)
    ? "The person acts in a browser here, outside the protocol messages."
    : "No example message for this step.";
  return (
    <div className="no-example">
      <p className="message-step">
        Step {step.number}: {name(step.from)} {step.response ? "⇢" : "→"} {name(step.to)}
      </p>
      <p className="no-example-label">{step.label}</p>
      <p className="muted">
        {why}
        {step.description ? ` ${step.description}` : ""}
      </p>
    </div>
  );
}
