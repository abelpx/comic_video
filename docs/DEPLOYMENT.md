# VidCraft Studio 部署指南

## 概述

本文档介绍如何部署 VidCraft Studio 项目到不同环境。

## 环境要求

### 系统要求

- **操作系统**: Linux (Ubuntu 20.04+ / CentOS 8+)
- **内存**: 最少 4GB，推荐 8GB+
- **存储**: 最少 50GB 可用空间
- **CPU**: 最少 2 核，推荐 4 核+

### 软件要求

- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **FFmpeg**: 4.4+
- **PostgreSQL**: 14+
- **Redis**: 6+
- **MinIO**: 最新版本
- **RabbitMQ**: 3.8+

## 快速部署

### 1. 克隆项目

```bash
git clone <repository-url>
cd comic_video
```

### 2. 配置环境变量

```bash
cp env.example .env
# 编辑 .env 文件，配置数据库、Redis等连接信息
```

### 3. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 4. 初始化数据库

```bash
# 连接到PostgreSQL容器
docker-compose exec postgres psql -U comic_video_user -d comic_video

# 执行迁移脚本
\i /scripts/migrations/001_initial_schema.sql
```

### 5. 访问服务

- **前端应用**: http://localhost:3000
- **API服务**: http://localhost:8080
- **MinIO控制台**: http://localhost:9001
- **RabbitMQ管理**: http://localhost:15672

## 生产环境部署

### 1. 服务器准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 安装FFmpeg
sudo apt install ffmpeg -y
```

### 2. 项目部署

```bash
# 创建项目目录
sudo mkdir -p /opt/comic_video
cd /opt/comic_video

# 克隆项目
git clone <repository-url> .

# 配置环境变量
cp env.example .env
nano .env
```

### 3. 生产环境配置

编辑 `.env` 文件，配置生产环境参数：

```bash
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=release

# 数据库配置（使用外部数据库）
DB_HOST=your-db-host
DB_PORT=5432
DB_NAME=comic_video
DB_USER=comic_video_user
DB_PASSWORD=strong-password
DB_SSLMODE=require

# Redis配置（使用外部Redis）
REDIS_HOST=your-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=strong-password
REDIS_DB=0

# MinIO配置（使用外部存储）
MINIO_ENDPOINT=your-minio-endpoint
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
MINIO_USE_SSL=true
MINIO_BUCKET_NAME=comic-video

# JWT配置
JWT_SECRET_KEY=your-very-secure-secret-key
JWT_EXPIRE=86400

# RabbitMQ配置
RABBITMQ_URL=amqp://user:password@your-rabbitmq-host:5672/

# 文件上传配置
MAX_FILE_SIZE=500MB
UPLOAD_PATH=/data/uploads
TEMP_PATH=/data/temp
OUTPUT_PATH=/data/output

# 日志配置
LOG_LEVEL=info
LOG_FILE=/var/log/comic_video/app.log

# 跨域配置
CORS_ALLOWED_ORIGINS=https://your-domain.com
```

### 4. 创建数据目录

```bash
# 创建数据目录
sudo mkdir -p /data/{uploads,temp,output,logs}
sudo chown -R 1001:1001 /data
```

### 5. 启动服务

```bash
# 构建并启动服务
docker-compose -f docker-compose.prod.yml up -d

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps
```

### 6. 配置Nginx反向代理

创建Nginx配置文件 `/etc/nginx/sites-available/comic_video`：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;

    # 前端应用
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API服务
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket
    location /ws/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 文件上传大小限制
    client_max_body_size 500M;
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/comic_video /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 7. 配置SSL证书

使用Let's Encrypt获取免费SSL证书：

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d your-domain.com
```

## 监控和日志

### 1. 日志管理

```bash
# 查看应用日志
docker-compose logs -f api
docker-compose logs -f worker

# 查看系统日志
sudo journalctl -u docker.service -f
```

### 2. 监控配置

创建监控脚本 `monitor.sh`：

```bash
#!/bin/bash

# 检查服务状态
check_service() {
    local service=$1
    if ! docker-compose ps $service | grep -q "Up"; then
        echo "$(date): $service is down, restarting..."
        docker-compose restart $service
    fi
}

# 检查磁盘空间
check_disk() {
    local usage=$(df /data | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ $usage -gt 80 ]; then
        echo "$(date): Disk usage is high: ${usage}%"
    fi
}

# 检查内存使用
check_memory() {
    local usage=$(free | grep Mem | awk '{printf("%.0f", $3/$2 * 100.0)}')
    if [ $usage -gt 80 ]; then
        echo "$(date): Memory usage is high: ${usage}%"
    fi
}

# 主循环
while true; do
    check_service api
    check_service worker
    check_disk
    check_memory
    sleep 300
done
```

### 3. 备份策略

创建备份脚本 `backup.sh`：

```bash
#!/bin/bash

BACKUP_DIR="/backup/$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# 备份数据库
docker-compose exec -T postgres pg_dump -U comic_video_user comic_video > $BACKUP_DIR/database.sql

# 备份上传文件
tar -czf $BACKUP_DIR/uploads.tar.gz /data/uploads

# 清理旧备份（保留7天）
find /backup -type d -mtime +7 -exec rm -rf {} \;
```

## 性能优化

### 1. 数据库优化

```sql
-- 创建索引
CREATE INDEX CONCURRENTLY idx_projects_user_status ON projects(user_id, status);
CREATE INDEX CONCURRENTLY idx_videos_user_type ON videos(user_id, type);
CREATE INDEX CONCURRENTLY idx_renders_user_status ON renders(user_id, status);

-- 优化查询
ANALYZE;
```

### 2. Redis优化

```bash
# 配置Redis
echo "maxmemory 1gb" >> /etc/redis/redis.conf
echo "maxmemory-policy allkeys-lru" >> /etc/redis/redis.conf
```

### 3. 视频处理优化

```bash
# 配置FFmpeg参数
export FFMPEG_THREADS=4
export FFMPEG_CPU_PRESET=medium
```

## 故障排除

### 常见问题

1. **服务无法启动**
   ```bash
   # 检查端口占用
   sudo netstat -tlnp | grep :8080
   
   # 检查日志
   docker-compose logs api
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker-compose exec postgres pg_isready
   
   # 检查连接配置
   docker-compose exec api env | grep DB_
   ```

3. **文件上传失败**
   ```bash
   # 检查存储权限
   ls -la /data/uploads
   
   # 检查磁盘空间
   df -h /data
   ```

4. **渲染任务失败**
   ```bash
   # 检查FFmpeg
   docker-compose exec api ffmpeg -version
   
   # 检查工作队列
   docker-compose exec rabbitmq rabbitmqctl list_queues
   ```

### 性能调优

1. **增加并发处理能力**
   ```bash
   # 启动多个worker实例
   docker-compose up -d --scale worker=3
   ```

2. **优化内存使用**
   ```bash
   # 限制容器内存
   docker-compose exec api sh -c "echo 'vm.max_map_count=262144' >> /etc/sysctl.conf"
   ```

3. **配置CDN**
   - 将静态资源上传到CDN
   - 配置MinIO作为CDN源站

## 安全配置

### 1. 网络安全

```bash
# 配置防火墙
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. 容器安全

```bash
# 使用非root用户运行容器
# 在Dockerfile中已配置

# 定期更新镜像
docker-compose pull
docker-compose up -d
```

### 3. 数据安全

```bash
# 加密敏感数据
# 使用环境变量存储密钥
# 定期备份数据
```

## 扩展部署

### 1. 负载均衡

使用HAProxy或Nginx进行负载均衡：

```nginx
upstream api_servers {
    server 192.168.1.10:8080;
    server 192.168.1.11:8080;
    server 192.168.1.12:8080;
}

server {
    listen 80;
    location /api/ {
        proxy_pass http://api_servers;
    }
}
```

### 2. 集群部署

使用Kubernetes进行集群部署：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: comic-video-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: comic-video-api
  template:
    metadata:
      labels:
        app: comic-video-api
    spec:
      containers:
      - name: api
        image: comic-video/api:latest
        ports:
        - containerPort: 8080
```

### 3. 微服务架构

将单体应用拆分为微服务：

- 用户服务 (User Service)
- 项目服务 (Project Service)
- 视频服务 (Video Service)
- 渲染服务 (Render Service)
- 模板服务 (Template Service)

## 维护和更新

### 1. 定期维护

```bash
# 清理日志
docker system prune -f

# 更新依赖
go mod tidy
npm update

# 安全更新
docker-compose pull
```

### 2. 版本升级

```bash
# 备份当前版本
docker-compose down
cp -r /opt/comic_video /opt/comic_video_backup

# 更新代码
git pull origin main

# 重新构建
docker-compose build --no-cache
docker-compose up -d

# 验证升级
curl http://localhost:8080/health
```

### 3. 回滚策略

```bash
# 快速回滚
docker-compose down
cd /opt/comic_video_backup
docker-compose up -d
``` 