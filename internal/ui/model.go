package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/404kernel/kernel/internal/ai"
	"github.com/404kernel/kernel/internal/pet"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type chatMessage struct {
	Role    string
	Content string
	Time    string
}

type Model struct {
	aiClient   *ai.Client
	pet        *pet.Pet
	textInput  textinput.Model
	viewport   viewport.Model
	spinner    spinner.Model
	streaming  bool
	streamChan chan ai.StreamChunk
	cancelFunc context.CancelFunc

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
	return tea.Tick(180*time.Millisecond, func(t time.Time) tea.Msg {
		return petTickMsg(t)
	})
}

func NewModel(client *ai.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "Ask kernel to explain code, generate commands, or solve problems..."
	ti.Focus()
	ti.CharLimit = 1024
	ti.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(SecondaryColor)

	p := pet.NewPet("Sparky")

	cwd, _ := os.Getwd()

	m := Model{
		aiClient:  client,
		pet:       p,
		textInput: ti,
		spinner:   s,
		Cwd:       cwd,
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
		return m, petTicker()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
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
		case tea.KeyEnter:
			if m.streaming {
				return m, nil
			}
			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}
			if input == "q" || input == "exit" {
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
			m.streamChan = make(chan ai.StreamChunk, 30)
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

	if m.streaming && m.CurrentResp.Len() > 0 {
		header := fmt.Sprintf("%s %s %s", AgentBadge.Render(" KERNEL "), m.spinner.View(), LogTimestamp.Render(time.Now().Format("15:04")))
		bubble := StreamingBubble.Render(fmt.Sprintf("%s\n\n%s", header, m.CurrentResp.String()))
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

	// Pet Companion Card (Header widget)
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

	// Viewport (Chat conversation)
	s.WriteString(m.viewport.View() + "\n\n")

	// Bottom Input Field
	var inputBorder = InputBoxActive
	if m.streaming {
		inputBorder = InputBoxInactive
	}
	s.WriteString(inputBorder.Render(m.textInput.View()) + "\n")

	// Footer Keybind Hints
	s.WriteString(FooterStyle.Render(
		fmt.Sprintf("%s send · %s cancel/exit · %s quit", FooterKey.Render("Enter"), FooterKey.Render("Esc"), FooterKey.Render("q")),
	))

	return s.String()
}
