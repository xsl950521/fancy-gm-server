#!/bin/bash

# GM Server 停止脚本

echo "=== GM Server 停止脚本 ==="
echo ""

# 查找并停止服务进程
PID=$(lsof -ti:8080 2>/dev/null || netstat -tlnp 2>/dev/null | grep :8080 | awk '{print $7}' | cut -d'/' -f1 | head -1)

if [ -n "$PID" ]; then
    echo "[信息] 找到服务进程 PID: $PID"
    kill -9 "$PID" 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "[成功] 服务已停止"
    else
        echo "[警告] 停止进程失败，可能需要管理员权限"
    fi
else
    echo "[信息] 未找到运行在 8080 端口的服务"
fi

# 也尝试通过进程名停止
pkill -f "gm-server" 2>/dev/null
if [ $? -eq 0 ]; then
    echo "[成功] 通过进程名停止服务"
fi

echo ""
echo "[完成] 停止操作完成"
