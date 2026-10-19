import type { ReactNode } from "react";
import { Truncated } from "./Truncated";

interface Props {
  value: unknown;
  // highlight maps a top-level key to a note, such as who checks a claim.
  highlight?: Map<string, string>;
}

// JsonView pretty-prints JSON in key order, shortening long strings such
// as JWTs, and can highlight top-level keys.
export function JsonView({ value, highlight }: Props) {
  return <pre className="json">{render(value, 0, highlight)}</pre>;
}

function render(v: unknown, depth: number, highlight?: Map<string, string>): ReactNode {
  const pad = "  ".repeat(depth);
  const inner = "  ".repeat(depth + 1);
  if (Array.isArray(v)) {
    if (v.length === 0) return <span className="j-punct">[]</span>;
    return (
      <>
        <span className="j-punct">[</span>
        {"\n"}
        {v.map((item, i) => (
          <span key={i}>
            {inner}
            {render(item, depth + 1)}
            {i < v.length - 1 ? <span className="j-punct">,</span> : null}
            {"\n"}
          </span>
        ))}
        {pad}
        <span className="j-punct">]</span>
      </>
    );
  }
  if (v !== null && typeof v === "object") {
    const entries = Object.entries(v as Record<string, unknown>);
    if (entries.length === 0) return <span className="j-punct">{"{}"}</span>;
    return (
      <>
        <span className="j-punct">{"{"}</span>
        {"\n"}
        {entries.map(([k, item], i) => {
          const note = depth === 0 ? highlight?.get(k) : undefined;
          return (
            <span key={k} className={note ? "j-hl" : undefined} title={note}>
              {inner}
              <span className="j-key">"{k}"</span>
              <span className="j-punct">: </span>
              {render(item, depth + 1)}
              {i < entries.length - 1 ? <span className="j-punct">,</span> : null}
              {note ? <span className="j-note">  ← {note}</span> : null}
              {"\n"}
            </span>
          );
        })}
        {pad}
        <span className="j-punct">{"}"}</span>
      </>
    );
  }
  if (typeof v === "string") {
    return (
      <span className="j-str">
        "<Truncated value={v} />"
      </span>
    );
  }
  if (typeof v === "number" || typeof v === "boolean") return <span className="j-num">{String(v)}</span>;
  return <span className="j-num">null</span>;
}
