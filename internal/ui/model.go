package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/404kernel/kernel/internal/ai"
	"github.com/404kernel/kernel/internal/config"
	"github.com/404kernel/kernel/internal/pet"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppMode int

const (
	ModeChat AppMode = iota
	ModeWizardSelectProvider
	ModeWizardEnterKey
	ModeWizardCustomBaseURL
	ModeWizardCustomModel
)

type chatMessage struct {
	Role    string
	Content string
	Time    string
}

type Model struct {
	cfg        *config.Config
	aiClient   *ai.Client
	pet        *pet.Pet
	textInput  textinput.Model
	viewport   viewport.Model
	spinner    spinner.Model
	streaming  bool
	streamChan chan ai.StreamChunk
	cancelFunc context.CancelFunc

	Mode           AppMode
	WizardProvider string
	TempCustomURL  string
	TempCustomKey  string
	TickCount      int

	Messages    []chatMessage
	Cwd         string
	CurrentResp strings.Builder

	width  int
	height int
	ready  bool
}

type streamMsg ai.StreamChunk
type petTickMsg time.Time

func petTicker() tea.Cmd {
	return tea.Tick(140*time.Millisecond, func(t time.Time) tea.Msg {
		return petTickMsg(t)
	})
}

func NewModel(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 4096
	ti.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(SecondaryColor)

	p := pet.NewPet("Sparky")
	cwd, _ := os.Getwd()

	initialMode := ModeChat
	if !cfg.HasValidKey() && cfg.Provider != "demo" {
		initialMode = ModeWizardSelectProvider
		ti.Placeholder = "Type 1-5 and press Enter..."
	} else {
		ti.Placeholder = "Ask kernel to explain code, generate a game, or run tasks..."
	}

	client := ai.NewClientFromConfig(cfg)

	m := Model{
		cfg:       cfg,
		aiClient:  client,
		pet:       p,
		textInput: ti,
		spinner:   s,
		Cwd:       cwd,
		Mode:      initialMode,
		Messages: []chatMessage{
			{
				Role:    "SYSTEM",
				Content: fmt.Sprintf("Kernel AI Agent initialized. Connected to [%s] using [%s]. Sparky is ready to assist.", client.Provider, client.Model),
				Time:    time.Now().Format("15:04"),
			},
		},
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		petTicker(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case petTickMsg:
		m.pet.Tick()
		m.TickCount++
		if m.streaming {
			m.updateViewport()
		}
		return m, petTicker()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			if m.streaming && m.cancelFunc != nil {
				m.cancelFunc()
				m.streaming = false
				m.pet.SetState(pet.StateIdle)
				m.Messages = append(m.Messages, chatMessage{
					Role:    "SYSTEM",
					Content: "Response generation canceled by user.",
					Time:    time.Now().Format("15:04"),
				})
				m.updateViewport()
				return m, nil
			}
			return m, tea.Quit
		case tea.KeyEsc:
			if m.streaming && m.cancelFunc != nil {
				m.cancelFunc()
				m.streaming = false
				m.pet.SetState(pet.StateIdle)
				m.updateViewport()
				return m, nil
			}
		case tea.KeyEnter:
			if m.Mode == ModeWizardSelectProvider {
				return m.handleWizardSelect()
			} else if m.Mode == ModeWizardEnterKey {
				return m.handleWizardEnterKey()
			} else if m.Mode == ModeWizardCustomBaseURL {
				return m.handleWizardCustomBaseURL()
			} else if m.Mode == ModeWizardCustomModel {
				return m.handleWizardCustomModel()
			}

			// Normal Chat Mode
			if m.streaming {
				return m, nil
			}
			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}
			// Only quit if user explicitly enters :q or /quit, NOT when asking a question containing 'q'
			if input == ":q" || input == "/quit" || input == "exit" {
				return m, tea.Quit
			}

			m.textInput.Reset()
			m.CurrentResp.Reset()
			m.streaming = true
			m.pet.SetState(pet.StateThinking)

			m.Messages = append(m.Messages, chatMessage{
				Role:    "YOU",
				Content: input,
				Time:    time.Now().Format("15:04"),
			})
			m.updateViewport()

			ctx, cancel := context.WithCancel(context.Background())
			m.cancelFunc = cancel
			m.streamChan = make(chan ai.StreamChunk, 50)
			go m.aiClient.Stream(ctx, input, m.streamChan)

			return m, waitForStream(m.streamChan)
		}

	case streamMsg:
		if msg.Error != nil {
			m.streaming = false
			m.pet.SetState(pet.StateIdle)
			m.Messages = append(m.Messages, chatMessage{
				Role:    "ERROR",
				Content: msg.Error.Error(),
				Time:    time.Now().Format("15:04"),
			})
			m.updateViewport()
			return m, nil
		}
		if msg.Done {
			m.streaming = false
			m.pet.SetState(pet.StateHappy)
			if m.CurrentResp.Len() > 0 {
				m.Messages = append(m.Messages, chatMessage{
					Role:    "KERNEL",
					Content: m.CurrentResp.String(),
					Time:    time.Now().Format("15:04"),
				})
				m.CurrentResp.Reset()
			}
			m.updateViewport()
			return m, nil
		}

		m.CurrentResp.WriteString(msg.Text)
		m.updateViewport()
		return m, waitForStream(m.streamChan)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		contentWidth := m.width - 4
		if contentWidth < 40 {
			contentWidth = 40
		}
		m.textInput.Width = contentWidth - 8

		vpHeight := m.height - 15
		if vpHeight < 6 {
			vpHeight = 6
		}

		if !m.ready {
			m.viewport = viewport.New(contentWidth, vpHeight)
			m.ready = true
		} else {
			m.viewport.Width = contentWidth
			m.viewport.Height = vpHeight
		}
		m.updateViewport()
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.spinner, spCmd = m.spinner.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

func (m *Model) handleWizardSelect() (tea.Model, tea.Cmd) {
	val := strings.TrimSpace(m.textInput.Value())
	m.textInput.Reset()

	switch val {
	case "1":
		m.WizardProvider = "gemini"
		m.Mode = ModeWizardEnterKey
		m.textInput.Placeholder = "Paste your Google Gemini API Key..."
	case "2":
		m.WizardProvider = "claude"
		m.Mode = ModeWizardEnterKey
		m.textInput.Placeholder = "Paste your Anthropic Claude API Key..."
	case "3":
		m.WizardProvider = "openai"
		m.Mode = ModeWizardEnterKey
		m.textInput.Placeholder = "Paste your OpenAI API Key..."
	case "4":
		m.cfg.Provider = "ollama"
		m.cfg.Model = "llama3"
		_ = config.SaveConfig(m.cfg)
		m.aiClient = ai.NewClientFromConfig(m.cfg)
		m.Mode = ModeChat
		m.textInput.Placeholder = "Ask kernel to explain code, generate a game, or run tasks..."
		m.Messages = append(m.Messages, chatMessage{
			Role:    "SYSTEM",
			Content: "Ollama configured (http://localhost:11434). Ready!",
			Time:    time.Now().Format("15:04"),
		})
		m.updateViewport()
	case "5":
		m.WizardProvider = "custom"
		m.Mode = ModeWizardCustomBaseURL
		m.textInput.Placeholder = "Enter Custom Base URL (e.g. https://api.deepseek.com/v1)..."
	default:
		m.cfg.Provider = "demo"
		m.aiClient = ai.NewClientFromConfig(m.cfg)
		m.Mode = ModeChat
		m.textInput.Placeholder = "Ask kernel to explain code, generate a game, or run tasks..."
		m.Messages = append(m.Messages, chatMessage{
			Role:    "SYSTEM",
			Content: "Running in Interactive Demo Mode.",
			Time:    time.Now().Format("15:04"),
		})
		m.updateViewport()
	}

	return m, nil
}

func (m *Model) handleWizardCustomBaseURL() (tea.Model, tea.Cmd) {
	url := strings.TrimSpace(m.textInput.Value())
	m.textInput.Reset()
	if url == "" {
		url = "https://api.deepseek.com/v1"
	}
	m.TempCustomURL = url
	m.Mode = ModeWizardEnterKey
	m.textInput.Placeholder = "Paste your API Key for this endpoint..."
	return m, nil
}

func (m *Model) handleWizardEnterKey() (tea.Model, tea.Cmd) {
	key := strings.TrimSpace(m.textInput.Value())
	m.textInput.Reset()

	if m.WizardProvider == "custom" {
		m.TempCustomKey = key
		m.Mode = ModeWizardCustomModel
		m.textInput.Placeholder = "Enter model name (e.g. deepseek-chat, mistral-large)..."
		return m, nil
	}

	if key != "" {
		m.cfg.Provider = m.WizardProvider
		switch m.WizardProvider {
		case "gemini":
			m.cfg.GeminiAPIKey = key
			m.cfg.Model = "gemini-flash-latest"
		case "claude":
			m.cfg.AnthropicKey = key
			m.cfg.Model = "claude-3-5-sonnet-latest"
		case "openai":
			m.cfg.OpenAIAPIKey = key
			m.cfg.Model = "gpt-4o"
		}
		_ = config.SaveConfig(m.cfg)
		m.aiClient = ai.NewClientFromConfig(m.cfg)
	}

	m.Mode = ModeChat
	m.pet.SetState(pet.StateHappy)
	m.textInput.Placeholder = "Ask kernel to explain code, generate a game, or run tasks..."
	m.Messages = append(m.Messages, chatMessage{
		Role:    "SYSTEM",
		Content: fmt.Sprintf("Setup complete! Connected to [%s]. Configuration saved to ~/.kernel/config.json", m.cfg.Provider),
		Time:    time.Now().Format("15:04"),
	})
	m.updateViewport()

	return m, nil
}

func (m *Model) handleWizardCustomModel() (tea.Model, tea.Cmd) {
	modelName := strings.TrimSpace(m.textInput.Value())
	m.textInput.Reset()
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	m.cfg.Provider = "custom"
	m.cfg.CustomEndpoint = m.TempCustomURL
	m.cfg.CustomKey = m.TempCustomKey
	m.cfg.CustomModel = modelName
	m.cfg.Model = modelName

	_ = config.SaveConfig(m.cfg)
	m.aiClient = ai.NewClientFromConfig(m.cfg)

	m.Mode = ModeChat
	m.pet.SetState(pet.StateHappy)
	m.textInput.Placeholder = "Ask kernel to explain code, generate a game, or run tasks..."
	m.Messages = append(m.Messages, chatMessage{
		Role:    "SYSTEM",
		Content: fmt.Sprintf("Custom API connected: [%s / %s]. Configuration saved to ~/.kernel/config.json", m.cfg.CustomEndpoint, m.cfg.CustomModel),
		Time:    time.Now().Format("15:04"),
	})
	m.updateViewport()

	return m, nil
}

func (m *Model) updateViewport() {
	var b strings.Builder

	for _, msg := range m.Messages {
		var roleTag string
		switch msg.Role {
		case "YOU":
			roleTag = UserBadge.Render(" USER ")
		case "KERNEL":
			roleTag = AgentBadge.Render(" KERNEL ")
		case "SYSTEM":
			roleTag = SystemBadge.Render(" SYSTEM ")
		case "ERROR":
			roleTag = ErrorBadge.Render(" ERROR ")
		}

		header := fmt.Sprintf("%s %s", roleTag, LogTimestamp.Render(msg.Time))
		bubble := MessageBubble.Render(fmt.Sprintf("%s\n\n%s", header, msg.Content))
		b.WriteString(bubble + "\n\n")
	}

	if m.streaming {
		dots := []string{"● ○ ○", "○ ● ○", "○ ○ ●", "● ● ●"}
		dotAnim := dots[(m.TickCount/2)%len(dots)]
		typingHeader := fmt.Sprintf("%s %s %s", AgentBadge.Render(" KERNEL "), TypingDotsStyle.Render(dotAnim), LogTimestamp.Render(time.Now().Format("15:04")))

		content := m.CurrentResp.String()
		if content == "" {
			content = SubtitleStyle.Render("Thinking and drafting solution...")
		} else {
			content += CursorStyle.Render(" ▋")
		}

		bubble := StreamingBubble.Render(fmt.Sprintf("%s\n\n%s", typingHeader, content))
		b.WriteString(bubble + "\n")
	}

	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func waitForStream(ch chan ai.StreamChunk) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return streamMsg{Done: true}
		}
		return streamMsg(chunk)
	}
}

func (m Model) View() string {
	var s strings.Builder

	// Top Header Bar
	title := TitleStyle.Render("⚡ KERNEL")
	providerBadge := ProviderBadgeStyle.Render(fmt.Sprintf("%s :: %s", strings.ToUpper(string(m.aiClient.Provider)), m.aiClient.Model))
	cwdBadge := SubtitleStyle.Render(fmt.Sprintf("cwd: %s", m.Cwd))
	headerBar := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", providerBadge, "  ", cwdBadge)
	s.WriteString(HeaderBox.Render(headerBar) + "\n\n")

	// Pet Companion Card
	petBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(0, 1).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				m.pet.Render(),
				"   ",
				fmt.Sprintf("%s\n%s", TitleStyle.Render("COMPANION: "+m.pet.Name), SubtitleStyle.Render("Status: "+m.pet.Mood)),
			),
		)
	s.WriteString(petBox + "\n\n")

	// If in Setup Wizard mode, display guided menu
	if m.Mode == ModeWizardSelectProvider {
		wizardMenu := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(WarningColor).
			Padding(1, 2).
			Render(
				TitleStyle.Render("SETUP WIZARD // CHOOSE AI PROVIDER") + "\n\n" +
					"Select an AI model provider:\n\n" +
					" [1] Google Gemini  (Free Tier API Key - gemini-flash-latest)\n" +
					" [2] Anthropic Claude (Sonnet 3.5 / Haiku)\n" +
					" [3] OpenAI (GPT-4o)\n" +
					" [4] Local Ollama (Free, runs offline on localhost:11434)\n" +
					" [5] Custom OpenAI-Compatible (DeepSeek, OpenRouter, Groq, Together)\n" +
					" [6] Try Demo Mode (Explore interface without a key)\n\n" +
					SubtitleStyle.Render("Type 1-6 and press Enter:"),
			)
		s.WriteString(wizardMenu + "\n\n")
	} else if m.Mode == ModeWizardCustomBaseURL {
		menu := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(1, 2).
			Render(
				TitleStyle.Render("CUSTOM API // ENTER BASE URL") + "\n\n" +
					"Examples:\n" +
					"  DeepSeek:   https://api.deepseek.com/v1\n" +
					"  OpenRouter: https://openrouter.ai/api/v1\n" +
					"  Groq:       https://api.groq.com/openai/v1\n\n" +
					SubtitleStyle.Render("Enter base URL and press Enter:"),
			)
		s.WriteString(menu + "\n\n")
	} else if m.Mode == ModeWizardCustomModel {
		menu := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(1, 2).
			Render(
				TitleStyle.Render("CUSTOM API // ENTER MODEL NAME") + "\n\n" +
					"Examples: deepseek-chat, deepseek-reasoner, llama-3.3-70b-versatile\n\n" +
					SubtitleStyle.Render("Enter model name and press Enter:"),
			)
		s.WriteString(menu + "\n\n")
	} else if m.Mode == ModeWizardEnterKey {
		keyMenu := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(1, 2).
			Render(
				TitleStyle.Render(fmt.Sprintf("ENTER %s API KEY", strings.ToUpper(m.WizardProvider))) + "\n\n" +
					"Key will be securely saved to ~/.kernel/config.json.\n" +
					SubtitleStyle.Render("Paste your key below and press Enter (or press Enter to skip):"),
			)
		s.WriteString(keyMenu + "\n\n")
	} else {
		// Viewport (Chat conversation)
		s.WriteString(m.viewport.View() + "\n\n")
	}

	// Bottom Input Field
	var inputBorder = InputBoxActive
	if m.streaming {
		inputBorder = InputBoxInactive
	}
	s.WriteString(inputBorder.Render(m.textInput.View()) + "\n")

	// Footer Keybind Hints
	s.WriteString(FooterStyle.Render(
		fmt.Sprintf("%s submit · %s cancel stream · %s quit", FooterKey.Render("Enter"), FooterKey.Render("Ctrl+C"), FooterKey.Render(":q")),
	))

	return s.String()
}
