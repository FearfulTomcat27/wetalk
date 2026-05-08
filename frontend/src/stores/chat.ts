import { create } from "zustand";
import type { Contact, Message } from "@/types/chat";

interface ChatState {
  contacts: Contact[];
  activeContactId: number | null;
  /** contactId → Message[] */
  messages: Record<number, Message[]>;
  /** contactId → 输入框草稿文本 */
  inputTexts: Record<number, string>;

  setContacts: (contacts: Contact[]) => void;
  selectContact: (id: number) => void;
  setInputText: (contactId: number, text: string) => void;
  sendMessage: () => void;
  addContact: (contact: Contact) => void;
  loadMessages: (contactId: number, msgs: Message[]) => void;
  receiveMessage: (msg: Message) => void;
}

let nextMessageId = 1;

export const useChatStore = create<ChatState>((set, get) => ({
  contacts: [],
  activeContactId: null,
  messages: {},
  inputTexts: {},

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
    const { activeContactId, inputTexts, contacts } = get();
    if (activeContactId === null) {
      return;
    }
    const text = (inputTexts[activeContactId] || "").trim();
    if (text === "") {
      return;
    }

    const newMsg: Message = {
      id: nextMessageId++,
      contactId: activeContactId,
      senderId: 0, // 0 = 自己
      content: text,
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
        c.id === activeContactId ? { ...c, lastMessage: text } : c
      ),
      inputTexts: { ...state.inputTexts, [activeContactId]: "" },
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
      messages: { ...state.messages, [contactId]: msgs },
    }));
  },

  receiveMessage: (msg: Message) => {
    set((state) => {
      const existing = state.messages[msg.contactId] || [];
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
            ? { ...c, lastMessage: msg.content, unread: isActive ? c.unread : c.unread + 1 }
            : c
        ),
      };
    });
  },
}));
