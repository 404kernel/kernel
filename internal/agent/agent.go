package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/404kernel/kernel/internal/ai"
	"github.com/404kernel/kernel/internal/exec"
	"github.com/404kernel/kernel/internal/tools"
)

type StepType string

const (
	StepThinking StepType = "thinking"
	StepWriting  StepType = "writing"
	StepRunning  StepType = "running"
	StepDone     StepType = "done"
	StepError    StepType = "error"
)

type AgentProgress struct {
	Type    StepType
	Message string
	Output  string
	Done    bool
}

var (
	// Look for code blocks that specify target file: ```language filename="path" or `FILE: path`
	fileBlockRegex = regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\\s+(?:FILE|path|filename)=[\"']?([^\"'\n]+)[\"']?\\s*\n(.*?)```")
)

// RunAutonomousTask executes an iterative agent workflow to satisfy a goal
func RunAutonomousTask(ctx context.Context, client *ai.Client, goal string, progressChan chan<- AgentProgress) {
	defer close(progressChan)

	progressChan <- AgentProgress{
		Type:    StepThinking,
		Message: fmt.Sprintf("Planning autonomous steps for: '%s'...", goal),
	}

	prompt := fmt.Sprintf(`You are Kernel Autonomous Agent. The user wants to achieve this goal:
"%s"

Provide a concrete, production-ready solution. If files need to be created or modified, format the code blocks explicitly with the target path like this:
`+"```"+`html FILE="index.html"
<!DOCTYPE html>
<html>...</html>
`+"```"+`

Or if commands should be run, specify them clearly in a bash/powershell block. Be concise and direct.`, goal)

	// Stream solution from AI
	ch := make(chan ai.StreamChunk, 50)
	go client.Stream(ctx, prompt, ch)

	var fullResponse strings.Builder
	for chunk := range ch {
		if chunk.Error != nil {
			progressChan <- AgentProgress{
				Type:    StepError,
				Message: fmt.Sprintf("AI Generation failed: %v", chunk.Error),
				Done:    true,
			}
			return
		}
		fullResponse.WriteString(chunk.Text)
	}

	content := fullResponse.String()

	// Parse out any target file blocks to write to disk
	matches := fileBlockRegex.FindAllStringSubmatch(content, -1)
	writtenFiles := 0
	for _, match := range matches {
		if len(match) >= 3 {
			targetFile := strings.TrimSpace(match[1])
			fileContent := match[2]

			progressChan <- AgentProgress{
				Type:    StepWriting,
				Message: fmt.Sprintf("Writing generated code to `%s`...", targetFile),
			}

			if err := tools.WriteWorkspaceFile(targetFile, fileContent); err != nil {
				progressChan <- AgentProgress{
					Type:    StepError,
					Message: fmt.Sprintf("Failed to write `%s`: %v", targetFile, err),
				}
			} else {
				writtenFiles++
				progressChan <- AgentProgress{
					Type:    StepWriting,
					Message: fmt.Sprintf("Created `%s` (%d bytes).", targetFile, len(fileContent)),
				}
			}
		}
	}

	// Look for bash commands to run if requested
	if strings.Contains(goal, "run") || strings.Contains(goal, "test") || strings.Contains(goal, "execute") {
		cmdRegex := regexp.MustCompile("(?s)```(?:bash|powershell|sh|cmd)\\s*\n(.*?)```")
		if cmdMatch := cmdRegex.FindStringSubmatch(content); len(cmdMatch) >= 2 {
			cmdStr := strings.TrimSpace(cmdMatch[1])
			lines := strings.Split(cmdStr, "\n")
			firstCmd := strings.TrimSpace(lines[0])
			if firstCmd != "" && !strings.Contains(firstCmd, "rm -rf") {
				progressChan <- AgentProgress{
					Type:    StepRunning,
					Message: fmt.Sprintf("Executing verification: `$ %s`...", firstCmd),
				}
				res := exec.RunShellCommand(ctx, firstCmd, 20*time.Second)
				progressChan <- AgentProgress{
					Type:    StepRunning,
					Message: fmt.Sprintf("Output (exit %d):\n%s", res.ExitCode, res.Stdout),
				}
			}
		}
	}

	summary := fmt.Sprintf("Autonomous task finished! Wrote %d file(s).\n\n%s", writtenFiles, content)
	progressChan <- AgentProgress{
		Type:    StepDone,
		Message: summary,
		Done:    true,
	}
}
