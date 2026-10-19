package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	PrimaryColor   = lipgloss.Color("#7D56F4") // Neon Purple / Violet
	SecondaryColor = lipgloss.Color("#00D7FE") // Electric Cyan
	AccentColor    = lipgloss.Color("#04B575") // Emerald Green
	WarningColor   = lipgloss.Color("#FFB86C") // Amber / Gold
	MutedColor     = lipgloss.Color("#6272A4") // Dim Slate
	DangerColor    = lipgloss.Color("#FF5555") // Bright Red
	BgDark         = lipgloss.Color("#13141C")
	BorderColor    = lipgloss.Color("#3B4252")

	// Header
	HeaderBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor)

	ProviderBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(PrimaryColor).
				Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Badges
	UserBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(SecondaryColor).
			Padding(0, 1)

	AgentBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(PrimaryColor).
			Padding(0, 1)

	SystemBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(WarningColor).
			Padding(0, 1)

	ErrorBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(DangerColor).
			Padding(0, 1)

	TypingDotsStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	CursorStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Blink(true)

	LogTimestamp = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Bubbles
	MessageBubble = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(0, 1)

	StreamingBubble = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(0, 1)

	InputBoxActive = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	InputBoxInactive = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(BorderColor).
				Padding(0, 1)

	// Footer
	FooterStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginTop(1)

	FooterKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor)
)
