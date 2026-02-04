package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GeminiProvider는 Gemini API를 사용하는 제공자입니다.
type GeminiProvider struct {
	apiKey  string
	baseURL string
}

// NewGeminiProvider는 새로운 GeminiProvider 인스턴스를 생성합니다.
func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent",
	}
}

// geminiRequest는 Gemini API 요청 구조체입니다.
type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

// geminiContent는 Gemini API 콘텐츠 구조체입니다.
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

// geminiPart는 Gemini API 파트 구조체입니다.
type geminiPart struct {
	Text string `json:"text"`
}

// geminiResponse는 Gemini API 응답 구조체입니다.
type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

// geminiCandidate는 Gemini API 후보 구조체입니다.
type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

// Generate는 Gemini API를 호출하여 커밋 메시지 후보들을 생성합니다.
func (g *GeminiProvider) Generate(prompt string) ([]string, error) {
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{
						Text: prompt,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// API 키를 URL 파라미터로 전달
	url := fmt.Sprintf("%s?key=%s", g.baseURL, g.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text
	messages := parseCommitMessages(text)

	if len(messages) == 0 {
		return []string{text}, nil
	}

	return messages, nil
}
