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
	if p.Mood != "*wags tail*" {
		t.Errorf("expected mood *wags tail*, got %s", p.Mood)
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
