package laya

import "time"

// PointerEvent describes pointer interaction dispatched through sprites.
type PointerEvent struct {
	Position Point
	Primary  bool
	Target   *Sprite
}

// MouseState describes mouse/touch input for the stage update.
type MouseState struct {
	X       float64
	Y       float64
	Primary bool
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

	if s.mouse.X != s.prevMouse.X || s.mouse.Y != s.prevMouse.Y {
		s.root.Dispatcher().Emit(EventMouseMove, PointerEvent{
			Position: point,
			Primary:  mouse.Primary,
			Target:   hit,
		})
	}

	if s.mouse.Primary && !s.prevMouse.Primary {
		s.pressed = hit
		event := PointerEvent{
			Position: point,
			Primary:  true,
			Target:   hit,
		}
		if hit != nil {
			hit.EmitWithBubble(EventMouseDown, event)
		}
		s.root.Dispatcher().Emit(EventStageMouseDown, event)
	}
	if !s.mouse.Primary && s.prevMouse.Primary {
		event := PointerEvent{
			Position: point,
			Primary:  false,
			Target:   hit,
		}
		target := s.pressed
		if target == nil {
			target = hit
		}
		if target != nil {
			target.EmitWithBubble(EventMouseUp, event)
			if target == hit && hit != nil {
				target.EmitWithBubble(EventClick, event)
			}
		}
		s.root.Dispatcher().Emit(EventStageMouseUp, event)
		s.pressed = nil
	}
	s.hover = hit
}

// HitTest returns the topmost sprite at the provided stage coordinates.
func (s *Stage) HitTest(pt Point) *Sprite {
	return s.hitTest(pt)
}

func (s *Stage) hitTest(pt Point) *Sprite {
	if hit := s.root.HitTest(pt); hit != nil && hit != s.root {
		return hit
	}
	return nil
}
