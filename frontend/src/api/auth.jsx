import axios from "axios";

const API = axios.create({
  baseURL: import.meta.env.VITE_API_URL || "http://localhost:8080",
  withCredentials: true,
});

export function login(email, password) {
  return API.post("/api/auth/sign-in", { email, password });
}

export function signup(email, password) {
  return API.post("/api/auth/sign-up", { email, password });
}

export function refreshToken(refresh_token) {
  return API.post("/api/auth/refresh", { refresh_token });
}

export function setAuthHeader(token) {
  API.defaults.headers.common["Authorization"] = `Bearer ${token}`;
}

export function clearAuthHeader() {
  delete API.defaults.headers.common["Authorization"];
}

// === ЛОГИКА REFRESH TOKEN ===
API.interceptors.response.use(
  (response) => response, // Все '2xx' ответы просто пропускаем
  async (error) => {
    const originalRequest = error.config;

    // Если ошибка 401 И это не был запрос на refresh (чтобы не зациклиться)
    if (error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true; // Помечаем, что мы уже пытались повторить
      
      const currentRefreshToken = localStorage.getItem("refreshToken");
      if (!currentRefreshToken) {
        // Если refresh токена нет, то просто выходим
        clearAuthHeader();
        localStorage.removeItem("authToken");
        localStorage.removeItem("refreshToken");
        window.location.href = '/login'; // Перезагружаем на логин
        return Promise.reject(error);
      }

      try {
        // 1. Пытаемся обновить токен
        const res = await refreshToken(currentRefreshToken);
        const { access_token, refresh_token } = res.data;

        // 2. Сохраняем новые токены
        localStorage.setItem("authToken", access_token);
        localStorage.setItem("refreshToken", refresh_token);
        
        // 3. Обновляем заголовок в axios
        setAuthHeader(access_token);
        
        // 4. Обновляем заголовок в оригинальном запросе
        originalRequest.headers["Authorization"] = `Bearer ${access_token}`;

        // 5. Повторяем оригинальный запрос с новым токеном
        return API(originalRequest);

      } catch (refreshError) {
        // Если refresh не удался, то все, разлогиниваем
        console.error("Refresh token failed", refreshError);
        clearAuthHeader();
        localStorage.removeItem("authToken");
        localStorage.removeItem("refreshToken");
        window.location.href = '/login'; // Перезагружаем на логин
        return Promise.reject(error);
      }
    }

    // Для всех других ошибок просто пробрасываем их
    return Promise.reject(error);
  }
);

export default API;