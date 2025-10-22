package widgets

import (
	"math"
	"testing"
)

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

func TestTextAutoSizeBoth(t *testing.T) {
	txt := NewText()
	txt.SetAutoSize(TextAutoSizeBoth)
	txt.GObject.SetSize(10, 10)
	txt.UpdateLayoutMetrics(42, 18, 40, 16)
	if width := txt.GObject.Width(); math.Abs(width-42) > 0.5 {
		t.Fatalf("expected width auto size to apply 42, got %.2f", width)
	}
	if height := txt.GObject.Height(); math.Abs(height-18) > 0.5 {
		t.Fatalf("expected height auto size to apply 18, got %.2f", height)
	}
	if tw := txt.TextWidth(); math.Abs(tw-40) > 0.5 {
		t.Fatalf("expected text width 40, got %.2f", tw)
	}
	if th := txt.TextHeight(); math.Abs(th-16) > 0.5 {
		t.Fatalf("expected text height 16, got %.2f", th)
	}
}

func TestTextAutoSizeHeightOnly(t *testing.T) {
	txt := NewText()
	txt.GObject.SetSize(100, 30)
	txt.SetAutoSize(TextAutoSizeHeight)
	txt.UpdateLayoutMetrics(45, 20, 40, 18)
	if width := txt.GObject.Width(); math.Abs(width-100) > 0.5 {
		t.Fatalf("width should remain 100 when width auto disabled, got %.2f", width)
	}
	if height := txt.GObject.Height(); math.Abs(height-20) > 0.5 {
		t.Fatalf("height should update to 20, got %.2f", height)
	}
}
