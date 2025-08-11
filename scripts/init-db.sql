-- =============================================================================
-- 🗄️ Comic Video 数据库初始化脚本
-- =============================================================================

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建项目表
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    novel_text TEXT NOT NULL,
    style VARCHAR(50) DEFAULT 'anime',
    status VARCHAR(20) DEFAULT 'created',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建角色表
CREATE TABLE IF NOT EXISTS characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    appearance TEXT,
    personality TEXT,
    voice_settings JSONB DEFAULT '{}',
    visual_features JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建场景表
CREATE TABLE IF NOT EXISTS scenes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sequence_number INTEGER NOT NULL,
    title VARCHAR(255),
    description TEXT NOT NULL,
    location VARCHAR(255),
    time_of_day VARCHAR(50),
    characters UUID[] DEFAULT '{}',
    dialogue TEXT,
    visual_elements JSONB DEFAULT '{}',
    audio_elements JSONB DEFAULT '{}',
    duration_seconds INTEGER DEFAULT 5,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, sequence_number)
);

-- 创建生成任务表
CREATE TABLE IF NOT EXISTS generation_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_type VARCHAR(50) NOT NULL, -- 'image', 'audio', 'video', 'complete'
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    progress FLOAT DEFAULT 0.0,
    current_stage VARCHAR(100),
    error_message TEXT,
    result_data JSONB DEFAULT '{}',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建文件表
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    task_id UUID REFERENCES generation_tasks(id) ON DELETE CASCADE,
    file_type VARCHAR(20) NOT NULL, -- 'image', 'audio', 'video', 'text'
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT,
    mime_type VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_characters_project_id ON characters(project_id);
CREATE INDEX IF NOT EXISTS idx_scenes_project_id ON scenes(project_id);
CREATE INDEX IF NOT EXISTS idx_scenes_sequence ON scenes(project_id, sequence_number);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON generation_tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON generation_tasks(status);
CREATE INDEX IF NOT EXISTS idx_files_project_id ON files(project_id);
CREATE INDEX IF NOT EXISTS idx_files_task_id ON files(task_id);
CREATE INDEX IF NOT EXISTS idx_files_type ON files(file_type);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_scenes_updated_at BEFORE UPDATE ON scenes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_generation_tasks_updated_at BEFORE UPDATE ON generation_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入默认系统配置
INSERT INTO system_configs (key, value, description) VALUES
    ('app_version', '"1.0.0"', '应用版本'),
    ('max_upload_size', '104857600', '最大上传文件大小(字节)'),
    ('supported_formats', '["txt", "pdf", "docx"]', '支持的文件格式'),
    ('default_video_settings', '{"resolution": "1920x1080", "fps": 30, "quality": "medium"}', '默认视频设置'),
    ('ai_service_config', '{"openai_model": "gpt-4", "tts_model": "spark-tts", "image_model": "stable-diffusion"}', 'AI服务配置')
ON CONFLICT (key) DO NOTHING;

-- 创建默认管理员用户 (密码: admin123)
INSERT INTO users (username, email, password_hash, role) VALUES
    ('admin', 'admin@comic-video.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin')
ON CONFLICT (username) DO NOTHING;

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE '✅ Comic Video 数据库初始化完成!';
    RAISE NOTICE '📊 创建的表: users, projects, characters, scenes, generation_tasks, files, system_configs';
    RAISE NOTICE '👤 默认管理员: admin / admin123';
    RAISE NOTICE '🔗 数据库连接: postgresql://comic_video:comic_video_password@localhost:5432/comic_video';
END $$;
