package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/404kernel/kernel/internal/config"
)

type Provider string

const (
	ProviderGemini Provider = "gemini"
	ProviderClaude Provider = "claude"
	ProviderOpenAI Provider = "openai"
	ProviderOllama Provider = "ollama"
	ProviderCustom Provider = "custom"
	ProviderDemo   Provider = "demo"
)

type Client struct {
	Provider Provider
	APIKey   string
	Model    string
	BaseURL  string
	HTTP     *http.Client
}

func NewClientFromConfig(cfg *config.Config) *Client {
	provider := Provider(cfg.Provider)
	var key string
	var baseURL string
	model := cfg.Model

	switch provider {
	case ProviderGemini:
		key = cfg.GeminiAPIKey
		if model == "" || model == "gemini-2.5-flash" || model == "gemini-1.5-flash" {
			model = "gemini-flash-latest"
		}
	case ProviderClaude:
		key = cfg.AnthropicKey
		if model == "" {
			model = "claude-3-5-sonnet-latest"
		}
	case ProviderOpenAI:
		key = cfg.OpenAIAPIKey
		if model == "" {
			model = "gpt-4o"
		}
	case ProviderOllama:
		baseURL = cfg.OllamaEndpoint
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		if model == "" {
			model = "llama3"
		}
	case ProviderCustom:
		key = cfg.CustomKey
		baseURL = cfg.CustomEndpoint
		model = cfg.CustomModel
		if model == "" {
			model = "default"
		}
	}

	if key == "" && provider != ProviderOllama && provider != ProviderCustom {
		provider = ProviderDemo
	}

	return &Client{
		Provider: provider,
		APIKey:   key,
		Model:    model,
		BaseURL:  baseURL,
		HTTP: &http.Client{
			Timeout: 180 * time.Second, // generous timeout for large code generation
		},
	}
}

type StreamChunk struct {
	Text  string
	Done  bool
	Error error
}

func (c *Client) Stream(ctx context.Context, prompt string, ch chan<- StreamChunk) {
	defer close(ch)

	if c.Provider == ProviderDemo {
		c.streamDemo(ctx, prompt, ch)
		return
	}

	switch c.Provider {
	case ProviderGemini:
		c.streamGemini(ctx, prompt, ch)
	case ProviderClaude:
		c.streamClaude(ctx, prompt, ch)
	case ProviderOpenAI, ProviderOllama, ProviderCustom:
		c.streamOpenAI(ctx, prompt, ch)
	default:
		c.streamDemo(ctx, prompt, ch)
	}
}

func (c *Client) streamDemo(ctx context.Context, prompt string, ch chan<- StreamChunk) {
	demoResponse := []string{
		"[KERNEL // DEMO MODE]: No active API key found.\n",
		"Run with your API key or configure it in the Setup Wizard:\n\n",
		"```bash\n",
		"export GEMINI_API_KEY=\"your-key\"\n",
		"```\n\n",
		"Sparky is ready to connect once you provide a key!\n",
	}

	for _, token := range demoResponse {
		select {
		case <-ctx.Done():
			ch <- StreamChunk{Error: ctx.Err(), Done: true}
			return
		case <-time.After(50 * time.Millisecond):
			ch <- StreamChunk{Text: token}
		}
	}
	ch <- StreamChunk{Done: true}
}

func (c *Client) streamGemini(ctx context.Context, prompt string, ch chan<- StreamChunk) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		ch <- StreamChunk{Error: err, Done: true}
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		ch <- StreamChunk{Error: err, Done: true}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		ch <- StreamChunk{Error: fmt.Errorf("gemini api error (%d): %s", resp.StatusCode, string(respBody)), Done: true}
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- StreamChunk{Error: err, Done: true}
		return
	}

	var data struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &data); err != nil {
		ch <- StreamChunk{Error: fmt.Errorf("failed to parse gemini response: %w", err), Done: true}
		return
	}

	var fullText string
	for _, cand := range data.Candidates {
		for _, part := range cand.Content.Parts {
			fullText += part.Text
		}
	}

	if fullText == "" {
		ch <- StreamChunk{Text: "(No response returned from model)", Done: true}
		return
	}

	// Smoothly stream chunks to give pleasant live typing UX
	words := strings.SplitAfter(fullText, " ")
	for _, word := range words {
		select {
		case <-ctx.Done():
			ch <- StreamChunk{Error: ctx.Err(), Done: true}
			return
		case <-time.After(15 * time.Millisecond):
			ch <- StreamChunk{Text: word}
		}
	}
	ch <- StreamChunk{Done: true}
}

func (c *Client) streamClaude(ctx context.Context, prompt string, ch chan<- StreamChunk) {
	url := "https://api.anthropic.com/v1/messages"

	payload := map[string]interface{}{
		"model":      c.Model,
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": true,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		ch <- StreamChunk{Error: err, Done: true}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		ch <- StreamChunk{Error: err, Done: true}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		ch <- StreamChunk{Error: fmt.Errorf("claude api error (%d): %s", resp.StatusCode, string(respBody)), Done: true}
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data: ") {
				jsonStr := strings.TrimPrefix(line, "data: ")
				var chunk struct {
					Type  string `json:"type"`
					Delta struct {
						Text string `json:"text"`
					} `json:"delta"`
				}
				if jsonErr := json.Unmarshal([]byte(jsonStr), &chunk); jsonErr == nil {
					if chunk.Delta.Text != "" {
						ch <- StreamChunk{Text: chunk.Delta.Text}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}
	ch <- StreamChunk{Done: true}
}

func (c *Client) streamOpenAI(ctx context.Context, prompt string, ch chan<- StreamChunk) {
	endpoint := "https://api.openai.com/v1/chat/completions"
	if c.Provider == ProviderOllama {
		endpoint = c.BaseURL + "/v1/chat/completions"
	} else if c.Provider == ProviderCustom {
		endpoint = strings.TrimSuffix(c.BaseURL, "/") + "/chat/completions"
	}

	payload := map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": true,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		ch <- StreamChunk{Error: err, Done: true}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		ch <- StreamChunk{Error: err, Done: true}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		ch <- StreamChunk{Error: fmt.Errorf("api error (%d): %s", resp.StatusCode, string(respBody)), Done: true}
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data: ") {
				jsonStr := strings.TrimPrefix(line, "data: ")
				if strings.TrimSpace(jsonStr) == "[DONE]" {
					ch <- StreamChunk{Done: true}
					return
				}
				var chunk struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}
				if jsonErr := json.Unmarshal([]byte(jsonStr), &chunk); jsonErr == nil {
					for _, choice := range chunk.Choices {
						if choice.Delta.Content != "" {
							ch <- StreamChunk{Text: choice.Delta.Content}
						}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}
	ch <- StreamChunk{Done: true}
}
