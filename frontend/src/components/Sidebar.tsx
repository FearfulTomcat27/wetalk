"use client";

import { useRouter, usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/auth";
import { MessageCircle, Users, LogOut } from "lucide-react";
import { Avatar } from "@/components/Avatar";
import { ProfilePopover } from "@/components/ProfilePopover";

export function Sidebar() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const router = useRouter();
  const pathname = usePathname();

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
          ? "text-[#3b82f6]"
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