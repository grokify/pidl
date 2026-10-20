import { useCallback, useEffect, useState } from "react";
import { isTyping } from "../host";

export type Theme = "auto" | "light" | "dark";
const THEMES: Theme[] = ["auto", "light", "dark"];

// Preferences are a per-viewer convenience; storage can be unavailable
// (private windows, file:// pages), so every access is guarded.
function load<T extends string>(key: string, allowed: readonly T[], fallback: T): T {
  try {
    const v = window.localStorage.getItem(key) as T | null;
    return v && allowed.includes(v) ? v : fallback;
  } catch {
    return fallback;
  }
}

function save(key: string, value: string) {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // Storage is optional; the preference simply is not remembered.
  }
}

// loadRep reads the remembered representation, such as "http" or "go".
function loadRep(): string {
  try {
    const v = window.localStorage.getItem("pidl.rep");
    return v && /^[a-z][a-z0-9+#-]*$/.test(v) ? v : "http";
  } catch {
    return "http";
  }
}

// usePrefs holds the viewer's remembered preferences and handles the
// present key sent to keyTarget. With page, the theme also applies to the
// page itself, for the standalone viewer.
export function usePrefs(keyTarget: EventTarget | null, page: boolean) {
  const [rep, setRepState] = useState<string>(loadRep);
  const setRep = useCallback((r: string) => {
    setRepState(r);
    save("pidl.rep", r);
  }, []);
  const [theme, setTheme] = useState<Theme>(() => load("pidl.theme", THEMES, "auto"));
  const [present, setPresent] = useState<boolean>(() => load("pidl.present", ["on", "off"] as const, "off") === "on");

  useEffect(() => {
    save("pidl.theme", theme);
    if (!page) return;
    const root = document.documentElement;
    if (theme === "auto") delete root.dataset.theme;
    else root.dataset.theme = theme;
  }, [theme, page]);

  useEffect(() => save("pidl.present", present ? "on" : "off"), [present]);

  const cycleTheme = useCallback(() => setTheme((t) => THEMES[(THEMES.indexOf(t) + 1) % THEMES.length]), []);
  const togglePresent = useCallback(() => setPresent((p) => !p), []);

  useEffect(() => {
    if (!keyTarget) return;
    const onKey = (e: KeyboardEvent) => {
      if (isTyping(e)) return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (e.key === "p" || e.key === "P") togglePresent();
    };
    const listener = onKey as EventListener;
    keyTarget.addEventListener("keydown", listener);
    return () => keyTarget.removeEventListener("keydown", listener);
  }, [keyTarget, togglePresent]);

  return { theme, cycleTheme, present, togglePresent, rep, setRep };
}
