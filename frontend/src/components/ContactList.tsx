"use client";

import { useState, useMemo } from "react";
import type { Contact } from "@/types/chat";
import { cn } from "@/lib/utils";
import { Avatar } from "@/components/Avatar";
import { UserPlus } from "lucide-react";
import { AddFriendDialog } from "@/components/AddFriendDialog";
import { ContactContextMenu } from "@/components/ContactContextMenu";

interface ContactListProps {
  contacts: Contact[];
  activeContactId: number | null;
  onSelectContact: (id: number) => void;
  className?: string;
  style?: React.CSSProperties;
  /** 显示模式：chat 含消息预览和未读，contacts 仅头像+用户名 */
  variant?: "chat" | "contacts";
  /** 是否显示添加好友按钮 */
  showAddFriend?: boolean;
  /** 搜索框下方的自定义内容 */
  headerContent?: React.ReactNode;
  /** 右键菜单删除聊天记录回调 */
  onDeleteChatHistory?: (contactId: number) => void;
  /** chatId → 未读消息数，作为未读数角标的唯一数据源 */
  unreadCounts?: Record<number, number>;
}

/** 格式化消息时间：今天→HH:mm，昨天→"昨天"，更早→MM-DD */
function formatTime(time?: string): string | null {
  if (!time) {return null;}
  const date = new Date(time);
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today.getTime() - 86400000);
  const msgDay = new Date(date.getFullYear(), date.getMonth(), date.getDate());
  if (msgDay.getTime() === today.getTime()) {
    return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
  }
  if (msgDay.getTime() === yesterday.getTime()) {
    return "昨天";
  }
  return `${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

export function ContactList({
  contacts,
  activeContactId,
  onSelectContact,
  className,
  style,
  variant = "chat",
  showAddFriend = false,
  headerContent,
  onDeleteChatHistory,
  unreadCounts = {},
}: ContactListProps) {
  const [search, setSearch] = useState("");
  const [addFriendOpen, setAddFriendOpen] = useState(false);

  // 右键菜单状态
  const [contextMenu, setContextMenu] = useState<{
    contactId: number;
    contactName: string;
    x: number;
    y: number;
  } | null>(null);

  const filteredContacts = useMemo(() => {
    if (!search.trim()) {return contacts;}
    const kw = search.toLowerCase();
    return contacts.filter(
      (c) =>
        c.nickname.toLowerCase().includes(kw) ||
        c.username.toLowerCase().includes(kw),
    );
  }, [contacts, search]);

  return (
    <aside className={cn("flex shrink-0 flex-col border-r bg-card", className)} style={style}>
      {/* 搜索框 + 添加好友 */}
      <div className="shrink-0 px-3 py-3">
        <div className="flex items-center gap-1.5">
          <div className="relative flex-1">
            <svg
              className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/40"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索联系人…"
              className="w-full rounded-lg border border-border bg-muted/50 py-2 pl-9 pr-8 text-sm text-foreground placeholder:text-muted-foreground/50 transition-colors focus:border-primary/30 focus:bg-background focus:outline-none focus:ring-2 focus:ring-primary/10"
            />
            {search && (
              <button
                onClick={() => setSearch("")}
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-muted-foreground/50 transition-colors hover:text-muted-foreground"
                aria-label="清除搜索"
              >
                <svg className="size-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            )}
          </div>

          {/* 添加好友按钮 — 搜索框右侧 */}
          {showAddFriend && (
            <button
              onClick={() => setAddFriendOpen(true)}
              className="flex size-9 shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
              title="添加好友"
            >
              <UserPlus className="size-5" />
            </button>
          )}
        </div>
      </div>

      {/* 自定义头部内容 */}
      {headerContent}

      {/* 联系人列表 */}
      <div className="flex-1 overflow-y-auto">
        {contacts.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <svg
              className="mb-3 size-10 text-muted-foreground/25"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
            <p className="text-sm text-muted-foreground">暂无联系人</p>
          </div>
        ) : filteredContacts.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <svg
              className="mb-3 size-10 text-muted-foreground/25"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <p className="text-sm text-muted-foreground">没有找到联系人</p>
            <p className="mt-1 text-xs text-muted-foreground/50">
              试试其他关键词
            </p>
          </div>
        ) : (
          <ul className="space-y-0.5 px-2 py-1">
            {filteredContacts.map((contact) => {
              const isActive = contact.id === activeContactId;
              const unread = contact.chat_id ? (unreadCounts[contact.chat_id] ?? 0) : 0;
              const hasUnread = unread > 0;
              const isChat = variant === "chat";
              const timeLabel = isChat ? formatTime(contact.lastMessageTime) : null;

              return (
                <li key={contact.id}>
                  <button
                    type="button"
                    onClick={() => onSelectContact(contact.id)}
                    onContextMenu={(e) => {
                      e.preventDefault();
                      setContextMenu({
                        contactId: contact.id,
                        contactName: contact.nickname || contact.username,
                        x: e.clientX,
                        y: e.clientY,
                      });
                    }}
                    className={cn(
                      "group flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-all duration-150",
                      isActive
                        ? "bg-[#3b82f6] hover:bg-[#3b82f6]/90 text-white"
                        : "hover:bg-muted/70",
                    )}
                  >
                    {/* 头像 */}
                    <div className="relative shrink-0">
                      <Avatar
                        src={contact.avatar}
                        alt={contact.nickname || contact.username}
                        size={36}
                        className={cn(
                          isActive ? "ring-2 ring-white/50" : "ring-2 ring-border",
                          isActive && "bg-white/30 text-white",
                        )}
                      />
                      {/* 未读红点 — 仅 chat 模式，显示在头像右上角 */}
                      {isChat && hasUnread && (
                        <span className="absolute -right-0.5 -top-0.5 flex min-w-[16px] items-center justify-center rounded-full bg-destructive px-1 py-0 text-[10px] font-bold leading-4 text-destructive-foreground shadow-sm">
                          {unread > 99 ? "99+" : unread}
                        </span>
                      )}
                    </div>

                    {/* 昵称 + 最后消息 */}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-baseline justify-between gap-2">
                        <p
                          className={cn(
                            "truncate text-sm",
                            isActive
                              ? "font-medium text-white"
                              : isChat && hasUnread && !isActive
                                ? "font-semibold text-foreground"
                                : "font-medium text-foreground",
                          )}
                        >
                          {contact.nickname || contact.username}
                        </p>
                        {/* 最后消息时间 — 仅 chat 模式 */}
                        {isChat && timeLabel && (
                          <span
                            className={cn(
                              "shrink-0 text-[11px]",
                              isActive ? "text-white/60" : "text-muted-foreground/60",
                            )}
                          >
                            {timeLabel}
                          </span>
                        )}
                      </div>
                      {/* 最后消息 — 仅 chat 模式 */}
                      {isChat && contact.lastMessage && (
                        <p
                          className={cn(
                            "mt-0.5 truncate text-xs",
                            isActive
                              ? "text-white/70"
                              : hasUnread && !isActive
                                ? "font-medium text-foreground/75"
                                : "text-muted-foreground",
                          )}
                        >
                          {contact.lastMessage}
                        </p>
                      )}
                    </div>
                  </button>
                </li>
              );
            })}
          </ul>
        )}
      </div>

      {/* 右键菜单 */}
      {contextMenu && (
        <ContactContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          contactId={contextMenu.contactId}
          contactName={contextMenu.contactName}
          onClose={() => setContextMenu(null)}
          onDeleteChatHistory={onDeleteChatHistory}
        />
      )}

      {/* 添加好友弹窗 */}
      <AddFriendDialog open={addFriendOpen} onOpenChange={setAddFriendOpen} />
    </aside>
  );
}
