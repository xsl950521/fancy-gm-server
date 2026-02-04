# 配置中心化文档

## 概述

项目使用Viper实现配置中心化，支持多种配置格式（YAML、JSON、TOML）和环境变量覆盖。

## 配置文件格式

### 支持格式
- **YAML** (推荐) - `config.yaml`
- **JSON** - `config.json`
- **TOML** - `config.toml`

### 配置文件位置

配置文件按以下顺序查找：
1. 命令行参数指定的路径：`--config /path/to/config.yaml`
2. 当前目录：`./config.yaml`
3. config目录：`./config/config.yaml`
4. 系统配置目录：`/etc/gm-server/config.yaml`

### 配置文件示例

#### YAML格式 (config.yaml)

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080

# 数据库配置
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "gm_server"
  sslmode: "disable"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  conn_max_idle_time: 300

# 日志配置
log:
  level: "info"        # debug, info, warn, error
  encoding: "json"      # json, console
  output_path: ""       # 空则输出到stdout
  error_path: ""         # 空则输出到stderr

# Web功能配置
web:
  host: "0.0.0.0"
  port: 8080
  resend_batch_size: 1000
  resend_batch_workers: 50
  resend_rank_workers: 32
  resend_http_timeout_sec: 30
```

#### JSON格式 (config.json)

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "postgres",
    "dbname": "gm_server",
    "sslmode": "disable",
    "max_open_conns": 100,
    "max_idle_conns": 10
  },
  "log": {
    "level": "info",
    "encoding": "json"
  },
  "web": {
    "host": "0.0.0.0",
    "port": 8080,
    "resend_batch_size": 1000,
    "resend_batch_workers": 50,
    "resend_rank_workers": 32,
    "resend_http_timeout_sec": 30
  }
}
```

## 环境变量配置

环境变量优先级高于配置文件，可以覆盖配置文件中的值。

### 环境变量命名规则

支持两种格式：
1. **GM_前缀格式**（推荐）：`GM_SERVER_HOST`、`GM_DB_HOST`
2. **直接格式**（兼容旧代码）：`DB_HOST`、`LOG_LEVEL`

### 环境变量列表

#### 服务器配置
- `GM_SERVER_HOST` 或 `SERVER_HOST` - 服务器监听地址
- `GM_SERVER_PORT` 或 `SERVER_PORT` 或 `PORT` - 服务器端口

#### 数据库配置
- `GM_DB_HOST` 或 `DB_HOST` - 数据库主机
- `GM_DB_PORT` 或 `DB_PORT` - 数据库端口
- `GM_DB_USER` 或 `DB_USER` - 数据库用户
- `GM_DB_PASSWORD` 或 `DB_PASSWORD` - 数据库密码
- `GM_DB_NAME` 或 `DB_NAME` - 数据库名称
- `GM_DB_SSLMODE` 或 `DB_SSLMODE` - SSL模式

#### 日志配置
- `GM_LOG_LEVEL` 或 `LOG_LEVEL` - 日志级别（debug/info/warn/error）
- `GM_LOG_ENCODING` 或 `LOG_ENCODING` - 日志格式（json/console）
- `GM_LOG_OUTPUT_PATH` 或 `LOG_OUTPUT_PATH` - 日志输出路径
- `GM_LOG_ERROR_PATH` 或 `LOG_ERROR_PATH` - 错误日志路径

#### Web配置
- `GM_WEB_HOST` 或 `WEB_HOST` - Web服务监听地址
- `GM_WEB_PORT` 或 `WEB_PORT` - Web服务监听端口
- `GM_RESEND_BATCH_SIZE` 或 `RESEND_BATCH_SIZE` - 批量补发大小
- `GM_RESEND_BATCH_WORKERS` 或 `RESEND_BATCH_WORKERS` - 批量补发工作协程数
- `GM_RESEND_RANK_WORKERS` 或 `RESEND_RANK_WORKERS` - 排名补发工作协程数
- `GM_RESEND_HTTP_TIMEOUT_SEC` 或 `RESEND_HTTP_TIMEOUT_SEC` - HTTP超时时间

## 使用方式

### 1. 使用配置文件

```bash
# 指定配置文件路径
./web --config config.yaml

# 或使用默认路径（自动查找）
./web
```

### 2. 使用环境变量

```bash
# Linux/macOS
export GM_DB_HOST=localhost
export GM_DB_PORT=5432
export GM_DB_USER=postgres
export GM_DB_PASSWORD=postgres
export GM_DB_NAME=gm_server
./web

# Windows
set GM_DB_HOST=localhost
set GM_DB_PORT=5432
set GM_DB_USER=postgres
set GM_DB_PASSWORD=postgres
set GM_DB_NAME=gm_server
web.exe
```

### 3. 混合使用

配置文件提供基础配置，环境变量覆盖特定值：

```bash
# config.yaml 中配置了基础设置
# 使用环境变量覆盖数据库密码
export GM_DB_PASSWORD=secret_password
./web --config config.yaml
```

## 配置优先级

配置优先级从高到低：
1. **环境变量** - 最高优先级
2. **命令行参数** - `--config` 指定的配置文件
3. **默认配置文件** - 自动查找的配置文件
4. **代码默认值** - 最低优先级

## 配置验证

配置加载时会自动验证：
- 必填字段检查
- 数值范围验证
- 格式验证

如果配置无效，服务启动会失败并显示错误信息。

## 配置热重载（可选）

配置支持热重载功能（需要手动启用）：

```go
config.WatchAppConfig(func(cfg *config.AppConfig) {
    // 配置文件变化时的回调
    // 可以在这里更新服务配置
})
```

## 最佳实践

1. **生产环境**：使用配置文件 + 环境变量（敏感信息如密码）
2. **开发环境**：使用配置文件
3. **Docker部署**：使用环境变量
4. **敏感信息**：永远不要写入配置文件，使用环境变量

## 示例

### 开发环境配置

创建 `config.dev.yaml`：

```yaml
server:
  host: "localhost"
  port: 8080

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "dev_password"
  dbname: "gm_server_dev"

log:
  level: "debug"
  encoding: "console"
```

启动：
```bash
./web --config config.dev.yaml
```

### 生产环境配置

创建 `config.prod.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "db.example.com"
  port: 5432
  user: "gm_user"
  dbname: "gm_server"
  # 密码通过环境变量设置

log:
  level: "info"
  encoding: "json"
  output_path: "/var/log/gm-server/app.log"
  error_path: "/var/log/gm-server/error.log"
```

启动：
```bash
export GM_DB_PASSWORD=production_password
./web --config config.prod.yaml
```

## 故障排查

### 问题1: 配置文件未找到

**现象**: 服务使用默认配置启动

**解决**: 
- 检查配置文件路径
- 使用 `--config` 参数指定完整路径
- 或确保配置文件在默认查找路径中

### 问题2: 环境变量未生效

**现象**: 环境变量设置后配置未改变

**解决**:
- 检查环境变量名称是否正确
- 确保环境变量在服务启动前设置
- 检查是否有配置文件覆盖了环境变量

### 问题3: 配置格式错误

**现象**: 服务启动失败，提示配置解析错误

**解决**:
- 检查YAML/JSON格式是否正确
- 使用在线YAML/JSON验证工具
- 参考示例配置文件
