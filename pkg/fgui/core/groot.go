package core

import (
	"math"
	"sync"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/tween"
)

// PopupDirection describes how a popup should be positioned relative to its target.
type PopupDirection int

const (
	// PopupDirectionAuto positions the popup below the target when possible.
	PopupDirectionAuto PopupDirection = iota
	// PopupDirectionUp positions the popup above the target.
	PopupDirectionUp
	// PopupDirectionDown positions the popup below the target.
	PopupDirectionDown
)

var (
	rootOnce sync.Once
	rootInst *GRoot
)

// ContentScaleLevel mirrors the TypeScript global that tracks stage scaling tiers.
var ContentScaleLevel int

// GRoot represents the top-level UI root responsible for stage wiring and popup management.
type GRoot struct {
	*GComponent

	stage *laya.Stage

	popupStack []*GObject
	justClosed []*GObject
	checking   bool

	stageMouseDown laya.Listener
	stageMouseUp   laya.Listener
}

// NewGRoot constructs a detached root. Use AttachStage to bind it to a stage.
func NewGRoot() *GRoot {
	comp := NewGComponent()
	if comp.DisplayObject() != nil {
		comp.DisplayObject().SetName("GRoot")
	}
	return &GRoot{
		GComponent: comp,
		popupStack: make([]*GObject, 0),
		justClosed: make([]*GObject, 0),
	}
}

// Inst returns the singleton root, mirroring the TypeScript API.
func Inst() *GRoot {
	rootOnce.Do(func() {
		rootInst = NewGRoot()
	})
	return rootInst
}

// Root is an alias of Inst for Go-idiomatic naming.
func Root() *GRoot {
	return Inst()
}

// Stage exposes the currently bound compat stage.
func (r *GRoot) Stage() *laya.Stage {
	return r.stage
}

// AttachStage binds the root to the provided stage and ensures the display tree is registered.
func (r *GRoot) AttachStage(stage *laya.Stage) {
	if r.stage == stage {
		if stage != nil {
			r.syncStageSize()
		}
		return
	}
	r.detachStage()
	r.stage = stage
	if stage == nil {
		return
	}
	stage.AddChild(r.DisplayObject())
	r.registerStageListeners()
	r.syncStageSize()
}

// Advance ticks the underlying stage scheduler and routes pointer events.
func (r *GRoot) Advance(delta time.Duration, mouse laya.MouseState) {
	r.AdvanceInput(delta, laya.InputState{Mouse: mouse})
}

// AdvanceInput ticks the stage with a full input payload.
func (r *GRoot) AdvanceInput(delta time.Duration, input laya.InputState) {
	if r.stage == nil {
		return
	}
	r.stage.UpdateInput(delta, input)
	tickAll(delta)
	tween.Advance(delta)

	// 更新所有 GTextInput 的光标闪烁状态
	r.updateTextInputCursors(delta.Seconds())
}

// updateTextInputCursors 递归遍历组件树,更新所有 GTextInput 的光标
func (r *GRoot) updateTextInputCursors(deltaTime float64) {
	r.visitTextInputs(r.GComponent, deltaTime)
}

// visitTextInputs 递归访问组件树中的所有 GTextInput
func (r *GRoot) visitTextInputs(comp *GComponent, deltaTime float64) {
	if comp == nil {
		return
	}
	for _, child := range comp.Children() {
		// 检查是否是 GTextInput
		if data := child.Data(); data != nil {
			// 使用类型断言检查是否实现了 UpdateCursor 方法
			if updater, ok := data.(interface{ UpdateCursor(float64) }); ok {
				updater.UpdateCursor(deltaTime)
			}
		}
		// 如果子对象也是组件,递归访问
		if childData := child.Data(); childData != nil {
			if childComp, ok := childData.(*GComponent); ok {
				r.visitTextInputs(childComp, deltaTime)
			}
		}
	}
}

// Scheduler returns the stage scheduler, if a stage has been attached.
func (r *GRoot) Scheduler() *laya.Scheduler {
	if r.stage == nil {
		return nil
	}
	return r.stage.Scheduler()
}

// Resize updates both the root and stage dimensions.
func (r *GRoot) Resize(width, height int) {
	if r.stage != nil {
		r.stage.SetSize(width, height)
	}
	r.SetSize(float64(width), float64(height))
	r.updateContentScaleLevel()
}

// ShowPopup displays the popup and tracks it on the stack. Target may be nil.
func (r *GRoot) ShowPopup(popup, target *GObject, dir PopupDirection) {
	if popup == nil {
		return
	}
	if idx := r.indexOfPopup(popup); idx != -1 {
		for i := len(r.popupStack) - 1; i >= idx; i-- {
			existing := r.popupStack[i]
			r.popupStack = r.popupStack[:len(r.popupStack)-1]
			r.closePopup(existing)
		}
	}
	r.popupStack = append(r.popupStack, popup)
	if popup.parent != r.GComponent {
		r.AddChild(popup)
	} else {
		// ensure popup is last
		r.RemoveChild(popup)
		r.AddChild(popup)
	}
	r.positionPopup(popup, target, dir)
}

// HidePopup removes the popup from the stack and detaches it from the root.
func (r *GRoot) HidePopup(popup *GObject) {
	if popup == nil {
		return
	}
	if idx := r.indexOfPopup(popup); idx != -1 {
		r.popupStack = append(r.popupStack[:idx], r.popupStack[idx+1:]...)
	}
	r.closePopup(popup)
}

// HideAllPopups closes every active popup.
func (r *GRoot) HideAllPopups() {
	for len(r.popupStack) > 0 {
		popup := r.popupStack[len(r.popupStack)-1]
		r.popupStack = r.popupStack[:len(r.popupStack)-1]
		r.closePopup(popup)
	}
}

// TogglePopup opens the popup if closed, or hides it otherwise.
func (r *GRoot) TogglePopup(popup, target *GObject, dir PopupDirection) {
	if popup == nil {
		return
	}
	for _, closed := range r.justClosed {
		if closed == popup {
			return
		}
	}
	if r.indexOfPopup(popup) != -1 || popup.parent == r.GComponent {
		r.HidePopup(popup)
		return
	}
	r.ShowPopup(popup, target, dir)
}

// HasAnyPopup reports whether at least one popup is currently open.
func (r *GRoot) HasAnyPopup() bool {
	return len(r.popupStack) > 0
}

// CheckPopups closes popups when the pointer target is outside the active stack.
func (r *GRoot) CheckPopups(target *laya.Sprite) {
	if r.checking {
		return
	}
	r.checking = true
	defer func() {
		r.checking = false
	}()
	r.justClosed = r.justClosed[:0]
	if len(r.popupStack) == 0 {
		return
	}

	for current := target; current != nil; current = current.Parent() {
		if owner := ownerAsGObject(current); owner != nil {
			if idx := r.indexOfPopup(owner); idx != -1 {
				for i := len(r.popupStack) - 1; i > idx; i-- {
					popup := r.popupStack[len(r.popupStack)-1]
					r.popupStack = r.popupStack[:len(r.popupStack)-1]
					r.closePopup(popup)
					r.justClosed = append(r.justClosed, popup)
				}
				return
			}
		}
	}

	for len(r.popupStack) > 0 {
		popup := r.popupStack[len(r.popupStack)-1]
		r.popupStack = r.popupStack[:len(r.popupStack)-1]
		r.closePopup(popup)
		r.justClosed = append(r.justClosed, popup)
	}
}

func (r *GRoot) positionPopup(popup, target *GObject, dir PopupDirection) {
	if popup == nil {
		return
	}
	rootSprite := r.DisplayObject()
	if rootSprite == nil {
		return
	}

	var global laya.Point
	var sizeW, sizeH float64
	if target != nil && target.DisplayObject() != nil {
		global = target.DisplayObject().LocalToGlobal(laya.Point{})
		sizeW, sizeH = approximateObjectSize(target)
	} else if r.stage != nil {
		mouse := r.stage.Mouse()
		global = laya.Point{X: mouse.X, Y: mouse.Y}
	}

	local := rootSprite.GlobalToLocal(global)
	popupW, popupH := approximateObjectSize(popup)
	rootW, rootH := r.rootDimensions()

	xx := local.X
	if popupW > 0 && rootW > 0 && xx+popupW > rootW {
		xx = xx + sizeW - popupW
	}
	if xx < 0 {
		xx = 0
	}

	yy := local.Y + sizeH
	switch dir {
	case PopupDirectionUp:
		yy = local.Y - popupH - 1
	case PopupDirectionDown:
		yy = local.Y + sizeH
	default:
		if popupH > 0 && rootH > 0 && local.Y+sizeH+popupH > rootH {
			yy = local.Y - popupH - 1
		} else {
			yy = local.Y + sizeH
		}
	}

	if yy < 0 {
		yy = 0
		if target != nil && sizeW > 0 {
			xx += sizeW / 2
		}
	}

	if popupW > 0 && rootW > 0 {
		if xx+popupW > rootW {
			xx = rootW - popupW
		}
		if xx < 0 {
			xx = 0
		}
	} else if xx < 0 {
		xx = 0
	}

	if popupH > 0 && rootH > 0 {
		if yy+popupH > rootH {
			yy = rootH - popupH
		}
		if yy < 0 {
			yy = 0
		}
	} else if yy < 0 {
		yy = 0
	}

	popup.SetPosition(xx, yy)
}

func (r *GRoot) closePopup(popup *GObject) {
	if popup == nil {
		return
	}
	if popup.parent != nil {
		popup.parent.RemoveChild(popup)
	}
}

func (r *GRoot) indexOfPopup(popup *GObject) int {
	for i, entry := range r.popupStack {
		if entry == popup {
			return i
		}
	}
	return -1
}

func (r *GRoot) registerStageListeners() {
	if r.stage == nil {
		return
	}
	root := r.stage.Root()
	r.stageMouseDown = func(evt *laya.Event) {
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			if pe.Hit != nil {
				r.CheckPopups(pe.Hit)
			} else {
				r.CheckPopups(pe.Target)
			}
		} else {
			r.CheckPopups(nil)
		}
	}
	r.stageMouseUp = func(evt *laya.Event) {
		r.justClosed = r.justClosed[:0]
	}
	root.Dispatcher().On(laya.EventStageMouseDown, r.stageMouseDown)
	root.Dispatcher().On(laya.EventStageMouseUp, r.stageMouseUp)
}

func (r *GRoot) detachStage() {
	if r.stage == nil {
		return
	}
	root := r.stage.Root()
	if r.stageMouseDown != nil {
		root.Dispatcher().Off(laya.EventStageMouseDown, r.stageMouseDown)
	}
	if r.stageMouseUp != nil {
		root.Dispatcher().Off(laya.EventStageMouseUp, r.stageMouseUp)
	}
	r.stageMouseDown = nil
	r.stageMouseUp = nil
	r.stage.RemoveChild(r.DisplayObject())
	r.stage = nil
}

func (r *GRoot) syncStageSize() {
	if r.stage == nil {
		return
	}
	w, h := r.stage.Size()
	r.SetSize(float64(w), float64(h))
	r.updateContentScaleLevel()
}

func (r *GRoot) updateContentScaleLevel() {
	if r.stage == nil {
		ContentScaleLevel = 0
		return
	}
	sx, sy := r.stage.Root().Scale()
	scale := math.Max(math.Abs(sx), math.Abs(sy))
	switch {
	case scale >= 3.5:
		ContentScaleLevel = 3
	case scale >= 2.5:
		ContentScaleLevel = 2
	case scale >= 1.5:
		ContentScaleLevel = 1
	default:
		ContentScaleLevel = 0
	}
}

func (r *GRoot) rootDimensions() (float64, float64) {
	width := r.Width()
	height := r.Height()
	if (width <= 0 || height <= 0) && r.stage != nil {
		if w, h := r.stage.Size(); w > 0 || h > 0 {
			if width <= 0 {
				width = float64(w)
			}
			if height <= 0 {
				height = float64(h)
			}
		}
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return width, height
}

func approximateObjectSize(obj *GObject) (float64, float64) {
	if obj == nil {
		return 0, 0
	}
	width := obj.Width()
	height := obj.Height()
	if (width <= 0 || height <= 0) && obj.DisplayObject() != nil {
		w, h := obj.DisplayObject().Size()
		if width <= 0 {
			width = w
		}
		if height <= 0 {
			height = h
		}
	}
	if width <= 0 || height <= 0 {
		if nested, ok := obj.Data().(*GComponent); ok && nested != nil {
			if width <= 0 {
				width = nested.Width()
			}
			if height <= 0 {
				height = nested.Height()
			}
		}
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return width, height
}

func ownerAsGObject(sprite *laya.Sprite) *GObject {
	if sprite == nil {
		return nil
	}
	if owner, ok := sprite.Owner().(*GObject); ok {
		return owner
	}
	return nil
}

// PlayOneShotSound plays a sound effect once without tracking.
// This uses the global buttonSoundPlayer if set, similar to transition sounds.
// 参见 TypeScript 版本: GRoot.inst.playOneShotSound(pi.file)
func (r *GRoot) PlayOneShotSound(url string, volume float64) {
	if url == "" {
		return
	}
	if buttonSoundPlayer != nil {
		buttonSoundPlayer(url, volume)
	}
}
