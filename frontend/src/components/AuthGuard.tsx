"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth";

interface AuthGuardProps {
  children: React.ReactNode;
}

export default function AuthGuard({ children }: AuthGuardProps) {
  const router = useRouter();
  const token = useAuthStore((s) => s.token);
  const _hydrated = useAuthStore((s) => s._hydrated);
  const init = useAuthStore((s) => s.init);

  // 初始化：从 localStorage 恢复 token
  useEffect(() => {
    init();
  }, [init]);

  useEffect(() => {
    if (_hydrated && token) {
      router.replace("/");
    }
  }, [_hydrated, token, router]);

  // 等待 hydration → 显示空白
  if (!_hydrated) {
    return null;
  }

  // 已登录 → 跳转中
  if (token) {
    return null;
  }

  return <>{children}</>;
}