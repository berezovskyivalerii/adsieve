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
