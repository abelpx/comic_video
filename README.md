# VidCraft Studio - AI视频创作平台

## 🎉 项目状态

**✅ 系统已成功部署并运行！**

VidCraft Studio是一个基于AI的视频创作平台，支持小说转视频、图像生成等功能。

## 🚀 快速开始

### 启动服务
```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker ps

# 查看日志
docker-compose logs -f
```

### 停止服务
```bash
# 停止所有服务
docker-compose down

# 停止并清理数据卷
docker-compose down -v
```

## 🌐 访问地址

- **前端应用**: http://localhost:3000
- **API服务**: http://localhost:8080
- **API健康检查**: http://localhost:8080/health
- **MinIO控制台**: http://localhost:9001 (minioadmin/minioadmin123)

## 📊 服务架构

### 核心服务
| 服务名称 | 端口 | 状态 | 说明 |
|---------|------|------|------|
| **VidCraft Web** | 3000 | ✅ 运行中 | React前端应用 |
| **VidCraft API** | 8080 | ✅ 运行中 | Go后端API服务 |
| **PostgreSQL** | 5432 | ✅ 运行中 | 主数据库 |
| **Redis** | 6379 | ✅ 运行中 | 缓存服务 |
| **MinIO** | 9000-9001 | ✅ 运行中 | 对象存储 |

### AI服务集成
| 服务名称 | 端口 | 端点 | 说明 |
|---------|------|------|------|
| **Ollama** | 11434 | `host.docker.internal:11434` | 大语言模型服务 |
| **Stable Diffusion** | 7860 | `host.docker.internal:7860` | 图像生成服务 |

## 🔧 技术栈

### 前端
- **框架**: React 18 + TypeScript
- **UI库**: Ant Design
- **构建工具**: Vite
- **状态管理**: Zustand

### 后端
- **语言**: Go 1.23
- **框架**: Gin
- **数据库**: PostgreSQL 15
- **缓存**: Redis 7
- **存储**: MinIO

### AI集成
- **LLM**: Ollama (DeepSeek-R1)
- **图像生成**: Stable Diffusion WebUI
- **网络**: Docker host.docker.internal

## 📋 API端点

### 认证相关
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/auth/profile` - 获取用户信息
- `POST /api/v1/auth/logout` - 用户登出

### 系统相关
- `GET /health` - 健康检查
- `GET /api/v1/info` - API信息
- `GET /api/v1/ai/quota` - 用户配额

## 🗂️ 项目结构

```
comic_video/
├── cmd/api/                 # API入口
├── internal/               # 内部代码
│   ├── api/               # API路由
│   ├── config/            # 配置
│   ├── domain/            # 领域模型
│   ├── repository/        # 数据访问层
│   ├── service/           # 业务逻辑层
│   └── utils/             # 工具函数
├── web/                   # 前端代码
│   ├── src/              # 源代码
│   ├── dist/             # 构建输出
│   └── Dockerfile        # 前端Docker配置
├── scripts/              # 脚本文件
├── docs/                 # 文档
├── docker-compose.yml    # Docker编排配置
├── Dockerfile           # 后端Docker配置
└── README.md           # 项目说明
```

## 🔍 故障排除

### 常见问题

1. **前端无法访问后端**
   - 检查API服务是否正常运行: `curl http://localhost:8080/health`
   - 检查CORS配置是否正确

2. **AI服务连接失败**
   - 确保Ollama服务运行: `ollama serve`
   - 确保SD WebUI运行: `python launch.py --api --listen`
   - 检查端口11434和7860是否开放

3. **数据库连接问题**
   - 检查PostgreSQL容器状态: `docker logs vidcraft_postgres`
   - 确认数据库配置正确

### 日志查看
```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker logs vidcraft_api
docker logs vidcraft_web
docker logs vidcraft_postgres
```

## 🛠️ 开发指南

### 本地开发

1. **后端开发**
   ```bash
   cd cmd/api
   go run main_simple.go
   ```

2. **前端开发**
   ```bash
   cd web
   npm install
   npm run dev
   ```

### 构建部署
```bash
# 构建所有服务
docker-compose build

# 重新构建并启动
docker-compose up -d --build
```

## 📝 更新日志

### v1.0.0 (2025-07-23)
- ✅ 完成基础架构搭建
- ✅ 实现用户认证系统
- ✅ 集成AI服务(Ollama + Stable Diffusion)
- ✅ 配置Docker容器化部署
- ✅ 修复前后端通信问题
- ✅ 清理无用文件和配置

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

**部署时间**: 2025-07-23  
**系统状态**: 🟢 全部正常运行  
**AI服务**: 🤖 已集成宿主机服务
