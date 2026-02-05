package config

import (
	"fmt"
	"os"
	"strings"
)

// Config는 애플리케이션 설정을 나타냅니다.
type Config struct {
	// API 키들 (우선순위 순서대로)
	APIKeys APIKeys

	// 선택할 LLM 모델 (비어있으면 첫 번째 유효한 키 사용)
	Model string
}

// APIKeys는 각 LLM Provider의 API 키를 나타냅니다.
type APIKeys struct {
	Claude string
	OpenAI string
	Gemini string
	Codex  string
}

// Load는 설정을 로드합니다.
// 환경 변수에서 직접 API 키를 읽어옵니다.
func Load() (*Config, error) {
	cfg := &Config{
		APIKeys: APIKeys{
			Claude: getEnvWithFallback("AI_COMMIT_CLAUDE_API_KEY", "CLAUDE_API_KEY", "ANTHROPIC_API_KEY"),
			OpenAI: getEnvWithFallback("AI_COMMIT_OPENAI_API_KEY", "OPENAI_API_KEY"),
			Gemini: getEnvWithFallback("AI_COMMIT_GEMINI_API_KEY", "GEMINI_API_KEY", "GOOGLE_API_KEY"),
			Codex:  getEnvWithFallback("AI_COMMIT_CODEX_API_KEY", "CODEX_API_KEY", "OPENAI_API_KEY"),
		},
		Model: getEnvWithFallback("AI_COMMIT_MODEL", "AI_COMMIT_LLM_MODEL"),
	}

	return cfg, nil
}

// GetFirstAvailableModel는 첫 번째 유효한 API 키를 가진 모델을 반환합니다.
// 우선순위: Claude > OpenAI > Gemini > Codex
func (c *Config) GetFirstAvailableModel() string {
	if c.Model != "" {
		return c.Model
	}

	if c.APIKeys.Claude != "" {
		return "claude"
	}
	if c.APIKeys.OpenAI != "" {
		return "openai"
	}
	if c.APIKeys.Gemini != "" {
		return "gemini"
	}
	if c.APIKeys.Codex != "" {
		return "codex"
	}

	return ""
}

// GetAPIKey는 지정된 모델의 API 키를 반환합니다.
func (c *Config) GetAPIKey(model string) (string, error) {
	model = strings.ToLower(model)

	switch model {
	case "claude":
		if c.APIKeys.Claude == "" {
			return "", fmt.Errorf("Claude API key not found")
		}
		return c.APIKeys.Claude, nil
	case "openai":
		if c.APIKeys.OpenAI == "" {
			return "", fmt.Errorf("OpenAI API key not found")
		}
		return c.APIKeys.OpenAI, nil
	case "gemini":
		if c.APIKeys.Gemini == "" {
			return "", fmt.Errorf("Gemini API key not found")
		}
		return c.APIKeys.Gemini, nil
	case "codex":
		if c.APIKeys.Codex == "" {
			return "", fmt.Errorf("Codex API key not found")
		}
		return c.APIKeys.Codex, nil
	default:
		return "", fmt.Errorf("unknown model: %s", model)
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
