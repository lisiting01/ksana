# ksana-web 设计文档（基于 Vue 3 + Vite + Element Plus）

版本：MVP v0.1（与 ksana-service v0.1 配套）

## 1. 目标与范围

- 目标：提供一个可视化的 Web 控制台，使用 ksana-service 的 REST API 完成任务的可视化管理（创建/查询/更新/删除/启停/立即执行），以及基础健康检查展示。
- 范围（MVP）：
  - 任务列表、搜索/筛选（按启用状态/类型/名称关键字）
  - 任务创建/编辑向导（HTTP 配置、调度 once/every、超时/重试）
  - 任务详情（只读信息、JSON 视图、快捷操作：run-now、pause/resume、delete）
  - 健康检查页（/health）
  - 全局基础错误提示与字段级表单校验
- 非范围（后续）：鉴权、执行历史明细/日志流、图表监控、角色权限、Cron表达式、WebSocket 实时更新、回调目标探活/验证。

## 2. 端到端交互流程（概览）

1) 用户在“任务列表”点击“新建任务” → 进入表单页 → 根据 schedule.kind 动态渲染 once/every 字段 → 前端完成基础校验 → 调用 `POST /jobs` 创建 → 返回后跳转到详情页或列表。
2) 在“任务列表/详情”执行快捷操作：run-now/pause/resume/delete → 调用对应 REST → 更新本地列表并气泡提示。
3) 页面初始化时拉取 `GET /jobs`；详情页按需 `GET /jobs/{id}`；全局定期轻量刷新（可选，MVP 默认手动刷新）。
4) “健康检查”直接调用 `GET /health` 并展示状态。

## 3. 前端架构与技术选型

- 框架：Vue 3（SFC + `<script setup>`）
- 构建：Vite
- 组件库：Element Plus
- 状态管理：Pinia（jobs、settings 两类 store）
- 路由：Vue Router（hash/history 任一，默认 history）
- HTTP 客户端：fetch（封一层）或 axios（二选一，MVP用 fetch 即可）
- 国际化：vue-i18n（预留，默认中文）
- 代码风格：ESLint + Prettier（项目内配置）

## 4. 页面与路由

- `/` Dashboard（可选，MVP 简化为跳转 `/jobs`）
- `/jobs` 任务列表页
  - 表格列：name、enabled、type、scheduleSummary、next_run_at、last_status、last_run_at、操作
  - 操作：查看、编辑、run-now、pause/resume、删除
  - 过滤：名称关键字、enabled、type（http）
- `/jobs/new` 新建任务
- `/jobs/:id` 任务详情（只读 + 操作）
- `/jobs/:id/edit` 任务编辑
- `/health` 健康检查
- `/settings` 设置（API 地址、语言，MVP 仅本地存储）

## 5. 数据模型与 TypeScript 类型

与 ksana-service 对齐（参见 `internal/api/dto.go` 与 `internal/model`）。时间一律使用 UTC 的 ISO 字符串（RFC3339）。

```ts
// src/types/job.ts
export type UUID = string;

export interface DurationString extends String {} // 例如 "5s"、"1m30s"，前端保持字符串透传

export interface HTTPConfig {
  method: 'GET' | 'POST';
  url: string;
  headers: Record<string, string>;
  body: string; // 原样字符串
}

export type ScheduleKind = 'once' | 'every';

export interface ScheduleOnce {
  kind: 'once';
  run_at: string; // RFC3339 UTC
}

export interface ScheduleEvery {
  kind: 'every';
  every: DurationString; // 必填
  start_at?: string; // RFC3339 UTC，可选
  jitter?: DurationString; // 可选
}

export type Schedule = ScheduleOnce | ScheduleEvery;

export interface CreateJobRequest {
  name: string;
  enabled?: boolean; // 默认 true
  type: 'http';
  http: HTTPConfig;
  schedule: Schedule;
  timeout?: DurationString; // 默认 10s
  max_retries?: number; // 默认 3
  retry_backoff?: DurationString; // 默认 5s
}

export interface UpdateJobRequest {
  name?: string;
  enabled?: boolean;
  http?: HTTPConfig;
  schedule?: Schedule;
  timeout?: DurationString;
  max_retries?: number;
  retry_backoff?: DurationString;
}

export interface JobResponse {
  id: UUID;
  name: string;
  enabled: boolean;
  type: 'http';
  http: HTTPConfig;
  schedule: Schedule;
  timeout: DurationString;
  max_retries: number;
  retry_backoff: DurationString;
  last_run_at?: string; // ISO
  next_run_at?: string; // ISO
  last_status: 'success' | 'failed' | 'timeout' | 'skipped' | 'paused' | 'missed' | '';
  last_error: string;
}
```

派生字段：`scheduleSummary`（前端渲染，便于表格展示，如 "every 5m (start now)" / "once 2025-09-20T03:00:00Z"）。

## 6. API 适配层

- 基础：统一的 `apiBase`（环境变量 `VITE_API_BASE_URL`，默认 `http://localhost:7100`）。
- 约定：
  - 成功：`2xx`；
  - 可重试错误：`408/429/5xx`（由后端用于回调；前端仅提示）；
  - 其他 4xx：参数错误，展示 `error/message`。

```ts
// src/api/http.ts
export const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:7100';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${apiBase}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
    ...init,
  });
  const text = await res.text();
  const data = text ? JSON.parse(text) : undefined;
  if (!res.ok) {
    throw new Error((data && (data.message || data.error)) || res.statusText);
  }
  return data as T;
}

// src/api/jobs.ts
import type { CreateJobRequest, UpdateJobRequest, JobResponse } from '@/types/job';

export const JobsAPI = {
  list: () => request<JobResponse[]>('/jobs'),
  get: (id: string) => request<JobResponse>(`/jobs/${id}`),
  create: (payload: CreateJobRequest) => request<JobResponse>('/jobs', { method: 'POST', body: JSON.stringify(payload) }),
  update: (id: string, payload: UpdateJobRequest) => request<JobResponse>(`/jobs/${id}`, { method: 'PATCH', body: JSON.stringify(payload) }),
  remove: (id: string) => request<void>(`/jobs/${id}`, { method: 'DELETE' }),
  runNow: (id: string) => request<{ message: string }>(`/jobs/${id}/run-now`, { method: 'POST' }),
  pause: (id: string) => request<{ message: string }>(`/jobs/${id}/pause`, { method: 'POST' }),
  resume: (id: string) => request<{ message: string }>(`/jobs/${id}/resume`, { method: 'POST' }),
  health: () => request<{ status: string }>('/health'),
};
```

## 7. Store 设计（Pinia）

- `useJobsStore`
  - state：`items: JobResponse[]`、`loading: boolean`、`error?: string`
  - actions：`fetchAll()`、`getById(id)`、`create(payload)`、`update(id, payload)`、`remove(id)`、`runNow(id)`、`pause(id)`、`resume(id)`
  - 逻辑：操作成功后刷新列表或局部替换；失败统一 toast。
- `useSettingsStore`
  - state：`apiBase`、`locale`；持久化到 `localStorage`。

## 8. 表单与校验（与服务端规则对齐）

- 通用：
  - `name`：必填、长度限制（前端 1~100，后端未给出则宽松）
  - `type`：固定 `http`
  - `http.method`：`GET|POST`；`GET` 时 `body` 可留空
  - `http.url`：必填、URL 格式校验（http/https）
  - `http.headers`：键名非空、去除首尾空白；键名大小写不敏感，前端保持大小写
  - `timeout`、`retry_backoff`、`every/jitter`：必须是 Go Duration 字符串（提供帮助提示/示例）
  - `max_retries`：>= 0 的整数
- 调度：
  - `schedule.kind`：`once` 或 `every`
  - `once`：必须提供 `run_at`（RFC3339 UTC），并提示“统一使用 UTC”
  - `every`：必须提供 `every`；可选 `start_at/jitter`
- 编辑时注意：服务端 PATCH 允许部分字段；切换 kind 需提示会重置 `next_run_at`（后端会清空并重算）。

## 9. 时间与时区

- 前端内部统一使用 UTC。输入/显示：
  - 输入：提供时区切换的提示但默认 UTC；或使用日期时间控件 + “以 UTC 提交”提示（MVP 先直接输入 ISO 字符串，提供转换帮助按钮）
  - 显示：在列表/详情中以 UTC ISO 展示，并可提供“本地时间”悬浮提示（可选）

## 10. UI 细节（Element Plus）

- 列表：`el-table` + 操作列（查看/编辑/运行/启停/删除）
- 表单：`el-form` + 分组（基础信息、HTTP 配置、调度、执行控制）
- 选择器：`el-select`（method/kind）、`el-input`（url/body）、`el-input-number`（max_retries）
- 反馈：`ElMessage`、`ElMessageBox`（删除二次确认）

## 11. 目录结构（建议）

```
ksana-web/
  ├─ index.html
  ├─ vite.config.ts
  ├─ package.json
  └─ src/
     ├─ main.ts
     ├─ router/
     │   └─ index.ts
     ├─ stores/
     │   ├─ jobs.ts
     │   └─ settings.ts
     ├─ api/
     │   ├─ http.ts
     │   └─ jobs.ts
     ├─ types/
     │   └─ job.ts
     ├─ views/
     │   ├─ JobsList.vue
     │   ├─ JobForm.vue
     │   ├─ JobDetail.vue
     │   ├─ Health.vue
     │   └─ Settings.vue
     ├─ components/
     │   ├─ ScheduleEditor.vue
     │   └─ HttpConfigEditor.vue
     └─ utils/
         ├─ schedule.ts   // scheduleSummary 等
         └─ time.ts       // UTC/ISO 转换、示例生成
```

## 12. 构建与本地开发

- Node 版本：≥ 18.x（Vite 推荐版本）
- 依赖：`vue@3`、`vue-router@4`、`pinia`、`element-plus`、`@vitejs/plugin-vue`、`typescript`
- 环境变量：`.env` / `.env.local`
  - `VITE_API_BASE_URL=http://localhost:7100`
- 常用脚本：
  - `dev`：本地开发（`vite`）
  - `build`：打包产物（静态资源，可 Nginx/静态托管）
  - `preview`：本地预览

## 13. 错误处理与可用性

- 全局捕获：request 层统一抛出 Error；组件使用 try/catch 显示 `ElMessage.error`
- 表单内联错误：Element Plus rule 验证 + 顶部 summary（可选）
- 加载态：列表/详情/提交过程中的 loading 状态与禁用按钮
- 空态：无数据/首屏引导

## 14. 安全与后续扩展

- 生产部署建议：
  - Web 静态资源可与 ksana-service 分离部署（跨域支持：可由反向代理统一域名解决，或补充 CORS 配置）
  - 后续引入 API Token 时，前端在请求头携带 `Authorization: Bearer <token>`（存储于安全位置，MVP 不实现）
- 输入安全：
  - Headers 键值对在前端做基础校验，防止注入；Body 文本不做转义直接透传
  - URL 仅允许 http/https
- 未来功能占位：
  - 执行历史与日志流：新增 `/runs` 相关接口后，前端扩展历史面板与实时流（WebSocket/EventSource）
  - 仪表盘：数量统计、状态分布、即将执行的任务
  - 多语言：内置 zh-CN/en-US 词条

## 15. 与 ksana-service 的对齐要点

- API 路由：严格按 `router.go` 定义：
  - `POST /jobs`、`GET /jobs`、`GET /jobs/{id}`、`PATCH /jobs/{id}`、`DELETE /jobs/{id}`
  - `POST /jobs/{id}/run-now`、`/pause`、`/resume`
  - `GET /health`
- DTO 字段：按 `dto.go` 与 `model/types.go`；前端不生成额外字段，仅在 UI 层派生摘要。
- 时间：展示 `last_run_at/next_run_at`；不存储历史（MVP）。

## 16. 验收标准（前端侧）

- 能创建 once/every 两类任务，参数通过服务端校验并成功落库。
- 列表准确展示关键字段，支持基本筛选；操作按钮可用。
- run-now/pause/resume/delete 操作可达并反馈结果；错误有提示。
- 健康页能正确反映后端状态。
- 所有时间显示为 UTC ISO 字符串，表单输入有格式提示。

---

注：本设计文档仅描述 ksana-web 的 MVP 实现规划，不修改任何后端代码；如后端 API 变更，请同步更新 src/types 与 API 适配层。

