package tools

import (
	"strings"
	"testing"
)

func TestHeuristicCommitMessage(t *testing.T) {
	changes := GitChanges{
		HasChanges: true,
		Files:      []string{"internal/ui/styles.go", "internal/ui/model.go"},
		Diff:       "+func ApplyTheme()",
	}

	msg := HeuristicCommitMessage(changes)
	if !strings.HasPrefix(msg, "feat(ui):") {
		t.Errorf("expected feat(ui): prefix, got: %s", msg)
	}

	docChanges := GitChanges{
		HasChanges: true,
		Files:      []string{"README.md"},
		Diff:       "+update docs",
	}
	docMsg := HeuristicCommitMessage(docChanges)
	if !strings.HasPrefix(docMsg, "docs:") {
		t.Errorf("expected docs: prefix, got: %s", docMsg)
	}
}

func TestBuildCommitPrompt(t *testing.T) {
	changes := GitChanges{
		HasChanges: true,
		Status:     "M internal/ui/model.go",
		Diff:       "+new feature",
	}
	prompt := BuildCommitPrompt(changes)
	if !strings.Contains(prompt, "Conventional Commit") {
		t.Errorf("expected prompt to mention Conventional Commit")
	}
}
