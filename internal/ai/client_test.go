package ai

import (
	"context"
	"testing"
	"time"
)

func TestNewClientDemoFallback(t *testing.T) {
	client := NewClient(ProviderGemini, "")
	if client.Provider != ProviderDemo && client.APIKey == "" {
		t.Errorf("expected demo provider fallback when api key is missing")
	}
}

func TestStreamDemo(t *testing.T) {
	client := NewClient(ProviderDemo, "")
	ch := make(chan StreamChunk, 20)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go client.Stream(ctx, "test prompt", ch)

	var chunks []string
	for chunk := range ch {
		if chunk.Error != nil {
			t.Fatalf("unexpected stream error: %v", chunk.Error)
		}
		if chunk.Text != "" {
			chunks = append(chunks, chunk.Text)
		}
	}

	if len(chunks) == 0 {
		t.Errorf("expected demo stream to emit chunks, got none")
	}
}
