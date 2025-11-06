import React from "react";

function DateRange({ from, to, setFrom, setTo, t }) {
  return (
    <div className="flex items-center gap-2">
      <input type="date" className="rounded-xl border px-3 py-2 text-sm" value={from} onChange={e => setFrom(e.target.value)} />
      <span className="text-xs text-zinc-500">{t.DateTo}</span>
      <input type="date" className="rounded-xl border px-3 py-2 text-sm" value={to} onChange={e => setTo(e.target.value)} />
    </div>
  );
}