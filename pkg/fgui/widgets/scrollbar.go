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

// ResolveChildren 手动触发子组件解析
// 用于 BuildComponent 中直接构建子组件的情况（不通过 SetTemplateComponent）
func (b *GScrollBar) ResolveChildren() {
	if b == nil {
		return
	}
	b.resolveTemplate()
	b.updateGrip()
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
	// 优先从 template 查找，如果没有 template 则从 GComponent 本身查找
	// 这支持两种模式：
	// 1. 通过 SetTemplateComponent 设置的独立模板（TypeScript 模式）
	// 2. 直接在 GComponent 中构建的子组件（Go BuildComponent 模式）
	searchRoot := b.template
	if searchRoot == nil {
		searchRoot = b.GComponent
	}
	if searchRoot == nil {
		return
	}

	if child := searchRoot.ChildByName("grip"); child != nil {
		b.setGrip(child)
	}
	if child := searchRoot.ChildByName("bar"); child != nil {
		b.bar = child
	}
	if child := searchRoot.ChildByName("arrow1"); child != nil {
		b.arrow1 = child
	}
	if child := searchRoot.ChildByName("arrow2"); child != nil {
		b.arrow2 = child
	}
	if searchRoot.GObject != nil {
		searchRoot.GObject.DisplayObject().Dispatcher().On(laya.EventMouseDown, b.onBarMouseDown)
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
	// 如果还没有绑定到 ScrollPane，跳过更新
	// 此时 displayPerc 还是默认值 0，无法计算正确的 gripLength
	// 等到 SetScrollPane() 调用 AddScrollListener() 后会再次触发 updateGrip()
	if b.target == nil {
		return
	}

	if b.bar == nil || b.grip == nil {
		return
	}

	// 与TypeScript版本保持一致：直接使用bar的高度/宽度
	// 不需要额外计算length()和extraMargin()
	var total float64
	if b.vertical {
		total = b.bar.Height()
	} else {
		total = b.bar.Width()
	}
	if total <= 0 {
		return
	}

	// 计算滑块长度
	gripLength := total
	if !b.fixedGrip {
		gripLength = total * b.displayPerc
	}

	// 应用最小尺寸限制（TypeScript版本中minSize用于限制gripLength）
	minSize := b.minSize()
	if gripLength < minSize {
		gripLength = minSize
	}
	if gripLength > total {
		gripLength = total
	}

	// 设置滑块尺寸和位置（与TypeScript版本逻辑一致）
	if b.vertical {
		b.grip.SetSize(b.grip.Width(), gripLength)
		// TypeScript: this._grip.y = this._bar.y + (this._bar.height - this._grip.height) * this._scrollPerc
		offset := (total - gripLength) * b.scrollPerc
		gripY := b.bar.Y() + offset
		b.grip.SetPosition(b.grip.X(), gripY)
	} else {
		b.grip.SetSize(gripLength, b.grip.Height())
		// TypeScript: this._grip.x = this._bar.x + (this._bar.width - this._grip.width) * this._scrollPerc
		offset := (total - gripLength) * b.scrollPerc
		gripX := b.bar.X() + offset
		b.grip.SetPosition(gripX, b.grip.Y())
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
	// 与TypeScript版本一致：minSize返回arrow按钮的高度/宽度总和
	// 注意：TypeScript版本没有额外的+10
	return b.extraMargin()
}

// getContainerDisplayObject 返回用于坐标转换的 DisplayObject
// 如果 template 存在则使用 template，否则使用 GComponent 本身
func (b *GScrollBar) getContainerDisplayObject() *laya.Sprite {
	if b.template != nil && b.template.GObject != nil {
		return b.template.GObject.DisplayObject()
	}
	if b.GComponent != nil && b.GComponent.GObject != nil {
		return b.GComponent.GObject.DisplayObject()
	}
	return nil
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
	// 与TypeScript版本一致：使用ScrollBar的DisplayObject进行坐标转换
	// TypeScript: this.globalToLocal(Laya.stage.mouseX, Laya.stage.mouseY, this._dragOffset)
	display := b.GComponent.GObject.DisplayObject()
	if display == nil {
		return
	}
	local := display.GlobalToLocal(event.Position)
	// 计算鼠标相对于滑块左上角的偏移量
	// 公式：dragOffset = 鼠标在ScrollBar中的位置 - 滑块在ScrollBar中的位置
	b.dragOffset = laya.Point{X: local.X - b.grip.X(), Y: local.Y - b.grip.Y()}
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
	// 与TypeScript版本一致：使用ScrollBar的DisplayObject进行坐标转换
	display := b.GComponent.GObject.DisplayObject()
	if display == nil {
		return
	}
	local := display.GlobalToLocal(pe.Position)
	if b.vertical {
		// 与TypeScript版本一致：直接使用bar的高度
		// TypeScript: var track: number = this._bar.height - this._grip.height;
		track := b.bar.Height() - b.grip.Height()
		if track <= 0 {
			return
		}
		curY := local.Y - b.dragOffset.Y
		perc := (curY - b.bar.Y()) / track
		b.target.SetPercY(perc, false)
	} else {
		// TypeScript: track = this._bar.width - this._grip.width
		track := b.bar.Width() - b.grip.Width()
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
	display := b.getContainerDisplayObject()
	if display == nil {
		return
	}
	local := display.GlobalToLocal(pe.Position)
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
