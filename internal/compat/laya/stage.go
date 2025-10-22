package laya

import (
	"time"
)

// PointerEvent describes pointer interaction dispatched through sprites.
type PointerEvent struct {
	Position Point
	Primary  bool
	Target   *Sprite
	WheelX   float64
	WheelY   float64
	TouchID  int
}

// MouseState describes mouse/touch input for the stage update.
type MouseState struct {
	X       float64
	Y       float64
	Primary bool
	WheelX  float64
	WheelY  float64
}

// Stage emulates the behaviour of Laya's global stage.
type Stage struct {
	root      *Sprite
	scheduler *Scheduler
	width     int
	height    int

	mouse     MouseState
	prevMouse MouseState
	hover     *Sprite
	pressed   *Sprite
	touchSeq  int
	activeID  int
}

// NewStage constructs a stage with the provided dimensions.
func NewStage(width, height int) *Stage {
	root := NewSprite()
	root.SetName("stage")
	root.SetSize(float64(width), float64(height))
	return &Stage{
		root:      root,
		scheduler: NewScheduler(),
		width:     width,
		height:    height,
		activeID:  -1,
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
	s.scheduler.Advance(delta)
	s.prevMouse = s.mouse
	s.mouse = mouse
	point := Point{X: mouse.X, Y: mouse.Y}
	hit := s.hitTest(point)

	if hit != s.hover {
		if s.hover != nil {
			event := s.pointerEvent(point, s.hover, mouse, s.activeID)
			s.hover.EmitWithBubble(EventRollOut, event)
		}
		if hit != nil {
			event := s.pointerEvent(point, hit, mouse, s.activeID)
			hit.EmitWithBubble(EventRollOver, event)
		}
		s.hover = hit
	}

	if s.mouse.X != s.prevMouse.X || s.mouse.Y != s.prevMouse.Y {
		event := s.pointerEvent(point, hit, mouse, s.activeID)
		if hit != nil {
			hit.EmitWithBubble(EventMouseMove, event)
			if s.activeID >= 0 {
				hit.EmitWithBubble(EventTouchMove, event)
			}
		}
		s.root.Dispatcher().Emit(EventMouseMove, event)
		if s.activeID >= 0 {
			s.root.Dispatcher().Emit(EventTouchMove, event)
		}
	}

	if s.mouse.Primary && !s.prevMouse.Primary {
		s.pressed = hit
		s.touchSeq++
		s.activeID = s.touchSeq
		event := s.pointerEvent(point, hit, mouse, s.activeID)
		if hit != nil {
			hit.EmitWithBubble(EventMouseDown, event)
			hit.EmitWithBubble(EventTouchBegin, event)
		}
		s.root.Dispatcher().Emit(EventStageMouseDown, event)
		s.root.Dispatcher().Emit(EventTouchBegin, event)
	}
	if !s.mouse.Primary && s.prevMouse.Primary {
		currentID := s.activeID
		event := s.pointerEvent(point, hit, mouse, currentID)
		target := s.pressed
		if target == nil {
			target = hit
		}
		if target != nil {
			target.EmitWithBubble(EventMouseUp, event)
			if target == hit && hit != nil {
				target.EmitWithBubble(EventClick, event)
			}
			target.EmitWithBubble(EventTouchEnd, event)
		}
		s.root.Dispatcher().Emit(EventStageMouseUp, event)
		s.root.Dispatcher().Emit(EventTouchEnd, event)
		s.pressed = nil
		s.activeID = -1
	}
	if mouse.WheelX != 0 || mouse.WheelY != 0 {
		event := s.pointerEvent(point, hit, mouse, s.activeID)
		if hit != nil {
			hit.EmitWithBubble(EventMouseWheel, event)
		}
		s.root.Dispatcher().Emit(EventMouseWheel, event)
	}
	s.hover = hit
}

// HitTest returns the topmost sprite at the provided stage coordinates.
func (s *Stage) HitTest(pt Point) *Sprite {
	return s.hitTest(pt)
}

func (s *Stage) pointerEvent(point Point, target *Sprite, mouse MouseState, touchID int) PointerEvent {
	if touchID < 0 {
		touchID = 0
	}
	return PointerEvent{
		Position: point,
		Primary:  mouse.Primary,
		Target:   target,
		WheelX:   mouse.WheelX,
		WheelY:   mouse.WheelY,
		TouchID:  touchID,
	}
}

func (s *Stage) hitTest(pt Point) *Sprite {
	if hit := s.root.HitTest(pt); hit != nil && hit != s.root {
		return hit
	}
	return nil
}
