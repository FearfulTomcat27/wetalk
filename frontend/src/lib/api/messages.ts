import { request } from "./request";
import type { Message } from "@/types/chat";

export interface SendMessageRequest {
  contactId: number;
  content: string;
}

/** 发送消息 */
export function sendMessage(data: SendMessageRequest) {
  return request<Message>({
    method: "POST",
    url: "/api/messages",
    data,
  });
}

/** 获取与某联系人的历史消息 */
export function getMessages(contactId: number) {
  return request<Message[]>({
    method: "GET",
    url: `/api/messages/${contactId}`,
  });
}
