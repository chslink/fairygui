package laya

import (
	"math"
	"testing"
	"time"
)

func TestSpriteLocalToGlobal(t *testing.T) {
	parent := NewSprite()
	child := NewSprite()
	parent.SetPosition(10, 20)
	child.SetPosition(5, 6)
	parent.AddChild(child)

	global := child.LocalToGlobal(Point{})
	if global.X != 15 || global.Y != 26 {
		t.Fatalf("expected global (15,26), got (%v,%v)", global.X, global.Y)
	}
}

func TestStageMouseEvents(t *testing.T) {
	stage := NewStage(800, 600)

	var downCalled, upCalled, moveCalled bool
	stage.Root().Dispatcher().On(EventStageMouseDown, func(_ Event) { downCalled = true })
	stage.Root().Dispatcher().On(EventStageMouseUp, func(_ Event) { upCalled = true })
	stage.Root().Dispatcher().On(EventMouseMove, func(_ Event) { moveCalled = true })

	stage.Update(time.Millisecond*16, MouseState{X: 100, Y: 50, Primary: false})
	stage.Update(time.Millisecond*16, MouseState{X: 120, Y: 70, Primary: true})
	stage.Update(time.Millisecond*16, MouseState{X: 120, Y: 70, Primary: false})

	if !moveCalled {
		t.Fatalf("expected mouse move event")
	}
	if !downCalled {
		t.Fatalf("expected mouse down event")
	}
	if !upCalled {
		t.Fatalf("expected mouse up event")
	}
}

func TestSpritePivotRotationBounds(t *testing.T) {
	sprite := NewSprite()
	sprite.SetSize(100, 50)
	sprite.SetPivot(0.5, 0.5)
	sprite.SetPosition(200, 300)
	sprite.SetRotation(math.Pi / 2) // 90 degrees

	centerGlobal := sprite.LocalToGlobal(Point{X: 50, Y: 25})
	if math.Abs(centerGlobal.X-200) > 1e-6 || math.Abs(centerGlobal.Y-300) > 1e-6 {
		t.Fatalf("expected center at (200,300), got (%v,%v)", centerGlobal.X, centerGlobal.Y)
	}

	bounds := sprite.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		t.Fatalf("expected positive bounds, got %+v", bounds)
	}

	local := sprite.GlobalToLocal(centerGlobal)
	if math.Abs(local.X-50) > 1e-6 || math.Abs(local.Y-25) > 1e-6 {
		t.Fatalf("global to local failed, got (%v,%v)", local.X, local.Y)
	}
}
