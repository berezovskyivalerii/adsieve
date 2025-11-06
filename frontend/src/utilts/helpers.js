export const classNames = (...a) => a.filter(Boolean).join(" ");
export const toDate = (s) => new Date(s + "T00:00:00");
export const fmtDDMMYYYY = (iso) => { const [y, m, d] = iso.split("-"); return `${d}.${m}.${y}`; };
export const getTodayDate = () => new Date().toISOString().split('T')[0];

export function currencyFmt(value, code) {
  try { return new Intl.NumberFormat(undefined, { style: "currency", currency: code, maximumFractionDigits: 2 }).format(value); }
  catch { return (Number(value) || 0).toFixed(2) + " " + code; }
}
export const convertAmount = (amountUSD, target, rates) => (rates?.rates?.[target] ?? 1) * amountUSD;