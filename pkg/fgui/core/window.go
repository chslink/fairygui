package core

import (
	"github.com/chslink/fairygui/internal/compat/laya"
)

// Window provides a modal popup container with optional close button, drag area,
// and content area. It mirrors the FairyGUI Window class.
type Window struct {
	*GComponent

	contentPane        *GComponent
	closeButton        *GObject
	dragArea           *GObject
	contentArea        *GObject
	frame              *GObject
	modalWaitingPane   *GObject

	modal               bool
	bringToFrontOnClick bool
	isShowing           bool
	isTop               bool
	initDone            bool

	dragging      bool
	dragStartPosX float64
	dragStartPosY float64
	dragMouseX    float64
	dragMouseY    float64

	uiSource     IUISource

	onInitHandler  func()
	onShownHandler func()
	onHideHandler  func()

	dragListenerID   laya.ListenerID
	moveListenerID   laya.ListenerID
	upListenerID     laya.ListenerID
}

// NewWindow constructs a Window. The content pane must be set before showing.
func NewWindow() *Window {
	comp := NewGComponent()
	w := &Window{
		GComponent:          comp,
		bringToFrontOnClick: GetUIConfig().BringWindowToFrontOnClick,
	}
	comp.SetData(w)
	return w
}

func (w *Window) ContentPane() *GComponent {
	return w.contentPane
}

// SetContentPane sets the window's content and automatically discovers
// standard child elements: "frame", "closeButton", "dragArea", "contentArea".
func (w *Window) SetContentPane(pane *GComponent) {
	if w == nil || pane == nil {
		return
	}
	if w.contentPane != nil {
		w.RemoveChild(w.contentPane.GObject)
	}
	w.contentPane = pane
	w.AddChild(pane.GObject)

	w.frame = pane.ChildByName("frame")
	w.closeButton = pane.ChildByName("closeButton")
	w.dragArea = pane.ChildByName("dragArea")
	w.contentArea = pane.ChildByName("contentArea")

	if w.closeButton != nil {
		w.closeButton.OnClick(func() { w.Hide() })
	}

	if w.dragArea != nil {
		w.setupDragHandlers()
	}

	w.initDone = true
	if w.onInitHandler != nil {
		w.onInitHandler()
	}
}

// SetUISource configures lazy loading via an IUISource.
func (w *Window) SetUISource(src IUISource) {
	w.uiSource = src
}

// Show displays the window on the root stage.
func (w *Window) Show() {
	if w == nil {
		return
	}
	root := Root()
	root.ShowWindow(w)
}

// ShowOn displays the window with a specific popup direction.
func (w *Window) ShowOn(root *GRoot) {
	if w == nil || root == nil {
		return
	}
	root.ShowWindow(w)
}

// Hide hides the window (with animation if configured).
func (w *Window) Hide() {
	if w == nil {
		return
	}
	root := Root()
	root.HideWindow(w)
}

// HideImmediately removes the window without animation.
func (w *Window) HideImmediately() {
	if w == nil {
		return
	}
	root := Root()
	root.HideWindowImmediately(w)
}

// ToggleStatus toggles the window visibility.
func (w *Window) ToggleStatus() {
	if w == nil {
		return
	}
	if w.isTop {
		w.Hide()
	} else {
		w.Show()
	}
}

// BringToFront moves the window to the top of the display stack.
func (w *Window) BringToFront() {
	if w == nil {
		return
	}
	root := Root()
	root.BringToFront(w)
}

// CenterOn centers the window on the specified root.
func (w *Window) CenterOn(root *GRoot) {
	if w == nil || root == nil {
		return
	}
	rw, rh := root.Width(), root.Height()
	ww, wh := w.Width(), w.Height()
	w.SetPosition((rw-ww)/2, (rh-wh)/2)
}

// SetModal sets whether the window has a modal background overlay.
func (w *Window) SetModal(modal bool) {
	w.modal = modal
}

// IsShowing reports whether the window is currently displayed.
func (w *Window) IsShowing() bool {
	if w == nil {
		return false
	}
	return w.isShowing
}

// IsTop reports whether the window is the topmost.
func (w *Window) IsTop() bool {
	if w == nil {
		return false
	}
	return w.isTop
}

// ShowModalWait displays a loading indicator on the window.
func (w *Window) ShowModalWait(msg string) {
	if w == nil {
		return
	}
	// Load from UIConfig.globalModalWaiting URL if available.
	_ = msg
}

// CloseModalWait hides the loading indicator.
func (w *Window) CloseModalWait() {
	if w == nil {
		return
	}
}

// SetOnInit registers a callback invoked after the content pane is set.
func (w *Window) SetOnInit(fn func()) {
	w.onInitHandler = fn
}

// SetOnShown registers a callback invoked when the window is shown.
func (w *Window) SetOnShown(fn func()) {
	w.onShownHandler = fn
}

// SetOnHide registers a callback invoked when the window is hidden.
func (w *Window) SetOnHide(fn func()) {
	w.onHideHandler = fn
}

// GObj returns the underlying GObject for compatibility.
func (w *Window) GObj() *GObject {
	return w.GObject
}

// ContentArea returns the content area child, if any.
func (w *Window) ContentArea() *GObject {
	return w.contentArea
}

// doShowAnimation runs the show animation (override for custom behavior).
func (w *Window) doShowAnimation() {}

// doHideAnimation runs the hide animation (override for custom behavior).
func (w *Window) doHideAnimation() {}

func (w *Window) onShown() {
	if w.onShownHandler != nil {
		w.onShownHandler()
	}
}

func (w *Window) onHide() {
	if w.onHideHandler != nil {
		w.onHideHandler()
	}
}

func (w *Window) setupDragHandlers() {
	if w.dragArea == nil {
		return
	}

	stage := Root().Stage()
	if stage == nil {
		return
	}

	w.dragListenerID = w.dragArea.OnWithID(laya.EventMouseDown, func(evt *laya.Event) {
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			w.dragging = true
			w.dragMouseX = pe.Position.X
			w.dragMouseY = pe.Position.Y
			w.dragStartPosX = w.X()
			w.dragStartPosY = w.Y()
		}
	})

	w.moveListenerID = stage.Root().Dispatcher().OnWithID(laya.EventMouseMove, func(evt *laya.Event) {
		if !w.dragging {
			return
		}
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			dx := pe.Position.X - w.dragMouseX
			dy := pe.Position.Y - w.dragMouseY
			w.SetPosition(w.dragStartPosX+dx, w.dragStartPosY+dy)
		}
	})

	w.upListenerID = stage.Root().Dispatcher().OnWithID(laya.EventStageMouseUp, func(evt *laya.Event) {
		w.dragging = false
	})
}
