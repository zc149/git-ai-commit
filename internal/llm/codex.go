package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CodexProvider는 Codex API를 사용하는 제공자입니다.
type CodexProvider struct {
	apiKey  string
	baseURL string
}

// NewCodexProvider는 새로운 CodexProvider 인스턴스를 생성합니다.
func NewCodexProvider(apiKey string) *CodexProvider {
	return &CodexProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1/completions",
	}
}

// codexRequest는 Codex API 요청 구조체입니다.
type codexRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// codexResponse는 Codex API 응답 구조체입니다.
type codexResponse struct {
	Choices []codexChoice `json:"choices"`
}

// codexChoice는 Codex API 응답 선택 구조체입니다.
type codexChoice struct {
	Text string `json:"text"`
}

// Generate는 Codex API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (c *CodexProvider) Generate(prompt string) ([]string, error) {
	reqBody := codexRequest{
		Model:       "code-davinci-003",
		Prompt:      prompt,
		MaxTokens:   1000,
		Temperature: 0.3,
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
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	var codexResp codexResponse
	if err := json.Unmarshal(body, &codexResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(codexResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	text := codexResp.Choices[0].Text
	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		return []string{text}, nil
	}

	return messages, nil
}
