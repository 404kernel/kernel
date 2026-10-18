package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/404kernel/kernel/internal/config"
	"github.com/404kernel/kernel/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
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
