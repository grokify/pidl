import { useEffect, useState } from "react";
import { coveredComponents, splitURL, statusLine } from "../http";
import { CURL, HTTP, hasRep, repName } from "../reps";
import type { Example, Header, Role, Sample, Step, TokenDef } from "../types";
import { CodeView } from "./CodeView";
import { JsonView } from "./JsonView";
import { Truncated } from "./Truncated";

interface Props {
  kind: "request" | "response";
  step: Step;
  roles: Role[];
  tokens: TokenDef[];
  // reps are the representations offered as tabs; rep is the reader's
  // choice, kept across steps and panels.
  reps: string[];
  rep: string;
  onRep: (rep: string) => void;
}

// MessagePanel shows one side of an exchange in the chosen representation
// (HTTP, curl or code), and the decoded tokens it carries.
export function MessagePanel({ kind, step, roles, tokens, reps, rep, onRep }: Props) {
  const ex = step.example as Example;
  const hasHttp = hasRep(step, kind, HTTP);
  const tokenCount = ex.tokens?.length ?? 0;
  // The chosen representation; otherwise HTTP, or the first form this step
  // has, such as code for a step with no HTTP message.
  const shown = hasRep(step, kind, rep) ? rep : hasHttp ? HTTP : (reps.find((r) => hasRep(step, kind, r)) ?? null);
  const [tokensOpen, setTokensOpen] = useState(shown === null);
  useEffect(() => setTokensOpen(shown === null), [step, shown]);
  const tab = tokensOpen || shown === null ? "tokens" : shown;
  const name = (id: string) => roles.find((r) => r.id === id)?.name ?? id;

  return (
    <div className="message">
      <div className="message-meta">
        <span className="message-step">
          Step {step.number}: {name(step.from)} {step.response ? "⇢" : "→"} {name(step.to)}
        </span>
        <span className={`kind kind-${ex.kind}`} title={ex.kind === "draft" ? "Taken from the source, values aligned to this scenario" : "Written from the source's normative text"}>
          {ex.kind}
        </span>
        <span className="source">{ex.source}</span>
      </div>
      <div className="tabs" role="tablist">
        {reps.map((r) => {
          const available = hasRep(step, kind, r);
          return (
            <button
              key={r}
              role="tab"
              aria-selected={tab === r}
              className={tab === r ? "on" : ""}
              disabled={!available}
              title={available ? undefined : `No ${repName(r)} form for this ${kind}`}
              onClick={() => {
                onRep(r);
                setTokensOpen(false);
              }}
            >
              {repName(r)}
            </button>
          );
        })}
        <button role="tab" aria-selected={tab === "tokens"} className={`tokens-tab${tab === "tokens" ? " on" : ""}`} disabled={tokenCount === 0} onClick={() => setTokensOpen(true)}>
          Tokens{tokenCount > 0 ? ` (${tokenCount})` : ""}
        </button>
      </div>
      <div className="message-body">
        {tab === HTTP && kind === "request" && ex.request && <RequestView req={ex.request} />}
        {tab === HTTP && kind === "response" && ex.response && <ResponseView res={ex.response} />}
        {tab === CURL && step.curl && <CodeView code={step.curl} lang="curl" />}
        {tab !== HTTP && tab !== CURL && tab !== "tokens" && (
          <Samples samples={(ex.samples ?? []).filter((s) => s.language === tab && s.code)} name={name} />
        )}
        {tab === "tokens" &&
          ex.tokens?.map((t, i) => {
            const def = tokens.find((d) => d.id === t.token);
            const checks = new Map<string, string>();
            for (const c of def?.claims ?? []) {
              if (c.checked_by?.includes(step.to)) {
                checks.set(c.name, `checked by ${name(step.to)}${c.description ? `: ${c.description}` : ""}`);
              }
            }
            return (
              <div className="token" key={i}>
                <div className="token-head">
                  <strong>{def?.name ?? t.token}</strong>
                  {t.label && <span className="token-label">{t.label}</span>}
                  {checks.size > 0 && <span className="token-checks">{name(step.to)} checks {checks.size} claims</span>}
                </div>
                {t.header != null && (
                  <>
                    <div className="token-part">Header</div>
                    <JsonView value={t.header} />
                  </>
                )}
                {t.payload != null && (
                  <>
                    <div className="token-part">Claims</div>
                    <JsonView value={t.payload} highlight={checks} />
                  </>
                )}
              </div>
            );
          })}
      </div>
      {ex.note && <p className="message-note">{ex.note}</p>}
    </div>
  );
}

// Samples shows each code sample in one language, titled by its label and
// the role whose developer writes it.
function Samples({ samples, name }: { samples: Sample[]; name: (id: string) => string }) {
  return (
    <div className="samples">
      {samples.map((s, i) => {
        const parts = [s.label, s.role && name(s.role) !== s.label ? name(s.role) : undefined].filter(Boolean);
        return <CodeView key={i} code={s.code ?? ""} lang={s.language} title={parts.join(" · ") || undefined} />;
      })}
    </div>
  );
}

function HeaderLines({ headers, covered }: { headers: Header[]; covered: Set<string> }) {
  return (
    <>
      {headers.map((h, i) => (
        <div className="hline" key={i}>
          <span className="hname">{h.name}</span>
          <span className="hcolon">: </span>
          <span className="hvalue">
            <Truncated value={h.value} max={64} />
          </span>
          {covered.has(h.name.toLowerCase()) && <span className="signed" title="Covered by the HTTP message signature">signed</span>}
        </div>
      ))}
    </>
  );
}

function RequestView({ req }: { req: NonNullable<Example["request"]> }) {
  const { target, host } = splitURL(req.url);
  const covered = coveredComponents(req.headers);
  const lineSigned = covered.has("@method") || covered.has("@path");
  return (
    <div className="http">
      <div className="hline start">
        <span className="method">{req.method}</span> {target} <span className="proto">HTTP/1.1</span>
        {lineSigned && <span className="signed" title="@method and @path are covered by the signature">signed</span>}
      </div>
      {host && (
        <div className="hline">
          <span className="hname">Host</span>
          <span className="hcolon">: </span>
          <span className="hvalue">{host}</span>
          {covered.has("@authority") && <span className="signed" title="@authority is covered by the signature">signed</span>}
        </div>
      )}
      <HeaderLines headers={req.headers ?? []} covered={covered} />
      {req.body !== undefined && <JsonView value={req.body} />}
    </div>
  );
}

function ResponseView({ res }: { res: NonNullable<Example["response"]> }) {
  return (
    <div className="http">
      <div className={`hline start status-${Math.floor(res.status / 100)}xx`}>{statusLine(res.status)}</div>
      <HeaderLines headers={res.headers ?? []} covered={new Set()} />
      {res.body !== undefined && <JsonView value={res.body} />}
    </div>
  );
}
