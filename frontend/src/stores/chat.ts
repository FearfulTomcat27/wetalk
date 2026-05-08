import { create } from "zustand";
import type { Contact, Message } from "@/types/chat";
import { wsClient } from "@/lib/ws";
import { sendMessage as sendMessageAPI } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";

interface ChatState {
  contacts: Contact[];
  activeContactId: number | null;
  /** contactId → Message[] */
  messages: Record<number, Message[]>;
  /** contactId → 输入框草稿文本 */
  inputTexts: Record<number, string>;
  connected: boolean;
  /** contactId → 是否正在发送 */
  sending: Record<number, boolean>;

  setContacts: (contacts: Contact[]) => void;
  selectContact: (id: number) => void;
  setInputText: (contactId: number, text: string) => void;
  sendMessage: () => void;
  addContact: (contact: Contact) => void;
  loadMessages: (contactId: number, msgs: Message[]) => void;
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

    const currentUserId = useAuthStore.getState().user?.id ?? 0;
    const clientMsgId = `c_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    const tempMsg: Message = {
      id: -Date.now(),
      sender_id: currentUserId,
      receiver_id: activeContactId,
      content: text,
      created_at: new Date().toISOString(),
      client_msg_id: clientMsgId,
    };

    // 乐观 UI：先插入临时消息，清空输入框
    set((state) => ({
      messages: {
        ...state.messages,
        [activeContactId]: [
          ...(state.messages[activeContactId] || []),
          tempMsg,
        ],
      },
      contacts: contacts.map((c) =>
        c.id === activeContactId ? { ...c, lastMessage: text } : c
      ),
      inputTexts: { ...state.inputTexts, [activeContactId]: "" },
    }));

    if (connected) {
      wsClient.send("message.send", {
        receiver_id: activeContactId,
        content: text,
        client_msg_id: clientMsgId,
      });
    } else {
      // HTTP 降级
      set((state) => ({
        sending: { ...state.sending, [activeContactId]: true },
      }));
      sendMessageAPI({ receiver_id: activeContactId, content: text })
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

  addContact: (contact: Contact) => {
    set((state) => {
      if (state.contacts.some((c) => c.id === contact.id)) {
        return state;
      }
      return { contacts: [...state.contacts, contact] };
    });
  },

  loadMessages: (contactId: number, msgs: Message[]) => {
    set((state) => ({
      messages: { ...state.messages, [contactId]: msgs },
    }));
  },

  receiveMessage: (msg: Message) => {
    set((state) => {
      const existing = state.messages[msg.sender_id] || [];
      if (existing.some((m) => m.id === msg.id)) {
        return state;
      }
      const isActive = state.activeContactId === msg.sender_id;
      return {
        messages: {
          ...state.messages,
          [msg.sender_id]: [...existing, msg],
        },
        contacts: state.contacts.map((c) =>
          c.id === msg.sender_id
            ? { ...c, lastMessage: msg.content, unread: isActive ? c.unread : c.unread + 1 }
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