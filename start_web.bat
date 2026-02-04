@echo off
REM GM Server 启动脚本
REM 设置UTF-8编码以支持中文显示
chcp 65001 >nul 2>&1

setlocal enabledelayedexpansion

set APP_NAME=gm-server
set BUILD_DIR=build
set EXE_PATH=%BUILD_DIR%\%APP_NAME%.exe

echo === GM Server 启动脚本 ===
echo.

REM 检查是否已编译
if exist "%EXE_PATH%" (
    echo [信息] 使用已编译的可执行文件: %EXE_PATH%
    set USE_EXE=1
) else (
    echo [信息] 未找到编译文件，将使用 go run 运行
    set USE_EXE=0
)

REM 检查配置文件
if exist "config.yaml" (
    set CONFIG_FILE=config.yaml
    echo [配置] 使用配置文件: config.yaml
) else if exist "config.yml" (
    set CONFIG_FILE=config.yml
    echo [配置] 使用配置文件: config.yml
) else if exist "config.json" (
    set CONFIG_FILE=config.json
    echo [配置] 使用配置文件: config.json
) else if exist "config.web.json" (
    set CONFIG_FILE=config.web.json
    echo [配置] 使用配置文件: config.web.json (旧格式)
) else (
    set CONFIG_FILE=
    echo [警告] 未找到配置文件，将使用默认配置和环境变量
)

echo.
echo [服务] 服务器将在 http://localhost:8080 启动
echo [提示] 按 Ctrl+C 停止服务器
echo.

REM 启动服务
if !USE_EXE! equ 1 (
    if defined CONFIG_FILE (
        "%EXE_PATH%" --config !CONFIG_FILE!
    ) else (
        "%EXE_PATH%"
    )
) else (
    if defined CONFIG_FILE (
        go run ./cmd/web --config !CONFIG_FILE!
    ) else (
        go run ./cmd/web
    )
)

endlocal
pause
