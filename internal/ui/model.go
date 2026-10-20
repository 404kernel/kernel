package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/404kernel/kernel/internal/agent"
	"github.com/404kernel/kernel/internal/ai"
	"github.com/404kernel/kernel/internal/audio"
	"github.com/404kernel/kernel/internal/config"
	"github.com/404kernel/kernel/internal/exec"
	"github.com/404kernel/kernel/internal/history"
	"github.com/404kernel/kernel/internal/pet"
	"github.com/404kernel/kernel/internal/tools"
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
	CurrentResp string

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
	if cfg.Theme != "" {
		_, _ = ApplyTheme(cfg.Theme)
	}

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
			// Handle Slash Commands and direct shell executions
			if strings.HasPrefix(input, "/") || strings.HasPrefix(input, "$") || strings.HasPrefix(input, ":") {
				if updatedModel, cmd, handled := m.handleSlashCommand(input); handled {
					return updatedModel, cmd
				}
			}

			m.textInput.Reset()
			m.CurrentResp = ""
			m.streaming = true
			m.pet.SetState(pet.StateThinking)

			m.Messages = append(m.Messages, chatMessage{
				Role:    "YOU",
				Content: input,
				Time:    time.Now().Format("15:04"),
			})
			m.updateViewport()

			// Build multi-turn context from current session (last 12 messages)
			var historyMsgs []ai.Message
			startIdx := 0
			if len(m.Messages) > 12 {
				startIdx = len(m.Messages) - 12
			}
			for _, msg := range m.Messages[startIdx:] {
				if msg.Role == "SYSTEM" || msg.Role == "ERROR" {
					continue
				}
				historyMsgs = append(historyMsgs, ai.Message{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}

			ctx, cancel := context.WithCancel(context.Background())
			m.cancelFunc = cancel
			m.streamChan = make(chan ai.StreamChunk, 50)
			go m.aiClient.StreamChat(ctx, historyMsgs, m.streamChan)

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
			if len(m.CurrentResp) > 0 {
				m.Messages = append(m.Messages, chatMessage{
					Role:    "KERNEL",
					Content: m.CurrentResp,
					Time:    time.Now().Format("15:04"),
				})
				m.CurrentResp = ""

				// Play animal-specific sound & award pet XP on successful response
				audio.PlayCompanionSound(string(m.pet.Skin))
				if leveledUp := m.pet.AddXP(15); leveledUp {
					audio.LevelUpChime()
					m.Messages = append(m.Messages, chatMessage{
						Role:    "SYSTEM",
						Content: fmt.Sprintf("🎉 Level Up! %s reached Level %d: [%s]!", m.pet.Name, m.pet.Level, m.pet.Title()),
						Time:    time.Now().Format("15:04"),
					})
				}
			}
			m.updateViewport()
			return m, nil
		}

		m.CurrentResp += msg.Text
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

	contentWidth := m.width - 12
	if contentWidth < 30 {
		contentWidth = 30
	}

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
		body := msg.Content
		if msg.Role == "KERNEL" {
			body = RenderMarkdown(msg.Content, contentWidth)
		}

		bubble := MessageBubble.Render(fmt.Sprintf("%s\n\n%s", header, body))
		b.WriteString(bubble + "\n\n")
	}

	if m.streaming {
		dots := []string{"● ○ ○", "○ ● ○", "○ ○ ●", "● ● ●"}
		dotAnim := dots[(m.TickCount/2)%len(dots)]
		typingHeader := fmt.Sprintf("%s %s %s", AgentBadge.Render(" KERNEL "), TypingDotsStyle.Render(dotAnim), LogTimestamp.Render(time.Now().Format("15:04")))

		content := m.CurrentResp
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

func (m *Model) handleSlashCommand(input string) (tea.Model, tea.Cmd, bool) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return m, nil, false
	}
	cmd := parts[0]
	arg := strings.TrimSpace(strings.TrimPrefix(input, cmd))

	// Clean quit and session archive
	if cmd == ":q" || cmd == "/quit" || cmd == "/exit" || cmd == "exit" {
		var hist []history.HistoryItem
		for _, msg := range m.Messages {
			hist = append(hist, history.HistoryItem{
				Timestamp: msg.Time,
				Role:      msg.Role,
				Content:   msg.Content,
			})
		}
		_ = history.AppendSession(hist)
		return m, tea.Quit, true
	}

	// Help reference
	if cmd == "/help" || cmd == "/?" {
		helpText := "### ⚡ Kernel Commands & Tools\n\n" +
			"- `/help` — Display this command reference\n" +
			"- `/clear` — Clear current chat screen\n" +
			"- `/pet` — Pet companion (+20 XP, Affection 100%)\n" +
			"- `/skin [cat|dragon|drone|dog]` — Change companion avatar\n" +
			"- `/tree [dir]` — Visual directory tree explorer\n" +
			"- `/diff <file>` — Compare workspace file modifications\n" +
			"- `/auto <goal>` — Autonomous coding agent loop\n" +
			"- `/read <file>` — Read and inspect file contents\n" +
			"- `/web <query>` — Live zero-key web search\n" +
			"- `/voice [on|off]` — Toggle audio synthesis & chimes\n" +
			"- `/speak <text>` — Read text aloud with system TTS\n" +
			"- `/stats` — Live system telemetry, RAM & CPU dashboard\n" +
			"- `/model <name>` — Switch active model dynamically\n" +
			"- `/theme [name]` — Switch color themes (`cyberpunk`, `matrix`, `dracula`, `nord`, `tokyo`, `synthwave`)\n" +
			"- `/commit [-y]` — AI Git Conventional Commit generator & committer\n" +
			"- `/save [filename]` — Export conversation to Markdown file\n" +
			"- `/exec <cmd>` or `$ <cmd>` — Run shell commands inside Kernel\n" +
			"- `/quit` or `:q` — Save session history and exit\n"
		m.textInput.Reset()
		m.Messages = append(m.Messages, chatMessage{
			Role:    "SYSTEM",
			Content: helpText,
			Time:    time.Now().Format("15:04"),
		})
		m.updateViewport()
		return m, nil, true
	}

	// Clear viewport
	if cmd == "/clear" {
		m.textInput.Reset()
		m.Messages = []chatMessage{
			{
				Role:    "SYSTEM",
				Content: "Chat cleared. Sparky is ready for new instructions.",
				Time:    time.Now().Format("15:04"),
			},
		}
		m.updateViewport()
		return m, nil, true
	}

	// Pet Sparky / Companion
	if cmd == "/pet" {
		m.textInput.Reset()
		feedback := m.pet.Pet()
		audio.PlayCompanionSound(string(m.pet.Skin))
		m.Messages = append(m.Messages, chatMessage{
			Role:    "SYSTEM",
			Content: fmt.Sprintf("%s\nLevel %d: [%s] | Affection: %d%%", feedback, m.pet.Level, m.pet.Title(), m.pet.Affection),
			Time:    time.Now().Format("15:04"),
		})
		m.updateViewport()
		return m, nil, true
	}

	// Dynamic Model Switching
	if cmd == "/model" {
		m.textInput.Reset()
		if arg == "" {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: fmt.Sprintf("Current Model: [%s]. To switch, type `/model <model-name>`.", m.aiClient.Model),
				Time:    time.Now().Format("15:04"),
			})
		} else {
			m.cfg.Model = arg
			_ = config.SaveConfig(m.cfg)
			m.aiClient = ai.NewClientFromConfig(m.cfg)
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: fmt.Sprintf("Switched active model to [%s]. Configuration updated.", m.cfg.Model),
				Time:    time.Now().Format("15:04"),
			})
		}
		m.updateViewport()
		return m, nil, true
	}

	// Export session
	if cmd == "/save" {
		m.textInput.Reset()
		var hist []history.HistoryItem
		for _, msg := range m.Messages {
			hist = append(hist, history.HistoryItem{
				Timestamp: msg.Time,
				Role:      msg.Role,
				Content:   msg.Content,
			})
		}
		savedFile, err := history.ExportMarkdown(arg, hist)
		if err != nil {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "ERROR",
				Content: fmt.Sprintf("Failed to export chat: %v", err),
				Time:    time.Now().Format("15:04"),
			})
		} else {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: fmt.Sprintf("Session successfully saved to: `%s`", savedFile),
				Time:    time.Now().Format("15:04"),
			})
		}
		m.updateViewport()
		return m, nil, true
	}

	// Live Web Search (/web <query>)
	if cmd == "/web" {
		m.textInput.Reset()
		if arg == "" {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "Usage: `/web <search-query>` (e.g. `/web golang bubbletea tutorials`)",
				Time:    time.Now().Format("15:04"),
			})
			m.updateViewport()
			return m, nil, true
		}

		m.Messages = append(m.Messages, chatMessage{
			Role:    "YOU",
			Content: "/web " + arg,
			Time:    time.Now().Format("15:04"),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		results, err := tools.SearchWeb(ctx, arg)
		if err != nil {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "ERROR",
				Content: fmt.Sprintf("Web search error: %v", err),
				Time:    time.Now().Format("15:04"),
			})
		} else {
			formatted := tools.FormatSearchResults(arg, results)
			m.Messages = append(m.Messages, chatMessage{
				Role:    "KERNEL",
				Content: formatted,
				Time:    time.Now().Format("15:04"),
			})
			m.pet.AddXP(10)
			m.pet.SetState(pet.StateHappy)
		}
		m.updateViewport()
		return m, nil, true
	}

	// Read File (/read <path>)
	if cmd == "/read" {
		m.textInput.Reset()
		if arg == "" {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "Usage: `/read <relative-or-absolute-filepath>`",
				Time:    time.Now().Format("15:04"),
			})
			m.updateViewport()
			return m, nil, true
		}

		m.Messages = append(m.Messages, chatMessage{
			Role:    "YOU",
			Content: "/read " + arg,
			Time:    time.Now().Format("15:04"),
		})

		content, err := tools.ReadWorkspaceFile(arg)
		if err != nil {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "ERROR",
				Content: fmt.Sprintf("Failed to read file: %v", err),
				Time:    time.Now().Format("15:04"),
			})
		} else {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "KERNEL",
				Content: fmt.Sprintf("### File: `%s`\n\n```\n%s\n```", arg, content),
				Time:    time.Now().Format("15:04"),
			})
			m.pet.AddXP(10)
		}
		m.updateViewport()
		return m, nil, true
	}

	// Companion Skin Switcher (/skin [dog|cat|dragon|drone])
	if cmd == "/skin" {
		m.textInput.Reset()
		if arg == "" {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "Available skins:\n" + pet.AvailableSkins() + "\n\nUsage: `/skin <skin-name>`",
				Time:    time.Now().Format("15:04"),
			})
		} else {
			feedback := m.pet.SetSkin(pet.SkinType(strings.ToLower(arg)))
			audio.PlayCompanionSound(string(m.pet.Skin))
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: feedback,
				Time:    time.Now().Format("15:04"),
			})
		}
		m.updateViewport()
		return m, nil, true
	}

	// Toggle Voice & Audio Synthesis (/voice [on|off])
	if cmd == "/voice" {
		m.textInput.Reset()
		if strings.ToLower(arg) == "off" || strings.ToLower(arg) == "false" {
			audio.VoiceEnabled = false
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "Audio chimes and speech synthesis muted.",
				Time:    time.Now().Format("15:04"),
			})
		} else {
			audio.VoiceEnabled = true
			audio.Bark()
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "Audio chimes and speech synthesis enabled!",
				Time:    time.Now().Format("15:04"),
			})
		}
		m.updateViewport()
		return m, nil, true
	}

	// Text-to-Speech Speak (/speak <text>)
	if cmd == "/speak" {
		m.textInput.Reset()
		if arg == "" {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "Usage: `/speak <phrase-to-speak>`",
				Time:    time.Now().Format("15:04"),
			})
		} else {
			audio.Speak(arg)
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: fmt.Sprintf("🗣️ Speaking: \"%s\"", arg),
				Time:    time.Now().Format("15:04"),
			})
		}
		m.updateViewport()
		return m, nil, true
	}

	// Live System Telemetry (/stats or /sys)
	if cmd == "/stats" || cmd == "/sys" {
		m.textInput.Reset()
		telem := tools.GetSystemTelemetry()
		formatted := tools.FormatSystemTelemetry(telem)
		m.Messages = append(m.Messages, chatMessage{
			Role:    "KERNEL",
			Content: formatted,
			Time:    time.Now().Format("15:04"),
		})
		m.pet.AddXP(10)
		m.updateViewport()
		return m, nil, true
	}

	// Interactive Directory Tree (/tree [dir])
	if cmd == "/tree" {
		m.textInput.Reset()
		dir := arg
		if dir == "" {
			dir = "."
		}
		treeOut, err := tools.RenderDirectoryTree(dir, 3)
		if err != nil {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "ERROR",
				Content: fmt.Sprintf("Tree error: %v", err),
				Time:    time.Now().Format("15:04"),
			})
		} else {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "KERNEL",
				Content: treeOut,
				Time:    time.Now().Format("15:04"),
			})
			m.pet.AddXP(10)
		}
		m.updateViewport()
		return m, nil, true
	}

	// File Diff Tool (/diff <file>)
	if cmd == "/diff" {
		m.textInput.Reset()
		if arg == "" {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "Usage: `/diff <file-path>`",
				Time:    time.Now().Format("15:04"),
			})
		} else {
			res := exec.RunShellCommand(context.Background(), fmt.Sprintf("git diff %s", arg), 10*time.Second)
			diffText := res.Stdout
			if diffText == "" {
				diffText = fmt.Sprintf("No unstaged git changes found for `%s`.", arg)
			}
			m.Messages = append(m.Messages, chatMessage{
				Role:    "KERNEL",
				Content: fmt.Sprintf("```diff\n%s\n```", diffText),
				Time:    time.Now().Format("15:04"),
			})
		}
		m.updateViewport()
		return m, nil, true
	}

	// Autonomous Agent Mode (/auto <goal>)
	if cmd == "/auto" {
		m.textInput.Reset()
		if arg == "" {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "Usage: `/auto <goal>` (e.g. `/auto create a responsive html5 clock`)",
				Time:    time.Now().Format("15:04"),
			})
			m.updateViewport()
			return m, nil, true
		}

		m.Messages = append(m.Messages, chatMessage{
			Role:    "YOU",
			Content: "/auto " + arg,
			Time:    time.Now().Format("15:04"),
		})
		m.pet.SetState(pet.StateThinking)
		m.updateViewport()

		progressChan := make(chan agent.AgentProgress, 20)
		go agent.RunAutonomousTask(context.Background(), m.aiClient, arg, progressChan)

		// Synchronously collect progress steps for display
		for p := range progressChan {
			if p.Type == agent.StepWriting || p.Type == agent.StepRunning {
				m.Messages = append(m.Messages, chatMessage{
					Role:    "SYSTEM",
					Content: p.Message,
					Time:    time.Now().Format("15:04"),
				})
			} else if p.Type == agent.StepDone {
				m.Messages = append(m.Messages, chatMessage{
					Role:    "KERNEL",
					Content: p.Message,
					Time:    time.Now().Format("15:04"),
				})
				audio.LevelUpChime()
				m.pet.AddXP(25)
				m.pet.SetState(pet.StateHappy)
			} else if p.Type == agent.StepError {
				m.Messages = append(m.Messages, chatMessage{
					Role:    "ERROR",
					Content: p.Message,
					Time:    time.Now().Format("15:04"),
				})
				m.pet.SetState(pet.StateIdle)
			}
		}

		m.updateViewport()
		return m, nil, true
	}

	// Shell command execution ($ or /exec)
	if cmd == "/exec" || strings.HasPrefix(input, "$") {
		m.textInput.Reset()
		cmdToRun := arg
		if strings.HasPrefix(input, "$") {
			cmdToRun = strings.TrimSpace(strings.TrimPrefix(input, "$"))
		}
		if cmdToRun == "" {
			return m, nil, true
		}

		m.Messages = append(m.Messages, chatMessage{
			Role:    "YOU",
			Content: "$ " + cmdToRun,
			Time:    time.Now().Format("15:04"),
		})

		res := exec.RunShellCommand(context.Background(), cmdToRun, 30*time.Second)
		formatted := exec.FormatResult(res)

		m.Messages = append(m.Messages, chatMessage{
			Role:    "SYSTEM",
			Content: fmt.Sprintf("```terminal\n%s\n```", formatted),
			Time:    time.Now().Format("15:04"),
		})
		m.pet.AddXP(10)
		m.pet.SetState(pet.StateHappy)
		m.updateViewport()
		return m, nil, true
	}

	// Color Theme Switcher (/theme [name])
	if cmd == "/theme" {
		m.textInput.Reset()
		if arg == "" {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: ListThemes(),
				Time:    time.Now().Format("15:04"),
			})
		} else {
			palette, err := ApplyTheme(arg)
			if err != nil {
				m.Messages = append(m.Messages, chatMessage{
					Role:    "ERROR",
					Content: fmt.Sprintf("Theme '%s' not recognized.\n\n%s", arg, ListThemes()),
					Time:    time.Now().Format("15:04"),
				})
			} else {
				m.cfg.Theme = palette.Name
				_ = config.SaveConfig(m.cfg)
				m.pet.AddXP(10)
				m.pet.SetState(pet.StateHappy)
				audio.PlayCompanionSound(string(m.pet.Skin))
				m.Messages = append(m.Messages, chatMessage{
					Role:    "SYSTEM",
					Content: fmt.Sprintf("🎨 Theme set to **%s** (%s)!\nSaved to `~/.kernel/config.json`.", palette.DisplayName, palette.Description),
					Time:    time.Now().Format("15:04"),
				})
			}
		}
		m.updateViewport()
		return m, nil, true
	}

	// AI Git Commit Assistant (/commit [-y])
	if cmd == "/commit" {
		m.textInput.Reset()
		changes := tools.InspectGitChanges()
		if !changes.HasChanges {
			m.Messages = append(m.Messages, chatMessage{
				Role:    "SYSTEM",
				Content: "No uncommitted git changes detected. Working tree is clean ✨",
				Time:    time.Now().Format("15:04"),
			})
			m.updateViewport()
			return m, nil, true
		}

		autoCommit := strings.Contains(arg, "-y") || strings.Contains(arg, "--yes")

		m.Messages = append(m.Messages, chatMessage{
			Role:    "YOU",
			Content: input,
			Time:    time.Now().Format("15:04"),
		})

		var commitMsg string
		if m.aiClient != nil && m.aiClient.Provider != ai.ProviderDemo {
			prompt := tools.BuildCommitPrompt(changes)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			ch := make(chan ai.StreamChunk, 20)
			go m.aiClient.Stream(ctx, prompt, ch)

			var sb strings.Builder
			for chunk := range ch {
				if chunk.Error != nil {
					break
				}
				sb.WriteString(chunk.Text)
			}
			commitMsg = strings.TrimSpace(sb.String())
		}

		if commitMsg == "" {
			commitMsg = tools.HeuristicCommitMessage(changes)
		}

		commitMsg = strings.Trim(commitMsg, "`\"'\n\r ")

		fileList := ""
		for i, f := range changes.Files {
			if i >= 6 {
				fileList += fmt.Sprintf("- ... and %d more files\n", len(changes.Files)-6)
				break
			}
			fileList += fmt.Sprintf("- `%s`\n", f)
		}

		if autoCommit {
			res := tools.RunGitCommit(commitMsg)
			if res.ExitCode == 0 {
				m.pet.AddXP(25)
				m.pet.SetState(pet.StateHappy)
				audio.Bark()
				m.Messages = append(m.Messages, chatMessage{
					Role:    "KERNEL",
					Content: fmt.Sprintf("### ⚡ Committed Successfully!\n\n**Commit Message:**\n```\n%s\n```\n\n**Committed Files (%d):**\n%s\n`git commit` exited cleanly (code 0).", commitMsg, len(changes.Files), fileList),
					Time:    time.Now().Format("15:04"),
				})
			} else {
				m.Messages = append(m.Messages, chatMessage{
					Role:    "ERROR",
					Content: fmt.Sprintf("Failed to run git commit: %s\n%s", res.Stderr, res.Stdout),
					Time:    time.Now().Format("15:04"),
				})
			}
		} else {
			m.pet.AddXP(10)
			m.Messages = append(m.Messages, chatMessage{
				Role:    "KERNEL",
				Content: fmt.Sprintf("### ⚡ AI Suggested Commit\n\n```\n%s\n```\n\n**Changed Files (%d):**\n%s\n\n*Tip:* Run `/commit -y` to stage all and commit immediately, or execute:\n```bash\ngit commit -m \"%s\"\n```", commitMsg, len(changes.Files), fileList, commitMsg),
				Time:    time.Now().Format("15:04"),
			})
		}

		m.updateViewport()
		return m, nil, true
	}

	return m, nil, false
}

func (m Model) View() string {
	var s strings.Builder

	// Top Header Bar
	title := TitleStyle.Render("⚡ KERNEL")
	providerBadge := ProviderBadgeStyle.Render(fmt.Sprintf("%s :: %s", strings.ToUpper(string(m.aiClient.Provider)), m.aiClient.Model))
	cwdBadge := SubtitleStyle.Render(fmt.Sprintf("cwd: %s", m.Cwd))
	headerBar := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", providerBadge, "  ", cwdBadge)
	s.WriteString(HeaderBox.Render(headerBar) + "\n\n")

	// Pet Companion Card with Level, Mood, and Affection Stats
	petStats := fmt.Sprintf("%s  %s\n%s\n%s",
		TitleStyle.Render("COMPANION: "+m.pet.Name),
		pet.PetMoodStyle.Render(fmt.Sprintf("[Lv. %d %s]", m.pet.Level, m.pet.Title())),
		SubtitleStyle.Render("Status: "+m.pet.Mood),
		SubtitleStyle.Render(fmt.Sprintf("Affection: %d%%  |  XP: %d/60", m.pet.Affection, m.pet.XP%60)),
	)

	petBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(0, 1).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				m.pet.RenderCustomSkin(),
				"   ",
				petStats,
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
