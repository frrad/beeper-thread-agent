package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ModelMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type OpenRouterClient struct {
	APIKey string
	Model  string
	Client *http.Client
}

func (o *OpenRouterClient) Complete(ctx context.Context, messages []ModelMessage) (string, error) {
	body, err := json.Marshal(map[string]any{"model": o.Model, "messages": messages})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+o.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "Beeper Thread Agent")
	client := o.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("OpenRouter %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 || strings.TrimSpace(result.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("OpenRouter returned an empty response")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
