import { create } from "zustand";
import type { Contact, Message } from "@/types/chat";

interface ChatState {
  contacts: Contact[];
  activeContactId: number | null;
  /** contactId → Message[] */
  messages: Record<number, Message[]>;
  inputText: string;

  setContacts: (contacts: Contact[]) => void;
  selectContact: (id: number) => void;
  setInputText: (text: string) => void;
  sendMessage: () => void;
  addContact: (contact: Contact) => void;
  /** 从服务端批量加载消息，合并到本地 state */
  loadMessages: (contactId: number, msgs: Message[]) => void;
  /** 由后端推送或轮询到达的新消息 */
  receiveMessage: (msg: Message) => void;
}

let nextMessageId = 1;

export const useChatStore = create<ChatState>((set, get) => ({
  contacts: [],
  activeContactId: null,
  messages: {},
  inputText: "",

  setContacts: (contacts: Contact[]) => {
    set({ contacts });
  },

  selectContact: (id: number) => {
    set({ activeContactId: id });
    // 清除该联系人的未读数
    set((state) => ({
      contacts: state.contacts.map((c) =>
        c.id === id ? { ...c, unread: 0 } : c
      ),
    }));
  },

  setInputText: (text: string) => {
    set({ inputText: text });
  },

  sendMessage: () => {
    const { activeContactId, inputText, contacts } = get();
    if (activeContactId === null || inputText.trim() === "") {
      return;
    }

    const newMsg: Message = {
      id: nextMessageId++,
      contactId: activeContactId,
      senderId: 0, // 0 = 自己
      content: inputText.trim(),
      timestamp: Date.now(),
    };

    set((state) => ({
      messages: {
        ...state.messages,
        [activeContactId]: [
          ...(state.messages[activeContactId] || []),
          newMsg,
        ],
      },
      contacts: contacts.map((c) =>
        c.id === activeContactId
          ? { ...c, lastMessage: inputText.trim() }
          : c
      ),
      inputText: "",
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

  loadMessages: (contactId: number, msgs: Message[]) => {
    set((state) => ({
      messages: {
        ...state.messages,
        [contactId]: msgs,
      },
    }));
  },

  receiveMessage: (msg: Message) => {
    set((state) => {
      const existing = state.messages[msg.contactId] || [];
      // 去重
      if (existing.some((m) => m.id === msg.id)) {
        return state;
      }

      const isActive = state.activeContactId === msg.contactId;

      return {
        messages: {
          ...state.messages,
          [msg.contactId]: [...existing, msg],
        },
        contacts: state.contacts.map((c) =>
          c.id === msg.contactId
            ? {
                ...c,
                lastMessage: msg.content,
                unread: isActive ? c.unread : c.unread + 1,
              }
            : c
        ),
      };
    });
  },
}));
