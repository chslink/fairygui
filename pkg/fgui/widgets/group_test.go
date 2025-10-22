package widgets

import "testing"

func TestGroupDefaults(t *testing.T) {
	grp := NewGroup()
	if grp == nil || grp.GObject == nil {
		t.Fatalf("expected group to wrap GObject")
	}
}
