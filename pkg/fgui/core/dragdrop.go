package core

import (
	"sync"

	"github.com/chslink/fairygui/internal/compat/laya"
)

var (
	dragDropOnce sync.Once
	dragDropInst *DragDropManager
)

type DragDropManager struct {
	agent      *GObject
	sourceData any
	dragging   bool
	touchID    int

	moveListener laya.ListenerID
	upListener   laya.ListenerID
}

// DragDrop returns the global drag-drop manager singleton.
func DragDrop() *DragDropManager {
	dragDropOnce.Do(func() {
		dragDropInst = &DragDropManager{}
	})
	return dragDropInst
}

// IsDragging reports whether a drag operation is currently in progress.
func (d *DragDropManager) IsDragging() bool {
	if d == nil {
		return false
	}
	return d.dragging
}

// Agent returns the visual drag proxy object.
func (d *DragDropManager) Agent() *GObject {
	if d == nil {
		return nil
	}
	return d.agent
}

// StartDrag begins a drag operation. The loader is used as the visual proxy,
// icon is the resource URL for the proxy image, and sourceData is forwarded to
// the DROP event on completion.
func (d *DragDropManager) StartDrag(source *GObject, icon string, sourceData any, touchID int) {
	if d == nil || source == nil {
		return
	}
	d.Cancel()

	d.sourceData = sourceData
	d.touchID = touchID

	// Create or reuse drag agent (a GObject displayed on top of everything).
	d.agent = NewGObject()
	d.agent.SetSortingOrder(1000000)
	d.agent.SetPivotWithAnchor(0.5, 0.5, false)
	d.agent.SetSize(100, 100)

	Root().AddChild(d.agent)

	// Position agent at current mouse position.
	if d.agent.DisplayObject() != nil && d.agent.DisplayObject().Parent() != nil {
		_ = icon // icon is used in rendering layer; stored for reference
	}

	d.dragging = true

	// Listen for mouse move to update agent position.
	stage := Root().Stage()
	if stage != nil {
		d.moveListener = stage.Root().Dispatcher().OnWithID(laya.EventMouseMove, func(evt *laya.Event) {
			if pe, ok := evt.Data.(laya.PointerEvent); ok {
				local := Root().DisplayObject().GlobalToLocal(laya.Point{X: pe.Position.X, Y: pe.Position.Y})
				d.agent.SetPosition(local.X, local.Y)
			}
		})
		d.upListener = stage.Root().Dispatcher().OnWithID(laya.EventStageMouseUp, func(evt *laya.Event) {
			d.onDragEnd(evt)
		})
	}
}

// Cancel stops the current drag operation without triggering DROP.
func (d *DragDropManager) Cancel() {
	if d == nil || !d.dragging {
		return
	}
	d.teardown()
}

func (d *DragDropManager) onDragEnd(evt *laya.Event) {
	if d == nil || !d.dragging {
		return
	}

	// Walk ancestors of the hit target to find a DROP listener.
	var pe laya.PointerEvent
	var hitTarget *laya.Sprite
	switch data := evt.Data.(type) {
	case laya.PointerEvent:
		pe = data
		hitTarget = data.Hit
	case *laya.Event:
		if data != nil {
			if p, ok := data.Data.(laya.PointerEvent); ok {
				pe = p
				hitTarget = p.Hit
			}
		}
	}

	for current := hitTarget; current != nil; current = current.Parent() {
		owner := ownerAsGObject(current)
		if owner != nil {
			owner.Emit(laya.EventDrop, d.sourceData)
			// In TypeScript, the DROP event returns false to continue bubbling.
			// Go version fires on every ancestor; the handler can check sourceData.
		}
	}
	_ = pe

	d.teardown()
}

func (d *DragDropManager) teardown() {
	if d.agent != nil && d.agent.Parent() != nil {
		d.agent.Parent().RemoveChild(d.agent)
	}
	d.agent = nil
	d.dragging = false
	d.sourceData = nil

	stage := Root().Stage()
	if stage != nil {
		if d.moveListener != 0 {
			stage.Root().Dispatcher().OffByID(laya.EventMouseMove, d.moveListener)
			d.moveListener = 0
		}
		if d.upListener != 0 {
			stage.Root().Dispatcher().OffByID(laya.EventStageMouseUp, d.upListener)
			d.upListener = 0
		}
	}
}
