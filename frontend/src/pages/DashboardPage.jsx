import React from 'react';
import { LineChart, Line, CartesianGrid, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer, BarChart, Bar } from "recharts";
import { BRAND } from '../lib/constants';
import { fmtDDMMYYYY, currencyFmt, classNames } from '../utils/helpers';
import { Card } from '../components/ui/Card';
import { DateRange } from '../components/ui/DateRange';

function Stat({ label, value, help }) {
  return (
    <div className="rounded-2xl border bg-white/70 dark:bg-zinc-900/40 p-4 shadow-sm flex flex-col transition-all">
      <div className="text-xs uppercase tracking-wider text-zinc-500">{label}</div>
      <div className="text-2xl font-semibold mt-1" style={{ color: BRAND.navy }}>{value}</div>
      {help && <div className="text-xs text-zinc-500 mt-1">{help}</div>}
    </div>
  );
}

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

function Charts({ rows }) {
  const series = rows.reduce((acc, r) => {
    const idx = acc.findIndex((x) => x.date === r.date);
    const add = { 
      clicks: parseInt(r.clicks, 10) || 0, 
      conv: parseInt(r.conversions, 10) || 0, 
      revenue: parseFloat(r.revenue) || 0, 
      spend: parseFloat(r.spend) || 0 
    };
    if (idx >= 0) { 
      acc[idx] = { 
        ...acc[idx], 
        clicks: acc[idx].clicks + add.clicks, 
        conv: acc[idx].conv + add.conv, 
        revenue: acc[idx].revenue + add.revenue, 
        spend: acc[idx].spend + add.spend 
      } 
    }
    else { acc.push({ date: r.date, ...add }); }
    return acc;
  }, []);
  
  return (
    <div className="grid md:grid-cols-2 gap-4">
      <Card title="Clicks vs Conversions">
        <div style={{ width: "100%", height: 260 }}>
          <ResponsiveContainer>
            <LineChart data={series} margin={{ left: 12, right: 12, top: 8, bottom: 8 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" tickFormatter={fmtDDMMYYYY} />
              <YAxis />
              <Tooltip labelFormatter={(v) => fmtDDMMYYYY(v)} />
              <Legend />
              <Line type="monotone" dataKey="clicks" stroke={BRAND.navy} dot={false} />
              <Line type="monotone" dataKey="conv" stroke={BRAND.red} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Card>
      <Card title="Revenue vs Spend">
        <div style={{ width: "100%", height: 260 }}>
          <ResponsiveContainer>
            <BarChart data={series} margin={{ left: 12, right: 12, top: 8, bottom: 8 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" tickFormatter={fmtDDMMYYYY} />
              <YAxis />
              <Tooltip labelFormatter={(v) => fmtDDMMYYYY(v)} />
              <Legend />
              <Bar dataKey="revenue" fill={BRAND.navy} />
              <Bar dataKey="spend" fill={BRAND.red} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Card>
    </div>
  );
}

export default function DashboardPage({
  t,
  error,
  aggUSD,
  aggRevenue,
  aggSpend,
  CPA,
  ROAS,
  currency,
  convertedRows,
  handleSyncToday,
  from, to, setFrom, setTo,
  adFilter, setAdFilter,
  ads,
  exportCSV,
  loading,
  density
}) {
  return (
    <div className="grid gap-4">
      {error && <p className="text-red-500 bg-red-100 p-3 rounded-xl">{error}</p>}
      <section className="grid md:grid-cols-3 gap-3">
        <Stat label={t.Clicks} value={aggUSD.clicks.toString()} help={t.SumInRange} />
        <Stat label={t.Conversions} value={aggUSD.conversions.toString()} help={t.SumInRange} />
        <Stat label={t.Revenue} value={currencyFmt(aggRevenue, currency)} help={t.SumInRange} />
        <Stat label={t.Spend} value={currencyFmt(aggSpend, currency)} help={t.SumInRange} />
        <Stat label={t.CPA} value={CPA > 0 ? currencyFmt(CPA, currency) : '—'} help="Spend / Conversions" />
        <Stat label={t.ROAS} value={ROAS.toFixed(2) + '×'} help="Revenue / Spend" />
      </section>
      
      <Charts rows={convertedRows} />
      
      <Card title={t.MetricsTable} right={
          <div className="flex items-center gap-3">
            <button className="rounded-xl border px-3 py-2 text-sm" onClick={handleSyncToday} style={{ borderColor: BRAND.steel, color: BRAND.navy }}>{t.SyncToday}</button>
            <DateRange from={from} to={to} setFrom={setFrom} setTo={setTo} t={t} />
            <select className="rounded-xl border px-3 py-2 text-sm" value={adFilter} onChange={e => setAdFilter(e.target.value)}>
              <option value="all">{t.AllAds}</option>
              {ads.map(a => <option key={a.ad_id} value={a.ad_id}>{a.name}</option>)}
            </select>
            <button className="rounded-xl border px-3 py-2 text-sm" onClick={() => exportCSV(convertedRows)} style={{ borderColor: BRAND.steel }}>{t.ExportCSV}</button>
          </div>
        }>
        <div className={classNames("rounded-xl border overflow-hidden", density === 'compact' && 'text-[13px]')}>
          <TableHeader t={t} />
          {loading ? (
             <div className="p-4 text-center">Загрузка метрик...</div>
          ) : (
            <div className="divide-y">{convertedRows.map((r, i) => <TableRow key={i + "/" + r.ad_id} row={r} ad={ads.find(a => a.ad_id === r.ad_id)} currency={currency} />)}</div>
          )}
        </div>
      </Card>
    </div>
  );
}