package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadWorkspaceFile reads file contents up to a safe size limit (256 KB)
func ReadWorkspaceFile(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path '%s' is a directory", cleanPath)
	}
	if info.Size() > 256*1024 {
		return "", fmt.Errorf("file '%s' exceeds 256KB safe limit", cleanPath)
	}

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteWorkspaceFile writes content to a file, creating parent directories if needed
func WriteWorkspaceFile(path string, content string) error {
	cleanPath := filepath.Clean(path)
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(cleanPath, []byte(content), 0644)
}

// ComputeSimpleDiff returns a unified-style diff between original and new content
func ComputeSimpleDiff(original, newContent, filename string) string {
	origLines := strings.Split(original, "\n")
	newLines := strings.Split(newContent, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s\n+++ %s (modified)\n", filename, filename))

	maxLen := len(origLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	hasDiff := false
	for i := 0; i < maxLen; i++ {
		var oLine, nLine string
		hasOld := i < len(origLines)
		hasNew := i < len(newLines)

		if hasOld {
			oLine = origLines[i]
		}
		if hasNew {
			nLine = newLines[i]
		}

		if hasOld && hasNew {
			if oLine == nLine {
				diff.WriteString(fmt.Sprintf("  %s\n", oLine))
			} else {
				hasDiff = true
				diff.WriteString(fmt.Sprintf("- %s\n", oLine))
				diff.WriteString(fmt.Sprintf("+ %s\n", nLine))
			}
		} else if hasOld && !hasNew {
			hasDiff = true
			diff.WriteString(fmt.Sprintf("- %s\n", oLine))
		} else if !hasOld && hasNew {
			hasDiff = true
			diff.WriteString(fmt.Sprintf("+ %s\n", nLine))
		}
	}

	if !hasDiff {
		return fmt.Sprintf("No differences found in %s", filename)
	}
	return diff.String()
}
