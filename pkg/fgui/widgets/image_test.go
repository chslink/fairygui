package widgets

import "testing"

func TestImageDefaults(t *testing.T) {
	img := NewImage()
	if img == nil || img.GObject == nil {
		t.Fatalf("expected GImage to wrap GObject")
	}
}
