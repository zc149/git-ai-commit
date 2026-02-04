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
	apiKey  string
	baseURL string
}

// NewClaudeProvider는 새로운 ClaudeProvider 인스턴스를 생성합니다.
func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1/messages",
	}
}

// claudeRequest는 Claude API 요청 구조체입니다.
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

// claudeMessage는 Claude API 메시지 구조체입니다.
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse는 Claude API 응답 구조체입니다.
type claudeResponse struct {
	Content []claudeContent `json:"content"`
}

// claudeContent는 Claude API 응답 콘텐츠 구조체입니다.
type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Generate는 Claude API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (c *ClaudeProvider) Generate(prompt string) ([]string, error) {
	reqBody := claudeRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1000,
		Messages: []claudeMessage{
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

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-dangerous-direct-browser-access", "true")

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

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// 응답 텍스트에서 메시지 후보 추출
	text := claudeResp.Content[0].Text
	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		// 파싱 실패 시 전체 텍스트를 하나의 메시지로 반환
		return []string{text}, nil
	}

	return messages, nil
}
