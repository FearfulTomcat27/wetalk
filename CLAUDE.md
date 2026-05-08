# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WeTalk — C/S 模式的在线聊天应用。前端 Next.js 16 (React 19) + Go/gin 后端 + MySQL + Redis。

## Commands

### Frontend (`frontend/`)
```bash
pnpm dev          # 开发服务器 (Turbopack, :3000)
pnpm build        # 生产构建 (TypeScript + Turbopack)
pnpm lint         # ESLint
pnpm add <pkg>    # 安装依赖（必须用 pnpm）
```
shadcn/ui 组件用 `npx shadcn@latest add <component>` 添加。

### Backend (`backend/`)
```bash
go build ./...    # 编译检查
go run ./cmd      # 启动服务 (:8080)
go mod tidy       # 整理依赖
```
格式化：`gofumpt` / `goimports`，lint：`golangci-lint`。工具在 `$HOME/go/bin/`。

**注意：** 配置在 `config.yaml`（不入库，含 DB 密码 + JWT secret），参考 README 示例。

## Architecture

```
wetalk/
├── frontend/                    # Next.js 16 (App Router, Turbopack)
│   └── src/
│       ├── app/
│       │   ├── layout.tsx       # 根布局 (RouteGuard + Toaster)
│       │   ├── page.tsx         # Dashboard 首页
│       │   ├── (auth)/          # /login, /register
│       │   └── (main)/          # /chat, /contacts (需登录)
│       ├── components/
│       │   ├── RouteGuard.tsx   # 全局路由守卫 (唯一跳转逻辑入口)
│       │   ├── Sidebar.tsx      # 左侧窄边栏 (68px, 头像+导航+退出)
│       │   ├── ContactList.tsx  # 联系人列表 (chat/contacts 双 variant)
│       │   ├── ChatArea.tsx     # 消息气泡区域 (微信绿色 #95EC69)
│       │   ├── ChatInput.tsx    # 输入框+底部图标栏 (Enter 发送)
│       │   ├── AddFriendDialog.tsx  # 搜索用户+发送好友请求弹窗
│       │   └── ui/              # shadcn/ui 组件 (button/card/dialog/input...)
│       ├── config/
│       │   └── routes.ts        # 路由权限配置
│       ├── lib/
│       │   ├── api/             # 模块化 API 客户端 (axios)
│       │   │   ├── request.ts   #   实例/拦截器/ApiError
│       │   │   ├── auth.ts      #   login/register/fetchCurrentUser
│       │   │   ├── users.ts     #   searchUsers
│       │   │   ├── messages.ts  #   sendMessage/getMessages/markAsRead
│       │   │   └── friends.ts   #   addFriend/getPendingRequests...
│       │   ├── avatar.ts        # getAvatarSrc(avatar?, seed?) — 优先后端 URL, DiceBear fallback
│       │   ├── validators.ts    # zod schemas (login/register)
│       │   └── ws.ts            # WebSocket 客户端单例 (connect/disconnect/send/on, auth frame, 心跳, 重连)
│       ├── stores/
│       │   ├── auth.ts          # zustand: token/user/_hydrated/_userFetched/init/login/register/logout
│       │   └── chat.ts          # zustand: contacts/messages/activeContactId/inputTexts/connected/sending + WS-first sendMessage + receiveMessage/updateMessageStatus
│       ├── types/chat.ts        # Contact, Message (含 client_msg_id) 类型
│       └── hooks/
├── backend/                     # Go 1.26, gin v1.12
│   ├── cmd/main.go              # 入口：依赖注入 + 启动/关闭
│   ├── config/config.go         # YAML 配置加载 (DB/Redis/JWT)
│   ├── db/                      # MySQL (sql.DB) + Redis (go-redis/v9) 连接池
│   ├── internal/
│   │   ├── router/router.go     # 路由注册（按模块分组，从 main.go 拆出）
│   │   ├── user/                # handler → service → repository 三层
│   │   ├── friend/              # 好友模块 (添加/接受/拒绝/待处理)
│   │   ├── message/             # 消息模块 (发送/查询聊天记录)
│   │   ├── ws/                  # WebSocket (Hub + Client + auth frame 认证)
│   │   └── middleware/auth.go   # JWT Bearer token 解析 → 注入 user_id
│   ├── pkg/
│   │   ├── errors/errors.go     # ErrNotFound/ErrConflict/ErrUnauthorized
│   │   └── utils/utils.go       # 统一响应 Success()/Error()
│   └── scripts/migrations/      # SQL 迁移 (users/friends/messages 表)
```

## Key Conventions

### Next.js 16
- 先读 `node_modules/next/dist/docs/` 指南，API 与训练数据可能不同。
- Turbopack 构建，App Router，Server Components 默认开启。
- Tailwind CSS v4 + `@tailwindcss/postcss` 插件。

### 路由权限 (RouteGuard)
封装在 `components/RouteGuard.tsx`，根 layout 使用。所有跳转逻辑只在此一处：

| 路由类型 | 路径 | 规则 |
|---------|------|------|
| `publicRoutes` | `/login`, `/register` | 已登录 → `/chat` |
| `redirectWhenLoggedInRoutes` | `/` | 已登录 → `/chat` |
| `protectedRoutes` | `/chat`, `/contacts` | 未登录 → `/login` |

各页面组件只管 UI 渲染，不做跳转判断。

### 头像逻辑
- 优先使用后端返回的 `avatar` URL。
- `getAvatarSrc(avatarUrl?, seed?)` — 无 URL 时 DiceBear (`micah?seed=`) fallback。
- 所有 `<img>` 有 `onError` fallback 到首字母占位。

### API 对接
- 前端 axios 实例 `baseURL: http://localhost:8080`（`NEXT_PUBLIC_API_URL` 可覆盖）。
- 请求拦截器自动注入 `Bearer token`。
- 响应拦截器统一 `toast.error(error.message)`，页面不重复 catch 弹 toast。
- 后端响应格式：`{ code, message, data }`。

### Go 后端规范
- 单行 if 必须加大括号。
- Handler → Service → Repository 三层分离，repository 负责 SQL。
- 路由注册集中在 `internal/router/router.go`，main.go 只做依赖组装和启动/关闭。
- 密码字段 `json:"-"`，密码哈希 bcrypt。
- 用户注册自动生成 DiceBear avatar URL。
- SQL 迁移按编号命名，`config.yaml` 不入库。
- WebSocket 认证采用 auth frame 方案（首条消息 `{"type":"auth","token":"<JWT>"}`），不走 JWT middleware。

### 前端状态管理 (zustand)
- `auth store`：token(userId+hydrated+_userFetched)→ init()→ localStorage→ fetchUser→ RouteGuard 等_userFetched
- `chat store`：contacts/messages/activeContactId/connected/sending/inputTexts，sendMessage WS-first + HTTP 降级，乐观 UI (临时消息 + updateMessageStatus)
- `sonner` toast 统一由 axios 拦截器处理

### WebSocket 实时聊天
- 后端 `ws` 包：Hub 模式（连接注册表 + 消息路由），Client（auth frame 认证 + ReadPump/WritePump）
- 前端 `ws.ts`：WSClient 单例，auth frame 认证，25s 心跳 ping，指数退避重连
- 消息协议：扁平 JSON 格式 `{"type":"message.new","id":42,"sender_id":1,...}`
- 路由注册集中在 `internal/router/router.go`，按模块分组（auth/api/friends/messages/ws）
- WS-first 发送：connected 时走 WS，离线时 HTTP POST 降级
- 乐观 UI：发送时生成临时消息（负 id + client_msg_id），收到 message.sent 后替换为服务端真实消息

### API 端点

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/auth/register` | 无 | 注册 (自动生成 DiceBear 头像) |
| POST | `/api/auth/login` | 无 | 登录 |
| GET | `/api/me` | JWT | 当前用户完整信息 |
| GET | `/api/users?keyword=` | JWT | 搜索用户 |
| POST | `/api/friends` | JWT | 发送好友请求 `{friend_id}` |
| GET | `/api/friends` | JWT | 好友列表 |
| GET | `/api/friends/pending` | JWT | 待处理好友请求 |
| PUT | `/api/friends/:id/accept` | JWT | 接受请求 |
| DELETE | `/api/friends/:id` | JWT | 删除/拒绝好友 |
| POST | `/api/messages` | JWT | 发送消息 `{receiver_id, content}` |
| GET | `/api/messages?friend_id=` | JWT | 聊天记录 (双向查询) |
| PUT | `/api/messages/read` | JWT | 标记已读 `{sender_id}` |
| GET | `/ws` | auth frame | WebSocket 连接 (实时消息推送) |
| GET | `/ping` | 无 | 健康检查 |
