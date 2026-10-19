package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	GeminiAPIKey   string `json:"gemini_api_key,omitempty"`
	AnthropicKey   string `json:"anthropic_api_key,omitempty"`
	OpenAIAPIKey   string `json:"openai_api_key,omitempty"`
	OllamaEndpoint string `json:"ollama_endpoint,omitempty"`
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".kernel")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return &Config{Provider: "gemini", Model: "gemini-flash-latest"}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Fallback to environment variables
		cfg := &Config{
			Provider:       "gemini",
			Model:          "gemini-flash-latest",
			GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
			AnthropicKey:   os.Getenv("ANTHROPIC_API_KEY"),
			OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
			OllamaEndpoint: "http://localhost:11434",
		}
		return cfg, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{Provider: "gemini", Model: "gemini-flash-latest"}, nil
	}

	if cfg.Model == "gemini-2.5-flash" || cfg.Model == "gemini-1.5-flash" || cfg.Model == "" {
		cfg.Model = "gemini-flash-latest"
	}

	// Environment variables always override config file if set
	if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
		cfg.GeminiAPIKey = envKey
	}
	if envKey := os.Getenv("ANTHROPIC_API_KEY"); envKey != "" {
		cfg.AnthropicKey = envKey
	}
	if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		cfg.OpenAIAPIKey = envKey
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Config) HasValidKey() bool {
	switch c.Provider {
	case "gemini":
		return c.GeminiAPIKey != ""
	case "claude":
		return c.AnthropicKey != ""
	case "openai":
		return c.OpenAIAPIKey != ""
	case "ollama":
		return true // Ollama runs locally without key
	default:
		return c.GeminiAPIKey != "" || c.AnthropicKey != "" || c.OpenAIAPIKey != ""
	}
}
