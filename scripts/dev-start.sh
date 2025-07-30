#!/bin/bash

# VidCraft Studio 开发环境启动脚本
# 支持热重载的开发环境，集成外部AI服务
# 跨平台兼容: Windows (Git Bash/WSL), Linux, macOS

set -e

# 解析命令行参数
SKIP_TTS=false
SHOW_LOGS=false
SKIP_NVIDIA_CHECK=false
SKIP_EXTERNAL_CHECK=false
BASIC_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-tts)
            SKIP_TTS=true
            shift
            ;;
        --show-logs)
            SHOW_LOGS=true
            shift
            ;;
        --skip-nvidia-check)
            SKIP_NVIDIA_CHECK=true
            shift
            ;;
        --skip-external-check)
            SKIP_EXTERNAL_CHECK=true
            shift
            ;;
        --basic-only)
            BASIC_ONLY=true
            shift
            ;;
        -h|--help)
            echo "VidCraft Studio 开发环境启动脚本"
            echo ""
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --skip-tts             跳过TTS服务启动"
            echo "  --show-logs            启动后显示实时日志"
            echo "  --skip-nvidia-check    跳过NVIDIA检查"
            echo "  --skip-external-check  跳过外部AI服务检查"
            echo "  --basic-only           仅启动基础服务(数据库、缓存等)"
            echo "  -h, --help             显示帮助信息"
            echo ""
            echo "示例:"
            echo "  $0                     # 完整启动"
            echo "  $0 --skip-tts          # 跳过TTS服务"
            echo "  $0 --basic-only        # 仅启动基础服务"
            echo ""
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            echo "使用 $0 --help 查看帮助"
            exit 1
            ;;
    esac
done

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

# 跨平台的curl函数
safe_curl() {
    local url="$1"
    if command_exists curl; then
        curl -f -s "$url" >/dev/null 2>&1
    elif command_exists wget; then
        wget -q --spider "$url" >/dev/null 2>&1
    else
        # 如果没有curl或wget，返回成功（跳过检查）
        return 0
    fi
}

# 跨平台的用户输入函数
read_user_input() {
    local prompt="$1"
    local default="$2"

    echo -n "$prompt"
    read -r reply

    if [ -z "$reply" ]; then
        reply="$default"
    fi
    echo "$reply"
}

detect_os

echo "🚀 启动 VidCraft Studio 开发环境 (支持热重载)..."
echo "🖥️  检测到操作系统: $OS"

# 检查Docker
if ! command_exists docker; then
    echo "❌ Docker 未安装，请先安装 Docker"
    echo "   下载地址: https://www.docker.com/products/docker-desktop"
    exit 1
fi

# 检查Docker Compose (支持新旧版本)
DOCKER_COMPOSE=""
if command_exists docker-compose; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    echo "   或使用 Docker Desktop (内置 docker compose 命令)"
    exit 1
fi

echo "🐳 使用 Docker Compose 命令: $DOCKER_COMPOSE"

# 检查外部AI服务连接
check_external_services() {
    if [ "$SKIP_EXTERNAL_CHECK" = true ]; then
        echo "⏭️  跳过外部AI服务检查"
        return 0
    fi

    echo ""
    echo "🔍 检查外部AI服务连接..."

    local ollama_ok=false
    local sd_ok=false

    # 检查Ollama服务
    echo "   检查Ollama服务 (127.0.0.1:11434)..."
    if safe_curl "http://127.0.0.1:11434/api/tags" 10; then
        echo "   ✅ Ollama服务连接正常"
        ollama_ok=true

        # 检查可用模型
        if command_exists curl; then
            local models=$(curl -s http://127.0.0.1:11434/api/tags 2>/dev/null | grep -o '"name":"[^"]*"' | cut -d'"' -f4 | head -3)
            if [ -n "$models" ]; then
                echo "      可用模型: $(echo $models | tr '\n' ', ' | sed 's/,$//')"
            fi
        fi
    else
        echo "   ❌ Ollama服务连接失败"
        echo "      请确保Ollama服务运行在 127.0.0.1:11434"
        echo "      启动命令: ollama serve"
    fi

    # 检查Stable Diffusion服务
    echo "   检查Stable Diffusion服务 (127.0.0.1:7860)..."
    if safe_curl "http://127.0.0.1:7860/sdapi/v1/options" 10; then
        echo "   ✅ Stable Diffusion服务连接正常"
        sd_ok=true

        # 检查当前模型
        if command_exists curl; then
            local options=$(curl -s http://127.0.0.1:7860/sdapi/v1/options 2>/dev/null)
            if echo "$options" | grep -q '"sd_model_checkpoint"'; then
                local model=$(echo "$options" | grep -o '"sd_model_checkpoint":"[^"]*"' | cut -d'"' -f4)
                echo "      当前模型: $model"
            fi
        fi
    else
        echo "   ❌ Stable Diffusion服务连接失败"
        echo "      请确保SD WebUI运行在 127.0.0.1:7860"
        echo "      启动命令: python launch.py --api"
    fi

    echo ""
    echo "📊 外部服务状态:"
    echo "   Ollama (文本生成): $([ "$ollama_ok" = true ] && echo "✅ 正常" || echo "❌ 异常")"
    echo "   Stable Diffusion (图像生成): $([ "$sd_ok" = true ] && echo "✅ 正常" || echo "❌ 异常")"

    if [ "$ollama_ok" = false ] || [ "$sd_ok" = false ]; then
        echo ""
        echo "⚠️  警告: 部分外部AI服务未运行，相关功能可能受限"
        reply=$(read_user_input "是否继续启动？(y/N): " "N")
        if [[ ! $reply =~ ^[Yy]$ ]]; then
            echo "💡 提示: 使用 --skip-external-check 跳过外部服务检查"
            exit 1
        fi
    fi
}

# 检查NVIDIA Docker支持（TTS服务需要GPU）
check_nvidia_support() {
    if [ "$SKIP_NVIDIA_CHECK" = true ] || [ "$SKIP_TTS" = true ]; then
        return 0
    fi

    echo ""
    echo "🎮 检查GPU支持..."

    nvidia_support=false
    if docker info 2>/dev/null | grep -q "nvidia"; then
        nvidia_support=true
        echo "   ✅ NVIDIA Docker支持已启用"

        # 检查GPU状态
        if command_exists nvidia-smi; then
            echo "   🔍 GPU状态:"
            nvidia-smi --query-gpu=name,memory.total,memory.used --format=csv,noheader,nounits 2>/dev/null | head -3 | while read line; do
                echo "      $line"
            done
        fi
    else
        echo "   ❌ NVIDIA Docker支持未检测到"
        echo "   请确保已安装 nvidia-docker2 和 nvidia-container-toolkit"

        reply=$(read_user_input "是否继续启动？(TTS服务可能无法使用GPU加速) (y/N): " "N")
        if [[ ! $reply =~ ^[Yy]$ ]]; then
            echo "💡 提示: 使用 --skip-tts 跳过TTS服务启动"
            exit 1
        fi
    fi
}

# 设置环境变量
export MODEL_ID=${MODEL_ID:-spark-tts}
if [ "$SKIP_TTS" = true ]; then
    export COMPOSE_PROFILES=""
else
    export COMPOSE_PROFILES=${COMPOSE_PROFILES:-ai-services}
fi

echo "📋 开发环境配置:"
echo "   操作系统: $OS"
echo "   Docker Compose: $DOCKER_COMPOSE"
echo "   MODEL_ID: $MODEL_ID"
echo "   COMPOSE_PROFILES: $COMPOSE_PROFILES"
echo "   跳过TTS: $SKIP_TTS"
echo "   仅基础服务: $BASIC_ONLY"
echo "   热重载: 已启用"

# 执行检查
if [ "$BASIC_ONLY" = false ]; then
    check_external_services
    check_nvidia_support
fi

# 创建必要的目录
echo "📁 创建开发数据目录..."
mkdir -p data/postgres-dev
mkdir -p data/redis-dev
mkdir -p data/minio-dev
mkdir -p tmp
mkdir -p logs

# 停止现有服务
echo "🛑 停止现有开发服务..."
$DOCKER_COMPOSE -f docker-compose.dev.yml down --remove-orphans 2>/dev/null || true

# 构建开发镜像
echo "🔨 构建开发镜像..."
$DOCKER_COMPOSE -f docker-compose.dev.yml build

# 启动基础服务
echo "🔧 启动基础服务 (PostgreSQL, Redis, MinIO)..."
$DOCKER_COMPOSE -f docker-compose.dev.yml up -d postgres redis minio

# 等待基础服务就绪
echo "⏳ 等待基础服务就绪..."
sleep 10

# 检查基础服务健康状态
echo "🔍 检查基础服务健康状态..."
for service in postgres redis minio; do
    echo "   检查 $service..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if $DOCKER_COMPOSE -f docker-compose.dev.yml ps $service | grep -q "healthy\|Up"; then
            echo "   ✅ $service 已就绪"
            break
        fi
        sleep 2
        timeout=$((timeout-2))
    done

    if [ $timeout -le 0 ]; then
        echo "   ❌ $service 启动超时"
        $DOCKER_COMPOSE -f docker-compose.dev.yml logs $service
        exit 1
    fi
done

# 启动TTS服务（如果需要）
if [ "$BASIC_ONLY" = false ] && [ "$SKIP_TTS" = false ] && [[ "$COMPOSE_PROFILES" == *"ai-services"* ]]; then
    echo "🎤 启动TTS服务..."
    $DOCKER_COMPOSE -f docker-compose.dev.yml up -d tts

    echo "⏳ 等待TTS服务启动（这可能需要几分钟）..."
    sleep 30

    echo "🔍 检查TTS服务状态..."
    timeout=180
    tts_ready=false
    while [ $timeout -gt 0 ]; do
        if safe_curl "http://localhost:8000/v2/health/ready" 10; then
            echo "   ✅ TTS服务已就绪"
            tts_ready=true
            break
        fi
        echo "   ⏳ TTS服务启动中... (剩余 ${timeout}s)"
        sleep 10
        timeout=$((timeout-10))
    done

    if [ "$tts_ready" = false ]; then
        echo "   ⚠️  TTS服务启动超时，但将继续启动其他服务"
        echo "   请检查TTS服务日志: $DOCKER_COMPOSE -f docker-compose.dev.yml logs tts"
    fi
fi

# 启动API和前端服务（如果不是仅基础模式）
if [ "$BASIC_ONLY" = false ]; then
    # 启动API开发服务
    echo "🌐 启动API开发服务 (支持热重载)..."
    $DOCKER_COMPOSE -f docker-compose.dev.yml up -d api-dev

    # 等待API服务就绪
    echo "⏳ 等待API服务就绪..."
    sleep 20

    # 检查API服务状态
    echo "🔍 检查API服务状态..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if safe_curl "http://localhost:8080/health" 5; then
            echo "   ✅ API服务已就绪"
            break
        fi
        sleep 3
        timeout=$((timeout-3))
    done

    if [ $timeout -le 0 ]; then
        echo "   ❌ API服务启动超时"
        $DOCKER_COMPOSE -f docker-compose.dev.yml logs api-dev
        exit 1
    fi

    # 启动前端开发服务
    echo "🖥️  启动前端开发服务 (支持热重载)..."
    $DOCKER_COMPOSE -f docker-compose.dev.yml up -d web-dev

    # 等待前端服务就绪
    echo "⏳ 等待前端服务就绪..."
    sleep 25
fi

# 显示服务状态
echo ""
echo "📊 开发服务状态:"
$DOCKER_COMPOSE -f docker-compose.dev.yml ps

echo ""
echo "🎉 VidCraft Studio 开发环境启动完成!"
echo ""

if [ "$BASIC_ONLY" = true ]; then
    echo "📱 基础服务访问地址:"
    echo "   📦 MinIO控制台: http://localhost:9001 (minioadmin/minioadmin123)"
    echo "   🗄️  PostgreSQL: localhost:5432 (vidcraft/vidcraft123)"
    echo "   🔴 Redis: localhost:6379"

    # 检查TTS服务
    if safe_curl "http://localhost:8000/v2/health/ready" 2; then
        echo "   🎤 TTS服务: http://localhost:8000 ✅"
    else
        echo "   🎤 TTS服务: 未启动 ⚠️"
    fi

    echo ""
    echo "🔧 下一步操作:"
    echo "   启动完整环境: $0"
    echo "   启动API服务: $DOCKER_COMPOSE -f docker-compose.dev.yml up -d api-dev"
    echo "   启动前端服务: $DOCKER_COMPOSE -f docker-compose.dev.yml up -d web-dev"
else
    echo "📱 完整服务访问地址:"
    echo "   🌐 前端应用: http://localhost:3000 (支持热重载)"
    echo "   🔌 API服务:  http://localhost:8080 (支持热重载)"
    echo "   📦 MinIO:    http://localhost:9001 (minioadmin/minioadmin123)"

    # 检查TTS服务
    if safe_curl "http://localhost:8000/v2/health/ready" 2; then
        echo "   🎤 TTS服务:  http://localhost:8000 ✅"
    else
        echo "   🎤 TTS服务:  未启动 ⚠️"
    fi

    echo ""
    echo "🎯 功能测试页面:"
    echo "   🎤 TTS测试:  http://localhost:3000/app/tts-test"
    echo "   📝 推文生成: http://localhost:3000/app/generate-tweet"
    echo "   📚 小说生成: http://localhost:3000/app/generate-novel"
    echo ""
    echo "🔥 热重载说明:"
    echo "   - 修改Go代码会自动重新编译和重启API服务"
    echo "   - 修改React代码会自动刷新浏览器"
    echo "   - 修改配置文件需要手动重启对应服务"
fi

echo ""
echo "🔧 开发工具:"
echo "   查看服务状态: $DOCKER_COMPOSE -f docker-compose.dev.yml ps"
echo "   查看API日志: $DOCKER_COMPOSE -f docker-compose.dev.yml logs -f api-dev"
echo "   查看TTS日志: $DOCKER_COMPOSE -f docker-compose.dev.yml logs -f tts"
echo "   进入API容器: $DOCKER_COMPOSE -f docker-compose.dev.yml exec api-dev sh"
echo "   重启API服务: $DOCKER_COMPOSE -f docker-compose.dev.yml restart api-dev"
echo "   停止开发环境: $DOCKER_COMPOSE -f docker-compose.dev.yml down"
echo "   开发工具脚本: ./scripts/dev-tools.sh help"
echo ""
echo "🤖 外部AI服务:"
echo "   Ollama (文本生成): localhost:11434"
echo "   Stable Diffusion (图像生成): localhost:7860"
echo "   Whisper (语音识别): localhost:9000"
echo ""

# 显示实时日志（可选）
if [ "$SHOW_LOGS" = true ]; then
    echo "📋 显示实时日志 (Ctrl+C 退出)..."
    $DOCKER_COMPOSE -f docker-compose.dev.yml logs -f api-dev web-dev
else
    reply=$(read_user_input "是否查看实时日志？(y/N): " "N")
    if [[ $reply =~ ^[Yy]$ ]]; then
        echo "📋 显示实时日志 (Ctrl+C 退出)..."
        $DOCKER_COMPOSE -f docker-compose.dev.yml logs -f api-dev web-dev
    fi
fi

echo "✨ 开发环境已就绪！开始愉快的开发吧！"
