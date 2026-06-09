# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 开发命令

```bash
# 前端 (端口 10050, 代理 /api → localhost:10040)
cd frontend && npm run dev        # 启动开发服务器
cd frontend && npm run build      # 类型检查 + 生产构建 (tsc -b && vite build)
cd frontend && npx tsc --noEmit   # 仅类型检查

# 后端 (开发端口 10040, 默认 8080)
cd backend && go run ./cmd/server   # 启动服务器

# Docker (生产部署)
docker compose up -d --build        # 构建并启动后端容器
```

**注意**: 本项目没有测试脚本（package.json 无 test 命令，Go 无 _test.go 文件），也没有 lint 配置。

## 配置

后端通过 Viper 加载配置，优先级：环境变量 > `.env` 文件 > 默认值。

| 配置文件 | 用途 |
|----------|------|
| `backend/.env` | 本地开发（后端直连外部 DB/Redis/MinIO） |
| `.env.production` | Docker 生产部署 |
| `frontend/.env` | 前端环境变量（`VITE_ONLYOFFICE_DS_URL`） |
| `.env.example` | 配置模板（注意：key 名与 config.go 不完全一致，仅供参考） |

**关键配置项**:
- `SERVER_PORT` — 后端端口（默认 8080，开发环境设为 10040）
- `STORAGE_DRIVER` — `local` 或 `minio`（MinIO S3 兼容存储）
- `REDIS_ADDR` — Redis 地址，不设置则跳过（权限缓存未实现，当前可忽略）
- `ONLYOFFICE_DS_URL` / `ONLYOFFICE_JWT_SECRET` / `ONLYOFFICE_CALLBACK_URL` — 不设置则禁用在线编辑
- `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` — 至少 32 字符

## 架构概览

**前端**: React 18 + Vite + TypeScript + Ant Design 5 + Zustand + React Query + React Router 6
**后端**: Go + Gin + pgx (PostgreSQL) + Viper (配置) — 标准三层架构 (handler → service → repository)
**基础设施**: PostgreSQL 16 (ltree/pg_trgm 扩展), Redis 7 (可选), MinIO (可选), OnlyOffice Document Server (可选)

数据库和 Redis 是外部服务，不在 docker-compose.yml 中管理。docker-compose 仅包含后端容器。

### 前端分层

```
pages/          → 页面组件 (路由级别)
components/     → 可复用 UI 组件
api/            → API 客户端函数 (每个 API 模块一个文件)
store/          → Zustand 状态 (authStore, teamStore)
hooks/          → 自定义 hooks
types/          → TypeScript 类型定义
```

- **路由体系**:
  - `/login`, `/register` — 公开
  - `/` — 首页（个人 + 团队双空间入口）
  - `/my/files`, `/my/files/:folderId`, `/my/trash` — 个人空间
  - `/teams` — 团队管理（列表 + 加入）
  - `/teams/:id` — 团队详情 + 成员
  - `/teams/:teamId/files`, `/teams/:teamId/files/:folderId`, `/teams/:teamId/trash` — 团队空间
  - `/editor/:fileId` — 编辑器（独立页面，无 AppLayout）
  - `/profile` — 个人设置
- **API 客户端** (`api/client.ts`): axios 实例，自动附加 Bearer token，401 时静默刷新 token 并重试队列（不拦截 auth 端点自身避免死循环）
- **状态管理**: `authStore` (zustand) 存储 token + user，持久化在内存; 其余状态由 React Query 管理服务端缓存
- **布局**: `AppLayout` 使用 antd `Layout` — 左侧白色 Sider（分区菜单）+ 顶部 Header + 灰色背景 Content（腾讯文档风格）
- **空间模型**: 个人空间（owner_id 隔离）+ 团队空间（team_id 隔离），FileBrowserPage / TrashPage 通过 URL 自动检测 scope
- **侧边栏分区**: 首页 / 个人空间（我的文件、回收站）/ 团队空间（动态列出用户所在团队 + 管理入口）/ 个人设置
- **路径别名**: `@/` 映射到 `frontend/src/`（vite resolve.alias）

### 后端分层

```
cmd/server/main.go       → 入口: 连接 DB, 运行迁移, 组装依赖, 优雅关闭
config/config.go         → Viper 配置加载 (.env + 环境变量), DSN 构建
internal/
  router/router.go       → Gin 路由定义 (所有 API 端点一览)
  handler/*_handler.go   → HTTP 层 (参数解析, 响应)
  service/*_service.go   → 业务逻辑
  repository/*_repo.go   → 数据访问 (SQL via pgx)
  model/models.go        → 数据模型
  middleware/             → JWT认证 (auth.go), CORS, 限流, 错误恢复
  storage/                → Storage 接口 + LocalStorage + MinioStorage
  dto/                    → 请求/响应 DTO
  util/                   → JWT, bcrypt, DB 连接池, 响应格式化
```

- **数据库迁移**: 内联在 `cmd/server/main.go` 的 `runMigrations()` 中，启动时自动执行（CREATE EXTENSION + CREATE TABLE IF NOT EXISTS），无独立迁移工具
- **存储抽象**: `storage.Storage` 接口（Store/Retrieve/Delete/Exists），支持 `local`（本地文件系统）和 `minio`（S3 兼容）两种驱动
- **路由分组**: `/api/v1` 下分为 auth（无需认证）、oo（OnlyOffice 回调，JWT in query）、protected（需 Bearer token）三组
- **优雅关闭**: 监听 SIGINT/SIGTERM，5 秒超时关闭 HTTP server

### 关键业务实体

- **Folder**: ltree 路径 (`folder_path`), 软删除, 权限控制
- **File**: 多版本 (FileVersion), 按扩展名分类图标, 软删除
- **Team**: 团队协作, 含成员角色管理, 支持加入申请/审批
- **Permission**: 文件夹/文件级别, 可按用户或团队授权
- **OnlyOffice**: JWT 加密回调, Document Server 集成编辑 .docx/.xlsx/.pptx

### 部署架构

- **开发**: Vite dev server (10050) 代理 `/api` 到 `localhost:10040`（Go 后端直连）
- **Docker 生产**: Nginx (80) 反向代理 `/api/` 到 backend 容器 (8080)，前端静态文件由 Nginx 直接 serve
- **裸机生产**: `nginx.production.conf` 示例 — 前端静态文件 + 代理 `/api/` 到 `127.0.0.1:10040`

## 界面美化指南

本项目参考**腾讯文档**设计风格，使用 **Ant Design 5** 组件库，所有样式通过 **内联 style 对象** 实现，无 CSS/Less 文件。

### 设计 Token 管理

统一设计 token 定义在 `frontend/src/theme.ts`，包含：
- `themeToken` — antd ConfigProvider 全局主题（主色 `#0052D9`、圆角 `8px`、中文字体栈）
- `colors` — 品牌色、文字色、边框色、背景色常量
- `fileIconColors` / `fileCardIconColors` — 文件类型图标颜色映射
- `spacing` — 间距体系 (xs:4, sm:8, md:12, lg:16, xl:24, xxl:32)
- `sidebar` — 侧边栏宽度配置

所有组件应从 `theme.ts` 引用颜色，不硬编码色值。

### 设计要点

- **主色**: `#0052D9`（腾讯蓝），通过 `ConfigProvider theme.token.colorPrimary` 全局注入
- **侧边栏**: 白色背景 + 右侧细线分隔，非深色背景。`theme="light"`，不使用 dark。
- **页头**: 白色背景 + 底部分隔线，无 boxShadow，高度 56px
- **内容区**: 灰色页面背景 (`#f5f6f7`)，白色圆角卡片 (`borderRadius: 12`) 承载内容
- **配色体系**:
  - 文字: 主 `#1d1d1f`、次要 `#88888a`、辅助 `#b0b0b2`
  - 边框: `#e5e6eb`
  - 背景: 页面 `#f5f6f7`、hover `#f0f2f5`
  - 文件图标: PDF `#e34d59`, Excel `#00a870`, PPT `#ed7b2f`, 图片 `#0052D9`, 文件夹 `#f5a623`, 默认 `#88888a`
- **图标**: 全部大图标使用圆形浅底容器包裹（`borderRadius: 10-16`, `background: #f5f6f7`），统一 44-64px 尺寸
- **间距**: 卡片内 padding 16-20px，内容区 margin 20px/padding 24px，列表项 gap 12px
- **登录/注册**: 蓝色系渐变背景 (`#e8f0fe → #d4e4fc → #c5d8f8`)，float 卡片圆角 16px，品牌 logo 渐变图标
- **交互**: hover 时浅蓝底色或蓝色边框 + 微微阴影 `boxShadow: '0 4px 16px rgba(0,82,217,0.1)'`
- **反馈**: react-hot-toast (全局) + antd message (操作确认)
- **空状态**: 虚线边框占位卡片 + 大图标 + 引导文案

### 添加 UI 新功能时的检查清单

1. 使用现有 antd 组件，不引入新的 UI 库
2. 内联 style 或引用 `theme.ts` 中的常量，不创建 CSS 文件
3. 所有文案使用中文
4. 响应式考虑：优先 antd Row/Col 栅格 (xs/sm/md/lg/xl)
5. 操作反馈：成功用 `message.success()`，失败用 `message.error()`
6. 图标选择 antd icons 已有集合，不引入图标库
7. 交互元素必须有 hover 效果（边框变色 + 微阴影）
8. 文件/文件夹使用圆形浅底容器包裹图标
