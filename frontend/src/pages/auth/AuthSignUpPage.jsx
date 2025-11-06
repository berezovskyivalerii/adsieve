import React, { useState } from 'react';

function AuthSignUpPage({ t, onSuccess, goLogin }) {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    setSuccess(false);

    try {
      const res = await signup(email, password);
      // Сразу логиним пользователя
      const { access_token, refresh_token } = res.data;
      
      setSuccess(true);
      setLoading(false);
      
      setTimeout(() => onSuccess({ email, name, access_token, refresh_token }), 1000);

    } catch (err) {
      console.error("Произошла ошибка при регистрации:", err);
      setLoading(false);
      if (err.code === "ERR_NETWORK") {
        setError("Ошибка сети. Сервер недоступен.");
      } else if (err.response) {
        setError(`Ошибка: ${err.response?.data?.error || 'Не удалось зарегистрироваться.'}`);
      } else {
        setError("Произошла неизвестная ошибка.");
      }
    }
  };

  return (
    <div className="mx-auto max-w-md w-full p-8 rounded-2xl border shadow bg-white/90 dark:bg-zinc-900/70 backdrop-blur transition">
      <div className="text-2xl font-semibold mb-1" style={{ color: BRAND.navy }}>{t.SignUp}</div>
      <div className="text-sm text-zinc-500 mb-4">Бесплатно в бете</div>
      
      {error && <p className="text-red-600 bg-red-100 p-3 rounded-xl mb-3 text-sm">{error}</p>}
      {success && <p className="text-emerald-600 bg-emerald-100 p-3 rounded-xl mb-3 text-sm">Успешно! Входим в аккаунт...</p>}
      
      <form onSubmit={handleSubmit} className="grid gap-3">
        <label className="text-sm"><div className="mb-1 text-zinc-500">{t.Name}</div><input value={name} onChange={e => setName(e.target.value)} className="w-full rounded-xl border px-3 py-2" placeholder="John Smith" /></label>
        <label className="text-sm"><div className="mb-1 text-zinc-500">{t.Email}</div><input type="email" value={email} onChange={e => setEmail(e.target.value)} className="w-full rounded-xl border px-3 py-2" placeholder="you@example.com" required /></label>
        <label className="text-sm"><div className="mb-1 text-zinc-500">{t.Password}</div><input type="password" value={password} onChange={e => setPassword(e.target.value)} className="w-full rounded-xl border px-3 py-2" placeholder="********" required /></label>
        <div className="flex items-center justify-between text-sm mt-2">
          <button type="button" onClick={goLogin} className="underline text-zinc-500 hover:text-zinc-700 transition">{t.Login}</button>
          <button 
            type="submit" 
            disabled={loading}
            className="rounded-xl px-4 py-2 text-white transition transform hover:translate-y-0.5 disabled:opacity-50" 
            style={{ backgroundColor: BRAND.navy }}
          >
            {loading ? "Создание..." : t.CreateAccount}
          </button>
        </div>
      </form>
      <SocialButtons t={t} variant="signup" />
    </div>
  );
}