import { useEffect, useState } from "react";
import { coveredComponents, splitURL, statusLine } from "../http";
import type { Example, Header, Role, Step, TokenDef } from "../types";
import { JsonView } from "./JsonView";
import { Truncated } from "./Truncated";

interface Props {
  kind: "request" | "response";
  step: Step;
  roles: Role[];
  tokens: TokenDef[];
}

// MessagePanel shows one side of an exchange: the HTTP message and the
// decoded tokens it carries.
export function MessagePanel({ kind, step, roles, tokens }: Props) {
  const ex = step.example as Example;
  const hasHttp = kind === "request" ? Boolean(ex.request) : Boolean(ex.response);
  const tokenCount = ex.tokens?.length ?? 0;
  const [tab, setTab] = useState<"http" | "tokens">(hasHttp ? "http" : "tokens");
  useEffect(() => setTab(hasHttp ? "http" : "tokens"), [step, hasHttp]);
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
        <button role="tab" aria-selected={tab === "http"} className={tab === "http" ? "on" : ""} disabled={!hasHttp} onClick={() => setTab("http")}>
          HTTP
        </button>
        <button role="tab" aria-selected={tab === "tokens"} className={tab === "tokens" ? "on" : ""} disabled={tokenCount === 0} onClick={() => setTab("tokens")}>
          Tokens{tokenCount > 0 ? ` (${tokenCount})` : ""}
        </button>
      </div>
      <div className="message-body">
        {tab === "http" && kind === "request" && ex.request && <RequestView req={ex.request} />}
        {tab === "http" && kind === "response" && ex.response && <ResponseView res={ex.response} />}
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
                {t.header && (
                  <>
                    <div className="token-part">Header</div>
                    <JsonView value={t.header} />
                  </>
                )}
                {t.payload && (
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
