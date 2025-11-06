import { useState, useEffect } from 'react';

export function useTheme() {
  const [dark, setDark] = useState(false);
  useEffect(() => {
    const saved = localStorage.getItem('adsieve_theme');
    setDark(saved ? saved === 'dark' : false);
  }, []);
  useEffect(() => {
    const root = document.documentElement;
    if (dark) root.classList.add("dark"); else root.classList.remove("dark");
    localStorage.setItem('adsieve_theme', dark ? 'dark' : 'light');
  }, [dark]);
  return { dark, setDark };
}