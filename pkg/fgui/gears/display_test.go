package gears

import "testing"

type mockController struct {
	selectedPageID string
	selectedIndex  int
}

func (m *mockController) SelectedPageID() string { return m.selectedPageID }
func (m *mockController) SelectedIndex() int      { return m.selectedIndex }

type mockOwner struct {
	x, y            float64
	width, height   float64
	scaleX, scaleY  float64
	alpha           float64
	rotation        float64
	visible         bool
	touchable       bool
	grayed          bool
	props           map[ObjectPropID]any
	gearLocked      bool
	displayLockCnt  int
	displayLockTok  uint32
}

func newMockOwner() *mockOwner {
	return &mockOwner{
		visible:   true,
		touchable: true,
		alpha:     1.0,
		scaleX:    1,
		scaleY:    1,
		props:     make(map[ObjectPropID]any),
	}
}

func (m *mockOwner) ID() string                                { return "mock" }
func (m *mockOwner) X() float64                                { return m.x }
func (m *mockOwner) Y() float64                                { return m.y }
func (m *mockOwner) SetPosition(x, y float64)                  { m.x, m.y = x, y }
func (m *mockOwner) Width() float64                            { return m.width }
func (m *mockOwner) Height() float64                           { return m.height }
func (m *mockOwner) SetSize(w, h float64)                      { m.width, m.height = w, h }
func (m *mockOwner) Scale() (float64, float64)                 { return m.scaleX, m.scaleY }
func (m *mockOwner) SetScale(sx, sy float64)                   { m.scaleX, m.scaleY = sx, sy }
func (m *mockOwner) ParentSize() (float64, float64)            { return 400, 300 }
func (m *mockOwner) SetGearLocked(l bool)                      { m.gearLocked = l }
func (m *mockOwner) GearLocked() bool                          { return m.gearLocked }
func (m *mockOwner) Visible() bool                             { return m.visible }
func (m *mockOwner) SetVisible(v bool)                         { m.visible = v }
func (m *mockOwner) Alpha() float64                            { return m.alpha }
func (m *mockOwner) SetAlpha(a float64)                        { m.alpha = a }
func (m *mockOwner) Rotation() float64                         { return m.rotation }
func (m *mockOwner) SetRotation(r float64)                     { m.rotation = r }
func (m *mockOwner) Grayed() bool                              { return m.grayed }
func (m *mockOwner) SetGrayed(g bool)                          { m.grayed = g }
func (m *mockOwner) Touchable() bool                           { return m.touchable }
func (m *mockOwner) SetTouchable(t bool)                       { m.touchable = t }
func (m *mockOwner) GetProp(id ObjectPropID) any               { return m.props[id] }
func (m *mockOwner) SetProp(id ObjectPropID, v any)            { m.props[id] = v }
func (m *mockOwner) CheckGearController(idx int, c Controller) bool { return false }
func (m *mockOwner) AddDisplayLock() uint32                    { m.displayLockCnt++; m.displayLockTok++; return m.displayLockTok }
func (m *mockOwner) ReleaseDisplayLock(tok uint32) {
	if tok == m.displayLockTok && m.displayLockCnt > 0 {
		m.displayLockCnt--
	}
}
func (m *mockOwner) hasActiveDisplayLock() bool { return m.displayLockCnt > 0 }

func TestGearDisplayConnected(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "up", selectedIndex: 0}
	gear := NewGearDisplay(owner)
	gear.SetController(ctrl)

	// No pages set → should be connected (visible)
	if !gear.Connected() {
		t.Error("expected connected when no pages set")
	}

	// Set pages to only "down"
	gear.SetPages([]string{"down"})
	// Current page is "up", not in pages → should be disconnected
	if gear.Connected() {
		t.Error("expected disconnected when page 'up' not in pages ['down']")
	}

	// Change page to "down"
	ctrl.selectedPageID = "down"
	ctrl.selectedIndex = 1
	gear.Apply()
	if !gear.Connected() {
		t.Error("expected connected when page 'down' is in pages ['down']")
	}
}

func TestGearDisplayIndexMatch(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "down", selectedIndex: 1}
	gear := NewGearDisplay(owner)
	gear.SetController(ctrl)
	gear.SetPages([]string{"0", "1", "2"})

	// Page index 1 should match stored string "1"
	gear.Apply()
	if !gear.Connected() {
		t.Error("expected connected when index 1 matches '1' in pages")
	}

	ctrl.selectedIndex = 3
	gear.Apply()
	if gear.Connected() {
		t.Error("expected disconnected when index 3 not in pages ['0','1','2']")
	}
}

func TestGearDisplayLock(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "up", selectedIndex: 0}
	gear := NewGearDisplay(owner)
	gear.SetController(ctrl)
	gear.SetPages([]string{"down"})

	// Not connected initially
	if gear.Connected() {
		t.Error("expected disconnected initially")
	}

	// With display lock → should be connected
	owner.AddDisplayLock()
	if !gear.Connected() {
		t.Error("expected connected with active display lock")
	}

	// Release lock → disconnected again
	owner.ReleaseDisplayLock(owner.displayLockTok)
	if gear.Connected() {
		t.Error("expected disconnected after releasing display lock")
	}
}

func TestGearDisplayVisibleCounter(t *testing.T) {
	owner := newMockOwner()
	ctrl := &mockController{selectedPageID: "up", selectedIndex: 0}
	gear := NewGearDisplay(owner)
	gear.SetController(ctrl)

	// No controller pages → visibleCounter = 1 → connected
	if !gear.Connected() {
		t.Error("expected connected when no pages")
	}

	// Add lock → visibleCounter = 2
	tok := gear.AddLock()
	if gear.visibleCounter != 2 {
		t.Errorf("expected visibleCounter=2, got %d", gear.visibleCounter)
	}

	// Release with matching token
	gear.ReleaseLock(tok)
	if gear.visibleCounter != 1 {
		t.Errorf("expected visibleCounter=1, got %d", gear.visibleCounter)
	}
}
