import { useEffect, useState, type ReactNode } from "react";

export const STAGE_W = 1600;
export const STAGE_H = 900;

// stageScale fits a 16:9 stage inside a window, never above 2x.
export function stageScale(width: number, height: number): number {
  return Math.min(width / STAGE_W, height / STAGE_H, 2);
}

// Stage renders its children on a fixed 1600x900 canvas scaled to fit the
// window, letterboxed, for presenting. When off, children fill the window.
export function Stage({ on, children }: { on: boolean; children: ReactNode }) {
  const [scale, setScale] = useState(() => stageScale(window.innerWidth, window.innerHeight));
  useEffect(() => {
    const onResize = () => setScale(stageScale(window.innerWidth, window.innerHeight));
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);
  if (!on) return <>{children}</>;
  return (
    <div className="stage-viewport">
      <div className="stage" style={{ width: STAGE_W, height: STAGE_H, transform: `scale(${scale})` }}>
        {children}
      </div>
    </div>
  );
}
