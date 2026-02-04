# Fancy GM Server

游戏管理工具后端服务

## 技术栈

- Go 1.19+
- Gin Web 框架
- excelize/v2 Excel 处理

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置文件

复制示例配置文件：

```bash
cp config.web.example.json config.web.json
```

### 3. 启动服务

```bash
# Windows
start_web.bat

# Linux/macOS
./start_web.sh
```

## 详细文档

- [编译说明](docs/BUILD.md)
- [后端技术文档](docs/BACKEND.md)
- [API接口文档](docs/API.md) - **客户端对接必读**
- [OpenAPI规范](docs/API_OPENAPI.yaml) - Swagger/OpenAPI格式
- [Makefile使用](docs/MAKEFILE.md)
- [配置文档](docs/CONFIG.md)
- [快速启动](docs/QUICKSTART.md)
- [中间件增强](docs/MIDDLEWARE.md) - 请求日志、性能监控、安全中间件
- [认证与授权](docs/AUTH.md) - JWT认证、RBAC权限管理

## API 文档

服务启动后访问：`http://localhost:8080`

**完整的API文档请查看：**
- [API接口文档](docs/API.md) - 详细的接口说明和参数
- [OpenAPI规范](docs/API_OPENAPI.yaml) - Swagger/OpenAPI格式
- [客户端对接示例](docs/API_CLIENT_EXAMPLES.md) - 各种编程语言的示例代码

主要 API 端点：
- `POST /api/upload` - 文件上传
- `POST /api/process/:jobId` - 数据处理
- `GET /api/status/:jobId` - 查询任务状态
- `GET /api/data` - 查询数据（分页）
- `GET /api/download/:jobId` - 下载文件
- `GET /api/history` - 历史记录
- `GET /api/analytics/statistics` - 数据分析
- `POST /api/missinglist/fetch` - 漏发名单拉取
- `POST /api/missinglist/resend` - 奖励补发

## License

MIT
