"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { isPublicRoute, isProtectedRoute, isRedirectRoute } from "@/config/routes";

interface RouteGuardProps {
  children: React.ReactNode;
}

export default function RouteGuard({ children }: RouteGuardProps) {
  const pathname = usePathname();
  const router = useRouter();
  const token = useAuthStore((s) => s.token);
  const _hydrated = useAuthStore((s) => s._hydrated);
  const _userFetched = useAuthStore((s) => s._userFetched);
  const init = useAuthStore((s) => s.init);

  // 客户端 hydration：从 localStorage 恢复 auth 状态
  useEffect(() => {
    init();
  }, [init]);

  // hydration 完成后根据路径 + 登录状态执行重定向
  useEffect(() => {
    if (!_hydrated) {
      return;
    }

    // 公开路由（登录/注册）→ 已登录用户跳转聊天页
    if (isPublicRoute(pathname) && token) {
      router.replace("/chat");
      return;
    }

    // 重定向路由（如 /）→ 已登录用户跳转聊天页
    if (isRedirectRoute(pathname) && token) {
      router.replace("/chat");
      return;
    }

    // 受保护路由 → 未登录用户跳转登录页
    if (isProtectedRoute(pathname) && !token) {
      router.replace("/login");
      return;
    }
  }, [_hydrated, pathname, token, router]);

  // 等待 hydration 完成
  if (!_hydrated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <div className="flex flex-col items-center gap-4">
          <div className="size-10 animate-spin rounded-full border-4 border-muted border-t-primary" />
          <p className="text-sm text-muted-foreground">加载中…</p>
        </div>
      </div>
    );
  }

  // 受保护路由 + 已登录但 user 尚未加载完成 → 等待
  if (isProtectedRoute(pathname) && token && !_userFetched) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <div className="flex flex-col items-center gap-4">
          <div className="size-10 animate-spin rounded-full border-4 border-muted border-t-primary" />
          <p className="text-sm text-muted-foreground">正在恢复登录…</p>
        </div>
      </div>
    );
  }

  // 公开路由 + 已登录 → 等待跳转
  if (isPublicRoute(pathname) && token) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-sm text-muted-foreground">正在跳转...</p>
      </div>
    );
  }

  // 重定向路由（/）+ 已登录 → 等待跳转
  if (isRedirectRoute(pathname) && token) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-sm text-muted-foreground">正在跳转...</p>
      </div>
    );
  }

  // 受保护路由 + 未登录 → 等待跳转
  if (isProtectedRoute(pathname) && !token) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black">
        <p className="text-sm text-muted-foreground">正在跳转...</p>
      </div>
    );
  }

  // 直接渲染
  return <>{children}</>;
}
