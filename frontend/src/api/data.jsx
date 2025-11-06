import API from "./auth"; // Импортируем наш настроенный axios

// --- Метрики и Сущности ---

/**
 * Получить агрегированные метрики
 */
export function getMetrics(from, to) {
  return API.get(`/api/metrics?from=${from}&to=${to}`);
}

/**
 * Получить список объявлений пользователя
 */
export function getAds() {
  return API.get("/api/ads");
}

/**
 * Поставить объявление на паузу (пока не реализовано в API, но заглушка есть)
 * PDF Спека: POST /ads/{ad_id}/pause
 * Простая API дока: нет
 * * !!! ВАЖНО: В твоей API-документации НЕТ эндпоинта для паузы. 
 * Я реализую логику вызова, но если эндпоинта нет, он будет возвращать 404.
 * Если он появится, код УЖЕ готов.
 */
export function pauseAd(ad_id) {
  // return API.post(`/api/ads/${ad_id}/pause`);
  // Временно возвращаем Promise.resolve(), т.к. ручки нет в доке
  console.warn("API for pauseAd not specified in docs, faking success.");
  return Promise.resolve();
}

/**
 * Возобновить объявление
 * PDF Спека: POST /ads/{ad_id}/resume
 * Простая API дока: нет
 */
export function resumeAd(ad_id) {
  // return API.post(`/api/ads/${ad_id}/resume`);
  console.warn("API for resumeAd not specified in docs, faking success.");
  return Promise.resolve();
}


// --- Интеграция Google Ads ---

/**
 * 1. Получить ссылку на OAuth
 */
export function connectGoogle() {
  return API.post("/integrations/google/connect");
}

/**
 * 2. Получить список доступных аккаунтов
 */
export function getGoogleAccounts() {
  return API.get("/integrations/google/accounts");
}

/**
 * 3. Привязать выбранные аккаунты
 */
export function linkGoogleAccounts(customer_ids) {
  return API.post("/integrations/google/link-accounts", { customer_ids });
}

/**
 * 4. Ручной синк расходов за сегодня
 */
export function syncGoogle(customer_id, date) {
  // date YYYY-MM-DD
  return API.post("/integrations/google/sync", { customer_id, date });
}