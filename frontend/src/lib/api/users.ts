import { request } from "./request";

export interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  avatar?: string;
}

/** 搜索用户 */
export function searchUsers(keyword: string) {
  return request<UserInfo[]>({
    method: "GET",
    url: "/api/users",
    params: { keyword },
  });
}
