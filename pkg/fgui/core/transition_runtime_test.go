package core

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/tween"
)

func TestTransitionPlayMovesChild(t *testing.T) {
	comp := NewGComponent()
	child := NewGObject()
	child.SetResourceID("child")
	comp.AddChild(child)

	info := TransitionInfo{
		Name: "move",
		Items: []TransitionItem{
			{
				Time:     0,
				TargetID: "child",
				Type:     TransitionActionXY,
				Tween: &TransitionTween{
					Duration: 0.5,
					Start: TransitionValue{
						B1: true,
						B2: true,
						F1: 0,
						F2: 0,
					},
					End: TransitionValue{
						B1: true,
						B2: true,
						F1: 120,
						F2: 45,
					},
				},
			},
		},
		TotalDuration: 0.5,
	}

	comp.AddTransition(info)
	tx := comp.Transition("move")
	if tx == nil {
		t.Fatalf("expected transition move to be registered")
	}

	child.SetPosition(0, 0)
	tx.Play(1, 0)

	tween.Advance(250 * time.Millisecond)
	if child.X() == 0 && child.Y() == 0 {
		t.Fatalf("expected child position to change during transition")
	}

	tween.Advance(400 * time.Millisecond)
	if tx.Playing() {
		t.Fatalf("expected transition to finish after duration")
	}

	if math.Abs(child.X()-120) > 0.1 || math.Abs(child.Y()-45) > 0.1 {
		t.Fatalf("expected child to reach target position, got (%.2f, %.2f)", child.X(), child.Y())
	}

	tx.Stop(false)
}

func TestTransitionShakeColorFilterSound(t *testing.T) {
	rand.Seed(1)

	comp := NewGComponent()
	child := NewGObject()
	child.SetResourceID("child")
	comp.AddChild(child)

	soundCalls := 0
	SetTransitionSoundPlayer(func(name string, volume float64) {
		if name != "ui://sound/click" {
			t.Fatalf("unexpected sound name %s", name)
		}
		if volume != 0.5 {
			t.Fatalf("unexpected volume %v", volume)
		}
		soundCalls++
	})
	t.Cleanup(func() {
		SetTransitionSoundPlayer(nil)
	})

	info := TransitionInfo{
		Name: "fx",
		Items: []TransitionItem{
			{
				Time:     0,
				TargetID: "child",
				Type:     TransitionActionShake,
				Tween: &TransitionTween{
					Duration: 0.2,
					Start:    TransitionValue{Amplitude: 10},
					End:      TransitionValue{Amplitude: 10},
				},
			},
			{
				Time: 0,
				Type: TransitionActionSound,
				Value: TransitionValue{
					Sound:  "ui://sound/click",
					Volume: 0.5,
				},
			},
			{
				Time:     0,
				TargetID: "child",
				Type:     TransitionActionColorFilter,
				Tween: &TransitionTween{
					Duration: 0.2,
					Start:    TransitionValue{F1: 0, F2: 0, F3: 0, F4: 0},
					End:      TransitionValue{F1: 1.1, F2: 0.9, F3: 0.8, F4: 1},
				},
			},
		},
		TotalDuration: 0.2,
	}

	comp.AddTransition(info)
	tx := comp.Transition("fx")
	if tx == nil {
		t.Fatalf("expected transition fx")
	}

	child.SetPosition(0, 0)
	tx.Play(1, 0)

	tween.Advance(50 * time.Millisecond)
	if soundCalls != 1 {
		t.Fatalf("expected sound callback once, got %d", soundCalls)
	}

	tween.Advance(100 * time.Millisecond)
	if child.X() == 0 && child.Y() == 0 {
		t.Fatalf("expected shake to offset position")
	}
	enabled, filter := child.ColorFilter()
	if !enabled {
		t.Fatalf("expected color filter enabled")
	}
	if filter[0] == 0 && filter[1] == 0 && filter[2] == 0 && filter[3] == 0 {
		t.Fatalf("expected color filter values to change")
	}

	tween.Advance(200 * time.Millisecond)
	if child.X() != 0 || child.Y() != 0 {
		t.Fatalf("expected shake offsets to reset, got (%.2f, %.2f)", child.X(), child.Y())
	}
	enabled, filter = child.ColorFilter()
	if !enabled {
		t.Fatalf("expected color filter to remain enabled")
	}
	if math.Abs(filter[0]-1.1) > 1e-6 || math.Abs(filter[1]-0.9) > 1e-6 {
		t.Fatalf("unexpected color filter final values %+v", filter)
	}
	tx.Stop(false)
}

func TestTransitionAnimationAppliesPlayingAndFrame(t *testing.T) {
	comp := NewGComponent()
	target := NewGObject()
	stub := &fakeAnimationWidget{
		playing: true,
		frame:   1,
	}
	target.SetData(stub)

	comp.AddChild(target)

	info := TransitionInfo{
		Name: "anim",
		Items: []TransitionItem{
			{
				Time:     0,
				TargetID: target.ID(),
				Type:     TransitionActionAnimation,
				Value: TransitionValue{
					Playing: false,
					Frame:   5,
				},
			},
			{
				Time:     0.2,
				TargetID: target.ID(),
				Type:     TransitionActionAnimation,
				Value: TransitionValue{
					Playing: true,
					Frame:   -1,
				},
			},
		},
		TotalDuration: 0.2,
	}

	comp.AddTransition(info)
	tx := comp.Transition("anim")
	if tx == nil {
		t.Fatalf("expected transition anim to be present")
	}

	tx.Play(1, 0)

	tween.Advance(10 * time.Millisecond)
	if stub.playing {
		t.Fatalf("expected loader to pause after first animation action")
	}
	if stub.frame != 5 {
		t.Fatalf("expected loader frame to update to 5, got %d", stub.frame)
	}
	if ts, ok := target.GetProp(gears.ObjectPropIDTimeScale).(float64); !ok || math.Abs(ts-1) > 1e-6 {
		t.Fatalf("expected default time scale 1, got %v", target.GetProp(gears.ObjectPropIDTimeScale))
	}
	tx.SetTimeScale(0.5)
	if math.Abs(tx.TimeScale()-0.5) > 1e-6 {
		t.Fatalf("expected transition time scale to update to 0.5, got %.3f", tx.TimeScale())
	}
	if ts, ok := target.GetProp(gears.ObjectPropIDTimeScale).(float64); !ok || math.Abs(ts-0.5) > 1e-6 {
		t.Fatalf("expected target time scale 0.5, got %v", target.GetProp(gears.ObjectPropIDTimeScale))
	}

	tween.Advance(200 * time.Millisecond)
	if stub.playing {
		t.Fatalf("expected loader to remain paused before second animation action")
	}
	tween.Advance(220 * time.Millisecond)
	if !stub.playing {
		t.Fatalf("expected loader to resume playing on second animation action")
	}
	if stub.frame != 5 {
		t.Fatalf("expected loader frame to remain at 5 when frame is -1, got %d", stub.frame)
	}

	tween.Advance(100 * time.Millisecond)
	if tx.Playing() {
		t.Fatalf("expected transition to finish after final action")
	}
	tx.Stop(false)
}

func TestTransitionAnimationStopCompleteAdvancesDelta(t *testing.T) {
	comp := NewGComponent()
	target := NewGObject()
	target.SetResourceID("anim-stop")
	stub := &fakeAnimationWidget{}
	target.SetData(stub)
	comp.AddChild(target)

	info := TransitionInfo{
		Name: "anim-stop",
		Items: []TransitionItem{
			{
				Time:     0,
				TargetID: target.ID(),
				Type:     TransitionActionAnimation,
				Value: TransitionValue{
					Playing: true,
					Frame:   0,
				},
			},
			{
				Time:     0.3,
				TargetID: target.ID(),
				Type:     TransitionActionAnimation,
				Value: TransitionValue{
					Playing: false,
					Frame:   6,
				},
			},
		},
		TotalDuration: 0.3,
	}

	comp.AddTransition(info)
	tx := comp.Transition("anim-stop")
	if tx == nil {
		t.Fatalf("expected transition to exist")
	}

	tx.Play(1, 0)
	tx.Stop(true)

	if stub.playing {
		t.Fatalf("expected stub to be paused after complete stop")
	}
	if stub.frame != 6 {
		t.Fatalf("expected frame 6 after complete stop, got %d", stub.frame)
	}
	if math.Abs(stub.delta-300) > 1e-3 {
		t.Fatalf("expected delta 300ms, got %.3f", stub.delta)
	}
}

func TestTransitionXYPathTween(t *testing.T) {
	comp := NewGComponent()
	child := NewGObject()
	child.SetResourceID("child")
	comp.AddChild(child)

	child.SetPosition(5, 5)

	info := TransitionInfo{
		Name: "path",
		Items: []TransitionItem{
			{
				Time:     0,
				TargetID: "child",
				Type:     TransitionActionXY,
				Tween: &TransitionTween{
					Duration: 1,
					Start:    TransitionValue{},
					End:      TransitionValue{},
					Path: []TransitionPathPoint{
						{CurveType: transitionCurveStraight, X: 0, Y: 0},
						{CurveType: transitionCurveStraight, X: 20, Y: 10},
					},
				},
			},
		},
		TotalDuration: 1,
	}

	comp.AddTransition(info)
	tx := comp.Transition("path")
	if tx == nil {
		t.Fatalf("expected path transition to be registered")
	}

	tx.Play(1, 0)

	tween.Advance(500 * time.Millisecond)
	if math.Abs(child.X()-15) > 1e-3 || math.Abs(child.Y()-10) > 1e-3 {
		t.Fatalf("unexpected midpoint position (%.3f, %.3f)", child.X(), child.Y())
	}

	tween.Advance(600 * time.Millisecond)
	if tx.Playing() {
		t.Fatalf("expected transition to finish after duration")
	}
	if math.Abs(child.X()-25) > 1e-3 || math.Abs(child.Y()-15) > 1e-3 {
		t.Fatalf("unexpected final position (%.3f, %.3f)", child.X(), child.Y())
	}
	tx.Stop(false)
}

type fakeAnimationWidget struct {
	playing bool
	frame   int
	delta   float64
}

func (f *fakeAnimationWidget) Playing() bool { return f.playing }
func (f *fakeAnimationWidget) SetPlaying(v bool) {
	f.playing = v
}
func (f *fakeAnimationWidget) Frame() int { return f.frame }
func (f *fakeAnimationWidget) SetFrame(v int) {
	f.frame = v
}
func (f *fakeAnimationWidget) DeltaTime() float64 { return f.delta }
func (f *fakeAnimationWidget) SetDeltaTime(v float64) {
	f.delta += v
}
