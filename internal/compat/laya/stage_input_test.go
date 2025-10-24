package laya_test

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
)

func TestStageTouchEvents(t *testing.T) {
	env := testutil.NewStageEnv(t, 400, 400)

	sprite := laya.NewSprite()
	sprite.SetSize(200, 200)
	env.Stage.AddChild(sprite)

	var begin, move, end int
	var lastBegin laya.PointerEvent
	sprite.Dispatcher().On(laya.EventTouchBegin, func(evt laya.Event) {
		begin++
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			lastBegin = pe
		} else {
			t.Fatalf("expected pointer event payload")
		}
	})
	sprite.Dispatcher().On(laya.EventTouchMove, func(evt laya.Event) {
		move++
	})
	sprite.Dispatcher().On(laya.EventTouchEnd, func(evt laya.Event) {
		end++
	})

	var stageBeginTarget *laya.Sprite
	env.Stage.Root().Dispatcher().On(laya.EventTouchBegin, func(evt laya.Event) {
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			stageBeginTarget = pe.Target
		}
	})

	env.AdvanceInput(16*time.Millisecond, laya.InputState{
		Touches: []laya.TouchInput{
			{ID: 1, Position: laya.Point{X: 50, Y: 50}, Phase: laya.TouchPhaseBegin, Primary: true},
		},
	})
	if begin != 1 {
		t.Fatalf("expected 1 touch begin, got %d", begin)
	}
	if lastBegin.Target != sprite {
		t.Fatalf("expected touch begin target sprite, got %v", lastBegin.Target)
	}
	if lastBegin.TouchID != 1 {
		t.Fatalf("expected touch ID 1, got %d", lastBegin.TouchID)
	}
	if stageBeginTarget != sprite {
		t.Fatalf("expected stage event target sprite, got %v", stageBeginTarget)
	}

	env.AdvanceInput(16*time.Millisecond, laya.InputState{
		Touches: []laya.TouchInput{
			{ID: 1, Position: laya.Point{X: 80, Y: 80}, Phase: laya.TouchPhaseMove, Primary: true},
		},
	})
	if move != 1 {
		t.Fatalf("expected move to fire once, got %d", move)
	}

	env.AdvanceInput(16*time.Millisecond, laya.InputState{
		Touches: []laya.TouchInput{
			{ID: 1, Position: laya.Point{X: 80, Y: 80}, Phase: laya.TouchPhaseEnd, Primary: true},
		},
	})
	if end != 1 {
		t.Fatalf("expected end to fire once, got %d", end)
	}
}

func TestStageKeyboardFocus(t *testing.T) {
	env := testutil.NewStageEnv(t, 200, 200)

	focus := laya.NewSprite()
	env.Stage.AddChild(focus)
	env.Stage.SetFocus(focus)

	var keyDown, keyUp int
	focus.Dispatcher().On(laya.EventKeyDown, func(evt laya.Event) {
		keyDown++
	})
	focus.Dispatcher().On(laya.EventKeyUp, func(evt laya.Event) {
		keyUp++
	})

	env.AdvanceInput(16*time.Millisecond, laya.InputState{
		Keys: []laya.KeyboardEvent{
			{Code: laya.KeyCode(65), Down: true},
		},
	})
	if keyDown != 1 {
		t.Fatalf("expected key down to reach focus target")
	}

	env.AdvanceInput(16*time.Millisecond, laya.InputState{
		Keys: []laya.KeyboardEvent{
			{Code: laya.KeyCode(65), Down: false},
		},
	})
	if keyUp != 1 {
		t.Fatalf("expected key up to reach focus target")
	}
}

func TestStagePointerCapture(t *testing.T) {
	env := testutil.NewStageEnv(t, 400, 400)

	capture := laya.NewSprite()
	capture.SetSize(50, 50)
	env.Stage.AddChild(capture)

	actual := laya.NewSprite()
	actual.SetPosition(200, 200)
	actual.SetSize(50, 50)
	env.Stage.AddChild(actual)

	env.Stage.SetCapture(capture)

	var captureDown int
	capture.Dispatcher().On(laya.EventMouseDown, func(evt laya.Event) {
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			if pe.Target != capture {
				t.Fatalf("expected capture target, got %+v", pe.Target)
			}
		}
		captureDown++
	})

	var rootTarget *laya.Sprite
	env.Stage.Root().Dispatcher().On(laya.EventStageMouseDown, func(evt laya.Event) {
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			rootTarget = pe.Target
		}
	})

	mouse := laya.MouseState{
		X:       float64(225),
		Y:       float64(225),
		Primary: true,
		Buttons: laya.MouseButtons{Left: true},
	}
	env.AdvanceInput(16*time.Millisecond, laya.InputState{Mouse: mouse})
	if captureDown != 1 {
		t.Fatalf("expected capture target to receive mouse down")
	}
	if rootTarget != actual {
		t.Fatalf("expected stage event to use actual hit sprite, got %v", rootTarget)
	}

	env.Stage.ReleaseCapture()
}
