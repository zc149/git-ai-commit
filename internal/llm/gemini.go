package llm

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider는 Gemini API를 사용하는 제공자입니다.
type GeminiProvider struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiProvider는 새로운 GeminiProvider 인스턴스를 생성합니다.
func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// gemini-2.0-flash 모델 사용
	model := client.GenerativeModel("gemini-2.0-flash")
	model.SetMaxOutputTokens(4096)

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

// Generate는 Gemini API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (g *GeminiProvider) Generate(prompt string) ([]string, error) {
	ctx := context.Background()

	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	if len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no parts in candidate content")
	}

	var text string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			text += string(txt)
		}
	}

	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		return []string{text}, nil
	}

	return messages, nil
}

// Close는 Gemini 클라이언트를 닫습니다.
func (g *GeminiProvider) Close() error {
	return g.client.Close()
}
