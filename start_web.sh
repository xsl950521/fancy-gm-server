#!/bin/bash

# GM Server 启动脚本

APP_NAME="gm-server"
BUILD_DIR="build"
EXE_PATH="$BUILD_DIR/$APP_NAME"

echo "=== GM Server 启动脚本 ==="
echo ""

# 检查是否已编译
if [ -f "$EXE_PATH" ]; then
    echo "[信息] 使用已编译的可执行文件: $EXE_PATH"
    USE_EXE=1
else
    echo "[信息] 未找到编译文件，将使用 go run 运行"
    USE_EXE=0
fi

# 检查配置文件
if [ -f "config.yaml" ]; then
    CONFIG_FILE="config.yaml"
    echo "[配置] 使用配置文件: config.yaml"
elif [ -f "config.yml" ]; then
    CONFIG_FILE="config.yml"
    echo "[配置] 使用配置文件: config.yml"
elif [ -f "config.json" ]; then
    CONFIG_FILE="config.json"
    echo "[配置] 使用配置文件: config.json"
elif [ -f "config.web.json" ]; then
    CONFIG_FILE="config.web.json"
    echo "[配置] 使用配置文件: config.web.json (旧格式)"
else
    CONFIG_FILE=""
    echo "[警告] 未找到配置文件，将使用默认配置和环境变量"
fi

echo ""
echo "[服务] 服务器将在 http://localhost:8080 启动"
echo "[提示] 按 Ctrl+C 停止服务器"
echo ""

# 启动服务
if [ "$USE_EXE" -eq 1 ]; then
    if [ -n "$CONFIG_FILE" ]; then
        "$EXE_PATH" --config "$CONFIG_FILE"
    else
        "$EXE_PATH"
    fi
else
    if [ -n "$CONFIG_FILE" ]; then
        go run cmd/web --config "$CONFIG_FILE"
    else
        go run cmd/web
    fi
fi
