# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WeTalk — C/S 模式的在线聊天应用。前端 Next.js 16 (React 19) + Go/gin 后端 + MySQL + MongoDB + Redis + 阿里云 OSS。

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
go build ./...                    # 编译检查
go run ./cmd                      # 启动服务 (:8080)
go run ./cmd/migrate_mongo        # MySQL → MongoDB 消息迁移
go mod tidy                       # 整理依赖

# 测试（需要 Docker Desktop 运行中）
go test ./...                     # 单元测试
go test --tags=integration ./...  # 集成测试（启动 MySQL/Redis/MongoDB 容器）
gotestsum -- -tags=integration -count=1 -timeout 120s ./...  # 集成测试（gotestsum 格式化输出）
```

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
│       │   ├── ChatArea.tsx     # 消息气泡区域（文本/图片缩略图/文件卡片，图片全屏预览，文件详情 Dialog）
│       │   ├── ChatInput.tsx    # 输入框+emoji面板+图片/文件上传 (Enter 发送, Shift+Enter 换行)
│       │   ├── ThemeProvider.tsx # next-themes 主题支持 (浅色/深色)
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
│       │   │   ├── friends.ts   #   addFriend/getPendingRequests...
│       │   │   └── upload.ts    #   uploadFile (multipart/form-data)
│       │   ├── avatar.ts        # getAvatarSrc(avatar?, seed?) — 优先后端 URL, DiceBear fallback
│       │   ├── validators.ts    # zod schemas (login/register)
│       │   └── ws.ts            # WebSocket 客户端单例 (connect/disconnect/send/on, auth frame, 心跳, 重连)
│       ├── stores/
│       │   ├── auth.ts          # zustand: token/user/_hydrated/_userFetched/init/login/register/logout
│       │   └── chat.ts          # zustand: contacts/messages/activeContactId/inputTexts/connected/sending + sendMessage(WS-first+HTTP降级) + sendMediaMessage(图片/文件) + receiveMessage/updateMessageStatus + formatMessagePreview
│       ├── types/chat.ts        # Contact, Message (含 client_msg_id) 类型
│       └── hooks/
├── backend/                     # Go 1.26, gin v1.12
│   ├── cmd/
│   │   ├── main.go              # 入口：依赖注入 + 启动/关闭
│   │   └── migrate_mongo/       # MySQL → MongoDB 数据迁移脚本
│   ├── config/config.go         # YAML 配置加载 (DB/Redis/JWT/Mongo/OSS)
│   ├── db/                      # MySQL (sql.DB) + Redis (go-redis/v9) + MongoDB 连接池
│   ├── controller/              # HTTP Handler 层 (user/friend/message/upload)
│   ├── service/                 # 业务逻辑层 (user/friend/message/chat)
│   ├── dto/                     # 数据访问层 (SQL 查询)
│   ├── model/                   # 数据模型 & 请求/响应结构体
│   ├── common/                  # 共享工具 (Response 辅助函数, OSS Client)
│   ├── types/                   # 类型定义 (AppError 业务错误码, Response 结构体)
│   ├── middleware/auth.go       # JWT Bearer token 解析 → 注入 user_id
│   ├── ws/                      # WebSocket (Hub + Client + auth frame 认证)
│   ├── integration/             # 集成测试（testcontainers, 78 测试用例）
│   └── scripts/migrations/      # SQL 迁移 (users/friends/messages/file_metadata 表)
```

## Code Formatting

代码格式化通过 Claude Code hooks 自动触发（`.claude/settings.json` + `.claude/format-on-edit.sh`）：

- **前端文件** (`frontend/**/*.{ts,tsx,js,jsx,css}`) — 编辑后自动运行 `pnpm lint --fix`
- **后端文件** (`backend/**/*.go`) — 编辑后自动运行 `gofumpt -w` + `goimports -w`

无需手动运行格式化命令。

### Next.js 16

- 先读 `node_modules/next/dist/docs/` 指南，API 与训练数据可能不同。
- Turbopack 构建，App Router，Server Components 默认开启。
- Tailwind CSS v4 + `@tailwindcss/postcss` 插件。

### 路由权限 (RouteGuard)

封装在 `components/RouteGuard.tsx`，根 layout 使用。所有跳转逻辑只在此一处：

| 路由类型                         | 路径                    | 规则             |
|------------------------------|-----------------------|----------------|
| `publicRoutes`               | `/login`, `/register` | 已登录 → `/chat`  |
| `redirectWhenLoggedInRoutes` | `/`                   | 已登录 → `/chat`  |
| `protectedRoutes`            | `/chat`, `/contacts`  | 未登录 → `/login` |

各页面组件只管 UI 渲染，不做跳转判断。

### 头像逻辑

- 优先使用后端返回的 `avatar` URL。
- `getAvatarSrc(avatarUrl?, seed?)` — 无 URL 时 DiceBear (`micah?seed=`) fallback。
- 所有 `<img>` 有 `onError` fallback 到首字母占位。

### 时间戳显示 (ChatArea.formatTime)

消息时间戳按以下优先级格式化：

| 条件         | 格式                | 示例                |
|------------|-------------------|-------------------|
| 昨天         | `昨天 HH:mm`        | 昨天 14:30          |
| 当前周（周一~周日） | `星期X HH:mm`       | 星期一 14:30         |
| 同年非当前周     | `M月D日 HH:mm`      | 5月10日 14:30       |
| 跨年         | `YYYY年M月D日 HH:mm` | 2025年12月28日 14:30 |

### 文件上传 (OSS)

- 前端 `ChatInput` 支持图片/文件上传按钮，图片 ≤10MB，文件 ≤20MB。
- 后端 `/api/upload` 接收 multipart 请求，上传至阿里云 OSS，key 格式 `uploads/{type}/{user_id}/{timestamp}_{filename}`。
- 上传成功后返回 `{url, content_type, file_name, file_size, file_type}`。
- 文件元数据嵌入在 MongoDB 消息文档的 `file_metadata` 字段中（MySQL `file_metadata` 表已废弃），无需额外 JOIN 查询。

### API 对接

- 前端 axios 实例 `baseURL: http://localhost:8080`（`NEXT_PUBLIC_API_URL` 可覆盖）。
- 请求拦截器自动注入 `Bearer token`。
- 响应拦截器统一 `toast.error(error.message)`，页面不重复 catch 弹 toast。
- 后端响应格式：`{ code, message, data }`。

### Go 后端规范

- 单行 if 必须加大括号。
- Controller → Service → DTO 三层分离，dto 负责数据访问（MySQL 用 GORM，MongoDB 用 mongo-go-driver）。
- 路由注册集中在 `router/router.go`，main.go 只做依赖组装和启动/关闭。
- 密码字段 `json:"-"`，密码哈希 bcrypt。
- 用户注册自动生成 DiceBear avatar URL。
- SQL 迁移按编号命名，`config.yaml` 不入库。
- WebSocket 认证采用 auth frame 方案（首条消息 `{"type":"auth","token":"<JWT>"}`），不走 JWT middleware。
- 业务错误使用 `types.AppError`（含 Code/Message/HTTPStatus），controller 通过 `errors.As` + `common.AppError` 统一响应。
- **新增 controller 模块时**：
  1. 在 handler 方法上添加 swagger 注释（参考 `controller/user.go` 的 `@Summary`/`@Description`/`@Tags`/`@Param`/`@Success`/`@Failure`/`@Router`），然后运行 `swag init -g cmd/main.go -o docs/` 重新生成文档。
  2. 在 `cmd/providers.go` 添加 provider 函数，在 `cmd/wire.go` 的 `wire.Build` 中注册，在 `router/router.go` 的 `Dependencies` 结构体中添加字段。
- **新增 service 模块时**：如需对接 DTO 层，在 `service/interfaces.go` 中定义 Repository 接口，在 `cmd/providers.go` 添加对应的 provider 实现绑定。

### 前端状态管理 (zustand)

- `auth store`：token(userId+hydrated+_userFetched)→ init()→ localStorage→ fetchUser→ RouteGuard 等_userFetched
- `chat store`：contacts/messages/activeContactId/connected/sending/inputTexts，sendMessage WS-first + HTTP 降级，乐观 UI (
  临时消息 + updateMessageStatus)
- `sonner` toast 统一由 axios 拦截器处理

### WebSocket 实时聊天

- 后端 `ws` 包：Hub 模式（连接注册表 + 消息路由），Client（auth frame 认证 + ReadPump/WritePump）
- 前端 `ws.ts`：WSClient 单例，auth frame 认证，25s 心跳 ping，指数退避重连
- 消息协议：扁平 JSON 格式，文本消息 `{"type":"message.new","id":42,"sender_id":1,...}`，图片/文件消息额外携带
  `content_type` 和 `file_metadata`
- 路由注册集中在 `internal/router/router.go`，按模块分组（auth/api/friends/messages/ws/upload）
- WS-first 发送：connected 时走 WS，离线时 HTTP POST 降级
- 乐观 UI：发送时生成临时消息（负 id + client_msg_id），收到 message.sent 后替换为服务端真实消息

### MongoDB 消息存储

- 消息数据存储在 MongoDB `messages` 集合中，`file_metadata` 直接嵌入消息文档（`file_metadata` 字段）。
- 引用消息内容在写入时预填充（`quoted_content` 字段），避免读取时的额外查询。
- `msg_id` 自增通过 MongoDB `counters` 集合实现（`FindOneAndUpdate` + `$inc`）。
- 索引：
  - `(chat_id, created_at)` — 聊天记录分页查询（倒序）
  - `(chat_id, sender_id, status)` — 未读统计和标记已读
  - `(msg_id)` — 唯一索引，兼容旧版 int64 ID 查询
- 启动时自动创建索引，创建失败不阻塞启动（仅日志警告）。

### MySQL 消息表清理

- 消息和 file_metadata 数据从 MySQL 迁移至 MongoDB 后，`messages` 和 `file_metadata` 表可删除。
- 迁移脚本：`go run ./cmd/migrate_mongo`
- DDL 清理脚本：`backend/sql/migrations/003_drop_mysql_messages.sql`

### API 端点

| 方法     | 路径                         | 认证         | 说明                                                        |
|--------|----------------------------|------------|-----------------------------------------------------------|
| POST   | `/api/auth/register`       | 无          | 注册 (自动生成 DiceBear 头像)                                     |
| POST   | `/api/auth/login`          | 无          | 登录                                                        |
| GET    | `/api/me`                  | JWT        | 当前用户完整信息                                                  |
| GET    | `/api/users?keyword=`      | JWT        | 搜索用户                                                      |
| POST   | `/api/friends`             | JWT        | 发送好友请求 `{friend_id}`                                      |
| GET    | `/api/friends`             | JWT        | 好友列表                                                      |
| GET    | `/api/friends/pending`     | JWT        | 待处理好友请求                                                   |
| PUT    | `/api/friends/:id/accept`  | JWT        | 接受请求                                                      |
| DELETE | `/api/friends/:id`         | JWT        | 删除/拒绝好友                                                   |
| POST   | `/api/messages`            | JWT        | 发送消息 `{chat_id, content, content_type?, quoted_message_id?}` |
| GET    | `/api/messages?chat_id=&offset=&limit=` | JWT        | 聊天记录 (MongoDB 分页, 按时间倒序)                        |
| GET    | `/api/messages/unread`     | JWT        | 所有聊天的未读消息数 `[{chat_id, count}]`                        |
| PUT    | `/api/messages/read`       | JWT        | 标记已读 `{chat_id}`                                          |
| POST   | `/api/upload`              | JWT        | 上传文件/图片 (multipart, `type=image\|file`, 图片≤10MB, 文件≤20MB) |
| GET    | `/ws`                      | auth frame | WebSocket 连接 (实时消息推送)                                     |
| GET    | `/ping`                    | 无          | 健康检查                                                      |
