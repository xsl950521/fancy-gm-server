# 后端技术文档

## 技术栈

### 核心框架
- **Go 1.19+** - 编程语言
- **Gin** - Web 框架
- **excelize/v2** - Excel 文件处理

### 主要依赖
```go
github.com/gin-gonic/gin          // Web 框架
github.com/xuri/excelize/v2       // Excel 处理
golang.org/x/net                  // 网络工具
```

## 项目结构

```
.
├── cmd/
│   └── web/
│       └── main.go              # 程序入口
├── api/
│   ├── handlers/                # 请求处理器
│   │   ├── upload.go           # 文件上传
│   │   ├── process.go          # 数据处理
│   │   ├── data.go             # 数据查询
│   │   ├── history.go          # 历史记录
│   │   ├── analytics.go        # 数据分析
│   │   └── missinglist.go      # 漏发名单
│   ├── middleware/             # 中间件
│   │   └── cors.go             # CORS 处理
│   └── models/                 # 数据模型
│       ├── job.go              # 任务模型
│       └── job_manager.go      # 任务管理
├── service/                     # 业务逻辑
│   ├── parser/                 # 数据解析
│   ├── history/                # 历史记录
│   ├── analytics/              # 数据分析
│   └── missinglist/            # 漏发名单
│       ├── missinglist.go      # 名单拉取
│       ├── resend.go           # 补发逻辑
│       ├── lua_parser.go       # Lua 配置解析
│       ├── config_loader.go    # 配置加载
│       └── item_mapper.go      # 道具映射
├── config/                     # 配置管理
│   ├── config.go              # 主配置
│   └── web_config.go          # Web 配置
├── dailyrank/                  # 日榜处理
├── totalrank/                  # 总榜处理
├── monthrank/                  # 月榜处理
├── clubpid/                    # 公会 PID 处理
└── storage/                    # 存储目录
    ├── uploads/               # 上传文件
    ├── results/               # 处理结果
    └── history/               # 历史记录
```

## 核心模块

### 1. Web 服务器 (cmd/web/main.go)

**功能**：
- 初始化 Gin 路由
- 配置静态文件服务
- 注册 API 路由
- 启动 HTTP 服务器

**关键代码**：
```go
func main() {
    // 加载配置
    webCfg, err := config.LoadWebConfig(*configPath)
    
    // 创建路由
    r := gin.Default()
    r.Use(middleware.CORS())
    
    // 静态文件
    r.Static("/static", "./web/static")
    
    // API 路由
    api := r.Group("/api")
    api.POST("/upload", uploadHandler.Upload)
    api.POST("/process", processHandler.Process)
    // ...
    
    // 启动服务器
    r.Run(fmt.Sprintf("%s:%d", webCfg.Host, webCfg.Port))
}
```

### 2. 文件上传 (api/handlers/upload.go)

**功能**：
- 接收文件上传请求
- 保存文件到临时目录
- 创建处理任务

**处理流程**：
1. 接收 multipart/form-data 请求
2. 保存文件到 `storage/uploads/`
3. 创建 Job 对象
4. 返回 Job ID

### 3. 数据处理 (api/handlers/process.go)

**功能**：
- 启动异步处理任务
- 返回任务状态
- 保存处理结果

**处理流程**：
1. 根据模式选择解析器
2. 异步处理文件
3. 保存结果到 Excel
4. 更新任务状态
5. 保存到历史记录

### 4. 数据解析 (service/parser/parser.go)

**解析器接口**：
```go
type Parser interface {
    Parse(filePath string) ([]SheetResult, error)
}
```

**支持的解析器**：
- `DailyRankParser` - 日榜数据
- `TotalRankParser` - 总榜数据
- `MonthRankParser` - 月榜数据
- `ClubPidParser` - 公会 PID 数据
- `RewardRecordParser` - 奖励记录
- `MailLogParser` - 邮件日志

### 5. 漏发名单 (service/missinglist/)

#### 5.1 名单拉取 (missinglist.go)

**功能**：
- 从外部 API 拉取漏发名单
- 数据过滤和分组
- 时间戳转换

**API 调用**：
```go
func FetchMissingList(url string, iid string) ([]MissingListItem, error) {
    resp, err := http.Get(url)
    // 解析响应
    // 转换数据格式
}
```

#### 5.2 补发逻辑 (resend.go)

**功能**：
- 解析 Excel 文件获取玩家信息
- 匹配奖励配置
- 发送补发请求

**两种补发模式**：

1. **批量补发 (ProcessResendV3)**
   - 使用 `send_mail_multi` API
   - 单次请求发送多个玩家
   - 支持分批处理

2. **分组补发 (ProcessResendV2)**
   - 使用 `send_mail` API
   - 按排名分组发送
   - 支持并发处理

**并发控制**：
```go
// 工作协程池
workerCount := batchWorkers
rankCh := make(chan int)
resultCh := make(chan []ResendResult, len(ranks))

// 启动工作协程
for i := 0; i < workerCount; i++ {
    go func() {
        for rank := range rankCh {
            resultCh <- processRankResend(...)
        }
    }()
}
```

#### 5.3 Lua 配置解析 (lua_parser.go)

**功能**：
- 解析 Lua 配置文件
- 提取标题、内容、奖励配置
- 支持多种游戏类型（bydr, dsc, cqsj, yxds）

**解析流程**：
1. 根据游戏类型选择解析器
2. 使用正则表达式提取配置
3. 解析嵌套的 Lua 表结构
4. 转换为 JSON 格式

**示例**：
```go
config, err := ParseLuaConfig(luaContent, "bydr")
// config.Titles["daily_person_super"]
// config.Contents["daily_person_super"]
// config.Rewards["daily_person_super"]
```

#### 5.4 配置加载 (config_loader.go)

**功能**：
- 从文件系统加载默认配置
- 支持项目内路径和外部路径
- 自动选择最新日期的配置文件

**文件命名规则**：
- `{iid}YYYYMMDD.lua`
- 例如：`bydr20260225.lua`

#### 5.5 道具映射 (item_mapper.go)

**功能**：
- 从 Excel 文件加载道具 ID 到名称的映射
- 提供线程安全的访问接口
- 使用 sync.Once 确保只加载一次

### 6. 数据分析 (service/analytics/analytics.go)

**功能**：
- 分析 Excel 数据
- 生成统计信息
- 排行榜专项分析

**分析维度**：
- 总体统计（记录数、列统计）
- 数据分布（直方图）
- 时间序列（趋势图）
- 排行榜分析（按组、按日、排名分布、分数分布、Top 玩家）

### 7. 历史记录 (service/history/history.go)

**功能**：
- 保存处理任务到 JSON 文件
- 查询历史记录
- 删除历史记录
- 自动清理旧记录

**存储格式**：
```json
{
  "id": "job-123",
  "status": "completed",
  "mode": "daily",
  "createdAt": "2026-02-04T10:00:00Z",
  "results": [...]
}
```

## API 接口

### 文件上传

```
POST /api/upload
Content-Type: multipart/form-data

参数：
- files: 文件列表
- mode: 处理模式 (daily/total/month/clubpid/...)
- archive: 是否归档 (true/false)
```

### 数据处理

```
POST /api/process
Content-Type: application/json

{
  "jobId": "job-123"
}
```

### 查询数据

```
GET /api/data/:jobId
GET /api/data/:jobId/sheet/:sheetName
```

### 下载文件

```
GET /api/download/:jobId
GET /api/download/:jobId/:sheetName?format=csv
```

### 历史记录

```
GET /api/history              # 获取所有历史记录
GET /api/history/:id          # 获取单个历史记录
DELETE /api/history/:id       # 删除历史记录
```

### 数据分析

```
GET /api/analytics/statistics?jobId=xxx&sheetName=xxx
GET /api/analytics/sheets?jobId=xxx
```

### 漏发名单

```
POST /api/missinglist/fetch              # 拉取漏发名单
POST /api/missinglist/filter             # 筛选数据
POST /api/missinglist/group              # 分组数据
POST /api/missinglist/parse-excel        # 解析 Excel
POST /api/missinglist/resend             # 发送补发
POST /api/missinglist/parse-lua          # 解析 Lua 配置
POST /api/missinglist/load-default-config # 加载默认配置
GET /api/missinglist/item-map            # 获取道具映射
```

## 配置管理

### Web 配置 (config/web_config.go)

**配置项**：
```go
type WebConfig struct {
    Host                 string `json:"host"`
    Port                 int    `json:"port"`
    ResendBatchSize      int    `json:"resend_batch_size"`
    ResendBatchWorkers   int    `json:"resend_batch_workers"`
    ResendRankWorkers    int    `json:"resend_rank_workers"`
    ResendHTTPTimeoutSec int    `json:"resend_http_timeout_sec"`
}
```

**加载顺序**：
1. 默认值
2. 配置文件 (`config.web.json`)
3. 环境变量（覆盖配置文件）

**环境变量**：
- `WEB_HOST`
- `WEB_PORT`
- `RESEND_BATCH_SIZE`
- `RESEND_BATCH_WORKERS`
- `RESEND_RANK_WORKERS`
- `RESEND_HTTP_TIMEOUT_SEC`

## 并发处理

### 任务处理

使用 Goroutine 异步处理任务：

```go
func (h *ProcessHandler) processJobAsync(job *models.Job) {
    go func() {
        // 处理逻辑
        job.SetStatus(models.JobStatusProcessing)
        // ...
        job.SetStatus(models.JobStatusCompleted)
    }()
}
```

### 补发并发

使用工作协程池处理批量补发：

```go
func processRanksConcurrently(...) {
    workerCount := rankWorkers
    rankCh := make(chan int)
    resultCh := make(chan []ResendResult, len(ranks))
    
    // 启动工作协程
    for i := 0; i < workerCount; i++ {
        go func() {
            for rank := range rankCh {
                resultCh <- processRankResend(...)
            }
        }()
    }
    
    // 收集结果
    var wg sync.WaitGroup
    // ...
}
```

## 错误处理

### 统一错误响应

```go
c.JSON(http.StatusBadRequest, gin.H{
    "success": false,
    "error": "错误信息",
})
```

### 错误日志

```go
log.Printf("错误: %v", err)
log.Printf("错误详情: %+v", err)
```

## 性能优化

### 1. HTTP 连接池

```go
httpTransport = &http.Transport{
    MaxIdleConns:        200,
    MaxIdleConnsPerHost: 100,
    IdleConnTimeout:     90 * time.Second,
    DisableKeepAlives:   false,
}
```

### 2. 字符串构建优化

使用 `strings.Builder` 减少内存分配：

```go
var builder strings.Builder
builder.WriteString("prefix")
builder.WriteString(strconv.Itoa(value))
result := builder.String()
```

### 3. 预分配切片

```go
results := make([]ResendResult, 0, len(players))
```

### 4. 批量处理

```go
const maxBatchSize = 1000
for i := 0; i < len(items); i += maxBatchSize {
    end := i + maxBatchSize
    if end > len(items) {
        end = len(items)
    }
    batch := items[i:end]
    processBatch(batch)
}
```

## 安全考虑

### 1. CORS 配置

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        // ...
    }
}
```

### 2. 文件上传验证

- 检查文件类型
- 限制文件大小
- 验证文件路径

### 3. 输入验证

```go
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

## 测试

### 单元测试

```go
func TestParseMonthRankKey(t *testing.T) {
    key := "hd_month_rank_rewards_820540800_2601132302_hd_month_stage_rank_1"
    info, err := monthrank.ParseMonthRankKey(key)
    assert.NoError(t, err)
    assert.Equal(t, 2601132302, info.GroupId)
}
```

### 集成测试

```go
func TestResendAPI(t *testing.T) {
    router := setupRouter()
    req := httptest.NewRequest("POST", "/api/missinglist/resend", body)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    assert.Equal(t, 200, w.Code)
}
```

## 部署

### 生产环境建议

1. **使用反向代理**（Nginx）
2. **配置 HTTPS**
3. **设置日志轮转**
4. **使用进程管理**（systemd/supervisor）
5. **监控和告警**

### 日志管理

```go
// 文件日志
logFile, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
log.SetOutput(logFile)

// 日志轮转（使用第三方库）
```

## 扩展开发

### 添加新的数据解析器

1. 在 `service/parser/` 创建新的解析器
2. 实现 `Parser` 接口
3. 在 `service/parser/parser.go` 注册
4. 在 `api/handlers/process.go` 添加处理逻辑

### 添加新的 API 端点

1. 在 `api/handlers/` 创建处理器
2. 在 `cmd/web/main.go` 注册路由
3. 添加相应的业务逻辑

### 添加新的配置项

1. 在 `config/web_config.go` 添加字段
2. 更新 `config.web.example.json`
3. 添加环境变量支持（如需要）
