# 使用示例（供被调用系统与调用方参考）

默认监听端口：7100（HTTP）。以下示例均以 `http://localhost:7100` 为前缀。

**注意：除 `/health` 端点外，所有API请求都需要提供API密钥进行鉴权。**

### 鉴权方式

支持两种方式提供API密钥：

1. **Authorization 头部**（推荐）:
```bash
curl -H "Authorization: ApiKey your-api-key-here" ...
```

2. **X-API-Key 头部**:
```bash
curl -H "X-API-Key: your-api-key-here" ...
```

## 一、管理 API 调用示例

- 创建周期任务（every，间隔 5 分钟）
```
curl -X POST http://localhost:7100/jobs \
  -H "Authorization: ApiKey your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ping-example",
    "enabled": true,
    "type": "http",
    "http": {
      "method": "GET",
      "url": "https://example.com/ping",
      "headers": {"Content-Type": "application/json"},
      "body": ""
    },
    "schedule": {"kind": "every", "every": "5m"},
    "timeout": "10s",
    "max_retries": 3,
    "retry_backoff": "5s"
  }'
```

- 创建一次性任务（once，指定 UTC 触发时间）
```
curl -X POST http://localhost:7100/jobs \
  -H "Authorization: ApiKey your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "once-notify",
    "enabled": true,
    "type": "http",
    "http": {
      "method": "POST",
      "url": "https://api.example.com/notify",
      "headers": {"Content-Type": "application/json"},
      "body": "{\n  \"event\": \"build.done\",\n  \"payload\": {\"id\": 123}\n}"
    },
    "schedule": {"kind": "once", "run_at": "2025-09-20T03:00:00Z"},
    "timeout": "10s",
    "max_retries": 3,
    "retry_backoff": "5s"
  }'
```

- 列出任务
```
curl -H "Authorization: ApiKey your-api-key-here" \
  http://localhost:7100/jobs
```

- 查看任务详情
```
curl -H "Authorization: ApiKey your-api-key-here" \
  http://localhost:7100/jobs/<job_id>
```

- 更新任务（部分字段）
```
curl -X PATCH http://localhost:7100/jobs/<job_id> \
  -H "Authorization: ApiKey your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

- 立即执行一次（不影响周期）
```
curl -X POST -H "Authorization: ApiKey your-api-key-here" \
  http://localhost:7100/jobs/<job_id>/run-now
```

- 暂停/恢复
```
# 暂停
curl -X POST -H "Authorization: ApiKey your-api-key-here" \
  http://localhost:7100/jobs/<job_id>/pause
# 恢复
curl -X POST -H "Authorization: ApiKey your-api-key-here" \
  http://localhost:7100/jobs/<job_id>/resume
```

- 删除任务
```
curl -X DELETE -H "Authorization: ApiKey your-api-key-here" \
  http://localhost:7100/jobs/<job_id>
```

- 健康检查
```
curl http://localhost:7100/health
```

说明：
- 时间一律为 UTC，格式为 RFC3339（如 `2025-09-20T03:00:00Z`）。
- 周期间隔使用 Go duration 字符串（如 `5s`、`1m30s`、`2h`）。

## 二、被调用系统如何对接（HTTP 回调）

定时服务会作为客户端，按任务配置对外发起 HTTP 请求。

- 请求方法：GET/POST（推荐 POST）。
- 请求头：可按任务自定义（例如 `Content-Type: application/json`）。服务会额外附带 `X-Ksana-Run-Id` 作为本次执行标识。
- 请求体：自由定义。建议包含以下字段，便于对端实现幂等与追踪：
  - `job_id`：任务标识
  - `run_id`：执行标识（与 `X-Ksana-Run-Id` 一致）
  - `scheduled_at`：计划触发时间（UTC）
  - `triggered_at`：实际触发时间（UTC）
  - `payload`：业务自定义内容

示例（定时服务向外部服务发起的 POST）：
```
POST /orders/notify HTTP/1.1
Host: api.example.com
Content-Type: application/json
X-Ksana-Run-Id: 5f2a7b7c1a1b4f0e9c0d1e2f3a4b5c6d

{
  "event": "order.received",
  "job_id": "8b8b68b0-4d1e-4dd9-bf0e-7b5b1d5b6f48",
  "run_id": "5f2a7b7c1a1b4f0e9c0d1e2f3a4b5c6d",
  "scheduled_at": "2025-09-19T11:05:00Z",
  "triggered_at": "2025-09-19T11:05:01Z",
  "payload": {"order_id": 12345}
}
```

返回约定：
- 2xx 视为成功。
- 408/429/5xx 视为可重试错误（按任务 `max_retries` 与 `retry_backoff` 重试）。
- 其他 4xx 视为永久失败，不再重试。

幂等建议：
- 以 `X-Ksana-Run-Id` 或 body 中的 `run_id` 作为幂等键，重复请求返回相同结果。

## 三、常见问题

- 服务是否有鉴权？
  - 支持基于API密钥的鉴权机制。在 `AUTH_KEYS_FILE` 文件中配置有效的API密钥，所有API请求（除 `/health` 外）都需要提供有效密钥。如果不配置密钥文件，服务启动时会记录警告信息，此时所有鉴权请求都会被拒绝。
- 宕机重启如何处理？
  - 周期任务（every）重启后直接滚动到下一次，不补偿历史多次；一次性任务（once）若过期则标记为 missed。
- 如何设置并发与超时？
  - 通过环境变量：`WORKERS`、`DEFAULT_TIMEOUT`、`MAX_RETRIES`、`RETRY_BACKOFF`。
