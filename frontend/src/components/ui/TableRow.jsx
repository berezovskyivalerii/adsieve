import React from "react";

function TableRow({ row, ad, currency }) {
  const revenue = parseFloat(row.revenue) || 0;
  const spend = parseFloat(row.spend) || 0;
  const conversions = parseInt(row.conversions, 10) || 0;
  
  const CPA = parseFloat(row.cpa) || (conversions > 0 && spend > 0 ? spend / conversions : null);
  const ROAS = parseFloat(row.roas) || (spend > 0 ? revenue / spend : 0);

  return (
    <div className="grid grid-cols-10 gap-3 items-center px-3 py-2 rounded-xl hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors">
      <div className="col-span-2 text-sm">{fmtDDMMYYYY(row.date)}</div>
      <div className="col-span-2">
        <div className="text-sm font-medium">{ad?.name ?? row.ad_id}</div>
        <div className={classNames("text-[11px] mt-0.5", ad?.status === "active" ? "text-emerald-600" : "text-zinc-500")}>{ad?.status ?? ""}</div>
      </div>
      <div className="text-sm">{row.clicks}</div>
      <div className="text-sm">{row.conversions}</div>
      <div className="text-sm">{currencyFmt(revenue, currency)}</div>
      <div className="text-sm">{currencyFmt(spend, currency)}</div>
      <div className="text-sm">{CPA != null ? currencyFmt(CPA, currency) : "—"}</div>
      <div className="text-sm">{ROAS.toFixed(2)}×</div>
    </div>
  );
}