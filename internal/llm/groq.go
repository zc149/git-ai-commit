package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

// GroqProvider는 Groq API를 사용하는 제공자입니다.
// Groq는 OpenAI 호환 API를 제공하므로 OpenAI SDK를 사용합니다.
type GroqProvider struct {
	client *openai.Client
	model  string
}

// NewGroqProvider는 새로운 GroqProvider 인스턴스를 생성합니다.
func NewGroqProvider(apiKey string) (*GroqProvider, error) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"

	client := openai.NewClientWithConfig(config)

	return &GroqProvider{
		client: client,
		model:  "llama-3.3-70b-versatile",
	}, nil
}

// Generate는 Groq API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (g *GroqProvider) Generate(prompt string) ([]string, error) {
	ctx := context.Background()

	resp, err := g.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: g.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   4096,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	text := resp.Choices[0].Message.Content

	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		return []string{text}, nil
	}

	return messages, nil
}

// Close는 Groq 클라이언트를 닫습니다.
func (g *GroqProvider) Close() error {
	// openai.Client에는 Close 메서드가 없음
	return nil
}

// NewGroqProviderFromEnv는 환경변수에서 API 키를 읽어 GroqProvider를 생성합니다.
func NewGroqProviderFromEnv() (*GroqProvider, error) {
	apiKey := os.Getenv("AI_COMMIT_GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AI_COMMIT_GROQ_API_KEY environment variable not set")
	}

	return NewGroqProvider(apiKey)
}
