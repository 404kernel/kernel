package pet

import (
	"strings"
	"testing"
)

func TestPetCreationAndTransitions(t *testing.T) {
	p := NewPet("Buddy")
	if p.Name != "Buddy" {
		t.Errorf("expected name Buddy, got %s", p.Name)
	}
	if p.State != StateIdle {
		t.Errorf("expected initial state StateIdle, got %v", p.State)
	}

	idleArt := p.Render()
	if !strings.Contains(idleArt, "/^ ^\\") {
		t.Errorf("expected dog ears in idle art, got: %s", idleArt)
	}

	p.SetState(StateThinking)
	if p.Mood != "focusing..." {
		t.Errorf("expected mood focusing..., got %s", p.Mood)
	}

	p.SetState(StateHappy)
	if p.Mood != "*wags tail happily*" {
		t.Errorf("expected mood *wags tail happily*, got %s", p.Mood)
	}
}

func TestPetAnimationTick(t *testing.T) {
	p := NewPet("Sparky")
	initialFrame := p.Frame
	p.Tick()
	if p.Frame != initialFrame+1 {
		t.Errorf("expected frame increment on tick")
	}
}

func TestPetXPAndLeveling(t *testing.T) {
	p := NewPet("Sparky")
	if p.Level != 1 {
		t.Fatalf("expected initial level 1, got %d", p.Level)
	}

	leveled := p.AddXP(70)
	if !leveled || p.Level != 2 {
		t.Errorf("expected level 2 after 70 XP, got %d", p.Level)
	}

	feedback := p.Pet()
	if !strings.Contains(feedback, "+20 XP") {
		t.Errorf("expected XP grant in pet feedback, got %s", feedback)
	}
}

