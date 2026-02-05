package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider는 OpenAI API를 사용하는 제공자입니다.
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAIProvider는 새로운 OpenAIProvider 인스턴스를 생성합니다.
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	config := openai.DefaultConfig(apiKey)
	// 필요한 경우 OpenAI API URL 커스터마이즈 가능
	// config.BaseURL = "https://api.openai.com/v1"

	return &OpenAIProvider{
		client: openai.NewClientWithConfig(config),
		model:  "gpt-4o-mini",
	}
}

// Generate는 OpenAI API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (o *OpenAIProvider) Generate(prompt string) ([]string, error) {
	ctx := context.Background()

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens: 4096,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// 응답 텍스트에서 메시지 후보 추출
	text := resp.Choices[0].Message.Content
	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		return []string{text}, nil
	}

	return messages, nil
}
