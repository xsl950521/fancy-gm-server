# Makefile 使用文档

本项目提供了 Makefile 用于简化构建、测试和部署流程。

## 前置要求

### Linux/macOS
- 已安装 `make` 工具（通常系统自带）
- Go 1.19+

### Windows
- 已安装 `make` 工具（推荐使用 [GnuWin32](http://gnuwin32.sourceforge.net/) 或 [Chocolatey](https://chocolatey.org/) 安装）
- 或使用 `Makefile.windows`（需要手动设置环境变量）
- Go 1.19+

## 快速开始

### 查看帮助

```bash
make help
```

或

```bash
make
```

## 常用命令

### 开发环境

```bash
# 运行开发服务器（使用默认配置）
make run

# 运行开发服务器（使用YAML配置）
make run-config

# 开发模式（自动重载，需要安装air）
make dev
```

### 构建

```bash
# 构建当前平台
make build

# 构建特定平台
make build-windows   # Windows版本
make build-linux     # Linux版本
make build-macos     # macOS版本
make build-macos-arm # macOS ARM64版本

# 构建所有平台
make build-all
```

### 测试

```bash
# 运行测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

### 依赖管理

```bash
# 下载依赖
make deps

# 更新依赖
make deps-update

# 整理依赖
make deps-tidy

# 创建vendor目录
make deps-vendor
```

### 代码质量

```bash
# 格式化代码
make fmt

# 运行代码检查
make lint

# 运行go vet
make vet
```

### 清理

```bash
# 清理构建文件
make clean

# 清理所有生成文件（包括缓存）
make clean-all
```

## 环境配置

### 创建配置文件

```bash
# 创建开发环境配置
make config-dev

# 创建生产环境配置
make config-prod
```

### 数据库初始化

```bash
# Linux/macOS
make db-init

# Windows
make db-init-windows
```

## 快速命令

```bash
# 快速构建（整理依赖+构建）
make quick-build

# 快速测试（整理依赖+测试）
make quick-test

# 快速运行（整理依赖+运行）
make quick-run
```

## 交叉编译

### 从 Linux/macOS 编译 Windows

```bash
make build-windows
```

### 从 Windows 编译 Linux

```bash
make build-linux
```

### 从 Linux 编译 macOS

```bash
make build-macos
make build-macos-arm  # ARM64版本（Apple Silicon）
```

## 构建输出

所有构建文件输出到 `build/` 目录：

```
build/
├── gm-server              # 当前平台
├── windows/
│   └── gm-server.exe
├── linux/
│   └── gm-server
├── macos/
│   └── gm-server
└── macos-arm/
    └── gm-server
```

## 版本信息

构建的可执行文件包含版本信息：

```bash
# 查看版本信息（如果实现了version命令）
./build/gm-server --version
```

版本信息从 Git 标签自动获取，如果没有标签则使用 "dev"。

## Windows 使用说明

### 方法1: 使用 Make（推荐）

如果已安装 Make：

```cmd
make build
make run-config
```

### 方法2: 使用 Makefile.windows

```cmd
# 需要手动设置环境变量
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-X main.Version=dev -s -w" -o build\gm-server.exe cmd\web\main.go
```

### 方法3: 直接使用 Go 命令

```cmd
# 构建
go build -o build\gm-server.exe cmd\web\main.go

# 运行
go run cmd\web\main.go --config config.yaml
```

## 高级用法

### 自定义构建参数

编辑 Makefile 中的变量：

```makefile
# 修改应用名称
APP_NAME := my-app

# 修改构建目录
BUILD_DIR := dist

# 添加额外的编译参数
BUILD_FLAGS := -ldflags "$(LDFLAGS)" -tags "production"
```

### 并行构建

```bash
# 使用 -j 参数并行构建多个平台
make -j4 build-all
```

### 静默构建

```bash
# 不显示命令，只显示结果
make -s build
```

## 故障排查

### 问题1: make: command not found

**解决方案**:
- Linux/macOS: 通常已安装，如果没有：`sudo apt-get install make` (Ubuntu) 或 `brew install make` (macOS)
- Windows: 安装 [GnuWin32](http://gnuwin32.sourceforge.net/) 或使用 Chocolatey: `choco install make`

### 问题2: 交叉编译失败

**解决方案**:
- 确保设置了正确的 `GOOS` 和 `GOARCH`
- 检查是否有 CGO 依赖（可能需要设置 `CGO_ENABLED=0`）

### 问题3: 版本信息显示不正确

**解决方案**:
- 确保项目在 Git 仓库中
- 或手动设置版本：`VERSION=1.0.0 make build`

## 集成到 CI/CD

### GitHub Actions 示例

```yaml
- name: Build
  run: make build-all

- name: Test
  run: make test-coverage
```

### GitLab CI 示例

```yaml
build:
  script:
    - make build-all
  artifacts:
    paths:
      - build/
```

## 最佳实践

1. **开发时**: 使用 `make run-config` 快速启动
2. **测试前**: 运行 `make quick-test` 确保代码质量
3. **发布前**: 运行 `make build-all` 构建所有平台
4. **部署前**: 运行 `make test-coverage` 检查测试覆盖率

## 扩展 Makefile

可以添加自定义任务：

```makefile
.PHONY: deploy
deploy: build ## 部署到服务器
	@echo "部署中..."
	@scp build/gm-server user@server:/opt/gm-server/
	@ssh user@server "systemctl restart gm-server"
```

然后在文档中说明如何使用。
