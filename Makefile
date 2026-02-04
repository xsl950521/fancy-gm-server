# GM Server Makefile
# 支持多环境编译和部署

# 变量定义
APP_NAME := gm-server
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%d\ %H:%M:%S)
BUILD_DIR := build
CMD_DIR := cmd/web
MAIN_FILE := $(CMD_DIR)/main.go

# Go 编译参数
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -s -w
BUILD_FLAGS := -ldflags "$(LDFLAGS)"

# 默认目标
.DEFAULT_GOAL := help

# 颜色定义（用于输出）
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

##@ 帮助信息

.PHONY: help
help: ## 显示帮助信息
	@echo "$(CYAN)GM Server Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)可用命令:$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(GREEN)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 开发环境

.PHONY: run
run: ## 运行开发服务器（使用默认配置）
	@echo "$(GREEN)启动开发服务器...$(NC)"
	@go run $(MAIN_FILE)

.PHONY: run-config
run-config: ## 运行开发服务器（使用YAML配置）
	@echo "$(GREEN)启动开发服务器（使用config.yaml）...$(NC)"
	@go run $(MAIN_FILE) --config config.yaml

.PHONY: dev
dev: ## 开发模式：运行并监听文件变化（需要安装air: go install github.com/cosmtrek/air@latest）
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)警告: air未安装，使用普通运行模式$(NC)"; \
		go run $(MAIN_FILE) --config config.yaml; \
	fi

.PHONY: test
test: ## 运行测试
	@echo "$(GREEN)运行测试...$(NC)"
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(GREEN)运行测试并生成覆盖率报告...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)覆盖率报告已生成: coverage.html$(NC)"

.PHONY: lint
lint: ## 运行代码检查
	@echo "$(GREEN)运行代码检查...$(NC)"
	@go vet ./...
	@go fmt ./...

##@ 构建

.PHONY: build
build: ## 构建当前平台的可执行文件
	@echo "$(GREEN)构建 $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "$(GREEN)构建完成: $(BUILD_DIR)/$(APP_NAME)$(NC)"

.PHONY: build-windows
build-windows: ## 构建 Windows 版本
	@echo "$(GREEN)构建 Windows 版本...$(NC)"
	@mkdir -p $(BUILD_DIR)/windows
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/windows/$(APP_NAME).exe $(MAIN_FILE)
	@echo "$(GREEN)构建完成: $(BUILD_DIR)/windows/$(APP_NAME).exe$(NC)"

.PHONY: build-linux
build-linux: ## 构建 Linux 版本
	@echo "$(GREEN)构建 Linux 版本...$(NC)"
	@mkdir -p $(BUILD_DIR)/linux
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/linux/$(APP_NAME) $(MAIN_FILE)
	@echo "$(GREEN)构建完成: $(BUILD_DIR)/linux/$(APP_NAME)$(NC)"

.PHONY: build-macos
build-macos: ## 构建 macOS 版本
	@echo "$(GREEN)构建 macOS 版本...$(NC)"
	@mkdir -p $(BUILD_DIR)/macos
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/macos/$(APP_NAME) $(MAIN_FILE)
	@echo "$(GREEN)构建完成: $(BUILD_DIR)/macos/$(APP_NAME)$(NC)"

.PHONY: build-macos-arm
build-macos-arm: ## 构建 macOS ARM64 版本
	@echo "$(GREEN)构建 macOS ARM64 版本...$(NC)"
	@mkdir -p $(BUILD_DIR)/macos-arm
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/macos-arm/$(APP_NAME) $(MAIN_FILE)
	@echo "$(GREEN)构建完成: $(BUILD_DIR)/macos-arm/$(APP_NAME)$(NC)"

.PHONY: build-all
build-all: ## 构建所有平台版本
	@echo "$(GREEN)构建所有平台版本...$(NC)"
	@$(MAKE) build-windows
	@$(MAKE) build-linux
	@$(MAKE) build-macos
	@$(MAKE) build-macos-arm
	@echo "$(GREEN)所有平台构建完成$(NC)"

##@ 环境配置

.PHONY: config-dev
config-dev: ## 创建开发环境配置文件
	@echo "$(GREEN)创建开发环境配置文件...$(NC)"
	@cp config/config.example.yaml config.dev.yaml
	@sed -i.bak 's/level: "info"/level: "debug"/' config.dev.yaml 2>/dev/null || \
	 sed -i '' 's/level: "info"/level: "debug"/' config.dev.yaml 2>/dev/null || \
	 echo "请手动编辑 config.dev.yaml，设置 log.level 为 debug"
	@rm -f config.dev.yaml.bak 2>/dev/null || true
	@echo "$(GREEN)配置文件已创建: config.dev.yaml$(NC)"

.PHONY: config-prod
config-prod: ## 创建生产环境配置文件
	@echo "$(GREEN)创建生产环境配置文件...$(NC)"
	@cp config/config.example.yaml config.prod.yaml
	@sed -i.bak 's/level: "info"/level: "warn"/' config.prod.yaml 2>/dev/null || \
	 sed -i '' 's/level: "info"/level: "warn"/' config.prod.yaml 2>/dev/null || \
	 echo "请手动编辑 config.prod.yaml，设置 log.level 为 warn"
	@sed -i.bak 's/encoding: "json"/encoding: "json"/' config.prod.yaml 2>/dev/null || \
	 sed -i '' 's/encoding: "json"/encoding: "json"/' config.prod.yaml 2>/dev/null || true
	@rm -f config.prod.yaml.bak 2>/dev/null || true
	@echo "$(GREEN)配置文件已创建: config.prod.yaml$(NC)"
	@echo "$(YELLOW)请编辑 config.prod.yaml 设置生产环境参数$(NC)"

##@ 数据库

.PHONY: db-init
db-init: ## 初始化数据库（Linux/macOS）
	@echo "$(GREEN)初始化数据库...$(NC)"
	@if [ -f scripts/init_db.sh ]; then \
		chmod +x scripts/init_db.sh && ./scripts/init_db.sh; \
	else \
		echo "$(RED)错误: scripts/init_db.sh 不存在$(NC)"; \
	fi

.PHONY: db-init-windows
db-init-windows: ## 初始化数据库（Windows）
	@echo "$(GREEN)初始化数据库...$(NC)"
	@if exist scripts\init_db.bat ( \
		scripts\init_db.bat \
	) else ( \
		echo 错误: scripts\init_db.bat 不存在 \
	)

##@ 依赖管理

.PHONY: deps
deps: ## 下载依赖
	@echo "$(GREEN)下载依赖...$(NC)"
	@go mod download

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "$(GREEN)更新依赖...$(NC)"
	@go get -u ./...
	@go mod tidy

.PHONY: deps-tidy
deps-tidy: ## 整理依赖
	@echo "$(GREEN)整理依赖...$(NC)"
	@go mod tidy

.PHONY: deps-vendor
deps-vendor: ## 创建vendor目录
	@echo "$(GREEN)创建vendor目录...$(NC)"
	@go mod vendor

##@ 清理

.PHONY: clean
clean: ## 清理构建文件
	@echo "$(GREEN)清理构建文件...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)清理完成$(NC)"

.PHONY: clean-all
clean-all: clean ## 清理所有生成文件（包括测试文件）
	@echo "$(GREEN)清理所有生成文件...$(NC)"
	@rm -rf vendor
	@go clean -cache
	@echo "$(GREEN)清理完成$(NC)"

##@ 部署

.PHONY: install
install: build ## 安装到系统路径（需要管理员权限）
	@echo "$(GREEN)安装 $(APP_NAME) 到系统路径...$(NC)"
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME) || \
	 echo "$(YELLOW)需要管理员权限，请手动复制$(NC)"

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "$(GREEN)构建 Docker 镜像...$(NC)"
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "$(GREEN)Docker 镜像构建完成$(NC)"

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	@echo "$(GREEN)运行 Docker 容器...$(NC)"
	@docker run -p 8080:8080 --rm $(APP_NAME):latest

##@ 工具

.PHONY: fmt
fmt: ## 格式化代码
	@echo "$(GREEN)格式化代码...$(NC)"
	@go fmt ./...

.PHONY: vet
vet: ## 运行 go vet
	@echo "$(GREEN)运行 go vet...$(NC)"
	@go vet ./...

.PHONY: mod-verify
mod-verify: ## 验证依赖
	@echo "$(GREEN)验证依赖...$(NC)"
	@go mod verify

.PHONY: info
info: ## 显示项目信息
	@echo "$(CYAN)=== 项目信息 ===$(NC)"
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Go 版本: $(shell go version)"
	@echo "工作目录: $(shell pwd)"

##@ 快速命令

.PHONY: quick-build
quick-build: deps-tidy build ## 快速构建（整理依赖+构建）

.PHONY: quick-test
quick-test: deps-tidy test ## 快速测试（整理依赖+测试）

.PHONY: quick-run
quick-run: deps-tidy run-config ## 快速运行（整理依赖+运行）
