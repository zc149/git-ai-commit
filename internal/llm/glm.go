package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GLMProvider는 GLM API를 사용하는 제공자입니다.
type GLMProvider struct {
	apiKey  string
	baseURL string
}

// NewGLMProvider는 새로운 GLMProvider 인스턴스를 생성합니다.
func NewGLMProvider(apiKey string) *GLMProvider {
	return &GLMProvider{
		apiKey:  apiKey,
		baseURL: "https://api.z.ai/api/paas/v4/chat/completions",
	}
}

// glmRequest는 GLM API 요청 구조체입니다.
type glmRequest struct {
	Model     string       `json:"model"`
	Messages  []glmMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens"`
}

// glmMessage는 GLM API 메시지 구조체입니다.
type glmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// glmResponse는 GLM API 응답 구조체입니다.
type glmResponse struct {
	Choices []glmChoice `json:"choices"`
}

// glmChoice는 GLM API 응답 선택 구조체입니다.
type glmChoice struct {
	Message glmMessage `json:"message"`
}

// Generate는 GLM API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (g *GLMProvider) Generate(prompt string) ([]string, error) {
	reqBody := glmRequest{
		Model: "glm-4-flash",
		Messages: []glmMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 4096,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", g.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

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

	var glmResp glmResponse
	if err := json.Unmarshal(body, &glmResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(glmResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	text := glmResp.Choices[0].Message.Content
	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		return []string{text}, nil
	}

	return messages, nil
}
