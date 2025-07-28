-- 添加角色重要性和出场频率字段
-- Migration: 20240127_add_character_importance.sql

-- 为characters表添加新字段
ALTER TABLE characters 
ADD COLUMN importance VARCHAR(50) DEFAULT '配角',
ADD COLUMN screen_time VARCHAR(20) DEFAULT '中';

-- 添加索引以提高查询性能
CREATE INDEX idx_characters_importance ON characters(importance);
CREATE INDEX idx_characters_screen_time ON characters(screen_time);

-- 更新现有数据的默认值
UPDATE characters 
SET importance = '主角' 
WHERE name IN (
    SELECT DISTINCT character_name 
    FROM storyboard_frames 
    GROUP BY character_name 
    HAVING COUNT(*) > 10
);

UPDATE characters 
SET screen_time = '高' 
WHERE importance = '主角';

UPDATE characters 
SET screen_time = '低' 
WHERE importance = '群众角色';

-- 添加约束
ALTER TABLE characters 
ADD CONSTRAINT chk_importance 
CHECK (importance IN ('主角', '重要配角', '次要角色', '群众角色'));

ALTER TABLE characters 
ADD CONSTRAINT chk_screen_time 
CHECK (screen_time IN ('高', '中', '低'));

-- 创建视图：按重要性统计角色
CREATE OR REPLACE VIEW character_importance_stats AS
SELECT 
    project_id,
    importance,
    COUNT(*) as character_count,
    AVG(CASE 
        WHEN screen_time = '高' THEN 3
        WHEN screen_time = '中' THEN 2
        WHEN screen_time = '低' THEN 1
        ELSE 0
    END) as avg_screen_time_score
FROM characters
GROUP BY project_id, importance;

-- 创建函数：获取项目主要角色
CREATE OR REPLACE FUNCTION get_main_characters(p_project_id UUID)
RETURNS TABLE(
    character_id UUID,
    character_name VARCHAR,
    importance VARCHAR,
    screen_time VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        c.id,
        c.name,
        c.importance,
        c.screen_time
    FROM characters c
    WHERE c.project_id = p_project_id
    AND c.importance IN ('主角', '重要配角')
    ORDER BY 
        CASE c.importance
            WHEN '主角' THEN 1
            WHEN '重要配角' THEN 2
            ELSE 3
        END,
        CASE c.screen_time
            WHEN '高' THEN 1
            WHEN '中' THEN 2
            WHEN '低' THEN 3
            ELSE 4
        END,
        c.created_at;
END;
$$ LANGUAGE plpgsql;

-- 添加注释
COMMENT ON COLUMN characters.importance IS '角色重要性：主角/重要配角/次要角色/群众角色';
COMMENT ON COLUMN characters.screen_time IS '预估出场频率：高/中/低';
COMMENT ON VIEW character_importance_stats IS '按重要性统计角色数量和平均出场频率';
COMMENT ON FUNCTION get_main_characters IS '获取项目的主要角色（主角和重要配角）';
