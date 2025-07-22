# VidCraft Studio 快速部署脚本 (Windows PowerShell)
# 用于快速设置开发环境和启动所有必要的服务

param(
    [switch]$SkipAI,
    [switch]$DevMode
)

Write-Host "🚀 VidCraft Studio 快速部署脚本" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Green

# 检查必要的工具
function Check-Requirements {
    Write-Host "📋 检查系统要求..." -ForegroundColor Yellow
    
    # 检查 Docker
    if (!(Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Docker 未安装，请先安装 Docker Desktop" -ForegroundColor Red
        exit 1
    }
    
    # 检查 Docker Compose
    if (!(Get-Command docker-compose -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Docker Compose 未安装，请先安装 Docker Compose" -ForegroundColor Red
        exit 1
    }
    
    # 检查 Go
    if (!(Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Go 未安装，请先安装 Go 1.21+" -ForegroundColor Red
        exit 1
    }
    
    # 检查 Node.js
    if (!(Get-Command node -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Node.js 未安装，请先安装 Node.js 18+" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ 系统要求检查通过" -ForegroundColor Green
}

# 创建必要的目录
function Create-Directories {
    Write-Host "📁 创建必要的目录..." -ForegroundColor Yellow
    
    $directories = @("data\postgres", "data\redis", "data\minio", "logs", "output", "uploads")
    
    foreach ($dir in $directories) {
        if (!(Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
    }
    
    Write-Host "✅ 目录创建完成" -ForegroundColor Green
}

# 设置环境变量
function Setup-Environment {
    Write-Host "⚙️ 设置环境变量..." -ForegroundColor Yellow
    
    if (!(Test-Path ".env")) {
        Write-Host "❌ .env 文件不存在，请先复制 .env.example 到 .env 并配置" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ 环境变量检查通过" -ForegroundColor Green
}

# 启动基础服务
function Start-Infrastructure {
    Write-Host "🐳 启动基础服务 (PostgreSQL, Redis, MinIO)..." -ForegroundColor Yellow
    
    $dockerComposeContent = @"
version: '3.8'

services:
  postgres:
    image: postgres:15
    container_name: vidcraft_postgres
    environment:
      POSTGRES_DB: vidcraft
      POSTGRES_USER: vidcraft
      POSTGRES_PASSWORD: vidcraft123
    ports:
      - "5432:5432"
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: vidcraft_redis
    ports:
      - "6379:6379"
    volumes:
      - ./data/redis:/data
    restart: unless-stopped

  minio:
    image: minio/minio:latest
    container_name: vidcraft_minio
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - ./data/minio:/data
    command: server /data --console-address ":9001"
    restart: unless-stopped
"@

    $dockerComposeContent | Out-File -FilePath "docker-compose.infrastructure.yml" -Encoding UTF8
    
    docker-compose -f docker-compose.infrastructure.yml up -d
    
    Write-Host "⏳ 等待服务启动..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
    
    Write-Host "✅ 基础服务启动完成" -ForegroundColor Green
}

# 启动AI服务提示
function Start-AIServices {
    if ($SkipAI) {
        Write-Host "⏭️ 跳过AI服务启动" -ForegroundColor Yellow
        return
    }
    
    Write-Host "🤖 AI服务配置提示..." -ForegroundColor Yellow
    Write-Host "📝 请手动启动以下AI服务：" -ForegroundColor Cyan
    Write-Host "   - Ollama: ollama serve" -ForegroundColor White
    Write-Host "   - Stable Diffusion WebUI: 在端口 7860" -ForegroundColor White
    Write-Host "   - TTS服务: 在端口 50021" -ForegroundColor White
    Write-Host "   - Whisper服务: 在端口 9000" -ForegroundColor White
    Write-Host ""
    Write-Host "⚠️  AI服务需要手动启动，请参考文档配置" -ForegroundColor Yellow
}

# 安装后端依赖
function Install-BackendDeps {
    Write-Host "📦 安装后端依赖..." -ForegroundColor Yellow
    
    go mod download
    go mod tidy
    
    Write-Host "✅ 后端依赖安装完成" -ForegroundColor Green
}

# 安装前端依赖
function Install-FrontendDeps {
    Write-Host "📦 安装前端依赖..." -ForegroundColor Yellow
    
    Push-Location web
    npm install
    Pop-Location
    
    Write-Host "✅ 前端依赖安装完成" -ForegroundColor Green
}

# 构建项目
function Build-Project {
    Write-Host "🔨 构建项目..." -ForegroundColor Yellow
    
    # 构建前端
    Write-Host "🎨 构建前端..." -ForegroundColor Cyan
    Push-Location web
    npm run build
    Pop-Location
    
    # 构建后端
    Write-Host "⚙️ 构建后端..." -ForegroundColor Cyan
    if (!(Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" -Force | Out-Null
    }
    go build -o bin/vidcraft.exe cmd/api/main.go
    
    Write-Host "✅ 项目构建完成" -ForegroundColor Green
}

# 启动应用
function Start-Application {
    Write-Host "🚀 启动应用..." -ForegroundColor Yellow
    
    if ($DevMode) {
        Write-Host "🔧 开发模式启动..." -ForegroundColor Cyan
        Write-Host "请在新的终端窗口中运行以下命令：" -ForegroundColor Yellow
        Write-Host "后端: go run cmd/api/main.go" -ForegroundColor White
        Write-Host "前端: cd web && npm run dev" -ForegroundColor White
    } else {
        Write-Host "⚙️ 启动后端服务..." -ForegroundColor Cyan
        Start-Process -FilePath ".\bin\vidcraft.exe" -WindowStyle Hidden
    }
    
    Write-Host ""
    Write-Host "🎉 VidCraft Studio 部署成功！" -ForegroundColor Green
    Write-Host ""
    Write-Host "📱 访问地址:" -ForegroundColor Cyan
    Write-Host "   - 前端应用: http://localhost:3000" -ForegroundColor White
    Write-Host "   - API服务: http://localhost:8080" -ForegroundColor White
    Write-Host "   - MinIO控制台: http://localhost:9001 (minioadmin/minioadmin123)" -ForegroundColor White
    Write-Host ""
    Write-Host "🔧 管理命令:" -ForegroundColor Cyan
    Write-Host "   - 停止基础服务: docker-compose -f docker-compose.infrastructure.yml down" -ForegroundColor White
    Write-Host "   - 查看容器状态: docker ps" -ForegroundColor White
    Write-Host ""
    Write-Host "📚 更多信息请查看 README.md" -ForegroundColor Cyan
}

# 清理函数
function Cleanup {
    Write-Host ""
    Write-Host "🛑 正在停止服务..." -ForegroundColor Yellow
    docker-compose -f docker-compose.infrastructure.yml down
    Write-Host "✅ 服务已停止" -ForegroundColor Green
}

# 主函数
function Main {
    try {
        Check-Requirements
        Create-Directories
        Setup-Environment
        Start-Infrastructure
        Start-AIServices
        Install-BackendDeps
        Install-FrontendDeps
        Build-Project
        Start-Application
    }
    catch {
        Write-Host "❌ 部署过程中出现错误: $($_.Exception.Message)" -ForegroundColor Red
        Cleanup
        exit 1
    }
}

# 处理 Ctrl+C
$null = Register-EngineEvent PowerShell.Exiting -Action { Cleanup }

# 运行主函数
Main

Write-Host ""
Write-Host "按任意键退出..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
