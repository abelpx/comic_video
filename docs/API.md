# VidCraft Studio API 文档

## 概述

VidCraft Studio API 提供视频编辑、模板管理、用户认证等功能。所有 API 都遵循 RESTful 设计原则。

### 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: JWT Bearer Token

### 响应格式

所有 API 响应都遵循以下格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 认证相关

### 用户注册

**POST** `/auth/register`

注册新用户账户。

**请求参数：**

```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "nickname": "测试用户"
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "id": "uuid",
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "测试用户",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 用户登录

**POST** `/auth/login`

用户登录获取访问令牌。

**请求参数：**

```json
{
  "username": "testuser",
  "password": "password123"
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid",
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "测试用户"
    }
  }
}
```

### 获取用户信息

**GET** `/auth/profile`

获取当前登录用户信息。

**请求头：**

```
Authorization: Bearer <token>
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "测试用户",
    "avatar": "https://example.com/avatar.jpg",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

## 项目管理

### 获取项目列表

**GET** `/projects`

获取当前用户的项目列表。

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |
| status | string | 否 | 项目状态过滤 |

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "uuid",
        "name": "我的第一个项目",
        "description": "项目描述",
        "status": "draft",
        "thumbnail": "https://example.com/thumb.jpg",
        "duration": 60,
        "resolution": "1920x1080",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "size": 10
  }
}
```

### 创建项目

**POST** `/projects`

创建新项目。

**请求参数：**

```json
{
  "name": "新项目",
  "description": "项目描述",
  "template_id": "uuid",
  "config": "{}"
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "项目创建成功",
  "data": {
    "id": "uuid",
    "name": "新项目",
    "description": "项目描述",
    "status": "draft",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 获取项目详情

**GET** `/projects/{id}`

获取指定项目的详细信息。

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "name": "项目名称",
    "description": "项目描述",
    "config": "{}",
    "status": "draft",
    "thumbnail": "https://example.com/thumb.jpg",
    "duration": 60,
    "resolution": "1920x1080",
    "template": {
      "id": "uuid",
      "name": "模板名称"
    },
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 更新项目

**PUT** `/projects/{id}`

更新项目信息。

**请求参数：**

```json
{
  "name": "更新后的项目名称",
  "description": "更新后的描述",
  "config": "{}"
}
```

### 删除项目

**DELETE** `/projects/{id}`

删除指定项目。

## 项目分享

### 创建项目分享

**POST** `/api/v1/projects/{id}/share`

请求参数：
```json
{
  "expires_at": "2024-12-31T23:59:59Z", // 可选，过期时间
  "password": "123456"                  // 可选，访问密码
}
```

响应示例：
```json
{
  "code": 200,
  "message": "项目分享成功",
  "data": {
    "share_url": "/share/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "token": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "expires_at": "2024-12-31T23:59:59Z"
  }
}
```

---

### 校验项目分享

**POST** `/api/v1/share/{token}/check`

请求参数：
```json
{
  "password": "123456" // 若分享设置了密码则必填
}
```

响应示例：
```json
{
  "code": 200,
  "message": "分享校验成功",
  "data": {
    "id": "项目ID",
    "name": "项目名称",
    "description": "项目描述",
    ... // 其余项目信息
  }
}
```

---

### 取消/失效项目分享

**POST** `/api/v1/share/{share_id}/cancel`

需登录，参数：无

响应示例：
```json
{
  "code": 200,
  "message": "分享已取消/失效",
  "data": null
}
```

## 视频管理

### 上传视频

**POST** `/videos/upload`

上传视频文件。

**请求参数：**

- `file`: 视频文件（multipart/form-data）
- `project_id`: 项目ID（可选）

**响应示例：**

```json
{
  "code": 200,
  "message": "上传成功",
  "data": {
    "id": "uuid",
    "file_name": "video.mp4",
    "original_name": "我的视频.mp4",
    "file_size": 1024000,
    "duration": 30.5,
    "width": 1920,
    "height": 1080,
    "format": "mp4",
    "status": "uploading"
  }
}
```

### 获取视频信息

**GET** `/videos/{id}`

获取视频详细信息。

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "file_name": "video.mp4",
    "original_name": "我的视频.mp4",
    "file_size": 1024000,
    "duration": 30.5,
    "width": 1920,
    "height": 1080,
    "format": "mp4",
    "codec": "h264",
    "bitrate": 2000000,
    "fps": 30.0,
    "thumbnail": "https://example.com/thumb.jpg",
    "status": "ready"
  }
}
```

### 获取视频列表

**GET** `/videos`

获取用户的视频列表。

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| size | int | 否 | 每页数量 |
| type | string | 否 | 视频类型 |
| status | string | 否 | 状态过滤 |

## 模板管理

### 获取模板列表

**GET** `/templates`

获取模板列表。

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| category | string | 否 | 分类过滤 |
| is_public | bool | 否 | 是否公开 |
| page | int | 否 | 页码 |
| size | int | 否 | 每页数量 |

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "uuid",
        "name": "模板名称",
        "description": "模板描述",
        "category": "business",
        "thumbnail": "https://example.com/thumb.jpg",
        "preview": "https://example.com/preview.mp4",
        "duration": 30,
        "resolution": "1920x1080",
        "is_premium": false,
        "download_count": 100,
        "rating": 4.5
      }
    ],
    "total": 1,
    "page": 1,
    "size": 10
  }
}
```

### 获取模板详情

**GET** `/templates/{id}`

获取模板详细信息。

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "name": "模板名称",
    "description": "模板描述",
    "category": "business",
    "thumbnail": "https://example.com/thumb.jpg",
    "preview": "https://example.com/preview.mp4",
    "config": "{}",
    "duration": 30,
    "resolution": "1920x1080",
    "tags": "business,corporate,professional",
    "is_premium": false,
    "download_count": 100,
    "rating": 4.5
  }
}
```

### 应用模板

**POST** `/templates/{id}/apply`

将模板应用到项目。

**请求参数：**

```json
{
  "project_id": "uuid"
}
```

## 渲染服务

### 创建渲染任务

**POST** `/renders`

创建视频渲染任务。

**请求参数：**

```json
{
  "project_id": "uuid",
  "name": "渲染任务名称",
  "quality": "high",
  "format": "mp4",
  "resolution": "1920x1080"
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "渲染任务创建成功",
  "data": {
    "id": "uuid",
    "name": "渲染任务名称",
    "status": "pending",
    "progress": 0,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 获取渲染状态

**GET** `/renders/{id}`

获取渲染任务状态。

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "name": "渲染任务名称",
    "status": "processing",
    "progress": 50,
    "output_path": "https://example.com/output.mp4",
    "output_size": 2048000,
    "duration": 30.5,
    "resolution": "1920x1080",
    "format": "mp4",
    "quality": "high",
    "started_at": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 获取渲染列表

**GET** `/renders`

获取用户的渲染任务列表。

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| size | int | 否 | 每页数量 |
| status | string | 否 | 状态过滤 |

### 下载渲染结果

**GET** `/renders/{id}/download`

下载渲染完成的视频文件。

## 素材管理

### 获取素材列表

**GET** `/materials`

获取素材列表。

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| category | string | 否 | 分类过滤 |
| type | string | 否 | 类型过滤 |
| page | int | 否 | 页码 |
| size | int | 否 | 每页数量 |

### 上传素材

**POST** `/materials/upload`

上传素材文件。

**请求参数：**

- `file`: 素材文件
- `name`: 素材名称
- `description`: 描述
- `category`: 分类
- `type`: 类型
- `tags`: 标签

## WebSocket API

### 实时状态更新

**WebSocket** `/ws/status`

用于实时获取渲染进度、上传状态等。

**连接参数：**

```
ws://localhost:8080/api/v1/ws/status?token=<jwt_token>
```

**消息格式：**

```json
{
  "type": "render_progress",
  "data": {
    "render_id": "uuid",
    "progress": 75,
    "status": "processing"
  }
}
```

## 项目配置（Project.Config）支持特效/滤镜/转场

### 结构说明

- `tracks`：轨道数组，每个轨道可为 video/audio/image 类型。
- `clips`：每个轨道下的素材片段。
- `effects`：每个 clip 可选，数组，支持多种滤镜/特效/转场。
  - `type`："filter"（滤镜）、"transition"（转场）、"effect"（特效，预留）。
  - `name`：滤镜/转场/特效名称。
  - `params`：参数对象（如淡入淡出时可指定 position、duration）。

### 支持的滤镜/转场

- 滤镜：
  - `grayscale`（黑白）
  - `boxblur`（模糊）
  - `negate`（反色）
- 转场：
  - `fade`（淡入淡出），参数：
    - `position`: "in" | "out"（淡入/淡出）
    - `duration`: 持续时间（秒）

### 示例 config

```json
{
  "tracks": [
    {
      "type": "video",
      "clips": [
        {
          "material_id": "素材ID1",
          "start": 0,
          "end": 5,
          "effects": [
            { "type": "filter", "name": "grayscale" },
            { "type": "transition", "name": "fade", "params": { "position": "in", "duration": 1 } }
          ]
        },
        {
          "material_id": "素材ID2",
          "start": 0,
          "end": 8,
          "effects": [
            { "type": "filter", "name": "boxblur" },
            { "type": "transition", "name": "fade", "params": { "position": "out", "duration": 1 } }
          ]
        }
      ]
    }
  ],
  "resolution": "1920x1080",
  "frame_rate": 30
}
```

### 使用建议

- 前端可为每个 clip 提供“添加滤镜/转场”功能，自动拼入 effects 字段。
- 后端自动解析并渲染，无需前端关心 FFmpeg 细节。

## 错误处理

### 错误响应格式

```json
{
  "code": 400,
  "message": "参数错误",
  "errors": [
    {
      "field": "username",
      "message": "用户名不能为空"
    }
  ],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 常见错误码

| 错误码 | 说明 |
|--------|------|
| 1001 | 用户名已存在 |
| 1002 | 邮箱已存在 |
| 1003 | 用户名或密码错误 |
| 1004 | 文件上传失败 |
| 1005 | 文件格式不支持 |
| 1006 | 文件大小超限 |
| 1007 | 渲染任务创建失败 |
| 1008 | 项目不存在 |
| 1009 | 权限不足 |
| 1010 | 资源不存在 | 