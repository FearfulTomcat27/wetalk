# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WeTalk — C/S 模式的在线聊天应用。前端 Next.js 16 (React 19) + Go/gin 后端 + MySQL。

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
go run .          # 启动服务 (:8080)
go mod tidy       # 整理依赖
```
格式化：`gofumpt` / `goimports`，lint：`golangci-lint`。工具在 `$HOME/go/bin/`。

## Architecture

```
wetalk/
├── frontend/          # Next.js 16 (App Router, Turbopack)
│   ├── src/
│   │   ├── app/       # 页面路由
│   │   │   ├── (auth)/login/    # /login
│   │   │   └── (auth)/register/ # /register
│   │   ├── components/ui/   # shadcn/ui 组件
│   │   ├── lib/        # api.ts (API 客户端), validators.ts (zod)
│   │   ├── stores/     # zustand stores
│   │   └── hooks/      # React hooks
│   └── components.json # shadcn/ui 配置 (base=zinc, lucide icons)
├── backend/           # Go 1.26, gin v1.12
│   ├── main.go        # 路由注册入口
│   ├── config/        # YAML 配置解析
│   ├── db/            # MySQL 连接池 (go-sql-driver/mysql)
│   ├── models/        # 数据模型
│   ├── handlers/      # HTTP 处理器
│   ├── middleware/     # JWT 认证中间件
│   ├── migrations/    # SQL 迁移文件
│   └── config.yaml    # 数据库 + JWT 配置（不入库）
```

## Key Conventions

### Next.js 16 注意事项
- **This is NOT standard Next.js** — AGENTS.md 要求先读 `node_modules/next/dist/docs/` 中的指南，API 和约定可能与训练数据不同。
- 使用 Turbopack 构建，App Router，Server Components 默认开启。
- Tailwind CSS v4，使用 `@tailwindcss/postcss` 插件。

### API 对接
- 前端通过 `NEXT_PUBLIC_API_URL` 环境变量或默认 `/api` 调用后端。
- 后端响应格式统一为 `{ code: number, message: string, data?: T }`。
- 认证：JWT Bearer token，登录/注册后存入 localStorage，请求时自动附加。
- 注册/登录端点：`POST /api/auth/register`、`POST /api/auth/login`。受保护端点需 `Authorization: Bearer <token>` 头。

### Go 后端代码规范
- 单行 if 必须加大括号。
- 密码字段用 `json:"-"` 不序列化。
- 配置通过 config.yaml 加载，JWT secret/expire 均可配置。
- SQL 迁移文件按编号命名（`001_xxx.sql`）。

### 前端状态管理
- zustand v5 管理全局状态，token 持久化到 localStorage。
- zod v4 做表单验证，schemas 定义在 `lib/validators.ts`。
- shadcn/ui 使用 `sonner` 做 toast 通知。
