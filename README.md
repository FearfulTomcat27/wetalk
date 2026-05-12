# WeTalk

在线聊天应用，C/S 架构。前端 Next.js 16 (React 19) + Go/gin 后端 + MySQL + MongoDB + Redis。

## 功能

- 用户注册/登录（JWT + bcrypt）
- Dashboard 首页
- 微信风格三列聊天面板（联系人列表 | 消息气泡 | 输入框）
- WebSocket 实时消息推送（auth frame 认证，心跳保活，断线重连）
- 乐观 UI 发送（临时消息 + 服务端确认替换）
- 离线降级（WS 断开时自动切换 HTTP POST）
- 图片/文件消息（缩略图预览、全屏查看、文件卡片、类型图标、一键下载）
- Emoji 面板（60 个常用 emoji，点击外部关闭）
- 消息时间戳智能格式化（昨天/星期几/月日/年月日）
- 暗色模式主题切换（next-themes）
- 消息输入框（Enter 发送 / Shift+Enter 换行）
- 联系人页面，好友添加/搜索/申请/同意/拒绝
- 好友请求待处理通知
- 全局路由守卫（登录/未登录自动跳转）
- DiceBear 头像自动生成
- 用户头像上传（阿里云 OSS 存储，支持 jpg/png/gif/webp，≤2MB）
- 联系人列表与聊天区域支持拖拽调整宽度

## 项目结构

```
wetalk/
├── frontend/                    # Next.js 16 前端 (App Router, Turbopack)
│   └── src/
│       ├── app/                 # 页面路由 (/, /login, /register, /chat, /contacts)
│       ├── components/          # 组件 (RouteGuard, Sidebar, ContactList, ChatArea...)
│       │   └── ui/              # shadcn/ui 组件
│       ├── stores/              # zustand 状态 (auth, chat)
│       ├── lib/api/             # 模块化 API 客户端 (axios, 含 upload)
│       └── config/              # 路由权限配置
├── backend/                     # Go 1.26 后端
│   ├── cmd/
│   │   ├── main.go              # 入口
│   │   └── migrate_mongo/       # MySQL → MongoDB 数据迁移
│   ├── config/                  # 配置加载 (DB/Redis/JWT/Mongo/OSS)
│   ├── db/                      # MySQL + Redis + MongoDB 连接
│   ├── controller/              # HTTP Handler 层
│   ├── service/                 # 业务逻辑层
│   ├── dto/                     # 数据访问层
│   ├── model/                   # 数据模型
│   ├── common/                  # 共享工具
│   ├── types/                   # 类型定义 (AppError 业务错误码)
│   ├── middleware/              # JWT 认证中间件
│   ├── ws/                      # WebSocket
│   ├── integration/             # 集成测试（testcontainers, 78 测试用例）
│   └── scripts/migrations/      # SQL 脚本 (建表)
└── CLAUDE.md                    # AI 助手指南
```

## 快速开始

### 前端

```bash
cd frontend
pnpm install
pnpm dev          # http://localhost:3000
```

环境变量：`NEXT_PUBLIC_API_URL`（后端地址，默认 `http://localhost:8080`），`NEXT_PUBLIC_WS_URL`（WebSocket 地址，默认 `ws://localhost:8080/ws`）

### 后端

```bash
cd backend
cp config.example.yaml config.yaml   # 编辑数据库和 JWT 配置
go run ./cmd                          # http://localhost:8080
```

**前置依赖：** MySQL + Redis + MongoDB 已启动，执行 `scripts/migrations/` SQL 脚本建表。

消息数据迁移（MySQL → MongoDB）：

```bash
cd backend
go run ./cmd/migrate_mongo   # 导入现有消息到 MongoDB
```

迁移完成后可选删除 MySQL 消息表：执行 `sql/migrations/003_drop_mysql_messages.sql`。

## 测试

需要 Docker Desktop 运行中（testcontainers 自动启动 MySQL/Redis/MongoDB 容器）。

```bash
cd backend

# 单元测试
go test ./...

# 集成测试
go test --tags=integration -count=1 -timeout 120s ./integration/...

# gotestsum 格式化输出
gotestsum -- -tags=integration -count=1 -timeout 120s ./integration/...
```

现有 78 个集成测试用例，覆盖用户注册/登录、好友管理、消息收发、未读统计等功能。

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

oss:
  endpoint: "oss-cn-shanghai.aliyuncs.com"
  bucket: "your-bucket-name"
  region: "cn-shanghai"
  access_key_id: "your-access-key-id"
  access_key_secret: "your-access-key-secret"

mongo:
  uri: "mongodb://localhost:27017"
  database: "wetalk"
```

**注意：** `config.yaml` 不入库，需手动创建。
```

## 技术栈

| 层 | 技术 |
|----|------|
| 前端框架 | Next.js 16 (React 19, TypeScript) |
| UI 组件 | shadcn/ui + Tailwind CSS v4 |
| 状态管理 | zustand v5 |
| 表单验证 | zod v4 |
| HTTP 客户端 | axios |
| 后端框架 | gin v1.12 |
| 数据库 | MySQL + MongoDB (mongo-go-driver) |
| 缓存 | Redis (go-redis/v9) |
| 认证 | JWT + bcrypt |
| 实时通信 | WebSocket (gorilla/websocket, auth frame) |
| 文件存储 | 阿里云 OSS |
| 主题 | next-themes (浅色/暗色) |
| 头像 | DiceBear (micah) + 阿里云 OSS |

## API 端点

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/auth/register` | 无 | 注册（自动生成 DiceBear 头像） |
| POST | `/api/auth/login` | 无 | 登录 |
| GET | `/api/me` | JWT | 当前用户完整信息 |
| POST | `/api/me/avatar` | JWT | 上传头像（multipart/form-data，≤2MB） |
| GET | `/api/users?keyword=` | JWT | 搜索用户 |
| POST | `/api/friends` | JWT | 发送好友请求 |
| GET | `/api/friends` | JWT | 好友列表 |
| GET | `/api/friends/pending` | JWT | 待处理好友请求 |
| PUT | `/api/friends/:id/accept` | JWT | 接受好友请求 |
| DELETE | `/api/friends/:id` | JWT | 删除/拒绝好友 |
| POST | `/api/messages` | JWT | 发送消息 `{chat_id, content, content_type?, quoted_message_id?}` |
| GET | `/api/messages?chat_id=&offset=&limit=` | JWT | 聊天记录（MongoDB 分页） |
| GET | `/api/messages/unread` | JWT | 所有聊天的未读消息数 |
| PUT | `/api/messages/read` | JWT | 标记已读 `{chat_id}` |
| POST | `/api/upload` | JWT | 上传文件/图片（≤10MB 图片, ≤20MB 文件） |
| GET | `/ws` | auth frame | WebSocket 连接（实时推送） |
| GET | `/ping` | 无 | 健康检查 |

## License

MIT
