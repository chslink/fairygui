package widgets

import "testing"

func TestButtonDefaults(t *testing.T) {
	btn := NewButton()
	if btn == nil || btn.GComponent == nil {
		t.Fatalf("expected GButton to wrap GComponent")
	}
}
