package audio

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

var (
	VoiceEnabled = true
)

// Bark plays a quick synthesized audio bark / chime in the background
func Bark() {
	if !VoiceEnabled {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if runtime.GOOS == "windows" {
			// Fast acoustic beep sequence via PowerShell
			cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", "[console]::beep(800, 100); [console]::beep(1100, 150)")
			_ = cmd.Run()
		} else if runtime.GOOS == "darwin" {
			cmd := exec.CommandContext(ctx, "afplay", "/System/Library/Sounds/Tink.aiff")
			_ = cmd.Run()
		} else {
			// Linux bell
			fmt.Print("\a")
		}
	}()
}

// LevelUpChime plays an ascending triumphant chime
func LevelUpChime() {
	if !VoiceEnabled {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if runtime.GOOS == "windows" {
			cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", "[console]::beep(523, 100); [console]::beep(659, 100); [console]::beep(784, 150); [console]::beep(1046, 250)")
			_ = cmd.Run()
		} else if runtime.GOOS == "darwin" {
			cmd := exec.CommandContext(ctx, "afplay", "/System/Library/Sounds/Glass.aiff")
			_ = cmd.Run()
		} else {
			fmt.Print("\a")
		}
	}()
}

// Speak uses native system speech synthesis to speak a phrase aloud
func Speak(text string) {
	if !VoiceEnabled || text == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if runtime.GOOS == "windows" {
			// Escape single quotes
			safeText := ""
			for _, r := range text {
				if r == '\'' {
					safeText += "''"
				} else if r == '"' || r == '`' || r == '$' {
					continue
				} else {
					safeText += string(r)
				}
			}
			script := fmt.Sprintf("Add-Type -AssemblyName System.Speech; $synth = New-Object System.Speech.Synthesis.SpeechSynthesizer; $synth.Rate = 1; $synth.Speak('%s')", safeText)
			cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
			_ = cmd.Run()
		} else if runtime.GOOS == "darwin" {
			cmd := exec.CommandContext(ctx, "say", text)
			_ = cmd.Run()
		} else {
			cmd := exec.CommandContext(ctx, "espeak", text)
			_ = cmd.Run()
		}
	}()
}
