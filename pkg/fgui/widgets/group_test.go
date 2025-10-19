package widgets

import "testing"

func TestGroupDefaults(t *testing.T) {
	grp := NewGroup()
	if grp == nil || grp.GComponent == nil {
		t.Fatalf("expected group to wrap GComponent")
	}
}
