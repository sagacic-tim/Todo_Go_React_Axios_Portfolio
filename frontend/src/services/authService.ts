// src/services/authService.ts
import { apiClient } from "./apiClient";
import type { AuthUser } from "../contexts/AuthContext";

interface AuthResponse {
  user: AuthUser;
}

export async function signup(email: string, password: string): Promise<AuthUser> {
  const { data } = await apiClient.post<AuthResponse>("/auth/signup", {
    email,
    password,
  });
  return data.user;
}

export async function login(email: string, password: string): Promise<AuthUser> {
  const { data } = await apiClient.post<AuthResponse>("/auth/login", {
    email,
    password,
  });
  return data.user;
}

export async function logout(): Promise<void> {
  await apiClient.post("/auth/logout");
}

export async function getMe(): Promise<AuthUser> {
  const { data } = await apiClient.get<AuthResponse>("/auth/me");
  return data.user;
}
