function AdsPage({ ads, onToggle, onView, t }) {
  return (
    <div className="grid gap-4">
      <Card title={t.Ads}>
        <div className="grid gap-2">
          {ads.map(a => (
            <div key={a.ad_id} className="flex items-center justify-between rounded-xl border px-3 py-2">
              <div>
                <button className="font-medium underline-offset-2 hover:underline" onClick={() => onView(a.ad_id)}>
                  {a.name}
                </button>
                <div className="text-xs text-zinc-500">ID {a.ad_id} · {a.platform}</div>
              </div>
              <div className="flex items-center gap-3">
                <button className="text-sm px-3 py-1.5 rounded-lg border hover:bg-zinc-50"
                  style={{ borderColor: BRAND.steel, color: BRAND.navy }}
                  onClick={() => onView(a.ad_id)}>
                  {t.ViewMetrics}
                </button>
                <Badge tone={a.status === "active" ? "navy" : "muted"}>{a.status === "active" ? t.Active : t.Paused}</Badge>
                <button onClick={() => onToggle(a.ad_id, a.status)} className="text-sm px-3 py-1.5 rounded-lg border hover:bg-zinc-50" style={{ borderColor: BRAND.steel, color: BRAND.navy }}>
                  {a.status === "active" ? t.Pause : t.Resume}
                </button>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
