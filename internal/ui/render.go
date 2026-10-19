package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// RenderMarkdown parses and renders markdown content with syntax highlighting.
// It wraps code blocks, bold text, lists, and headings cleanly within the provided terminal width.
func RenderMarkdown(content string, width int) string {
	if width < 30 {
		width = 30
	}

	// Create glamour renderer tailored for dark terminal themes
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}

	out, err := r.Render(content)
	if err != nil {
		return content
	}

	return strings.TrimSpace(out)
}
