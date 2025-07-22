#!/bin/bash

# VidCraft Studio 工作流测试脚本
# 用于测试完整的AI工作流功能

set -e

API_BASE_URL="http://localhost:8080/api/v1"
TOKEN=""
PROJECT_ID=""
WORKFLOW_ID=""

echo "🧪 VidCraft Studio 工作流测试"
echo "=============================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
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

# 检查API服务是否运行
check_api_service() {
    log_info "检查API服务状态..."
    
    if curl -s "$API_BASE_URL/health" > /dev/null 2>&1; then
        log_success "API服务运行正常"
    else
        log_error "API服务未运行，请先启动后端服务"
        exit 1
    fi
}

# 用户注册和登录
authenticate_user() {
    log_info "进行用户认证..."
    
    # 尝试注册测试用户
    REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "test_user",
            "email": "test@example.com",
            "password": "test123456"
        }' || echo '{"code": 400}')
    
    # 登录获取token
    LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "test@example.com",
            "password": "test123456"
        }')
    
    TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // empty')
    
    if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
        log_success "用户认证成功"
    else
        log_error "用户认证失败"
        echo "Response: $LOGIN_RESPONSE"
        exit 1
    fi
}

# 创建测试项目
create_test_project() {
    log_info "创建测试项目..."
    
    PROJECT_RESPONSE=$(curl -s -X POST "$API_BASE_URL/projects" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d '{
            "name": "AI工作流测试项目",
            "description": "用于测试完整AI工作流的项目",
            "type": "ai_workflow"
        }')
    
    PROJECT_ID=$(echo "$PROJECT_RESPONSE" | jq -r '.data.id // empty')
    
    if [ -n "$PROJECT_ID" ] && [ "$PROJECT_ID" != "null" ]; then
        log_success "测试项目创建成功: $PROJECT_ID"
    else
        log_error "测试项目创建失败"
        echo "Response: $PROJECT_RESPONSE"
        exit 1
    fi
}

# 启动小说转视频工作流
start_novel_to_video_workflow() {
    log_info "启动小说转视频工作流..."
    
    # 测试小说文本
    NOVEL_TEXT="在一个遥远的魔法王国里，住着一位年轻的魔法师艾莉娅。她有着金色的长发和碧绿的眼睛，总是穿着一件深蓝色的魔法袍。

艾莉娅从小就展现出了非凡的魔法天赋，但她的魔法总是会出现一些意想不到的效果。比如，当她想要召唤一只小鸟时，却召唤出了一条彩虹色的小龙。

有一天，王国遭到了黑暗魔法师的威胁。艾莉娅决定踏上冒险之旅，去寻找传说中的光明之石，以拯救她的家园。

在森林深处，她遇到了一只会说话的白兔，白兔告诉她：'真正的力量不在于魔法的强大，而在于内心的勇气和善良。'

经过重重考验，艾莉娅终于找到了光明之石，并成功击败了黑暗魔法师，拯救了王国。从此，她成为了王国最受尊敬的魔法师。"
    
    WORKFLOW_RESPONSE=$(curl -s -X POST "$API_BASE_URL/workflows/novel-to-video" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"project_id\": \"$PROJECT_ID\",
            \"novel_text\": \"$NOVEL_TEXT\",
            \"title\": \"魔法师艾莉娅的冒险\",
            \"description\": \"一个关于年轻魔法师拯救王国的奇幻故事\",
            \"settings\": {
                \"video_theme\": \"fantasy\",
                \"target_audience\": \"young\",
                \"content_type\": \"short\",
                \"auto_publish\": false
            }
        }")
    
    WORKFLOW_ID=$(echo "$WORKFLOW_RESPONSE" | jq -r '.data.workflow_id // empty')
    
    if [ -n "$WORKFLOW_ID" ] && [ "$WORKFLOW_ID" != "null" ]; then
        log_success "工作流启动成功: $WORKFLOW_ID"
    else
        log_error "工作流启动失败"
        echo "Response: $WORKFLOW_RESPONSE"
        exit 1
    fi
}

# 监控工作流进度
monitor_workflow_progress() {
    log_info "监控工作流进度..."
    
    local max_attempts=60  # 最多等待10分钟
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        PROGRESS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/workflows/$WORKFLOW_ID/progress" \
            -H "Authorization: Bearer $TOKEN")
        
        STATUS=$(echo "$PROGRESS_RESPONSE" | jq -r '.data.workflow.status // empty')
        PROGRESS=$(echo "$PROGRESS_RESPONSE" | jq -r '.data.overall_progress // 0')
        CURRENT_STEP=$(echo "$PROGRESS_RESPONSE" | jq -r '.data.workflow.current_step // empty')
        
        log_info "工作流状态: $STATUS, 进度: $PROGRESS%, 当前步骤: $CURRENT_STEP"
        
        case "$STATUS" in
            "completed")
                log_success "工作流执行完成！"
                return 0
                ;;
            "failed")
                log_error "工作流执行失败"
                echo "Progress Response: $PROGRESS_RESPONSE"
                return 1
                ;;
            "cancelled")
                log_warning "工作流被取消"
                return 1
                ;;
        esac
        
        sleep 10
        ((attempt++))
    done
    
    log_warning "工作流监控超时"
    return 1
}

# 获取工作流结果
get_workflow_results() {
    log_info "获取工作流结果..."
    
    WORKFLOW_RESPONSE=$(curl -s -X GET "$API_BASE_URL/workflows/$WORKFLOW_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "工作流详情:"
    echo "$WORKFLOW_RESPONSE" | jq '.'
    
    TASKS_RESPONSE=$(curl -s -X GET "$API_BASE_URL/workflows/$WORKFLOW_ID/tasks" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "工作流任务:"
    echo "$TASKS_RESPONSE" | jq '.'
}

# 清理测试数据
cleanup_test_data() {
    log_info "清理测试数据..."
    
    if [ -n "$PROJECT_ID" ]; then
        curl -s -X DELETE "$API_BASE_URL/projects/$PROJECT_ID" \
            -H "Authorization: Bearer $TOKEN" > /dev/null
        log_success "测试项目已删除"
    fi
}

# 主测试流程
main() {
    log_info "开始工作流测试..."
    
    check_api_service
    authenticate_user
    create_test_project
    start_novel_to_video_workflow
    monitor_workflow_progress
    get_workflow_results
    
    log_success "工作流测试完成！"
}

# 错误处理
trap 'log_error "测试过程中出现错误，正在清理..."; cleanup_test_data; exit 1' ERR

# 运行测试
main

# 询问是否清理测试数据
echo ""
read -p "是否删除测试数据？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    cleanup_test_data
else
    log_info "测试数据保留，项目ID: $PROJECT_ID"
fi

log_success "测试脚本执行完成！"
