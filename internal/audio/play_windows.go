//go:build windows

package audio

import (
	"syscall"
	"unsafe"
)

var (
	winmm          = syscall.NewLazyDLL("winmm.dll")
	procPlaySoundW = winmm.NewProc("PlaySoundW")
)

const (
	SND_ASYNC  = 0x0001
	SND_MEMORY = 0x0004
)

func playWAV(data []byte) {
	if len(data) == 0 {
		return
	}
	procPlaySoundW.Call(
		uintptr(unsafe.Pointer(&data[0])),
		0,
		uintptr(SND_MEMORY|SND_ASYNC),
	)
}
