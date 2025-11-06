import React from 'react';

export function Card({ title, children, right }) {
  return (
    <section className="rounded-2xl border bg-white/80 dark:bg-zinc-900/50 p-4 shadow-sm transition-all">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold" style={{ color: "#5b6b89" }}>{title}</h3>
        {right}
      </div>
      {children}
    </section>
  );
}