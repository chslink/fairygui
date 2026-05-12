package fgui

import "github.com/chslink/fairygui/internal/compat/laya"

// CancelFunc unregisters an event listener when called.
type CancelFunc func()

// ListenClick registers a click listener and returns a cancel function.
func ListenClick(obj *GObject, fn func()) CancelFunc {
	id := obj.OnClick(fn)
	return func() { obj.OffClick(id) }
}

// ListenClickWithData registers a click listener with event data.
func ListenClickWithData(obj *GObject, fn func(*laya.Event)) CancelFunc {
	id := obj.OnClickWithData(fn)
	return func() { obj.OffClick(id) }
}

// ListenLink registers a link-click listener.
func ListenLink(obj *GObject, fn func(link string)) CancelFunc {
	id := obj.OnLink(fn)
	return func() { obj.OffLink(id) }
}

// ListenDrop registers a drop listener.
func ListenDrop(obj *GObject, fn func(data any)) CancelFunc {
	id := obj.OnWithID(laya.EventDrop, func(evt *laya.Event) {
		fn(evt.Data)
	})
	return func() { obj.OffByID(laya.EventDrop, id) }
}

// ListenStateChanged registers a state-change listener.
func ListenStateChanged(obj *GObject, fn func(*laya.Event)) CancelFunc {
	id := obj.OnStateChanged(fn)
	return func() { obj.OffStateChanged(id) }
}
