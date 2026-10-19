import { useState } from "react";

// Truncated shows a long value shortened, expanding on click.
export function Truncated({ value, max = 72 }: { value: string; max?: number }) {
  const [open, setOpen] = useState(false);
  if (value.length <= max) return <>{value}</>;
  return (
    <span className="truncated" onClick={() => setOpen(!open)} title={open ? "Click to shorten" : "Click to show the full value"}>
      {open ? value : `${value.slice(0, max - 20)}…${value.slice(-12)}`}
    </span>
  );
}
