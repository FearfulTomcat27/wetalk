"use client";

import { useEffect, useState } from "react";
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
  const [mounted, setMounted] = useState(false);

  // 初始化：从 localStorage 恢复 token
  useEffect(() => {
    init();
  }, [init]);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (mounted && _hydrated && token) {
      router.replace("/");
    }
  }, [mounted, _hydrated, token, router]);

  // 等待 hydration 或未挂载 → 显示空白
  if (!_hydrated || !mounted) {
    return null;
  }

  // 已登录 → 跳转中
  if (token) {
    return null;
  }

  return <>{children}</>;
}
