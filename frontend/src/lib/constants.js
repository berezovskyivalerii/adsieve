export const BRAND = { navy: "#02285c", red: "#b7291e", mist: "#ddd7da", steel: "#5b6b89" };
export const CURRENCIES = ["USD", "EUR", "UAH", "CHF", "GBP", "CNY"];
export const DEFAULT_RATES = { base: 'USD', updatedAt: new Date(), rates: { USD: 1, EUR: 0.92, UAH: 41.334, CHF: 0.86, GBP: 0.77, CNY: 7.10 } };
export const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";