// A small, dependency-free highlighter: comments, strings, numbers and
// keywords, which is enough to make short samples readable.

export type TokenKind = "plain" | "comment" | "string" | "number" | "keyword";
export interface CodeToken {
  kind: TokenKind;
  text: string;
}

const KEYWORDS: Record<string, string[]> = {
  go: ["break", "case", "chan", "const", "continue", "default", "defer", "else", "fallthrough", "for", "func", "go", "goto", "if", "import", "interface", "map", "package", "range", "return", "select", "struct", "switch", "type", "var", "nil", "true", "false", "err"],
  typescript: ["async", "await", "break", "case", "catch", "class", "const", "continue", "default", "else", "export", "extends", "false", "finally", "for", "from", "function", "if", "import", "interface", "let", "new", "null", "return", "switch", "this", "throw", "true", "try", "type", "undefined", "while"],
  python: ["and", "as", "async", "await", "class", "def", "elif", "else", "except", "False", "finally", "for", "from", "if", "import", "in", "is", "lambda", "None", "not", "or", "raise", "return", "True", "try", "while", "with", "yield"],
  csharp: ["async", "await", "class", "const", "else", "false", "for", "foreach", "if", "in", "namespace", "new", "null", "private", "public", "return", "static", "string", "true", "try", "catch", "using", "var", "void"],
  curl: ["curl"],
};
KEYWORDS.javascript = KEYWORDS.typescript;

const HASH_COMMENTS = new Set(["python", "curl", "ruby", "shell", "bash"]);

export function tokenize(code: string, lang: string): CodeToken[] {
  const keywords = new Set(KEYWORDS[lang] ?? []);
  const hash = HASH_COMMENTS.has(lang);
  const out: CodeToken[] = [];
  const push = (kind: TokenKind, text: string) => {
    const last = out[out.length - 1];
    if (last && last.kind === kind) last.text += text;
    else out.push({ kind, text });
  };
  let i = 0;
  while (i < code.length) {
    const rest = code.slice(i);
    let m: RegExpMatchArray | null;
    if ((hash && rest[0] === "#") || rest.startsWith("//")) {
      const end = code.indexOf("\n", i);
      const stop = end < 0 ? code.length : end;
      push("comment", code.slice(i, stop));
      i = stop;
    } else if (rest.startsWith("/*")) {
      const end = code.indexOf("*/", i + 2);
      const stop = end < 0 ? code.length : end + 2;
      push("comment", code.slice(i, stop));
      i = stop;
    } else if ((m = rest.match(/^("(?:\\.|[^"\\\n])*"|'(?:\\.|[^'\\\n])*'|`[^`]*`)/))) {
      push("string", m[0]);
      i += m[0].length;
    } else if ((m = rest.match(/^\d[\d_.xXa-fA-F]*/))) {
      push("number", m[0]);
      i += m[0].length;
    } else if ((m = rest.match(/^[A-Za-z_$][\w$]*/))) {
      push(keywords.has(m[0]) ? "keyword" : "plain", m[0]);
      i += m[0].length;
    } else {
      push("plain", code[i]);
      i++;
    }
  }
  return out;
}
