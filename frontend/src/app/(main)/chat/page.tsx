"use client";

import { useState, useCallback } from "react";
import { ContactList } from "@/components/ContactList";
import { ChatArea } from "@/components/ChatArea";
import { ChatInput } from "@/components/ChatInput";
import type { Contact, Message } from "@/types/chat";
import {useAuthStore} from "@/stores/auth";

// 模拟当前登录用户 ID（后续从 store 获取）
const CURRENT_USER_ID = 1;

// 模拟联系人数据
const mockContacts: Contact[] = [
  {
    id: 2,
    username: "alice",
    nickname: "Alice",
    lastMessage: "好的，明天见！",
    unread: 3,
  },
  {
    id: 3,
    username: "bob",
    nickname: "Bob",
    lastMessage: "那个项目进展如何？",
    unread: 0,
  },
  {
    id: 4,
    username: "carol",
    nickname: "Carol",
    lastMessage: "谢谢你的帮助 🙏",
    unread: 1,
  },
  {
    id: 5,
    username: "dave",
    nickname: "Dave",
    lastMessage: "晚上一起吃饭吗？",
    unread: 0,
  },
  {
    id: 6,
    username: "eve",
    nickname: "Eve",
    lastMessage: "文件我已经发你了",
    unread: 5,
  },
];

// 模拟消息数据
const mockMessages: Record<number, Message[]> = {
  2: [
    { id: 1, contactId: 2, senderId: 2, content: "你好！", timestamp: Date.now() - 3600000 },
    { id: 2, contactId: 2, senderId: 1, content: "你好 Alice！", timestamp: Date.now() - 3500000 },
    { id: 3, contactId: 2, senderId: 2, content: "明天有空吗？", timestamp: Date.now() - 3400000 },
    { id: 4, contactId: 2, senderId: 1, content: "有的，几点？", timestamp: Date.now() - 3300000 },
    { id: 5, contactId: 2, senderId: 2, content: "下午三点可以吗？", timestamp: Date.now() - 3200000 },
    { id: 6, contactId: 2, senderId: 1, content: "没问题 👌", timestamp: Date.now() - 3100000 },
    { id: 7, contactId: 2, senderId: 2, content: "好的，明天见！", timestamp: Date.now() - 3000000 },
  ],
  3: [
    { id: 8, contactId: 3, senderId: 3, content: "那个项目进展如何？", timestamp: Date.now() - 7200000 },
  ],
  4: [
    { id: 9, contactId: 4, senderId: 4, content: "能帮我看下这个 bug 吗？", timestamp: Date.now() - 86400000 },
    { id: 10, contactId: 4, senderId: 1, content: "当然，发我看看", timestamp: Date.now() - 86000000 },
    { id: 11, contactId: 4, senderId: 4, content: "谢谢你的帮助 🙏", timestamp: Date.now() - 85000000 },
  ],
};

export default function ChatPage() {
  const [contacts] = useState<Contact[]>(mockContacts);
  const [activeContactId, setActiveContactId] = useState<number | null>(null);
  const [messages, setMessages] = useState<Record<number, Message[]>>(mockMessages);

  const activeContact = contacts.find((c) => c.id === activeContactId) ?? null;
  const activeMessages = activeContactId ? messages[activeContactId] ?? [] : [];

  const handleSendMessage = useCallback(
    (content: string) => {
      if (!activeContactId) {return;}
      const newMsg: Message = {
        id: Date.now(),
        contactId: activeContactId,
        senderId: CURRENT_USER_ID,
        content,
        timestamp: Date.now(),
      };
      setMessages((prev) => ({
        ...prev,
        [activeContactId]: [...(prev[activeContactId] ?? []), newMsg],
      }));
    },
    [activeContactId],
  );

  return (
    <div className="flex flex-1 overflow-hidden">
      {/* 左侧联系人列表 */}
      <ContactList
        contacts={contacts}
        activeContactId={activeContactId}
        onSelectContact={setActiveContactId}
      />

      {/* 右侧聊天区域 */}
      <div className="flex flex-1 flex-col overflow-hidden bg-muted/30">
        {activeContact ? (
          <>
            {/* 聊天头部 */}
            <div className="flex h-14 shrink-0 items-center gap-3 border-b bg-card px-4">
              {/* 头像占位 */}
              <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-medium text-primary">
                {activeContact.nickname.charAt(0)}
              </div>
              <div className="flex flex-col">
                <p className="text-sm font-medium">{activeContact.nickname}</p>
                <p className="text-xs text-muted-foreground">
                  @{activeContact.username}
                </p>
              </div>
            </div>

            {/* 消息列表 */}
            <ChatArea
              messages={activeMessages}
              currentUserId={CURRENT_USER_ID}
              contactName={activeContact.nickname}
            />

            {/* 输入框 */}
            <ChatInput onSend={handleSendMessage} />
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
