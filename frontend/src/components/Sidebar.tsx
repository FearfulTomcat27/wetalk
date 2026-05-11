"use client";

import { useRouter, usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { useChatStore } from "@/stores/chat";
import { MessageCircle, Users, LogOut, Sun, Moon } from "lucide-react";
import { useTheme } from "next-themes";
import { Avatar } from "@/components/Avatar";
import { ProfilePopover } from "@/components/ProfilePopover";

export function Sidebar() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const { theme, setTheme } = useTheme();
  const router = useRouter();
  const pathname = usePathname();

  const isChatActive = pathname === "/chat";
  const isContactsActive = pathname === "/contacts";
  const pendingRequestsCount = useChatStore((s) => s.pendingRequestsCount);
  const unreadCounts = useChatStore((s) => s.unreadCounts);
  const totalUnread = Object.values(unreadCounts).reduce((sum, c) => sum + c, 0);

  function handleLogout() {
    logout();
    router.replace("/login");
  }

  return (
    <aside className="flex w-[68px] shrink-0 flex-col items-center justify-between bg-muted/50 py-4">
      {/* 顶部：用户头像 + 昵称 */}
      <div className="flex flex-col items-center gap-2">
        {user ? (
          <>
            <ProfilePopover>
              <Avatar
                src={user.avatar}
                alt={user?.nickname || user?.username || "?"}
                size={36}
              />
            </ProfilePopover>
          </>
        ) : (
          <>
            <div className="size-10 shrink-0 animate-pulse rounded-lg bg-muted" />
            <span className="h-3 w-12 animate-pulse rounded bg-muted" />
          </>
        )}
      </div>

      {/* 中间：导航按钮 */}
      <div className="flex flex-1 flex-col items-center gap-1 pt-6">
        <button
          onClick={() => router.push("/chat")}
          className={`relative flex size-10 items-center justify-center rounded-xl transition-colors ${
            isChatActive
              ? "text-[#3b82f6]"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
          title="聊天"
        >
          <MessageCircle className="size-5" />
          {totalUnread > 0 && (
            <span className="absolute -right-0.5 -top-0.5 flex min-w-[16px] items-center justify-center rounded-full bg-destructive px-1 py-0 text-[10px] font-bold leading-4 text-destructive-foreground shadow-sm">
              {totalUnread > 99 ? "99+" : totalUnread}
            </span>
          )}
        </button>
        <button
          onClick={() => router.push("/contacts")}
          className={`relative flex size-10 items-center justify-center rounded-xl transition-colors ${
            isContactsActive
              ? "text-[#3b82f6]"
              : "text-muted-foreground hover:bg-muted hover:text-foreground"
          }`}
          title="联系人"
        >
          <Users className="size-5" />
          {pendingRequestsCount > 0 && (
            <span className="absolute -right-0.5 -top-0.5 flex min-w-[16px] items-center justify-center rounded-full bg-destructive px-1 py-0 text-[10px] font-bold leading-4 text-destructive-foreground shadow-sm">
              {pendingRequestsCount > 99 ? "99+" : pendingRequestsCount}
            </span>
          )}
        </button>
      </div>

      {/* 底部：主题切换 + 退出按钮 */}
      <div className="flex flex-col items-center gap-1">
        <button
          onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
          className="flex size-10 items-center justify-center rounded-xl text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          title={theme === "dark" ? "切换亮色" : "切换暗色"}
        >
          {theme === "dark" ? <Sun className="size-5" /> : <Moon className="size-5" />}
        </button>
        <button
        onClick={handleLogout}
        className="flex size-10 items-center justify-center rounded-xl text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
        title="退出登录"
      >
        <LogOut className="size-5" />
      </button>
      </div>
    </aside>
  );
}