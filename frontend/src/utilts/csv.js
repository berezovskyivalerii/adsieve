function csvString(rows) {
  if (rows.length === 0) return "";
  const headers = Object.keys(rows[0]);
  const lines = [headers.join(",")].concat(rows.map(r => headers.map(h => JSON.stringify(r[h])).join(",")));
  return lines.join("\n");
}

export function exportCSV(rows) {
  const blob = new Blob([csvString(rows)], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a'); a.href = url; a.download = 'metrics.csv'; a.click(); URL.revokeObjectURL(url);
}