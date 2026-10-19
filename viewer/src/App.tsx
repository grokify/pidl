import { useState } from "react";
import { SequenceDiagram } from "./components/SequenceDiagram";
import type { Walkthrough } from "./types";

export default function App({ walkthrough }: { walkthrough: Walkthrough }) {
  const scenario = walkthrough.scenarios[0];
  const [step, setStep] = useState(0);
  return (
    <main>
      <h1>
        {walkthrough.name}: {scenario.name}
      </h1>
      <SequenceDiagram scenario={scenario} roles={walkthrough.roles} width={960} active={step} onSelect={setStep} />
    </main>
  );
}
