@echo off
REM GM Server 启动脚本 - 使用YAML配置
REM 设置UTF-8编码以支持中文显示
chcp 65001 >nul 2>&1

setlocal enabledelayedexpansion

set APP_NAME=gm-server
set BUILD_DIR=build
set EXE_PATH=%BUILD_DIR%\%APP_NAME%.exe

echo === GM Server 启动脚本 (YAML配置) ===
echo.

REM 检查YAML配置文件
if not exist "config.yaml" (
    if not exist "config.yml" (
        echo [错误] 未找到配置文件 config.yaml 或 config.yml
        echo.
        echo 请先创建配置文件：
        echo   1. 复制示例配置: copy config\config.example.yaml config.yaml
        echo   2. 编辑配置文件: notepad config.yaml
        echo.
        pause
        exit /b 1
    ) else (
        set CONFIG_FILE=config.yml
    )
) else (
    set CONFIG_FILE=config.yaml
)

echo [配置] 使用配置文件: %CONFIG_FILE%

REM 检查是否已编译
if exist "%EXE_PATH%" (
    echo [信息] 使用已编译的可执行文件: %EXE_PATH%
    set USE_EXE=1
) else (
    echo [信息] 未找到编译文件，将使用 go run 运行
    set USE_EXE=0
)

echo.
echo [服务] 服务器将在 http://localhost:8080 启动
echo [提示] 按 Ctrl+C 停止服务器
echo.

REM 启动服务
if !USE_EXE! equ 1 (
    "%EXE_PATH%" --config %CONFIG_FILE%
) else (
    go run ./cmd/web --config %CONFIG_FILE%
)

endlocal
pause
