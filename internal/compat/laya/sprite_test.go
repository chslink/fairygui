package laya_test

import (
	"math"
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
)

func TestSpriteLocalToGlobal(t *testing.T) {
	parent := laya.NewSprite()
	child := laya.NewSprite()
	parent.SetPosition(10, 20)
	child.SetPosition(5, 6)
	parent.AddChild(child)

	global := child.LocalToGlobal(laya.Point{})
	if global.X != 15 || global.Y != 26 {
		t.Fatalf("expected global (15,26), got (%v,%v)", global.X, global.Y)
	}
}

func TestStageMouseEvents(t *testing.T) {
	env := testutil.NewStageEnv(t, 800, 600)
	stage := env.Stage

	container := laya.NewSprite()
	container.SetSize(200, 200)
	stage.AddChild(container)

	child := laya.NewSprite()
	child.SetSize(100, 100)
	child.SetPosition(50, 50)
	container.AddChild(child)

	var order []string
	stageMoveCalled := false
	stage.Root().Dispatcher().On(laya.EventMouseMove, func(evt laya.Event) {
		stageMoveCalled = true
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			if pe.Target != child {
				t.Fatalf("expected move target child, got %v", pe.Target)
			}
		} else {
			t.Fatalf("expected PointerEvent payload")
		}
	})

	stage.Root().Dispatcher().On(laya.EventStageMouseDown, func(evt laya.Event) {
		order = append(order, "stage-down")
		if pe, ok := evt.Data.(laya.PointerEvent); !ok || pe.Target != child {
			t.Fatalf("stage down should target child")
		}
	})
	stage.Root().Dispatcher().On(laya.EventStageMouseUp, func(evt laya.Event) {
		order = append(order, "stage-up")
	})

	container.Dispatcher().On(laya.EventMouseDown, func(evt laya.Event) {
		order = append(order, "parent-down")
	})
	container.Dispatcher().On(laya.EventMouseUp, func(evt laya.Event) {
		order = append(order, "parent-up")
	})
	container.Dispatcher().On(laya.EventClick, func(evt laya.Event) {
		order = append(order, "parent-click")
	})

	child.Dispatcher().On(laya.EventMouseDown, func(evt laya.Event) {
		order = append(order, "child-down")
	})
	child.Dispatcher().On(laya.EventMouseUp, func(evt laya.Event) {
		order = append(order, "child-up")
	})
	child.Dispatcher().On(laya.EventClick, func(evt laya.Event) {
		order = append(order, "child-click")
	})

	env.Advance(time.Millisecond*16, laya.MouseState{X: 80, Y: 80, Primary: false})
	env.Advance(time.Millisecond*16, laya.MouseState{X: 80, Y: 80, Primary: true})
	env.Advance(time.Millisecond*16, laya.MouseState{X: 80, Y: 80, Primary: false})

	if !stageMoveCalled {
		t.Fatalf("expected mouse move event")
	}

	expectedOrder := []string{
		"child-down",
		"parent-down",
		"stage-down",
		"child-up",
		"parent-up",
		"child-click",
		"parent-click",
		"stage-up",
	}
	if len(order) != len(expectedOrder) {
		t.Fatalf("unexpected order length: got %v", order)
	}
	for i, expected := range expectedOrder {
		if order[i] != expected {
			t.Fatalf("unexpected event order at %d: expected %s got %s", i, expected, order[i])
		}
	}
}

func TestSpritePivotRotationBounds(t *testing.T) {
	sprite := laya.NewSprite()
	sprite.SetSize(100, 50)
	sprite.SetPivotWithAnchor(0.5, 0.5, true)
	sprite.SetPosition(200, 300)
	sprite.SetRotation(math.Pi / 2) // 90 degrees

	centerGlobal := sprite.LocalToGlobal(laya.Point{X: 50, Y: 25})
	wantX, wantY := expectedPivotGlobal(100, 50, 0.5, 0.5, 200, 300, math.Pi/2, 0, 0, 1, 1)
	if math.Abs(centerGlobal.X-wantX) > 1e-6 || math.Abs(centerGlobal.Y-wantY) > 1e-6 {
		t.Fatalf("unexpected pivot location got (%v,%v) want (%v,%v)", centerGlobal.X, centerGlobal.Y, wantX, wantY)
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

func TestStageHitTestOutside(t *testing.T) {
	env := testutil.NewStageEnv(t, 800, 600)
	stage := env.Stage
	child := laya.NewSprite()
	child.SetSize(100, 100)
	child.SetPosition(10, 10)
	stage.AddChild(child)

	// keep event log to ensure nothing fires
	log := &testutil.EventLog{}
	testutil.AttachEventLog(log, child, laya.EventMouseDown, laya.EventClick)

	env.Advance(time.Millisecond*16, laya.MouseState{X: 500, Y: 500, Primary: false})
	env.Advance(time.Millisecond*16, laya.MouseState{X: 500, Y: 500, Primary: true})
	env.Advance(time.Millisecond*16, laya.MouseState{X: 500, Y: 500, Primary: false})

	if len(log.Records) != 0 {
		t.Fatalf("expected no events on child, got %v", log.Records)
	}
}

func TestSpriteHitTestVisibility(t *testing.T) {
	sprite := laya.NewSprite()
	sprite.SetSize(100, 100)
	sprite.SetVisible(false)

	if hit := sprite.HitTest(laya.Point{X: 10, Y: 10}); hit != nil {
		t.Fatalf("expected no hit on invisible sprite")
	}
	sprite.SetVisible(true)
	if hit := sprite.HitTest(laya.Point{X: 200, Y: 200}); hit != nil {
		t.Fatalf("expected no hit outside bounds")
	}
}

func TestSpriteCustomHitTester(t *testing.T) {
	sprite := laya.NewSprite()
	sprite.SetSize(10, 10)
	sprite.SetHitTester(func(x, y float64) bool {
		return x < 5
	})

	if sprite.HitTest(laya.Point{X: 4, Y: 5}) != sprite {
		t.Fatalf("expected hit in custom region")
	}
	if sprite.HitTest(laya.Point{X: 8, Y: 4}) != nil {
		t.Fatalf("expected miss outside custom region")
	}
}

func TestSpritePivotAnchorPosition(t *testing.T) {
	sprite := laya.NewSprite()
	sprite.SetSize(120, 80)
	sprite.SetScale(1.4, 0.8)
	sprite.SetSkew(0.2, -0.1)
	sprite.SetRotation(0.35)
	sprite.SetPivotWithAnchor(0.3, 0.6, true)

	sprite.SetPosition(200, 150)

	offX, offY := expectedPivotOffset(120, 80, 0.3, 0.6, 0.35, 0.2, -0.1, 1.4, 0.8)
	wantX := 200 + offX
	wantY := 150 + offY
	got := sprite.Position()
	if math.Abs(got.X-wantX) > 1e-6 || math.Abs(got.Y-wantY) > 1e-6 {
		t.Fatalf("unexpected anchored position: got (%v,%v) want (%v,%v)", got.X, got.Y, wantX, wantY)
	}

	sprite.Move(10, -20)
	got = sprite.Position()
	wantX = 210 + offX
	wantY = 130 + offY
	if math.Abs(got.X-wantX) > 1e-6 || math.Abs(got.Y-wantY) > 1e-6 {
		t.Fatalf("unexpected anchored position after move: got (%v,%v) want (%v,%v)", got.X, got.Y, wantX, wantY)
	}

	sprite.SetSize(90, 40)
	offX, offY = expectedPivotOffset(90, 40, 0.3, 0.6, 0.35, 0.2, -0.1, 1.4, 0.8)
	got = sprite.Position()
	wantX = 210 + offX
	wantY = 130 + offY
	if math.Abs(got.X-wantX) > 1e-6 || math.Abs(got.Y-wantY) > 1e-6 {
		t.Fatalf("unexpected anchored position after resize: got (%v,%v) want (%v,%v)", got.X, got.Y, wantX, wantY)
	}
}

func TestSpritePivotWithoutAnchor(t *testing.T) {
	sprite := laya.NewSprite()
	sprite.SetSize(100, 60)
	sprite.SetScale(1.2, 0.9)
	sprite.SetRotation(0.25)
	sprite.SetSkew(-0.05, 0.1)

	sprite.SetPivotWithAnchor(0.4, 0.2, false)
	sprite.SetPosition(80, 40)

	offX, offY := expectedPivotOffset(100, 60, 0.4, 0.2, 0.25, -0.05, 0.1, 1.2, 0.9)
	got := sprite.Position()
	wantX := 80 + offX
	wantY := 40 + offY
	if math.Abs(got.X-wantX) > 1e-6 || math.Abs(got.Y-wantY) > 1e-6 {
		t.Fatalf("unexpected position with pivot (non-anchor): got (%v,%v) want (%v,%v)", got.X, got.Y, wantX, wantY)
	}
}

func TestSpriteOwner(t *testing.T) {
	sprite := laya.NewSprite()
	if sprite.Owner() != nil {
		t.Fatalf("expect nil owner by default")
	}

	owner := &struct {
		ID string
	}{"owner"}

	sprite.SetOwner(owner)
	if sprite.Owner() != owner {
		t.Fatalf("owner mismatch, got %#v want %#v", sprite.Owner(), owner)
	}

	sprite.SetOwner(nil)
	if sprite.Owner() != nil {
		t.Fatalf("expect nil owner after reset")
	}
}

func expectedPivotOffset(width, height, pivotX, pivotY, rotation, skewX, skewY, scaleX, scaleY float64) (float64, float64) {
	px := pivotX * width
	py := pivotY * height

	cosY := math.Cos(rotation + skewY)
	sinY := math.Sin(rotation + skewY)
	cosX := math.Cos(rotation - skewX)
	sinX := math.Sin(rotation - skewX)

	a := cosY * scaleX
	b := sinY * scaleX
	c := -sinX * scaleY
	d := cosX * scaleY

	transformedX := a*px + c*py
	transformedY := b*px + d*py
	return px - transformedX, py - transformedY
}

func expectedPivotGlobal(width, height, pivotX, pivotY float64, rawX, rawY, rotation, skewX, skewY, scaleX, scaleY float64) (float64, float64) {
	offX, offY := expectedPivotOffset(width, height, pivotX, pivotY, rotation, skewX, skewY, scaleX, scaleY)
	baseX := rawX + offX
	baseY := rawY + offY
	px := pivotX * width
	py := pivotY * height

	cosY := math.Cos(rotation + skewY)
	sinY := math.Sin(rotation + skewY)
	cosX := math.Cos(rotation - skewX)
	sinX := math.Sin(rotation - skewX)

	a := cosY * scaleX
	b := sinY * scaleX
	c := -sinX * scaleY
	d := cosX * scaleY

	tx := baseX - px*a - py*c
	ty := baseY - px*b - py*d

	globalX := a*px + c*py + tx
	globalY := b*px + d*py + ty
	return globalX, globalY
}
