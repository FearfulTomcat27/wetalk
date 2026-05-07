export interface Contact {
  id: number;
  username: string;
  nickname: string;
  avatar?: string;
  lastMessage?: string;
  unread: number;
}

export interface Message {
  id: number;
  contactId: number;
  senderId: number;
  content: string;
  timestamp: number; // Unix ms
}

export interface SendMessagePayload {
  contactId: number;
  content: string;
}
