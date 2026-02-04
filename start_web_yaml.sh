#!/bin/bash

# GM Server 启动脚本 - 使用YAML配置

APP_NAME="gm-server"
BUILD_DIR="build"
EXE_PATH="$BUILD_DIR/$APP_NAME"

echo "=== GM Server 启动脚本 (YAML配置) ==="
echo ""

# 检查YAML配置文件
if [ -f "config.yaml" ]; then
    CONFIG_FILE="config.yaml"
elif [ -f "config.yml" ]; then
    CONFIG_FILE="config.yml"
else
    echo "[错误] 未找到配置文件 config.yaml 或 config.yml"
    echo ""
    echo "请先创建配置文件："
    echo "  1. 复制示例配置: cp config/config.example.yaml config.yaml"
    echo "  2. 编辑配置文件: vim config.yaml"
    echo ""
    exit 1
fi

echo "[配置] 使用配置文件: $CONFIG_FILE"

# 检查是否已编译
if [ -f "$EXE_PATH" ]; then
    echo "[信息] 使用已编译的可执行文件: $EXE_PATH"
    USE_EXE=1
else
    echo "[信息] 未找到编译文件，将使用 go run 运行"
    USE_EXE=0
fi

echo ""
echo "[服务] 服务器将在 http://localhost:8080 启动"
echo "[提示] 按 Ctrl+C 停止服务器"
echo ""

# 启动服务
if [ "$USE_EXE" -eq 1 ]; then
    "$EXE_PATH" --config "$CONFIG_FILE"
else
    go run cmd/web --config "$CONFIG_FILE"
fi
