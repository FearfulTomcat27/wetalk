import { request } from "./request";
import type { UserInfo } from "./users";

/** 好友信息 — 匹配后端 { friend_id, friend_name, friend_avatar, created_at, last_message, unread_count } */
export interface FriendInfo {
  friend_id: number;
  friend_name: string;
  friend_avatar?: string;
  created_at?: string;
  last_message?: string;
  unread_count?: number;
}

/** 待处理的好友请求 — 匹配后端 { id, user: {...}, createdAt } */
export interface PendingRequest {
  /** 请求 ID */
  id: number;
  /** 请求发送时间 */
  createdAt?: string;
  /** 发起请求的用户信息 */
  user: UserInfo;
}

/** 发送好友请求 */
export function addFriend(friendId: number) {
  return request({ method: "POST", url: "/api/friends", data: { friend_id: friendId } });
}

/** 获取好友列表 */
export function getFriends() {
  return request<FriendInfo[]>({ method: "GET", url: "/api/friends" });
}

/** 获取待处理的好友请求 */
export function getPendingRequests() {
  return request<PendingRequest[]>({ method: "GET", url: "/api/friends/pending" });
}

/** 接受好友请求 */
export function acceptFriendRequest(friendId: number) {
  return request({ method: "PUT", url: `/api/friends/${friendId}/accept` });
}

/** 拒绝/删除好友请求 */
export function declineFriendRequest(friendId: number) {
  return request({ method: "DELETE", url: `/api/friends/${friendId}` });
}
