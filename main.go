package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/404kernel/kernel/internal/ai"
	"github.com/404kernel/kernel/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	providerFlag := flag.String("provider", "demo", "AI provider (demo, gemini, claude, openai, ollama)")
	modelFlag := flag.String("model", "", "Model name override")
	flag.Parse()

	provider := ai.Provider(*providerFlag)
	client := ai.NewClient(provider, *modelFlag)

	m := ui.NewModel(client)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running kernel: %v\n", err)
		os.Exit(1)
	}
}
