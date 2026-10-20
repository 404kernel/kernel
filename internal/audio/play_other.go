//go:build !windows

package audio

import (
	"os"
	"os/exec"
	"runtime"
)

func playWAV(data []byte) {
	if len(data) == 0 {
		return
	}
	tmp, err := os.CreateTemp("", "kernel_sound_*.wav")
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name())
	_, _ = tmp.Write(data)
	_ = tmp.Close()

	if runtime.GOOS == "darwin" {
		_ = exec.Command("afplay", tmp.Name()).Run()
	} else {
		_ = exec.Command("aplay", tmp.Name()).Run()
	}
}
