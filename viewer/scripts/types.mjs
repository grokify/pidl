// Generates src/walkthrough.gen.ts from the walkthrough JSON Schema, which
// is itself generated from the Go types (schema/gen). Run: npm run types.
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { compile } from "json-schema-to-typescript";

const here = dirname(fileURLToPath(import.meta.url));
const schemaPath = resolve(here, "../../schema/walkthrough.schema.json");
const out = resolve(here, "../src/walkthrough.gen.ts");

const file = JSON.parse(readFileSync(schemaPath, "utf8"));
// The file's root is a $ref to the Walkthrough definition; compile that
// definition as the root so it is emitted as an interface.
const { Walkthrough: root, ...defs } = file.$defs;
const schema = { ...root, title: "Walkthrough", description: file.description, $defs: defs };

// dedupe adjusts the schema for TypeScript. A property that references a definition and also carries its own
// description would become a duplicate type (Narration1, Narration2, ...).
// Keep the definition's type and move the description onto the property's
// doc comment only.
function dedupe(node) {
  if (Array.isArray(node)) {
    node.forEach(dedupe);
    return;
  }
  if (node === null || typeof node !== "object") {
    return;
  }
  if (typeof node.$ref === "string" && node.description !== undefined) {
    node.tsType = node.$ref.replace("#/$defs/", "");
    delete node.$ref;
  }
  for (const [key, value] of Object.entries(node)) {
    // A property schema with no constraints (true, or only a description)
    // accepts any JSON value, such as a body that may be text or JSON.
    if (key === "properties") {
      for (const [name, prop] of Object.entries(value)) {
        if (prop === true) {
          value[name] = { tsType: "unknown" };
        } else if (Object.keys(prop).every((k) => k === "description")) {
          prop.tsType = "unknown";
        }
      }
    }
    dedupe(value);
  }
}
dedupe(schema);

const ts = await compile(schema, "Walkthrough", {
  additionalProperties: false,
  unreachableDefinitions: true,
  bannerComment: `// Code generated from schema/walkthrough.schema.json by scripts/types.mjs. DO NOT EDIT.
// Regenerate with: npm run types`,
});
writeFileSync(out, ts);
console.log(`Wrote ${out}`);
