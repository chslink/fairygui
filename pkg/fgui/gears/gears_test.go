package gears

import "testing"

func TestGearSizeApply(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "p1", selectedIndex: 0}
	gear := NewGearSize(owner)
	gear.SetController(ctrl)

	owner.SetSize(200, 100)
	owner.SetScale(1.5, 1.5)
	gear.UpdateState()

	owner.SetSize(0, 0)
	owner.SetScale(1, 1)
	gear.Apply()

	if owner.Width() != 200 || owner.Height() != 100 {
		t.Errorf("expected size (200,100), got (%.0f,%.0f)", owner.Width(), owner.Height())
	}
	scaleX, scaleY := owner.Scale()
	if scaleX != 1.5 || scaleY != 1.5 {
		t.Errorf("expected scale (1.5,1.5), got (%.1f,%.1f)", scaleX, scaleY)
	}
}

func TestGearSizeDefaultValue(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "unknown", selectedIndex: -1}
	gear := NewGearSize(owner)
	gear.SetController(ctrl)

	owner.SetSize(300, 200)
	gear.UpdateState()

	owner.SetSize(0, 0)
	gear.Apply()
	if owner.Width() != 300 || owner.Height() != 200 {
		t.Errorf("expected default size (300,200), got (%.0f,%.0f)", owner.Width(), owner.Height())
	}
}

func TestGearColorApply(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "red", selectedIndex: 0}
	gear := NewGearColor(owner)
	gear.SetController(ctrl)

	owner.SetProp(ObjectPropIDColor, "#ff0000")
	owner.SetProp(ObjectPropIDOutlineColor, "#00ff00")
	gear.UpdateState()

	owner.SetProp(ObjectPropIDColor, "#000000")
	owner.SetProp(ObjectPropIDOutlineColor, "#000000")
	gear.Apply()

	if color, ok := owner.GetProp(ObjectPropIDColor).(string); !ok || color != "#ff0000" {
		t.Errorf("expected color '#ff0000', got '%v'", owner.GetProp(ObjectPropIDColor))
	}
	if outline, ok := owner.GetProp(ObjectPropIDOutlineColor).(string); !ok || outline != "#00ff00" {
		t.Errorf("expected outline '#00ff00', got '%v'", owner.GetProp(ObjectPropIDOutlineColor))
	}
}

func TestGearLookApply(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "hidden", selectedIndex: 0}
	gear := NewGearLook(owner)
	gear.SetController(ctrl)

	owner.SetAlpha(0.5)
	owner.SetRotation(45)
	owner.SetGrayed(true)
	owner.SetTouchable(false)
	gear.UpdateState()

	owner.SetAlpha(1)
	owner.SetRotation(0)
	owner.SetGrayed(false)
	owner.SetTouchable(true)
	gear.Apply()

	if owner.Alpha() != 0.5 {
		t.Errorf("expected alpha 0.5, got %.2f", owner.Alpha())
	}
	if owner.Rotation() != 45 {
		t.Errorf("expected rotation 45, got %.0f", owner.Rotation())
	}
	if !owner.Grayed() {
		t.Error("expected grayed=true")
	}
	if owner.Touchable() {
		t.Error("expected touchable=false")
	}
}

func TestGearTextSwitch(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "p1", selectedIndex: 0}
	gear := NewGearText(owner)
	gear.SetController(ctrl)

	owner.SetProp(ObjectPropIDText, "Hello")
	gear.UpdateState()

	owner.SetProp(ObjectPropIDText, "")
	gear.Apply()

	if text, ok := owner.GetProp(ObjectPropIDText).(string); !ok || text != "Hello" {
		t.Errorf("expected text 'Hello', got '%v'", owner.GetProp(ObjectPropIDText))
	}
}

func TestGearIconSwitch(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "p1", selectedIndex: 0}
	gear := NewGearIcon(owner)
	gear.SetController(ctrl)

	owner.SetProp(ObjectPropIDIcon, "ui://icon1")
	gear.UpdateState()

	owner.SetProp(ObjectPropIDIcon, "")
	gear.Apply()

	if icon, ok := owner.GetProp(ObjectPropIDIcon).(string); !ok || icon != "ui://icon1" {
		t.Errorf("expected icon 'ui://icon1', got '%v'", owner.GetProp(ObjectPropIDIcon))
	}
}

func TestGearAnimationSwitch(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "play", selectedIndex: 0}
	gear := NewGearAnimation(owner)
	gear.SetController(ctrl)

	owner.SetProp(ObjectPropIDPlaying, true)
	owner.SetProp(ObjectPropIDFrame, 5)
	gear.UpdateState()

	owner.SetProp(ObjectPropIDPlaying, false)
	owner.SetProp(ObjectPropIDFrame, 0)
	gear.Apply()

	if playing, ok := owner.GetProp(ObjectPropIDPlaying).(bool); !ok || !playing {
		t.Errorf("expected playing=true, got '%v'", owner.GetProp(ObjectPropIDPlaying))
	}
	if frame, ok := owner.GetProp(ObjectPropIDFrame).(int); !ok || frame != 5 {
		t.Errorf("expected frame=5, got '%v'", owner.GetProp(ObjectPropIDFrame))
	}
}

func TestGearFontSizeSwitch(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "p1", selectedIndex: 0}
	gear := NewGearFontSize(owner)
	gear.SetController(ctrl)

	owner.SetProp(ObjectPropIDFontSize, 24)
	gear.UpdateState()

	owner.SetProp(ObjectPropIDFontSize, 12)
	gear.Apply()

	if fs, ok := owner.GetProp(ObjectPropIDFontSize).(int); !ok || fs != 24 {
		t.Errorf("expected font size 24, got '%v'", owner.GetProp(ObjectPropIDFontSize))
	}
}
