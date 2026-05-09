import { request } from "./request";
import type { ApiResponse } from "./request";

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  confirmPassword: string;
  nickname?: string;
}

export interface AuthResponse {
  token: string;
  user: {
    id: number;
    username: string;
    nickname: string;
    avatar?: string;
  };
}

export function login(data: LoginRequest) {
  return request<AuthResponse>({ method: "POST", url: "/api/auth/login", data });
}

export function register(data: RegisterRequest) {
  const { confirmPassword: _, ...payload } = data;
  return request<AuthResponse>({ method: "POST", url: "/api/auth/register", data: payload });
}

export function fetchCurrentUser() {
  return request<Omit<AuthResponse, "token">>({ method: "GET", url: "/api/me" });
}
