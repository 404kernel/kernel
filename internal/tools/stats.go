package tools

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var startTime = time.Now()

type SystemTelemetry struct {
	Uptime       time.Duration
	OS           string
	Arch         string
	GoVersion    string
	NumCPU       int
	Goroutines   int
	AllocMB      float64
	TotalAllocMB float64
	SysMB        float64
	NumGC        uint32
}

// GetSystemTelemetry collects live runtime and hardware telemetry metrics
func GetSystemTelemetry() SystemTelemetry {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemTelemetry{
		Uptime:       time.Since(startTime),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		Goroutines:   runtime.NumGoroutine(),
		AllocMB:      float64(m.Alloc) / (1024 * 1024),
		TotalAllocMB: float64(m.TotalAlloc) / (1024 * 1024),
		SysMB:        float64(m.Sys) / (1024 * 1024),
		NumGC:        m.NumGC,
	}
}

// FormatSystemTelemetry renders a formatted cyberpunk stats dashboard
func FormatSystemTelemetry(t SystemTelemetry) string {
	uptimeStr := t.Uptime.Round(time.Second).String()

	var b strings.Builder
	b.WriteString("### 📊 System Telemetry & Hardware Status\n\n")
	b.WriteString("```text\n")
	b.WriteString("╭──────────────────────────┬──────────────────────────╮\n")
	b.WriteString(fmt.Sprintf("│ Platform:   %-12s │ Uptime:     %-12s │\n", t.OS+"/"+t.Arch, uptimeStr))
	b.WriteString(fmt.Sprintf("│ CPU Cores:  %-12d │ Goroutines: %-12d │\n", t.NumCPU, t.Goroutines))
	b.WriteString(fmt.Sprintf("│ Mem Alloc:  %-9.2f MB │ Mem Sys:    %-9.2f MB │\n", t.AllocMB, t.SysMB))
	b.WriteString(fmt.Sprintf("│ Garbage C:  %-12d │ Go Version: %-12s │\n", t.NumGC, t.GoVersion))
	b.WriteString("╰──────────────────────────┴──────────────────────────╯\n")
	b.WriteString("```\n")

	return b.String()
}
