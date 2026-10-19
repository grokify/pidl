import { describe, expect, it } from "vitest";
import { tokenize } from "./highlight";

const kinds = (code: string, lang: string) => tokenize(code, lang).filter((t) => t.kind !== "plain").map((t) => [t.kind, t.text]);

describe("tokenize", () => {
  it("marks Go keywords, strings, numbers and comments", () => {
    expect(kinds('if err := f("a//b", 42); err != nil { // done', "go")).toEqual([
      ["keyword", "if"],
      ["keyword", "err"],
      ["string", '"a//b"'],
      ["number", "42"],
      ["keyword", "err"],
      ["keyword", "nil"],
      ["comment", "// done"],
    ]);
  });
  it("uses hash comments where the language does", () => {
    expect(kinds("x = 1  # note", "python")).toEqual([["number", "1"], ["comment", "# note"]]);
    expect(kinds("x = '#1'", "python")).toEqual([["string", "'#1'"]]);
    expect(kinds("# replayed\ncurl -X GET", "curl")).toEqual([["comment", "# replayed"], ["keyword", "curl"]]);
  });
  it("keeps the text intact", () => {
    const code = "func main() {\n\t/* block */ s := `raw\nstring`\n}";
    expect(tokenize(code, "go").map((t) => t.text).join("")).toBe(code);
  });
  it("treats unknown languages as plain apart from strings and comments", () => {
    expect(kinds('if "x" // c', "elixir")).toEqual([["string", '"x"'], ["comment", "// c"]]);
  });
});
