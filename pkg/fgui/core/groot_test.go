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
	child.DisplayObject().Dispatcher().On(laya.EventMouseDown, func(evt laya.Event) {
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
