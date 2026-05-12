package gears

import "testing"

func TestGearDisplay2ConnectedWithController(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "a", selectedIndex: 0}

	gear := NewGearDisplay2(owner)
	gear.pages = []string{"a", "b"}
	gear.SetController(ctrl)

	// ctrl on "a" → visibleCounter = 1 → visible
	if !gear.Evaluate(true) {
		t.Error("expected Evaluate(true)=true when controller page matches")
	}

	// Change to non-matching page
	ctrl.selectedPageID = "c"
	gear.Apply()
	if gear.Evaluate(true) {
		t.Error("expected Evaluate(true)=false when page 'c' not in ['a','b']")
	}
}

func TestGearDisplay2NoController(t *testing.T) {
	owner := newMockOwner()
	gear := NewGearDisplay2(owner)

	// No controller → GearDisplay2 returns connected as-is
	if !gear.Evaluate(true) {
		t.Error("expected Evaluate(true)=true with no controller")
	}
	if gear.Evaluate(false) {
		t.Error("expected Evaluate(false)=false")
	}
}

func TestGearDisplay2OrCondition(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "c", selectedIndex: 2}

	gear := NewGearDisplay2(owner)
	gear.pages = []string{"a", "b"}
	gear.condition = 1 // OR mode
	gear.SetController(ctrl) // calls Apply internally

	// Page "c" not in pages → visibleCounter = 0
	// condition=1 (OR): visible || connected = false || true = true
	if !gear.Evaluate(true) {
		t.Error("expected Evaluate(true)=true in OR mode regardless of page match")
	}
	// OR mode: false || false = false
	if gear.Evaluate(false) {
		t.Error("expected Evaluate(false)=false in OR mode (both false)")
	}
}
