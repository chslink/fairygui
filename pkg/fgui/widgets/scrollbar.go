package widgets

import (
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// GScrollBar 代表 ScrollPane 使用的滚动条。
type GScrollBar struct {
	*core.GComponent
	packageItem *assets.PackageItem
	template    *core.GComponent

	grip      *core.GObject
	bar       *core.GObject
	arrow1    *core.GObject
	arrow2    *core.GObject
	fixedGrip bool

	target      *core.ScrollPane
	vertical    bool
	scrollPerc  float64
	displayPerc float64

	dragOffset laya.Point
	dragging   bool

	listenerID int

	stageMoveListener laya.Listener
	stageUpListener   laya.Listener
}

// ComponentRoot exposes the embedded component for helpers.
func (b *GScrollBar) ComponentRoot() *core.GComponent {
	if b == nil {
		return nil
	}
	return b.GComponent
}

// NewScrollBar 创建滚动条。
func NewScrollBar() *GScrollBar {
	comp := core.NewGComponent()
	bar := &GScrollBar{
		GComponent: comp,
	}
	comp.SetData(bar)
	return bar
}

// SetPackageItem 记录模板来源。
func (b *GScrollBar) SetPackageItem(item *assets.PackageItem) {
	b.packageItem = item
}

// PackageItem 返回模板。
func (b *GScrollBar) PackageItem() *assets.PackageItem {
	return b.packageItem
}

// SetTemplateComponent 挂载模板。
func (b *GScrollBar) SetTemplateComponent(comp *core.GComponent) {
	if b.template != nil && b.GComponent != nil {
		b.GComponent.RemoveChild(b.template.GObject)
	}
	b.template = comp
	if comp != nil && b.GComponent != nil {
		b.GComponent.AddChild(comp.GObject)
	}
	b.resolveTemplate()
	b.updateGrip()
}

// TemplateComponent 返回模板。
func (b *GScrollBar) TemplateComponent() *core.GComponent {
	return b.template
}

// SetFixedGrip 标记 Grip 是否固定尺寸。
func (b *GScrollBar) SetFixedGrip(fixed bool) {
	b.fixedGrip = fixed
	b.updateGrip()
}

// SetScrollPane 绑定目标 ScrollPane。
func (b *GScrollBar) SetScrollPane(pane *core.ScrollPane, vertical bool) {
	if b.target == pane && b.vertical == vertical {
		return
	}
	if b.target != nil {
		b.target.RemoveScrollListener(b.listenerID)
		b.listenerID = 0
	}
	b.target = pane
	b.vertical = vertical
	if pane != nil {
		b.listenerID = pane.AddScrollListener(func(info core.ScrollInfo) {
			b.SyncFromPane(info)
		})
	}
	b.updateGrip()
}

// SyncFromPane 根据 ScrollPane 状态刷新视图。
func (b *GScrollBar) SyncFromPane(info core.ScrollInfo) {
	if b.target == nil {
		return
	}
	if b.vertical {
		b.setScrollPerc(info.PercY)
		b.setDisplayPerc(info.DisplayPercY)
	} else {
		b.setScrollPerc(info.PercX)
		b.setDisplayPerc(info.DisplayPercX)
	}
}

func (b *GScrollBar) resolveTemplate() {
	if b.template == nil {
		return
	}
	if child := b.template.ChildByName("grip"); child != nil {
		b.setGrip(child)
	}
	if child := b.template.ChildByName("bar"); child != nil {
		b.bar = child
	}
	if child := b.template.ChildByName("arrow1"); child != nil {
		b.arrow1 = child
	}
	if child := b.template.ChildByName("arrow2"); child != nil {
		b.arrow2 = child
	}
	if b.template.GObject != nil {
		b.template.GObject.DisplayObject().Dispatcher().On(laya.EventMouseDown, b.onBarMouseDown)
	}
}

func (b *GScrollBar) setGrip(obj *core.GObject) {
	if b.grip != nil && b.grip.DisplayObject() != nil {
		b.grip.DisplayObject().Dispatcher().Off(laya.EventMouseDown, b.onGripMouseDown)
	}
	b.grip = obj
	if obj != nil && obj.DisplayObject() != nil {
		obj.DisplayObject().Dispatcher().On(laya.EventMouseDown, b.onGripMouseDown)
	}
}

func (b *GScrollBar) updateGrip() {
	if b.bar == nil || b.grip == nil {
		return
	}
	total := b.length()
	if total <= 0 {
		return
	}
	gripLength := total
	if !b.fixedGrip {
		gripLength = total * b.displayPerc
	}
	minSize := b.minSize()
	if gripLength < minSize {
		gripLength = minSize
	}
	if gripLength > total {
		gripLength = total
	}
	if b.vertical {
		b.grip.SetSize(b.grip.Width(), gripLength)
		offset := (total - gripLength) * b.scrollPerc
		b.grip.SetPosition(b.grip.X(), b.bar.Y()+offset)
	} else {
		b.grip.SetSize(gripLength, b.grip.Height())
		offset := (total - gripLength) * b.scrollPerc
		b.grip.SetPosition(b.bar.X()+offset, b.grip.Y())
	}
	b.grip.DisplayObject().SetVisible(b.displayPerc > 0 && b.displayPerc < 1)
}

func (b *GScrollBar) setScrollPerc(value float64) {
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}
	b.scrollPerc = value
	b.updateGrip()
}

func (b *GScrollBar) setDisplayPerc(value float64) {
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}
	b.displayPerc = value
	b.updateGrip()
}

func (b *GScrollBar) length() float64 {
	if b.bar == nil {
		return 0
	}
	if b.vertical {
		return b.bar.Height() - b.extraMargin()
	}
	return b.bar.Width() - b.extraMargin()
}

func (b *GScrollBar) extraMargin() float64 {
	total := 0.0
	if b.vertical {
		if b.arrow1 != nil {
			total += b.arrow1.Height()
		}
		if b.arrow2 != nil {
			total += b.arrow2.Height()
		}
	} else {
		if b.arrow1 != nil {
			total += b.arrow1.Width()
		}
		if b.arrow2 != nil {
			total += b.arrow2.Width()
		}
	}
	return total
}

func (b *GScrollBar) minSize() float64 {
	if b.vertical {
		return 10 + b.extraMargin()
	}
	return 10 + b.extraMargin()
}

func (b *GScrollBar) onGripMouseDown(evt laya.Event) {
	if b.grip == nil || b.target == nil {
		return
	}
	evt.Data = evt.Data // ensure not nil
	event, ok := evt.Data.(laya.PointerEvent)
	if !ok {
		return
	}
	b.dragging = true
	local := b.template.GObject.DisplayObject().GlobalToLocal(event.Position)
	if b.vertical {
		b.dragOffset = laya.Point{X: 0, Y: local.Y - b.grip.Y()}
	} else {
		b.dragOffset = laya.Point{X: local.X - b.grip.X(), Y: 0}
	}
	b.registerStageDrag()
}

func (b *GScrollBar) onStageMouseMove(evt laya.Event) {
	if !b.dragging || b.grip == nil || b.target == nil {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok {
		return
	}
	local := b.template.GObject.DisplayObject().GlobalToLocal(pe.Position)
	if b.vertical {
		track := b.length() - b.grip.Height()
		if track <= 0 {
			return
		}
		curY := local.Y - b.dragOffset.Y
		perc := (curY - b.bar.Y()) / track
		b.target.SetPercY(perc, false)
	} else {
		track := b.length() - b.grip.Width()
		if track <= 0 {
			return
		}
		curX := local.X - b.dragOffset.X
		perc := (curX - b.bar.X()) / track
		b.target.SetPercX(perc, false)
	}
}

func (b *GScrollBar) onStageMouseUp(evt laya.Event) {
	if !b.dragging {
		return
	}
	b.dragging = false
	b.unregisterStageDrag()
}

func (b *GScrollBar) onBarMouseDown(evt laya.Event) {
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok || b.target == nil || b.bar == nil {
		return
	}
	local := b.template.GObject.DisplayObject().GlobalToLocal(pe.Position)
	if b.vertical {
		if local.Y < b.grip.Y() {
			b.target.ScrollUp()
		} else {
			b.target.ScrollDown()
		}
	} else {
		if local.X < b.grip.X() {
			b.target.ScrollLeft()
		} else {
			b.target.ScrollRight()
		}
	}
}

func (b *GScrollBar) registerStageDrag() {
	root := core.Root()
	if root == nil {
		return
	}
	stage := root.Stage()
	if stage == nil {
		return
	}
	dispatcher := stage.Root().Dispatcher()
	if b.stageMoveListener == nil {
		b.stageMoveListener = func(evt laya.Event) {
			b.onStageMouseMove(evt)
		}
	}
	if b.stageUpListener == nil {
		b.stageUpListener = func(evt laya.Event) {
			b.onStageMouseUp(evt)
		}
	}
	dispatcher.On(laya.EventMouseMove, b.stageMoveListener)
	dispatcher.On(laya.EventStageMouseUp, b.stageUpListener)
}

func (b *GScrollBar) unregisterStageDrag() {
	root := core.Root()
	if root == nil {
		return
	}
	stage := root.Stage()
	if stage == nil {
		return
	}
	dispatcher := stage.Root().Dispatcher()
	if b.stageMoveListener != nil {
		dispatcher.Off(laya.EventMouseMove, b.stageMoveListener)
	}
	if b.stageUpListener != nil {
		dispatcher.Off(laya.EventStageMouseUp, b.stageUpListener)
	}
}
