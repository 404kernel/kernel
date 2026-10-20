package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/404kernel/kernel/internal/exec"
)

type GitChanges struct {
	HasChanges bool
	Status     string
	Diff       string
	Files      []string
}

// InspectGitChanges queries git status and git diff in the current workspace.
func InspectGitChanges() GitChanges {
	ctx := context.Background()

	statusRes := exec.RunShellCommand(ctx, "git status --porcelain", 5*time.Second)
	statusText := strings.TrimSpace(statusRes.Stdout)
	if statusText == "" {
		return GitChanges{HasChanges: false}
	}

	var files []string
	for _, line := range strings.Split(statusText, "\n") {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 3 {
			files = append(files, strings.TrimSpace(trimmed[2:]))
		}
	}

	// First try staged diff
	diffRes := exec.RunShellCommand(ctx, "git diff --cached", 5*time.Second)
	diffText := strings.TrimSpace(diffRes.Stdout)

	// If no staged diff, get unstaged working diff
	if diffText == "" {
		diffRes = exec.RunShellCommand(ctx, "git diff", 5*time.Second)
		diffText = strings.TrimSpace(diffRes.Stdout)
	}

	// Truncate diff if massive to prevent token explosion
	if len(diffText) > 6000 {
		diffText = diffText[:6000] + "\n... [diff truncated for length] ..."
	}

	return GitChanges{
		HasChanges: true,
		Status:     statusText,
		Diff:       diffText,
		Files:      files,
	}
}

// BuildCommitPrompt crafts an instruction prompt for the LLM to output a conventional commit.
func BuildCommitPrompt(changes GitChanges) string {
	return fmt.Sprintf(`You are an expert Git and software release assistant.
Based on the following git status and git diff, write a concise, professional Conventional Commit message.

Rules:
1. First line must be formatted as: type(scope): short imperative summary (max 72 chars).
   Types: feat, fix, docs, style, refactor, perf, test, chore.
2. If helpful, include 2-3 short bullet points after an empty line describing key changes.
3. Output ONLY the commit message. No explanations, no markdown code fences, no surrounding quotes.

Git Status:
%s

Git Diff:
%s
`, changes.Status, changes.Diff)
}

// HeuristicCommitMessage generates an intelligent fallback commit message offline or in demo mode.
func HeuristicCommitMessage(changes GitChanges) string {
	if len(changes.Files) == 0 {
		return "chore: update workspace files"
	}

	scope := "core"
	firstFile := changes.Files[0]
	if strings.Contains(firstFile, "ui/") || strings.Contains(firstFile, "styles") {
		scope = "ui"
	} else if strings.Contains(firstFile, "tools/") {
		scope = "tools"
	} else if strings.Contains(firstFile, "config/") {
		scope = "config"
	} else if strings.HasSuffix(firstFile, ".md") {
		return "docs: update documentation and usage guides"
	}

	// Detect if features were added
	hasAdditions := strings.Contains(changes.Diff, "+")
	if hasAdditions {
		return fmt.Sprintf("feat(%s): implement new functionality and enhancements across %d files", scope, len(changes.Files))
	}

	return fmt.Sprintf("refactor(%s): update and refine %s", scope, firstFile)
}

// RunGitCommit executes git add and git commit with the given message.
func RunGitCommit(message string) exec.ExecResult {
	// Escape double quotes for shell execution
	escaped := strings.ReplaceAll(message, `"`, `\"`)
	cmd := fmt.Sprintf(`git add -A && git commit -m "%s"`, escaped)
	return exec.RunShellCommand(context.Background(), cmd, 10*time.Second)
}
