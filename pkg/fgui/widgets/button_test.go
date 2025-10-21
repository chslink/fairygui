package widgets

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestButtonDefaults(t *testing.T) {
	btn := NewButton()
	if btn == nil || btn.GComponent == nil {
		t.Fatalf("expected GButton to wrap GComponent")
	}
	if btn.Mode() != ButtonModeCommon {
		t.Fatalf("unexpected default mode: %v", btn.Mode())
	}
	if !btn.ChangeStateOnClick() {
		t.Fatalf("expected changeStateOnClick to default to true")
	}
	if btn.DownEffectValue() != 0.8 {
		t.Fatalf("unexpected default down effect value: %v", btn.DownEffectValue())
	}
	if btn.SoundVolumeScale() != 1 {
		t.Fatalf("unexpected default sound volume scale: %v", btn.SoundVolumeScale())
	}
	if btn.Selected() {
		t.Fatalf("expected button to start unselected")
	}
}

func TestButtonSelectionAndMode(t *testing.T) {
	btn := NewButton()
	btn.SetMode(ButtonModeCheck)
	if btn.Mode() != ButtonModeCheck {
		t.Fatalf("expected mode to switch to check, got %v", btn.Mode())
	}
	btn.SetSelected(true)
	if !btn.Selected() {
		t.Fatalf("expected selection to stick in check mode")
	}
	btn.SetMode(ButtonModeCommon)
	if btn.Selected() {
		t.Fatalf("expected selection to reset when switching to common mode")
	}
	btn.SetSelected(true)
	if btn.Selected() {
		t.Fatalf("expected selection to remain false in common mode")
	}
}

func TestButtonMetadataAccessors(t *testing.T) {
	btn := NewButton()
	buttonCtrl := core.NewController("button")
	relatedCtrl := core.NewController("related")
	popup := core.NewGObject()

	btn.SetTitle("Title")
	btn.SetSelectedTitle("SelectedTitle")
	btn.SetIcon("ui://pkg/icon")
	btn.SetSelectedIcon("ui://pkg/icon_selected")
	btn.SetSound("sound://click")
	btn.SetSoundVolumeScale(0.5)
	btn.SetChangeStateOnClick(false)
	btn.SetButtonController(buttonCtrl)
	btn.SetRelatedController(relatedCtrl)
	btn.SetRelatedPageID("page")
	btn.SetLinkedPopup(popup)
	btn.SetDownEffect(1)
	btn.SetDownEffectValue(0.6)
	btn.SetDownScaled(true)

	if btn.Title() != "Title" {
		t.Fatalf("unexpected title: %s", btn.Title())
	}
	if btn.SelectedTitle() != "SelectedTitle" {
		t.Fatalf("unexpected selected title: %s", btn.SelectedTitle())
	}
	if btn.Icon() != "ui://pkg/icon" {
		t.Fatalf("unexpected icon: %s", btn.Icon())
	}
	if btn.SelectedIcon() != "ui://pkg/icon_selected" {
		t.Fatalf("unexpected selected icon: %s", btn.SelectedIcon())
	}
	if btn.Sound() != "sound://click" {
		t.Fatalf("unexpected sound: %s", btn.Sound())
	}
	if btn.SoundVolumeScale() != 0.5 {
		t.Fatalf("unexpected sound volume scale: %v", btn.SoundVolumeScale())
	}
	if btn.ChangeStateOnClick() {
		t.Fatalf("expected changeStateOnClick to be false after setter")
	}
	if btn.ButtonController() != buttonCtrl {
		t.Fatalf("expected button controller to persist")
	}
	if btn.RelatedController() != relatedCtrl {
		t.Fatalf("expected related controller to persist")
	}
	if btn.RelatedPageID() != "page" {
		t.Fatalf("unexpected related page id: %s", btn.RelatedPageID())
	}
	if btn.LinkedPopup() != popup {
		t.Fatalf("expected linked popup to persist")
	}
	if btn.DownEffect() != 1 {
		t.Fatalf("unexpected down effect: %d", btn.DownEffect())
	}
	if btn.DownEffectValue() != 0.6 {
		t.Fatalf("unexpected down effect value: %v", btn.DownEffectValue())
	}
	if !btn.DownScaled() {
		t.Fatalf("expected downScaled to be true after setter")
	}
}

func TestButtonTitleObjectSync(t *testing.T) {
	btn := NewButton()
	text := NewText()
	text.GObject.SetData(text)
	btn.SetTitleObject(text.GObject)
	btn.SetTitle("Hello")
	if got := text.Text(); got != "Hello" {
		t.Fatalf("expected title object to receive text, got %q", got)
	}
	btn.SetSelectedTitle("World")
	btn.SetMode(ButtonModeCheck)
	btn.SetSelected(true)
	if got := text.Text(); got != "World" {
		t.Fatalf("expected selected title to propagate, got %q", got)
	}
	btn.SetSelected(false)
	if got := text.Text(); got != "Hello" {
		t.Fatalf("expected title to revert after deselect, got %q", got)
	}
}

func TestButtonIconObjectSync(t *testing.T) {
	btn := NewButton()
	loader := NewLoader()
	loader.GObject.SetData(loader)
	btn.SetIconObject(loader.GObject)
	btn.SetIcon("ui://pkg/icon")
	if got := loader.URL(); got != "ui://pkg/icon" {
		t.Fatalf("expected icon url ui://pkg/icon, got %q", got)
	}
	btn.SetSelectedIcon("ui://pkg/icon_selected")
	btn.SetMode(ButtonModeCheck)
	btn.SetSelected(true)
	if got := loader.URL(); got != "ui://pkg/icon_selected" {
		t.Fatalf("expected selected icon to propagate, got %q", got)
	}
	btn.SetSelected(false)
	if got := loader.URL(); got != "ui://pkg/icon" {
		t.Fatalf("expected icon to revert after deselect, got %q", got)
	}
}

func TestButtonClickTogglesSelection(t *testing.T) {
	env := testutil.NewStageEnv(t, 200, 200)
	stage := env.Stage

	btn := NewButton()
	btn.SetMode(ButtonModeCheck)
	obj := btn.GComponent.GObject
	obj.SetSize(60, 30)
	obj.SetPosition(20, 20)
	stage.AddChild(obj.DisplayObject())

	env.Advance(16*time.Millisecond, laya.MouseState{X: 10, Y: 10, Primary: false})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 30, Y: 30, Primary: true})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 30, Y: 30, Primary: false})

	if !btn.Selected() {
		t.Fatalf("expected button to toggle selection on click")
	}

	btn.SetChangeStateOnClick(false)
	env.Advance(16*time.Millisecond, laya.MouseState{X: 30, Y: 30, Primary: true})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 30, Y: 30, Primary: false})

	if !btn.Selected() {
		t.Fatalf("expected selection to remain when changeStateOnClick disabled")
	}
}
