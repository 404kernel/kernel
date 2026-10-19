package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/404kernel/kernel/internal/config"
)

type HistoryItem struct {
	Timestamp string `json:"timestamp"`
	Role      string `json:"role"`
	Content   string `json:"content"`
}

type SessionRecord struct {
	ID        string        `json:"id"`
	StartedAt string        `json:"started_at"`
	Messages  []HistoryItem `json:"messages"`
}

func getHistoryPath() (string, error) {
	dir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

// AppendSession records a completed chat session to ~/.kernel/history.json
func AppendSession(messages []HistoryItem) error {
	if len(messages) == 0 {
		return nil
	}

	path, err := getHistoryPath()
	if err != nil {
		return err
	}

	var sessions []SessionRecord
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &sessions)
	}

	// Keep last 25 sessions to avoid unbounded file growth
	if len(sessions) > 25 {
		sessions = sessions[len(sessions)-25:]
	}

	newSession := SessionRecord{
		ID:        fmt.Sprintf("session-%d", time.Now().Unix()),
		StartedAt: time.Now().Format(time.RFC3339),
		Messages:  messages,
	}

	sessions = append(sessions, newSession)
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ExportMarkdown exports current chat messages to a local markdown file
func ExportMarkdown(filename string, messages []HistoryItem) (string, error) {
	if filename == "" {
		filename = fmt.Sprintf("kernel_chat_%s.md", time.Now().Format("20060102_150405"))
	}

	var out string
	out += "# Kernel Chat Session\n\n"
	out += fmt.Sprintf("Generated on: `%s`\n\n---\n\n", time.Now().Format(time.RFC1123))

	for _, m := range messages {
		out += fmt.Sprintf("### %s (%s)\n\n%s\n\n", m.Role, m.Timestamp, m.Content)
	}

	err := os.WriteFile(filename, []byte(out), 0644)
	if err != nil {
		return "", err
	}
	return filename, nil
}
