package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RenderDirectoryTree builds a formatted ASCII tree of files and directories
func RenderDirectoryTree(root string, maxDepth int) (string, error) {
	if root == "" {
		root = "."
	}
	if maxDepth <= 0 {
		maxDepth = 3
	}

	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		cleanRoot = root
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("📁 **%s**\n", filepath.Base(cleanRoot)))

	err = buildTree(&b, cleanRoot, "", 0, maxDepth)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(b.String()), nil
}

func buildTree(b *strings.Builder, currentPath string, prefix string, currentDepth int, maxDepth int) error {
	if currentDepth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return err
	}

	// Filter out noisy directories
	var filtered []os.DirEntry
	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" || name == "node_modules" || name == "dist" || name == ".kernel" || strings.HasPrefix(name, ".") {
			continue
		}
		filtered = append(filtered, entry)
	}

	for i, entry := range filtered {
		isLast := i == len(filtered)-1
		connector := "├── "
		childPrefix := "│   "
		if isLast {
			connector = "└── "
			childPrefix = "    "
		}

		name := entry.Name()
		entryPath := filepath.Join(currentPath, name)

		if entry.IsDir() {
			b.WriteString(fmt.Sprintf("%s%s📁 %s/\n", prefix, connector, name))
			_ = buildTree(b, entryPath, prefix+childPrefix, currentDepth+1, maxDepth)
		} else {
			info, _ := entry.Info()
			sizeStr := ""
			if info != nil {
				sizeStr = formatBytes(info.Size())
			}
			icon := getFileIcon(name)
			b.WriteString(fmt.Sprintf("%s%s%s %s  *(%s)*\n", prefix, connector, icon, name, sizeStr))
		}
	}

	return nil
}

func getFileIcon(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "🐹"
	case ".py":
		return "🐍"
	case ".js", ".ts", ".jsx", ".tsx":
		return "📜"
	case ".json", ".yaml", ".yml":
		return "⚙️"
	case ".md":
		return "📝"
	case ".exe", ".bin":
		return "⚡"
	default:
		return "📄"
	}
}

func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
}
