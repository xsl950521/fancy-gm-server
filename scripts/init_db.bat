@echo off
REM 数据库一键初始化脚本 (Windows)
REM 用于创建数据库和用户

REM 设置代码页为UTF-8以支持中文显示
chcp 65001 >nul 2>&1

setlocal enabledelayedexpansion

echo === GM Server 数据库初始化脚本 ===
echo.

REM 默认配置
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=5432
if "%DB_USER%"=="" set DB_USER=postgres
if "%DB_NAME%"=="" set DB_NAME=gm_server

REM 检查psql是否安装
where psql >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 未找到 psql 命令，请先安装 PostgreSQL 客户端
    exit /b 1
)

echo 连接信息:
echo   主机: %DB_HOST%
echo   端口: %DB_PORT%
echo   用户: %DB_USER%
echo   数据库: %DB_NAME%
echo.

REM 提示输入密码
if "%PGPASSWORD%"=="" (
    set /p PGPASSWORD="请输入 PostgreSQL 管理员密码: "
)

set PGHOST=%DB_HOST%
set PGPORT=%DB_PORT%
set PGUSER=%DB_USER%

REM 检查数据库是否已存在
psql -lqt 2>nul | findstr /C:"%DB_NAME%" >nul
if %errorlevel% equ 0 (
    echo 数据库 '%DB_NAME%' 已存在
    set /p confirm="是否删除并重新创建? (y/N): "
    if /i "!confirm!"=="y" (
        echo 正在删除数据库 '%DB_NAME%'...
        psql -c "DROP DATABASE IF EXISTS %DB_NAME%;" 2>nul
        if !errorlevel! neq 0 (
            echo [错误] 删除数据库失败，请检查权限和连接
            exit /b 1
        )
        echo 数据库已删除
    ) else (
        echo 跳过数据库创建
        exit /b 0
    )
)

REM 创建数据库
echo 正在创建数据库 '%DB_NAME%'...
psql -c "CREATE DATABASE %DB_NAME%;" 2>nul
if %errorlevel% neq 0 (
    echo [错误] 数据库创建失败，请检查权限和连接
    exit /b 1
)
echo [成功] 数据库创建成功

REM 创建扩展
echo 正在创建数据库扩展...
psql -d %DB_NAME% -c "CREATE EXTENSION IF NOT EXISTS ""uuid-ossp"";" >nul 2>&1
if %errorlevel% equ 0 (
    echo [成功] 扩展创建成功
) else (
    echo [警告] 扩展创建失败，可能已存在或权限不足
)

echo.
echo === 数据库初始化完成 ===
echo 数据库名称: %DB_NAME%
echo 现在可以启动服务，服务会自动执行表结构迁移
echo.

endlocal
