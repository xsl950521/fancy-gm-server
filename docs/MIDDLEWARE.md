# 中间件增强文档

## 概述

本文档介绍已实现的中间件增强功能，包括请求日志、性能监控和安全中间件。

## 已实现的中间件

### 1. 请求日志中间件 (`RequestLogger`)

**功能**：
- 记录每个HTTP请求的详细信息
- 记录请求参数、响应时间、响应状态
- 支持请求体记录（自动跳过文件上传）
- 根据状态码选择日志级别

**配置**：
```yaml
middleware:
  request_logger:
    enabled: true  # 是否启用请求日志
```

**日志字段**：
- `method`: HTTP方法
- `path`: 请求路径
- `query`: 查询参数
- `ip`: 客户端IP
- `status`: HTTP状态码
- `latency`: 请求耗时
- `user_agent`: 用户代理
- `request_id`: 请求ID（如果已设置）
- `request_body`: 请求体（仅非文件上传的POST/PUT/PATCH请求）
- `errors`: 错误信息（如果有）

**示例日志**：
```json
{
  "level": "info",
  "ts": 1234567890,
  "msg": "HTTP Request",
  "method": "POST",
  "path": "/api/upload",
  "ip": "127.0.0.1",
  "status": 200,
  "latency": "45.2ms",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. 性能监控中间件 (`PerformanceMonitor`)

**功能**：
- 统计每个接口的调用次数
- 记录接口的平均、最大、最小耗时
- 自动识别慢请求并记录
- 提供API查看性能统计

**配置**：
```yaml
middleware:
  performance_monitor:
    enabled: true              # 是否启用性能监控
    slow_threshold: "1s"        # 慢请求阈值（如：1s, 500ms）
    max_slow_requests: 100     # 最大慢请求记录数
```

**API端点**：
- `GET /api/metrics` - 获取性能统计信息
- `POST /api/metrics/reset` - 重置性能统计

**统计信息示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "routes": {
      "GET /api/history": {
        "count": 150,
        "avg_latency": "25.3ms",
        "max_latency": "120.5ms",
        "min_latency": "10.2ms",
        "total_latency": "3.795s"
      }
    },
    "slow_requests_count": 5,
    "slow_requests": [
      {
        "path": "/api/process/123",
        "method": "POST",
        "latency": "1.5s",
        "status": 200,
        "timestamp": "2024-01-01T12:00:00Z",
        "request_id": "550e8400-e29b-41d4-a716-446655440000"
      }
    ]
  }
}
```

### 3. 安全中间件 (`Security`)

**功能**：
- 限制请求体大小，防止DoS攻击
- 限制请求头大小
- 设置安全响应头（X-Content-Type-Options, X-Frame-Options, X-XSS-Protection）

**配置**：
```yaml
middleware:
  security:
    max_request_body_size: 10485760  # 最大请求体大小（字节），10MB
    max_header_size: 8192            # 最大请求头大小（字节），8KB
```

**安全响应头**：
- `X-Content-Type-Options: nosniff` - 防止MIME类型嗅探
- `X-Frame-Options: DENY` - 防止点击劫持
- `X-XSS-Protection: 1; mode=block` - XSS保护

## 中间件执行顺序

中间件的执行顺序很重要，当前顺序为：

1. **Recovery** - 捕获panic，防止服务崩溃
2. **RequestID** - 生成请求ID，用于追踪
3. **Security** - 安全检查（请求体大小等）
4. **PerformanceMonitor** - 性能监控（如果启用）
5. **RequestLogger** - 请求日志（如果启用）
6. **ErrorHandler** - 统一错误处理
7. **CORS** - 跨域处理

## 配置示例

完整配置示例：

```yaml
# 中间件配置
middleware:
  # 请求日志配置
  request_logger:
    enabled: true

  # 性能监控配置
  performance_monitor:
    enabled: true
    slow_threshold: "1s"        # 1秒以上的请求视为慢请求
    max_slow_requests: 100       # 最多记录100个慢请求

  # 安全配置
  security:
    max_request_body_size: 10485760  # 10MB
    max_header_size: 8192            # 8KB
```

## 使用建议

### 开发环境
- 启用请求日志，便于调试
- 启用性能监控，识别性能瓶颈
- 可以设置较小的慢请求阈值（如500ms）

### 生产环境
- 启用请求日志，但考虑日志轮转
- 启用性能监控，定期查看统计信息
- 设置合理的慢请求阈值（如1-2秒）
- 根据实际需求调整请求体大小限制

## 性能影响

- **请求日志中间件**：轻微性能影响（主要是I/O操作），建议在生产环境使用JSON格式日志
- **性能监控中间件**：几乎无性能影响（内存操作），可以放心启用
- **安全中间件**：几乎无性能影响（仅检查请求头）

## 注意事项

1. **请求体记录**：文件上传请求的请求体不会记录，避免日志过大
2. **慢请求记录**：慢请求列表有大小限制，超过限制会移除最旧的记录
3. **性能统计**：统计信息存储在内存中，服务重启后会重置
4. **请求体大小**：如果请求体超过限制，请求会被立即拒绝

## 未来扩展

根据 `docs/IMPROVEMENTS.md` 的建议，未来可以考虑：

1. **限流中间件**：API限流、IP限流、用户限流
2. **认证中间件**：JWT认证、API密钥认证
3. **审计日志中间件**：记录操作审计日志
4. **缓存中间件**：响应缓存、查询结果缓存
