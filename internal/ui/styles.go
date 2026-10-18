package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Palettes
	PrimaryColor   = lipgloss.Color("#7D56F4") // Neon Purple / Violet
	SecondaryColor = lipgloss.Color("#00D7FE") // Electric Cyan
	AccentColor    = lipgloss.Color("#04B575") // Emerald Green
	WarningColor   = lipgloss.Color("#FFB86C") // Amber / Orange
	MutedColor     = lipgloss.Color("#6272A4") // Dim Gray / Muted Purple
	DangerColor    = lipgloss.Color("#FF5555") // Bright Red
	BgDark         = lipgloss.Color("#13141C")
	BorderColor    = lipgloss.Color("#3B4252")

	// Box Styles matching the modern terminal screenshot
	HeaderBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1).
			Bold(true).
			Foreground(SecondaryColor)

	StatBox = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		Padding(0, 1).
		MarginBottom(1)

	ActivityBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	StatLabel = lipgloss.NewStyle().
			Foreground(MutedColor).
			Width(14)

	StatValue = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	ProgressFull = lipgloss.NewStyle().
			Foreground(SecondaryColor)

	ProgressEmpty = lipgloss.NewStyle().
			Foreground(BorderColor)

	LogTimestamp = lipgloss.NewStyle().
			Foreground(MutedColor)

	LogAgent = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	LogSuccess = lipgloss.NewStyle().
			Foreground(AccentColor)

	LogInfo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2"))

	FooterStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginTop(1)

	FooterKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor)
)
