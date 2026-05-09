import { create } from "zustand";
import { login as loginApi, register as registerApi, fetchCurrentUser, ApiError } from "@/lib/api";
import type { LoginRequest, RegisterRequest } from "@/lib/api";

interface User {
  id: number;
  username: string;
  nickname: string;
  avatar?: string;
}

interface AuthState {
  token: string | null;
  user: User | null;
  isLoading: boolean;
  error: string | null;
  /** 是否已完成客户端 hydration（从 localStorage 恢复 token） */
  _hydrated: boolean;
  /** user 信息是否已加载完成（含 fetchUser 成功/失败/无需加载） */
  _userFetched: boolean;
  
  init: () => void;
  login: (data: LoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => void;
  fetchUser: () => Promise<void>;
  clearError: () => void;
  setUser: (user: User, token: string) => void;
  updateAvatar: (avatarUrl: string) => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  isLoading: false,
  error: null,
  _hydrated: false,
  _userFetched: false,
  
  init: () => {
    // 防止重复初始化
    if (get()._hydrated) {
      return;
    }
    const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
    set({ token, _hydrated: true });

    // 无 token → 无需拉取 user
    if (!token) {
      set({ _userFetched: true });
      return;
    }

    // token 存在但 user 为空 → 异步获取用户信息
    if (!get().user) {
      set({ isLoading: true });
      fetchCurrentUser()
        .then((res) => {
          const { user } = res.data!;
          set({ user, isLoading: false, _userFetched: true });
        })
        .catch(() => {
          // token 过期或无效 → 清除
          localStorage.removeItem("token");
          set({ token: null, user: null, isLoading: false, _userFetched: true });
        });
    } else {
      set({ _userFetched: true });
    }
  },

  login: async (data: LoginRequest) => {
    set({ isLoading: true, error: null });
    try {
      const res = await loginApi(data);
      const { token, user } = res.data!;
      localStorage.setItem("token", token);
      set({ token, user, isLoading: false, _userFetched: true });
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "登录失败，请稍后重试";
      set({ error: message, isLoading: false });
      throw err;
    }
  },

  register: async (data: RegisterRequest) => {
    set({ isLoading: true, error: null });
    try {
      const res = await registerApi(data);
      const { token, user } = res.data!;
      localStorage.setItem("token", token);
      set({ token, user, isLoading: false, _userFetched: true });
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "注册失败，请稍后重试";
      set({ error: message, isLoading: false });
      throw err;
    }
  },

  logout: () => {
    localStorage.removeItem("token");
    set({ token: null, user: null, error: null, _userFetched: false });
  },

  fetchUser: async () => {
    set({ isLoading: true, error: null });
    try {
      const res = await fetchCurrentUser();
      const { user } = res.data!;
      set({ user, isLoading: false });
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "获取用户信息失败";
      set({ error: message, isLoading: false });
      throw err;
    }
  },

  clearError: () => set({ error: null }),

  setUser: (user: User, token: string) => {
    localStorage.setItem("token", token);
    set({ user, token });
  },

  updateAvatar: (avatarUrl: string) => {
    const user = get().user;
    if (user) {
      set({ user: { ...user, avatar: avatarUrl } });
    }
  },
}));
