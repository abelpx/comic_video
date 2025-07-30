#!/bin/bash

# VidCraft Studio 开发工具脚本
# 提供各种开发和调试命令
# 跨平台兼容: Windows (Git Bash/WSL), Linux, macOS

set -e

# 检测操作系统
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS=Linux;;
        Darwin*)    OS=Mac;;
        CYGWIN*|MINGW*|MSYS*) OS=Windows;;
        *)          OS="Unknown";;
    esac
}

# 跨平台的命令检查函数
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 设置Docker Compose命令
setup_docker_compose() {
    if command_exists docker-compose; then
        DOCKER_COMPOSE="docker-compose"
    elif docker compose version >/dev/null 2>&1; then
        DOCKER_COMPOSE="docker compose"
    else
        echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
}

detect_os
setup_docker_compose

COMPOSE_FILE="docker-compose.dev.yml"

# 显示帮助信息
show_help() {
    echo "🛠️  VidCraft Studio 开发工具"
    echo ""
    echo "用法: $0 <命令> [参数]"
    echo ""
    echo "可用命令:"
    echo "  start          启动开发环境"
    echo "  stop           停止开发环境"
    echo "  restart        重启开发环境"
    echo "  status         查看服务状态"
    echo "  logs [service] 查看日志"
    echo "  shell [service] 进入容器shell"
    echo "  build          重新构建镜像"
    echo "  clean          清理开发环境"
    echo "  db-reset       重置数据库"
    echo "  test           运行测试"
    echo "  lint           代码检查"
    echo "  format         代码格式化"
    echo "  deps           更新依赖"
    echo "  backup         备份开发数据"
    echo "  restore        恢复开发数据"
    echo "  tts-start      启动TTS服务"
    echo "  tts-stop       停止TTS服务"
    echo "  tts-restart    重启TTS服务"
    echo "  tts-logs       查看TTS日志"
    echo "  tts-test       测试TTS服务"
    echo ""
    echo "示例:"
    echo "  $0 start                    # 启动开发环境"
    echo "  $0 logs api-dev             # 查看API服务日志"
    echo "  $0 shell api-dev            # 进入API容器"
    echo "  $0 restart api-dev          # 重启API服务"
    echo "  $0 tts-test                 # 测试TTS服务"
}

# 启动开发环境
start_dev() {
    echo "🚀 启动开发环境..."
    ./scripts/dev-start.sh
}

# 停止开发环境
stop_dev() {
    echo "🛑 停止开发环境..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE down
    echo "✅ 开发环境已停止"
}

# 重启开发环境
restart_dev() {
    echo "🔄 重启开发环境..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE restart
    echo "✅ 开发环境已重启"
}

# 查看服务状态
show_status() {
    echo "📊 服务状态:"
    $DOCKER_COMPOSE -f $COMPOSE_FILE ps
    echo ""
    echo "🔍 健康检查:"

    # 跨平台的curl函数
    safe_curl() {
        local url="$1"
        if command_exists curl; then
            curl -f -s "$url" >/dev/null 2>&1
        elif command_exists wget; then
            wget -q --spider "$url" >/dev/null 2>&1
        else
            return 1
        fi
    }

    # 检查API服务
    if safe_curl "http://localhost:8080/health"; then
        echo "   ✅ API服务: 正常"
    else
        echo "   ❌ API服务: 异常"
    fi

    # 检查前端服务
    if safe_curl "http://localhost:3000"; then
        echo "   ✅ 前端服务: 正常"
    else
        echo "   ❌ 前端服务: 异常"
    fi

    # 检查TTS服务
    if safe_curl "http://localhost:8000/v2/health/ready"; then
        echo "   ✅ TTS服务: 正常"
    else
        echo "   ⚠️  TTS服务: 异常或未启动"
        # 检查TTS容器状态
        if $DOCKER_COMPOSE -f $COMPOSE_FILE ps tts 2>/dev/null | grep -q "Up"; then
            echo "      TTS容器运行中，可能正在初始化..."
        elif $DOCKER_COMPOSE -f $COMPOSE_FILE ps tts 2>/dev/null | grep -q "Exit"; then
            echo "      TTS容器已退出，请检查日志"
        else
            echo "      TTS容器未启动"
        fi
    fi
}

# 查看日志
show_logs() {
    local service=$1
    if [ -z "$service" ]; then
        echo "📋 显示所有服务日志:"
        $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f
    else
        echo "📋 显示 $service 服务日志:"
        $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f $service
    fi
}

# 进入容器shell
enter_shell() {
    local service=$1
    if [ -z "$service" ]; then
        service="api-dev"
    fi

    echo "🐚 进入 $service 容器..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE exec $service sh
}

# 重新构建镜像
build_images() {
    echo "🔨 重新构建开发镜像..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE build --no-cache
    echo "✅ 镜像构建完成"
}

# 清理开发环境
clean_dev() {
    echo "🧹 清理开发环境..."

    # 停止并删除容器
    $DOCKER_COMPOSE -f $COMPOSE_FILE down --volumes --remove-orphans

    # 删除开发镜像（跨平台兼容）
    if command_exists docker && command_exists grep && command_exists awk; then
        docker images | grep vidcraft | grep dev | awk '{print $3}' | while read -r image_id; do
            if [ -n "$image_id" ]; then
                docker rmi "$image_id" 2>/dev/null || true
            fi
        done
    fi

    # 清理临时文件
    rm -rf tmp/* 2>/dev/null || true
    rm -rf logs/* 2>/dev/null || true

    echo "✅ 开发环境已清理"
}

# 重置数据库
reset_db() {
    echo "🗄️  重置数据库..."

    # 停止API服务
    $DOCKER_COMPOSE -f $COMPOSE_FILE stop api-dev

    # 重启数据库
    $DOCKER_COMPOSE -f $COMPOSE_FILE restart postgres

    # 等待数据库就绪
    sleep 10

    # 重启API服务
    $DOCKER_COMPOSE -f $COMPOSE_FILE start api-dev

    echo "✅ 数据库已重置"
}

# 运行测试
run_tests() {
    echo "🧪 运行测试..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE exec api-dev go test ./...
}

# 代码检查
run_lint() {
    echo "🔍 运行代码检查..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE exec api-dev golangci-lint run
}

# 代码格式化
format_code() {
    echo "✨ 格式化代码..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE exec api-dev go fmt ./...
    echo "✅ 代码格式化完成"
}

# 更新依赖
update_deps() {
    echo "📦 更新Go依赖..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE exec api-dev go mod tidy
    $DOCKER_COMPOSE -f $COMPOSE_FILE exec api-dev go mod download

    echo "📦 更新前端依赖..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE exec web-dev npm update

    echo "✅ 依赖更新完成"
}

# 备份开发数据
backup_data() {
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"

    echo "💾 备份开发数据到 $backup_dir..."

    # 备份数据库
    $DOCKER_COMPOSE -f $COMPOSE_FILE exec postgres pg_dump -U vidcraft vidcraft > "$backup_dir/database.sql"

    # 备份MinIO数据（如果mc命令可用）
    if $DOCKER_COMPOSE -f $COMPOSE_FILE exec minio which mc >/dev/null 2>&1; then
        $DOCKER_COMPOSE -f $COMPOSE_FILE exec minio mc mirror /data "$backup_dir/minio/" 2>/dev/null || true
    fi

    echo "✅ 数据备份完成: $backup_dir"
}

# 恢复开发数据
restore_data() {
    local backup_dir=$1
    if [ -z "$backup_dir" ]; then
        echo "❌ 请指定备份目录"
        echo "用法: $0 restore <backup_directory>"
        exit 1
    fi

    if [ ! -d "$backup_dir" ]; then
        echo "❌ 备份目录不存在: $backup_dir"
        exit 1
    fi

    echo "📥 从 $backup_dir 恢复数据..."

    # 恢复数据库
    if [ -f "$backup_dir/database.sql" ]; then
        $DOCKER_COMPOSE -f $COMPOSE_FILE exec -T postgres psql -U vidcraft vidcraft < "$backup_dir/database.sql"
        echo "✅ 数据库恢复完成"
    fi

    # 恢复MinIO数据
    if [ -d "$backup_dir/minio" ]; then
        if $DOCKER_COMPOSE -f $COMPOSE_FILE exec minio which mc >/dev/null 2>&1; then
            $DOCKER_COMPOSE -f $COMPOSE_FILE exec minio mc mirror "$backup_dir/minio/" /data 2>/dev/null || true
            echo "✅ MinIO数据恢复完成"
        fi
    fi
}

# TTS服务管理函数
tts_start() {
    echo "🎤 启动TTS服务..."

    # 检查NVIDIA支持
    if ! docker info 2>/dev/null | grep -q "nvidia"; then
        echo "⚠️  警告: 未检测到NVIDIA Docker支持"
        read -p "是否继续启动？(y/N): " -r reply
        if [[ ! $reply =~ ^[Yy]$ ]]; then
            return 1
        fi
    fi

    # 设置环境变量
    export COMPOSE_PROFILES=ai-services

    # 启动TTS服务
    $DOCKER_COMPOSE -f $COMPOSE_FILE up -d tts

    echo "⏳ 等待TTS服务启动..."
    sleep 30

    # 检查服务状态
    local timeout=120
    while [ $timeout -gt 0 ]; do
        if safe_curl "http://localhost:8000/v2/health/ready"; then
            echo "✅ TTS服务启动成功"
            return 0
        fi
        echo "   等待中... (剩余 ${timeout}s)"
        sleep 5
        timeout=$((timeout-5))
    done

    echo "❌ TTS服务启动超时，请检查日志"
    $DOCKER_COMPOSE -f $COMPOSE_FILE logs tts
    return 1
}

tts_stop() {
    echo "🛑 停止TTS服务..."
    $DOCKER_COMPOSE -f $COMPOSE_FILE stop tts
    echo "✅ TTS服务已停止"
}

tts_restart() {
    echo "🔄 重启TTS服务..."
    tts_stop
    sleep 5
    tts_start
}

tts_logs() {
    echo "📋 显示TTS服务日志:"
    $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f tts
}

tts_test() {
    echo "🧪 测试TTS服务..."

    # 检查服务状态
    if ! safe_curl "http://localhost:8000/v2/health/ready"; then
        echo "❌ TTS服务未运行，请先启动服务"
        echo "   运行: $0 tts-start"
        return 1
    fi

    echo "✅ TTS服务运行正常"

    # 测试API端点
    echo "🔍 测试API端点..."

    # 测试健康检查
    if safe_curl "http://localhost:8080/api/v1/tts/health"; then
        echo "✅ API健康检查: 正常"
    else
        echo "❌ API健康检查: 失败"
    fi

    # 测试服务信息
    if safe_curl "http://localhost:8080/api/v1/tts/info"; then
        echo "✅ 服务信息接口: 正常"
    else
        echo "❌ 服务信息接口: 失败"
    fi

    # 测试语音模型列表
    if safe_curl "http://localhost:8080/api/v1/tts/voices"; then
        echo "✅ 语音模型接口: 正常"
    else
        echo "❌ 语音模型接口: 失败"
    fi

    # 测试语音生成
    if command_exists curl; then
        echo "🎵 测试语音生成..."
        local test_response=$(curl -s -w "%{http_code}" -o /dev/null "http://localhost:8080/api/v1/tts/test?voice_model=spark_tts_zh&language=zh")
        if [ "$test_response" = "200" ]; then
            echo "✅ 语音生成测试: 成功"
        else
            echo "❌ 语音生成测试: 失败 (HTTP $test_response)"
        fi
    fi

    echo ""
    echo "🎉 TTS服务测试完成！"
    echo "💡 可以访问测试页面: http://localhost:3000/app/tts-test"
}

# 主函数
main() {
    case $1 in
        start)
            start_dev
            ;;
        stop)
            stop_dev
            ;;
        restart)
            if [ -n "$2" ]; then
                echo "🔄 重启 $2 服务..."
                $DOCKER_COMPOSE -f $COMPOSE_FILE restart $2
            else
                restart_dev
            fi
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs $2
            ;;
        shell)
            enter_shell $2
            ;;
        build)
            build_images
            ;;
        clean)
            clean_dev
            ;;
        db-reset)
            reset_db
            ;;
        test)
            run_tests
            ;;
        lint)
            run_lint
            ;;
        format)
            format_code
            ;;
        deps)
            update_deps
            ;;
        backup)
            backup_data
            ;;
        restore)
            restore_data $2
            ;;
        tts-start)
            tts_start
            ;;
        tts-stop)
            tts_stop
            ;;
        tts-restart)
            tts_restart
            ;;
        tts-logs)
            tts_logs
            ;;
        tts-test)
            tts_test
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo "❌ 未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 检查是否在项目根目录
if [ ! -f "go.mod" ]; then
    echo "❌ 请在项目根目录运行此脚本"
    exit 1
fi

# 执行主函数
main "$@"
