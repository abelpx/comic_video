#!/bin/bash

# =============================================================================
# 🎬 Comic Video 本地开发启动脚本
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查开发环境依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go 1.19+"
        exit 1
    fi
    
    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    log_success "Go 版本: $go_version"
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker 服务未运行，请启动 Docker"
        exit 1
    fi
    
    log_success "Docker 运行正常"
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi
    
    log_success "Docker Compose 可用"
}

# 初始化环境
init_environment() {
    log_info "初始化开发环境..."
    
    # 创建环境配置文件
    if [ ! -f .env ]; then
        cp .env.dev .env
        log_success "已创建 .env 配置文件"
    else
        log_info ".env 文件已存在"
    fi
    
    # 创建必要目录
    mkdir -p logs/{app,postgres,redis,minio}
    mkdir -p output/{videos,images,audio}
    mkdir -p temp
    mkdir -p config
    
    log_success "目录结构创建完成"
    
    # 安装Go依赖
    log_info "安装Go依赖..."
    go mod tidy
    log_success "Go依赖安装完成"
}

# 启动基础设施服务
start_infrastructure() {
    log_info "启动基础设施服务..."
    
    # 启动Docker服务
    docker-compose up -d
    
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    check_services_health
}

# 检查服务健康状态
check_services_health() {
    log_info "检查服务健康状态..."
    
    local all_healthy=true
    
    # 检查PostgreSQL
    if pg_isready -h localhost -p 5432 -U comic_video > /dev/null 2>&1; then
        log_success "✅ PostgreSQL 连接正常"
    else
        log_error "❌ PostgreSQL 连接失败"
        all_healthy=false
    fi
    
    # 检查Redis
    if redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then
        log_success "✅ Redis 连接正常"
    else
        log_error "❌ Redis 连接失败"
        all_healthy=false
    fi
    
    # 检查MinIO
    if curl -f -s http://localhost:9000/minio/health/live > /dev/null 2>&1; then
        log_success "✅ MinIO 连接正常"
    else
        log_error "❌ MinIO 连接失败"
        all_healthy=false
    fi
    
    if [ "$all_healthy" = true ]; then
        log_success "🎉 所有基础设施服务运行正常!"
    else
        log_error "❌ 部分服务启动失败，请检查日志"
        return 1
    fi
}

# 启动应用
start_application() {
    log_info "启动Comic Video应用..."
    
    # 检查main.go是否存在
    if [ ! -f "cmd/server/main.go" ]; then
        log_error "应用入口文件不存在: cmd/server/main.go"
        exit 1
    fi
    
    log_info "🚀 启动Go应用服务器..."
    go run cmd/server/main.go
}

# 显示服务信息
show_service_info() {
    echo ""
    log_success "🎉 Comic Video 开发环境启动完成!"
    echo ""
    echo "📊 服务地址:"
    echo "  🌐 应用服务:      http://localhost:8080"
    echo "  🗄️ PostgreSQL:   localhost:5432"
    echo "  🔴 Redis:        localhost:6379"
    echo "  📁 MinIO API:    http://localhost:9000"
    echo "  🌐 MinIO控制台:   http://localhost:9001"
    echo ""
    echo "🔐 默认账户:"
    echo "  数据库: comic_video / comic_video_password"
    echo "  Redis: (密码: comic_video_redis)"
    echo "  MinIO: comic_video / comic_video_minio_password"
    echo ""
    echo "📝 管理命令:"
    echo "  make status          - 检查服务状态"
    echo "  make logs            - 查看服务日志"
    echo "  make stop-services   - 停止基础设施服务"
    echo "  make clean-services  - 清理服务数据"
    echo ""
}

# 停止服务
stop_services() {
    log_info "停止基础设施服务..."
    docker-compose down
    log_success "服务已停止"
}

# 清理服务
clean_services() {
    log_info "清理服务数据..."
    docker-compose down -v --remove-orphans
    docker system prune -f
    log_success "服务数据清理完成"
}

# 显示帮助信息
show_help() {
    echo "Comic Video 开发环境管理脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  start     启动完整开发环境 (默认)"
    echo "  stop      停止基础设施服务"
    echo "  restart   重启基础设施服务"
    echo "  status    检查服务状态"
    echo "  clean     清理服务数据"
    echo "  init      仅初始化环境"
    echo "  services  仅启动基础设施服务"
    echo "  app       仅启动应用"
    echo "  help      显示此帮助信息"
    echo ""
}

# 主函数
main() {
    local command=${1:-"start"}
    
    case $command in
        "start")
            echo "🎬 启动Comic Video开发环境"
            echo "================================"
            check_dependencies
            init_environment
            start_infrastructure
            show_service_info
            start_application
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            log_info "重启基础设施服务..."
            docker-compose restart
            check_services_health
            ;;
        "status")
            check_services_health
            ;;
        "clean")
            clean_services
            ;;
        "init")
            check_dependencies
            init_environment
            ;;
        "services")
            check_dependencies
            start_infrastructure
            show_service_info
            ;;
        "app")
            start_application
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 捕获Ctrl+C信号
trap 'echo ""; log_info "正在停止服务..."; stop_services; exit 0' INT

# 执行主函数
main "$@"
