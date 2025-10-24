package laya

import "reflect"

// EventType identifies an event emitted by the compatibility layer.
type EventType string

// Common event names referenced by FairyGUI.
const (
	EventAdded          EventType = "added"
	EventRemoved        EventType = "removed"
	EventDisplay        EventType = "display"
	EventUndisplay      EventType = "undisplay"
	EventMouseDown      EventType = "mousedown"
	EventMouseUp        EventType = "mouseup"
	EventMouseMove      EventType = "mousemove"
	EventClick          EventType = "click"
	EventMouseWheel     EventType = "mousewheel"
	EventRollOver       EventType = "rollover"
	EventRollOut        EventType = "rollout"
	EventTouchBegin     EventType = "touchBegin"
	EventTouchEnd       EventType = "touchEnd"
	EventScroll         EventType = "scroll"
	EventDragStart      EventType = "dragStart"
	EventDragMove       EventType = "dragMove"
	EventDragEnd        EventType = "dragEnd"
	EventTouchMove      EventType = "touchMove"
	EventTouchCancel    EventType = "touchCancel"
	EventXYChanged      EventType = "xyChanged"
	EventSizeChanged    EventType = "sizeChanged"
	EventStateChanged   EventType = "stateChanged"
	EventStageMouseDown EventType = "stageMouseDown"
	EventStageMouseUp   EventType = "stageMouseUp"
	EventKeyDown        EventType = "keyDown"
	EventKeyUp          EventType = "keyUp"
	EventKeyPress       EventType = "keyPress"
	EventFocusIn        EventType = "focusIn"
	EventFocusOut       EventType = "focusOut"
	EventLink           EventType = "link"
)

// Event carries the runtime payload for a dispatched event.
type Event struct {
	Type EventType
	Data any
}

// Listener reacts to an emitted event.
type Listener func(Event)

// EventDispatcher dispatches named events to registered listeners.
type EventDispatcher interface {
	On(evt EventType, fn Listener)
	Once(evt EventType, fn Listener)
	Off(evt EventType, fn Listener)
	Emit(evt EventType, data any)
}

type listenerEntry struct {
	fn   Listener
	key  uintptr
	once bool
}

// BasicEventDispatcher is a simple in-memory dispatcher. It is not thread safe.
type BasicEventDispatcher struct {
	listeners map[EventType][]listenerEntry
}

// NewEventDispatcher returns an empty dispatcher instance.
func NewEventDispatcher() *BasicEventDispatcher {
	return &BasicEventDispatcher{
		listeners: make(map[EventType][]listenerEntry),
	}
}

// On registers a listener for the given event type.
func (d *BasicEventDispatcher) On(evt EventType, fn Listener) {
	d.addListener(evt, fn, false)
}

// Once registers a listener that will be removed after it runs once.
func (d *BasicEventDispatcher) Once(evt EventType, fn Listener) {
	d.addListener(evt, fn, true)
}

func (d *BasicEventDispatcher) addListener(evt EventType, fn Listener, once bool) {
	if fn == nil {
		return
	}
	entry := listenerEntry{
		fn:   fn,
		key:  reflect.ValueOf(fn).Pointer(),
		once: once,
	}
	d.listeners[evt] = append(d.listeners[evt], entry)
}

// Off removes a listener from the event list.
func (d *BasicEventDispatcher) Off(evt EventType, fn Listener) {
	if fn == nil {
		return
	}
	key := reflect.ValueOf(fn).Pointer()
	list := d.listeners[evt]
	out := list[:0]
	for _, entry := range list {
		if entry.key != key {
			out = append(out, entry)
		}
	}
	if len(out) == 0 {
		delete(d.listeners, evt)
	} else {
		d.listeners[evt] = out
	}
}

// Emit dispatches an event to registered listeners.
func (d *BasicEventDispatcher) Emit(evt EventType, data any) {
	list := d.listeners[evt]
	if len(list) == 0 {
		return
	}
	event := Event{Type: evt, Data: data}
	remaining := list[:0]
	for _, entry := range list {
		if entry.fn != nil {
			entry.fn(event)
		}
		if !entry.once {
			remaining = append(remaining, entry)
		}
	}
	if len(remaining) == 0 {
		delete(d.listeners, evt)
	} else {
		d.listeners[evt] = remaining
	}
}
