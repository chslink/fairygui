package tween

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(1)
}

func resetManager() {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()
	globalManager.tweeners = nil
}

func approxEqual(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.001 {
		t.Fatalf("value mismatch: got %.4f want %.4f", got, want)
	}
}

func TestTweenBasicCompletion(t *testing.T) {
	resetManager()
	var starts, updates, completes int
	var last Value
	tw := To2(0, 0, 10, 20, 1).
		OnStart(func(*GTweener) { starts++ }).
		OnUpdate(func(tw *GTweener) {
			updates++
			last = tw.Value()
		}).
		OnComplete(func(tw *GTweener) {
			completes++
			last = tw.Value()
		})

	Advance(500 * time.Millisecond)
	if starts != 1 {
		t.Fatalf("expected start once after first advance, got %d", starts)
	}
	if updates == 0 {
		t.Fatal("expected update callbacks after first advance")
	}

	Advance(600 * time.Millisecond)
	if completes != 1 {
		t.Fatalf("expected complete once, got %d", completes)
	}
	approxEqual(t, last.X, 10)
	approxEqual(t, last.Y, 20)
	if !tw.AllCompleted() {
		t.Fatal("expected tween to report completion")
	}
	if IsTweening(nil) {
		t.Fatal("nil target should not report tweening")
	}
}

func TestTweenDelay(t *testing.T) {
	resetManager()
	var starts, completes int
	tw := To(0, 10, 1).SetDelay(0.5)
	tw.OnStart(func(*GTweener) { starts++ })
	tw.OnComplete(func(*GTweener) { completes++ })

	Advance(400 * time.Millisecond)
	if starts != 0 {
		t.Fatalf("expected no start before delay, got %d", starts)
	}

	Advance(200 * time.Millisecond)
	if starts != 1 {
		t.Fatalf("expected start after delay, got %d", starts)
	}

	Advance(1 * time.Second)
	if completes != 1 {
		t.Fatalf("expected complete once, got %d", completes)
	}
}

func TestTweenRepeatYoyo(t *testing.T) {
	resetManager()
	var completes int
	tw := To(0, 100, 1).SetRepeat(1, true)
	tw.OnComplete(func(*GTweener) { completes++ })

	Advance(500 * time.Millisecond)
	value := tw.Value().X
	if value <= 0 || value >= 100 {
		t.Fatalf("expected intermediate value during first pass, got %.2f", value)
	}

	Advance(600 * time.Millisecond)
	if tw.Completed() {
		t.Fatalf("tween should not be completed halfway through yoyo (state=%d value=%.2f)", tw.endedState, tw.Value().X)
	}

	Advance(1 * time.Second)
	if completes != 1 {
		t.Fatalf("expected complete once after yoyo, got %d", completes)
	}
	approxEqual(t, tw.Value().X, 0)
	if !tw.AllCompleted() {
		t.Fatal("expected AllCompleted after yoyo finishes")
	}
}

func TestTweenSeek(t *testing.T) {
	resetManager()
	tw := To(0, 100, 2)
	tw.Seek(1)
	approxEqual(t, tw.NormalizedTime(), 0.75) // QuadOut at half time
	approxEqual(t, tw.Value().X, 75)
}

func TestGlobalKill(t *testing.T) {
	resetManager()
	type dummy struct{}
	obj := &dummy{}
	To(0, 10, 1).SetTarget(obj, "x")
	To(0, 20, 1).SetTarget(obj, "y")
	if !IsTweening(obj) {
		t.Fatal("expected tweening before kill")
	}
	if !Kill(obj, false) {
		t.Fatal("expected kill to return true")
	}
	if IsTweening(obj) {
		t.Fatal("expected no tweens after kill")
	}
}
