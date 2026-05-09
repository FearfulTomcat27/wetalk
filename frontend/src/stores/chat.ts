import { create } from "zustand";
import type { Contact, Message, FileMetadata } from "@/types/chat";
import { wsClient } from "@/lib/ws";
import { sendMessage as sendMessageAPI } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";

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
function formatMessagePreview(content: string, contentType?: string): string {
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

  setContacts: (contacts: Contact[]) => void;
  selectContact: (id: number) => void;
  setActiveContactId: (id: number) => void;
  setInputText: (contactId: number, text: string) => void;
  sendMessage: () => void;
  sendMediaMessage: (content: string, contentType: string, fileMetadata?: FileMetadata) => void;
  addContact: (contact: Contact) => void;
  loadMessages: (chatId: number, msgs: Message[]) => void;
  receiveMessage: (msg: Message) => void;
  updateMessageStatus: (clientMsgId: string, serverMsg: Message) => void;
  setConnected: (connected: boolean) => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  contacts: [],
  activeContactId: null,
  messages: {},
  inputTexts: {},
  connected: false,
  sending: {},

  setContacts: (contacts: Contact[]) => {
    set({ contacts });
  },

  selectContact: (id: number) => {
    set({ activeContactId: id });
    set((state) => ({
      contacts: state.contacts.map((c) =>
        c.id === id ? { ...c, unread: 0 } : c
      ),
    }));
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
    const { activeContactId, inputTexts, connected, contacts } = get();
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

    const currentUserId = useAuthStore.getState().user?.id ?? 0;
    const clientMsgId = `c_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    const tempMsg: Message = {
      id: -Date.now(),
      chat_id: chatId,
      sender_id: currentUserId,
      content: text,
      created_at: new Date().toISOString(),
      client_msg_id: clientMsgId,
    };

    // 乐观 UI：先插入临时消息，清空输入框
    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: [
          ...(state.messages[chatId] || []),
          tempMsg,
        ],
      },
      contacts: contacts.map((c) =>
        c.id === activeContactId ? { ...c, lastMessage: text, lastMessageTime: new Date().toISOString() } : c
      ),
      inputTexts: { ...state.inputTexts, [activeContactId]: "" },
    }));

    if (connected) {
      wsClient.send("message.send", {
        chat_id: chatId,
        content: text,
        client_msg_id: clientMsgId,
      });
    } else {
      // HTTP 降级
      set((state) => ({
        sending: { ...state.sending, [activeContactId]: true },
      }));
      sendMessageAPI({ chat_id: chatId, content: text })
        .then((res) => {
          const serverMsg = res.data!;
          get().updateMessageStatus(clientMsgId, serverMsg);
        })
        .catch(() => {
          // 错误已在拦截器 toast
        })
        .finally(() => {
          set((state) => ({
            sending: { ...state.sending, [activeContactId]: false },
          }));
        });
    }
  },

  sendMediaMessage: (content: string, contentType: string, fileMetadata?: FileMetadata) => {
    const { activeContactId, connected, contacts } = get();
    if (activeContactId === null) {
      return;
    }

    const activeContact = contacts.find((c) => c.id === activeContactId);
    const chatId = activeContact?.chat_id;
    if (!chatId) {
      return;
    }

    const currentUserId = useAuthStore.getState().user?.id ?? 0;
    const clientMsgId = `c_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    const lastMsgLabel = fileMetadata?.original_name ? `[文件] ${fileMetadata.original_name}` : formatMessagePreview(content, contentType);

    const tempMsg: Message = {
      id: -Date.now(),
      chat_id: chatId,
      sender_id: currentUserId,
      content,
      content_type: contentType,
      created_at: new Date().toISOString(),
      client_msg_id: clientMsgId,
      file_metadata: fileMetadata,
    };

    set((state) => ({
      messages: {
        ...state.messages,
        [chatId]: [
          ...(state.messages[chatId] || []),
          tempMsg,
        ],
      },
      contacts: contacts.map((c) =>
        c.id === activeContactId ? { ...c, lastMessage: lastMsgLabel, lastMessageTime: new Date().toISOString() } : c
      ),
    }));

    const fmPayload = fileMetadata ? {
      url: fileMetadata.url!,
      original_name: fileMetadata.original_name!,
      file_size: fileMetadata.file_size!,
      mime_type: fileMetadata.mime_type!,
    } : undefined;

    if (connected) {
      wsClient.send("message.send", {
        chat_id: chatId,
        content,
        content_type: contentType,
        client_msg_id: clientMsgId,
        file_metadata: fmPayload,
      });
    } else {
      set((state) => ({
        sending: { ...state.sending, [activeContactId]: true },
      }));
      sendMessageAPI({ chat_id: chatId, content, content_type: contentType, client_msg_id: clientMsgId, file_metadata: fmPayload })
        .then((res) => {
          const serverMsg = res.data!;
          get().updateMessageStatus(clientMsgId, serverMsg);
        })
        .catch(() => {})
        .finally(() => {
          set((state) => ({
            sending: { ...state.sending, [activeContactId]: false },
          }));
        });
    }
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
    set((state) => {
      // 根据 chat_id 查找对应联系人
      const contact = state.contacts.find((c) => c.chat_id === msg.chat_id);
      if (!contact) {
        return state;
      }

      const existing = state.messages[msg.chat_id] || [];
      if (existing.some((m) => m.id === msg.id)) {
        return state;
      }
      const isActive = state.activeContactId === contact.id;
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
      };
    });
  },

  updateMessageStatus: (clientMsgId: string, serverMsg: Message) => {
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
  },

  setConnected: (connected: boolean) => {
    set({ connected });
  },
}));