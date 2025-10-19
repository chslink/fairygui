package fgui

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestRootAliasesSingleton(t *testing.T) {
	if Root() != core.Root() {
		t.Fatalf("Root alias did not match core singleton")
	}
	if Instance() != Root() {
		t.Fatalf("Instance alias should match Root")
	}
}

func TestStageAttachmentThroughAPI(t *testing.T) {
	stage := NewStage(320, 240)
	AttachStage(stage)
	if CurrentStage() != stage {
		t.Fatalf("stage accessor mismatch")
	}

	Resize(800, 600)
	if root := Root(); root.Width() != 800 || root.Height() != 600 {
		t.Fatalf("resize should propagate to root, got (%v,%v)", root.Width(), root.Height())
	}
}

func TestPopupShortcuts(t *testing.T) {
	stage := NewStage(400, 300)
	AttachStage(stage)
	HideAllPopups()

	popup := NewGObject()
	ShowPopup(popup, nil, PopupDirectionAuto)
	if !HasAnyPopup() {
		t.Fatalf("expected popup to be tracked")
	}

	HidePopup(popup)
	if HasAnyPopup() {
		t.Fatalf("popup should be hidden via shortcut")
	}
}

func TestAdvanceShortcut(t *testing.T) {
	stage := NewStage(200, 200)
	AttachStage(stage)

	obj := NewGObject()
	obj.SetSize(20, 20)
	stage.AddChild(obj.DisplayObject())

	var clicked bool
	obj.DisplayObject().Dispatcher().Once(laya.EventMouseDown, func(evt laya.Event) {
		clicked = true
	})

	Advance(time.Millisecond*16, MouseState{X: 5, Y: 5, Primary: true})
	if !clicked {
		t.Fatalf("advance shortcut should route stage events")
	}
}

func TestContentScaleAlias(t *testing.T) {
	stage := NewStage(120, 90)
	AttachStage(stage)

	stage.Root().SetScale(1.6, 1.6)
	core.Root().Resize(120, 90)
	if ContentScale() != 1 {
		t.Fatalf("content scale should reflect stage scaling, got %d", ContentScale())
	}

	stage.Root().SetScale(1, 1)
	core.Root().Resize(120, 90)
	if ContentScale() != core.ContentScaleLevel {
		t.Fatalf("content scale wrapper mismatch")
	}
}
