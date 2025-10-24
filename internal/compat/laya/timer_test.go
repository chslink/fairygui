package laya_test

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
)

func TestSchedulerAfter(t *testing.T) {
	s := laya.NewScheduler()
	count := 0
	s.After(time.Millisecond*10, func() {
		count++
	})
	s.Advance(time.Millisecond * 5)
	if count != 0 {
		t.Fatalf("expected no trigger yet, got %d", count)
	}
	s.Advance(time.Millisecond * 5)
	if count != 1 {
		t.Fatalf("expected single trigger, got %d", count)
	}
	s.Advance(time.Millisecond * 10)
	if count != 1 {
		t.Fatalf("after should not repeat, got %d", count)
	}
}

func TestSchedulerEvery(t *testing.T) {
	s := laya.NewScheduler()
	count := 0
	h := s.Every(time.Millisecond*10, func() {
		count++
	})
	for i := 0; i < 5; i++ {
		s.Advance(time.Millisecond * 10)
	}
	if count != 5 {
		t.Fatalf("expected 5 triggers, got %d", count)
	}
	s.Cancel(h)
	s.Advance(time.Millisecond * 10)
	if count != 5 {
		t.Fatalf("expected cancel to stop repetitions, got %d", count)
	}
}

func TestSchedulerCallLater(t *testing.T) {
	s := laya.NewScheduler()
	count := 0
	s.CallLater(func() {
		count++
	})
	s.Advance(time.Millisecond)
	if count != 1 {
		t.Fatalf("expected callLater to execute once, got %d", count)
	}
	s.Advance(time.Millisecond)
	if count != 1 {
		t.Fatalf("callLater should not repeat, got %d", count)
	}
}

func TestSchedulerFrameLoop(t *testing.T) {
	s := laya.NewScheduler()
	count := 0
	s.FrameLoop(2, func() {
		count++
	})
	for i := 0; i < 6; i++ {
		s.Advance(time.Millisecond)
	}
	if count != 3 {
		t.Fatalf("expected frame loop every 2 frames to run 3 times, got %d", count)
	}
}
