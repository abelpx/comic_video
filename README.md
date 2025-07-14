# VidCraft Studio (Video Crafting Platform)

一个功能强大的在线视频编辑和生成平台，支持视频剪辑、特效添加、模板应用等功能。

## 项目概述

VidCraft Studio 是一个基于Web的视频编辑平台，提供专业的视频编辑工具，支持多种视频格式，内置丰富的特效和模板，让用户能够轻松创建高质量的视频内容。

### 核心功能

- 🎬 **视频编辑**: 支持视频剪辑、拼接、分割等基础操作
- ✨ **特效系统**: 内置多种视频特效和转场效果
- 📱 **模板系统**: 提供丰富的视频模板，快速生成专业视频
- 🎨 **素材库**: 海量音乐、图片、视频素材
- 👥 **用户系统**: 用户注册、登录、作品管理
- 🚀 **云端渲染**: 支持云端视频渲染和导出
- 📊 **项目管理**: 项目保存、版本管理、协作编辑

## 技术架构

### 后端架构

```
comic_video/
├── cmd/                    # 应用程序入口
│   ├── api/               # API服务器
│   ├── worker/            # 后台任务处理器
│   └── scheduler/         # 定时任务调度器
├── internal/              # 内部包
│   ├── api/              # API层
│   │   ├── handlers/     # HTTP处理器
│   │   ├── middleware/   # 中间件
│   │   └── routes/       # 路由定义
│   ├── service/          # 业务逻辑层
│   │   ├── auth/         # 认证服务
│   │   ├── video/        # 视频处理服务
│   │   ├── template/     # 模板服务
│   │   ├── user/         # 用户服务
│   │   └── render/       # 渲染服务
│   ├── repository/       # 数据访问层
│   │   ├── postgres/     # PostgreSQL存储
│   │   └── redis/        # Redis缓存
│   ├── domain/           # 领域模型
│   │   ├── entity/       # 实体
│   │   ├── dto/          # 数据传输对象
│   │   └── vo/           # 视图对象
│   ├── config/           # 配置管理
│   ├── utils/            # 工具函数
│   └── pkg/              # 公共包
├── web/                  # 前端代码
│   ├── src/
│   │   ├── components/   # React组件
│   │   ├── pages/        # 页面组件
│   │   ├── services/     # API服务
│   │   ├── store/        # Redux状态管理
│   │   └── utils/        # 工具函数
│   ├── public/           # 静态资源
│   └── package.json
├── scripts/              # 脚本文件
├── docs/                 # 文档
├── docker/               # Docker配置
├── deploy/               # 部署配置
├── go.mod
├── go.sum
└── README.md
```

### 数据库设计

#### 核心表结构

1. **users** - 用户表
2. **projects** - 项目表
3. **videos** - 视频资源表
4. **templates** - 模板表
5. **materials** - 素材表
6. **renders** - 渲染任务表
7. **user_projects** - 用户项目关联表

## 开发环境搭建

### 前置要求

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- FFmpeg
- PostgreSQL 14+
- Redis 6+

### 快速开始

1. **克隆项目**
```bash
git clone <repository-url>
cd comic_video
```

2. **启动依赖服务**
```bash
docker-compose up -d postgres redis minio rabbitmq
```

3. **安装后端依赖**
```bash
go mod download
```

4. **安装前端依赖**
```bash
cd web
npm install
```

5. **运行开发服务器**
```bash
# 后端
go run cmd/api/main.go

# 前端
cd web
npm run dev
```

## API文档

### 认证相关

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/logout` - 用户登出
- `GET /api/v1/auth/profile` - 获取用户信息

### 项目管理

- `GET /api/v1/projects` - 获取项目列表
- `POST /api/v1/projects` - 创建新项目
- `GET /api/v1/projects/{id}` - 获取项目详情
- `PUT /api/v1/projects/{id}` - 更新项目
- `DELETE /api/v1/projects/{id}` - 删除项目

### 视频处理

- `POST /api/v1/videos/upload` - 上传视频
- `GET /api/v1/videos/{id}` - 获取视频信息
- `POST /api/v1/videos/{id}/process` - 处理视频
- `GET /api/v1/videos/{id}/status` - 获取处理状态

### 模板系统

- `GET /api/v1/templates` - 获取模板列表
- `GET /api/v1/templates/{id}` - 获取模板详情
- `POST /api/v1/templates/{id}/apply` - 应用模板

### 渲染服务

- `POST /api/v1/renders` - 创建渲染任务
- `GET /api/v1/renders/{id}` - 获取渲染状态
- `GET /api/v1/renders/{id}/download` - 下载渲染结果

## 部署指南

### Docker部署

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d
```

### 生产环境配置

1. 配置环境变量
2. 设置数据库连接
3. 配置对象存储
4. 设置CDN
5. 配置负载均衡

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

- 项目维护者: [Your Name]
- 邮箱: [your.email@example.com]
- 项目地址: [GitHub Repository URL] 