package core

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
)

func TestGRootSingleton(t *testing.T) {
	a := Root()
	b := Inst()
	if a != b {
		t.Fatalf("expected singleton instance, got %p and %p", a, b)
	}
}

func TestGRootAttachStage(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)

	root.AttachStage(stage)
	if root.Stage() != stage {
		t.Fatalf("stage not attached")
	}
	if stage.Root().Children()[0] != root.DisplayObject() {
		t.Fatalf("root display not added to stage")
	}

	w := root.Width()
	h := root.Height()
	if w != 800 || h != 600 {
		t.Fatalf("root size mismatch got (%v,%v)", w, h)
	}
}

func TestGRootPopupLifecycle(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)
	root.AttachStage(stage)

	popup := NewGObject()
	root.ShowPopup(popup, nil, PopupDirectionAuto)
	if !root.HasAnyPopup() {
		t.Fatalf("expected popup stack")
	}
	if popup.parent != root.GComponent {
		t.Fatalf("popup not attached to root")
	}

	root.HidePopup(popup)
	if root.HasAnyPopup() {
		t.Fatalf("expected popup removed")
	}
	if popup.parent != nil {
		t.Fatalf("popup should be detached")
	}
}

func TestGRootCheckPopups(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)
	root.AttachStage(stage)

	p1 := NewGObject()
	p2 := NewGObject()
	root.ShowPopup(p1, nil, PopupDirectionAuto)
	root.ShowPopup(p2, p1, PopupDirectionAuto)

	root.CheckPopups(p1.DisplayObject())
	if !root.HasAnyPopup() {
		t.Fatalf("primary popup should remain")
	}
	if root.indexOfPopup(p1) != 0 {
		t.Fatalf("expected primary popup to stay in stack")
	}
	if root.indexOfPopup(p2) != -1 {
		t.Fatalf("secondary popup should close when clicking primary")
	}

	root.CheckPopups(nil)
	if root.HasAnyPopup() {
		t.Fatalf("expected all popups closed")
	}
}

func TestGRootStageEventsClosePopups(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)
	root.AttachStage(stage)

	popup := NewGObject()
	root.ShowPopup(popup, nil, PopupDirectionAuto)

	stage.Root().Dispatcher().Emit(laya.EventStageMouseDown, laya.PointerEvent{})
	if root.HasAnyPopup() {
		t.Fatalf("popups expected to close on stage mouse down")
	}
	if len(root.justClosed) != 1 {
		t.Fatalf("justClosed should contain popup")
	}

	stage.Root().Dispatcher().Emit(laya.EventStageMouseUp, laya.PointerEvent{})
	if len(root.justClosed) != 0 {
		t.Fatalf("stage mouse up should clear justClosed")
	}
}

func TestGRootTogglePopupRespectsJustClosed(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)
	root.AttachStage(stage)

	popup := NewGObject()
	root.ShowPopup(popup, nil, PopupDirectionAuto)

	stage.Root().Dispatcher().Emit(laya.EventStageMouseDown, laya.PointerEvent{})
	if root.HasAnyPopup() {
		t.Fatalf("expected popup to close on stage down")
	}

	root.TogglePopup(popup, nil, PopupDirectionAuto)
	if root.HasAnyPopup() {
		t.Fatalf("toggle should be ignored while popup flagged as just closed")
	}

	stage.Root().Dispatcher().Emit(laya.EventStageMouseUp, laya.PointerEvent{})
	root.TogglePopup(popup, nil, PopupDirectionAuto)
	if !root.HasAnyPopup() {
		t.Fatalf("toggle should reopen popup after stage up")
	}
}

func TestGRootAdvanceRoutesStage(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(400, 300)
	root.AttachStage(stage)

	child := NewGObject()
	child.SetSize(50, 50)
	stage.AddChild(child.DisplayObject())

	var down bool
	child.DisplayObject().Dispatcher().On(laya.EventMouseDown, func(evt *laya.Event) {
		down = true
	})

	root.Advance(time.Millisecond*16, laya.MouseState{X: 10, Y: 10, Primary: true})
	if !down {
		t.Fatalf("expected stage update to deliver mouse events")
	}
}

func TestGRootContentScaleLevel(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)
	root.AttachStage(stage)

	stage.Root().SetScale(2.6, 2.6)
	root.updateContentScaleLevel()
	if ContentScaleLevel != 2 {
		t.Fatalf("expected content scale level 2, got %d", ContentScaleLevel)
	}

	stage.Root().SetScale(1, 1)
	root.updateContentScaleLevel()
	if ContentScaleLevel != 0 {
		t.Fatalf("expected content scale level reset, got %d", ContentScaleLevel)
	}
}

func TestGRootPositionPopupClampsWithinStage(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)
	root.AttachStage(stage)

	target := NewGObject()
	target.SetSize(40, 40)
	target.SetPosition(760, 120)
	root.AddChild(target)

	popup := NewGObject()
	popup.SetSize(120, 80)
	root.ShowPopup(popup, target, PopupDirectionAuto)

	wantX := root.Width() - popup.Width()
	if popup.X() != wantX {
		t.Fatalf("expected popup X %v, got %v", wantX, popup.X())
	}
	if popup.Y() != target.Y()+target.Height() {
		t.Fatalf("expected popup below target, got %v", popup.Y())
	}
}

func TestGRootPositionPopupAutoFlipsUp(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)
	root.AttachStage(stage)

	target := NewGObject()
	target.SetSize(80, 40)
	target.SetPosition(200, 540)
	root.AddChild(target)

	popup := NewGObject()
	popup.SetSize(150, 120)
	root.ShowPopup(popup, target, PopupDirectionAuto)

	if popup.Y() >= target.Y() {
		t.Fatalf("popup should be positioned above target when lacking space below, got %v", popup.Y())
	}
	if popup.X() < 0 || popup.X()+popup.Width() > root.Width() {
		t.Fatalf("popup X should remain within root bounds, got %v", popup.X())
	}
}

func TestGRootPositionPopupForcedUpClamp(t *testing.T) {
	root := NewGRoot()
	stage := laya.NewStage(800, 600)
	root.AttachStage(stage)

	target := NewGObject()
	target.SetSize(60, 30)
	target.SetPosition(100, 20)
	root.AddChild(target)

	popup := NewGObject()
	popup.SetSize(120, 150)
	root.ShowPopup(popup, target, PopupDirectionUp)

	if popup.Y() != 0 {
		t.Fatalf("popup should clamp to top edge, got %v", popup.Y())
	}
	if popup.X() <= target.X() {
		t.Fatalf("popup X should shift right when clamped, got %v", popup.X())
	}
	if popup.X()+popup.Width() > root.Width() {
		t.Fatalf("popup should remain within stage width, got %v", popup.X())
	}
}
