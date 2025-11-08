package scenes

import (
	"context"
	"fmt"
	"math"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/tween"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// JoystickEvent identifiers emitted by JoystickModule.
const (
	JoystickEventMoving laya.EventType = "JoystickMoving"
	JoystickEventUp     laya.EventType = "JoystickUp"
)

// JoystickDemo mirrors the interactive joystick sample.
type JoystickDemo struct {
	component   *core.GComponent
	module      *JoystickModule
	textField   *widgets.GTextField
	moveHandler laya.Listener
	upHandler   laya.Listener
}

// NewJoystickDemo constructs the joystick demo scene.
func NewJoystickDemo() Scene {
	return &JoystickDemo{}
}

func (d *JoystickDemo) Name() string {
	return "JoystickDemo"
}

func (d *JoystickDemo) Load(ctx context.Context, mgr *Manager) (*core.GComponent, error) {
	env := mgr.Environment()
	pkg, err := env.Package(ctx, "Joystick")
	if err != nil {
		return nil, err
	}
	item := chooseComponent(pkg, "Main")
	if item == nil {
		return nil, newMissingComponentError("Joystick", "Main")
	}
	component, err := env.Factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil, err
	}
	d.component = component

	stage := core.Root().Stage()
	if stage == nil {
		return nil, fmt.Errorf("joystick demo: stage not attached")
	}

	module, err := NewJoystickModule(component, stage)
	if err != nil {
		return nil, err
	}
	d.module = module

	if child := component.ChildByName("n9"); child != nil {
		if tf, ok := child.Data().(*widgets.GTextField); ok {
			d.textField = tf
			d.textField.SetText("")
		}
	}

	d.moveHandler = func(evt *laya.Event) {
		if d.textField == nil {
			return
		}
		if degree, ok := evt.Data.(float64); ok {
			d.textField.SetText(fmt.Sprintf("%.0f", degree))
		}
	}
	d.upHandler = func(laya.Event) {
		if d.textField != nil {
			d.textField.SetText("")
		}
	}

	module.On(JoystickEventMoving, d.moveHandler)
	module.On(JoystickEventUp, d.upHandler)

	return component, nil
}

func (d *JoystickDemo) Dispose() {
	if d.module != nil {
		if d.moveHandler != nil {
			d.module.Off(JoystickEventMoving, d.moveHandler)
		}
		if d.upHandler != nil {
			d.module.Off(JoystickEventUp, d.upHandler)
		}
		d.module.Dispose()
		d.module = nil
	}
	d.component = nil
	d.textField = nil
	d.moveHandler = nil
	d.upHandler = nil
}

// JoystickModule replicates the behaviour of the FairyGUI joystick helper.
type JoystickModule struct {
	dispatcher *laya.BasicEventDispatcher

	view      *core.GComponent
	stage     *laya.Stage
	button    *widgets.GButton
	touchArea *core.GObject
	thumb     *core.GObject
	center    *core.GObject

	initX float64
	initY float64

	startPos laya.Point
	lastPos  laya.Point

	touchID int
	radius  float64

	tweener *tween.GTweener

	touchListener laya.Listener
	moveListener  laya.Listener
	upListener    laya.Listener
}

// NewJoystickModule wires joystick interactions for the given component.
func NewJoystickModule(view *core.GComponent, stage *laya.Stage) (*JoystickModule, error) {
	if view == nil {
		return nil, fmt.Errorf("joystick module: nil component")
	}
	if stage == nil {
		return nil, fmt.Errorf("joystick module: nil stage")
	}
	button := childButton(view, "joystick")
	if button == nil {
		return nil, fmt.Errorf("joystick module: button 'joystick' not found")
	}
	touchArea := view.ChildByName("joystick_touch")
	if touchArea == nil {
		return nil, fmt.Errorf("joystick module: touch area 'joystick_touch' not found")
	}
	center := view.ChildByName("joystick_center")
	if center == nil {
		return nil, fmt.Errorf("joystick module: center 'joystick_center' not found")
	}
	thumb := buttonChild(button, "thumb")
	if thumb == nil {
		return nil, fmt.Errorf("joystick module: thumb not found")
	}

	button.SetChangeStateOnClick(false)

	module := &JoystickModule{
		dispatcher: laya.NewEventDispatcher(),
		view:       view,
		stage:      stage,
		button:     button,
		touchArea:  touchArea,
		thumb:      thumb,
		center:     center,
		touchID:    -1,
		radius:     150,
	}

	module.initX = center.X() + center.Width()/2
	module.initY = center.Y() + center.Height()/2

	module.touchListener = func(evt *laya.Event) {
		module.onTouchDown(evt)
	}
	touchArea.On(laya.EventMouseDown, module.touchListener)

	return module, nil
}

// Dispose releases listeners and running tweens.
func (m *JoystickModule) Dispose() {
	if m.touchArea != nil && m.touchListener != nil {
		m.touchArea.Off(laya.EventMouseDown, m.touchListener)
	}
	m.unregisterStageListeners()
	if m.tweener != nil {
		m.tweener.Kill(false)
		m.tweener = nil
	}
	if m.button != nil {
		m.button.SetSelected(false)
	}
	m.touchID = -1
}

// On registers a listener for joystick events.
func (m *JoystickModule) On(evt laya.EventType, fn laya.Listener) {
	if m.dispatcher != nil {
		m.dispatcher.On(evt, fn)
	}
}

// Off removes a previously registered joystick listener.
func (m *JoystickModule) Off(evt laya.EventType, fn laya.Listener) {
	if m.dispatcher != nil {
		m.dispatcher.Off(evt, fn)
	}
}

func (m *JoystickModule) emit(evt laya.EventType, data any) {
	if m.dispatcher != nil {
		m.dispatcher.Emit(evt, data)
	}
}

func (m *JoystickModule) onTouchDown(evt laya.Event) {
	if m.touchID != -1 {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok {
		return
	}
	if pe.TouchID <= 0 {
		pe.TouchID = 1
	}
	m.touchID = pe.TouchID

	if m.tweener != nil {
		m.tweener.Kill(false)
		m.tweener = nil
	}

	local := m.componentSpace(pe.Position)
	bounds := m.touchBounds()
	bx := clamp(local.X, bounds.X, bounds.X+bounds.W)
	by := clamp(local.Y, bounds.Y, bounds.Y+bounds.H)

	m.startPos = laya.Point{X: bx, Y: by}
	m.lastPos = m.startPos

	button := m.button.GComponent.GObject
	center := m.center

	m.button.SetSelected(true)
	button.SetPosition(bx-button.Width()/2, by-button.Height()/2)
	center.SetVisible(true)
	center.SetPosition(bx-center.Width()/2, by-center.Height()/2)

	m.updateThumb(bx, by)
	m.registerStageListeners()
}

func (m *JoystickModule) onStageMove(evt laya.Event) {
	if m.touchID == -1 {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok || pe.TouchID != m.touchID {
		return
	}
	m.handleMove(pe)
}

func (m *JoystickModule) onStageUp(evt laya.Event) {
	if m.touchID == -1 {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok || pe.TouchID != m.touchID {
		return
	}
	m.handleRelease()
}

func (m *JoystickModule) handleMove(pe laya.PointerEvent) {
	local := m.componentSpace(pe.Position)
	moveX := local.X - m.lastPos.X
	moveY := local.Y - m.lastPos.Y
	m.lastPos = local

	button := m.button.GComponent.GObject
	buttonX := button.X() + moveX
	buttonY := button.Y() + moveY

	centerX := buttonX + button.Width()/2
	centerY := buttonY + button.Height()/2

	offsetX := centerX - m.startPos.X
	offsetY := centerY - m.startPos.Y

	rad := math.Atan2(offsetY, offsetX)
	degree := rad * 180 / math.Pi

	maxX := m.radius * math.Cos(rad)
	maxY := m.radius * math.Sin(rad)
	if math.Abs(offsetX) > math.Abs(maxX) {
		offsetX = maxX
	}
	if math.Abs(offsetY) > math.Abs(maxY) {
		offsetY = maxY
	}

	centerX = m.startPos.X + offsetX
	centerY = m.startPos.Y + offsetY

	if centerX < 0 {
		centerX = 0
	}
	rootHeight := core.Root().Height()
	if centerY > rootHeight {
		centerY = rootHeight
	}

	button.SetPosition(centerX-button.Width()/2, centerY-button.Height()/2)
	m.updateThumb(centerX, centerY)
	m.emit(JoystickEventMoving, degree)
}

func (m *JoystickModule) handleRelease() {
	if m.touchID == -1 {
		return
	}
	m.touchID = -1
	m.unregisterStageListeners()

	if m.tweener != nil {
		m.tweener.Kill(false)
		m.tweener = nil
	}

	button := m.button.GComponent.GObject
	center := m.center

	center.SetVisible(false)

	startX := button.X()
	startY := button.Y()
	endX := m.initX - button.Width()/2
	endY := m.initY - button.Height()/2

	m.tweener = tween.To2(startX, startY, endX, endY, 0.3).
		SetEase(tween.EaseTypeCircOut).
		OnUpdate(func(tw *tween.GTweener) {
			val := tw.Value()
			button.SetPosition(val.X, val.Y)
		}).
		OnComplete(func(*tween.GTweener) {
			button.SetPosition(endX, endY)
			m.thumb.SetRotation(0)
			center.SetVisible(true)
			center.SetPosition(m.initX-center.Width()/2, m.initY-center.Height()/2)
			m.button.SetSelected(false)
			m.tweener = nil
		})

	m.emit(JoystickEventUp, nil)
}

func (m *JoystickModule) registerStageListeners() {
	root := m.stage.Root()
	if root == nil {
		return
	}
	dispatcher := root.Dispatcher()
	if dispatcher == nil {
		return
	}
	if m.moveListener == nil {
		m.moveListener = func(evt *laya.Event) {
			m.onStageMove(evt)
		}
	}
	if m.upListener == nil {
		m.upListener = func(evt *laya.Event) {
			m.onStageUp(evt)
		}
	}
	dispatcher.On(laya.EventMouseMove, m.moveListener)
	dispatcher.On(laya.EventStageMouseUp, m.upListener)
}

func (m *JoystickModule) unregisterStageListeners() {
	root := m.stage.Root()
	if root == nil {
		return
	}
	dispatcher := root.Dispatcher()
	if dispatcher == nil {
		return
	}
	if m.moveListener != nil {
		dispatcher.Off(laya.EventMouseMove, m.moveListener)
	}
	if m.upListener != nil {
		dispatcher.Off(laya.EventStageMouseUp, m.upListener)
	}
}

func (m *JoystickModule) componentSpace(global laya.Point) laya.Point {
	if m.view == nil || m.view.DisplayObject() == nil {
		return global
	}
	return m.view.DisplayObject().GlobalToLocal(global)
}

func (m *JoystickModule) touchBounds() laya.Rect {
	return laya.Rect{
		X: m.touchArea.X(),
		Y: m.touchArea.Y(),
		W: m.touchArea.Width(),
		H: m.touchArea.Height(),
	}
}

func (m *JoystickModule) updateThumb(centerX, centerY float64) {
	deltaX := centerX - m.initX
	deltaY := centerY - m.initY
	degree := math.Atan2(deltaY, deltaX)*180/math.Pi + 90
	m.thumb.SetRotation(degree * math.Pi / 180)
}

func childButton(parent *core.GComponent, name string) *widgets.GButton {
	if parent == nil {
		return nil
	}
	if obj := parent.ChildByName(name); obj != nil {
		if btn, ok := obj.Data().(*widgets.GButton); ok {
			return btn
		}
	}
	return nil
}

func buttonChild(btn *widgets.GButton, name string) *core.GObject {
	if btn == nil {
		return nil
	}
	if tmpl := btn.TemplateComponent(); tmpl != nil {
		if child := tmpl.ChildByName(name); child != nil {
			return child
		}
	}
	if btn.GComponent != nil {
		return btn.GComponent.ChildByName(name)
	}
	return nil
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
