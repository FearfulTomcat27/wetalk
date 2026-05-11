"use client";

import { useState, useCallback, useRef, useEffect, type MouseEvent as ReactMouseEvent } from "react";
import { useSearchParams } from "next/navigation";
import { ContactList } from "@/components/ContactList";
import { ChatArea } from "@/components/ChatArea";
import { ChatInput } from "@/components/ChatInput";
import { useChatStore } from "@/stores/chat";
import { useAuthStore } from "@/stores/auth";
import { wsClient } from "@/lib/ws";
import { getFriends, getMessages, markAsRead } from "@/lib/api";
import type { FriendInfo } from "@/lib/api";
import type { Contact } from "@/types/chat";
import { Loader2 } from "lucide-react";

// 宽度常量
const DEFAULT_CONTACT_WIDTH = 280;
const MIN_CONTACT_WIDTH = 200;
const MAX_CONTACT_WIDTH = 480;
const MIN_CHAT_WIDTH = 300;

/** 从 OSS URL 提取文件名：uploads/file/1/1712345678_photo.jpg → photo.jpg */
function extractFilename(url: string): string {
  try {
    const lastSegment = new URL(url).pathname.split("/").pop() || "";
    const idx = lastSegment.indexOf("_");
    return idx !== -1 ? lastSegment.slice(idx + 1) : lastSegment;
  } catch {
    return url;
  }
}

/** 根据消息类型格式化联系人列表预览文本 */
function formatMessagePreview(content: string, contentType?: string): string {
  if (!contentType || contentType === "text") {return content;}
  if (contentType === "image") {return "[图片]";}
  if (contentType === "file") {return `[文件] ${extractFilename(content)}`;}
  return content;
}

function friendToContact(f: FriendInfo): Contact {
  return {
    id: f.friend_id,
    chat_id: f.chat_id,
    username: f.friend_name,
    nickname: f.friend_name,
    avatar: f.friend_avatar,
    lastMessage: f.last_message ? formatMessagePreview(f.last_message, f.last_message_type) : undefined,
    lastMessageType: f.last_message_type,
    lastMessageTime: f.last_message_time,
    unread: f.unread_count ?? 0,
  };
}

export default function ChatPage() {
  const searchParams = useSearchParams();
  const currentUserId = useAuthStore((s) => s.user?.id);
  const token = useAuthStore((s) => s.token);
  const userFetched = useAuthStore((s) => s._userFetched);

  // Store 状态
  const contacts = useChatStore((s) => s.contacts);
  const activeContactId = useChatStore((s) => s.activeContactId);
  const messages = useChatStore((s) => s.messages);
  const inputTexts = useChatStore((s) => s.inputTexts);
  const sending = useChatStore((s) => s.sending);
  const connected = useChatStore((s) => s.connected);
  const quotedMessages = useChatStore((s) => s.quotedMessages);
  const clearQuotedMessage = useChatStore((s) => s.clearQuotedMessage);
  const setContacts = useChatStore((s) => s.setContacts);
  const setActiveContactId = useChatStore((s) => s.setActiveContactId);
  const selectContact = useChatStore((s) => s.selectContact);
  const setInputText = useChatStore((s) => s.setInputText);
  const sendMessage = useChatStore((s) => s.sendMessage);
  const sendMediaMessage = useChatStore((s) => s.sendMediaMessage);
  const loadMessages = useChatStore((s) => s.loadMessages);
  const receiveMessage = useChatStore((s) => s.receiveMessage);
  const updateMessageStatus = useChatStore((s) => s.updateMessageStatus);
  const setConnected = useChatStore((s) => s.setConnected);

  // 仅保留 UI 相关的局部 state
  const [contactsLoading, setContactsLoading] = useState(true);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [contactWidth, setContactWidth] = useState(DEFAULT_CONTACT_WIDTH);
  const [dragging, setDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // 加载好友列表 → 写入 store
  useEffect(() => {
    async function load() {
      setContactsLoading(true);
      try {
        const res = await getFriends();
        setContacts((res.data || []).map(friendToContact));
      } catch {
        // 错误已在拦截器 toast
      } finally {
        setContactsLoading(false);
      }
    }
    load();
  }, [setContacts]);

  // 通过 URL query param ?contact=xxx 自动选中联系人
  useEffect(() => {
    const contactParam = searchParams.get("contact");
    if (contactParam && contacts.length > 0) {
      const contactId = Number(contactParam);
      if (!Number.isNaN(contactId) && contacts.some((c) => c.id === contactId)) {
        setActiveContactId(contactId);
      }
    }
  }, [searchParams, contacts.length, setActiveContactId]);

  // 选中联系人时加载历史消息 → 写入 store
  useEffect(() => {
    if (activeContactId === null) {
      return;
    }
    async function load() {
      setMessagesLoading(true);
      try {
        const activeChatId = contacts.find((c) => c.id === activeContactId)?.chat_id;
        if (!activeChatId) {
          setMessagesLoading(false);
          return;
        }
        const res = await getMessages(activeChatId);
        loadMessages(activeChatId, (res.data || []).slice().reverse());
      } catch {
        // 错误已在拦截器 toast
      } finally {
        setMessagesLoading(false);
      }
    }
    load();
  }, [activeContactId, loadMessages, contacts]);

  // 选中联系人时自动聚焦输入框
  useEffect(() => {
    if (activeContactId !== null) {
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [activeContactId]);

  // WS 初始化：等 auth store _userFetched && token
  useEffect(() => {
    if (!userFetched || !token) {
      return;
    }

    wsClient.connect();

    const unsubNew = wsClient.on("message.new", (payload) => {
      receiveMessage(payload as import("@/types/chat").Message);
    });
    const unsubSent = wsClient.on("message.sent", (payload) => {
      const msg = payload as import("@/types/chat").Message;
      if (msg.client_msg_id) {
        updateMessageStatus(msg.client_msg_id, msg);
      }
    });
    const unsubAuthOk = wsClient.on("auth.ok", () => {
      setConnected(true);
    });
    const unsubDisconnect = wsClient.on("disconnect", () => {
      setConnected(false);
    });

    return () => {
      unsubNew();
      unsubSent();
      unsubAuthOk();
      unsubDisconnect();
      wsClient.disconnect();
    };
  }, [userFetched, token, receiveMessage, updateMessageStatus, setConnected]);

  function handleSelectContact(id: number) {
    selectContact(id);
    const chatId = contacts.find((c) => c.id === id)?.chat_id;
    if (chatId) {
      markAsRead(chatId).catch(() => {});
    }
  }

  const chatInputValue = activeContactId ? (inputTexts[activeContactId] ?? "") : "";
  const isSending = activeContactId ? (sending[activeContactId] ?? false) : false;
  const activeContact = contacts.find((c) => c.id === activeContactId) ?? null;
  const activeMessages = activeContact ? messages[activeContact.chat_id] ?? [] : [];

  // --- 拖拽处理 ---
  const handleMouseDown = useCallback((e: ReactMouseEvent) => {
    e.preventDefault();
    setDragging(true);
  }, []);

  useEffect(() => {
    if (!dragging) {return;}

    function handleMouseMove(e: globalThis.MouseEvent) {
      if (!containerRef.current) {return;}
      const rect = containerRef.current.getBoundingClientRect();
      const newWidth = e.clientX - rect.left;
      const maxWidth = Math.min(
        rect.width - MIN_CHAT_WIDTH,
        MAX_CONTACT_WIDTH,
      );
      setContactWidth(Math.min(Math.max(newWidth, MIN_CONTACT_WIDTH), maxWidth));
    }

    function handleMouseUp() {
      setDragging(false);
    }

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    document.body.style.userSelect = "none";
    document.body.style.cursor = "col-resize";

    return () => {
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
      document.body.style.userSelect = "";
      document.body.style.cursor = "";
    };
  }, [dragging]);

  return (
    <div ref={containerRef} className="flex flex-1 overflow-hidden">
      {/* 联系人列表 */}
      {contactsLoading ? (
        <div className="flex items-center justify-center" style={{ width: contactWidth }}>
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <ContactList
          contacts={contacts}
          activeContactId={activeContactId}
          onSelectContact={handleSelectContact}
          style={{ width: contactWidth }}
          showAddFriend
        />
      )}

      {/* 拖拽手柄 */}
      <div
        onMouseDown={handleMouseDown}
        className={`relative shrink-0 cursor-col-resize transition-colors ${dragging ? " bg-primary/10" : "hover:bg-primary/30"}`}
      >
        <div
          className={`absolute inset-y-0 left-1/2 w-px -translate-x-1/2 transition-colors ${dragging ? "bg-primary/10" : "bg-border group-hover:bg-primary/40"}`}
        />
      </div>

      {/* 第三列：聊天区域 */}
      <div className="flex flex-1 flex-col overflow-hidden bg-muted/30">
        {activeContact ? (
          <>
            <div className="flex h-14 shrink-0 items-center bg-card px-4">
              <div className="flex flex-col">
                <p className="text-sm font-medium">{activeContact.nickname}</p>
                <p className="text-xs text-muted-foreground">
                  @{activeContact.username}
                </p>
              </div>
              {!connected && (
                <span className="ml-2 text-xs text-muted-foreground/60">（离线模式）</span>
              )}
            </div>

            {messagesLoading ? (
              <div className="flex flex-1 items-center justify-center">
                <Loader2 className="size-6 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <ChatArea
                messages={activeMessages}
                currentUserId={currentUserId ?? 0}
                contactName={activeContact.nickname}
                contactUsername={activeContact.username}
                contactAvatar={activeContact.avatar}
              />
            )}

            <ChatInput
              ref={inputRef}
              value={chatInputValue}
              disabled={isSending}
              onChange={(text) => {
                if (activeContactId !== null) {
                  setInputText(activeContactId, text);
                }
              }}
              onSend={sendMessage}
              onSendMedia={sendMediaMessage}
              quotedMessage={
                activeContactId !== null
                  ? quotedMessages[String(activeContactId)] ?? null
                  : null
              }
              onClearQuote={() => {
                if (activeContactId !== null) {
                  clearQuotedMessage(activeContactId);
                }
              }}
            />
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center">
            <div className="text-center">
              <div className="mx-auto mb-4 flex size-20 items-center justify-center rounded-full bg-muted">
                <svg
                  className="size-10 text-muted-foreground/40"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={1.5}
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                  />
                </svg>
              </div>
              <p className="text-lg font-medium text-muted-foreground">
                选择一个联系人开始聊天
              </p>
              <p className="mt-1 text-sm text-muted-foreground/60">
                从左侧列表中选择一个联系人
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}