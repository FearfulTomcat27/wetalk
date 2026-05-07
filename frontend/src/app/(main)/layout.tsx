"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth";

interface MainLayoutProps {
  children: React.ReactNode;
}

export default function MainLayout({ children }: MainLayoutProps) {
  const router = useRouter();
  const { token, user, _hydrated } = useAuthStore();
  const init = useAuthStore((s) => s.init);
  // token 即代表已登录，user 由 init() 异步加载无需阻塞路由
  const isLoggedIn = Boolean(token);

  // 初始化：从 localStorage 恢复 token
  useEffect(() => {
    init();
  }, [init]);

  // hydration 完成后 → 未登录则跳转
  useEffect(() => {
    if (_hydrated && !isLoggedIn) {
      router.replace("/login");
    }
  }, [_hydrated, isLoggedIn, router]);

  // 等待 hydration
  if (!_hydrated) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-4">
          <div className="size-10 animate-spin rounded-full border-4 border-muted border-t-primary" />
          <p className="text-sm text-muted-foreground">加载中…</p>
        </div>
      </div>
    );
  }

  // 未登录 → 显示跳转 loading（useEffect 中执行 replace）
  if (!isLoggedIn) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">正在跳转...</p>
      </div>
    );
  }

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-background">
      {/* 顶部导航栏 */}
      <header className="flex h-14 shrink-0 items-center justify-between border-b bg-card px-5">
        <div className="flex items-center gap-2.5">
          <div className="flex size-8 items-center justify-center rounded-lg bg-primary">
            <svg
              className="size-4 text-primary-foreground"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
          </div>
          <h1 className="text-base font-semibold tracking-tight">WeTalk</h1>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">
            {user?.nickname || user?.username}
          </span>
          <button
            onClick={() => useAuthStore.getState().logout()}
            className="rounded-md px-2 py-1 text-sm text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
          >
            退出
          </button>
        </div>
      </header>
      {children}
    </div>
  );
}
