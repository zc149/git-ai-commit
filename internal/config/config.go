package config

import (
	"fmt"
	"os"
	"strings"
)

// Config는 애플리케이션 설정을 나타냅니다.
type Config struct {
	// API 키
	GroqAPIKey string

	// 사용할 LLM 모델 (현재는 groq만 지원)
	Model string
}

// Load는 설정을 로드합니다.
// 환경 변수에서 API 키를 읽어옵니다.
func Load() (*Config, error) {
	cfg := &Config{
		GroqAPIKey: getEnvWithFallback("AI_COMMIT_GROQ_API_KEY", "GROQ_API_KEY"),
		Model:      os.Getenv("AI_COMMIT_MODEL"),
	}

	// 기본 모델은 groq
	if cfg.Model == "" {
		cfg.Model = "groq"
	}

	return cfg, nil
}

// GetFirstAvailableModel는 첫 번째 유효한 API 키를 가진 모델을 반환합니다.
func (c *Config) GetFirstAvailableModel() string {
	if c.GroqAPIKey != "" {
		return "groq"
	}
	return ""
}

// GetAPIKey는 지정된 모델의 API 키를 반환합니다.
func (c *Config) GetAPIKey(model string) (string, error) {
	model = strings.ToLower(model)

	switch model {
	case "groq":
		if c.GroqAPIKey == "" {
			return "", fmt.Errorf("Groq API key not found. Please set AI_COMMIT_GROQ_API_KEY environment variable")
		}
		return c.GroqAPIKey, nil
	default:
		return "", fmt.Errorf("unknown model: %s (only 'groq' is supported)", model)
	}
}

// getEnvWithFallback은 여러 환경 변수 이름 중 첫 번째로 존재하는 값을 반환합니다.
func getEnvWithFallback(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}
