package script

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"comic_video/internal/service/ai"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service 剧本服务
type Service struct {
	db         *gorm.DB
	scriptRepo *postgres.ScriptRepository
	sceneRepo  *postgres.ScriptSceneRepository
	aiService  *ai.Service
}

// NewService 创建剧本服务
func NewService(db *gorm.DB, aiService *ai.Service) *Service {
	return &Service{
		db:         db,
		scriptRepo: postgres.NewScriptRepository(db),
		sceneRepo:  postgres.NewScriptSceneRepository(db),
		aiService:  aiService,
	}
}

// AdaptNovelToScript 将小说改编为剧本
func (s *Service) AdaptNovelToScript(ctx context.Context, req *AdaptScriptRequest) (*entity.Script, error) {
	log.Printf("[ScriptService] 开始改编小说为剧本: project=%s", req.ProjectID)

	// 1. 分析小说结构
	analysis, err := s.analyzeNovelStructure(ctx, req.NovelText)
	if err != nil {
		return nil, fmt.Errorf("分析小说结构失败: %w", err)
	}

	// 2. 生成剧本内容
	scriptContent, err := s.generateScriptContent(ctx, req.NovelText, analysis)
	if err != nil {
		return nil, fmt.Errorf("生成剧本内容失败: %w", err)
	}

	// 3. 创建剧本实体
	projectUUID, _ := uuid.Parse(req.ProjectID)
	script := &entity.Script{
		ProjectID:  projectUUID,
		Title:      req.Title,
		Type:       entity.ScriptTypeAdapted,
		Content:    scriptContent,
		SourceText: req.NovelText,
		Metadata:   s.buildMetadata(analysis),
		Version:    1,
		Status:     "draft",
	}

	if err := s.scriptRepo.Create(ctx, script); err != nil {
		return nil, fmt.Errorf("保存剧本失败: %w", err)
	}

	// 4. 解析并创建场景
	scenes, err := s.parseScriptScenes(ctx, script.ID, scriptContent)
	if err != nil {
		log.Printf("[ScriptService] 解析场景失败: %v", err)
		// 不阻断流程，继续执行
	} else {
		for _, scene := range scenes {
			if err := s.sceneRepo.Create(ctx, scene); err != nil {
				log.Printf("[ScriptService] 保存场景失败: %v", err)
			}
		}
	}

	log.Printf("[ScriptService] 剧本改编完成: script=%s, scenes=%d", script.ID, len(scenes))
	return script, nil
}

// analyzeNovelStructure 分析小说结构
func (s *Service) analyzeNovelStructure(ctx context.Context, novelText string) (*NovelAnalysis, error) {
	prompt := fmt.Sprintf(`请分析以下小说的结构，提取关键信息：

要求：
1. 识别所有重要角色（包括主角、配角、重要的次要角色）
2. 分析故事情节结构
3. 识别主要场景和地点
4. 分析对话和叙述比例
5. 确定故事的时间线

输出JSON格式：
{
  "characters": [{"name": "角色名", "role": "角色定位", "importance": "主要/次要"}],
  "plot_structure": {
    "beginning": "开头概述",
    "development": "发展概述", 
    "climax": "高潮概述",
    "ending": "结尾概述"
  },
  "scenes": [{"location": "地点", "description": "场景描述"}],
  "dialogue_ratio": 0.3,
  "timeline": "时间线描述",
  "themes": ["主题1", "主题2"],
  "tone": "整体基调"
}

小说内容：
%s`, novelText)

	response, err := s.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析JSON响应
	var analysis NovelAnalysis
	cleanedResponse := s.cleanJSONResponse(response)
	if err := json.Unmarshal([]byte(cleanedResponse), &analysis); err != nil {
		return nil, fmt.Errorf("解析分析结果失败: %w", err)
	}

	return &analysis, nil
}

// generateScriptContent 生成剧本内容
func (s *Service) generateScriptContent(ctx context.Context, novelText string, analysis *NovelAnalysis) (string, error) {
	prompt := fmt.Sprintf(`基于以下小说和分析结果，改编为标准剧本格式：

改编要求：
1. 保持原作核心情节和人物关系
2. 增强对话的戏剧性和可视化效果
3. 添加必要的场景描述和动作指导
4. 适合视频制作的节奏和结构
5. 每个场景包含：场景标题、地点、时间、角色、对话、动作

剧本格式：
场景一：[场景标题]
地点：[具体地点]
时间：[具体时间]
角色：[出场角色]

[场景描述]

角色A：[对话内容]
（动作：[动作描述]）

角色B：[对话内容]
（动作：[动作描述]）

[转场说明]

小说原文：
%s

分析结果：
%s

请生成完整的剧本：`, novelText, s.formatAnalysisForPrompt(analysis))

	response, err := s.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("生成剧本失败: %w", err)
	}

	return response, nil
}

// parseScriptScenes 解析剧本场景
func (s *Service) parseScriptScenes(ctx context.Context, scriptID uuid.UUID, scriptContent string) ([]*entity.ScriptScene, error) {
	var scenes []*entity.ScriptScene
	sceneNumber := 1

	// 简单的场景分割逻辑（可以后续优化）
	lines := strings.Split(scriptContent, "\n")
	var currentScene *entity.ScriptScene
	var sceneContent strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检测场景开始
		if strings.HasPrefix(line, "场景") || strings.HasPrefix(line, "SCENE") {
			// 保存上一个场景
			if currentScene != nil {
				currentScene.Description = sceneContent.String()
				scenes = append(scenes, currentScene)
			}

			// 创建新场景
			currentScene = &entity.ScriptScene{
				ScriptID:    scriptID,
				SceneNumber: sceneNumber,
				Title:       line,
			}
			sceneNumber++
			sceneContent.Reset()
		} else if currentScene != nil {
			// 解析场景内容
			if strings.HasPrefix(line, "地点：") {
				currentScene.Location = strings.TrimPrefix(line, "地点：")
			} else if strings.HasPrefix(line, "时间：") {
				currentScene.TimeOfDay = strings.TrimPrefix(line, "时间：")
			} else {
				sceneContent.WriteString(line + "\n")
			}
		}
	}

	// 保存最后一个场景
	if currentScene != nil {
		currentScene.Description = sceneContent.String()
		scenes = append(scenes, currentScene)
	}

	return scenes, nil
}

// buildMetadata 构建元数据
func (s *Service) buildMetadata(analysis *NovelAnalysis) string {
	metadata := map[string]interface{}{
		"analysis":    analysis,
		"created_by":  "ai_adaptation",
		"version":     "1.0",
		"timestamp":   fmt.Sprintf("%d", time.Now().Unix()),
	}

	bytes, _ := json.Marshal(metadata)
	return string(bytes)
}

// cleanJSONResponse 清理JSON响应
func (s *Service) cleanJSONResponse(response string) string {
	// 移除可能的markdown标记
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")
	response = strings.TrimSpace(response)

	// 查找JSON对象的开始和结束
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start != -1 && end != -1 && end > start {
		return response[start : end+1]
	}

	return response
}

// formatAnalysisForPrompt 格式化分析结果用于提示词
func (s *Service) formatAnalysisForPrompt(analysis *NovelAnalysis) string {
	bytes, _ := json.MarshalIndent(analysis, "", "  ")
	return string(bytes)
}

// AdaptScriptRequest 改编剧本请求
type AdaptScriptRequest struct {
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	Novel     string `json:"novel"`
	NovelText string `json:"novel_text"`
	Style     string `json:"style"`
}

// NovelAnalysis 小说分析结果
type NovelAnalysis struct {
	Characters []struct {
		Name       string `json:"name"`
		Role       string `json:"role"`
		Importance string `json:"importance"`
	} `json:"characters"`
	PlotStructure struct {
		Beginning   string `json:"beginning"`
		Development string `json:"development"`
		Climax      string `json:"climax"`
		Ending      string `json:"ending"`
	} `json:"plot_structure"`
	Scenes []struct {
		Location    string `json:"location"`
		Description string `json:"description"`
	} `json:"scenes"`
	DialogueRatio float64  `json:"dialogue_ratio"`
	Timeline      string   `json:"timeline"`
	Themes        []string `json:"themes"`
	Tone          string   `json:"tone"`
}
