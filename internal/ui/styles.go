package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ThemePalette struct {
	Name        string
	DisplayName string
	Description string
	Primary     string
	Secondary   string
	Accent      string
	Warning     string
	Muted       string
	Danger      string
	BgDark      string
	Border      string
}

var AvailableThemes = map[string]ThemePalette{
	"cyberpunk": {
		Name:        "cyberpunk",
		DisplayName: "⚡ Cyberpunk",
		Description: "Neon violet & electric cyan",
		Primary:     "#7D56F4",
		Secondary:   "#00D7FE",
		Accent:      "#04B575",
		Warning:     "#FFB86C",
		Muted:       "#6272A4",
		Danger:      "#FF5555",
		BgDark:      "#13141C",
		Border:      "#3B4252",
	},
	"matrix": {
		Name:        "matrix",
		DisplayName: "🟢 Matrix",
		Description: "Phosphor green mainframe",
		Primary:     "#00FF66",
		Secondary:   "#00DD44",
		Accent:      "#33FF33",
		Warning:     "#CCFF00",
		Muted:       "#1F6B38",
		Danger:      "#FF3333",
		BgDark:      "#0A140D",
		Border:      "#0D471C",
	},
	"dracula": {
		Name:        "dracula",
		DisplayName: "🧛 Dracula",
		Description: "Gothic vampire purple & pastel",
		Primary:     "#BD93F9",
		Secondary:   "#8BE9FD",
		Accent:      "#50FA7B",
		Warning:     "#F1FA8C",
		Muted:       "#6272A4",
		Danger:      "#FF5555",
		BgDark:      "#282A36",
		Border:      "#44475A",
	},
	"nord": {
		Name:        "nord",
		DisplayName: "❄️ Nord",
		Description: "Arctic blue & icy aurora",
		Primary:     "#88C0D0",
		Secondary:   "#81A1C1",
		Accent:      "#A3BE8C",
		Warning:     "#EBCB8B",
		Muted:       "#4C566A",
		Danger:      "#BF616A",
		BgDark:      "#2E3440",
		Border:      "#3B4252",
	},
	"tokyo": {
		Name:        "tokyo",
		DisplayName: "🗼 Tokyo Night",
		Description: "Tokyo night storm & lavender",
		Primary:     "#7AA2F7",
		Secondary:   "#BB9AF7",
		Accent:      "#9ECE6A",
		Warning:     "#E0AF68",
		Muted:       "#565F89",
		Danger:      "#F7768E",
		BgDark:      "#1A1B26",
		Border:      "#292E42",
	},
	"synthwave": {
		Name:        "synthwave",
		DisplayName: "🌆 Synthwave",
		Description: "Sunset magenta & neon grid",
		Primary:     "#FF007F",
		Secondary:   "#00F0FF",
		Accent:      "#FFE600",
		Warning:     "#FF7700",
		Muted:       "#7957D5",
		Danger:      "#FF1744",
		BgDark:      "#18092B",
		Border:      "#4B0082",
	},
}

var (
	CurrentThemeName = "cyberpunk"

	// Colors
	PrimaryColor   lipgloss.TerminalColor
	SecondaryColor lipgloss.TerminalColor
	AccentColor    lipgloss.TerminalColor
	WarningColor   lipgloss.TerminalColor
	MutedColor     lipgloss.TerminalColor
	DangerColor    lipgloss.TerminalColor
	BgDark         lipgloss.TerminalColor
	BorderColor    lipgloss.TerminalColor

	// Header
	HeaderBox          lipgloss.Style
	TitleStyle         lipgloss.Style
	ProviderBadgeStyle lipgloss.Style
	SubtitleStyle      lipgloss.Style

	// Badges
	UserBadge       lipgloss.Style
	AgentBadge      lipgloss.Style
	SystemBadge     lipgloss.Style
	ErrorBadge      lipgloss.Style
	TypingDotsStyle lipgloss.Style
	CursorStyle     lipgloss.Style
	LogTimestamp    lipgloss.Style

	// Bubbles
	MessageBubble    lipgloss.Style
	StreamingBubble  lipgloss.Style
	InputBoxActive   lipgloss.Style
	InputBoxInactive lipgloss.Style

	// Footer
	FooterStyle lipgloss.Style
	FooterKey   lipgloss.Style
)

func rebuildStyles() {
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

	FooterStyle = lipgloss.NewStyle().
		Foreground(MutedColor).
		MarginTop(1)

	FooterKey = lipgloss.NewStyle().
		Bold(true).
		Foreground(SecondaryColor)
}

// ApplyTheme sets the active UI palette and dynamically updates all styles.
func ApplyTheme(name string) (ThemePalette, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "tokyonight" || key == "tokyo-night" {
		key = "tokyo"
	}

	palette, exists := AvailableThemes[key]
	if !exists {
		return AvailableThemes["cyberpunk"], fmt.Errorf("unknown theme '%s'", name)
	}

	CurrentThemeName = key
	PrimaryColor = lipgloss.Color(palette.Primary)
	SecondaryColor = lipgloss.Color(palette.Secondary)
	AccentColor = lipgloss.Color(palette.Accent)
	WarningColor = lipgloss.Color(palette.Warning)
	MutedColor = lipgloss.Color(palette.Muted)
	DangerColor = lipgloss.Color(palette.Danger)
	BgDark = lipgloss.Color(palette.BgDark)
	BorderColor = lipgloss.Color(palette.Border)

	rebuildStyles()
	return palette, nil
}

// ListThemes renders an ASCII preview table of all available themes.
func ListThemes() string {
	var b strings.Builder
	b.WriteString("### 🎨 Available Kernel Themes\n\n")
	b.WriteString("| Theme | Name | Description | Status |\n")
	b.WriteString("|---|---|---|---|\n")

	keys := []string{"cyberpunk", "matrix", "dracula", "nord", "tokyo", "synthwave"}
	for _, k := range keys {
		t := AvailableThemes[k]
		status := " "
		if k == CurrentThemeName {
			status = "★ **ACTIVE**"
		}
		b.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", t.Name, t.DisplayName, t.Description, status))
	}

	b.WriteString("\n*Usage:* `/theme <name>` (e.g. `/theme matrix`, `/theme nord`)")
	return b.String()
}

func init() {
	_, _ = ApplyTheme("cyberpunk")
}
