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
	s.Every(time.Millisecond*10, func() {
		count++
	})
	for i := 0; i < 5; i++ {
		s.Advance(time.Millisecond * 10)
	}
	if count != 5 {
		t.Fatalf("expected 5 triggers, got %d", count)
	}
}
