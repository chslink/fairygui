package laya

import (
	"time"
)

// PointerEvent describes pointer interaction dispatched through sprites.
type PointerEvent struct {
	Position  Point
	Primary   bool
	Target    *Sprite
	Hit       *Sprite
	WheelX    float64
	WheelY    float64
	TouchID   int
	Buttons   MouseButtons
	Modifiers KeyModifiers
	Phase     TouchPhase
}

// MouseState describes mouse/touch input for the stage update.
type MouseState struct {
	X         float64
	Y         float64
	Primary   bool
	Buttons   MouseButtons
	Modifiers KeyModifiers
	WheelX    float64
	WheelY    float64
}

// Stage emulates the behaviour of Laya's global stage.
type Stage struct {
	root      *Sprite
	scheduler *Scheduler
	width     int
	height    int

	mouse        MouseState
	prevMouse    MouseState
	hover        *Sprite
	pressed      *Sprite
	touchSeq     int
	activeID     int
	capture      *Sprite
	focus        *Sprite
	touchTargets map[int]*Sprite
	keysDown     map[KeyCode]bool
	modState     KeyModifiers
}

// NewStage constructs a stage with the provided dimensions.
func NewStage(width, height int) *Stage {
	root := NewSprite()
	root.SetName("stage")
	root.SetSize(float64(width), float64(height))
	return &Stage{
		root:         root,
		scheduler:    NewScheduler(),
		width:        width,
		height:       height,
		activeID:     -1,
		touchTargets: make(map[int]*Sprite),
		keysDown:     make(map[KeyCode]bool),
	}
}

// Root returns the root sprite.
func (s *Stage) Root() *Sprite {
	return s.root
}

// Scheduler returns the frame scheduler.
func (s *Stage) Scheduler() *Scheduler {
	return s.scheduler
}

// SetSize updates the stage dimensions.
func (s *Stage) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.root.SetSize(float64(width), float64(height))
}

// Size returns the current stage dimensions.
func (s *Stage) Size() (int, int) {
	return s.width, s.height
}

// Mouse returns the most recent mouse state processed by the stage.
func (s *Stage) Mouse() MouseState {
	return s.mouse
}

// Focus returns the sprite that currently holds keyboard focus.
func (s *Stage) Focus() *Sprite {
	return s.focus
}

// SetFocus updates the focus target and emits focus events.
func (s *Stage) SetFocus(sprite *Sprite) {
	if s.focus == sprite {
		return
	}
	prev := s.focus
	s.focus = sprite
	if prev != nil {
		prev.Emit(EventFocusOut, prev)
	}
	if sprite != nil {
		sprite.Emit(EventFocusIn, sprite)
	}
}

// Capture returns the sprite currently capturing pointer events.
func (s *Stage) Capture() *Sprite {
	return s.capture
}

// SetCapture routes future pointer events to the provided sprite.
func (s *Stage) SetCapture(sprite *Sprite) {
	s.capture = sprite
}

// ReleaseCapture clears the current pointer capture target.
func (s *Stage) ReleaseCapture() {
	s.capture = nil
}

// AddChild attaches a sprite to the root.
func (s *Stage) AddChild(child *Sprite) {
	s.root.AddChild(child)
}

// RemoveChild detaches a sprite from the root.
func (s *Stage) RemoveChild(child *Sprite) {
	s.root.RemoveChild(child)
}

// Update advances the scheduler and processes input events.
func (s *Stage) Update(delta time.Duration, mouse MouseState) {
	s.UpdateInput(delta, InputState{Mouse: mouse})
}

// UpdateInput advances the scheduler and processes a full frame of input.
func (s *Stage) UpdateInput(delta time.Duration, input InputState) {
	s.scheduler.Advance(delta)
	s.prevMouse = s.mouse
	s.mouse = input.Mouse
	s.modState = input.Mouse.Modifiers

	s.handleMouseInput(input.Mouse)
	if len(input.Touches) > 0 {
		s.handleTouches(input.Touches)
	}
	if len(input.Keys) > 0 {
		s.handleKeyEvents(input.Keys)
	}
}

func (s *Stage) handleMouseInput(mouse MouseState) {
	point := Point{X: mouse.X, Y: mouse.Y}
	actual := s.hitTest(point)

	if actual != s.hover {
		if s.hover != nil {
			event := s.pointerEvent(point, s.hover, actual, s.prevMouse, s.activeID, TouchPhaseMove)
			s.hover.EmitWithBubble(EventRollOut, event)
		}
		if actual != nil {
			event := s.pointerEvent(point, actual, actual, s.mouse, s.activeID, TouchPhaseMove)
			actual.EmitWithBubble(EventRollOver, event)
		}
		s.hover = actual
	}

	target := actual
	if s.capture != nil {
		target = s.capture
	}

	moved := s.mouse.X != s.prevMouse.X || s.mouse.Y != s.prevMouse.Y
	if moved {
		moveEvent := s.pointerEvent(point, target, actual, s.mouse, s.activeID, TouchPhaseMove)
		if target != nil {
			target.EmitWithBubble(EventMouseMove, moveEvent)
			if s.activeID >= 0 {
				target.EmitWithBubble(EventTouchMove, moveEvent)
			}
		}
		stageEvent := moveEvent
		stageEvent.Target = actual
		stageEvent.Hit = actual
		s.root.Dispatcher().Emit(EventMouseMove, stageEvent)
		if s.activeID >= 0 {
			s.root.Dispatcher().Emit(EventTouchMove, stageEvent)
		}
	}

	primary := mouse.Primary || mouse.Buttons.Left
	prevPrimary := s.prevMouse.Primary || s.prevMouse.Buttons.Left

	if primary && !prevPrimary {
		s.pressed = target
		s.touchSeq++
		s.activeID = s.touchSeq
		downEvent := s.pointerEvent(point, target, actual, s.mouse, s.activeID, TouchPhaseBegin)

		if target != nil {
			target.EmitWithBubble(EventMouseDown, downEvent)
			target.EmitWithBubble(EventTouchBegin, downEvent)
		}
		stageEvent := downEvent
		stageEvent.Target = actual
		stageEvent.Hit = actual
		s.root.Dispatcher().Emit(EventStageMouseDown, stageEvent)
		s.root.Dispatcher().Emit(EventTouchBegin, stageEvent)
	}

	if !primary && prevPrimary {
		currentID := s.activeID
		upEvent := s.pointerEvent(point, target, actual, s.mouse, currentID, TouchPhaseEnd)
		releaseTarget := s.pressed
		if releaseTarget == nil {
			releaseTarget = target
		}
		if releaseTarget != nil {
			event := upEvent
			event.Target = releaseTarget
			releaseTarget.EmitWithBubble(EventMouseUp, event)
			if releaseTarget == target && target != nil {
				clickEvent := event
				releaseTarget.EmitWithBubble(EventClick, clickEvent)
			}
			releaseTarget.EmitWithBubble(EventTouchEnd, event)
		}
		stageEvent := upEvent
		stageEvent.Target = actual
		stageEvent.Hit = actual
		s.root.Dispatcher().Emit(EventStageMouseUp, stageEvent)
		s.root.Dispatcher().Emit(EventTouchEnd, stageEvent)
		s.pressed = nil
		s.activeID = -1
	}

	if mouse.WheelX != 0 || mouse.WheelY != 0 {
		wheelEvent := s.pointerEvent(point, target, actual, s.mouse, s.activeID, TouchPhaseMove)
		if target != nil {
			target.EmitWithBubble(EventMouseWheel, wheelEvent)
		}
		stageEvent := wheelEvent
		stageEvent.Target = actual
		stageEvent.Hit = actual
		s.root.Dispatcher().Emit(EventMouseWheel, stageEvent)
	}
}

func (s *Stage) handleTouches(inputs []TouchInput) {
	for _, touch := range inputs {
		point := touch.Position
		actual := s.hitTest(point)
		target := actual
		if existing := s.touchTargets[touch.ID]; existing != nil {
			target = existing
		}
		if target == nil {
			target = actual
		}
		if touch.Phase == TouchPhaseBegin {
			if s.capture != nil {
				target = s.capture
			}
			s.touchTargets[touch.ID] = target
		} else if s.capture != nil {
			target = s.capture
		}
		event := s.touchPointerEvent(touch, target, actual)

		switch touch.Phase {
		case TouchPhaseBegin:
			if target != nil {
				target.EmitWithBubble(EventTouchBegin, event)
			}
			stageEvent := event
			stageEvent.Target = actual
			stageEvent.Hit = actual
			s.root.Dispatcher().Emit(EventTouchBegin, stageEvent)
		case TouchPhaseMove:
			if target != nil {
				target.EmitWithBubble(EventTouchMove, event)
			}
			stageEvent := event
			stageEvent.Target = actual
			stageEvent.Hit = actual
			s.root.Dispatcher().Emit(EventTouchMove, stageEvent)
		case TouchPhaseEnd:
			if target != nil {
				target.EmitWithBubble(EventTouchEnd, event)
			}
			stageEvent := event
			stageEvent.Target = actual
			stageEvent.Hit = actual
			s.root.Dispatcher().Emit(EventTouchEnd, stageEvent)
			delete(s.touchTargets, touch.ID)
		case TouchPhaseCancel:
			if target != nil {
				target.EmitWithBubble(EventTouchEnd, event)
			}
			stageEvent := event
			stageEvent.Target = actual
			stageEvent.Hit = actual
			s.root.Dispatcher().Emit(EventTouchCancel, stageEvent)
			delete(s.touchTargets, touch.ID)
		}
	}
}

func (s *Stage) handleKeyEvents(events []KeyboardEvent) {
	if s.keysDown == nil {
		s.keysDown = make(map[KeyCode]bool)
	}
	for _, ke := range events {
		if ke.Down {
			s.keysDown[ke.Code] = true
		} else {
			delete(s.keysDown, ke.Code)
		}
		s.modState = ke.Modifiers

		target := s.focus
		if target == nil {
			target = s.root
		}

		if ke.Down {
			if target != nil {
				target.EmitWithBubble(EventKeyDown, ke)
			}
			if ke.Rune != 0 {
				if target != nil {
					target.EmitWithBubble(EventKeyPress, ke)
				}
				s.root.Dispatcher().Emit(EventKeyPress, ke)
			}
			if target != s.root && s.root != nil {
				s.root.Dispatcher().Emit(EventKeyDown, ke)
			}
		} else {
			if target != nil {
				target.EmitWithBubble(EventKeyUp, ke)
			}
			if target != s.root && s.root != nil {
				s.root.Dispatcher().Emit(EventKeyUp, ke)
			}
		}
	}
}

// HitTest returns the topmost sprite at the provided stage coordinates.
func (s *Stage) HitTest(pt Point) *Sprite {
	return s.hitTest(pt)
}

func (s *Stage) pointerEvent(point Point, target, hit *Sprite, mouse MouseState, touchID int, phase TouchPhase) PointerEvent {
	if touchID < 0 {
		touchID = 0
	}
	buttons := mouse.Buttons
	if !buttons.Left && !buttons.Right && !buttons.Middle && mouse.Primary {
		buttons.Left = true
	}
	return PointerEvent{
		Position:  point,
		Primary:   mouse.Primary || buttons.Left,
		Target:    target,
		Hit:       hit,
		WheelX:    mouse.WheelX,
		WheelY:    mouse.WheelY,
		TouchID:   touchID,
		Buttons:   buttons,
		Modifiers: mouse.Modifiers,
		Phase:     phase,
	}
}

func (s *Stage) hitTest(pt Point) *Sprite {
	if hit := s.root.HitTest(pt); hit != nil && hit != s.root {
		return hit
	}
	return nil
}

func (s *Stage) touchPointerEvent(touch TouchInput, target, hit *Sprite) PointerEvent {
	return PointerEvent{
		Position:  touch.Position,
		Primary:   touch.Primary,
		Target:    target,
		Hit:       hit,
		TouchID:   touch.ID,
		Modifiers: s.modState,
		Phase:     touch.Phase,
	}
}
