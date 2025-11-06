import React, { useState } from 'react';

function SocialButtons({ t, variant = 'login' }) {
  const label = variant === 'login' ? t.OrWith : t.OrRegWith;
  const btn = (src, text) => (
    <button className="w-full flex items-center justify-center border rounded-xl py-2 hover:bg-zinc-50 transition">
      <img src={src} alt={text} className="h-5 w-5 mr-2" /> {text}
    </button>
  );
  return (
    <div className="mt-5">
      <div className="text-sm text-zinc-500 text-center mb-2 border-t pt-3">{label}</div>
      <div className="grid gap-2">
        {btn("/google-icon.png", "Google")}
        {btn("/apple-icon.png", "Apple")}
        {btn("/microsoft-icon.png", "Microsoft")}
        {btn("/facebook-icon.png", "Facebook")}
      </div>
    </div>
  );
}

function AuthLoginPage({ t, onSubmit, goReset, goSignUp }) {
  const [email, setEmail] = useState(''); 
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await onSubmit(email, password);
    } catch (err) {
      console.error(err);
      setError(err.response?.data?.error || "Неверный логин или пароль");
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-md w-full p-8 rounded-2xl border shadow bg-white/90 dark:bg-zinc-900/70 backdrop-blur transition">
      <div className="text-2xl font-semibold mb-1" style={{ color: BRAND.navy }}>{t.Login}</div>
      <div className="text-sm text-zinc-500 mb-4">Добро пожаловать в AdSieve</div>
      
      {error && <p className="text-red-600 bg-red-100 p-3 rounded-xl mb-3 text-sm">{error}</p>}

      <form onSubmit={handleSubmit} className="grid gap-3">
        <label className="text-sm">
          <div className="mb-1 text-zinc-500">{t.Email}</div>
          <input value={email} onChange={e => setEmail(e.target.value)} className="w-full rounded-xl border px-3 py-2" placeholder="you@example.com" required />
        </label>
        <label className="text-sm">
          <div className="mb-1 text-zinc-500">{t.Password}</div>
          <input type="password" value={password} onChange={e => setPassword(e.target.value)} className="w-full rounded-xl border px-3 py-2" placeholder="********" required />
        </label>
        <div className="flex items-center justify-between text-sm">
          <button type="button" onClick={goReset} className="underline text-zinc-500 hover:text-zinc-700 transition">{t.Forgot}</button>
          <button 
            type="submit"
            disabled={loading}
            className="rounded-xl px-4 py-2 text-white transition transform hover:translate-y-0.5 disabled:opacity-50" 
            style={{ backgroundColor: BRAND.navy }}
          >
            {loading ? "Вход..." : t.Continue}
          </button>
        </div>
        <div className="text-sm text-zinc-500">Нет аккаунта? <button type="button" onClick={goSignUp} className="underline">Создать</button></div>
      </form>
      <SocialButtons t={t} variant="login" />
    </div>
  );
}