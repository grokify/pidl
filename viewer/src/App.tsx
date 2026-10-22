import { useState } from "react";
import { Card } from "./components/Card";
import { Stage } from "./components/Stage";
import { MessagePanel } from "./components/MessagePanel";
import { NoExample } from "./components/NoExample";
import { SequenceDiagram } from "./components/SequenceDiagram";
import { ScenarioIntro, StepNarrative } from "./components/StepCaption";
import { StepControls } from "./components/StepControls";
import { TopBar } from "./components/TopBar";
import { usePosition } from "./hooks/usePosition";
import { usePrefs } from "./hooks/usePrefs";
import { useWidth } from "./hooks/useWidth";
import { PAGE_HOST, type Host } from "./host";
import { exchangeFor } from "./nav";
import { representations } from "./reps";
import type { Walkthrough } from "./types";

export default function App({ walkthrough, host = PAGE_HOST }: { walkthrough: Walkthrough; host?: Host }) {
  // An embedded viewer takes keys only while it has focus.
  const [root, setRoot] = useState<HTMLDivElement | null>(null);
  const keyTarget = host.keys === "page" ? window : root;
  const { scenario, step, goTo, selectScenario } = usePosition(walkthrough, host.hash, keyTarget);
  const { ref, width } = useWidth<HTMLDivElement>(900);
  const prefs = usePrefs(keyTarget, host.page);
  const reps = representations(walkthrough);
  const rep = reps.includes(prefs.rep) ? prefs.rep : reps[0];
  const repProps = { reps, rep, onRep: prefs.setRep };
  const ex = step > 0 ? exchangeFor(scenario, step - 1) : null;
  const req = ex?.request != null ? scenario.steps[ex.request] : undefined;
  const res = ex?.response != null ? scenario.steps[ex.response] : undefined;
  const active = step > 0 ? scenario.steps[step - 1] : undefined;
  const isPerson = (id: string) => walkthrough.roles.find((r) => r.id === id)?.type === "user";
  const involvesPerson = active ? isPerson(active.from) || isPerson(active.to) : false;

  return (
    <div
      className="pidl-viewer"
      ref={setRoot}
      data-theme={prefs.theme === "auto" ? undefined : prefs.theme}
      tabIndex={host.keys === "element" ? 0 : undefined}
    >
    <Stage on={prefs.present}>
    <div className="app">
      <TopBar
        walkthrough={walkthrough}
        source={host.source}
        scenario={scenario}
        onScenario={selectScenario}
        theme={prefs.theme}
        onTheme={prefs.cycleTheme}
        present={prefs.present}
        onPresent={prefs.togglePresent}
      />
      <main className="content">
        <Card title={scenario.name} meta={scenario.fragment ? "Shared steps" : `${scenario.participants.length} roles · ${scenario.steps.length} steps`} className="diagram-card">
          <ScenarioIntro scenario={scenario} />
          <div className="diagram-wrap" ref={ref}>
            <SequenceDiagram scenario={scenario} roles={walkthrough.roles} width={width} active={step} onSelect={goTo} />
          </div>
          <StepNarrative scenario={scenario} roles={walkthrough.roles} step={step} />
          <StepControls step={step} total={scenario.steps.length} onStep={goTo} />
        </Card>
        <div className="side">
          <Card title="Request" className="message-card">
            {req ? (
              req.example ? (
                <MessagePanel kind="request" step={req} roles={walkthrough.roles} tokens={walkthrough.tokens ?? []} {...repProps} />
              ) : (
                <NoExample step={req} roles={walkthrough.roles} />
              )
            ) : (
              <p className="empty">
                {step === 0
                  ? "Step through the scenario to see each request."
                  : "This response answers a request made before this scenario's steps begin."}
              </p>
            )}
          </Card>
          <Card title="Response" className="message-card">
            {res ? (
              res.example ? (
                <MessagePanel kind="response" step={res} roles={walkthrough.roles} tokens={walkthrough.tokens ?? []} {...repProps} />
              ) : (
                <NoExample step={res} roles={walkthrough.roles} />
              )
            ) : (
              <p className="empty">
                {ex?.responseLater != null
                  ? `The response arrives in step ${ex.responseLater + 1}.`
                  : step === 0
                    ? "Responses appear as you step through."
                    : involvesPerson
                    ? "The person acts in a browser here; there is no protocol response to show."
                    : "This step is not answered directly; the scenario continues with the next step."}
              </p>
            )}
          </Card>
        </div>
      </main>
    </div>
    </Stage>
    </div>
  );
}
