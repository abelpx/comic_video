package tweet

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"comic_video/internal/domain/entity"
	"comic_video/internal/repository/postgres"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TemplateService 推文模板服务
type TemplateService struct {
	templateRepo *postgres.TweetTemplateRepository
	db           *gorm.DB
}

// NewTemplateService 创建推文模板服务
func NewTemplateService(db *gorm.DB) *TemplateService {
	return &TemplateService{
		templateRepo: postgres.NewTweetTemplateRepository(db),
		db:           db,
	}
}

// CreateTemplate 创建推文模板
func (s *TemplateService) CreateTemplate(ctx context.Context, userID uuid.UUID, req *CreateTemplateRequest) (*entity.TweetTemplate, error) {
	log.Printf("[TemplateService] 创建推文模板: user_id=%s, name=%s", userID, req.Name)

	// 解析模板变量
	variables := s.extractTemplateVariables(req.Template)
	variablesJSON, _ := json.Marshal(variables)

	template := &entity.TweetTemplate{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Template:    req.Template,
		Variables:   string(variablesJSON),
		Platform:    req.Platform,
		Style:       req.Style,
		MaxLength:   req.MaxLength,
		IsPublic:    req.IsPublic,
		IsPremium:   req.IsPremium,
		CreatedBy:   userID,
		Status:      "active",
	}

	err := s.templateRepo.Create(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("创建模板失败: %w", err)
	}

	log.Printf("[TemplateService] 推文模板创建成功: id=%s", template.ID)
	return template, nil
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Template    string `json:"template"`
	Platform    string `json:"platform"`
	Style       string `json:"style"`
	MaxLength   int    `json:"max_length"`
	IsPublic    bool   `json:"is_public"`
	IsPremium   bool   `json:"is_premium"`
}

// UpdateTemplate 更新推文模板
func (s *TemplateService) UpdateTemplate(ctx context.Context, userID uuid.UUID, templateID uuid.UUID, req *UpdateTemplateRequest) (*entity.TweetTemplate, error) {
	// 获取现有模板
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("获取模板失败: %w", err)
	}

	// 检查权限
	if template.CreatedBy != userID {
		return nil, fmt.Errorf("无权限修改此模板")
	}

	// 更新字段
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Category != "" {
		template.Category = req.Category
	}
	if req.Template != "" {
		template.Template = req.Template
		// 重新解析变量
		variables := s.extractTemplateVariables(req.Template)
		variablesJSON, _ := json.Marshal(variables)
		template.Variables = string(variablesJSON)
	}
	if req.Platform != "" {
		template.Platform = req.Platform
	}
	if req.Style != "" {
		template.Style = req.Style
	}
	if req.MaxLength > 0 {
		template.MaxLength = req.MaxLength
	}

	err = s.templateRepo.Update(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("更新模板失败: %w", err)
	}

	log.Printf("[TemplateService] 推文模板更新成功: id=%s", template.ID)
	return template, nil
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Template    string `json:"template"`
	Platform    string `json:"platform"`
	Style       string `json:"style"`
	MaxLength   int    `json:"max_length"`
}

// GetTemplate 获取推文模板
func (s *TemplateService) GetTemplate(ctx context.Context, templateID uuid.UUID) (*entity.TweetTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("获取模板失败: %w", err)
	}

	return template, nil
}

// ListTemplates 获取推文模板列表
func (s *TemplateService) ListTemplates(ctx context.Context, req *ListTemplatesRequest) (*ListTemplatesResponse, error) {
	templates, total, err := s.templateRepo.List(ctx, req.Category, req.Platform, req.IsPublic, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("获取模板列表失败: %w", err)
	}

	return &ListTemplatesResponse{
		Templates: templates,
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
}

// ListTemplatesRequest 模板列表请求
type ListTemplatesRequest struct {
	Category string `json:"category"`
	Platform string `json:"platform"`
	IsPublic *bool  `json:"is_public"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// ListTemplatesResponse 模板列表响应
type ListTemplatesResponse struct {
	Templates []*entity.TweetTemplate `json:"templates"`
	Total     int64                   `json:"total"`
	Limit     int                     `json:"limit"`
	Offset    int                     `json:"offset"`
}

// SearchTemplates 搜索推文模板
func (s *TemplateService) SearchTemplates(ctx context.Context, req *SearchTemplatesRequest) (*ListTemplatesResponse, error) {
	templates, total, err := s.templateRepo.Search(ctx, req.Keyword, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("搜索模板失败: %w", err)
	}

	return &ListTemplatesResponse{
		Templates: templates,
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
}

// SearchTemplatesRequest 搜索模板请求
type SearchTemplatesRequest struct {
	Keyword string `json:"keyword"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

// GetPopularTemplates 获取热门模板
func (s *TemplateService) GetPopularTemplates(ctx context.Context, platform string, limit int) ([]*entity.TweetTemplate, error) {
	return s.templateRepo.GetPopular(ctx, platform, limit)
}

// GetTemplatesByCategory 根据分类获取模板
func (s *TemplateService) GetTemplatesByCategory(ctx context.Context, category string, limit int) ([]*entity.TweetTemplate, error) {
	return s.templateRepo.GetByCategory(ctx, category, limit)
}

// UseTemplate 使用模板生成推文
func (s *TemplateService) UseTemplate(ctx context.Context, templateID uuid.UUID, variables map[string]string) (string, error) {
	// 获取模板
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return "", fmt.Errorf("获取模板失败: %w", err)
	}

	// 增加使用次数
	_ = s.templateRepo.IncrementUseCount(ctx, templateID)

	// 替换变量
	content := template.Template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		content = strings.ReplaceAll(content, placeholder, value)
	}

	// 检查长度限制
	if template.MaxLength > 0 && len(content) > template.MaxLength {
		content = s.truncateContent(content, template.MaxLength)
	}

	log.Printf("[TemplateService] 模板使用成功: template_id=%s, content_length=%d", templateID, len(content))
	return content, nil
}

// DeleteTemplate 删除推文模板
func (s *TemplateService) DeleteTemplate(ctx context.Context, userID uuid.UUID, templateID uuid.UUID) error {
	// 获取模板检查权限
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("获取模板失败: %w", err)
	}

	if template.CreatedBy != userID {
		return fmt.Errorf("无权限删除此模板")
	}

	err = s.templateRepo.Delete(ctx, templateID)
	if err != nil {
		return fmt.Errorf("删除模板失败: %w", err)
	}

	log.Printf("[TemplateService] 推文模板删除成功: id=%s", templateID)
	return nil
}

// GetTemplateCategories 获取模板分类
func (s *TemplateService) GetTemplateCategories(ctx context.Context) ([]string, error) {
	return s.templateRepo.GetCategories(ctx)
}

// extractTemplateVariables 提取模板变量
func (s *TemplateService) extractTemplateVariables(template string) []string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)
	
	var variables []string
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 {
			variable := strings.TrimSpace(match[1])
			if !seen[variable] {
				variables = append(variables, variable)
				seen[variable] = true
			}
		}
	}
	
	return variables
}

// truncateContent 截断内容
func (s *TemplateService) truncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	// 尝试在句号、感叹号或问号处截断
	punctuation := []string{"。", "！", "？", ".", "!", "?"}
	for i := maxLength - 1; i > maxLength/2; i-- {
		for _, p := range punctuation {
			if i < len(content) && string(content[i]) == p {
				return content[:i+1]
			}
		}
	}

	// 如果找不到合适的截断点，在最后一个空格处截断
	for i := maxLength - 1; i > maxLength/2; i-- {
		if i < len(content) && content[i] == ' ' {
			return content[:i] + "..."
		}
	}

	// 最后的选择：硬截断
	return content[:maxLength-3] + "..."
}
