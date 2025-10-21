package fgui

import (
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// Public aliases to mirror the TypeScript API surface.
type (
	Stage          = laya.Stage
	Scheduler      = laya.Scheduler
	MouseState     = laya.MouseState
	PointerEvent   = laya.PointerEvent
	EventType      = laya.EventType
	GRoot          = core.GRoot
	GComponent     = core.GComponent
	GObject        = core.GObject
	PopupDirection = core.PopupDirection
)

const (
	// PopupDirectionAuto positions the popup below the target when possible.
	PopupDirectionAuto = core.PopupDirectionAuto
	// PopupDirectionUp positions the popup above the target.
	PopupDirectionUp = core.PopupDirectionUp
	// PopupDirectionDown positions the popup below the target.
	PopupDirectionDown = core.PopupDirectionDown
)

// NewStage constructs a compat stage suitable for attaching to the root.
func NewStage(width, height int) *Stage {
	return laya.NewStage(width, height)
}

// NewGObject creates a bare UI object backed by a compat sprite.
func NewGObject() *core.GObject {
	return core.NewGObject()
}

// NewGComponent constructs an empty component container.
func NewGComponent() *core.GComponent {
	return core.NewGComponent()
}

// Root returns the singleton GRoot instance (alias of core.Root()).
func Root() *core.GRoot {
	return core.Root()
}

// Instance is an alias to Root for parity with the TypeScript API.
func Instance() *core.GRoot {
	return core.Inst()
}

// AttachStage binds the singleton root to the provided stage.
func AttachStage(stage *Stage) *core.GRoot {
	root := core.Root()
	root.AttachStage(stage)
	return root
}

// CurrentStage returns the compat stage currently attached to the root, if any.
func CurrentStage() *Stage {
	return core.Root().Stage()
}

// Advance ticks the singleton root and underlying stage scheduler.
func Advance(delta time.Duration, mouse MouseState) {
	core.Root().Advance(delta, mouse)
}

// CurrentScheduler exposes the stage scheduler for timer integrations.
func CurrentScheduler() *Scheduler {
	return core.Root().Scheduler()
}

// ShowPopup displays the popup using the singleton root.
func ShowPopup(popup, target *core.GObject, dir PopupDirection) {
	core.Root().ShowPopup(popup, target, dir)
}

// HidePopup hides the specified popup via the singleton root.
func HidePopup(popup *core.GObject) {
	core.Root().HidePopup(popup)
}

// HideAllPopups hides all active popups on the singleton root.
func HideAllPopups() {
	core.Root().HideAllPopups()
}

// TogglePopup toggles the popup on the singleton root.
func TogglePopup(popup, target *core.GObject, dir PopupDirection) {
	core.Root().TogglePopup(popup, target, dir)
}

// HasAnyPopup reports whether the singleton root currently has visible popups.
func HasAnyPopup() bool {
	return core.Root().HasAnyPopup()
}

// Resize updates both root and stage dimensions for the singleton root.
func Resize(width, height int) {
	core.Root().Resize(width, height)
}

// ContentScale reports the current content scale level.
func ContentScale() int {
	return core.ContentScaleLevel
}
