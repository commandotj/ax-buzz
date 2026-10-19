package ui

import "testing"

func TestOutputViewShortOutput(t *testing.T) {
	m := NewOutputModel()
	m.Append("one")
	if m.View() == "" {
		t.Fatal("short output should render")
	}
}
