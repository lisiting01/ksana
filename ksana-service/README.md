# Ksana 定时调度服务

基于 Go 实现的轻量级定时调度服务，支持 HTTP 回调任务的一次性（once）和周期性（every）执行。

## 特性

- 支持一次性任务（once）和周期性任务（every）
- HTTP 回调执行方式（GET/POST）
- JSON 文件持久化存储
- 任务失败重试机制
- 优雅关闭和生命周期管理
- REST API 管理接口
- 结构化日志输出
- API 密钥鉴权机制

## 快速开始

### 构建

```bash
go build -o ksana-service .
```

### 运行

```bash
./ksana-service
```

服务默认在 7100 端口启动。

### 环境变量配置

- `PORT`: 服务端口 (默认: 7100)
- `DATA_DIR`: 数据存储目录 (默认: ./data)
- `WORKERS`: 工作线程数 (默认: 4)
- `DEFAULT_TIMEOUT`: 默认超时时间 (默认: 10s)
- `MAX_RETRIES`: 最大重试次数 (默认: 3)
- `RETRY_BACKOFF`: 重试退避时间 (默认: 5s)
- `LOG_LEVEL`: 日志级别 (默认: info)
- `AUTH_KEYS_FILE`: API 密钥文件路径 (默认: ./config/api_keys.txt)

## 鉴权配置

Ksana 服务支持基于 API 密钥的鉴权机制。

### API 密钥文件格式

在 `AUTH_KEYS_FILE` 指定的路径创建密钥文件（默认：`./config/api_keys.txt`），格式如下：

```
# API 密钥文件 - 每行一个密钥
# 以 # 开头的行为注释，空行将被忽略

api-key-for-admin-system
api-key-for-monitoring-system

# 可以添加更多密钥
another-valid-api-key
```

### 鉴权说明

- 除 `/health` 端点外，所有 API 请求都需要提供有效的 API 密钥
- 支持两种鉴权方式：
  - `Authorization: ApiKey <your-api-key>`
  - `X-API-Key: <your-api-key>`
- 缺失或无效的密钥将返回 401/403 错误
- 鉴权失败会记录客户端 IP、路径等信息到日志

## 架构概览

- 单进程服务，由 HTTP 管理接口、调度器、执行器和 JSON 存储组成
- HTTP 层负责 REST API、请求校验、日志记录与鉴权
- 调度器基于最小堆维护任务的 `next_run_at`，使用单个计时器驱动到期分发
- 执行器是带重试的工作池，复用共享 `http.Client`，按任务配置控制超时与退避
- 存储层将任务保存在 JSON 文件中，提供内存镜像与原子落盘能力

## 任务数据模型

默认数据文件位于 `./data/jobs.json`（可通过 `DATA_DIR` 调整）。存储格式示例：

```json
{
  "version": 1,
  "updated_at": "2025-09-19T11:00:00Z",
  "jobs": [
    {
      "id": "uuid-1",
      "name": "ping service",
      "enabled": true,
      "type": "http",
      "http": {
        "method": "GET",
        "url": "https://example.com/ping",
        "headers": {"X-Token": "abc"},
        "body": ""
      },
      "schedule": {
        "kind": "every",
        "every": "5m",
        "start_at": "",
        "jitter": "5s"
      },
      "timeout": "10s",
      "max_retries": 3,
      "retry_backoff": "5s",
      "last_run_at": "",
      "next_run_at": "2025-09-19T11:05:00Z",
      "last_status": "",
      "last_error": ""
    }
  ]
}
```

关键信息：

- `schedule.kind` 支持 `once` 和 `every`；`every` 任务可选 `start_at` 与 `jitter`
- 所有时间字段使用 UTC RFC3339 字符串；持续时间使用 Go duration 语法（如 `5m30s`）
- 运行状态字段：`last_status` 取值包括 `success`、`failed`、`timeout`、`skipped`、`paused`、`missed`

## 调度与执行行为

- once 任务按计划运行一次，若服务启动时已过期会标记为 `missed` 不再补偿
- every 任务执行后基于计划时间滚动到下一次，重启时会跳过已过期的窗口
- `run-now` 命令立即触发执行，但不会改变任务的周期计划
- 失败或超时将按照 `MAX_RETRIES` 与 `RETRY_BACKOFF` 重试；超过阈值后记录最终状态
- 执行阶段会记录 `last_run_at` 与最新错误摘要，便于排查

## 持久化与运行注意事项

- JSON 文件使用临时文件 + 原子重命名写入，减少崩溃时的数据损坏风险
- 服务启动时会加载全部任务并构建内存堆；保存失败会阻止启动
- 优雅关闭：拦截信号后依次关闭 HTTP、停止调度器、等待执行器完成收尾
- 建议通过结构化日志（`log/slog`）收集关键字段：`job_id`、`name`、`status`、`latency_ms`
- 单进程部署场景，不提供跨节点竞争与补偿机制

## 测试建议

- once/every 两类任务的调度与状态更新
- run-now 操作在存在周期任务时的行为
- 超时与重试路径是否符合期望
- JSON 存储在高频更新下的完整性与恢复能力

## API 文档

详细的 API 使用示例请参考 [EXAMPLES.md](EXAMPLES.md)

### 主要端点

- `POST /jobs` - 创建任务
- `GET /jobs` - 列出所有任务
- `GET /jobs/{id}` - 获取单个任务
- `PATCH /jobs/{id}` - 更新任务
- `DELETE /jobs/{id}` - 删除任务
- `POST /jobs/{id}/run-now` - 立即执行任务
- `POST /jobs/{id}/pause` - 暂停任务
- `POST /jobs/{id}/resume` - 恢复任务
- `GET /health` - 健康检查

## Docker 部署

使用 Docker 部署时，需要挂载 API 密钥文件：

```bash
# 创建配置目录和密钥文件
mkdir -p ./config
echo "your-api-key-here" > ./config/api_keys.txt

# 运行容器
docker run -d \
  --name ksana-service \
  -p 7100:7100 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config:/app/config \
  -e AUTH_KEYS_FILE=/app/config/api_keys.txt \
  ksana-service:latest
```

## 后续规划

- 扩展 Cron 表达式与更灵活的调度策略
- 增强执行历史、监控与告警
- 支持多实例选主与幂等保障
- 提供更强的鉴权与签名校验能力

