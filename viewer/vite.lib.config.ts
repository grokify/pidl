import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// The embeddable viewer: one script with React and the styles bundled, as
// an IIFE (global PidlViewer, for a plain script tag) and an ES module.
export default defineConfig({
  plugins: [react()],
  define: { "process.env.NODE_ENV": JSON.stringify("production") },
  build: {
    outDir: "dist-lib",
    emptyOutDir: true,
    lib: {
      entry: "src/lib.tsx",
      name: "PidlViewer",
      formats: ["iife", "es"],
      fileName: (format) => (format === "es" ? "pidl-viewer.mjs" : "pidl-viewer.js"),
    },
  },
});
