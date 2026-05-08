"use client";

import { useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { MessageCircle, Users, LogOut } from "lucide-react";
import { getAvatarSrc } from "@/lib/avatar";

export function Sidebar() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const router = useRouter();
  const pathname = usePathname();
  const [avatarError, setAvatarError] = useState(false);

  const initial = (user?.nickname || user?.username || "?").charAt(0).toUpperCase();
  const username = user?.username ?? "me";
  const showAvatarImg = !avatarError;

  const isChatActive = pathname === "/chat";
  const isContactsActive = pathname === "/contacts";

  function handleLogout() {
    logout();
    router.replace("/login");
  }

  const navButton = (
    icon: React.ReactNode,
    label: string,
    active: boolean,
    href: string,
  ) => (
    <button
      onClick={() => router.push(href)}
      className={`flex size-10 items-center justify-center rounded-xl transition-colors ${
        active
          ? "bg-primary/10 text-primary"
          : "text-muted-foreground hover:bg-muted hover:text-foreground"
      }`}
      title={label}
    >
      {icon}
    </button>
  );

  return (
    <aside className="flex w-[68px] shrink-0 flex-col items-center justify-between bg-muted/50 py-4">
      {/* 顶部：用户头像 + 昵称 */}
      <div className="flex flex-col items-center gap-2">
        {user ? (
          <>
            {showAvatarImg ? (
              <img
                src={getAvatarSrc(user?.avatar, username)}
                alt={user.nickname || user.username}
                onError={() => setAvatarError(true)}
                className="size-10 shrink-0 rounded-full object-cover ring-2 ring-border"
              />
            ) : (
              <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary/15 text-sm font-semibold text-primary ring-2 ring-border">
                {initial}
              </div>
            )}
            <span className="max-w-[60px] truncate text-center text-[11px] font-medium leading-tight text-muted-foreground">
              {user.nickname || user.username}
            </span>
          </>
        ) : (
          <>
            <div className="size-10 shrink-0 animate-pulse rounded-full bg-muted" />
            <span className="h-3 w-12 animate-pulse rounded bg-muted" />
          </>
        )}
      </div>

      {/* 中间：导航按钮 */}
      <div className="flex flex-1 flex-col items-center gap-1 pt-6">
        {navButton(
          <MessageCircle className="size-5" />,
          "聊天",
          isChatActive,
          "/chat",
        )}
        {navButton(
          <Users className="size-5" />,
          "联系人",
          isContactsActive,
          "/contacts",
        )}
      </div>

      {/* 底部：退出按钮 */}
      <button
        onClick={handleLogout}
        className="flex size-10 items-center justify-center rounded-xl text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
        title="退出登录"
      >
        <LogOut className="size-5" />
      </button>
    </aside>
  );
}
