import React from 'react';

export function Badge({ children, tone }) {
  const BRAND = { navy: "#02285c", red: "#b7291e" };
  const styles =
    tone === "red" ? { backgroundColor: BRAND.red + "20", color: BRAND.red } :
      tone === "navy" ? { backgroundColor: BRAND.navy + "20", color: BRAND.navy } :
        { backgroundColor: "#e5e7eb", color: "#374151" };
  return <span className="inline-flex items-center rounded-lg px-2.5 py-1 text-xs border" style={styles}>{children}</span>;
}