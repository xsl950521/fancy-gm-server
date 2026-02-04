#!/bin/bash

echo "启动 Redis 数据解析工具 Web 服务器..."
echo ""
echo "服务器将在 http://localhost:8080 启动"
echo "使用配置文件: config.web.json"
echo "按 Ctrl+C 停止服务器"
echo ""

go run cmd/web/main.go --config config.web.json
