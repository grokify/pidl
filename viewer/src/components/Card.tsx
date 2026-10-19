import type { ReactNode } from "react";

interface Props {
  title: ReactNode;
  meta?: ReactNode;
  className?: string;
  children: ReactNode;
}

// Card is a bordered, padded section of the page with a title row.
export function Card({ title, meta, className = "", children }: Props) {
  return (
    <section className={`card ${className}`}>
      <div className="card-head">
        <h2>{title}</h2>
        {meta && <div className="card-meta">{meta}</div>}
      </div>
      <div className="card-body">{children}</div>
    </section>
  );
}
