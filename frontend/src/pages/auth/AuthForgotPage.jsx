import React, { useState } from 'react';

function AuthForgotPage({ t, goLogin }) {
  const [email, setEmail] = useState('');
  const onSend = () => alert('Reset link sent to ' + email);
  return (
    <div className="mx-auto max-w-md w-full p-8 rounded-2xl border shadow bg-white/90 dark:bg-zinc-900/70 backdrop-blur transition">
      <div className="text-2xl font-semibold mb-1" style={{ color: BRAND.navy }}>{t.ResetPassword}</div>
      <div className="text-sm text-zinc-500 mb-4">Мы отправим письмо со ссылкой для восстановления</div>
      <div className="grid gap-3">
        <label className="text-sm"><div className="mb-1 text-zinc-500">{t.Email}</div><input value={email} onChange={e => setEmail(e.target.value)} className="w-full rounded-xl border px-3 py-2" placeholder="you@example.com" /></label>
        <div className="flex items-center justify-between text-sm">
          <button onClick={goLogin} className="underline text-zinc-500 hover:text-zinc-700 transition">{t.Login}</button>
          <button onClick={onSend} className="rounded-xl px-4 py-2 text-white transition transform hover:translate-y-0.5" style={{ backgroundColor: BRAND.navy }}>{t.SendLink}</button>
        </div>
      </div>
    </div>
  );
}