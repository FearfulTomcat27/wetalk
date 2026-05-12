import { request } from "./request";
import type { Message } from "@/types/chat";

export interface SendMessageRequest {
  chat_id: number;
  content: string;
  content_type?: string;
  client_msg_id?: string;
  quoted_message_id?: number;
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

/** 未读消息数（每个聊天一个） */
export interface UnreadCount {
  chat_id: number;
  count: number;
}

/** 获取当前用户所有聊天的未读消息数 */
export function getUnreadCounts() {
  return request<UnreadCount[]>({
    method: "GET",
    url: "/api/messages/unread",
  });
}

/** 删除消息 */
export function deleteMessage(messageId: number) {
  return request({
    method: "DELETE",
    url: `/api/messages/${messageId}`,
  });
}

/** 删除某个聊天的全部聊天记录（仅对当前用户生效） */
export function deleteChatHistory(chatId: number) {
  return request({
    method: "DELETE",
    url: `/api/messages/history/${chatId}`,
  });
}
