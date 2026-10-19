import { Card } from "./components/Card";
import { SequenceDiagram } from "./components/SequenceDiagram";
import { StepCaption } from "./components/StepCaption";
import { TopBar } from "./components/TopBar";
import { usePosition } from "./hooks/usePosition";
import { useWidth } from "./hooks/useWidth";
import { exchangeFor } from "./nav";
import type { Walkthrough } from "./types";

export default function App({ walkthrough }: { walkthrough: Walkthrough }) {
  const { scenario, step, goTo, selectScenario } = usePosition(walkthrough);
  const { ref, width } = useWidth<HTMLDivElement>(900);
  const ex = step > 0 ? exchangeFor(scenario, step - 1) : null;
  const req = ex?.request != null ? scenario.steps[ex.request] : undefined;
  const res = ex?.response != null ? scenario.steps[ex.response] : undefined;

  return (
    <div className="app">
      <TopBar walkthrough={walkthrough} scenario={scenario} step={step} onScenario={selectScenario} onStep={goTo} />
      <main className="content">
        <Card title="Sequence" meta={scenario.fragment ? "Shared steps" : `${scenario.participants.length} roles`} className="diagram-card">
          <div className="diagram-wrap" ref={ref}>
            <SequenceDiagram scenario={scenario} roles={walkthrough.roles} width={width} active={step} onSelect={goTo} />
          </div>
          <StepCaption scenario={scenario} roles={walkthrough.roles} step={step} />
        </Card>
        <div className="side">
          <Card title="Request" className="message-card">
            {req?.example?.request ? (
              <code>
                {req.example.request.method} {req.example.request.url}
              </code>
            ) : (
              <p className="empty">{step === 0 ? "Step through the scenario to see each request." : "No HTTP request for this step."}</p>
            )}
          </Card>
          <Card title="Response" className="message-card">
            {res?.example?.response ? (
              <code>HTTP {res.example.response.status}</code>
            ) : (
              <p className="empty">
                {ex?.responseLater != null ? `The response arrives in step ${ex.responseLater + 1}.` : "No HTTP response for this step."}
              </p>
            )}
          </Card>
        </div>
      </main>
    </div>
  );
}
