export interface Contact {
  id: number;
  username: string;
  nickname: string;
  avatar?: string;
  lastMessage?: string;
  lastMessageType?: string;
  lastMessageTime?: string;
  unread: number;
}

/** 匹配后端 JSON: { id, sender_id, receiver_id, content, content_type, status, created_at } */
export interface Message {
  id: number;
  sender_id: number;
  receiver_id: number;
  content: string;
  content_type?: string;
  status?: string;
  created_at: string; // ISO 8601
  /** 乐观 UI 匹配：发送时生成，收到 message.sent 后用于替换临时消息 */
  client_msg_id?: string;
  /** 文件消息元数据 */
  file_metadata?: FileMetadata;
}

export interface FileMetadata {
  url?: string;
  original_name?: string;
  file_size?: number;
  mime_type?: string;
  width?: number;
  height?: number;
}
