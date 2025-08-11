# Comic Video Makefile
# 简化开发和部署流程

.PHONY: help build test clean dev run lint format deps migrate swagger
.PHONY: install setup start stop restart status logs clean-logs

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
APP_NAME := comic-video
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Docker相关
DOCKER_REGISTRY := registry.example.com
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(APP_NAME)
DOCKER_TAG := $(VERSION)

# 数据库相关
DB_HOST := localhost
DB_PORT := 5432
DB_NAME := comic_video
DB_USER := comic_video_user
DB_PASSWORD := comic_video_password

# 应用配置
APP_NAME := comic-video
APP_PORT := 8080
APP_HOST := localhost

# 本地服务配置
TTS_LOCAL_PATH := F:/pycharm/project/Spark-TTS
TTS_PORT := 8000
POSTGRES_PORT := 5432
REDIS_PORT := 6379

## help: 显示帮助信息
help:
	@echo "Comic Video 开发工具"
	@echo ""
	@echo "可用命令:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: 构建应用程序
build:
	@echo "构建应用程序..."
	@go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/api
	@echo "构建完成: bin/$(APP_NAME)"

## build-all: 构建所有平台的二进制文件
build-all:
	@echo "构建多平台二进制文件..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-linux-amd64 ./cmd/api
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-darwin-amd64 ./cmd/api
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-windows-amd64.exe ./cmd/api
	@echo "多平台构建完成"

## test: 运行测试
test:
	@echo "运行测试..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "测试完成，覆盖率报告: coverage.html"

## test-unit: 运行单元测试
test-unit:
	@echo "运行单元测试..."
	@go test -v -short ./...

## test-integration: 运行集成测试
test-integration:
	@echo "运行集成测试..."
	@go test -v -tags=integration ./...

## benchmark: 运行基准测试
benchmark:
	@echo "运行基准测试..."
	@go test -bench=. -benchmem ./...

## clean: 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/ dist/ coverage.out coverage.html
	@go clean -cache -testcache -modcache
	@echo "清理完成"

## setup: 初始化开发环境
setup:
	@echo "🛠️ 初始化开发环境..."
	@echo "📋 检查依赖..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go 未安装"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker 未安装"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose 未安装"; exit 1; }
	@echo "✅ 依赖检查完成"
	@echo "📝 设置环境配置..."
	@if [ ! -f .env ]; then \
		cp .env.dev .env; \
		echo "✅ 已创建 .env 文件"; \
	else \
		echo "ℹ️ .env 文件已存在"; \
	fi
	@echo "📁 创建必要目录..."
	@mkdir -p logs/{app,postgres,redis,minio}
	@mkdir -p output/{videos,images,audio}
	@mkdir -p temp
	@mkdir -p config
	@echo "📦 安装Go依赖..."
	@go mod tidy
	@echo "🎉 开发环境初始化完成!"

## install: 安装依赖
install:
	@echo "📦 安装项目依赖..."
	@go mod download
	@go mod tidy
	@echo "✅ 依赖安装完成"

## start-services: 启动基础设施服务
start-services:
	@echo "🚀 启动基础设施服务..."
	@docker-compose up -d
	@echo "⏳ 等待服务启动..."
	@sleep 10
	@make status
	@echo "🎉 基础设施服务已启动!"
	@echo ""
	@echo "📊 服务地址:"
	@echo "  🗄️ PostgreSQL:   localhost:5432"
	@echo "  🔴 Redis:        localhost:6379"
	@echo "  📁 MinIO API:    http://localhost:9000"
	@echo "  🌐 MinIO控制台:   http://localhost:9001"
	@echo ""
	@echo "📝 查看日志: make logs"
	@echo "🛑 停止服务: make stop-services"

## dev: 启动开发服务器
dev: start-services
	@echo "🚀 启动开发服务器..."
	@echo "📋 检查基础设施服务..."
	@make status
	@echo "🏃 启动Go应用..."
	@go run cmd/server/main.go

## run: 运行应用
run:
	@echo "🏃 运行应用..."
	@go run cmd/server/main.go

## dev-minimal: 启动最小开发环境(仅核心服务)
dev-minimal: dev-setup
	@echo "🚀 启动最小开发环境..."
	@docker-compose -f docker-compose.dev.yml up -d postgres-dev redis-dev minio-dev vidcraft-api vidcraft-web
	@echo "✅ 最小开发环境已启动!"
	@echo ""
	@echo "📊 服务地址:"
	@echo "  🌐 API服务:      http://localhost:8080"
	@echo "  🎨 前端界面:     http://localhost:3000"
	@echo "  🗄️ MinIO控制台:  http://localhost:9001"

## dev-ai: 启动AI服务环境
dev-ai:
	@echo "🤖 启动AI服务环境..."
	@echo "📋 检查TTS目录..."
	@if [ ! -d "$(TTS_LOCAL_PATH)" ]; then \
		echo "❌ TTS目录不存在: $(TTS_LOCAL_PATH)"; \
		echo "请确保Spark-TTS项目已克隆到指定路径"; \
		exit 1; \
	fi
	@docker-compose -f docker-compose.dev.yml --profile ai-services up -d tts ollama-dev
	@echo "⏳ 等待AI服务启动..."
	@sleep 15
	@make ai-services-check
	@echo "✅ AI服务环境已启动!"

## stop-services: 停止基础设施服务
stop-services:
	@echo "🛑 停止基础设施服务..."
	@docker-compose down
	@echo "✅ 基础设施服务已停止"

## restart-services: 重启基础设施服务
restart-services:
	@echo "🔄 重启基础设施服务..."
	@docker-compose restart
	@echo "✅ 基础设施服务已重启"

## logs: 查看服务日志
logs:
	@echo "📋 查看服务日志..."
	@docker-compose logs -f

## status: 检查服务状态
status:
	@echo "📊 检查服务状态..."
	@echo ""
	@echo "🐳 Docker容器状态:"
	@docker-compose ps
	@echo ""
	@echo "🔗 服务连接测试:"
	@echo -n "  PostgreSQL: "
	@if pg_isready -h localhost -p 5432 -U comic_video > /dev/null 2>&1; then \
		echo "✅ 连接正常"; \
	else \
		echo "❌ 连接失败"; \
	fi
	@echo -n "  Redis: "
	@if redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then \
		echo "✅ 连接正常"; \
	else \
		echo "❌ 连接失败"; \
	fi
	@echo -n "  MinIO: "
	@if curl -f -s http://localhost:9000/minio/health/live > /dev/null 2>&1; then \
		echo "✅ 连接正常"; \
	else \
		echo "❌ 连接失败"; \
	fi

## clean-services: 清理服务数据
clean-services:
	@echo "🧹 清理服务数据..."
	@docker-compose down -v --remove-orphans
	@docker system prune -f
	@echo "✅ 服务数据清理完成"

## prod: 启动生产环境
prod:
	@echo "启动生产环境..."
	@docker-compose up -d
	@echo "生产环境已启动"

## prod-stop: 停止生产环境
prod-stop:
	@echo "停止生产环境..."
	@docker-compose down
	@echo "生产环境已停止"

## docker-build: 构建Docker镜像
docker-build:
	@echo "构建Docker镜像..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "Docker镜像构建完成"

## docker-push: 推送Docker镜像
docker-push: docker-build
	@echo "推送Docker镜像..."
	@docker tag $(APP_NAME):$(VERSION) $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker tag $(APP_NAME):$(VERSION) $(DOCKER_IMAGE):latest
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_IMAGE):latest
	@echo "Docker镜像推送完成"

## lint: 代码检查
lint:
	@echo "运行代码检查..."
	@golangci-lint run
	@echo "代码检查完成"

## format: 格式化代码
format:
	@echo "格式化代码..."
	@go fmt ./...
	@goimports -w .
	@echo "代码格式化完成"

## deps: 更新依赖
deps:
	@echo "更新Go依赖..."
	@go mod tidy
	@go mod download
	@echo "更新前端依赖..."
	@cd web && npm install
	@echo "依赖更新完成"

## deps-upgrade: 升级依赖
deps-upgrade:
	@echo "升级Go依赖..."
	@go get -u ./...
	@go mod tidy
	@echo "升级前端依赖..."
	@cd web && npm update
	@echo "依赖升级完成"

## migrate-up: 执行数据库迁移
migrate-up:
	@echo "执行数据库迁移..."
	@migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up
	@echo "数据库迁移完成"

## migrate-down: 回滚数据库迁移
migrate-down:
	@echo "回滚数据库迁移..."
	@migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down 1
	@echo "数据库回滚完成"

## migrate-create: 创建新的迁移文件
migrate-create:
	@read -p "请输入迁移文件名: " name; \
	migrate create -ext sql -dir migrations $$name
	@echo "迁移文件创建完成"

## swagger: 生成API文档
swagger:
	@echo "生成Swagger文档..."
	@swag init -g cmd/api/main.go -o docs/swagger
	@echo "API文档生成完成"

## install-tools: 安装开发工具
install-tools:
	@echo "安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/air-verse/air@latest
	@echo "开发工具安装完成"

## security: 安全检查
security:
	@echo "运行安全检查..."
	@gosec ./...
	@echo "安全检查完成"

## performance: 性能分析
performance:
	@echo "运行性能分析..."
	@go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "性能分析完成，查看结果:"
	@echo "  go tool pprof cpu.prof"
	@echo "  go tool pprof mem.prof"

## logs: 查看应用日志
logs:
	@echo "查看应用日志..."
	@docker-compose logs -f vidcraft-api

## logs-web: 查看前端日志
logs-web:
	@echo "查看前端日志..."
	@docker-compose logs -f vidcraft-web

## db-backup: 备份数据库
db-backup:
	@echo "备份数据库..."
	@mkdir -p backups
	@pg_dump -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) > backups/$(APP_NAME)-$(shell date +%Y%m%d_%H%M%S).sql
	@echo "数据库备份完成"

## db-restore: 恢复数据库
db-restore:
	@echo "恢复数据库..."
	@read -p "请输入备份文件路径: " file; \
	psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) < $$file
	@echo "数据库恢复完成"





## health: 检查服务健康状态
health:
	@echo "🏥 检查服务健康状态..."
	@echo ""
	@echo "🌐 核心服务:"
	@if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then \
		echo "  ✅ API服务运行正常 (http://localhost:8080)"; \
	else \
		echo "  ❌ API服务不可用 (http://localhost:8080)"; \
	fi
	@if curl -f -s http://localhost:3000 > /dev/null 2>&1; then \
		echo "  ✅ 前端服务运行正常 (http://localhost:3000)"; \
	else \
		echo "  ❌ 前端服务不可用 (http://localhost:3000)"; \
	fi
	@echo ""
	@make ai-services-check

## version: 显示版本信息
version:
	@echo "应用版本信息:"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)"
	@echo "  Go版本: $(GO_VERSION)"

## release: 创建发布版本
release:
	@echo "创建发布版本..."
	@read -p "请输入版本号 (例如: v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	git push origin $$version
	@echo "发布版本创建完成"

## setup: 初始化项目环境
setup: install-tools deps
	@echo "初始化项目环境..."
	@cp .env.example .env
	@echo "请编辑 .env 文件配置环境变量"
	@echo "然后运行: make dev"

## ci: CI/CD流水线
ci: lint test build
	@echo "CI/CD流水线执行完成"

## deploy: 部署到生产环境
deploy: docker-build docker-push
	@echo "部署到生产环境..."
	@echo "请在生产服务器上运行: docker-compose pull && docker-compose up -d"

# 内部目标（不显示在help中）
.PHONY: _check-tools
_check-tools:
	@command -v go >/dev/null 2>&1 || { echo "请安装Go"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "请安装Docker"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "请安装Docker Compose"; exit 1; }
