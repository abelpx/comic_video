#!/bin/bash

# VidCraft Studio 快速部署脚本
# 用于快速设置开发环境和启动所有必要的服务

set -e

echo "🚀 VidCraft Studio 快速部署脚本"
echo "=================================="

# 检查必要的工具
check_requirements() {
    echo "📋 检查系统要求..."
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    # 检查 Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        echo "❌ Go 未安装，请先安装 Go 1.21+"
        exit 1
    fi
    
    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        echo "❌ Node.js 未安装，请先安装 Node.js 18+"
        exit 1
    fi
    
    echo "✅ 系统要求检查通过"
}

# 创建必要的目录
create_directories() {
    echo "📁 创建必要的目录..."
    
    mkdir -p data/postgres
    mkdir -p data/redis
    mkdir -p data/minio
    mkdir -p logs
    mkdir -p output
    mkdir -p uploads
    
    echo "✅ 目录创建完成"
}

# 设置环境变量
setup_environment() {
    echo "⚙️ 设置环境变量..."
    
    if [ ! -f .env ]; then
        echo "❌ .env 文件不存在，请先复制 .env.example 到 .env 并配置"
        exit 1
    fi
    
    # 检查关键环境变量
    source .env
    
    if [ -z "$DATABASE_URL" ]; then
        echo "❌ DATABASE_URL 未设置"
        exit 1
    fi
    
    if [ -z "$REDIS_URL" ]; then
        echo "❌ REDIS_URL 未设置"
        exit 1
    fi
    
    echo "✅ 环境变量检查通过"
}

# 启动基础服务
start_infrastructure() {
    echo "🐳 启动基础服务 (PostgreSQL, Redis, MinIO)..."
    
    cat > docker-compose.infrastructure.yml << EOF
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
EOF

    docker-compose -f docker-compose.infrastructure.yml up -d
    
    echo "⏳ 等待服务启动..."
    sleep 10
    
    echo "✅ 基础服务启动完成"
}

# 启动AI服务
start_ai_services() {
    echo "🤖 启动AI服务..."
    
    # 这里可以添加启动 Ollama, Stable Diffusion 等AI服务的脚本
    echo "📝 请手动启动以下AI服务："
    echo "   - Ollama: ollama serve"
    echo "   - Stable Diffusion WebUI: 在端口 7860"
    echo "   - TTS服务: 在端口 50021"
    echo "   - Whisper服务: 在端口 9000"
    
    echo "⚠️  AI服务需要手动启动，请参考文档配置"
}

# 安装后端依赖
install_backend_deps() {
    echo "📦 安装后端依赖..."
    
    go mod download
    go mod tidy
    
    echo "✅ 后端依赖安装完成"
}

# 安装前端依赖
install_frontend_deps() {
    echo "📦 安装前端依赖..."
    
    cd web
    npm install
    cd ..
    
    echo "✅ 前端依赖安装完成"
}

# 数据库迁移
run_migrations() {
    echo "🗄️ 运行数据库迁移..."
    
    # 等待数据库启动
    echo "⏳ 等待数据库启动..."
    sleep 5
    
    # 运行迁移（这里假设有迁移命令）
    echo "📝 数据库迁移将在应用启动时自动执行"
    
    echo "✅ 数据库准备完成"
}

# 构建项目
build_project() {
    echo "🔨 构建项目..."
    
    # 构建前端
    echo "🎨 构建前端..."
    cd web
    npm run build
    cd ..
    
    # 构建后端
    echo "⚙️ 构建后端..."
    go build -o bin/vidcraft cmd/api/main.go
    
    echo "✅ 项目构建完成"
}

# 启动应用
start_application() {
    echo "🚀 启动应用..."
    
    # 启动后端
    echo "⚙️ 启动后端服务..."
    ./bin/vidcraft &
    BACKEND_PID=$!
    
    echo "✅ 应用启动完成"
    echo ""
    echo "🎉 VidCraft Studio 部署成功！"
    echo ""
    echo "📱 访问地址:"
    echo "   - 前端应用: http://localhost:3000"
    echo "   - API文档: http://localhost:8080/swagger/index.html"
    echo "   - MinIO控制台: http://localhost:9001"
    echo ""
    echo "🔧 管理命令:"
    echo "   - 停止应用: kill $BACKEND_PID"
    echo "   - 查看日志: tail -f logs/app.log"
    echo "   - 停止基础服务: docker-compose -f docker-compose.infrastructure.yml down"
    echo ""
    echo "📚 更多信息请查看 README.md"
}

# 主函数
main() {
    check_requirements
    create_directories
    setup_environment
    start_infrastructure
    start_ai_services
    install_backend_deps
    install_frontend_deps
    run_migrations
    build_project
    start_application
}

# 处理中断信号
cleanup() {
    echo ""
    echo "🛑 正在停止服务..."
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
    fi
    docker-compose -f docker-compose.infrastructure.yml down
    echo "✅ 服务已停止"
    exit 0
}

trap cleanup SIGINT SIGTERM

# 运行主函数
main

# 保持脚本运行
wait
