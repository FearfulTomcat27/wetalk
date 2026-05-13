import { create } from "zustand";
import type { Contact, Message, MessageStatus, FileMetadata } from "@/types/chat";
import { wsClient } from "@/lib/ws";
import { sendMessage as sendMessageAPI } from "@/lib/api";
import type { SendMessageRequest } from "@/lib/api/messages";
import { useAuthStore } from "@/stores/auth";

const WS_SEND_TIMEOUT_MS = 5000;
const HTTP_RETRY_MAX = 3;
const HTTP_RETRY_BASE_MS = 1000;
const PENDING_QUEUE_KEY = "wetalk_pending_messages";

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

/** 根据消息类型格式化预览文本 */
export function formatMessagePreview(content: string, contentType?: string): string {
  if (!contentType || contentType === "text") {return content;}
  if (contentType === "image") {return "[图片]";}
  if (contentType === "file") {return `[文件] ${extractFilename(content)}`;}
  return content;
}

interface ChatState {
  contacts: Contact[];
  activeContactId: number | null;
  /** chatId → Message[] */
  messages: Record<number, Message[]>;
  /** contactId → 输入框草稿文本 */
  inputTexts: Record<number, string>;
  connected: boolean;
  /** contactId → 是否正在发送 */
  sending: Record<number, boolean>;
  /** contactId → 该联系人下的引用消息 */
  quotedMessages: Record<string, { id: number; content: string } | null>;
  /** 待处理的好友请求数量 */
  pendingRequestsCount: number;
  /** chatId → 未读消息数（来自单独接口，供侧边栏角标使用） */
  unreadCounts: Record<number, number>;
  /** client_msg_id → 超时定时器 ID（WS 发送超时用） */
  pendingTimeouts: Record<string, ReturnType<typeof setTimeout>>;

  setContacts: (contacts: Contact[]) => void;
  selectContact: (id: number) => void;
  setActiveContactId: (id: number) => void;
  setInputText: (contactId: number, text: string) => void;
  sendMessage: () => void;
  sendMediaMessage: (content: string, contentType: string, fileMetadata?: FileMetadata) => void;
  setQuotedMessage: (contactId: number, msg: { id: number; content: string }) => void;
  clearQuotedMessage: (contactId: number) => void;
  addContact: (contact: Contact) => void;
  loadMessages: (chatId: number, msgs: Message[]) => void;
  receiveMessage: (msg: Message) => void;
  updateMessageStatus: (clientMsgId: string, serverMsg: Message) => void;
  setConnected: (connected: boolean) => void;
  setPendingRequestsCount: (count: number) => void;
  incrementPendingRequests: () => void;
  decrementPendingRequests: () => void;
  setUnreadCounts: (counts: Record<number, number>) => void;
  removeChatHistory: (chatId: number) => void;
  setMessages: (messages: Record<number, Message[]>) => void;
  sendViaHttp: (clientMsgId: string, data: SendMessageRequest, retryCount: number) => void;
  markMessageFailed: (clientMsgId: string) => void;
  retryMessage: (msg: Message) => void;
  clearAllPendingTimeouts: () => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  contacts: [],
  activeContactId: null,
  messages: {},
  inputTexts: {},
  connected: false,
  sending: {},
  quotedMessages: {},
  pendingRequestsCount: 0,
  unreadCounts: {},
  pendingTimeouts: {},

  setContacts: (contacts: Contact[]) => {
    set({ contacts });
  },

  selectContact: (id: number) => {
    set((state) => {
      const contact = state.contacts.find((c) => c.id === id);
      const chatId = contact?.chat_id;
      return {
        activeContactId: id,
        contacts: state.contacts.map((c) =>
          c.id === id ? { ...c, unread: 0 } : c
        ),
        unreadCounts: chatId
          ? { ...state.unreadCounts, [chatId]: 0 }
          : state.unreadCounts,
      };
    });
  },

  setActiveContactId: (id: number) => {
    set({ activeContactId: id });
  },

  setInputText: (contactId: number, text: string) => {
    set((state) => ({
      inputTexts: { ...state.inputTexts, [contactId]: text },
    }));
  },

  sendMessage: () => {
    const { activeContactId, inputTexts, connected, contacts, quotedMessages } = get();
    if (activeContactId === null) {
      return;
    }
    const text = (inputTexts[activeContactId] || "").trim();
    if (text === "") {
      return;
    }

    const activeContact = contacts.find((c) => c.id === activeContactId);
    const chatId = activeContact?.chat_id;
    if (!chatId) {
      return;
    }

    const quoted = quotedMessages[String(activeContactId)];
    const quotedMessageId = quoted?.id;
    const quotedContent = quoted?.content;

    const currentUserId = useAuthStore.getState().user?.id ?? 0;
    const clientMsgId = `c_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    const tempMsg: Message = {
      id: -Date.now(),
      chat_id: chatId,
      sender_id: currentUserId,
      content: text,
      status: "sending",
      created_at: new Date().toISOString(),
      client_msg_id: clientMsgId,
      quoted_message_id: quotedMessageId,
      quoted_content: quotedContent,
    };

    // 乐观 UI：先插入临时消息，清空输入框，清除引用
    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: [
          ...(state.messages[chatId] || []),
          tempMsg,
        ],
      },
      contacts: state.contacts.map((c) =>
        c.id === activeContactId ? { ...c, lastMessage: text, lastMessageTime: new Date().toISOString() } : c
      ),
      inputTexts: { ...state.inputTexts, [activeContactId]: "" },
      quotedMessages: { ...state.quotedMessages, [String(activeContactId)]: null },
    }));

    // 持久化发送队列
    persistPendingQueue(get());

    if (connected) {
      wsClient.send("message.send", {
        chat_id: chatId,
        content: text,
        client_msg_id: clientMsgId,
        quoted_message_id: quotedMessageId,
      });

      // WS 超时定时器：N 秒内未收到 message.sent 则降级到 HTTP
      const timeoutId = setTimeout(() => {
        const { pendingTimeouts } = get();
        if (!pendingTimeouts[clientMsgId]) return; // 已经被清理（message.sent 先到了）
        get().sendViaHttp(clientMsgId, { chat_id: chatId, content: text, quoted_message_id: quotedMessageId }, 0);
        set((state) => ({
          pendingTimeouts: (() => {
            const next = { ...state.pendingTimeouts };
            delete next[clientMsgId];
            return next;
          })(),
        }));
      }, WS_SEND_TIMEOUT_MS);

      set((state) => ({
        pendingTimeouts: { ...state.pendingTimeouts, [clientMsgId]: timeoutId },
      }));
    } else {
      get().sendViaHttp(clientMsgId, { chat_id: chatId, content: text, quoted_message_id: quotedMessageId }, 0);
    }
  },

  sendMediaMessage: (content: string, contentType: string, fileMetadata?: FileMetadata) => {
    const { activeContactId, connected, contacts, quotedMessages } = get();
    if (activeContactId === null) {
      return;
    }

    const activeContact = contacts.find((c) => c.id === activeContactId);
    const chatId = activeContact?.chat_id;
    if (!chatId) {
      return;
    }

    const quoted = quotedMessages[String(activeContactId)];
    const quotedMessageId = quoted?.id;
    const quotedContent = quoted?.content;

    const currentUserId = useAuthStore.getState().user?.id ?? 0;
    const clientMsgId = `c_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    const lastMsgLabel = fileMetadata?.original_name ? `[文件] ${fileMetadata.original_name}` : formatMessagePreview(content, contentType);

    const tempMsg: Message = {
      id: -Date.now(),
      chat_id: chatId,
      sender_id: currentUserId,
      content,
      content_type: contentType,
      status: "sending",
      created_at: new Date().toISOString(),
      client_msg_id: clientMsgId,
      file_metadata: fileMetadata,
      quoted_message_id: quotedMessageId,
      quoted_content: quotedContent,
    };

    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: [
          ...(state.messages[chatId] || []),
          tempMsg,
        ],
      },
      contacts: state.contacts.map((c) =>
        c.id === activeContactId ? { ...c, lastMessage: lastMsgLabel, lastMessageTime: new Date().toISOString() } : c
      ),
      quotedMessages: { ...state.quotedMessages, [String(activeContactId)]: null },
    }));

    // 持久化发送队列
    persistPendingQueue(get());

    const fmPayload = fileMetadata ? {
      url: fileMetadata.url!,
      original_name: fileMetadata.original_name!,
      file_size: fileMetadata.file_size!,
      mime_type: fileMetadata.mime_type!,
    } : undefined;

    const httpData: SendMessageRequest = {
      chat_id: chatId, content, content_type: contentType,
      client_msg_id: clientMsgId, quoted_message_id: quotedMessageId,
      file_metadata: fmPayload as SendMessageRequest["file_metadata"],
    };

    if (connected) {
      wsClient.send("message.send", {
        chat_id: chatId,
        content,
        content_type: contentType,
        client_msg_id: clientMsgId,
        quoted_message_id: quotedMessageId,
        file_metadata: fmPayload,
      });

      const timeoutId = setTimeout(() => {
        const { pendingTimeouts } = get();
        if (!pendingTimeouts[clientMsgId]) return;
        get().sendViaHttp(clientMsgId, httpData, 0);
        set((state) => ({
          pendingTimeouts: (() => {
            const next = { ...state.pendingTimeouts };
            delete next[clientMsgId];
            return next;
          })(),
        }));
      }, WS_SEND_TIMEOUT_MS);

      set((state) => ({
        pendingTimeouts: { ...state.pendingTimeouts, [clientMsgId]: timeoutId },
      }));
    } else {
      get().sendViaHttp(clientMsgId, httpData, 0);
    }
  },

  setQuotedMessage: (contactId: number, msg: { id: number; content: string }) => {
    set((state) => ({
      quotedMessages: { ...state.quotedMessages, [String(contactId)]: msg },
    }));
  },

  clearQuotedMessage: (contactId: number) => {
    set((state) => ({
      quotedMessages: { ...state.quotedMessages, [String(contactId)]: null },
    }));
  },

  addContact: (contact: Contact) => {
    set((state) => {
      if (state.contacts.some((c) => c.id === contact.id)) {
        return state;
      }
      return { contacts: [...state.contacts, contact] };
    });
  },

  loadMessages: (chatId: number, msgs: Message[]) => {
    set((state) => ({
      messages: { ...state.messages, [chatId]: msgs },
    }));
  },

  receiveMessage: (msg: Message) => {
    // 忽略自己发出的消息（确认通过 message.sent 或 HTTP 响应处理）
    const currentUserId = useAuthStore.getState().user?.id;
    if (msg.sender_id === currentUserId) {
      return;
    }

    set((state) => {
      const contact = state.contacts.find((c) => c.chat_id === msg.chat_id);
      if (!contact) {
        return state;
      }

      const existing = state.messages[msg.chat_id] || [];
      if (existing.some((m) => m.id === msg.id)) {
        return state;
      }

      // 判断当前活跃聊天是否就是消息所属的聊天
      // 直接比较活跃联系人的 chat_id 与消息的 chat_id，避免间接比较 id 导致的偏差
      let isActive = false;
      if (state.activeContactId !== null) {
        const activeContact = state.contacts.find((c) => c.id === state.activeContactId);
        if (activeContact !== undefined) {
          isActive = activeContact.chat_id === msg.chat_id;
        }
      }

      const lastMsgLabel = formatMessagePreview(msg.content, msg.content_type);
      return {
        messages: {
          ...state.messages,
          [msg.chat_id]: [...existing, msg],
        },
        contacts: state.contacts.map((c) =>
          c.chat_id === msg.chat_id
            ? { ...c, lastMessage: lastMsgLabel, lastMessageTime: msg.created_at, unread: isActive ? c.unread : c.unread + 1 }
            : c
        ),
        unreadCounts: {
          ...state.unreadCounts,
          [msg.chat_id]: (state.unreadCounts[msg.chat_id] ?? 0) + (isActive ? 0 : 1),
        },
      };
    });
  },

  updateMessageStatus: (clientMsgId: string, serverMsg: Message) => {
    // 清除该消息的 WS 超时定时器
    const { pendingTimeouts } = get();
    const timer = pendingTimeouts[clientMsgId];
    if (timer) {
      clearTimeout(timer);
      const next = { ...pendingTimeouts };
      delete next[clientMsgId];
      set({ pendingTimeouts: next });
    }

    set((state) => {
      const newMessages: Record<number, Message[]> = {};
      for (const [cid, msgs] of Object.entries(state.messages)) {
        const idx = msgs.findIndex((m) => m.client_msg_id === clientMsgId);
        if (idx !== -1) {
          newMessages[Number(cid)] = [
            ...msgs.slice(0, idx),
            { ...serverMsg, client_msg_id: clientMsgId },
            ...msgs.slice(idx + 1),
          ];
        } else {
          newMessages[Number(cid)] = msgs;
        }
      }
      return { messages: newMessages };
    });

    // 发送成功后清理持久化队列
    persistPendingQueue(get());
  },

  setConnected: (connected: boolean) => {
    set({ connected });
  },

  removeChatHistory: (chatId: number) => {
    set((state) => ({
      messages: { ...state.messages, [chatId]: [] },
      contacts: state.contacts.map((c) =>
        c.chat_id === chatId
          ? { ...c, lastMessage: undefined, lastMessageTime: undefined, lastMessageType: undefined, unread: 0 }
          : c
      ),
      unreadCounts: { ...state.unreadCounts, [chatId]: 0 },
    }));
  },

  setPendingRequestsCount: (count: number) => {
    set({ pendingRequestsCount: count });
  },

  incrementPendingRequests: () => {
    set((state) => ({ pendingRequestsCount: state.pendingRequestsCount + 1 }));
  },

  decrementPendingRequests: () => {
    set((state) => ({ pendingRequestsCount: Math.max(0, state.pendingRequestsCount - 1) }));
  },

  setUnreadCounts: (counts: Record<number, number>) => {
    set({ unreadCounts: counts });
  },

  setMessages: (messages: Record<number, Message[]>) => {
    set({ messages });
  },

  /** HTTP 发送（带指数退避重试），由 sendMessage/sendMediaMessage/retryMessage 调用 */
  sendViaHttp: (clientMsgId: string, data: SendMessageRequest, retryCount: number) => {
    const contactId = get().activeContactId;
    if (contactId !== null) {
      set((state) => ({
        sending: { ...state.sending, [contactId]: true },
      }));
    }
    sendMessageAPI(data)
      .then((res) => {
        const serverMsg = res.data!;
        get().updateMessageStatus(clientMsgId, serverMsg);
      })
      .catch(() => {
        if (retryCount < HTTP_RETRY_MAX) {
          const delay = HTTP_RETRY_BASE_MS * Math.pow(2, retryCount);
          setTimeout(() => {
            get().sendViaHttp(clientMsgId, data, retryCount + 1);
          }, delay);
        } else {
          get().markMessageFailed(clientMsgId);
        }
      })
      .finally(() => {
        if (contactId !== null) {
          set((state) => ({
            sending: { ...state.sending, [contactId]: false },
          }));
        }
      });
  },

  /** 标记消息为发送失败 */
  markMessageFailed: (clientMsgId: string) => {
    // 清理可能的残留定时器
    const { pendingTimeouts } = get();
    const timer = pendingTimeouts[clientMsgId];
    if (timer) {
      clearTimeout(timer);
      const next = { ...pendingTimeouts };
      delete next[clientMsgId];
      set({ pendingTimeouts: next });
    }

    set((state) => {
      const newMessages: Record<number, Message[]> = {};
      for (const [cid, msgs] of Object.entries(state.messages)) {
        const idx = msgs.findIndex((m) => m.client_msg_id === clientMsgId);
        if (idx !== -1) {
          newMessages[Number(cid)] = [
            ...msgs.slice(0, idx),
            { ...msgs[idx], status: "failed" as MessageStatus },
            ...msgs.slice(idx + 1),
          ];
        } else {
          newMessages[Number(cid)] = msgs;
        }
      }
      return { messages: newMessages };
    });

    // 失败后清理持久化队列
    persistPendingQueue(get());
  },

  /** 手动重发失败消息 */
  retryMessage: (msg: Message) => {
    if (!msg.client_msg_id) return;
    const clientMsgId = msg.client_msg_id;

    // 重置状态为 sending
    set((state) => {
      const newMessages: Record<number, Message[]> = {};
      for (const [cid, msgs] of Object.entries(state.messages)) {
        const idx = msgs.findIndex((m) => m.client_msg_id === clientMsgId);
        if (idx !== -1) {
          newMessages[Number(cid)] = [
            ...msgs.slice(0, idx),
            { ...msgs[idx], status: "sending" as MessageStatus },
            ...msgs.slice(idx + 1),
          ];
        } else {
          newMessages[Number(cid)] = msgs;
        }
      }
      return { messages: newMessages };
    });

    const { connected } = get();
    const data: SendMessageRequest = {
      chat_id: msg.chat_id,
      content: msg.content,
      content_type: msg.content_type,
      client_msg_id: clientMsgId,
      quoted_message_id: msg.quoted_message_id,
    };
    if (msg.file_metadata) {
      data.file_metadata = {
        url: msg.file_metadata.url || msg.content,
        original_name: msg.file_metadata.original_name || "",
        file_size: msg.file_metadata.file_size || 0,
        mime_type: msg.file_metadata.mime_type || "",
        width: msg.file_metadata.width,
        height: msg.file_metadata.height,
      };
    }

    if (connected) {
      wsClient.send("message.send", {
        chat_id: msg.chat_id,
        content: msg.content,
        content_type: msg.content_type,
        client_msg_id: clientMsgId,
        quoted_message_id: msg.quoted_message_id,
        file_metadata: msg.file_metadata ? {
          url: msg.file_metadata.url || msg.content,
          original_name: msg.file_metadata.original_name || "",
          file_size: msg.file_metadata.file_size || 0,
          mime_type: msg.file_metadata.mime_type || "",
        } : undefined,
      });

      const timeoutId = setTimeout(() => {
        const { pendingTimeouts } = get();
        if (!pendingTimeouts[clientMsgId]) return;
        get().sendViaHttp(clientMsgId, data, 0);
        set((state) => ({
          pendingTimeouts: (() => {
            const next = { ...state.pendingTimeouts };
            delete next[clientMsgId];
            return next;
          })(),
        }));
      }, WS_SEND_TIMEOUT_MS);

      set((state) => ({
        pendingTimeouts: { ...state.pendingTimeouts, [clientMsgId]: timeoutId },
      }));
    } else {
      get().sendViaHttp(clientMsgId, data, 0);
    }

    persistPendingQueue(get());
  },

  /** 清除所有待处理的超时定时器（用于清理） */
  clearAllPendingTimeouts: () => {
    const { pendingTimeouts } = get();
    for (const clientMsgId of Object.keys(pendingTimeouts)) {
      clearTimeout(pendingTimeouts[clientMsgId]);
    }
    set({ pendingTimeouts: {} });
  },
}));

/** 将 sending 状态的消息持久化到 localStorage */
function persistPendingQueue(state: ChatState) {
  try {
    const sendingMessages: Message[] = [];
    for (const msgs of Object.values(state.messages)) {
      for (const m of msgs) {
        if (m.status === "sending") {
          sendingMessages.push(m);
        }
      }
    }
    if (sendingMessages.length > 0) {
      localStorage.setItem(PENDING_QUEUE_KEY, JSON.stringify(sendingMessages));
    } else {
      localStorage.removeItem(PENDING_QUEUE_KEY);
    }
  } catch {
    // localStorage 不可用时忽略
  }
}

/** 页面加载时从 localStorage 恢复待发送消息，逐个重试 */
export function restorePendingQueue() {
  try {
    const raw = localStorage.getItem(PENDING_QUEUE_KEY);
    if (!raw) return;
    const msgs: Message[] = JSON.parse(raw);
    if (msgs.length === 0) return;

    // 恢复消息到对应 chat 的 message 列表（如果还没有的话）
    const store = useChatStore.getState();
    const newMessages: Record<number, Message[]> = { ...store.messages };
    for (const m of msgs) {
      const existing = newMessages[m.chat_id] || [];
      if (!existing.some((em) => em.client_msg_id === m.client_msg_id)) {
        newMessages[m.chat_id] = [...existing, m];
      }
    }
    store.setMessages(newMessages);

    // 逐个重试发送（间隔 500ms 避免并发）
    msgs.forEach((m, i) => {
      setTimeout(() => {
        if (m.client_msg_id) {
          store.retryMessage(m);
        }
      }, i * 500);
    });

    localStorage.removeItem(PENDING_QUEUE_KEY);
  } catch {
    localStorage.removeItem(PENDING_QUEUE_KEY);
  }
}