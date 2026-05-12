package gears

import "testing"

func TestGearXYApply(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "page1", selectedIndex: 0}
	gear := NewGearXY(owner)
	gear.SetController(ctrl)

	// Store position for "page1"
	owner.SetPosition(100, 200)
	gear.UpdateState()

	// Move owner away
	owner.SetPosition(0, 0)

	// Apply should restore position
	gear.Apply()
	if owner.X() != 100 || owner.Y() != 200 {
		t.Errorf("expected position (100,200), got (%.0f,%.0f)", owner.X(), owner.Y())
	}
}

func TestGearXYPercentMode(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "p1", selectedIndex: 0}
	gear := NewGearXY(owner)
	gear.SetController(ctrl)

	gear.SetPositionsInPercent(true)
	if !gear.PositionsInPercent() {
		t.Error("expected PositionsInPercent=true")
	}

	owner.SetPosition(50, 75)
	gear.UpdateState()

	// In percent mode, values should reflect percentages
	owner.SetPosition(0, 0)
	gear.Apply()
	// In percent mode, position = percentage * parentSize
	expectedX := 50.0 // stored percentage
	expectedY := 75.0
	if owner.X() != expectedX || owner.Y() != expectedY {
		t.Errorf("percent mode: expected (%.0f,%.0f), got (%.0f,%.0f)", expectedX, expectedY, owner.X(), owner.Y())
	}
}

func TestGearXYDefaultValue(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "unknown", selectedIndex: 99}
	gear := NewGearXY(owner)
	gear.SetController(ctrl)

	// Set default value
	owner.SetPosition(42, 84)
	gear.UpdateState()

	// Apply with unknown page should use default
	owner.SetPosition(0, 0)
	gear.Apply()
	if owner.X() != 42 || owner.Y() != 84 {
		t.Errorf("expected default (42,84), got (%.0f,%.0f)", owner.X(), owner.Y())
	}
}

func TestGearXYUpdateFromRelations(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "p1", selectedIndex: 0}
	gear := NewGearXY(owner)
	gear.SetController(ctrl)

	owner.SetPosition(10, 20)
	gear.UpdateState()

	// UpdateFromRelations adjusts stored values when relations move the owner
	// Should not panic
	gear.UpdateFromRelations(5, 5)
}
