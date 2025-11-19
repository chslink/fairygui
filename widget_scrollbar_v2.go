package fairygui

import (
	"sync"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// ScrollInfo - 滚动信息
type ScrollInfo struct {
	PercX        float64
	PercY        float64
	DisplayPercX float64
	DisplayPercY float64
}

// ScrollListener - 滚动监听器
type ScrollListener func(ScrollInfo)

// ============================================================================
// ScrollBar - 滚动条控件 V2
// ============================================================================

type ScrollBar struct {
	*ComponentImpl

	// 资源相关
	packageItem      *PackageItemWrapper
	template         *ComponentImpl

	// 子对象
	grip      DisplayObject
	bar       DisplayObject
	arrow1    DisplayObject
	arrow2    DisplayObject
	fixedGrip bool

	// 滚动目标
	target      *ScrollPaneV2
	vertical    bool
	scrollPerc  float64
	displayPerc float64

	// 拖拽状态
	dragOffset Point
	dragging   bool

	// 事件监听器
	stageMoveHandler  EventHandler
	stageUpHandler    EventHandler
	stageMoveListener laya.Listener
	stageUpListener   laya.Listener

	// 内部状态
	dragMutex sync.Mutex
}

// NewScrollBar 创建新的滚动条
func NewScrollBar() *ScrollBar {
	sb := &ScrollBar{
		ComponentImpl: NewComponent(),
		scrollPerc:    0,
		displayPerc:   1,
		fixedGrip:     false,
		dragging:      false,
	}

	return sb
}

// ============================================================================
// 资源相关
// ============================================================================

// SetPackageItem 设置资源项
func (sb *ScrollBar) SetPackageItem(item PackageItem) {
	if item == nil {
		sb.packageItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		sb.packageItem = wrapper
	}
}

// PackageItem 返回资源项
func (sb *ScrollBar) PackageItem() PackageItem {
	return sb.packageItem
}

// SetTemplateComponent 设置模板组件
func (sb *ScrollBar) SetTemplateComponent(comp *ComponentImpl) {
	if sb.template != nil {
		sb.RemoveChild(sb.template)
	}
	sb.template = comp
	if comp != nil {
		comp.SetPosition(0, 0)
		sb.AddChild(comp)
	}
	sb.resolveTemplate()
	sb.updateGrip()
}

// TemplateComponent 返回模板组件
func (sb *ScrollBar) TemplateComponent() *ComponentImpl {
	return sb.template
}

// ResolveChildren 手动触发子组件解析
func (sb *ScrollBar) ResolveChildren() {
	sb.resolveTemplate()
	sb.updateGrip()
}

// SetFixedGrip 设置滑块是否固定尺寸
func (sb *ScrollBar) SetFixedGrip(fixed bool) {
	sb.fixedGrip = fixed
	sb.updateGrip()
}

// ============================================================================
// 滚动目标管理
// ============================================================================

// SetScrollPane 绑定滚动目标
func (sb *ScrollBar) SetScrollPane(pane *ScrollPaneV2, vertical bool) {
	if sb.target == pane && sb.vertical == vertical {
		return
	}

	sb.target = pane
	sb.vertical = vertical

	if pane != nil {
		pane.AddScrollListener(func(info ScrollInfo) {
			sb.SyncFromPane(info)
		})
	}

	sb.updateGrip()
}

// SyncFromPane 根据 ScrollPane 状态刷新
func (sb *ScrollBar) SyncFromPane(info ScrollInfo) {
	if sb.target == nil {
		return
	}

	if sb.vertical {
		sb.setScrollPerc(info.PercY)
		sb.setDisplayPerc(info.DisplayPercY)
	} else {
		sb.setScrollPerc(info.PercX)
		sb.setDisplayPerc(info.DisplayPercX)
	}
}

// ============================================================================
// 组件解析
// ============================================================================

func (sb *ScrollBar) resolveTemplate() {
	if sb.template == nil {
		return
	}

	// 查找 grip
	if child := sb.template.GetChildByName("grip"); child != nil {
		sb.setGrip(child)
	}

	// 查找 bar
	if child := sb.template.GetChildByName("bar"); child != nil {
		sb.bar = child
	}

	// 查找 arrow1
	if child := sb.template.GetChildByName("arrow1"); child != nil {
		sb.arrow1 = child
		if btn, ok := child.(*Button); ok {
			btn.OnClick(func() {
				if sb.target == nil {
					return
				}
				if sb.vertical {
					sb.target.ScrollUp()
				} else {
					sb.target.ScrollLeft()
				}
			})
		}
	}

	// 查找 arrow2
	if child := sb.template.GetChildByName("arrow2"); child != nil {
		sb.arrow2 = child
		if btn, ok := child.(*Button); ok {
			btn.OnClick(func() {
				if sb.target == nil {
					return
				}
				if sb.vertical {
					sb.target.ScrollDown()
				} else {
					sb.target.ScrollRight()
				}
			})
		}
	}

	// 绑定 bar 点击
	if sb.bar != nil {
		if comp, ok := sb.bar.(*ComponentImpl); ok {
			comp.On("mousedown", func(event Event) {
				if mouseEvent, ok := event.(*MouseEvent); ok && sb.target != nil {
					sb.onBarMouseDown(Point{X: mouseEvent.LocalX, Y: mouseEvent.LocalY})
				}
			})
		}
	}
}

func (sb *ScrollBar) setGrip(obj DisplayObject) {
	sb.grip = obj

	if sb.grip != nil {
		if comp, ok := sb.grip.(*ComponentImpl); ok {
			comp.On("mousedown", func(event Event) {
				event.StopPropagation()
				if mouseEvent, ok := event.(*MouseEvent); ok {
					sb.onGripMouseDown(Point{X: mouseEvent.LocalX, Y: mouseEvent.LocalY})
				}
			})
		}
	}
}

func (sb *ScrollBar) setScrollPerc(value float64) {
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}

	sb.scrollPerc = value
	sb.updateGrip()
}

func (sb *ScrollBar) setDisplayPerc(value float64) {
	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}

	sb.displayPerc = value
	sb.updateGrip()
}

func (sb *ScrollBar) minSize() float64 {
	return sb.extraMargin()
}

func (sb *ScrollBar) extraMargin() float64 {
	var total float64
	if sb.vertical {
		if sb.arrow1 != nil {
			if sizable, ok := sb.arrow1.(interface{ Height() float64 }); ok {
				total += sizable.Height()
			}
		}
		if sb.arrow2 != nil {
			if sizable, ok := sb.arrow2.(interface{ Height() float64 }); ok {
				total += sizable.Height()
			}
		}
	} else {
		if sb.arrow1 != nil {
			if sizable, ok := sb.arrow1.(interface{ Width() float64 }); ok {
				total += sizable.Width()
			}
		}
		if sb.arrow2 != nil {
			if sizable, ok := sb.arrow2.(interface{ Width() float64 }); ok {
				total += sizable.Width()
			}
		}
	}
	return total
}

// ============================================================================
// 事件处理
// ============================================================================

func (sb *ScrollBar) onGripMouseDown(position Point) {
	if sb.grip == nil || sb.target == nil {
		return
	}

	sb.dragMutex.Lock()
	defer sb.dragMutex.Unlock()

	sb.dragging = true

	// 计算鼠标相对于滑块左上角的偏移量
	localX, localY := sb.GlobalToLocal(position.X, position.Y)

	// 使用断言获取坐标方法
	var gripX, gripY float64
	if pos, ok := sb.grip.(interface{ X() float64 }); ok {
		gripX = pos.X()
	}
	if pos, ok := sb.grip.(interface{ Y() float64 }); ok {
		gripY = pos.Y()
	}

	sb.dragOffset = Point{X: localX - gripX, Y: localY - gripY}

	sb.registerStageDrag()
}

func (sb *ScrollBar) updateGrip() {
	sb.dragMutex.Lock()
	defer sb.dragMutex.Unlock()

	if sb.target == nil {
		return
	}

	if sb.bar == nil || sb.grip == nil {
		return
	}

	// 获取 bar 的尺寸
	var total float64
	if sb.vertical {
		if sizable, ok := sb.bar.(interface{ Height() float64 }); ok {
			total = sizable.Height()
		}
	} else {
		if sizable, ok := sb.bar.(interface{ Width() float64 }); ok {
			total = sizable.Width()
		}
	}

	if total <= 0 {
		return
	}

	// 计算滑块长度
	gripLength := total
	if !sb.fixedGrip {
		gripLength = total * sb.displayPerc
	}

	// 应用最小尺寸限制
	minSize := sb.minSize()
	if gripLength < minSize {
		gripLength = minSize
	}
	if gripLength > total {
		gripLength = total
	}

	// 设置滑块尺寸和位置
	if sb.vertical {
		// 设置高度
		var currentWidth float64
		if sizable, ok := sb.grip.(interface{ Width() float64 }); ok {
			currentWidth = sizable.Width()
		}
		if sizable, ok := sb.grip.(Sizable); ok {
			sizable.SetSize(currentWidth, gripLength)
		}

		offset := (total - gripLength) * sb.scrollPerc

		// 获取 bar 的 Y 坐标
		var barY float64
		if pos, ok := sb.bar.(interface{ Y() float64 }); ok {
			barY = pos.Y()
		}

		// 获取 grip 的 X 坐标
		var gripX float64
		if pos, ok := sb.grip.(interface{ X() float64 }); ok {
			gripX = pos.X()
		}

		gripY := barY + offset
		if pos, ok := sb.grip.(Positionable); ok {
			pos.SetPosition(gripX, gripY)
		}
	} else {
		// 水平方向
		var currentHeight float64
		if sizable, ok := sb.grip.(interface{ Height() float64 }); ok {
			currentHeight = sizable.Height()
		}
		if sizable, ok := sb.grip.(Sizable); ok {
			sizable.SetSize(gripLength, currentHeight)
		}

		offset := (total - gripLength) * sb.scrollPerc

		// 获取 bar 的 X 坐标
		var barX float64
		if pos, ok := sb.bar.(interface{ X() float64 }); ok {
			barX = pos.X()
		}

		// 获取 grip 的 Y 坐标
		var gripY float64
		if pos, ok := sb.grip.(interface{ Y() float64 }); ok {
			gripY = pos.Y()
		}

		gripX := barX + offset
		if pos, ok := sb.grip.(Positionable); ok {
			pos.SetPosition(gripX, gripY)
		}
	}

	sb.grip.SetVisible(sb.displayPerc > 0 && sb.displayPerc < 1)
}

func (sb *ScrollBar) onStageMouseMove(position Point) {
	if !sb.dragging || sb.grip == nil || sb.target == nil {
		return
	}

	sb.dragMutex.Lock()
	defer sb.dragMutex.Unlock()

	localX, localY := sb.GlobalToLocal(position.X, position.Y)

	// 获取 bar 的尺寸
	var barSize float64
	if sb.vertical {
		if sizable, ok := sb.bar.(interface{ Height() float64 }); ok {
			barSize = sizable.Height()
		}
	} else {
		if sizable, ok := sb.bar.(interface{ Width() float64 }); ok {
			barSize = sizable.Width()
		}
	}

	// 获取 grip 的尺寸
	var gripSize float64
	if sb.vertical {
		if sizable, ok := sb.grip.(interface{ Height() float64 }); ok {
			gripSize = sizable.Height()
		}
	} else {
		if sizable, ok := sb.grip.(interface{ Width() float64 }); ok {
			gripSize = sizable.Width()
		}
	}

	track := barSize - gripSize
	if track <= 0 {
		return
	}

	if sb.vertical {
		// 获取 bar 的 Y 坐标
		var barY float64
		if pos, ok := sb.bar.(interface{ Y() float64 }); ok {
			barY = pos.Y()
		}

		curY := localY - sb.dragOffset.Y
		perc := (curY - barY) / track
		sb.target.SetPercY(perc, false)
	} else {
		// 获取 bar 的 X 坐标
		var barX float64
		if pos, ok := sb.bar.(interface{ X() float64 }); ok {
			barX = pos.X()
		}

		curX := localX - sb.dragOffset.X
		perc := (curX - barX) / track
		sb.target.SetPercX(perc, false)
	}
}

func (sb *ScrollBar) onStageMouseUp() {
	if !sb.dragging {
		return
	}

	sb.dragMutex.Lock()
	defer sb.dragMutex.Unlock()

	sb.dragging = false
	sb.unregisterStageDrag()
}

func (sb *ScrollBar) onBarMouseDown(position Point) {
	if sb.target == nil || sb.bar == nil || sb.grip == nil {
		return
	}

	localX, localY := sb.GlobalToLocal(position.X, position.Y)

	// 获取 grip 的坐标
	var gripX, gripY float64
	if pos, ok := sb.grip.(interface{ X() float64 }); ok {
		gripX = pos.X()
	}
	if pos, ok := sb.grip.(interface{ Y() float64 }); ok {
		gripY = pos.Y()
	}

	if sb.vertical {
		if localY < gripY {
			sb.target.ScrollUp()
		} else {
			sb.target.ScrollDown()
		}
	} else {
		if localX < gripX {
			sb.target.ScrollLeft()
		} else {
			sb.target.ScrollRight()
		}
	}
}

func (sb *ScrollBar) registerStageDrag() {
	if sb.stageMoveHandler == nil {
		sb.stageMoveHandler = func(event Event) {
			if mouseEvent, ok := event.(*MouseEvent); ok {
				sb.onStageMouseMove(Point{X: mouseEvent.LocalX, Y: mouseEvent.LocalY})
			}
		}
	}

	if sb.stageUpHandler == nil {
		sb.stageUpHandler = func(event Event) {
			sb.onStageMouseUp()
		}
	}

	root := core.Root()
	if root != nil {
		// 将 EventHandler 转换为 laya.Listener
		if sb.stageMoveHandler != nil && sb.stageMoveListener == nil {
			sb.stageMoveListener = func(evt *laya.Event) {
				if evt != nil && sb.stageMoveHandler != nil {
					sb.stageMoveHandler(&BaseEvent{typ: string(evt.Type), target: root})
				}
			}
		}
		if sb.stageUpHandler != nil && sb.stageUpListener == nil {
			sb.stageUpListener = func(evt *laya.Event) {
				if evt != nil && sb.stageUpHandler != nil {
					sb.stageUpHandler(&BaseEvent{typ: string(evt.Type), target: root})
				}
			}
		}
		if sb.stageMoveListener != nil {
			root.On("mousemove", sb.stageMoveListener)
		}
		if sb.stageUpListener != nil {
			root.On("mouseup", sb.stageUpListener)
		}
	}
}

func (sb *ScrollBar) unregisterStageDrag() {
	root := core.Root()
	if root != nil {
		if sb.stageMoveListener != nil {
			root.Off("mousemove", sb.stageMoveListener)
			sb.stageMoveListener = nil
		}
		if sb.stageUpListener != nil {
			root.Off("mouseup", sb.stageUpListener)
			sb.stageUpListener = nil
		}
	}
}

// ============================================================================
// 公开 API
// ============================================================================

// ScrollPerc 返回滚动百分比
func (sb *ScrollBar) ScrollPerc() float64 {
	return sb.scrollPerc
}

// DisplayPerc 返回显示百分比
func (sb *ScrollBar) DisplayPerc() float64 {
	return sb.displayPerc
}

// IsVertical 返回是否是垂直方向
func (sb *ScrollBar) IsVertical() bool {
	return sb.vertical
}

// IsDragging 返回是否正在拖拽
func (sb *ScrollBar) IsDragging() bool {
	sb.dragMutex.Lock()
	defer sb.dragMutex.Unlock()
	return sb.dragging
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertScrollBar 类型断言
func AssertScrollBar(obj DisplayObject) (*ScrollBar, bool) {
	scrollbar, ok := obj.(*ScrollBar)
	return scrollbar, ok
}

// IsScrollBar 检查是否是 ScrollBar
func IsScrollBar(obj DisplayObject) bool {
	_, ok := obj.(*ScrollBar)
	return ok
}
