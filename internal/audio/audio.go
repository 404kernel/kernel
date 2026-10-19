package audio

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	VoiceEnabled = true
)

// PlayCompanionSound plays an animal-specific vocalization or sci-fi sound effect
func PlayCompanionSound(skin string) {
	if !VoiceEnabled {
		return
	}

	skin = strings.ToLower(skin)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		if runtime.GOOS == "windows" {
			var script string
			switch skin {
			case "cat":
				// Cat cute meow vocalization
				script = `(New-Object -ComObject SAPI.SpVoice).Speak('Meow!')`
			case "dragon":
				// Dragon roar
				script = `[console]::beep(200, 150); (New-Object -ComObject SAPI.SpVoice).Speak('Rawr!')`
			case "drone":
				// Futuristic digital sonar chirp sequence
				script = `[console]::beep(1800, 40); [console]::beep(2400, 60); [console]::beep(3200, 80)`
			default: // "dog"
				// Dog enthusiastic Woof vocalization
				script = `(New-Object -ComObject SAPI.SpVoice).Speak('Woof! Woof!')`
			}
			cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
			_ = cmd.Run()
		} else if runtime.GOOS == "darwin" {
			switch skin {
			case "cat":
				cmd := exec.CommandContext(ctx, "say", "-v", "Victoria", "Meow")
				_ = cmd.Run()
			case "dragon":
				cmd := exec.CommandContext(ctx, "say", "-v", "Zarvox", "Rawr")
				_ = cmd.Run()
			case "drone":
				cmd := exec.CommandContext(ctx, "afplay", "/System/Library/Sounds/Submarine.aiff")
				_ = cmd.Run()
			default:
				cmd := exec.CommandContext(ctx, "say", "-v", "Fred", "Woof woof")
				_ = cmd.Run()
			}
		} else {
			// Linux bell fallback
			fmt.Print("\a")
		}
	}()
}

// Bark is a backwards-compatible alias for the dog vocalization
func Bark() {
	PlayCompanionSound("dog")
}

// LevelUpChime plays an ascending triumphant fanfare
func LevelUpChime() {
	if !VoiceEnabled {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		if runtime.GOOS == "windows" {
			script := `& { [console]::beep(523, 100); [console]::beep(659, 100); [console]::beep(784, 120); [console]::beep(1046, 220) }`
			cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
			_ = cmd.Run()
		} else if runtime.GOOS == "darwin" {
			cmd := exec.CommandContext(ctx, "afplay", "/System/Library/Sounds/Glass.aiff")
			_ = cmd.Run()
		} else {
			fmt.Print("\a")
		}
	}()
}

// Speak uses native system speech synthesis to speak any phrase aloud
func Speak(text string) {
	if !VoiceEnabled || text == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if runtime.GOOS == "windows" {
			safeText := strings.ReplaceAll(text, "'", "''")
			safeText = strings.ReplaceAll(safeText, "$", "")
			safeText = strings.ReplaceAll(safeText, "`", "")
			script := fmt.Sprintf("(New-Object -ComObject SAPI.SpVoice).Speak('%s')", safeText)
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
