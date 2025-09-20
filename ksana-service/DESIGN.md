# Ksana 定时调度服务（JSON 持久化）设计

版本：v0.1-MVP（单进程，JSON 文件持久化）

## 1. 概述
- 目标：提供一个轻量、可落地的单机定时调度服务，使用 JSON 文件持久化任务；支持一次性（once）与固定间隔（every）两类任务，采用 HTTP 回调作为执行方式。
- 非目标（本阶段不做）：分布式高可用、Cron 表达式、复杂并发/速率限制、GUI 控制台、鉴权体系。

## 2. MVP 范围
- 任务类型：HTTP 回调（GET/POST，支持 headers 与 body）。
- 调度类型：
  - once：在指定时间运行一次。
  - every：固定间隔周期开启，可选 `start_at` 与 `jitter`。
- 基础能力：启停（enabled）、手动触发（run-now）、超时控制、固定次数重试（固定退避）。
- 持久化：单个 JSON 文件存储全部任务与元数据；内存镜像 + 原子写回。
- 单进程：不考虑多实例竞争，不做跨进程锁与补偿。

### 2.1 开发清单（MVP）
- model：Job/Schedule/HTTPConfig，自定义 Duration JSON 编解码与字段校验。
- store：加载/保存 JSON、原子写回、损坏备份、内存镜像（RWMutex）。
- scheduler：最小堆 + 单定时器循环；once/every 触发与 misfire 规则。
- executor：http.Client 复用、context 超时、固定退避重试、结果回写。
- api：/jobs CRUD、/run-now、/pause|/resume、/health（本阶段无鉴权）。
- lifecycle：优雅关闭（停止接收->等待/取消->最终保存）。

### 2.2 验收标准
- once 任务按时执行；过期 once 标记为 missed 不补偿。
- every 任务按间隔滚动；重启后直接滚动到“下一次”。
- run-now 触发即时执行且不影响周期 next_run_at。
- 失败与超时按配置重试，超过次数后记录 failed/timeout。
- JSON 原子写：高频更新/执行下文件未损坏；异常重启可完整恢复。
- 管理 API 可创建、查询、更新、删除、启停任务；/health 返回 200。

## 3. 架构与组件
- 进程结构：
  - HTTP API（管理接口）
  - Scheduler（小根堆优先队列，按 `next_run_at` 排序）
  - Worker 池（固定并发，执行 HTTP 回调）
  - Store（JSON 读写，原子落盘）
- 数据流：API 变更 -> 更新内存模型 -> 原子写 JSON -> 通知 Scheduler 更新堆项 -> 到期分发到 Worker -> 执行结果回写。
- 退出流程：停止接收新任务 -> 等待进行中任务（或超时取消） -> Flush JSON -> 退出。

## 4. 数据模型（JSON）
- 文件：`data/jobs.json`（可通过环境变量 `KSANA_DATA`/`DATA_DIR` 覆盖目录）
- 顶层结构示例：
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
- 字段说明（关键）：
  - job.type 固定为 `http`；未来可扩展其他执行器。
  - schedule:
    - kind: `once` 或 `every`。
    - once: `run_at`（RFC3339）。
    - every: `every`（Go duration），可选 `start_at`（RFC3339，空则以创建时刻为起点），可选 `jitter`（duration，默认 0）。
  - 时间与时区：存储统一使用 UTC（RFC3339）。
  - bookkeeping: `last_run_at`、`next_run_at`、`last_status`（success|failed|timeout|skipped|paused|missed）、`last_error`。

## 5. 调度器设计
- 数据结构：最小堆（优先队列），键为 `next_run_at`（UTC）。
- 启动：加载 JSON -> 校验 -> 为 every 计算初始 `next_run_at`（若为空）-> 将 enabled 且未过期/需继续的 job 入堆。
- 触发：主循环阻塞等待最早到期；到期弹出 -> 投递到 Worker；
  - once：执行后标记完成（不再入堆）；若当前时间已过期，标记 `missed` 不补偿。
  - every：执行后基于当前时间滚动到下一次，写回 `next_run_at`，再入堆。
- tick 与兜底：当无任务时阻塞；额外每 1s/5s 低频 tick 防止边界误差。
- Misfire 策略（MVP）：
  - once：过期直接 `missed`，不补偿。
  - every：按“现在”滚动到下一次，不补偿历史多次。

## 6. 执行器（HTTP）
- Worker 池：固定并发度（默认 4，可配置 `WORKERS`）。
- 超时：执行使用 context + HTTP 客户端超时（默认 `DEFAULT_TIMEOUT`）。
- 重试：失败时按固定退避 `RETRY_BACKOFF`，最多 `MAX_RETRIES` 次；超时计为失败。
- 结果回写：更新 `last_run_at`、`last_status`、`last_error`；for every 计算下次 `next_run_at` 并写回。
- 幂等：由回调方负责（本服务不去重）。

## 7. 持久化与原子性
- 内存镜像：服务运行时维护 jobs 列表与索引（map[id]job）。
- 原子写：每次变更落盘时先写临时文件 `jobs.json.tmp` -> 调用 `rename` 覆盖 `jobs.json`，减少损坏概率。
- 并发控制：进程内互斥（mutex）序列化写操作；读使用内存镜像。
- 启动恢复：
  - JSON 解析失败：将原文件重命名为 `jobs.bad.json`，新建空模板文件继续启动。
  - every：若 `next_run_at` 落后当前时间，直接从“现在”滚动到下一次。

## 8. 管理 API（REST）
- 端点与负载（简化约定）：
  - POST `/jobs`
    - 输入：job（除系统字段外），`id` 可留空由服务生成；时间均为 UTC。
    - 输出：创建后的完整 job。
  - GET `/jobs`
    - 输出：job 列表（分页暂不做）。
  - GET `/jobs/{id}`
    - 输出：单个 job。
  - PATCH `/jobs/{id}`
    - 输入：部分字段更新（如 `name`、`enabled`、`schedule`、`http`、`timeout`、`max_retries`、`retry_backoff`）。
    - 输出：更新后的 job。
  - DELETE `/jobs/{id}`
    - 输出：204 无内容。
  - POST `/jobs/{id}/run-now`
    - 行为：立即提交一次执行（不改变周期与 `next_run_at`）。
  - POST `/jobs/{id}/pause` | `/jobs/{id}/resume`
    - 行为：设置 `enabled=false/true`，并更新堆项。
  - GET `/health`
    - 用于存活/就绪探测（返回 200）。

## 9. 配置
- 环境变量：
  - `PORT`（默认 7100）
  - `DATA_DIR` 或 `KSANA_DATA`（默认 `./data`）
  - `WORKERS`（默认 4）
  - `DEFAULT_TIMEOUT`（默认 `10s`）
  - `MAX_RETRIES`（默认 3）
  - `RETRY_BACKOFF`（默认 `5s`）
  - `LOG_LEVEL`（默认 `info`）

## 10. 目录结构（建议）
```
ksana-service/
  main.go            # 入口与 HTTP 路由注册、依赖装配
  README.md
  DESIGN.md          # 本文件
  internal/
    api/             # handler 与请求校验、路由
      router.go      # 路由注册与中间件（日志）
      handlers.go    # /jobs、/health 等 HTTP 处理函数
      dto.go         # 请求/响应 DTO 与校验
    scheduler/       # 堆、调度循环、时间滚动
      scheduler.go   # 调度器核心循环（timer+heap）
      heap.go        # 最小堆封装（container/heap）
      clock.go       # 时钟接口，便于测试注入
    store/           # JSON 读写、原子落盘
      store.go       # 接口定义（Load/Save/List/Get/Put/Delete）
      json_store.go  # JSON 文件实现（原子写、损坏备份）
    executor/        # 执行器与重试
      http_exec.go   # HTTP 回调执行（client、超时、重试）
    model/           # Job 数据结构与校验
      types.go       # Job、Schedule、HTTPConfig 等类型
      duration.go    # 自定义 Duration 的 JSON 编解码
      validate.go    # URL、时间、枚举校验
  data/
    jobs.json        # 运行时生成/持久化文件（不必预置）
```

命名与组织约定：
- internal 仅对本模块可见，避免外部依赖耦合；包名小写、短名（api/scheduler/store/executor/model）。
- 每个子包提供清晰的对外接口（interface 或函数集），main.go 负责装配依赖并注入配置。
- 错误处理优先返回显式错误值，日志在边界层（api/executor/scheduler）记录；model/store 尽量无日志。

## 11. 开发里程碑与任务清单
- D1 数据模型与校验：定义模型结构、默认值、字段校验（duration/URL/时间）。
- D2 JSON 存储：加载/保存、原子写、损坏备份与空模板初始化。
- D3 调度器：小根堆与时间滚动；启动初始化；once/every 策略与 misfire。
- D4 执行器：Worker 池、HTTP 请求、超时、重试、结果回写。
- D5 API：CRUD、启停、run-now、health。
- D6 生命周期：优雅关闭、进行中任务等待/取消、最终 flush。
- D7 自测与日志：基本端到端验证，结构化日志与错误信息对齐。

## 12. 测试要点（最小集）
- once：未来时执行；过期时 `missed` 不补偿。
- every：`start_at` 为空与设定场景；`jitter` 生效；宕机重启后滚动到“下一次”。
- 超时与重试：超时视为失败，按退避重试，最终状态正确回写。
- 原子写一致性：高频更新/执行下文件未损坏，重启可恢复。
- run-now：不影响周期 `next_run_at`；并发执行时状态更新正确。

## 13. 运行与运维
- 日志：结构化（时间、job_id、name、status、latency、error 摘要）。
- 健康检查：`/health` 返回 200；未来可扩展 `/metrics`。
- 资源参数：根据负载调节 `WORKERS` 与超时；避免过小间隔导致空转。
- 平台假设：单进程部署，Windows/Linux 均可运行。

## 14. 未来扩展（超出 MVP）
- Cron 表达式与日历排除。
- 运行历史表（runs）与更丰富可观测性。
- 多实例下的选主/锁与幂等。
- 失败告警与通知渠道（Webhook/Email）。

## 15. Go 实现要点（stdlib 优先）
- Go 版本：建议 Go 1.21+（可用 `log/slog`，`signal.NotifyContext`）。
- 依赖偏好：MVP 尽量仅使用标准库；如需 UUID 可先用 `crypto/rand` 生成 16 字节并编码为十六进制/URL 安全 Base64，避免外部依赖。
- 包/模块：
  - `internal/model`：`Job`/`Schedule`/`HTTPConfig`/自定义 `Duration` 类型（`time.Duration` 的 JSON 字符串封装，支持 "5s" 形式）。
  - `internal/store`：`Load(ctx)`/`Save(ctx)`/`List`/`Get`/`Put`/`Delete`，内存镜像 + `sync.RWMutex`；落盘使用原子写（见下）。
  - `internal/scheduler`：容器堆（`container/heap`）+ 单个 `time.Timer` 动态 `Reset`；一个 goroutine 驱动调度循环。
  - `internal/executor`：共享 `http.Client`（连接池+超时），执行与重试逻辑。
  - `internal/api`：`net/http` + `encoding/json` 处理 REST，表单/JSON 校验与错误响应。
  - `main.go`：装配、配置读取（`os.Getenv`）、路由、优雅退出。
- 计时与时区：所有持久化时间使用 `time.Time` 的 UTC（`time.Now().UTC()`/`time.Parse(time.RFC3339)`）。
- 调度循环建议：
  - 维护一个最小堆；取堆顶时间 `t0`，用单个 `time.Timer` 等待；到期后批量取出<=`now` 的任务并投递到工作队列。
  - 每次新增/更新任务时：若其 `next_run_at` 早于当前堆顶，`Reset` 定时器。
  - 对 `every` 任务的下一次触发，以“计划时间滚动”而非“当前时间 + every”避免漂移：`next = last_scheduled.Add(every)`，直到 `next > now`；再可选叠加 `jitter`。
- Worker 池：
  - `workers = make(chan JobID, N)` 或使用 `taskCh chan *Job`，`N` 个 goroutine 读取并执行。
  - 执行使用 `context.WithTimeout`；重试通过 `for i := 0; i < max; i++ { ... time.Sleep(backoff) }` 或内部 `time.Timer`（在 worker 内部同步重试即可）。
- HTTP 客户端：
  - 共享一个 `http.Client{ Timeout: defaultTimeout, Transport: &http.Transport{ MaxIdleConns: 100, IdleConnTimeout: 90 * time.Second, TLSHandshakeTimeout: 10 * time.Second, ExpectContinueTimeout: 1 * time.Second } }`。
  - 构造请求时拷贝 headers；`Content-Type` 由 job 指定或根据 body 推断。
- JSON 编码/解码：
  - 自定义 `Duration` 类型实现 `MarshalJSON/UnmarshalJSON` 以支持字符串形式；`time.Time` 使用 RFC3339 文本。
  - PATCH 局部更新可使用输入 DTO 的指针字段或 `map[string]any` + 校验后落库。
- 原子写（Windows/Unix 兼容）：
  - 步骤：写入同目录临时文件 -> `file.Sync()` 确保持久化 -> 关闭 -> 目录 `Sync`（可选，Unix 常见）-> 使用 `os.Rename(tmp, dst)` 进行替换。
  - Windows 注意：若 `dst` 已存在，部分环境下 `os.Rename` 可能失败；可先尝试 `os.Rename`，失败时回退为 `os.Remove(dst); os.Rename(tmp, dst)`；确保整个过程在同一卷且同一目录以最大化原子性。
- 优雅退出：
  - `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)`；Unix 可加 `syscall.SIGTERM`。
  - 关闭对外服务后，`WaitGroup` 等待进行中任务，超时则取消其 context。
- 日志：优先使用 `log/slog`（Go 1.21）；降级可用 `log`。日志字段包括：`job_id`、`name`、`status`、`latency_ms`、`error` 摘要。
- 测试：table-driven；`httptest.Server` 模拟 HTTP 回调；可为时间封装 `Clock` 接口以便注入测试时钟；并发下用 `-race` 运行。

## 16. 项目文件组织（细化与演进）
- 单二进制：保留根目录 `main.go`，不引入 `cmd/` 结构（MVP 简化）。
- 配置装配：`main.go` 读取环境变量 -> 实例化 `store.JSONStore`、`executor.HTTP`、`scheduler.Scheduler`、注册 `api.Router`。
- 依赖注入：无需引入 DI 框架；通过构造函数传入依赖（如 `NewScheduler(store, exec, clock)`）。
- 版本演进：未来若需要多个可执行体，再拆分为 `cmd/ksana-service/main.go`。

## 17. 与外部服务对接（HTTP 回调约定）
- 触发方式：定时服务作为客户端，按任务配置主动向外部服务发起 HTTP 请求。
- 支持方法：`GET`、`POST`（MVP 主推 `POST`）。
- Headers：可在任务中自定义（如 `Content-Type: application/json` 等）。
- Body：自由字符串；常见为 JSON。建议包含以下字段以便对端幂等：
  - `job_id`：任务标识
  - `run_id`：本次执行标识（定时服务生成，全局唯一）
  - `scheduled_at`：计划触发时间（UTC）
  - `triggered_at`：实际触发时间（UTC）
- 响应与重试约定：
  - 2xx：视为成功（记录 success）。
  - 408/429/5xx：视为可重试错误，按 `max_retries + retry_backoff` 重试。
  - 4xx（除 408/429）：视为永久失败（记录 failed，不再重试）。
- 超时：由定时服务控制（`timeout`），超过则记为 `timeout` 并按可重试处理。
- 幂等：
  - 定时服务在请求头附带 `X-Ksana-Run-Id: <run_id>`（或在 body 中携带），对端应基于 `run_id` 实现幂等处理。
  - 定时服务不做重复请求去重（MVP 简化），以对端幂等为准。

示例任务（POST JSON）：
```json
{
  "id": "", 
  "name": "notify-order",
  "enabled": true,
  "type": "http",
  "http": {
    "method": "POST",
    "url": "https://api.example.com/orders/notify",
    "headers": {"Content-Type": "application/json"},
    "body": "{\n  \"event\": \"order.received\",\n  \"job_id\": \"<filled-by-service>\",\n  \"run_id\": \"<filled-by-service>\",\n  \"payload\": {\"order_id\": 12345}\n}"
  },
  "schedule": {"kind": "every", "every": "5m"},
  "timeout": "10s",
  "max_retries": 3,
  "retry_backoff": "5s"
}
```

## 18. MVP 确认
- 目标场景：由“其他服务提供 HTTP 接口”，本定时服务按计划调用这些接口触发动作。
- 结论：本设计的 MVP 完全支持该方式（HTTP 回调型任务），包含方法、Header/Body 自定义、超时与重试、基础幂等约定。
- 已知取舍：
  - 不做“精确一次”保证，幂等需由被调服务实现。
  - 宕机期间的历史触发不补偿（every 滚动到下一次；once 过期记 missed）。
  - 本阶段无鉴权与签名校验（如需可后续扩展）。
