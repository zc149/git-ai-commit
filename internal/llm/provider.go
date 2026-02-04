package llm

// Provider는 LLM 제공자를 위한 인터페이스입니다.
type Provider interface {
	// Generate는 주어진 프롬프트로부터 커밋 메시지 후보들을 생성합니다.
	Generate(prompt string) ([]string, error)
}

// NewProvider는 설정에 따른 Provider 인스턴스를 반환합니다.
func NewProvider(model string, apiKey string) (Provider, error) {
	switch model {
	case "claude":
		return NewClaudeProvider(apiKey), nil
	case "openai":
		return NewOpenAIProvider(apiKey), nil
	case "codex":
		return NewCodexProvider(apiKey), nil
	case "glm":
		return NewGLMProvider(apiKey), nil
	case "gemini":
		return NewGeminiProvider(apiKey), nil
	default:
		return NewClaudeProvider(apiKey), nil
	}
}
