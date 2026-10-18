package pet

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	StateIdle State = iota
	StateThinking
	StateHappy
	StateSleeping
)

type Pet struct {
	Name       string
	State      State
	Frame      int
	Mood       string
	BlinkTimer int
}

var (
	PetColor    = lipgloss.Color("#FFB86C") // Amber / Golden dog
	AccentColor = lipgloss.Color("#00D7FE") // Cyan collar/eyes
	MutedColor  = lipgloss.Color("#6272A4")

	PetBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BD93F9")).
			Padding(0, 1)

	PetMoodStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)
)

func NewPet(name string) *Pet {
	if name == "" {
		name = "Sparky"
	}
	return &Pet{
		Name:  name,
		State: StateIdle,
		Mood:  "curious",
	}
}

func (p *Pet) Tick() {
	p.Frame++
}

func (p *Pet) SetState(s State) {
	p.State = s
	switch s {
	case StateThinking:
		p.Mood = "focusing..."
	case StateHappy:
		p.Mood = "*wags tail*"
	case StateSleeping:
		p.Mood = "zzZ..."
	default:
		p.Mood = "listening"
	}
}

// Render returns a 4-line ASCII dog companion that animates smoothly
func (p *Pet) Render() string {
	switch p.State {
	case StateThinking:
		return p.renderThinking()
	case StateHappy:
		return p.renderHappy()
	case StateSleeping:
		return p.renderSleeping()
	default:
		return p.renderIdle()
	}
}

func (p *Pet) renderIdle() string {
	tail := "~"
	if p.Frame%4 >= 2 {
		tail = "-"
	}
	eyes := "• •"
	if p.Frame%12 == 0 {
		eyes = "- -" // blink
	}

	return lipgloss.NewStyle().Foreground(PetColor).Render(
		"  /^ ^\\\n" +
			" / " + eyes + " \\\n" +
			" V\\ Y /V\n" +
			"  / - \\  " + tail + "\n" +
			" (____)m",
	)
}

func (p *Pet) renderThinking() string {
	// Animated sniffing & ears twitching
	ears := "/^ ^\\"
	if p.Frame%2 == 0 {
		ears = "\\^ ^/"
	}
	gear := "⠋"
	gears := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	gear = gears[p.Frame%len(gears)]

	return lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FE")).Render(
		"  " + ears + "  " + gear + "\n" +
			" / o o \\\n" +
			" V\\ ^ /V\n" +
			"  / ~ \\  |\n" +
			" (____)m",
	)
}

func (p *Pet) renderHappy() string {
	tails := []string{"\\", "|", "/", "-"}
	tail := tails[p.Frame%len(tails)]

	return lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render(
		"  /^ ^\\   !\n" +
			" / ^ ^ \\\n" +
			" V\\ U /V\n" +
			"  / - \\  " + tail + "\n" +
			" (____)m",
	)
}

func (p *Pet) renderSleeping() string {
	snore := "z"
	if p.Frame%2 == 0 {
		snore = "Z"
	}

	return lipgloss.NewStyle().Foreground(MutedColor).Render(
		"  /- -\\   " + snore + "\n" +
			" / - - \\\n" +
			" V\\ . /V\n" +
			"  / - \\  ~\n" +
			" (____)m",
	)
}

func RandomPetDialogue(state State) string {
	switch state {
	case StateThinking:
		quotes := []string{"Crunching tokens...", "Digging through the codebase...", "Sniffing out solutions..."}
		return quotes[rand.Intn(len(quotes))]
	case StateHappy:
		quotes := []string{"All done, boss!", "Woof! Check that out.", "Tail wagging at 100%!"}
		return quotes[rand.Intn(len(quotes))]
	default:
		return "Ready for your next instruction."
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
