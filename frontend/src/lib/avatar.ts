/**
 * 获取头像图片地址。
 * 优先使用后端返回的 avatar URL，无则 fallback 到 DiceBear 生成的头像。
 *
 * @param avatarUrl 后端返回的头像 URL（可为空）
 * @param seed 用户名，用作 DiceBear 随机种子
 * @returns 头像图片 URL
 */
export function getAvatarSrc(avatarUrl?: string, seed?: string): string {
  if (avatarUrl) {
    return avatarUrl;
  }
  const s = seed || "default";
  return `https://api.dicebear.com/9.x/micah/svg?seed=${encodeURIComponent(s)}`;
}
