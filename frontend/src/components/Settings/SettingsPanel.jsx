function SettingsPanel({ lang, setLang, currency, setCurrency, density, setDensity, dark, setDark, t }) {
  return (
    <div className="grid gap-4">
      <Card title={t.Language}><select className="rounded-xl border px-3 py-2 text-sm" value={lang} onChange={e => setLang(e.target.value)}><option value="ru">Русский</option><option value="en">English</option><option value="uk">Українська</option></select></Card>
      <Card title={t.Currency}><select className="rounded-xl border px-3 py-2 text-sm" value={currency} onChange={e => setCurrency(e.target.value)}>{CURRENCIES.map(c => <option key={c} value={c}>{c}</option>)}</select></Card>
      <Card title={t.Theme}><div className="flex items-center gap-3"><button className="rounded-xl px-3 py-2 border" style={{ borderColor: BRAND.steel }} onClick={() => setDark(false)}>{t.Light}</button><button className="rounded-xl px-3 py-2 border" style={{ borderColor: BRAND.steel, backgroundColor: dark ? BRAND.navy : undefined, color: dark ? 'white' : undefined }} onClick={() => setDark(true)}>{t.Dark}</button></div></Card>
      <Card title={t.Density}><div className="flex items-center gap-3"><button className="rounded-xl px-3 py-2 border" onClick={() => setDensity("comfortable")} style={{ borderColor: BRAND.steel }}>{t.Comfortable}</button><button className="rounded-xl px-3 py-2 border" onClick={() => setDensity("compact")} style={{ borderColor: BRAND.steel }}>{t.Compact}</button></div></Card>
    </div>
  );
}