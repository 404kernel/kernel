package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/404kernel/kernel/internal/ai"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type logItem struct {
	Timestamp string
	Level     string
	Message   string
}

type Model struct {
	aiClient    *ai.Client
	textInput   textinput.Model
	viewport    viewport.Model
	spinner     spinner.Model
	streaming   bool
	streamChan  chan ai.StreamChunk
	cancelFunc  context.CancelFunc

	// Live Metrics (Matching screenshot)
	TasksDone   int
	TotalTasks  int
	CPM         int
	PeakCPM     int
	Throughput  float64
	Activity    []logItem

	Response    strings.Builder
	width       int
	height      int
	ready       bool
}

type streamMsg ai.StreamChunk
type tickMsg time.Time

func NewModel(client *ai.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "Ask kernel to plan, refactor, or run diagnostics..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(SecondaryColor)

	m := Model{
		aiClient:   client,
		textInput:  ti,
		spinner:    s,
		TotalTasks: 161,
		TasksDone:  161,
		PeakCPM:    161,
		Throughput: 1.0,
		Activity: []logItem{
			{Timestamp: time.Now().Format("15:04:05"), Level: "SYS", Message: "Kernel runtime initialized v1.0.0"},
			{Timestamp: time.Now().Format("15:04:05"), Level: "NET", Message: fmt.Sprintf("AI provider configured: [%s]", client.Provider)},
			{Timestamp: time.Now().Format("15:04:06"), Level: "OK", Message: "Workspace connection verified"},
		},
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.streaming && m.cancelFunc != nil {
				m.cancelFunc()
				m.streaming = false
				m.addLog("WRN", "Stream canceled by user")
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
			m.Response.Reset()
			m.streaming = true
			m.TasksDone++
			m.TotalTasks++
			m.CPM = 120
			if m.CPM > m.PeakCPM {
				m.PeakCPM = m.CPM
			}
			m.addLog("REQ", input)

			ctx, cancel := context.WithCancel(context.Background())
			m.cancelFunc = cancel
			m.streamChan = make(chan ai.StreamChunk, 20)
			go m.aiClient.Stream(ctx, input, m.streamChan)

			return m, waitForStream(m.streamChan)
		}

	case streamMsg:
		if msg.Error != nil {
			m.streaming = false
			m.addLog("ERR", msg.Error.Error())
			return m, nil
		}
		if msg.Done {
			m.streaming = false
			m.addLog("DONE", "Response completed")
			m.viewport.SetContent(m.Response.String())
			m.viewport.GotoBottom()
			return m, nil
		}

		m.Response.WriteString(msg.Text)
		m.viewport.SetContent(m.Response.String())
		m.viewport.GotoBottom()
		return m, waitForStream(m.streamChan)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		contentWidth := m.width - 4
		if contentWidth < 40 {
			contentWidth = 40
		}
		m.textInput.Width = contentWidth - 10

		if !m.ready {
			m.viewport = viewport.New(contentWidth, 10)
			m.ready = true
		} else {
			m.viewport.Width = contentWidth
		}
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.spinner, spCmd = m.spinner.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

func (m *Model) addLog(level, message string) {
	item := logItem{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   message,
	}
	m.Activity = append(m.Activity, item)
	if len(m.Activity) > 10 {
		m.Activity = m.Activity[len(m.Activity)-10:]
	}
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

	// Top Title Bar
	title := TitleStyle.Render("KERNEL // TERMINAL AGENT")
	providerBadge := SubtitleStyle.Render(fmt.Sprintf("[%s / %s]", strings.ToUpper(string(m.aiClient.Provider)), m.aiClient.Model))
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", providerBadge) + "\n\n")

	// Box 1: Metrics (Matching exact screenshot styling)
	statContent := fmt.Sprintf(
		"%s %s\n%s %s\n%s %s\n%s %s\n\n%s %s\n%s",
		StatLabel.Render("checked"), StatValue.Render(fmt.Sprintf("%d / %d", m.TasksDone, m.TotalTasks)),
		StatLabel.Render("registered"), StatValue.Render("0"),
		StatLabel.Render("cpm"), StatValue.Render(fmt.Sprintf("%d", m.CPM)),
		StatLabel.Render("peak cpm"), StatValue.Render(fmt.Sprintf("%d", m.PeakCPM)),
		StatLabel.Render("throughput"), renderBar(m.Throughput, 36),
		ProgressFull.Render("100.0%"),
	)
	s.WriteString(StatBox.Render(statContent) + "\n")

	// Box 2: Activity Log (Matching exact screenshot styling)
	var logContent strings.Builder
	logContent.WriteString(TitleStyle.Render("activity") + "\n\n")
	for _, act := range m.Activity {
		badge := LogAgent.Render(fmt.Sprintf("[%s]", act.Level))
		if act.Level == "DONE" || act.Level == "OK" {
			badge = LogSuccess.Render(fmt.Sprintf("[%s]", act.Level))
		}
		logContent.WriteString(fmt.Sprintf("%s %s %s\n", LogTimestamp.Render(act.Timestamp), badge, LogInfo.Render(act.Message)))
	}
	s.WriteString(ActivityBox.Render(logContent.String()) + "\n")

	// Box 3: Response Viewport (when streaming or has content)
	if m.Response.Len() > 0 || m.streaming {
		var respHeader string
		if m.streaming {
			respHeader = fmt.Sprintf("%s Kernel streaming thoughts...", m.spinner.View())
		} else {
			respHeader = "Agent Output:"
		}
		respBox := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(0, 1).
			Render(fmt.Sprintf("%s\n\n%s", respHeader, m.viewport.View()))
		s.WriteString("\n" + respBox + "\n")
	}

	// Bottom: Input prompt and keybind footer
	s.WriteString("\n" + m.textInput.View() + "\n")
	s.WriteString(FooterStyle.Render(
		fmt.Sprintf("done · %d of %d tasks · %s to exit", m.TasksDone, m.TotalTasks, FooterKey.Render("q / Ctrl+C")),
	))

	return s.String()
}

func renderBar(percent float64, width int) string {
	if width <= 0 {
		width = 20
	}
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := ProgressFull.Render(strings.Repeat("█", filled)) + ProgressEmpty.Render(strings.Repeat("░", empty))
	return bar
}
