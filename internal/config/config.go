package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	AI       AIConfig       `mapstructure:"ai"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	Expire    int    `mapstructure:"expire"`
}

// AIConfig AI配置
type AIConfig struct {
	SDEndpoint      string
	OllamaEndpoint  string
	OllamaModel     string
	OllamaApiKey    string // 新增
	WhisperEndpoint string
	TTSEndpoint     string
}

// Load 加载配置
func Load() *Config {
	// 加载环境变量文件
	godotenv.Load()

	// 设置默认值
	setDefaults()

	// 从环境变量读取配置
	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "comic_video_user"),
			Password: getEnv("DB_PASSWORD", "comic_video_password"),
			DBName:   getEnv("DB_NAME", "comic_video"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
			UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
			BucketName:      getEnv("MINIO_BUCKET_NAME", "comic-video"),
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET_KEY", "your-secret-key"),
			Expire:    getEnvAsInt("JWT_EXPIRE", 24*60*60), // 24小时
		},
		AI: AIConfig{
			SDEndpoint:      getEnv("SD_ENDPOINT", "http://127.0.0.1:7860"),
			OllamaEndpoint:  getEnv("OLLAMA_ENDPOINT", "http://127.0.0.1:11434"),
			OllamaModel:     getEnv("OLLAMA_MODEL", "llama2"),
			OllamaApiKey:    getEnv("OLLAMA_API_KEY", ""),
			WhisperEndpoint: getEnv("WHISPER_ENDPOINT", "http://127.0.0.1:9000"),
			TTSEndpoint:     getEnv("TTS_ENDPOINT", "http://127.0.0.1:50021"),
		},
	}

	return config
}

// setDefaults 设置默认值
func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket_name", "comic-video")
	viper.SetDefault("jwt.expire", 24*60*60)
	viper.SetDefault("ai.sd_endpoint", "http://127.0.0.1:7860")
	viper.SetDefault("ai.ollama_endpoint", "http://127.0.0.1:11434")
	viper.SetDefault("ai.ollama_model", "llama2")
	viper.SetDefault("ai.ollama_api_key", "")
	viper.SetDefault("ai.whisper_endpoint", "http://127.0.0.1:9000")
	viper.SetDefault("ai.tts_endpoint", "http://127.0.0.1:50021")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
} 