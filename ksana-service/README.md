# Ksana 定时调度服务

基于Go实现的轻量级定时调度服务，支持HTTP回调任务的一次性（once）和周期性（every）执行。

## 特性

- 支持一次性任务（once）和周期性任务（every）
- HTTP回调执行方式（GET/POST）
- JSON文件持久化存储
- 任务失败重试机制
- 优雅关闭和生命周期管理
- REST API管理接口
- 结构化日志输出

## 快速开始

### 构建

```bash
go build -o ksana-service .
```

### 运行

```bash
./ksana-service
```

服务默认在7100端口启动。

### 环境变量配置

- `PORT`: 服务端口 (默认: 7100)
- `DATA_DIR`: 数据存储目录 (默认: ./data)
- `WORKERS`: 工作线程数 (默认: 4)
- `DEFAULT_TIMEOUT`: 默认超时时间 (默认: 10s)
- `MAX_RETRIES`: 最大重试次数 (默认: 3)
- `RETRY_BACKOFF`: 重试退避时间 (默认: 5s)
- `LOG_LEVEL`: 日志级别 (默认: info)

## API文档

详细的API使用示例请参考 [EXAMPLES.md](EXAMPLES.md)

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

## 设计文档

详细的系统设计和架构说明请参考 [DESIGN.md](DESIGN.md)