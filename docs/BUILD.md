# 项目编译说明

## 环境要求

### 必需环境
- **Go 1.19+** - 后端开发语言
- **Node.js** (可选) - 仅用于前端资源管理，本项目前端为纯静态文件，无需构建

### 操作系统
- Windows 10/11
- Linux (Ubuntu 20.04+, CentOS 7+)
- macOS 10.15+

## 编译步骤

### 1. 克隆项目

```bash
git clone <repository-url>
cd redis_data
```

### 2. 安装依赖

项目使用 Go Modules 管理依赖，首次编译会自动下载依赖：

```bash
go mod download
```

### 3. 编译后端服务

#### Windows

```bash
# 编译到 cmd/web/web.exe
cd cmd/web
go build -o web.exe main.go

# 或者使用完整路径
go build -o ../../cmd/web/web.exe ./cmd/web/main.go
```

#### Linux/macOS

```bash
# 编译到 cmd/web/web
cd cmd/web
go build -o web main.go

# 或者使用完整路径
go build -o ../../cmd/web/web ./cmd/web/main.go
```

### 4. 配置文件

#### 创建 Web 配置文件

复制示例配置文件：

```bash
# Windows
copy config.web.example.json config.web.json

# Linux/macOS
cp config.web.example.json config.web.json
```

编辑 `config.web.json`：

```json
{
  "host": "0.0.0.0",
  "port": 8080,
  "resend_batch_size": 1000,
  "resend_batch_workers": 50,
  "resend_rank_workers": 32,
  "resend_http_timeout_sec": 30
}
```

#### 配置说明

- `host`: 服务器监听地址，`0.0.0.0` 表示监听所有网络接口
- `port`: 服务器端口号
- `resend_batch_size`: 批量补发时每批处理的玩家数量
- `resend_batch_workers`: 批量补发的并发工作协程数
- `resend_rank_workers`: 按排名补发的并发工作协程数
- `resend_http_timeout_sec`: HTTP 请求超时时间（秒）

### 5. 启动服务

#### Windows

```bash
# 使用启动脚本
start_web.bat

# 或直接运行
cmd\web\web.exe --config config.web.json
```

#### Linux/macOS

```bash
# 使用启动脚本
chmod +x start_web.sh
./start_web.sh

# 或直接运行
./cmd/web/web --config config.web.json
```

### 6. 验证服务

打开浏览器访问：`http://localhost:8080`

## 编译选项

### 交叉编译

#### 编译 Windows 版本（在 Linux/macOS 上）

```bash
GOOS=windows GOARCH=amd64 go build -o web.exe ./cmd/web/main.go
```

#### 编译 Linux 版本（在 Windows 上）

```bash
set GOOS=linux
set GOARCH=amd64
go build -o web ./cmd/web/main.go
```

#### 编译 macOS 版本（在 Linux/Windows 上）

```bash
GOOS=darwin GOARCH=amd64 go build -o web ./cmd/web/main.go
```

### 优化编译

#### 启用优化和减小体积

```bash
go build -ldflags="-s -w" -o web.exe ./cmd/web/main.go
```

- `-s`: 去除符号表
- `-w`: 去除调试信息

#### 静态链接

```bash
CGO_ENABLED=0 go build -a -ldflags="-s -w" -o web.exe ./cmd/web/main.go
```

## 依赖管理

### 主要依赖

项目依赖通过 `go.mod` 管理，主要依赖包括：

- `github.com/gin-gonic/gin` - Web 框架
- `github.com/xuri/excelize/v2` - Excel 文件处理
- `golang.org/x/net` - 网络工具包

### 更新依赖

```bash
# 更新所有依赖到最新版本
go get -u ./...

# 更新特定依赖
go get -u github.com/gin-gonic/gin

# 整理依赖
go mod tidy
```

### 查看依赖

```bash
# 查看所有依赖
go list -m all

# 查看依赖树
go mod graph
```

## 常见问题

### 1. 编译错误：找不到包

**问题**：`cannot find package "xxx"`

**解决**：
```bash
go mod download
go mod tidy
```

### 2. 编译错误：CGO 相关

**问题**：CGO 编译错误

**解决**：禁用 CGO（如果不需要）
```bash
CGO_ENABLED=0 go build ./cmd/web/main.go
```

### 3. 端口被占用

**问题**：`bind: address already in use`

**解决**：
- 修改 `config.web.json` 中的 `port` 配置
- 或关闭占用端口的进程

### 4. 配置文件不存在

**问题**：`config file not found`

**解决**：
- 确保 `config.web.json` 存在于项目根目录
- 或使用 `--config` 参数指定配置文件路径

## 开发环境设置

### IDE 推荐

- **VS Code** - 安装 Go 扩展
- **GoLand** - JetBrains 官方 IDE
- **Vim/Neovim** - 配合 vim-go 插件

### VS Code 扩展

- Go (golang.go)
- Go Test (可选)

### 代码格式化

```bash
# 格式化代码
go fmt ./...

# 代码检查
go vet ./...
```

## 构建脚本

### Windows 构建脚本 (build.bat)

```batch
@echo off
echo 编译 Redis 数据解析工具...
cd cmd\web
go build -ldflags="-s -w" -o web.exe main.go
if %errorlevel% == 0 (
    echo 编译成功！
    move web.exe ..\..\web.exe
) else (
    echo 编译失败！
    pause
)
```

### Linux/macOS 构建脚本 (build.sh)

```bash
#!/bin/bash
echo "编译 Redis 数据解析工具..."
cd cmd/web
go build -ldflags="-s -w" -o web main.go
if [ $? -eq 0 ]; then
    echo "编译成功！"
    mv web ../../web
else
    echo "编译失败！"
    exit 1
fi
```

## 版本信息

查看编译后的版本信息：

```bash
# 添加版本信息编译
go build -ldflags="-X main.Version=1.0.0 -X main.BuildTime=$(date +%Y-%m-%d\ %H:%M:%S)" -o web.exe ./cmd/web/main.go
```

## 部署建议

### 生产环境

1. **使用反向代理**（Nginx/Apache）
2. **配置 HTTPS**
3. **设置防火墙规则**
4. **使用进程管理工具**（systemd/supervisor）
5. **配置日志轮转**

### Docker 部署（可选）

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o web ./cmd/web/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/web .
COPY --from=builder /app/config.web.json .
EXPOSE 8080
CMD ["./web", "--config", "config.web.json"]
```
