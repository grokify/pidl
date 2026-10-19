import type { Role, Scenario } from "./types";

// Diagram geometry, in SVG user units. Kept free of React so it can be
// unit tested.
export const HEADER_H = 64;
export const ROW_H = 54;
export const TOP_PAD = 20;
export const BOTTOM_PAD = 28;
export const MARGIN_X = 80;
export const BOX_W = 132;
const BAND_PAD_X = 58;
const BAND_INSET = 7;
// BAND_LABEL_H is the strip above a band's first step that holds the band's
// title, so the title never sits on the first message label.
export const BAND_LABEL_H = 22;

export interface Column {
  id: string;
  name: string;
  x: number;
}

export interface Row {
  index: number;
  y: number;
  x1: number;
  x2: number;
  self: boolean;
}

export interface Band {
  name: string;
  depth: number;
  first: number;
  last: number;
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface Layout {
  width: number;
  height: number;
  columns: Column[];
  rows: Row[];
  bands: Band[];
}

// rowY is a row's y before any band label strips are added.
export function rowY(index: number): number {
  return HEADER_H + TOP_PAD + index * ROW_H + ROW_H / 2;
}

// bandStarts counts, for each step, how many fragment bands begin there:
// a step that enters a fragment, or several nested ones at once.
function bandStarts(sc: Scenario): number[] {
  return sc.steps.map((s, i) => {
    const cur = s.fragments ?? [];
    const prev = sc.steps[i - 1]?.fragments ?? [];
    let shared = 0;
    while (shared < cur.length && shared < prev.length && cur[shared] === prev[shared]) shared++;
    return cur.length - shared;
  });
}

// layoutScenario places the scenario's participants as columns, in role
// order, and its steps as rows. Consecutive steps inside the same fragment
// form a band, with nested fragments inset.
export function layoutScenario(sc: Scenario, roles: Role[], width: number): Layout {
  const order = new Map(roles.map((r, i) => [r.id, i]));
  const ids = [...sc.participants].sort((a, b) => (order.get(a) ?? 1e9) - (order.get(b) ?? 1e9));
  const names = new Map(roles.map((r) => [r.id, r.name]));
  const span = width - 2 * MARGIN_X;
  const columns: Column[] = ids.map((id, i) => ({
    id,
    name: names.get(id) ?? id,
    x: ids.length === 1 ? width / 2 : MARGIN_X + (i * span) / (ids.length - 1),
  }));
  const xOf = new Map(columns.map((c) => [c.id, c.x]));
  const x = (id: string) => xOf.get(id) ?? width / 2;

  // Each band starting at a step pushes that step and the rest down by one
  // label strip.
  const starts = bandStarts(sc);
  let offset = 0;
  const rows: Row[] = sc.steps.map((s, i) => {
    offset += starts[i] * BAND_LABEL_H;
    return {
      index: i,
      y: rowY(i) + offset,
      x1: x(s.from),
      x2: x(s.to),
      self: s.from === s.to,
    };
  });

  const bands: Band[] = [];
  const maxDepth = Math.max(0, ...sc.steps.map((s) => s.fragments?.length ?? 0));
  for (let depth = 0; depth < maxDepth; depth++) {
    let start = -1;
    const key = (i: number) => (sc.steps[i]?.fragments ?? []).slice(0, depth + 1).join("\u0000");
    for (let i = 0; i <= sc.steps.length; i++) {
      const inBand = i < sc.steps.length && (sc.steps[i].fragments?.length ?? 0) > depth;
      if (start >= 0 && (!inBand || key(i) !== key(start))) {
        bands.push(makeBand(sc, rows, start, i - 1, depth, x, starts[start]));
        start = -1;
      }
      if (inBand && start < 0) {
        start = i;
      }
    }
  }

  const lastY = rows.length > 0 ? rows[rows.length - 1].y : rowY(0);
  return {
    width,
    height: lastY + ROW_H / 2 + BOTTOM_PAD,
    columns,
    rows,
    bands,
  };
}

// makeBand sizes a band around its steps. startsHere is how many bands
// begin at the band's first step; their label strips stack above that step,
// outermost first, so this band's strip is (startsHere - depthAmongThem) up.
function makeBand(
  sc: Scenario,
  rows: Row[],
  first: number,
  last: number,
  depth: number,
  x: (id: string) => number,
  startsHere: number,
): Band {
  let minX = Infinity;
  let maxX = -Infinity;
  for (let i = first; i <= last; i++) {
    for (const id of [sc.steps[i].from, sc.steps[i].to]) {
      minX = Math.min(minX, x(id));
      maxX = Math.max(maxX, x(id));
    }
  }
  const inset = depth * BAND_INSET;
  const left = minX - BAND_PAD_X + inset;
  const right = maxX + BAND_PAD_X - inset;
  // Bands that began at an earlier step and still enclose this one sit
  // outside this band's strips; count only those starting here.
  const enclosingDepth = (sc.steps[first].fragments?.length ?? 0) - startsHere;
  const strips = startsHere - (depth - enclosingDepth);
  const top = rows[first].y - ROW_H / 2 + 3 - strips * BAND_LABEL_H;
  const bottom = rows[last].y + ROW_H / 2 - 3 - inset;
  return {
    name: sc.steps[first].fragments?.[depth] ?? "",
    depth,
    first,
    last,
    x: left,
    y: top,
    width: right - left,
    height: bottom - top,
  };
}
