package audio

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

//go:embed assets/dog_bark.wav
var dogBarkWAV []byte

//go:embed assets/cat_meow.wav
var catMeowWAV []byte

var (
	VoiceEnabled = true
)

// PlayCompanionSound plays an authentic animal vocalization or sci-fi sound effect
func PlayCompanionSound(skin string) {
	if !VoiceEnabled {
		return
	}

	skin = strings.ToLower(skin)

	switch skin {
	case "dog":
		playWAV(dogBarkWAV)
	case "cat":
		playWAV(catMeowWAV)
	case "drone":
		go func() {
			if runtime.GOOS == "windows" {
				script := `[console]::beep(1800, 40); [console]::beep(2400, 60); [console]::beep(3200, 80)`
				cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
				_ = cmd.Run()
			} else {
				fmt.Print("\a")
			}
		}()
	case "dragon":
		go func() {
			if runtime.GOOS == "windows" {
				script := `[console]::beep(220, 100); [console]::beep(180, 150); [console]::beep(140, 200)`
				cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
				_ = cmd.Run()
			} else {
				fmt.Print("\a")
			}
		}()
	default:
		playWAV(dogBarkWAV)
	}
}

// Bark is a backwards-compatible alias for the dog bark
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
