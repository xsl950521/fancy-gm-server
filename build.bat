@echo off
REM GM Server 构建脚本 (Windows)
REM 设置UTF-8编码
chcp 65001 >nul 2>&1

setlocal enabledelayedexpansion

set APP_NAME=gm-server
set BUILD_DIR=build
set CMD_DIR=cmd\web
set MAIN_FILE=%CMD_DIR%\main.go

if "%1"=="" goto help
if "%1"=="help" goto help
if "%1"=="build" goto build
if "%1"=="build-windows" goto build-windows
if "%1"=="build-linux" goto build-linux
if "%1"=="build-macos" goto build-macos
if "%1"=="build-all" goto build-all
if "%1"=="run" goto run
if "%1"=="run-config" goto run-config
if "%1"=="test" goto test
if "%1"=="clean" goto clean
if "%1"=="deps" goto deps
if "%1"=="deps-tidy" goto deps-tidy
goto help

:help
echo === GM Server 构建脚本 ===
echo.
echo 用法: build.bat [命令]
echo.
echo 可用命令:
echo   build          - 构建当前平台的可执行文件
echo   build-windows  - 构建 Windows 版本
echo   build-linux    - 构建 Linux 版本
echo   build-macos    - 构建 macOS 版本
echo   build-all      - 构建所有平台版本
echo   run            - 运行开发服务器
echo   run-config     - 运行开发服务器（使用YAML配置）
echo   test           - 运行测试
echo   clean          - 清理构建文件
echo   deps           - 下载依赖
echo   deps-tidy      - 整理依赖
echo.
goto end

:build
echo [构建] 构建 %APP_NAME%...
if not exist %BUILD_DIR% mkdir %BUILD_DIR%
go build -ldflags "-s -w" -o %BUILD_DIR%\%APP_NAME%.exe ./%CMD_DIR%
if %errorlevel% equ 0 (
    echo [成功] 构建完成: %BUILD_DIR%\%APP_NAME%.exe
) else (
    echo [错误] 构建失败
    exit /b 1
)
goto end

:build-windows
echo [构建] 构建 Windows 版本...
if not exist %BUILD_DIR%\windows mkdir %BUILD_DIR%\windows
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o %BUILD_DIR%\windows\%APP_NAME%.exe ./%CMD_DIR%
if %errorlevel% equ 0 (
    echo [成功] 构建完成: %BUILD_DIR%\windows\%APP_NAME%.exe
) else (
    echo [错误] 构建失败
    exit /b 1
)
goto end

:build-linux
echo [构建] 构建 Linux 版本...
if not exist %BUILD_DIR%\linux mkdir %BUILD_DIR%\linux
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o %BUILD_DIR%\linux\%APP_NAME% ./%CMD_DIR%
if %errorlevel% equ 0 (
    echo [成功] 构建完成: %BUILD_DIR%\linux\%APP_NAME%
) else (
    echo [错误] 构建失败
    exit /b 1
)
goto end

:build-macos
echo [构建] 构建 macOS 版本...
if not exist %BUILD_DIR%\macos mkdir %BUILD_DIR%\macos
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w" -o %BUILD_DIR%\macos\%APP_NAME% ./%CMD_DIR%
if %errorlevel% equ 0 (
    echo [成功] 构建完成: %BUILD_DIR%\macos\%APP_NAME%
) else (
    echo [错误] 构建失败
    exit /b 1
)
goto end

:build-all
echo [构建] 构建所有平台版本...
call %0 build-windows
call %0 build-linux
call %0 build-macos
echo [成功] 所有平台构建完成
goto end

:run
echo [运行] 启动开发服务器...
go run ./%CMD_DIR%
goto end

:run-config
echo [运行] 启动开发服务器（使用config.yaml）...
go run ./%CMD_DIR% --config config.yaml
goto end

:test
echo [测试] 运行测试...
go test -v ./...
goto end

:clean
echo [清理] 清理构建文件...
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
if exist coverage.out del coverage.out
if exist coverage.html del coverage.html
echo [成功] 清理完成
goto end

:deps
echo [依赖] 下载依赖...
go mod download
goto end

:deps-tidy
echo [依赖] 整理依赖...
go mod tidy
goto end

:end
endlocal
