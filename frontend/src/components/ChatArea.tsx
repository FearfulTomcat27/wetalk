"use client";

import { useEffect, useRef } from "react";
import type { Message } from "@/types/chat";
import { cn } from "@/lib/utils";
import { Avatar } from "@/components/Avatar";
import { useAuthStore } from "@/stores/auth";

interface ChatAreaProps {
  messages: Message[];
  currentUserId: number;
  /** 联系人昵称，用于头像 fallback 首字母 */
  contactName?: string;
  /** 联系人用户名，用于 DiceBear 头像 seed */
  contactUsername?: string;
  /** 联系人头像 URL（优先使用） */
  contactAvatar?: string;
}

/**
 * 格式化时间戳，当天显示 HH:mm，跨天显示 MM-DD HH:mm
 */
function formatTime(ts: string | number): string {
  const date = new Date(ts);
  const now = new Date();
  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");

  const isToday =
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate();

  if (isToday) {
    return `${hours}:${minutes}`;
  }
  const month = (date.getMonth() + 1).toString().padStart(2, "0");
  const day = date.getDate().toString().padStart(2, "0");
  return `${month}-${day} ${hours}:${minutes}`;
}

/**
 * 判断是否需要在两条消息之间展示时间分隔符（间隔 > 2 分钟）
 */
function shouldShowTime(messages: Message[], index: number): boolean {
  if (index === 0) {return true;}
  return new Date(messages[index].created_at).getTime() - new Date(messages[index - 1].created_at).getTime() > 120_000;
}

export function ChatArea({ messages, currentUserId, contactName, contactUsername, contactAvatar }: ChatAreaProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const prevCountRef = useRef(messages.length);

  const user = useAuthStore((s) => s.user);

  const selfUsername = user?.username ?? "me";
  const otherUsername = contactUsername ?? contactName ?? "?";

  
  // 新消息到达时自动滚动到底部
  useEffect(() => {
    const isNew = messages.length > prevCountRef.current;
    prevCountRef.current = messages.length;
    bottomRef.current?.scrollIntoView({
      behavior: isNew ? "smooth" : "auto",
    });
  }, [messages]);

  // 首次渲染滚动到底部
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "auto" });
  }, []);

  if (messages.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <div className="text-center">
          <div className="mx-auto mb-3 flex size-16 items-center justify-center rounded-full bg-muted/60">
            <svg
              className="size-8 text-muted-foreground/25"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"
              />
            </svg>
          </div>
          <p className="text-sm text-muted-foreground">暂无消息</p>
          <p className="mt-1 text-xs text-muted-foreground/50">
            发送一条消息开始对话
          </p>
        </div>
      </div>
    );
  }

  return (
    <div
      ref={scrollRef}
      className="flex-1 overflow-y-auto px-4 py-3"
    >
      <div className="mx-auto">
        {messages.map((msg, index) => {
          const isSelf = msg.sender_id === currentUserId;
          const showTime = shouldShowTime(messages, index);

          return (
            <div key={msg.id}>
              {/* 时间分隔符 */}
              {showTime && (
                <div className="flex items-center justify-center py-3">
                  <span className="select-none rounded-md bg-muted/50 px-3 py-0.5 text-[11px] leading-relaxed text-muted-foreground/80">
                    {formatTime(msg.created_at)}
                  </span>
                </div>
              )}

              {/* 消息行 */}
              <div
                className={cn(
                  "flex animate-message-in items-start gap-2.5 mb-4",
                  isSelf ? "flex-row-reverse" : "flex-row",
                )}
              >
                {/* 头像 */}
                <Avatar
                  src={isSelf ? user?.avatar : contactAvatar}
                  alt={isSelf ? selfUsername : otherUsername}
                  fallbackInitial={isSelf ? "我" : contactName?.[0]?.toUpperCase() || "?"}
                  size={36}
                />

                {/* 气泡 */}
                <div className="relative max-w-[62%]">
                  <div
                    className={cn(
                      "rounded-md px-3.5 py-2 text-sm leading-normal break-words whitespace-pre-wrap shadow-md",
                      isSelf
                        ? "bg-[#3b82f6] text-white"
                        : "bg-[#eeeef0] text-gray-900",
                    )}
                  >
                    <p>{msg.content}</p>
                  </div>
                  {/* 曲线箭头 — 与头像居中对齐 */}
                  {isSelf ? (
                    <svg
                      className="absolute"
                      width="5"
                      height="15"
                      viewBox="0 0 5 15"
                      style={{ right: -5, top: 11 }}
                    >
                      <path d="M 0,0 C 0,3 5,5.5 5,7.5 C 5,9.5 0,12 0,15" fill="#3b82f6" />
                    </svg>
                  ) : (
                    <svg
                      className="absolute"
                      width="5"
                      height="15"
                      viewBox="0 0 5 15"
                      style={{ left: -5, top: 11 }}
                    >
                      <path d="M 5,0 C 5,3 0,5.5 0,7.5 C 0,9.5 5,12 5,15" fill="#eeeef0" />
                    </svg>
                  )}
                </div>
              </div>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
