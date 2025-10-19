package widgets

import "testing"

func TestLoaderDefaults(t *testing.T) {
	loader := NewLoader()
	if loader == nil || loader.GObject == nil {
		t.Fatalf("expected GLoader to wrap GObject")
	}
}
