# 快速启动指南

## 使用YAML配置启动服务

### Windows系统

#### 方法1: 使用专用启动脚本（推荐）

```cmd
start_web_yaml.bat
```

#### 方法2: 使用通用启动脚本

```cmd
start_web.bat
```

脚本会自动查找配置文件（优先级：config.yaml > config.yml > config.json > config.web.json）

#### 方法3: 手动启动

```cmd
# 使用YAML配置文件
go run cmd/web/main.go --config config.yaml

# 或使用默认配置（自动查找）
go run cmd/web/main.go
```

### Linux/macOS系统

#### 方法1: 使用专用启动脚本（推荐）

```bash
chmod +x start_web_yaml.sh
./start_web_yaml.sh
```

#### 方法2: 使用通用启动脚本

```bash
chmod +x start_web.sh
./start_web.sh
```

#### 方法3: 手动启动

```bash
# 使用YAML配置文件
go run cmd/web/main.go --config config.yaml

# 或使用默认配置（自动查找）
go run cmd/web/main.go
```

## 配置文件准备

### 1. 创建配置文件

**Windows:**
```cmd
copy config\config.example.yaml config.yaml
notepad config.yaml
```

**Linux/macOS:**
```bash
cp config/config.example.yaml config.yaml
vim config.yaml
```

### 2. 编辑配置

主要需要修改的配置项：

```yaml
database:
  host: "localhost"        # 数据库主机
  port: 5432              # 数据库端口
  user: "postgres"        # 数据库用户
  password: "postgres"    # 数据库密码（建议使用环境变量）
  dbname: "gm_server"     # 数据库名称

log:
  level: "info"           # 开发环境可设置为 "debug"
  encoding: "console"     # 开发环境建议用 "console"，生产环境用 "json"
```

### 3. 使用环境变量（推荐用于敏感信息）

```cmd
REM Windows
set GM_DB_PASSWORD=your_password
start_web_yaml.bat
```

```bash
# Linux/macOS
export GM_DB_PASSWORD=your_password
./start_web_yaml.sh
```

## 启动前检查

### 1. 检查数据库

确保PostgreSQL服务已启动，并且数据库已创建：

```cmd
REM Windows
scripts\init_db.bat
```

```bash
# Linux/macOS
./scripts/init_db.sh
```

### 2. 检查配置文件

```cmd
REM Windows - 检查配置文件是否存在
dir config.yaml
```

```bash
# Linux/macOS - 检查配置文件是否存在
ls -la config.yaml
```

### 3. 检查Go环境

```bash
go version
```

## 启动服务

### 开发环境启动

```cmd
REM Windows
start_web_yaml.bat
```

```bash
# Linux/macOS
./start_web_yaml.sh
```

### 生产环境启动

```bash
# 编译
go build -o gm-server cmd/web/main.go

# 启动
./gm-server --config config.yaml
```

## 验证服务

服务启动后，访问：

- Web界面: http://localhost:8080
- API文档: http://localhost:8080/api

## 常见问题

### 问题1: 配置文件未找到

**现象**: 提示 "Failed to load config"

**解决**: 
- 确保配置文件在项目根目录
- 或使用 `--config` 参数指定完整路径

### 问题2: 数据库连接失败

**现象**: "Failed to connect database"

**解决**:
1. 检查PostgreSQL服务是否运行
2. 检查配置文件中的数据库连接信息
3. 使用环境变量设置密码：`set GM_DB_PASSWORD=your_password`

### 问题3: 端口被占用

**现象**: "bind: address already in use"

**解决**:
1. 修改配置文件中的端口号
2. 或关闭占用端口的进程

### 问题4: 中文乱码（Windows）

**现象**: 控制台中文显示乱码

**解决**:
- 启动脚本已自动设置UTF-8编码
- 如果仍有问题，检查控制台字体设置

## 停止服务

在运行服务的控制台中按 `Ctrl+C` 停止服务。

## 日志查看

### 开发环境（console格式）

日志直接输出到控制台。

### 生产环境（json格式）

如果配置了日志文件路径：

```yaml
log:
  output_path: "logs/app.log"
  error_path: "logs/error.log"
```

查看日志：

```bash
# Linux/macOS
tail -f logs/app.log

# Windows
type logs\app.log
```

## 下一步

- 查看 [配置文档](CONFIG.md) 了解详细配置选项
- 查看 [编译说明](BUILD.md) 了解如何编译和部署
- 查看 [改进建议](IMPROVEMENTS.md) 了解功能增强方向
