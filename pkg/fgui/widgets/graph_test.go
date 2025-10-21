package widgets

import "testing"

func TestGraphDefaults(t *testing.T) {
	g := NewGraph()
	if g == nil || g.GObject == nil {
		t.Fatalf("expected graph to wrap GObject")
	}
	if g.Type() != GraphTypeEmpty {
		t.Fatalf("expected default type empty")
	}
	if g.FillColor() == "" {
		t.Fatalf("expected default fill colour")
	}
	if g.LineColor() == "" {
		t.Fatalf("expected default line colour")
	}
}
