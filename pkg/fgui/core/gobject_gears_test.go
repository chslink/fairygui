package core

import (
	"math"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/gears"
)

func TestGObjectGearXYRelationsUpdate(t *testing.T) {
	parent := NewGComponent()
	parent.SetSize(200, 100)
	obj := NewGObject()
	parent.AddChild(obj)
	obj.SetPosition(10, 20)

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"Page0", "Page1"})
	ctrl.SetSelectedIndex(0)

	gear := obj.GetGear(gears.IndexXY)
	if gear == nil {
		t.Fatalf("expected GearXY instance")
	}
	xy, ok := gear.(*gears.GearXY)
	if !ok {
		t.Fatalf("expected *gears.GearXY, got %T", gear)
	}
	xy.SetController(ctrl)
	xy.UpdateState()

	ctrl.SetSelectedIndex(1)
	obj.SetPosition(30, 45)
	xy.UpdateState()

	ctrl.SetSelectedIndex(0)
	obj.SetPosition(12, 26)
	obj.updateGearFromRelationsSafe(gears.IndexXY, 2, 6)

	v0 := xy.Value("p0")
	if v0.X != 12 || v0.Y != 26 {
		t.Fatalf("page0 position mismatch: got (%v,%v)", v0.X, v0.Y)
	}
	v1 := xy.Value("p1")
	if v1.X != 32 || v1.Y != 51 {
		t.Fatalf("page1 position not adjusted by relations: got (%v,%v)", v1.X, v1.Y)
	}
}

func TestGObjectGearSizeRelationsUpdate(t *testing.T) {
	obj := NewGObject()
	obj.SetSize(100, 50)

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"Page0", "Page1"})

	gear := obj.GetGear(gears.IndexSize)
	if gear == nil {
		t.Fatalf("expected GearSize instance")
	}
	sizeGear, ok := gear.(*gears.GearSize)
	if !ok {
		t.Fatalf("expected *gears.GearSize, got %T", gear)
	}
	sizeGear.SetController(ctrl)
	sizeGear.UpdateState()

	ctrl.SetSelectedIndex(1)
	obj.SetSize(200, 120)
	sizeGear.UpdateState()

	ctrl.SetSelectedIndex(0)
	obj.SetSize(110, 70)
	obj.updateGearFromRelationsSafe(gears.IndexSize, 10, 20)

	v0 := sizeGear.Value("p0")
	if v0.Width != 110 || v0.Height != 70 {
		t.Fatalf("page0 size mismatch: got (%v,%v)", v0.Width, v0.Height)
	}
	v1 := sizeGear.Value("p1")
	if v1.Width != 210 || v1.Height != 140 {
		t.Fatalf("page1 size not adjusted by relations: got (%v,%v)", v1.Width, v1.Height)
	}
}

func TestControllerAppliesGearXY(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()
	parent.AddChild(child)

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"Page0", "Page1"})
	parent.AddController(ctrl)

	gear := child.GetGear(gears.IndexXY)
	xy, ok := gear.(*gears.GearXY)
	if !ok {
		t.Fatalf("expected *gears.GearXY, got %T", gear)
	}
	xy.SetController(ctrl)

	ctrl.SetSelectedIndex(0)
	child.SetPosition(10, 20)
	xy.UpdateState()

	ctrl.SetSelectedIndex(1)
	child.SetPosition(40, 60)
	xy.UpdateState()

	child.SetGearLocked(true)
	child.SetPosition(0, 0)
	child.SetGearLocked(false)
	ctrl.SetSelectedIndex(0)
	if child.X() != 10 || child.Y() != 20 {
		t.Fatalf("expected controller to restore page0 position, got (%v,%v)", child.X(), child.Y())
	}
	ctrl.SetSelectedIndex(1)
	if child.X() != 40 || child.Y() != 60 {
		t.Fatalf("expected controller to restore page1 position, got (%v,%v)", child.X(), child.Y())
	}
}

func TestControllerAppliesGearSize(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()
	parent.AddChild(child)
	child.SetSize(100, 50)
	child.SetScale(1, 1)

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"Page0", "Page1"})
	parent.AddController(ctrl)

	gear := child.GetGear(gears.IndexSize)
	sizeGear, ok := gear.(*gears.GearSize)
	if !ok {
		t.Fatalf("expected *gears.GearSize, got %T", gear)
	}
	sizeGear.SetController(ctrl)

	ctrl.SetSelectedIndex(0)
	child.SetSize(120, 70)
	child.SetScale(1.2, 1.1)
	sizeGear.UpdateState()

	ctrl.SetSelectedIndex(1)
	child.SetSize(200, 150)
	child.SetScale(0.8, 1.5)
	sizeGear.UpdateState()

	child.SetGearLocked(true)
	child.SetSize(10, 10)
	child.SetScale(1, 1)
	child.SetGearLocked(false)
	ctrl.SetSelectedIndex(0)
	if child.Width() != 120 || child.Height() != 70 {
		t.Fatalf("expected page0 size restored, got (%v,%v)", child.Width(), child.Height())
	}
	sx, sy := child.Scale()
	if sx != 1.2 || sy != 1.1 {
		t.Fatalf("expected page0 scale restored, got (%v,%v)", sx, sy)
	}
	ctrl.SetSelectedIndex(1)
	if child.Width() != 200 || child.Height() != 150 {
		t.Fatalf("expected page1 size restored, got (%v,%v)", child.Width(), child.Height())
	}
	sx, sy = child.Scale()
	if sx != 0.8 || sy != 1.5 {
		t.Fatalf("expected page1 scale restored, got (%v,%v)", sx, sy)
	}
}

func TestGearDisplayVisibility(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()
	parent.AddChild(child)

	gear := child.GetGear(gears.IndexDisplay)
	displayGear, ok := gear.(*gears.GearDisplay)
	if !ok {
		t.Fatalf("expected *gears.GearDisplay, got %T", gear)
	}
	displayGear.SetPages([]string{"page-visible"})

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"page-hidden", "page-visible"}, []string{"Hidden", "Visible"})
	parent.AddController(ctrl)
	displayGear.SetController(ctrl)

	ctrl.SetSelectedIndex(0)
	if child.Visible() {
		t.Fatalf("expected child hidden for page-hidden")
	}
	ctrl.SetSelectedIndex(1)
	if !child.Visible() {
		t.Fatalf("expected child visible for page-visible")
	}
}

func TestGearLookAppearance(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()
	parent.AddChild(child)
	child.SetAlpha(0.5)
	child.SetRotation(0.2)
	child.SetGrayed(false)
	child.SetTouchable(true)

	gear := child.GetGear(gears.IndexLook)
	lookGear, ok := gear.(*gears.GearLook)
	if !ok {
		t.Fatalf("expected *gears.GearLook, got %T", gear)
	}

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"page-a", "page-b"}, []string{"A", "B"})
	parent.AddController(ctrl)
	lookGear.SetController(ctrl)

	ctrl.SetSelectedIndex(0)
	child.SetAlpha(0.4)
	child.SetRotation(0.1)
	child.SetGrayed(true)
	child.SetTouchable(false)
	lookGear.UpdateState()

	ctrl.SetSelectedIndex(1)
	child.SetAlpha(0.9)
	child.SetRotation(0.6)
	child.SetGrayed(false)
	child.SetTouchable(true)
	lookGear.UpdateState()

	child.SetAlpha(1.0)
	child.SetRotation(1.0)
	child.SetGrayed(false)
	child.SetTouchable(true)
	ctrl.SetSelectedIndex(0)
	if child.Alpha() != 0.4 || child.Rotation() != 0.1 || !child.Grayed() || child.Touchable() {
		t.Fatalf("expected page-a appearance restored")
	}
	ctrl.SetSelectedIndex(1)
	if child.Alpha() != 0.9 || child.Rotation() != 0.6 || child.Grayed() || !child.Touchable() {
		t.Fatalf("expected page-b appearance restored")
	}
}

func TestGearTextSwitch(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()
	parent.AddChild(child)

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"P0", "P1"})
	parent.AddController(ctrl)

	gear := child.GetGear(gears.IndexText)
	textGear, ok := gear.(*gears.GearText)
	if !ok {
		t.Fatalf("expected GearText, got %T", gear)
	}
	textGear.SetController(ctrl)

	child.SetProp(gears.ObjectPropIDText, "foo")
	textGear.UpdateState()

	ctrl.SetSelectedIndex(1)
	child.SetProp(gears.ObjectPropIDText, "bar")
	textGear.UpdateState()

	child.SetProp(gears.ObjectPropIDText, "baz")
	ctrl.SetSelectedIndex(0)
	if got, _ := child.GetProp(gears.ObjectPropIDText).(string); got != "foo" {
		t.Fatalf("expected page0 text 'foo', got %q", got)
	}
	ctrl.SetSelectedIndex(1)
	if got, _ := child.GetProp(gears.ObjectPropIDText).(string); got != "bar" {
		t.Fatalf("expected page1 text 'bar', got %q", got)
	}
}

func TestGearIconSwitch(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()
	parent.AddChild(child)

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"P0", "P1"})
	parent.AddController(ctrl)

	gear := child.GetGear(gears.IndexIcon)
	iconGear, ok := gear.(*gears.GearIcon)
	if !ok {
		t.Fatalf("expected GearIcon, got %T", gear)
	}
	iconGear.SetController(ctrl)

	child.SetProp(gears.ObjectPropIDIcon, "icon0")
	iconGear.UpdateState()
	ctrl.SetSelectedIndex(1)
	child.SetProp(gears.ObjectPropIDIcon, "icon1")
	iconGear.UpdateState()
	child.SetProp(gears.ObjectPropIDIcon, "icon-x")
	ctrl.SetSelectedIndex(0)
	if got, _ := child.GetProp(gears.ObjectPropIDIcon).(string); got != "icon0" {
		t.Fatalf("expected icon for page0 to be 'icon0', got %q", got)
	}
	ctrl.SetSelectedIndex(1)
	if got, _ := child.GetProp(gears.ObjectPropIDIcon).(string); got != "icon1" {
		t.Fatalf("expected icon for page1 to be 'icon1', got %q", got)
	}
}

func TestGearFontSizeSwitch(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()
	parent.AddChild(child)

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"P0", "P1"})
	parent.AddController(ctrl)

	gear := child.GetGear(gears.IndexFontSize)
	fontGear, ok := gear.(*gears.GearFontSize)
	if !ok {
		t.Fatalf("expected GearFontSize, got %T", gear)
	}
	fontGear.SetController(ctrl)

	child.SetProp(gears.ObjectPropIDFontSize, 12)
	fontGear.UpdateState()
	ctrl.SetSelectedIndex(1)
	child.SetProp(gears.ObjectPropIDFontSize, 18)
	fontGear.UpdateState()
	child.SetProp(gears.ObjectPropIDFontSize, 22)
	ctrl.SetSelectedIndex(0)
	if got, _ := child.GetProp(gears.ObjectPropIDFontSize).(int); got != 12 {
		t.Fatalf("expected font size 12 for page0, got %d", got)
	}
	ctrl.SetSelectedIndex(1)
	if got, _ := child.GetProp(gears.ObjectPropIDFontSize).(int); got != 18 {
		t.Fatalf("expected font size 18 for page1, got %d", got)
	}
}

func TestGearAnimationTimeScaleDelta(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()
	parent.AddChild(child)

	ctrl := NewController("ctrl")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"P0", "P1"})
	parent.AddController(ctrl)

	gear := child.GetGear(gears.IndexAnimation)
	anim, ok := gear.(*gears.GearAnimation)
	if !ok {
		t.Fatalf("expected GearAnimation, got %T", gear)
	}
	anim.SetController(ctrl)

	child.SetProp(gears.ObjectPropIDPlaying, true)
	child.SetProp(gears.ObjectPropIDFrame, 2)
	child.SetProp(gears.ObjectPropIDTimeScale, 1.0)
	child.SetProp(gears.ObjectPropIDDeltaTime, 0.0)
	anim.UpdateState()

	ctrl.SetSelectedIndex(1)
	child.SetProp(gears.ObjectPropIDPlaying, false)
	child.SetProp(gears.ObjectPropIDFrame, 4)
	child.SetProp(gears.ObjectPropIDTimeScale, 0.25)
	child.SetProp(gears.ObjectPropIDDeltaTime, 33.0)
	anim.UpdateState()

	ctrl.SetSelectedIndex(0)
	if ts, _ := child.GetProp(gears.ObjectPropIDTimeScale).(float64); math.Abs(ts-1) > 1e-6 {
		t.Fatalf("expected page0 time scale 1, got %.3f", ts)
	}
	if dt, _ := child.GetProp(gears.ObjectPropIDDeltaTime).(float64); math.Abs(dt) > 1e-6 {
		t.Fatalf("expected page0 deltaTime reset, got %.3f", dt)
	}

	ctrl.SetSelectedIndex(1)
	if ts, _ := child.GetProp(gears.ObjectPropIDTimeScale).(float64); math.Abs(ts-0.25) > 1e-6 {
		t.Fatalf("expected page1 time scale 0.25, got %.3f", ts)
	}
	if dt, _ := child.GetProp(gears.ObjectPropIDDeltaTime).(float64); math.Abs(dt-33.0) > 1e-6 {
		t.Fatalf("expected page1 deltaTime 33, got %.3f", dt)
	}
}
