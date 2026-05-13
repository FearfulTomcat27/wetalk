export interface Contact {
  id: number;
  chat_id: number;
  username: string;
  nickname: string;
  avatar?: string;
  lastMessage?: string;
  lastMessageType?: string;
  lastMessageTime?: string;
  unread: number;
}

export type MessageStatus = "sending" | "sent" | "delivered" | "read" | "failed";

/** 匹配后端 JSON: { id, chat_id, sender_id, content, content_type, status, created_at } */
export interface Message {
  id: number;
  chat_id: number;
  sender_id: number;
  content: string;
  content_type?: string;
  status: MessageStatus;
  created_at: string; // ISO 8601
  /** 乐观 UI 匹配：发送时生成，收到 message.sent 后用于替换临时消息 */
  client_msg_id?: string;
  /** 文件消息元数据 */
  file_metadata?: FileMetadata;
  /** 引用回复的消息 ID */
  quoted_message_id?: number;
  /** 被引用消息的内容（用于预览展示） */
  quoted_content?: string;
  /** 被引用消息的发送者 ID */
  quoted_sender_id?: number;
  /** 被引用消息的发送者用户名 */
  quoted_sender_name?: string;
  /** 被引用消息的类型 (text/image/file) */
  quoted_content_type?: string;
  /** 被引用消息的文件元数据（图片缩略图/文件图标用） */
  quoted_file_metadata?: FileMetadata;
}

export interface FileMetadata {
  url?: string;
  original_name?: string;
  file_size?: number;
  mime_type?: string;
  width?: number;
  height?: number;
}
