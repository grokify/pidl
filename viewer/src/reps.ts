import type { Step, Walkthrough } from "./types";

// A representation is a form of a step's message: the HTTP exchange, an
// equivalent curl command, or code in a language.
export const HTTP = "http";
export const CURL = "curl";

const NAMES: Record<string, string> = {
  http: "HTTP",
  curl: "cURL",
  go: "Go",
  typescript: "TypeScript",
  javascript: "JavaScript",
  python: "Python",
  csharp: "C#",
  java: "Java",
  rust: "Rust",
  kotlin: "Kotlin",
  swift: "Swift",
  ruby: "Ruby",
  php: "PHP",
};

export function repName(rep: string): string {
  return NAMES[rep] ?? rep.charAt(0).toUpperCase() + rep.slice(1);
}

// representations lists HTTP, then cURL when any request has one, then
// each sample language in the order it first appears.
export function representations(w: Walkthrough): string[] {
  const out = [HTTP];
  let curl = false;
  const langs: string[] = [];
  for (const sc of w.scenarios) {
    for (const s of sc.samples ?? []) if (!langs.includes(s.language)) langs.push(s.language);
    for (const st of sc.steps) {
      if (st.curl) curl = true;
      for (const s of st.example?.samples ?? []) if (!langs.includes(s.language)) langs.push(s.language);
    }
  }
  if (curl) out.push(CURL);
  return out.concat(langs);
}

// hasRep reports whether a step's panel can show rep. kind is the panel the
// step is shown in: curl exists only for requests.
export function hasRep(step: Step, kind: "request" | "response", rep: string): boolean {
  const ex = step.example;
  if (!ex) return false;
  if (rep === HTTP) return kind === "request" ? Boolean(ex.request) : Boolean(ex.response);
  if (rep === CURL) return kind === "request" && Boolean(step.curl);
  return (ex.samples ?? []).some((s) => s.language === rep && Boolean(s.code));
}
