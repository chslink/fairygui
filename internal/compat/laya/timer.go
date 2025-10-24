package laya

import "time"

// TimerHandle identifies a scheduled timer task.
type TimerHandle struct {
	id uint64
}

// Scheduler implements a frame-based timer similar to Laya.timer.
type Scheduler struct {
	tasks []*scheduledTask
	byID  map[uint64]*scheduledTask
	now   time.Duration
	seq   uint64
}

type scheduledTask struct {
	id              uint64
	delay           time.Duration
	remain          time.Duration
	repeat          bool
	frame           bool
	frameInterval   int
	framesRemaining int
	handler         func()
	cancelled       bool
}

// NewScheduler creates an empty scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		byID: make(map[uint64]*scheduledTask),
	}
}

// Advance advances the scheduler by the supplied delta time and processes frame-based tasks.
func (s *Scheduler) Advance(delta time.Duration) {
	if s == nil {
		return
	}
	if delta < 0 {
		delta = 0
	}
	s.now += delta
	if len(s.tasks) == 0 {
		return
	}
	pending := s.tasks[:0]
	for _, task := range s.tasks {
		if task == nil || task.handler == nil || task.cancelled {
			delete(s.byID, task.id)
			continue
		}
		if task.frame {
			if task.framesRemaining <= 0 {
				task.framesRemaining = 1
			}
			task.framesRemaining--
			if task.framesRemaining <= 0 {
				handler := task.handler
				handler()
				if task.repeat && !task.cancelled && task.handler != nil {
					task.framesRemaining = task.frameInterval
					if task.framesRemaining <= 0 {
						task.framesRemaining = 1
					}
					pending = append(pending, task)
					continue
				}
				delete(s.byID, task.id)
				continue
			}
			pending = append(pending, task)
			continue
		}

		if delta > 0 {
			task.remain -= delta
		}
		if task.remain <= 0 {
			handler := task.handler
			handler()
			if task.repeat && !task.cancelled && task.handler != nil {
				task.remain = task.delay
				if task.remain <= 0 {
					task.remain = task.delay
				}
				pending = append(pending, task)
				continue
			}
			delete(s.byID, task.id)
			continue
		}
		pending = append(pending, task)
	}
	s.tasks = pending
}

// After schedules a one-shot callback after the given delay.
func (s *Scheduler) After(delay time.Duration, fn func()) TimerHandle {
	return s.scheduleTime(delay, fn, false)
}

// Every schedules a repeating callback with the provided period.
func (s *Scheduler) Every(period time.Duration, fn func()) TimerHandle {
	return s.scheduleTime(period, fn, true)
}

// Once registers a one-shot timer using time duration.
func (s *Scheduler) Once(delay time.Duration, fn func()) TimerHandle {
	return s.scheduleTime(delay, fn, false)
}

// Loop registers a repeating timer using time duration.
func (s *Scheduler) Loop(period time.Duration, fn func()) TimerHandle {
	return s.scheduleTime(period, fn, true)
}

// CallLater schedules a callback for the next scheduler tick.
func (s *Scheduler) CallLater(fn func()) TimerHandle {
	return s.FrameOnce(1, fn)
}

// FrameOnce schedules a callback to run after the specified number of frames.
func (s *Scheduler) FrameOnce(frames int, fn func()) TimerHandle {
	if s == nil || fn == nil {
		return TimerHandle{}
	}
	if frames <= 0 {
		frames = 1
	}
	return s.addTask(&scheduledTask{
		frame:           true,
		frameInterval:   frames,
		framesRemaining: frames,
		repeat:          false,
		handler:         fn,
	})
}

// FrameLoop schedules a callback to run every N frames.
func (s *Scheduler) FrameLoop(frames int, fn func()) TimerHandle {
	if s == nil || fn == nil {
		return TimerHandle{}
	}
	if frames <= 0 {
		frames = 1
	}
	return s.addTask(&scheduledTask{
		frame:           true,
		frameInterval:   frames,
		framesRemaining: frames,
		repeat:          true,
		handler:         fn,
	})
}

// Cancel stops the task associated with the provided handle.
func (s *Scheduler) Cancel(handle TimerHandle) {
	if s == nil || handle.id == 0 {
		return
	}
	if task, ok := s.byID[handle.id]; ok {
		task.cancelled = true
		task.handler = nil
		delete(s.byID, handle.id)
	}
}

// CancelAll stops every scheduled task.
func (s *Scheduler) CancelAll() {
	if s == nil {
		return
	}
	for id, task := range s.byID {
		if task != nil {
			task.cancelled = true
			task.handler = nil
		}
		delete(s.byID, id)
	}
	s.tasks = s.tasks[:0]
}

func (s *Scheduler) scheduleTime(delay time.Duration, fn func(), repeat bool) TimerHandle {
	if s == nil || fn == nil {
		return TimerHandle{}
	}
	if delay < 0 {
		delay = 0
	}
	if delay == 0 && !repeat {
		return s.CallLater(fn)
	}
	return s.addTask(&scheduledTask{
		delay:   delay,
		remain:  delay,
		repeat:  repeat,
		handler: fn,
	})
}

func (s *Scheduler) addTask(task *scheduledTask) TimerHandle {
	if task == nil {
		return TimerHandle{}
	}
	s.seq++
	task.id = s.seq
	if task.frameInterval <= 0 {
		task.frameInterval = 1
	}
	if task.frame && task.framesRemaining <= 0 {
		task.framesRemaining = task.frameInterval
	}
	s.byID[task.id] = task
	s.tasks = append(s.tasks, task)
	return TimerHandle{id: task.id}
}
