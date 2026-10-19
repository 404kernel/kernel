package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/404kernel/kernel/internal/config"
	"github.com/404kernel/kernel/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\n[FATAL PANIC]: %v\n", r)
			if f, err := os.OpenFile("kernel_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				fmt.Fprintf(f, "[%s] PANIC: %v\n", time.Now().Format(time.RFC3339), r)
				f.Close()
			}
		}
	}()

	providerFlag := flag.String("provider", "", "AI provider override (gemini, claude, openai, ollama, demo)")
	modelFlag := flag.String("model", "", "Model name override")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}

	if *providerFlag != "" {
		cfg.Provider = *providerFlag
	}
	if *modelFlag != "" {
		cfg.Model = *modelFlag
	}

	m := ui.NewModel(cfg)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running kernel: %v\n", err)
		os.Exit(1)
	}
}
