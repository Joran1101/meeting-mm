package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 存储应用程序配置
type Config struct {
	// 服务器配置
	Port string
	Env  string

	// DeepSeek配置
	DeepSeekAPIKey  string
	DeepSeekBaseURL string

	// Notion配置
	NotionAPIKey     string
	NotionDatabaseID string

	// Whisper配置
	WhisperModelPath string
	UseLocalWhisper  bool
}

// 全局配置实例
var AppConfig Config

// LoadConfig 从环境变量加载配置
func LoadConfig(envFile string) error {
	// 加载.env文件
	var err error
	if envFile != "" {
		err = godotenv.Load(envFile)
	} else {
		err = godotenv.Load()
	}

	if err != nil {
		log.Println("警告: .env文件未找到，使用环境变量")
	}

	// 服务器配置
	AppConfig.Port = getEnv("PORT", "8080")
	AppConfig.Env = getEnv("ENV", "development")

	// DeepSeek配置
	AppConfig.DeepSeekAPIKey = getEnv("DEEPSEEK_API_KEY", "")
	AppConfig.DeepSeekBaseURL = getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1")

	// Notion配置
	AppConfig.NotionAPIKey = getEnv("NOTION_API_KEY", "")
	AppConfig.NotionDatabaseID = getEnv("NOTION_DATABASE_ID", "")

	// Whisper配置
	AppConfig.WhisperModelPath = getEnv("WHISPER_MODEL_PATH", "../whisper/models/ggml-base.bin")
	AppConfig.UseLocalWhisper = getEnv("USE_LOCAL_WHISPER", "true") == "true"

	return nil
}

// GetConfig 获取全局配置实例
func GetConfig() *Config {
	return &AppConfig
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
