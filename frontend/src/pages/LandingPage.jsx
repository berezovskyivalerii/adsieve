import React, { useState, useEffect } from 'react';

function LandingPage({ onSignUp }) {
  const [slideIndex, setSlideIndex] = useState(0);
  const slides = [
    { image: "/slide1.png", title: "Добро пожаловать в AdSieve", subtitle: "Объедините данные вашей рекламы в одном месте." },
    { image: "/slide2.png", title: "Оптимизируйте рекламу", subtitle: "Клики, конверсии и ROI — под контролем." },
    { image: "/slide3.png", title: "Единая платформа", subtitle: "Подключения за минуты. Аналитика — сразу." },
    { image: "/slide4.png", title: "Гибкие интеграции", subtitle: "Facebook Ads и другие источники данных." },
    { image: "/slide5.png", title: "Готово к росту", subtitle: "От MVP к продакшену без боли." }
  ];
  useEffect(() => { const id = setInterval(() => setSlideIndex(i => (i + 1) % slides.length), 5000); return () => clearInterval(id); }, []);
  const prev = () => setSlideIndex(i => (i - 1 + slides.length) % slides.length);
  const next = () => setSlideIndex(i => (i + 1) % slides.length);

  return (
    <div className="landing-page">
      <section
        className="relative w-screen h-[60vh] md:h-[70vh] overflow-hidden text-white
                    left-1/2 right-1/2 -translate-x-1/2 mb-10">
        {slides.map((s, i) => (
          <div key={i}
            className={`absolute inset-0 transition-opacity duration-700 ${i === slideIndex ? 'opacity-100 visible' : 'opacity-0 invisible'}`}
            style={{ background: `url(${s.image}) center/cover no-repeat` }}>
            <div className="absolute inset-0 bg-black/40 flex flex-col items-center justify-center p-4">
              <h1 className="text-3xl md:text-5xl font-bold mb-2 text-center">{s.title}</h1>
              <p className="text-base md:text-xl mb-5 text-center opacity-90">{s.subtitle}</p>
              {i === slideIndex && (
                <button onClick={onSignUp} className="bg-white text-black font-semibold py-2.5 px-5 rounded-xl transition hover:translate-y-0.5">
                  Начать сейчас
                </button>
              )}
            </div>
          </div>
        ))}
        <button onClick={prev} aria-label="Prev"
          className="absolute left-4 top-1/2 -translate-y-1/2 bg-white/80 hover:bg-white text-black rounded-full w-10 h-10 flex items-center justify-center border transition">‹</button>
        <button onClick={next} aria-label="Next"
          className="absolute right-4 top-1/2 -translate-y-1/2 bg-white/80 hover:bg-white text-black rounded-full w-10 h-10 flex items-center justify-center border transition">›</button>
        <div className="absolute bottom-5 w-full flex justify-center gap-2">
          {slides.map((_, i) => (
            <button key={i} onClick={() => setSlideIndex(i)} className={`h-3 w-3 rounded-full border transition ${i === slideIndex ? 'bg-white border-white' : 'bg-white/50 border-white/80'}`} />
          ))}
        </div>
      </section>

      <div className="mx-auto max-w-7xl px-6">
        <section id="what-is" className="mb-10">
          <h2 className="text-3xl font-semibold mb-3">Что такое AdSieve?</h2>
          <p className="text-zinc-700 dark:text-zinc-200">
            AdSieve — платформа для объединения и анализа рекламных метрик из разных источников в едином дашборде.
            Видите картину целиком, оптимизируйте кампании и повышайте эффективность.
          </p>
        </section>

        <section id="features" className="mb-12">
          <h2 className="text-2xl font-semibold mb-4">Возможности</h2>
          <div className="grid md:grid-cols-3 gap-5">
            {[
              { img: '/feat1.png', title: 'Единый дашборд', text: 'Все кампании, объявления и метрики в одном месте.' },
              { img: '/feat2.png', title: 'Живые метрики', text: 'Клики, конверсии, выручка, расход — обновляются регулярно.' },
              { img: '/feat3.png', title: 'Интеграции', text: 'Быстрые подключения к популярным рекламным платформам.' },
            ].map((f, i) => (
              <div key={i} className="rounded-2xl border overflow-hidden bg-white/80 dark:bg-zinc-900/50">
                <div className="h-40 bg-zinc-200 dark:bg-zinc-800" style={{ background: `url(${f.img}) center/cover no-repeat` }} />
                <div className="p-5">
                  <div className="font-medium mb-1">{f.title}</div>
                  <div className="text-sm text-zinc-600 dark:text-zinc-300">{f.text}</div>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section id="how-to" className="mb-12">
          <h2 className="text-2xl font-semibold mb-3">Как использовать</h2>
          <ol className="list-decimal list-inside text-zinc-700 dark:text-zinc-200 space-y-1">
            <li>Создайте аккаунт и войдите.</li>
            <li>Подключите рекламные источники (например, Google Ads).</li>
            <li>Разместите пиксель/настройте S2S-конверсию на сайте.</li>
            <li>Анализируйте дашборд и оптимизируйте кампании.</li>
          </ol>
        </section>

        <section id="pricing" className="mb-12">
          <h2 className="text-2xl font-semibold mb-4">Цены</h2>
          <div className="grid md:grid-cols-3 gap-5">
            {[
              { name: 'Free (Beta)', price: '0$', features: ['Полный доступ в бетe', 'Без ограничений по проектам', 'Сообщество'] },
              { name: 'Pro', price: '—', features: ['Приоритетная поддержка', 'Доп. интеграции', 'Экспорт и API'] },
              { name: 'Enterprise', price: '—', features: ['SLA и SSO', 'Расширенные лимиты', 'White-label'] },
            ].map((p, idx) => (
              <div key={idx} className="rounded-2xl border p-6 bg-white/80 dark:bg-zinc-900/50">
                <div className="text-lg font-semibold mb-1">{p.name}</div>
                <div className="text-3xl font-bold mb-3">{p.price}</div>
                <ul className="text-sm text-zinc-600 dark:text-zinc-300 list-disc list-inside space-y-1">
                  {p.features.map((f, i) => <li key={i}>{f}</li>)}
                </ul>
              </div>
            ))}
          </div>
        </section>

        <section id="gallery" className="mb-12">
          <h2 className="text-2xl font-semibold mb-4">Как это выглядит</h2>
          <div className="grid md:grid-cols-3 gap-4">
            {['/gal1.png', '/gal2.png', '/gal3.png'].map((g, i) => (
              <div key={i} className="rounded-2xl overflow-hidden border h-48 bg-zinc-200 dark:bg-zinc-800" style={{ background: `url(${g}) center/cover no-repeat` }} />
            ))}
          </div>
        </section>

        <section id="faq" className="mb-16">
          <h2 className="text-2xl font-semibold mb-4">FAQ</h2>
          <div className="grid md:grid-cols-3 gap-4">
            <details className="rounded-xl border p-4"><summary className="font-medium">Можно ли добавить другие источники?</summary><div className="text-sm mt-2">Да, через интеграции можно расширять список источников.</div></details>
            <details className="rounded-xl border p-4"><summary className="font-medium">Нужен ли разработчик?</summary><div className="text-sm mt-2">Базовую интеграцию можно выполнить самостоятельно. Для S2S лучше привлечь разработчика.</div></details>
            <details className="rounded-xl border p-4"><summary className="font-medium">Будет ли мобильная версия?</summary><div className="text-sm mt-2">Да, позже. Сейчас приоритет — десктоп.</div></details>
          </div>
        </section>
      </div>

      <footer className="w-screen left-1/2 right-1/2 -translate-x-1/2 relative border-t mt-10">
        <div className="mx-auto max-w-7xl px-6 py-8 text-sm text-zinc-600 dark:text-zinc-400">
          <div className="grid md:grid-cols-4 gap-6">
            <div>
              <div className="text-lg font-semibold text-zinc-800 dark:text-zinc-100">AdSieve</div>
              <p className="mt-2">Единая аналитика рекламных метрик. Прозрачно. Быстро. Масштабируемо.</p>
              <p className="mt-2">© 2025 AdSieve</p>
            </div>
            <div>
              <div className="font-medium text-zinc-800 dark:text-zinc-100">Продукт</div>
              <ul className="mt-2 space-y-1">
                <li><a href="#features" className="hover:underline">Возможности</a></li>
                <li><a href="#pricing" className="hover:underline">Цены</a></li>
                <li><a href="#gallery" className="hover:underline">Скриншоты</a></li>
              </ul>
            </div>
            <div>
              <div className="font-medium text-zinc-800 dark:text-zinc-100">Ресурсы</div>
              <ul className="mt-2 space-y-1">
                <li><a href="#faq" className="hover:underline">FAQ</a></li>
                <li><a href="#docs" onClick={(e) => e.preventDefault()} className="hover:underline">Документация</a></li>
                <li><a href="#support" onClick={(e) => e.preventDefault()} className="hover:underline">Поддержка</a></li>
              </ul>
            </div>
            <div>
              <div className="font-medium text-zinc-800 dark:text-zinc-100">Компания</div>
              <ul className="mt-2 space-y-1">
                <li><a href="#about" onClick={(e) => e.preventDefault()} className="hover:underline">О нас</a></li>
                <li><a href="#contact" onClick={(e) => e.preventDefault()} className="hover:underline">Контакты</a></li>
                <li><a href="#legal" onClick={(e) => e.preventDefault()} className="hover:underline">Условия и конфиденциальность</a></li>
              </ul>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}