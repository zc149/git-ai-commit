package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAIProvider는 OpenAI API를 사용하는 제공자입니다.
type OpenAIProvider struct {
	apiKey  string
	baseURL string
}

// NewOpenAIProvider는 새로운 OpenAIProvider 인스턴스를 생성합니다.
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1/chat/completions",
	}
}

// openaiRequest는 OpenAI API 요청 구조체입니다.
type openaiRequest struct {
	Model     string          `json:"model"`
	Messages  []openaiMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

// openaiMessage는 OpenAI API 메시지 구조체입니다.
type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openaiResponse는 OpenAI API 응답 구조체입니다.
type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
}

// openaiChoice는 OpenAI API 응답 선택 구조체입니다.
type openaiChoice struct {
	Message openaiMessage `json:"message"`
}

// Generate는 OpenAI API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (o *OpenAIProvider) Generate(prompt string) ([]string, error) {
	reqBody := openaiRequest{
		Model: "gpt-4",
		Messages: []openaiMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 1000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", o.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var openaiResp openaiResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// 응답 텍스트에서 메시지 후보 추출
	text := openaiResp.Choices[0].Message.Content
	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		return []string{text}, nil
	}

	return messages, nil
}
