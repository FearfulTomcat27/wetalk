"use client";

interface MainLayoutProps {
  children: React.ReactNode;
}

/**
 * /chat 页面的主布局 — 全屏三列布局，路由守卫由 RouteGuard 统一处理。
 * 此 layout 仅包裹 children，不做额外 UI，具体布局由 chat/page.tsx 负责。
 */
export default function MainLayout({ children }: MainLayoutProps) {
  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {children}
    </div>
  );
}
