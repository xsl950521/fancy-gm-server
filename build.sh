#!/bin/bash

# GM Server 构建脚本 (Linux/macOS)

set -e

APP_NAME="gm-server"
BUILD_DIR="build"
CMD_DIR="cmd/web"
MAIN_FILE="$CMD_DIR/main.go"

# 颜色定义
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

show_help() {
    echo -e "${CYAN}=== GM Server 构建脚本 ===${NC}"
    echo ""
    echo "用法: ./build.sh [命令]"
    echo ""
    echo -e "${GREEN}可用命令:${NC}"
    echo "  build          - 构建当前平台的可执行文件"
    echo "  build-windows  - 构建 Windows 版本"
    echo "  build-linux    - 构建 Linux 版本"
    echo "  build-macos    - 构建 macOS 版本"
    echo "  build-macos-arm - 构建 macOS ARM64 版本"
    echo "  build-all      - 构建所有平台版本"
    echo "  run            - 运行开发服务器"
    echo "  run-config     - 运行开发服务器（使用YAML配置）"
    echo "  test           - 运行测试"
    echo "  clean          - 清理构建文件"
    echo "  deps           - 下载依赖"
    echo "  deps-tidy      - 整理依赖"
    echo ""
}

build_current() {
    echo -e "${GREEN}[构建]${NC} 构建 $APP_NAME..."
    mkdir -p "$BUILD_DIR"
    go build -ldflags "-s -w" -o "$BUILD_DIR/$APP_NAME" "$CMD_DIR"
    echo -e "${GREEN}[成功]${NC} 构建完成: $BUILD_DIR/$APP_NAME"
}

build_windows() {
    echo -e "${GREEN}[构建]${NC} 构建 Windows 版本..."
    mkdir -p "$BUILD_DIR/windows"
    GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o "$BUILD_DIR/windows/$APP_NAME.exe" "$CMD_DIR"
    echo -e "${GREEN}[成功]${NC} 构建完成: $BUILD_DIR/windows/$APP_NAME.exe"
}

build_linux() {
    echo -e "${GREEN}[构建]${NC} 构建 Linux 版本..."
    mkdir -p "$BUILD_DIR/linux"
    GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "$BUILD_DIR/linux/$APP_NAME" "$CMD_DIR"
    echo -e "${GREEN}[成功]${NC} 构建完成: $BUILD_DIR/linux/$APP_NAME"
}

build_macos() {
    echo -e "${GREEN}[构建]${NC} 构建 macOS 版本..."
    mkdir -p "$BUILD_DIR/macos"
    GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o "$BUILD_DIR/macos/$APP_NAME" "$CMD_DIR"
    echo -e "${GREEN}[成功]${NC} 构建完成: $BUILD_DIR/macos/$APP_NAME"
}

build_macos_arm() {
    echo -e "${GREEN}[构建]${NC} 构建 macOS ARM64 版本..."
    mkdir -p "$BUILD_DIR/macos-arm"
    GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o "$BUILD_DIR/macos-arm/$APP_NAME" "$CMD_DIR"
    echo -e "${GREEN}[成功]${NC} 构建完成: $BUILD_DIR/macos-arm/$APP_NAME"
}

build_all() {
    echo -e "${GREEN}[构建]${NC} 构建所有平台版本..."
    build_windows
    build_linux
    build_macos
    build_macos_arm
    echo -e "${GREEN}[成功]${NC} 所有平台构建完成"
}

run_dev() {
    echo -e "${GREEN}[运行]${NC} 启动开发服务器..."
    go run "$CMD_DIR"
}

run_config() {
    echo -e "${GREEN}[运行]${NC} 启动开发服务器（使用config.yaml）..."
    go run "$CMD_DIR" --config config.yaml
}

run_test() {
    echo -e "${GREEN}[测试]${NC} 运行测试..."
    go test -v ./...
}

clean_build() {
    echo -e "${GREEN}[清理]${NC} 清理构建文件..."
    rm -rf "$BUILD_DIR"
    rm -f coverage.out coverage.html
    echo -e "${GREEN}[成功]${NC} 清理完成"
}

download_deps() {
    echo -e "${GREEN}[依赖]${NC} 下载依赖..."
    go mod download
}

tidy_deps() {
    echo -e "${GREEN}[依赖]${NC} 整理依赖..."
    go mod tidy
}

# 主逻辑
case "${1:-help}" in
    build)
        build_current
        ;;
    build-windows)
        build_windows
        ;;
    build-linux)
        build_linux
        ;;
    build-macos)
        build_macos
        ;;
    build-macos-arm)
        build_macos_arm
        ;;
    build-all)
        build_all
        ;;
    run)
        run_dev
        ;;
    run-config)
        run_config
        ;;
    test)
        run_test
        ;;
    clean)
        clean_build
        ;;
    deps)
        download_deps
        ;;
    deps-tidy)
        tidy_deps
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}[错误]${NC} 未知命令: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
