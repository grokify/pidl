import type { Walkthrough } from "./types";

export default function App({ walkthrough }: { walkthrough: Walkthrough }) {
  return (
    <main>
      <h1>{walkthrough.name}</h1>
      <ol>
        {walkthrough.scenarios.map((s) => (
          <li key={s.id}>
            {s.name} ({s.steps.length} steps)
          </li>
        ))}
      </ol>
    </main>
  );
}
