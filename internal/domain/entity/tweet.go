package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tweet 推文实体
type Tweet struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	ProjectID   *uuid.UUID     `json:"project_id" gorm:"type:uuid"`
	Project     *Project       `json:"project" gorm:"foreignKey:ProjectID"`
	
	// 推文内容
	Content     string         `json:"content" gorm:"type:text;not null"`
	Title       string         `json:"title"`
	Platform    string         `json:"platform" gorm:"default:'twitter'"` // twitter/weibo/douyin
	Style       string         `json:"style"`                              // 推文风格
	Theme       string         `json:"theme"`                              // 推文主题/角度
	
	// 推文元数据
	Hashtags    string         `json:"hashtags" gorm:"type:text"`          // JSON数组
	Length      int            `json:"length"`                             // 字符长度
	Quality     float64        `json:"quality"`                            // 质量评分
	
	// 来源信息
	SourceType  string         `json:"source_type"`                        // novel/manual/template
	SourceData  string         `json:"source_data" gorm:"type:text"`       // 源数据（小说内容等）
	
	// 状态和统计
	Status      string         `json:"status" gorm:"default:'draft'"`      // draft/published/archived
	ViewCount   int            `json:"view_count" gorm:"default:0"`
	LikeCount   int            `json:"like_count" gorm:"default:0"`
	ShareCount  int            `json:"share_count" gorm:"default:0"`
	
	// 发布信息
	PublishedAt *time.Time     `json:"published_at"`
	PublishURL  string         `json:"publish_url"`                        // 发布后的URL
	
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TweetTemplate 推文模板实体
type TweetTemplate struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Category    string         `json:"category" gorm:"not null"`           // 模板分类
	
	// 模板内容
	Template    string         `json:"template" gorm:"type:text;not null"` // 模板内容，支持变量
	Variables   string         `json:"variables" gorm:"type:text"`         // JSON数组，模板变量
	
	// 模板配置
	Platform    string         `json:"platform" gorm:"default:'twitter'"`
	Style       string         `json:"style"`
	MaxLength   int            `json:"max_length" gorm:"default:280"`
	
	// 使用统计
	UseCount    int            `json:"use_count" gorm:"default:0"`
	Rating      float64        `json:"rating" gorm:"default:0"`
	
	// 状态
	IsPublic    bool           `json:"is_public" gorm:"default:true"`
	IsPremium   bool           `json:"is_premium" gorm:"default:false"`
	Status      string         `json:"status" gorm:"default:'active'"`
	
	CreatedBy   uuid.UUID      `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TweetAnalysis 推文分析实体
type TweetAnalysis struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TweetID     uuid.UUID      `json:"tweet_id" gorm:"type:uuid;not null"`
	Tweet       Tweet          `json:"tweet" gorm:"foreignKey:TweetID"`
	
	// 分析结果
	Sentiment   string         `json:"sentiment"`                          // positive/negative/neutral
	Keywords    string         `json:"keywords" gorm:"type:text"`          // JSON数组
	Topics      string         `json:"topics" gorm:"type:text"`            // JSON数组
	Emotions    string         `json:"emotions" gorm:"type:text"`          // JSON数组
	
	// 质量分析
	ReadabilityScore float64   `json:"readability_score"`
	EngagementScore  float64   `json:"engagement_score"`
	ViralityScore    float64   `json:"virality_score"`
	
	// 建议
	Suggestions string         `json:"suggestions" gorm:"type:text"`       // JSON数组
	
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TableName 指定表名
func (Tweet) TableName() string {
	return "tweets"
}

func (TweetTemplate) TableName() string {
	return "tweet_templates"
}

func (TweetAnalysis) TableName() string {
	return "tweet_analyses"
}

// BeforeCreate 创建前的钩子
func (t *Tweet) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (tt *TweetTemplate) BeforeCreate(tx *gorm.DB) error {
	if tt.ID == uuid.Nil {
		tt.ID = uuid.New()
	}
	return nil
}

func (ta *TweetAnalysis) BeforeCreate(tx *gorm.DB) error {
	if ta.ID == uuid.Nil {
		ta.ID = uuid.New()
	}
	return nil
}
