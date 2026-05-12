"use client";

import { useState, useEffect, useRef } from "react";
import Image from "next/image";
import type { Message } from "@/types/chat";
import { cn } from "@/lib/utils";
import { Avatar } from "@/components/Avatar";
import { useAuthStore } from "@/stores/auth";
import { Download, X } from "lucide-react";
import { MessageContextMenu } from "@/components/MessageContextMenu";
import { useChatStore } from "@/stores/chat";
import { deleteMessage } from "@/lib/api/messages";
import { toast } from "sonner";

/** 根据文件扩展名返回对应的图标颜色和显示文字 */
function getFileTypeInfo(mimeType?: string, fileName?: string): { color: string; label: string } {
  if (!mimeType && !fileName) {
    return { color: "#8b8b8b", label: "?" };
  }
  // 从 mime_type 或文件名提取扩展名
  let ext = "";
  if (fileName) {
    const parts = fileName.split(".");
    if (parts.length > 1) {
      ext = parts[parts.length - 1].toUpperCase();
    }
  }
  if (!ext && mimeType) {
    const mimeMap: Record<string, string> = {
      "application/pdf": "PDF",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document": "DOCX",
      "application/msword": "DOC",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "XLSX",
      "application/vnd.ms-excel": "XLS",
      "application/vnd.openxmlformats-officedocument.presentationml.presentation": "PPTX",
      "application/vnd.ms-powerpoint": "PPT",
      "application/zip": "ZIP",
      "application/x-rar-compressed": "RAR",
      "application/x-7z-compressed": "7Z",
      "text/plain": "TXT",
      "application/json": "JSON",
    };
    ext = mimeMap[mimeType] || mimeType.split("/").pop()?.toUpperCase() || "";
  }
  // 扩展名对应颜色
  const colorMap: Record<string, string> = {
    PDF: "#e74c3c",
    DOC: "#2979ff",
    DOCX: "#2979ff",
    XLS: "#27ae60",
    XLSX: "#27ae60",
    PPT: "#e67e22",
    PPTX: "#e67e22",
    ZIP: "#f39c12",
    RAR: "#f39c12",
    "7Z": "#f39c12",
    TXT: "#8b8b8b",
    JSON: "#8b8b8b",
  };
  return { color: colorMap[ext] || "#8b8b8b", label: ext || "?" };
}
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import * as VisuallyHidden from "@radix-ui/react-visually-hidden";

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
 * 格式化时间戳
 * - 昨天：显示"昨天 HH:mm"
 * - 当前周（周一~周日）：显示"星期X HH:mm"（如"星期一 14:30"）
 * - 非当前周但同年：显示"M月D日 HH:mm"（如"5月10日 14:30"）
 * - 非今年：显示"YYYY年M月D日 HH:mm"（如"2025年12月28日 14:30"）
 */
/** 引用消息预览组件：显示发送者 + 右侧竖线 + 图片缩略图/文件图标/文字 */
function QuotedPreview({ msg, isSelf }: { msg: Message; isSelf: boolean }) {
  const otherName = msg.quoted_sender_name || "未知用户";
  const selfName = "你";
  const displayName = msg.quoted_sender_id === msg.sender_id ? selfName : otherName;

  const nameColor = isSelf ? "text-blue-400" : "text-gray-500";
  const contentColor = "text-muted-foreground";

  return (
    <div className={cn(
      "mt-1 flex items-center gap-2 rounded border-r-2 pr-2 text-xs leading-snug max-w-xs",
      isSelf ? "border-blue-500/60" : "border-gray-400",
    )}>
      <span className={cn("shrink-0 font-medium", nameColor)}>
        {displayName}：
      </span>
      {msg.quoted_content_type === "image" ? (
        (() => {
          const imgUrl = msg.quoted_file_metadata?.url || msg.quoted_content;
          return imgUrl ? (
            <div className="relative size-[60px] shrink-0 overflow-hidden rounded border">
              <Image
                src={imgUrl}
                alt="引用图片"
                fill
                sizes="60px"
                className="object-cover"
                unoptimized={imgUrl.includes("oss-cn-shanghai")}
              />
            </div>
          ) : (
            <span className={cn("truncate", contentColor)}>[图片]</span>
          );
        })()
      ) : msg.quoted_content_type === "file" ? (
        <div className="flex items-center gap-1.5 min-w-0">
          <span className={cn("truncate", contentColor)}>
            {msg.quoted_file_metadata?.original_name || "[文件]"}
          </span>
          {(() => {
            const typeInfo = getFileTypeInfo(
              msg.quoted_file_metadata?.mime_type,
              msg.quoted_file_metadata?.original_name,
            );
            return (
              <div
                className="flex size-6 shrink-0 items-center justify-center rounded text-white text-[9px] font-bold leading-none"
                style={{ backgroundColor: typeInfo.color }}
              >
                {typeInfo.label}
              </div>
            );
          })()}
        </div>
      ) : (
        <span className={cn("truncate", contentColor)}>
          {msg.quoted_content && msg.quoted_content.length > 50
            ? msg.quoted_content.slice(0, 50) + "..."
            : msg.quoted_content}
        </span>
      )}
    </div>
  );
}

function formatTime(ts: string | number): string {
  const date = new Date(ts);
  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");
  const timeStr = `${hours}:${minutes}`;

  const now = new Date();

  // 计算本周一 00:00:00
  const dayOfWeek = now.getDay(); // 0=周日, 1=周一, ...
  const mondayDiff = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
  const monday = new Date(now);
  monday.setDate(now.getDate() - mondayDiff);
  monday.setHours(0, 0, 0, 0);

  // 本周日 23:59:59
  const sunday = new Date(monday);
  sunday.setDate(monday.getDate() + 6);
  sunday.setHours(23, 59, 59, 999);

  if (date >= monday && date <= sunday) {
    // 今天只显示时间
    if (
      date.getFullYear() === now.getFullYear() &&
      date.getMonth() === now.getMonth() &&
      date.getDate() === now.getDate()
    ) {
      return timeStr;
    }
    // 昨天
    const yesterday = new Date(now);
    yesterday.setDate(now.getDate() - 1);
    if (
      date.getFullYear() === yesterday.getFullYear() &&
      date.getMonth() === yesterday.getMonth() &&
      date.getDate() === yesterday.getDate()
    ) {
      return `昨天 ${timeStr}`;
    }
    const weekdays = ["星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"];
    return `${weekdays[date.getDay()]} ${timeStr}`;
  }

  // 非当前周
  const month = (date.getMonth() + 1).toString();
  const day = date.getDate().toString();

  if (date.getFullYear() === now.getFullYear()) {
    return `${month}月${day}日 ${timeStr}`;
  }
  return `${date.getFullYear()}年${month}月${day}日 ${timeStr}`;
}

/**
 * 格式化文件大小
 * < 1KB → "xxx B", < 1MB → "xxx KB", ≥ 1MB → "xxx.x MB"
 */
function formatFileSize(bytes?: number): string {
  if (bytes === undefined || bytes === null) {return "";}
  if (bytes < 1024) {return `${bytes} B`;}
  if (bytes < 1024 * 1024) {return `${Math.floor(bytes / 1024)} KB`;}
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** 从 OSS URL 提取文件名 */
function extractFilename(url: string): string {
  try {
    const lastSegment = new URL(url).pathname.split("/").pop() || "";
    const idx = lastSegment.indexOf("_");
    return idx !== -1 ? lastSegment.slice(idx + 1) : lastSegment;
  } catch {
    return url;
  }
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
  const [previewImage, setPreviewImage] = useState<string | null>(null);
  const [filePreview, setFilePreview] = useState<Message | null>(null);

  // 右键菜单状态
  const [contextMenu, setContextMenu] = useState<{
    message: Message;
    x: number;
    y: number;
  } | null>(null);

  const user = useAuthStore((s) => s.user);
  const activeContactId = useChatStore((s) => s.activeContactId);
  const setQuotedMessage = useChatStore((s) => s.setQuotedMessage);

  const selfUsername = user?.username ?? "me";
  const otherUsername = contactUsername ?? contactName ?? "?";

  // Esc 关闭图片预览
  useEffect(() => {
    if (!previewImage) {return;}
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {setPreviewImage(null);}
    }
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [previewImage]);

  // Esc 关闭文件预览
  useEffect(() => {
    if (!filePreview) {return;}
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {setFilePreview(null);}
    }
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [filePreview]);

  
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
    <>
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
                onContextMenu={(e) => {
                  e.preventDefault();
                  setContextMenu({ message: msg, x: e.clientX, y: e.clientY });
                }}
              >
                {/* 头像 */}
                <Avatar
                  src={isSelf ? user?.avatar : contactAvatar}
                  alt={isSelf ? selfUsername : otherUsername}
                  fallbackInitial={isSelf ? "我" : contactName?.[0]?.toUpperCase() || "?"}
                  size={36}
                />

                {/* 气泡 / 图片缩略图 */}
                <div className="relative max-w-[62%]">
                  {msg.content_type === "image" ? (
                    <div>
                      <div
                        className="relative overflow-hidden rounded-md cursor-pointer shadow-md"
                        style={{ width: 200, height: 150 }}
                        onClick={() => setPreviewImage(msg.content)}
                      >
                        <Image
                          src={msg.content}
                          alt="图片消息"
                          fill
                          sizes="200px"
                          className="object-cover"
                          unoptimized={msg.content.includes("oss-cn-shanghai")}
                        />
                      </div>
                      {/* 被引用消息预览 */}
                      {msg.quoted_content && (
                        <QuotedPreview msg={msg} isSelf={isSelf} />
                      )}
                    </div>
                  ) : (
                    <>
                      <div
                        className={cn(
                          "rounded-md px-3.5 py-2 text-sm leading-normal break-words whitespace-pre-wrap shadow-md",
                          isSelf
                            ? "bg-[#3b82f6] text-white"
                            : "bg-[#eeeef0] text-gray-900",
                          msg.content_type === "file" && "p-2.5",
                        )}
                      >

                        {msg.content_type === "file" ? (
                          <div
                            onClick={() => setFilePreview(msg)}
                            className="flex items-center gap-2 cursor-pointer w-[180px] h-[60px]"
                          >
                            {/* 左侧：文件名 + 文件大小 */}
                            <div className="min-w-0 flex-1">
                              <p className={cn("truncate text-[13px] font-medium leading-snug", isSelf ? "text-white" : "text-gray-800")}>
                                {msg.file_metadata?.original_name || extractFilename(msg.content) || "文件"}
                              </p>
                              {msg.file_metadata?.file_size != null && (
                                <p className={cn("text-[11px] mt-0.5", isSelf ? "text-white/70" : "text-gray-500")}>
                                  {formatFileSize(msg.file_metadata?.file_size)}
                                </p>
                              )}
                            </div>
                            {/* 右侧：文件类型图标方块 */}
                            {(() => {
                              const typeInfo = getFileTypeInfo(msg.file_metadata?.mime_type, msg.file_metadata?.original_name || extractFilename(msg.content));
                              return (
                                <div
                                  className="flex size-10 shrink-0 items-center justify-center rounded-md text-white text-[11px] font-bold leading-none"
                                  style={{ backgroundColor: typeInfo.color }}
                                >
                                  {typeInfo.label}
                                </div>
                              );
                            })()}
                          </div>
                        ) : (
                          <p>{msg.content}</p>
                        )}
                      </div>
                      {/* 被引用消息预览 */}
                      {msg.quoted_content && (
                        <QuotedPreview msg={msg} isSelf={isSelf} />
                      )}
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
                    </>
                  )}
                </div>
              </div>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>
    </div>

    {/* 右键菜单 */}
    {contextMenu && (
      <MessageContextMenu
        x={contextMenu.x}
        y={contextMenu.y}
        message={contextMenu.message}
        isSelf={contextMenu.message.sender_id === currentUserId}
        onClose={() => setContextMenu(null)}
        onDelete={(msg) => {
          deleteMessage(msg.id)
            .then(() => toast.success("消息已删除"))
            .catch(() => {});
        }}
        onQuote={(msg) => {
          if (activeContactId !== null) {
            setQuotedMessage(activeContactId, { id: msg.id, content: msg.content });
          }
        }}
      />
    )}

    {/* 图片预览 */}
    {previewImage && (
      <div
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm"
        onClick={() => setPreviewImage(null)}
      >
        <div className="relative" style={{ width: "90vw", height: "90vh" }}>
          <Image
            src={previewImage}
            alt="图片预览"
            fill
            className="object-contain"
            unoptimized={previewImage.includes("oss-cn-shanghai")}
          />
        </div>
        <button
          className="absolute top-4 right-4 rounded-full bg-white/10 p-2 text-white/80 transition-colors hover:bg-white/20 hover:text-white"
          onClick={() => setPreviewImage(null)}
        >
          <X className="size-6" />
        </button>
      </div>
    )}

    {/* 文件预览 Dialog */}
    <Dialog open={filePreview !== null} onOpenChange={(open) => { if (!open) setFilePreview(null); }}>
      <DialogContent className="sm:max-w-md">
        <VisuallyHidden.Root>
          <DialogTitle>文件预览</DialogTitle>
        </VisuallyHidden.Root>
        <div className="space-y-4">
          {/* 文件信息 — 纵向布局 */}
          <div className="flex flex-col items-center gap-4 p-4">
            {/* 文件类型图标 */}
            {(() => {
              const typeInfo = getFileTypeInfo(filePreview?.file_metadata?.mime_type, filePreview?.file_metadata?.original_name || (filePreview?.content ? extractFilename(filePreview.content) : ""));
              return (
                <div
                  className="flex size-16 shrink-0 items-center justify-center rounded-lg text-white text-base font-bold leading-none"
                  style={{ backgroundColor: typeInfo.color }}
                >
                  {typeInfo.label}
                </div>
              );
            })()}
            {/* 文件名 + 文件大小 */}
            <div className="text-center w-full">
              <p className="truncate text-sm font-medium">{filePreview?.file_metadata?.original_name || (filePreview?.content ? extractFilename(filePreview.content) : "") || "文件"}</p>
              <p className="text-xs text-muted-foreground mt-1">
                {formatFileSize(filePreview?.file_metadata?.file_size)}
              </p>
            </div>
          </div>
          <div className="flex justify-center">
          <button
            onClick={() => {
              const url = filePreview?.content || "";
              const name = filePreview?.file_metadata?.original_name || (filePreview?.content ? extractFilename(filePreview.content) : "") || "file";
              if (url) {
                const a = document.createElement("a");
                a.href = url;
                a.download = name;
                a.click();
              }
            }}
            className="inline-flex cursor-pointer items-center justify-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
          >
            <Download className="size-4" />
            接收文件
          </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </>
  );
}
