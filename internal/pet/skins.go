package pet

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type SkinType string

const (
	SkinDog    SkinType = "dog"
	SkinCat    SkinType = "cat"
	SkinDragon SkinType = "dragon"
	SkinDrone  SkinType = "drone"
)

// SetSkin changes the pet's visual form and default name
func (p *Pet) SetSkin(skin SkinType) string {
	switch skin {
	case SkinCat:
		p.Skin = SkinCat
		if p.Name == "Sparky" {
			p.Name = "Nyx"
		}
		return "Transformed into Nyx the Cyber Cat! 🐱"
	case SkinDragon:
		p.Skin = SkinDragon
		if p.Name == "Sparky" {
			p.Name = "Draco"
		}
		return "Transformed into Draco the Pixel Dragon! 🐉"
	case SkinDrone:
		p.Skin = SkinDrone
		if p.Name == "Sparky" {
			p.Name = "Orbit"
		}
		return "Transformed into Orbit the Sentinel Drone! 🛸"
	default:
		p.Skin = SkinDog
		p.Name = "Sparky"
		return "Switched to Sparky the Terminal Doggo! 🐶"
	}
}

// RenderCustomSkin routes the frame render based on current skin
func (p *Pet) RenderCustomSkin() string {
	switch p.Skin {
	case SkinCat:
		return p.renderCat()
	case SkinDragon:
		return p.renderDragon()
	case SkinDrone:
		return p.renderDrone()
	default:
		return p.Render()
	}
}

func (p *Pet) renderCat() string {
	whisker := "="
	if p.Frame%4 >= 2 {
		whisker = "~"
	}
	eyes := "• •"
	if p.Frame%14 == 0 {
		eyes = "- -"
	}
	if p.State == StateThinking {
		eyes = "o o"
	} else if p.State == StateHappy {
		eyes = "^ ^"
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Render(
		fmt.Sprintf("  /\\_/\\\n ( %s )\n%s(   )%s\n  (__)m", eyes, whisker, whisker),
	)
}

func (p *Pet) renderDragon() string {
	wings := "\\/  \\/"
	if p.Frame%4 >= 2 {
		wings = "--  --"
	}
	flame := ""
	if p.State == StateThinking || p.State == StateHappy {
		flame = " ~🔥"
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render(
		fmt.Sprintf(" %s\n <' )~%s\n  (  )\n  ^--^", wings, flame),
	)
}

func (p *Pet) renderDrone() string {
	beacons := []string{"[ ⊙ ]", "[ ◎ ]", "[ ● ]", "[ ○ ]"}
	beacon := beacons[(p.Frame/2)%len(beacons)]
	rings := "<" + beacon + ">"

	pulse := "···"
	if p.State == StateThinking {
		pulse = "⚡⚡⚡"
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FE")).Render(
		fmt.Sprintf("   ===\n  %s\n   %s\n   ===", rings, pulse),
	)
}

// AvailableSkins returns the list of skins for user display
func AvailableSkins() string {
	skins := []string{
		"• `dog` — Sparky the Terminal Doggo 🐶",
		"• `cat` — Nyx the Cyber Cat 🐱",
		"• `dragon` — Draco the Pixel Dragon 🐉",
		"• `drone` — Orbit the Sentinel Drone 🛸",
	}
	return strings.Join(skins, "\n")
}
