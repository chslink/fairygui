package testutil

import "github.com/chslink/fairygui/internal/compat/laya"

// EventRecord captures a dispatched event for assertions.
type EventRecord struct {
	Source *laya.Sprite
	Type   laya.EventType
	Data   any
}

// EventLog collects events fired by one or more sprites.
type EventLog struct {
	Records []EventRecord
}

// AttachEventLog registers listeners on the provided sprite and records events of the given types.
func AttachEventLog(log *EventLog, sprite *laya.Sprite, events ...laya.EventType) {
	dispatcher := sprite.Dispatcher()
	for _, evt := range events {
		func(e laya.EventType) {
			dispatcher.On(e, func(ev *laya.Event) {
				log.Records = append(log.Records, EventRecord{
					Source: sprite,
					Type:   e,
					Data:   ev.Data,
				})
			})
		}(evt)
	}
}
