import React from "react";

function TableHeader({ t }) {
  return (
    <div className="grid grid-cols-10 gap-3 text-xs font-medium text-zinc-500 px-3 py-2 bg-zinc-50 dark:bg-zinc-800/50">
      <div className="col-span-2">Date</div>
      <div className="col-span-2">Ad</div>
      <div>{t.Clicks}</div>
      <div>{t.Conversions}</div>
      <div>{t.Revenue}</div>
      <div>{t.Spend}</div>
      <div>{t.CPA}</div>
      <div>{t.ROAS}</div>
    </div>
  );
}