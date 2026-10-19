package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestInitThemeDark(t *testing.T) {
	*flagTheme = "dark"
	InitTheme()
	if !lipgloss.HasDarkBackground() {
		t.Error("expected dark background after InitTheme with 'dark'")
	}
}

func TestInitThemeInvalid(t *testing.T) {
	oldTheme, oldExit := *flagTheme, exitInvalidTheme
	t.Cleanup(func() { *flagTheme, exitInvalidTheme = oldTheme, oldExit })
	*flagTheme = "invalid"
	exitCode := 0
	exitInvalidTheme = func(code int) { exitCode = code }
	InitTheme()
	if exitCode != 1 {
		t.Fatalf("invalid theme exit code=%d", exitCode)
	}
}

func TestInitThemeLight(t *testing.T) {
	*flagTheme = "light"
	InitTheme()
	if lipgloss.HasDarkBackground() {
		t.Error("expected light background after InitTheme with 'light'")
	}
}

func TestInitThemeAuto(t *testing.T) {
	// auto should not panic
	*flagTheme = "auto"
	InitTheme()
}
