"use client";

import { useEffect, useRef, useState } from "react";
import type { Message } from "@/types/chat";
import { cn } from "@/lib/utils";
import { getAvatarSrc } from "@/lib/avatar";
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
 * 格式化时间戳为 HH:mm
 */
function formatTime(ts: number): string {
  const date = new Date(ts);
  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");
  return `${hours}:${minutes}`;
}

/**
 * 判断是否需要在两条消息之间展示时间分隔符（间隔 > 5 分钟）
 */
function shouldShowTime(messages: Message[], index: number): boolean {
  if (index === 0) {return true;}
  return messages[index].timestamp - messages[index - 1].timestamp > 300_000;
}

export function ChatArea({ messages, currentUserId, contactName, contactUsername, contactAvatar }: ChatAreaProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const prevCountRef = useRef(messages.length);

  const user = useAuthStore((s) => s.user);
  const [selfAvatarError, setSelfAvatarError] = useState(false);
  const [otherAvatarError, setOtherAvatarError] = useState(false);

  const selfUsername = user?.username ?? "me";
  const otherUsername = contactUsername ?? contactName ?? "?";
  const otherInitial = (contactName || "?").charAt(0).toUpperCase();

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
      <div className="mx-auto space-y-0.5">
        {messages.map((msg, index) => {
          const isSelf = msg.senderId === currentUserId;
          const showTime = shouldShowTime(messages, index);
          const initial = contactName?.charAt(0).toUpperCase() || "?";

          return (
            <div key={msg.id}>
              {/* 时间分隔符 */}
              {showTime && (
                <div className="flex items-center justify-center py-3">
                  <span className="select-none rounded-md bg-muted/50 px-3 py-0.5 text-[11px] leading-relaxed text-muted-foreground/80">
                    {formatTime(msg.timestamp)}
                  </span>
                </div>
              )}

              {/* 消息行 */}
              <div
                className={cn(
                  "flex animate-message-in items-end gap-2",
                  isSelf ? "flex-row-reverse" : "flex-row",
                )}
              >
                {/* 头像 */}
                <div className="mb-0.5 shrink-0">
                  {isSelf ? (
                    selfAvatarError ? (
                      <div className="flex size-8 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary" title="我">
                        我
                      </div>
                    ) : (
                      <img
                        src={getAvatarSrc(user?.avatar, selfUsername)}
                        alt="我"
                        onError={() => setSelfAvatarError(true)}
                        className="size-8 rounded-full object-cover"
                      />
                    )
                  ) : otherAvatarError ? (
                    <div className="flex size-8 items-center justify-center rounded-full bg-muted text-xs font-semibold text-muted-foreground" title={contactName}>
                      {otherInitial}
                    </div>
                  ) : (
                    <img
                      src={getAvatarSrc(contactAvatar, otherUsername)}
                      alt={contactName || "?"}
                      onError={() => setOtherAvatarError(true)}
                      className="size-8 rounded-full object-cover"
                    />
                  )}
                </div>

                {/* 气泡 */}
                <div
                  className={cn(
                    "group relative max-w-[68%] px-4 py-2.5 text-sm leading-relaxed shadow-sm transition-shadow hover:shadow-md",
                    isSelf
                      ? "rounded-2xl rounded-br-md bg-[#95EC69] text-black"
                      : "rounded-2xl rounded-bl-md bg-card text-foreground",
                  )}
                >
                  <p className="whitespace-pre-wrap break-words">
                    {msg.content}
                  </p>
                  {/* 气泡内时间 */}
                  <span
                    className={cn(
                      "mt-1 flex justify-end text-[10px] opacity-45",
                      isSelf ? "text-black" : "text-muted-foreground",
                    )}
                  >
                    {formatTime(msg.timestamp)}
                  </span>
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
