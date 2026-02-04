# 数据库初始化脚本

本目录包含数据库一键初始化脚本，用于快速创建和初始化数据库。

## 脚本说明

### 1. init_db.sh (Linux/macOS)

Bash脚本，适用于Linux和macOS系统。

**使用方法：**

```bash
# 使用默认配置
chmod +x scripts/init_db.sh
./scripts/init_db.sh

# 使用自定义配置
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_NAME=gm_server \
./scripts/init_db.sh
```

**环境变量：**
- `DB_HOST`: 数据库主机（默认: localhost）
- `DB_PORT`: 数据库端口（默认: 5432）
- `DB_USER`: 数据库用户（默认: postgres）
- `DB_NAME`: 数据库名称（默认: gm_server）
- `PGPASSWORD`: PostgreSQL管理员密码

### 2. init_db.bat (Windows)

批处理脚本，适用于Windows系统。已设置UTF-8代码页以支持中文显示。

**使用方法：**

```cmd
REM 使用默认配置
scripts\init_db.bat

REM 使用自定义配置
set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=postgres
set DB_NAME=gm_server
scripts\init_db.bat
```

**注意：**
- 如果中文显示乱码，请确保：
  1. 脚本文件以UTF-8编码保存
  2. 控制台支持UTF-8（脚本会自动设置代码页65001）
  3. 如果仍有问题，可以尝试使用 `init_db_utf8.bat` 版本

**环境变量：**
- `DB_HOST`: 数据库主机（默认: localhost）
- `DB_PORT`: 数据库端口（默认: 5432）
- `DB_USER`: 数据库用户（默认: postgres）
- `DB_NAME`: 数据库名称（默认: gm_server）
- `PGPASSWORD`: PostgreSQL管理员密码

### 3. init_db.sql (通用SQL脚本)

纯SQL脚本，可以在任何支持PostgreSQL的环境中执行。

**使用方法：**

```bash
# 方法1: 使用psql执行
psql -U postgres -f scripts/init_db.sql

# 方法2: 直接执行SQL
psql -U postgres -c "CREATE DATABASE gm_server;"
psql -U postgres -d gm_server -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
```

## 注意事项

1. **权限要求**: 执行脚本需要PostgreSQL管理员权限（通常是postgres用户）
2. **数据库存在**: 如果数据库已存在，脚本会提示是否删除重建
3. **表结构迁移**: 脚本只创建数据库，表结构会在服务启动时自动创建（通过GORM迁移）
4. **密码输入**: 如果未设置`PGPASSWORD`环境变量，脚本会提示输入密码

## 快速开始

### Linux/macOS

```bash
# 1. 给脚本添加执行权限
chmod +x scripts/init_db.sh

# 2. 执行初始化
./scripts/init_db.sh

# 3. 输入PostgreSQL管理员密码
```

### Windows

```cmd
REM 1. 打开命令提示符或PowerShell
REM 2. 切换到项目目录
cd E:\xsl_code\fancy-gm-server

REM 3. 执行初始化脚本
scripts\init_db.bat

REM 4. 输入PostgreSQL管理员密码
```

### 使用SQL脚本

```bash
# 使用psql执行
psql -U postgres -f scripts/init_db.sql
```

## 验证

初始化完成后，可以验证数据库是否创建成功：

```bash
# 列出所有数据库
psql -U postgres -l

# 连接到数据库
psql -U postgres -d gm_server

# 查看扩展
\dx
```

## 故障排查

### 问题0: 中文显示乱码（Windows）

**现象**: 执行init_db.bat时中文显示为乱码

**解决方案**:
1. 脚本已自动设置UTF-8代码页（chcp 65001）
2. 如果仍有问题，检查控制台字体是否支持中文
3. 可以尝试在PowerShell中执行：`[Console]::OutputEncoding = [System.Text.Encoding]::UTF8`
4. 或使用 `init_db_utf8.bat` 版本

### 问题1: 找不到psql命令

**解决方案**: 确保PostgreSQL客户端已安装并添加到PATH环境变量中。

- Linux: `sudo apt-get install postgresql-client` (Ubuntu/Debian)
- macOS: `brew install postgresql`
- Windows: 安装PostgreSQL时选择安装命令行工具

### 问题2: 连接被拒绝

**解决方案**: 
1. 检查PostgreSQL服务是否运行
2. 检查防火墙设置
3. 检查`pg_hba.conf`配置

### 问题3: 权限不足

**解决方案**: 使用具有创建数据库权限的用户（通常是postgres用户）

```bash
# 切换到postgres用户（Linux）
sudo -u postgres psql

# 或使用-U参数指定用户
psql -U postgres
```
