package core

import (
	"git-ai-commit/internal/git"
	"git-ai-commit/internal/llm"
)

// Generator는 커밋 메시지를 생성하는 역할을 합니다.
type Generator struct {
	provider llm.Provider
}

// NewGenerator는 새로운 Generator 인스턴스를 생성합니다.
func NewGenerator(provider llm.Provider) *Generator {
	return &Generator{
		provider: provider,
	}
}

// Generate는 diff를 분석하여 커밋 메시지 후보들을 생성합니다.
func (g *Generator) Generate(diff *git.DiffResult, detail string) ([]string, error) {
	// 프롬프트 생성
	prompt := GeneratePrompt(diff, detail)

	// LLM 호출
	messages, err := g.provider.Generate(prompt)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
