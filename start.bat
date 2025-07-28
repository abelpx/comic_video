@echo off
REM VidCraft API 启动脚本

echo ========================================
echo    VidCraft API - 小说转视频平台
echo ========================================
echo.

REM 设置Go环境
set GOROOT=C:\Users\Administrator\sdk\go1.24.3
set GOPATH=C:\Users\Administrator\go
set PATH=%GOROOT%\bin;%GOPATH%\bin;%PATH%

echo 检查Go环境...
go version
if errorlevel 1 (
    echo 错误: Go环境未正确配置
    pause
    exit /b 1
)

echo.
echo 检查项目依赖...
go mod tidy
if errorlevel 1 (
    echo 错误: 依赖检查失败
    pause
    exit /b 1
)

echo.
echo 构建项目...
go build -o vidcraft-api.exe ./cmd/api
if errorlevel 1 (
    echo 错误: 项目构建失败
    pause
    exit /b 1
)

echo.
echo 启动VidCraft API服务器...
echo 服务器将在 http://localhost:8080 启动
echo 按 Ctrl+C 停止服务器
echo.

vidcraft-api.exe

pause
