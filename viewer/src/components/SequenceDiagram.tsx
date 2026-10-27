import { useEffect, useMemo, useRef } from "react";
import { BOX_W, HEADER_H, ROW_H, layoutScenario } from "../layout";
import type { Role, Scenario } from "../types";

interface Props {
  scenario: Scenario;
  roles: Role[];
  width: number;
  // active is the 1-based active step; 0 shows the overview.
  active: number;
  onSelect: (step: number) => void;
  // observed marks each row with its observed status: extra exchanges
  // dashed, weak matches flagged, steps the run did not perform dimmed.
  observed?: boolean;
}

type RowState = "overview" | "past" | "active" | "future";

function rowState(index: number, active: number): RowState {
  if (active === 0) return "overview";
  if (index + 1 < active) return "past";
  if (index + 1 === active) return "active";
  return "future";
}

// observedClass names a row's observed status for styling.
function observedClass(status?: string, strength?: string): string {
  if (!status) return "";
  return `obs obs-${status}${strength === "weak" ? " obs-weak" : ""}`;
}

export function SequenceDiagram({ scenario, roles, width, active, onSelect, observed = false }: Props) {
  const layout = useMemo(() => layoutScenario(scenario, roles, width), [scenario, roles, width]);
  const scroller = useRef<HTMLDivElement>(null);
  const activeStep = active > 0 ? scenario.steps[active - 1] : undefined;
  const activeRoles = new Set(activeStep ? [activeStep.from, activeStep.to] : []);

  // Keep the active row in the middle of the scrolling area.
  useEffect(() => {
    const el = scroller.current;
    if (!el) return;
    if (active === 0) {
      el.scrollTo({ top: 0, behavior: "smooth" });
      return;
    }
    const y = layout.rows[active - 1].y - HEADER_H;
    el.scrollTo({ top: Math.max(0, y - el.clientHeight / 2), behavior: "smooth" });
  }, [active, layout]);

  const bodyHeight = layout.height - HEADER_H;
  const boxH = 40;

  return (
    <div className="diagram">
      <svg className="diagram-header" width={width} height={HEADER_H} role="img" aria-label="Roles">
        {layout.columns.map((c) => {
          const state = !activeStep ? "" : c.id === activeStep.from ? "sender" : c.id === activeStep.to ? "receiver" : "idle";
          return (
            <g key={c.id} className={`role ${state}`}>
              <rect x={c.x - BOX_W / 2} y={(HEADER_H - boxH) / 2} width={BOX_W} height={boxH} rx={8} />
              <text x={c.x} y={HEADER_H / 2 + 5} textAnchor="middle">
                {c.name}
              </text>
            </g>
          );
        })}
      </svg>
      <div className="diagram-body" ref={scroller}>
        <svg width={width} height={bodyHeight} viewBox={`0 ${HEADER_H} ${width} ${bodyHeight}`} role="img" aria-label={`${scenario.name} sequence`}>
          <defs>
            <marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="8" markerHeight="8" orient="auto-start-reverse">
              <path d="M 0 0 L 10 5 L 0 10 z" className="arrowhead" />
            </marker>
            <marker id="arrow-active" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="8" markerHeight="8" orient="auto-start-reverse">
              <path d="M 0 0 L 10 5 L 0 10 z" className="arrowhead active" />
            </marker>
          </defs>

          {layout.bands.map((b) => (
            <g key={`${b.depth}-${b.first}`} className={`band depth-${b.depth}`}>
              <rect x={b.x} y={b.y} width={b.width} height={b.height} rx={8} />
              <text x={b.x + 10} y={b.y + 15}>
                {b.name}
              </text>
            </g>
          ))}

          {layout.columns.map((c) => (
            <line
              key={c.id}
              className={`lifeline ${activeRoles.has(c.id) ? "active" : ""}`}
              x1={c.x}
              x2={c.x}
              y1={HEADER_H}
              y2={layout.height}
            />
          ))}

          {layout.rows.map((r) => {
            const s = scenario.steps[r.index];
            const state = rowState(r.index, active);
            const marker = state === "active" ? "url(#arrow-active)" : "url(#arrow)";
            const labelX = r.self ? r.x1 + 46 : (r.x1 + r.x2) / 2;
            const findings = s.findings?.length ?? 0;
            const obs = observed ? observedClass(s.observed_status, s.observed_strength) : s.observed_status === "extra" ? "obs obs-extra" : "";
            const repeats = observed && s.observed?.repeats ? ` ×${s.observed.repeats}` : "";
            return (
              <g
                key={r.index}
                className={`row ${state} ${s.response ? "response" : ""} ${obs}`}
                onClick={() => onSelect(r.index + 1)}
                role="button"
                aria-label={`Step ${s.number}: ${s.label}`}
              >
                <rect className="hit" x={0} y={r.y - ROW_H / 2} width={width} height={ROW_H} />
                {r.self ? (
                  <path className="message" d={`M ${r.x1} ${r.y - 10} h 34 v 20 h -32`} markerEnd={marker} />
                ) : (
                  <line className="message" x1={r.x1} y1={r.y} x2={r.x2} y2={r.y} markerEnd={marker} />
                )}
                <text className="label" x={labelX} y={r.y - 8} textAnchor={r.self ? "start" : "middle"}>
                  {s.label}
                  {repeats}
                </text>
                <g className="badge" transform={`translate(${r.x1 + (r.x2 >= r.x1 ? 14 : -14)}, ${r.y + 13})`}>
                  <circle r={9} />
                  <text y={4} textAnchor="middle">
                    {s.number}
                  </text>
                </g>
                {findings > 0 && (
                  <g className="finding-mark" transform={`translate(${labelX + (r.self ? -12 : 0)}, ${r.y + 12})`}>
                    <title>{s.findings!.map((f) => `${f.id}: ${f.title}`).join("\n")}</title>
                    <circle r={7} />
                    <text y={4} textAnchor="middle">
                      !
                    </text>
                  </g>
                )}
              </g>
            );
          })}
        </svg>
      </div>
    </div>
  );
}
