@echo off
REM GM Server 停止脚本
REM 设置UTF-8编码以支持中文显示
chcp 65001 >nul 2>&1

setlocal enabledelayedexpansion

echo === GM Server 停止脚本 ===
echo.

REM 查找并停止服务进程
for /f "tokens=2" %%a in ('netstat -ano ^| findstr :8080 ^| findstr LISTENING') do (
    set PID=%%a
    echo [信息] 找到服务进程 PID: !PID!
    taskkill /F /PID !PID! >nul 2>&1
    if !errorlevel! equ 0 (
        echo [成功] 服务已停止
    ) else (
        echo [警告] 停止进程失败，可能需要管理员权限
    )
)

REM 也尝试通过进程名停止
taskkill /F /IM gm-server.exe >nul 2>&1
if !errorlevel! equ 0 (
    echo [成功] 通过进程名停止服务
)

echo.
echo [完成] 停止操作完成
endlocal
pause
