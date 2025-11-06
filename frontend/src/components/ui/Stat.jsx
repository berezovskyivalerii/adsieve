import React from "react";

function Stat({ label, value, help }) {
  return (
    <div className="rounded-2xl border bg-white/70 dark:bg-zinc-900/40 p-4 shadow-sm flex flex-col transition-all">
      <div className="text-xs uppercase tracking-wider text-zinc-500">{label}</div>
      <div className="text-2xl font-semibold mt-1" style={{ color: BRAND.navy }}>{value}</div>
      {help && <div className="text-xs text-zinc-500 mt-1">{help}</div>}
    </div>
  );
}