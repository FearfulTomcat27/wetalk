/**
 * 路由权限配置。
 *
 * - publicRoutes:              已登录用户不能访问（自动跳转 /chat）
 * - redirectWhenLoggedInRoutes: 已登录时重定向到 /chat（不等于 publicRoutes，因为公开路由无需拦截登录页）
 * - protectedRoutes:           必须登录才能访问（未登录跳转 /login）
 *                              支持前缀匹配，例如 "/chat" 会匹配 "/chat"、"/chat/xxx"
 */

/** 已登录用户不能访问（精确匹配） */
export const publicRoutes = ["/login", "/register"];

/** 已登录时自动跳转到 /chat（精确匹配） */
export const redirectWhenLoggedInRoutes = ["/"];

/** 必须登录才能访问（前缀匹配） */
export const protectedRoutes = ["/chat", "/contacts"];

export function isPublicRoute(pathname: string): boolean {
  return publicRoutes.includes(pathname);
}

export function isRedirectRoute(pathname: string): boolean {
  return redirectWhenLoggedInRoutes.includes(pathname);
}

export function isProtectedRoute(pathname: string): boolean {
  return protectedRoutes.some((route) => pathname.startsWith(route));
}
