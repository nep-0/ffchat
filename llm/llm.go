package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float64
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func New(baseURL, apiKey, model string, temperature float64) *Client {
	if temperature == 0 {
		temperature = 0.1
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &Client{
		BaseURL:     baseURL,
		APIKey:      apiKey,
		Model:       model,
		Temperature: temperature,
	}
}

func (c *Client) GenerateFFmpegCommand(prompt string) (string, error) {
	systemPrompt := `You are a helpful assistant that generates ffmpeg commands. 
Given a user's description of a media processing task, output ONLY the ffmpeg command to accomplish it.
Do NOT include any explanations, markdown formatting, or code blocks unless necessary.
If the input is not a media processing task, respond with exactly "NOT_FFMPEG: <explanation>".
The output should be a valid shell command that can be executed directly.
Do not use -y flag unless explicitly requested by the user.
Do not use quotes around the entire command.`

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	reqBody := ChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: c.Temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.BaseURL + "chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("API error: %s", errorResp.Error.Message)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	if strings.HasPrefix(content, "NOT_FFMPEG:") {
		return "", fmt.Errorf("%s", strings.TrimPrefix(content, "NOT_FFMPEG: "))
	}

	return content, nil
}
