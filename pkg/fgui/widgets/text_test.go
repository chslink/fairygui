package widgets

import "testing"

func TestTextDefaults(t *testing.T) {
	txt := NewText()
	if txt == nil || txt.GObject == nil {
		t.Fatalf("expected GTextField to wrap GObject")
	}
	txt.SetText("hello")
	if txt.Text() != "hello" {
		t.Fatalf("text setter/getter mismatch")
	}
}
