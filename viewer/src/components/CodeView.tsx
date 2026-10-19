import { useState } from "react";
import { tokenize } from "../highlight";

// CodeView shows code with light highlighting and a copy button.
export function CodeView({ code, lang, title }: { code: string; lang: string; title?: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard
      ?.writeText(code)
      .then(() => {
        setCopied(true);
        window.setTimeout(() => setCopied(false), 1500);
      })
      .catch(() => setCopied(false));
  };
  return (
    <div className="code-block">
      <div className="code-head">
        {title && <span className="code-title">{title}</span>}
        <button className="copy" onClick={copy} title="Copy to clipboard">
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <pre className="code">
        <code>
          {tokenize(code, lang).map((t, i) =>
            t.kind === "plain" ? t.text : (
              <span key={i} className={`tok-${t.kind}`}>
                {t.text}
              </span>
            ),
          )}
        </code>
      </pre>
    </div>
  );
}
