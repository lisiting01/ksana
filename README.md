# 项目说明

Ksana 是一个轻量级的定时调度系统，由后端服务 `ksana-service` 与前端控制台 `ksana-web` 组成：

- ksana-service：Go 实现的调度与执行核心，提供 REST API、任务持久化与重试机制。
- ksana-web：基于 Vue 3 + Vite + Element Plus 的 Web 控制台（已实现），用于可视化管理任务、健康检查和系统设置。

更多细节：
- `ksana-service/README.md`：服务端使用与配置
- `ksana-service/EXAMPLES.md`：完整 API 示例（curl）
- `ksana-service/DESIGN.md`：服务端设计与架构
- `ksana-web/README.md`：前端功能与使用说明
- `ksana-web/DESIGN.md`：前端设计文档

## 项目结构

- `ksana-service`：Go 实现的调度服务与 REST API
- `ksana-web`：Web 管理控制台

## 技术栈

- 服务端：Go（见 `ksana-service/go.mod`）
- 前端：Vue 3 + Element Plus + Pinia + Vue Router + Vite + TypeScript

## 快速开始（本地运行）

1) 启动后端服务（默认端口 7100）

```bash
cd ksana-service
go build -o ksana-service .
./ksana-service    # Windows 下运行 ksana-service.exe
```

健康检查：`GET http://localhost:7100/health`

2) 启动前端控制台

```bash
cd ksana-web
npm install

# 可选：设置后端 API 地址（默认 http://localhost:7100）
# 在项目根目录创建 .env.local 并写入：
# VITE_API_BASE_URL=http://localhost:7100

npm run dev
```

在浏览器打开 Vite 输出的本地地址，进入“健康检查”页面确认服务状态为 ok。

3) 创建一个周期任务示例（每 5 分钟执行一次 GET 请求）

```bash
curl -X POST http://localhost:7100/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ping-example",
    "enabled": true,
    "type": "http",
    "http": {"method": "GET", "url": "https://example.com/ping", "headers": {}, "body": ""},
    "schedule": {"kind": "every", "every": "5m"},
    "timeout": "10s",
    "max_retries": 3,
    "retry_backoff": "5s"
  }'
```

也可在 `ksana-web` 的“新建任务”页面可视化创建。

## 运行配置（服务端环境变量）

来自 `ksana-service` 的配置（括号内为默认值）：
- `PORT`：服务端口（`7100`）
- `DATA_DIR`：数据目录（`./data`），等价别名 `KSANA_DATA`
- `WORKERS`：执行并发度（`4`）
- `DEFAULT_TIMEOUT`：HTTP 执行默认超时（`10s`）
- `MAX_RETRIES`：失败最大重试次数（`3`）
- `RETRY_BACKOFF`：固定退避时长（`5s`）
- `LOG_LEVEL`：日志级别（`info`，可选 `debug|info|warn|error`）

数据默认持久化到 `ksana-service/data/jobs.json`（自动创建，JSON 原子写入）。

## 前端配置（环境变量）

- `VITE_API_BASE_URL`：后端 API 根地址，默认 `http://localhost:7100`。
- 运行期也可在“设置”页面修改 API 地址，保存到浏览器 localStorage。

## API 概览（服务端）

- `POST /jobs` 创建任务
- `GET /jobs` 列出任务
- `GET /jobs/{id}` 任务详情
- `PATCH /jobs/{id}` 更新任务（部分字段）
- `DELETE /jobs/{id}` 删除任务
- `POST /jobs/{id}/run-now` 立即执行一次（不影响周期）
- `POST /jobs/{id}/pause` 暂停任务
- `POST /jobs/{id}/resume` 恢复任务
- `GET /health` 健康检查

字段与校验规则见 `ksana-service/internal/api/dto.go` 与 `ksana-service/internal/model/*`。

## 前端功能概览（ksana-web）

- 任务列表：筛选、快捷操作（查看/编辑/run-now/启停/删除）
- 任务表单：HTTP 配置、调度（once/every）、执行控制（超时/重试）
- 任务详情：只读信息、JSON 视图、操作菜单
- 健康检查：显示后端状态
- 系统设置：API 地址与语言（本地持久化）

更多内容见 `ksana-web/README.md` 与 `ksana-web/DESIGN.md`。

## 开发与要求

- Go 版本：见 `ksana-service/go.mod`
- Node 版本：见 `ksana-web/package.json` 的 engines（建议 Node 20.19+ 或 22.12+）
- 时间处理：统一使用 UTC；`every`/`timeout`/`retry_backoff` 使用 Go Duration（如 `5s`、`1m30s`）

## 路线图（Roadmap）

- 鉴权与访问控制（API Token 等）
- 更多调度类型（如 Cron 表达式）
- 执行历史与日志流、指标与可观测性

## AI 声明

本项目的绝大部分实现由 AI 辅助生成，项目动机之一是探索“与多个 AI 协同工作”的模式与流程；作者主要负责功能/架构设计与人工审阅。详情见 `AI_STATEMENT.md`。
