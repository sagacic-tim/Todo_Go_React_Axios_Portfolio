// src/services/apiClient.ts

import axios, { AxiosError } from "axios";

const baseURL = import.meta.env.VITE_API_BASE_URL ?? "/api";

export const apiClient = axios.create({
  baseURL,
  timeout: 15_000,
  headers: {
    "Content-Type": "application/json",
  },
  // HttpOnly cookie is sent automatically for same-domain requests.
  withCredentials: true,
});

// ─── Response interceptor ─────────────────────────────────────────────────────
// When the backend returns 401 (session expired / not logged in), dispatch
// a custom event so the App shell can open the auth modal.
// We skip the interception for auth routes themselves to avoid redirect loops.

apiClient.interceptors.response.use(
  (res) => res,
  (err: AxiosError) => {
    const url = err.config?.url ?? "";
    const isAuthRoute = url.startsWith("/auth/");

    if (err.response?.status === 401 && !isAuthRoute) {
      window.dispatchEvent(new CustomEvent("auth:required"));
    }

    return Promise.reject(err);
  }
);

// ─── Error helper ─────────────────────────────────────────────────────────────

export function toMessage(err: unknown): string {
  if (!err) return "Unknown error";
  if (axios.isAxiosError(err)) {
    const ax = err as AxiosError<any>;
    return (
      ax.response?.data?.error ||
      ax.response?.data?.message ||
      ax.message ||
      "Request failed"
    );
  }
  return err instanceof Error ? err.message : String(err);
}
