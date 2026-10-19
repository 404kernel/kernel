package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ExecResult struct {
	Command  string
	Stdout   string
	Stderr   string
	Duration time.Duration
	ExitCode int
	Error    error
}

// RunShellCommand executes a command using powershell on Windows, or bash/sh on Unix.
func RunShellCommand(ctx context.Context, cmdStr string, timeout time.Duration) ExecResult {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "powershell", "-NoProfile", "-NonInteractive", "-Command", cmdStr)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", cmdStr)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return ExecResult{
		Command:  cmdStr,
		Stdout:   strings.TrimSpace(stdoutBuf.String()),
		Stderr:   strings.TrimSpace(stderrBuf.String()),
		Duration: duration,
		ExitCode: exitCode,
		Error:    err,
	}
}

// FormatResult formats command execution output for terminal display
func FormatResult(r ExecResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("$ %s (in %v, exit: %d)\n\n", r.Command, r.Duration.Round(time.Millisecond), r.ExitCode))

	if r.Stdout != "" {
		b.WriteString(r.Stdout + "\n")
	}
	if r.Stderr != "" {
		if r.Stdout != "" {
			b.WriteString("\n--- STDERR ---\n")
		}
		b.WriteString(r.Stderr + "\n")
	}
	if r.Error != nil && r.Stdout == "" && r.Stderr == "" {
		b.WriteString(fmt.Sprintf("Error: %v\n", r.Error))
	}
	return strings.TrimSpace(b.String())
}
