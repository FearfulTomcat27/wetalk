"use client";

import { useState, useMemo } from "react";
import type { Contact } from "@/types/chat";
import { cn } from "@/lib/utils";

interface ContactListProps {
  contacts: Contact[];
  activeContactId: number | null;
  onSelectContact: (id: number) => void;
}

export function ContactList({ contacts, activeContactId, onSelectContact }: ContactListProps) {
  const [search, setSearch] = useState("");

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
    <aside className="flex w-80 shrink-0 flex-col border-r bg-card">
      {/* 搜索框 */}
      <div className="shrink-0 px-3 py-3">
        <div className="relative">
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
      </div>

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
              const hasUnread = contact.unread > 0;

              return (
                <li key={contact.id}>
                  <button
                    type="button"
                    onClick={() => onSelectContact(contact.id)}
                    className={cn(
                      "group flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-all duration-150",
                      isActive
                        ? "bg-primary/10 hover:bg-primary/15"
                        : "hover:bg-muted/70",
                    )}
                  >
                    {/* 头像 */}
                    <div className="relative shrink-0">
                      <div
                        className={cn(
                          "flex size-11 items-center justify-center rounded-full text-sm font-semibold transition-colors",
                          isActive
                            ? "bg-primary text-primary-foreground"
                            : "bg-muted text-muted-foreground group-hover:bg-muted-foreground/15",
                        )}
                      >
                        {(contact.nickname || contact.username).charAt(0).toUpperCase()}
                      </div>
                      {/* 未读红点 - 微信风格 */}
                      {hasUnread && !isActive && (
                        <span className="absolute -right-0.5 -top-0.5 flex min-w-[16px] items-center justify-center rounded-full bg-destructive px-1 py-0 text-[10px] font-bold leading-4 text-destructive-foreground shadow-sm">
                          {contact.unread > 99 ? "99+" : contact.unread}
                        </span>
                      )}
                    </div>

                    {/* 昵称 + 最后消息 */}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-baseline justify-between gap-2">
                        <p
                          className={cn(
                            "truncate text-sm",
                            hasUnread && !isActive
                              ? "font-semibold text-foreground"
                              : "font-medium text-foreground",
                          )}
                        >
                          {contact.nickname || contact.username}
                        </p>
                        {/* 活跃状态下的未读标记 */}
                        {hasUnread && isActive && (
                          <span className="shrink-0 rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-bold leading-none text-primary-foreground">
                            {contact.unread > 99 ? "99+" : contact.unread}
                          </span>
                        )}
                      </div>
                      {contact.lastMessage && (
                        <p
                          className={cn(
                            "mt-0.5 truncate text-xs",
                            hasUnread && !isActive
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
    </aside>
  );
}
