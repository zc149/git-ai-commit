package llm

// Provider는 LLM 제공자를 위한 인터페이스입니다.
type Provider interface {
	// Generate는 주어진 프롬프트로부터 커밋 메시지 후보들을 생성합니다.
	Generate(prompt string) ([]string, error)
	// Close는 리소스를 정리합니다.
	Close() error
}

// NewProvider는 설정에 따른 Provider 인스턴스를 반환합니다.
// 현재는 Groq만 지원합니다.
func NewProvider(model string, apiKey string) (Provider, error) {
	switch model {
	case "groq":
		return NewGroqProvider(apiKey)
	default:
		// 기본값은 Groq
		return NewGroqProvider(apiKey)
	}
}
