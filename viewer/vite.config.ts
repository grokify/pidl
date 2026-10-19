import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { viteSingleFile } from "vite-plugin-singlefile";

// The build is one self-contained HTML file: scripts, styles and fonts
// inlined, so `pidl generate -f viewer` can embed it and add the data.
export default defineConfig({
  base: "./",
  plugins: [react(), viteSingleFile()],
  build: { outDir: "dist", emptyOutDir: true, assetsInlineLimit: 100_000_000 },
  test: { environment: "node" },
});
