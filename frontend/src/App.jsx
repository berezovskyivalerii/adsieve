import React, { useEffect, useMemo, useState, useCallback } from "react";
import { useNavigate, useLocation } from "react-router-dom";

import API, { signup, login, setAuthHeader, clearAuthHeader } from "./api/auth";
import { getMetrics, getAds, pauseAd, resumeAd, connectGoogle, getGoogleAccounts, linkGoogleAccounts, syncGoogle } from "./api/data";

import { I18N } from './lib/i18n';
import { BRAND, CURRENCIES, DEFAULT_RATES } from './lib/constants';
import { useTheme } from './hooks/useTheme';
import { classNames, toDate, getTodayDate, convertAmount, currencyFmt } from './utils/helpers';
import { exportCSV } from './utils/csv';

import { Badge } from './components/ui/Badge';
import SettingsDrawer from './components/settings/SettingsDrawer'; 
import SettingsPanel from './components/settings/SettingsPanel'; 

import LandingPage from './pages/LandingPage';
import AuthLoginPage from './pages/auth/AuthLoginPage';
import AuthSignUpPage from './pages/auth/AuthSignUpPage';
import AuthForgotPage from './pages/auth/AuthForgotPage';
import DashboardPage from './pages/DashboardPage';
import AdsPage from './pages/AdsPage';
import IntegrationsPage from './pages/IntegrationsPage';
import DocsPage from './pages/DocsPage';


export default function AdSieveUI() {
  const { dark, setDark } = useTheme();
  const [lang, setLang] = useState('ru');
  const t = I18N[lang];

  const navigate = useNavigate();
  const location = useLocation();
  const pathFor = {
    Landing: '/',
    Login: '/login',
    SignUp: '/signup',
    Forgot: '/forgot',
    Dashboard: '/app/dashboard',
    Ads: '/app/ads',
    Integrations: '/app/integrations',
    Docs: '/app/docs',
  };
  function pageForPath(pathname) {
    if (pathname === '/') return 'Landing';
    if (pathname.startsWith('/login')) return 'Login';
    if (pathname.startsWith('/signup')) return 'SignUp';
    if (pathname.startsWith('/forgot')) return 'Forgot';
    if (pathname.startsWith('/app/dashboard')) return 'Dashboard';
    if (pathname.startsWith('/app/ads')) return 'Ads';
    if (pathname.startsWith('/app/integrations')) return 'Integrations';
    if (pathname.startsWith('/app/docs')) return 'Docs';
    if (pathname.startsWith('/app')) return 'Dashboard'; 
    return 'Landing';
  }

  const [page, setPage] = useState(pageForPath(location.pathname));
  const [settingsOpen, setSettingsOpen] = useState(false);
  
  const [ads, setAds] = useState([]);
  const [metrics, setMetrics] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [adFilter, setAdFilter] = useState('all');
  const [from, setFrom] = useState("2025-07-20");
  const [to, setTo] = useState(getTodayDate());
  const [currency, setCurrency] = useState('USD');
  const [density, setDensity] = useState('comfortable');
  const [lastSync, setLastSync] = useState(new Date());
  const [user, setUser] = useState(null);
  const [fx, setFx] = useState(DEFAULT_RATES);
  
  useEffect(() => { const id = setInterval(() => { const now = new Date(); setLastSync(now); setFx(prev => ({ ...prev, updatedAt: now })); }, 30000); return () => clearInterval(id); }, []);
  
  useEffect(() => {
    const p = pageForPath(location.pathname);
    if (p !== page) setPage(p);
  }, [location.pathname, page]);
  
  useEffect(() => {
    const target = pathFor[page] || '/';
    if (location.pathname !== target) {
      navigate(target, { replace: false });
    }
  }, [page, location.pathname, navigate, pathFor]);

  const fetchData = useCallback(async () => {
    if (!user) return;
      
    setLoading(true);
    setError(null);
    try {
      const [adsRes, metricsRes] = await Promise.all([
        getAds(),
        getMetrics(from, to)
      ]);
      
      setAds(adsRes.data?.items || []); 
      setMetrics(metricsRes.data || []);
      setLastSync(new Date());
  
    } catch (err) {
      console.error("Failed to fetch data", err);
      setError(err.response?.data?.error || "Ошибка загрузки данных");
    }
    setLoading(false);
  }, [user, from, to]);

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    const token = localStorage.getItem("authToken");
    if (token) {
      setAuthHeader(token);
      setUser({ email: localStorage.getItem("userEmail") || 'user' });
    }
  }, []);
  
  const protectedPages = ['Dashboard', 'Ads', 'Integrations', 'Docs'];
  useEffect(() => { 
    if (!user && protectedPages.includes(page)) {
      setPage('Login');
    }
  }, [user, page]);  

  const handleLogin = async (email, password) => {
    const res = await login(email, password);
    const { access_token, refresh_token } = res.data;
    
    localStorage.setItem("authToken", access_token);
    localStorage.setItem("refreshToken", refresh_token);
    localStorage.setItem("userEmail", email);
    
    setAuthHeader(access_token);
    setUser({ email });
    setPage('Dashboard');
  };
  const handleSignUp = ({ email, name, access_token, refresh_token }) => {
    localStorage.setItem("authToken", access_token);
    localStorage.setItem("refreshToken", refresh_token);
    localStorage.setItem("userEmail", email);
    
    setAuthHeader(access_token);
    setUser({ email, name });
    setPage('Dashboard');
  };
  const handleLogout = () => {
    clearAuthHeader();
    localStorage.removeItem("authToken");
    localStorage.removeItem("refreshToken");
    localStorage.removeItem("userEmail");
    setUser(null);
    setPage('Landing');
  };
  
  const rows = useMemo(() => {
    const fromD = toDate(from); const toD = toDate(to);
    
    return metrics.filter(m => { 
        const matchesAd = adFilter === 'all' || m.ad_id === adFilter; 
        return matchesAd; 
      })
      .map(m => ({ 
        ...m, 
      }))
      .sort((a, b) => a.date.localeCompare(b.date) || a.ad_id.localeCompare(b.ad_id));
  }, [metrics, from, to, adFilter]); 

  const convertedRows = useMemo(() => rows.map(r => ({ 
      ...r, 
      revenue: convertAmount(parseFloat(r.revenue) || 0, currency, fx), 
      spend: convertAmount(parseFloat(r.spend) || 0, currency, fx) 
  })), [rows, currency, fx]);

  const aggUSD = useMemo(() => rows.reduce((acc, r) => ({ 
      clicks: acc.clicks + (parseInt(r.clicks, 10) || 0), 
      conversions: acc.conversions + (parseInt(r.conversions, 10) || 0), 
      revenue: acc.revenue + (parseFloat(r.revenue) || 0), 
      spend: acc.spend + (parseFloat(r.spend) || 0) 
  }), { clicks: 0, conversions: 0, revenue: 0, spend: 0 }), [rows]);
  
  const CPAusd = aggUSD.conversions > 0 && aggUSD.spend > 0 ? aggUSD.spend / aggUSD.conversions : 0;
  const ROAS = aggUSD.spend > 0 ? aggUSD.revenue / aggUSD.spend : 0;
  const aggRevenue = convertAmount(aggUSD.revenue, currency, fx);
  const aggSpend = convertAmount(aggUSD.spend, currency, fx);
  const CPA = CPAusd > 0 ? convertAmount(CPAusd, currency, fx) : 0;

  // --- ОБРАБОТЧИКИ ДЕЙСТВИЙ ---
  const onToggle = async (id, currentStatus) => {
    const action = currentStatus === 'active' ? pauseAd : resumeAd;
    try {
      await action(id);
      const res = await getAds();
      setAds(res.data?.items || []);
    } catch (err) {
      console.error("Failed to toggle ad status", err);
      alert("Ошибка при смене статуса");
    }
  };

  const handleSyncToday = async () => {
    const firstAd = ads[0];
    if (!firstAd) {
      alert("Сначала привяжите аккаунты Google Ads");
      return;
    }
    const customer_id_from_example = "123-456-7890";
    const today = getTodayDate();
    
    alert(`Отправка запроса на синк для ${customer_id_from_example} за ${today}...`);
    try {
      await syncGoogle(customer_id_from_example, today);
      alert("Синхронизация завершена! Обновляю метрики...");
      await fetchData();
    } catch (err) {
      console.error("Sync failed", err);
      alert(`Ошибка синхронизации: ${err.response?.data?.error}`);
    }
  };
  
  const goHome = () => {
    if (!user) {
      setPage('Landing');
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } else {
      setPage('Dashboard');
    }
  };
  
  const [fadeKey, setFadeKey] = useState(0);
  useEffect(() => { setFadeKey(k => k + 1); }, [page]);

  const isAuth = page === 'Login' || page === 'SignUp' || page === 'Forgot';
  
  // --- РЕНДЕРИНГ СТРАНИЦЫ ---
  
  const current = isAuth ? (
    page === 'Login' ? <AuthLoginPage t={t} onSubmit={handleLogin} goReset={() => setPage('Forgot')} goSignUp={() => setPage('SignUp')} />
        : page === 'SignUp' ? <AuthSignUpPage t={t} onSuccess={handleSignUp} goLogin={() => setPage('Login')} />
        : <AuthForgotPage t={t} goLogin={() => setPage('Login')} />
  ) : page === 'Landing' ? (
    <LandingPage onSignUp={() => setPage('SignUp')} />
  ) : page === 'Dashboard' ? (
    <DashboardPage
      t={t}
      error={error}
      aggUSD={aggUSD}
      aggRevenue={aggRevenue}
      aggSpend={aggSpend}
      CPA={CPA}
      ROAS={ROAS}
      currency={currency}
      convertedRows={convertedRows}
      handleSyncToday={handleSyncToday}
      from={from} to={to} setFrom={setFrom} setTo={setTo}
      adFilter={adFilter} setAdFilter={setAdFilter}
      ads={ads}
      exportCSV={exportCSV}
      loading={loading}
      density={density}
    />
  ) : page === 'Ads' ? (
    <AdsPage ads={ads} onToggle={onToggle} onView={(id) => { setAdFilter(id); setPage('Dashboard'); }} t={t} />
  ) : page === 'Integrations' ? (
    <IntegrationsPage t={t} />
  ) : (
    <DocsPage t={t} />
  );

  // --- ОСНОВНАЯ JSX-ВЕРСТКА ---
  return (
    <div className={classNames(dark ? 'dark' : undefined)}>
      <div className="min-h-screen text-zinc-900 dark:text-zinc-50">
        <header className="sticky top-0 z-10 ...">
          <div className="flex items-center justify-between h-14 max-w-7xl mx-auto px-6">
            <button onClick={goHome} /* ... */ >
              <img src="/logo.png" alt="AdSieve Logo" /* ... */ />
              <div className="text-lg font-semibold ..." style={{ color: BRAND.navy }}>
                AdSieve
              </div>
              <Badge tone="red">MVP</Badge>
            </button>
            {user && page !== 'Landing' && (
              <nav className="hidden md:flex items-center gap-1">
                {(['Dashboard', 'Ads', 'Integrations', 'Docs']).map(p => (
                  <button key={p}
                    onClick={() => user ? setPage(p) : setPage('Login')}
                    className={classNames("px-3 py-1.5 rounded-xl text-sm border transition", page === p ? "text-white" : "hover:bg-zinc-100 dark:hover:bg-zinc-800")}
                    style={{ backgroundColor: page === p ? BRAND.navy : undefined, borderColor: BRAND.mist }}>
                    {t[p]}
                  </button>
                ))}
                <button onClick={() => user ? setSettingsOpen(true) : setPage('Login')}
                  className="px-3 py-1.5 rounded-xl text-sm border hover:bg-zinc-100 dark:hover:bg-zinc-800 transition"
                  style={{ borderColor: BRAND.mist }}>{t.Settings}</button>
              </nav>
            )}
            <div className="flex items-center gap-3">
              {!user ? (
                <>
                  <button className="rounded-xl px-3 py-1.5 text-sm border transition" style={{ borderColor: BRAND.steel }}  onClick={() => { setPage('SignUp'); }}>{t.SignUp}</button>
                  <button className="rounded-xl px-3 py-1.5 text-sm border transition" style={{ borderColor: BRAND.steel }}  onClick={() => { setPage('Login'); }}>{t.Login}</button>
                </>
              ) : (
                <div className="flex items-center gap-2 text-xs">
                  <Badge tone="muted">{user.email}</Badge>
                  <button className="rounded-xl px-2.5 py-1.5 text-sm border transition" onClick={handleLogout}>{t.Logout}</button>
                </div>
              )}
              <button className="rounded-xl px-2.5 py-1.5 text-sm border transition" onClick={() => setDark(!dark)}>{dark ? t.Light : t.Dark}</button>
            </div>
          </div>
        </header>

        {/* Content wrapper */}
        <main key={fadeKey} className={classNames(/* ... */)}>
          {page === 'Landing' ? (
            <>{current}</>
          ) : (
            <div className="mx-auto max-w-7xl px-6 pb-16">
              {!isAuth && (
                <div className="text-xs text-zinc-500 mt-4">
                  <span className="mr-2">{t.LastSync}:</span>
                  <Badge tone="muted">{new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' }).format(lastSync)}</Badge>
                  <span className="mx-2">• FX:</span>
                  <Badge tone="muted">{new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' }).format(fx.updatedAt)} ({currency})</Badge>
                </div>
              )}
              <div className="mt-6 grid gap-4">
                {current}
              </div>
            </div>
          )}
        </main>

        {page !== 'Landing' && (
          <footer className="mt-12 text-xs ...">
            <div>AdSieve MVP UI prototype · Connected to backend API.</div>
          </footer>
        )}
      </div>

      <SettingsDrawer open={settingsOpen} onClose={() => setSettingsOpen(false)}>
        <SettingsPanel 
            lang={lang} setLang={setLang} 
            currency={currency} setCurrency={setCurrency}
            density={density} setDensity={setDensity} 
            dark={dark} setDark={setDark} t={t} />
      </SettingsDrawer>
    </div>
  );
}