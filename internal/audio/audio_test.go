package audio

import (
	"testing"
	"time"
)

func TestPlayCompanionSound(t *testing.T) {
	if len(dogBarkWAV) == 0 {
		t.Fatal("expected embedded dogBarkWAV to be non-empty")
	}
	if len(catMeowWAV) == 0 {
		t.Fatal("expected embedded catMeowWAV to be non-empty")
	}

	// Test calls without panic
	PlayCompanionSound("dog")
	PlayCompanionSound("cat")
	PlayCompanionSound("dragon")
	PlayCompanionSound("drone")
	Bark()
	time.Sleep(100 * time.Millisecond)
}
