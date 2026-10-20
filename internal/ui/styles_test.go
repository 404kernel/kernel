package ui

import (
	"testing"
)

func TestApplyTheme(t *testing.T) {
	themes := []string{"cyberpunk", "matrix", "dracula", "nord", "tokyo", "synthwave"}
	for _, name := range themes {
		palette, err := ApplyTheme(name)
		if err != nil {
			t.Fatalf("expected theme '%s' to apply without error, got %v", name, err)
		}
		if palette.Name != name {
			t.Errorf("expected palette.Name to be '%s', got '%s'", name, palette.Name)
		}
		if CurrentThemeName != name {
			t.Errorf("expected CurrentThemeName to be '%s', got '%s'", name, CurrentThemeName)
		}
	}

	// Test aliases
	aliasPalette, err := ApplyTheme("tokyo-night")
	if err != nil || aliasPalette.Name != "tokyo" {
		t.Fatalf("expected tokyo-night alias to resolve to tokyo, got %v", err)
	}

	// Test unknown theme
	_, err = ApplyTheme("non-existent-theme")
	if err == nil {
		t.Errorf("expected error for unknown theme, got nil")
	}
}

func TestListThemes(t *testing.T) {
	out := ListThemes()
	if len(out) == 0 {
		t.Fatal("expected ListThemes output to be non-empty")
	}
}
