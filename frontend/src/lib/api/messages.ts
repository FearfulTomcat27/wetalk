import { request } from "./request";
import type { Message } from "@/types/chat";

export interface SendMessageRequest {
  chat_id: number;
  content: string;
  content_type?: string;
  client_msg_id?: string;
  file_metadata?: {
    url: string;
    original_name: string;
    file_size: number;
    mime_type: string;
    width?: number;
    height?: number;
  };
}

/** 发送消息 */
export function sendMessage(data: SendMessageRequest) {
  return request<Message>({
    method: "POST",
    url: "/api/messages",
    data,
  });
}

/** 标记某个聊天的消息为已读 */
export function markAsRead(chatId: number) {
  return request({
    method: "PUT",
    url: "/api/messages/read",
    data: { chat_id: chatId },
  });
}

/** 获取某个聊天的历史消息 */
export function getMessages(
  chatId: number,
  before?: number,
  limit?: number,
) {
  return request<Message[]>({
    method: "GET",
    url: "/api/messages",
    params: { chat_id: chatId, before, limit },
  });
}
