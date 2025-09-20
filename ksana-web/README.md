# ksana-web

基于 Vue 3 + Vite + Element Plus 的 ksana-service 可视化 Web 控制台（MVP v0.1）。用于通过 REST API 管理定时任务：创建、查询、更新、删除、启停、立即执行，以及基础健康检查与系统设置。

配套后端：ksana-service（默认运行在 `http://localhost:7100`）。

## 功能特性

- 任务列表：展示 name、enabled、type、调度摘要、next_run_at、last_status、last_run_at，并支持名称关键字、启用状态、类型筛选
- 任务创建/编辑：
  - HTTP 配置（方法、URL、Headers、Body）
  - 调度配置（once 指定 UTC 时间；every 指定 Go Duration，支持 start_at/jitter）
  - 执行控制（timeout、max_retries、retry_backoff）
- 任务详情：只读信息 + 快捷操作（run-now、pause/resume、delete） + JSON 视图
- 健康检查：调用 `/health` 展示服务状态
- 系统设置：本地保存 API 地址、语言等（localStorage 持久化）

## 目录结构

```
ksana-web/
  ├─ index.html
  ├─ vite.config.ts
  ├─ package.json
  └─ src/
     ├─ main.ts
     ├─ router/
     │  └─ index.ts
     ├─ stores/
     │  ├─ jobs.ts
     │  └─ settings.ts
     ├─ api/
     │  ├─ http.ts
     │  └─ jobs.ts
     ├─ types/
     │  └─ job.ts
     ├─ views/
     │  ├─ JobsList.vue
     │  ├─ JobForm.vue
     │  ├─ JobDetail.vue
     │  ├─ Health.vue
     │  └─ Settings.vue
     ├─ components/
     │  ├─ ScheduleEditor.vue
     │  └─ HttpConfigEditor.vue
     └─ utils/
         ├─ schedule.ts
         └─ time.ts
```

## 环境要求

- Node.js：遵循 `package.json` engines（建议 Node 20.19+ 或 22.12+）
- 包管理：npm（或兼容工具）

## 快速开始

1) 安装依赖

```bash
npm install
```

2) 配置后端 API 地址（二选一）

- 方式 A：在项目根目录创建 `.env.local`，设置变量：

```
VITE_API_BASE_URL=http://localhost:7100
```

- 方式 B：运行后进入“设置”页面，直接在 UI 中修改 API 地址，保存后会写入浏览器 localStorage。

3) 本地开发

```bash
npm run dev
```

访问开发地址（Vite 输出的本地 URL）。首次进入会跳转到“任务列表”。

4) 生产构建与预览

```bash
npm run build
npm run preview
```

构建产物为纯静态资源，可直接部署到 Nginx/静态托管服务。

## 主要脚本

- `dev`：本地开发（Vite）
- `build`：类型检查 + 构建
- `preview`：预览构建产物
- `type-check`：使用 `vue-tsc` 进行类型检查
- `lint`：ESLint 校验与自动修复
- `format`：Prettier 针对 `src/` 的格式化

## 环境变量

- `VITE_API_BASE_URL`：后端 API 根地址，默认 `http://localhost:7100`

说明：应用启动时会读取该变量作为默认值，运行期也会读取并使用“设置”页面保存到 localStorage 的 `apiBase`。

## 与后端 API 的约定

统一基地址：`apiBase`（由 `VITE_API_BASE_URL` 或设置页决定）。

使用的接口（需与 ksana-service 对齐）：

- `GET /jobs`、`POST /jobs`
- `GET /jobs/{id}`、`PATCH /jobs/{id}`、`DELETE /jobs/{id}`
- `POST /jobs/{id}/run-now`、`POST /jobs/{id}/pause`、`POST /jobs/{id}/resume`
- `GET /health`

请求/响应：

- 请求默认 `Content-Type: application/json`
- 失败时抛出统一 `ApiError`，组件捕获后通过 Element Plus 的 `ElMessage` 友好提示

时间与调度：

- 所有时间以 UTC ISO 字符串（RFC3339）展示与提交
- 周期配置使用 Go Duration 字符串：示例 `5s`、`1m`、`1h`、`30m`、`2h30m`

## 使用指引（MVP）

1) 健康检查：
   - 打开“健康检查”页确认后端可达，状态应为 `ok`

2) 创建任务：
   - “任务列表” → “新建任务”
   - 填写“HTTP 配置”：方法、URL（需 http/https）、可选 Headers/Body
   - 选择“调度类型”：
     - once：填写 UTC 时间（如 `2025-09-20T03:00:00Z`）
     - every：填写周期（如 `5m`），可选 `start_at`（UTC）和 `jitter`
   - 配置“执行控制”：`timeout`（默认 10s）、`max_retries`（默认 3）、`retry_backoff`（默认 5s）

3) 列表与详情：
   - 列表支持筛选、快捷操作；详情页提供 JSON 视图与更多操作

4) 启停与立即执行：
   - 列表或详情页使用“更多”菜单：run-now、pause/resume、delete

## 常见问题

- 无法连接后端：
  - 确认 ksana-service 已启动并监听 `7100`
  - 确认“设置”页或 `.env.local` 中的 `VITE_API_BASE_URL` 正确
  - 若跨域，建议通过反向代理统一域名或在后端开启 CORS

- 时间格式不通过校验：
  - once/start_at 必须是 RFC3339 UTC 字符串（例如 `2025-09-20T03:00:00Z`）
  - every/jitter/timeout/retry_backoff 必须是合法 Go Duration（例如 `5s`、`1m30s`）

## 技术栈

- Vue 3（`<script setup>`）
- Vite 7
- Element Plus
- Pinia
- Vue Router 4
- TypeScript

## 备注

- 更详细的交互与类型定义可参考 `DESIGN.md`
- 当前版本仅支持任务类型 `http`
- 本项目不包含后端，需配合 ksana-service 使用
