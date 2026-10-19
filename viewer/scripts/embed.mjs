// Copies the single-file viewer build to render/viewer/viewer.html, where
// the Go renderer embeds it for `pidl generate -f viewer`.
import { copyFileSync, readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const src = resolve(here, "../dist/index.html");
const dst = resolve(here, "../../render/viewer/viewer.html");

const html = readFileSync(src, "utf8");
for (const marker of ['<script type="application/json" id="pidl-walkthrough"></script>', "<title>PIDL Walkthrough</title>"]) {
  if (!html.includes(marker)) {
    console.error(`viewer build is missing ${marker}`);
    process.exit(1);
  }
}
copyFileSync(src, dst);
console.log(`Wrote ${dst}`);
