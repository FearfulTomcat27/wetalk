# WeTalk

在线聊天应用，C/S 架构。前端 Next.js 16 (React 19) + Go/gin 后端 + MySQL + Redis。

## 项目结构

```
wetalk/
├── frontend/                 # Next.js 16 前端 (App Router, Turbopack)
│   └── src/
│       ├── app/              # 页面路由 (/, /login, /register, /chat)
│       ├── components/       # UI 组件 (shadcn/ui)
│       ├── stores/           # zustand 状态管理
│       ├── lib/              # API 客户端 (axios), 验证 (zod)
│       └── hooks/            # React hooks
├── backend/                  # Go 1.26 后端
│   ├── cmd/                  # 入口 main.go
│   ├── internal/             # 业务模块 (user, message, friend, middleware)
│   ├── config/               # 配置加载
│   ├── db/                   # MySQL + Redis 连接
│   ├── pkg/                  # 公共包 (errors, utils)
│   └── scripts/migrations/   # SQL 迁移
└── CLAUDE.md                 # AI 助手指南
```

## 快速开始

### 前端

```bash
cd frontend
pnpm install
pnpm dev          # http://localhost:3000
pnpm build        # 生产构建
```

环境变量：`NEXT_PUBLIC_API_URL`（后端地址，默认 `http://localhost:8080`）

### 后端

```bash
cd backend
# 1. 创建 config.yaml（参考下方配置）
# 2. 确保 MySQL 和 Redis 已启动
# 3. 执行 scripts/migrations/ 中的 SQL 迁移
go run ./cmd     # http://localhost:8080
```

**config.yaml 示例：**

```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: ""
  dbname: wetalk
  charset: utf8mb4
  max_open_conns: 25
  max_idle_conns: 10

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10

jwt:
  secret: "your-jwt-secret"
  expire_hours: 72
```

### 格式化 & Lint

```bash
# 前端
cd frontend && pnpm lint

# 后端（工具在 $HOME/go/bin/）
gofumpt -w . && goimports -w .
golangci-lint run
```

## 技术栈

| 层 | 技术 |
|----|------|
| 前端框架 | Next.js 16 (React 19, TypeScript) |
| UI 组件 | shadcn/ui + Tailwind CSS v4 |
| 状态管理 | zustand |
| 表单验证 | zod |
| HTTP 客户端 | axios |
| 后端框架 | gin (Go) |
| 数据库 | MySQL |
| 缓存 | Redis |
| 认证 | JWT + bcrypt |

## API 端点

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/auth/register` | 无 | 注册 |
| POST | `/api/auth/login` | 无 | 登录 |
| GET | `/api/me` | JWT | 当前用户 |
| GET | `/api/users?keyword=` | JWT | 搜索用户 |
| POST | `/api/friends` | JWT | 添加好友 |
| GET | `/api/friends` | JWT | 好友列表 |
| PUT | `/api/friends/:id/accept` | JWT | 接受请求 |
| DELETE | `/api/friends/:id` | JWT | 删除好友 |
| POST | `/api/messages` | JWT | 发送消息 |
| GET | `/api/messages?friend_id=` | JWT | 聊天记录 |

## License

MIT