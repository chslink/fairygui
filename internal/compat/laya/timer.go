package laya

import "time"

// Scheduler implements a frame-based timer similar to Laya.timer.
type Scheduler struct {
	tasks []*scheduledTask
	now   time.Duration
}

type scheduledTask struct {
	delay   time.Duration
	remain  time.Duration
	repeat  bool
	handler func()
}

// NewScheduler creates an empty scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// Advance advances the scheduler by the supplied delta time.
func (s *Scheduler) Advance(delta time.Duration) {
	if delta <= 0 {
		return
	}
	s.now += delta
	pending := s.tasks[:0]
	for _, task := range s.tasks {
		task.remain -= delta
		if task.remain <= 0 {
			if task.handler != nil {
				task.handler()
			}
			if task.repeat {
				task.remain = task.delay
				pending = append(pending, task)
			}
			continue
		}
		pending = append(pending, task)
	}
	s.tasks = pending
}

// After schedules a one-shot callback after the given delay.
func (s *Scheduler) After(delay time.Duration, fn func()) {
	if fn == nil || delay < 0 {
		return
	}
	s.tasks = append(s.tasks, &scheduledTask{
		delay:   delay,
		remain:  delay,
		repeat:  false,
		handler: fn,
	})
}

// Every schedules a repeating callback with the provided period.
func (s *Scheduler) Every(period time.Duration, fn func()) {
	if fn == nil || period <= 0 {
		return
	}
	s.tasks = append(s.tasks, &scheduledTask{
		delay:   period,
		remain:  period,
		repeat:  true,
		handler: fn,
	})
}
