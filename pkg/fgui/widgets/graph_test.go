package widgets

import "testing"

func TestGraphDefaults(t *testing.T) {
	g := NewGraph()
	if g == nil || g.GObject == nil {
		t.Fatalf("expected graph to wrap GObject")
	}
}
