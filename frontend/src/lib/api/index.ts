// 基础
export { client, request, ApiError } from "./request";
export type { ApiResponse } from "./request";

// 认证
export { login, register, fetchCurrentUser } from "./auth";
export type { LoginRequest, RegisterRequest, AuthResponse } from "./auth";

// 用户
export { searchUsers, uploadAvatar } from "./users";
export type { UserInfo } from "./users";

// 好友
export {
  addFriend,
  getFriends,
  getPendingRequests,
  acceptFriendRequest,
  declineFriendRequest,
} from "./friends";
export type { FriendInfo, PendingRequest } from "./friends";

// 上传
export { uploadFile } from "./upload";
export type { UploadResult } from "./upload";

// 消息
export { sendMessage, getMessages, markAsRead, getUnreadCounts, deleteChatHistory } from "./messages";
export type { SendMessageRequest, UnreadCount } from "./messages";
