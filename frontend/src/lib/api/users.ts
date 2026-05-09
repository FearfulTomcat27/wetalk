import { request } from "./request";

export interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  avatar?: string;
}

/** 上传头像 */
export function uploadAvatar(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  return request<{ avatar: string }>({
    method: "POST",
    url: "/api/me/avatar",
    data: formData,
    headers: { "Content-Type": "multipart/form-data" },
  });
}

/** 搜索用户 */
export function searchUsers(keyword: string) {
  return request<UserInfo[]>({
    method: "GET",
    url: "/api/users",
    params: { keyword },
  });
}
