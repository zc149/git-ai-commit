package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ClaudeProvider는 Claude API를 사용하는 제공자입니다.
type ClaudeProvider struct {
	apiKey string
	client *http.Client
	model  string
}

// ClaudeResponse는 Claude API 응답 구조체입니다.
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason   string  `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ClaudeRequest는 Claude API 요청 구조체입니다.
type ClaudeRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Messages  []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

// NewClaudeProvider는 새로운 ClaudeProvider 인스턴스를 생성합니다.
func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey: apiKey,
		client: &http.Client{},
		model:  "claude-3-5-sonnet-20241022",
	}
}

// Generate는 Claude API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (c *ClaudeProvider) Generate(prompt string) ([]string, error) {
	reqBody := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 4096,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-dangerous-direct-browser-access", "true")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// 응답 텍스트 추출
	var text string
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	// 응답 텍스트에서 메시지 후보 추출
	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		// 파싱 실패 시 전체 텍스트를 하나의 메시지로 반환
		return []string{text}, nil
	}

	return messages, nil
}
