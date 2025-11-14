package core

import (
	"math"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
)

// debugLog 调试日志（空实现，用于开发调试）
func debugLog(format string, args ...interface{}) {
	// 空实现，调试时可以取消注释输出日志
	// fmt.Printf("[DEBUG] "+format+"\n", args...)
}

// ScrollType mirrors FairyGUI 的滚动方向枚举。
type ScrollType int

const (
	// ScrollTypeHorizontal 仅允许横向滚动。
	ScrollTypeHorizontal ScrollType = iota
	// ScrollTypeVertical 仅允许纵向滚动。
	ScrollTypeVertical
	// ScrollTypeBoth 允许双向滚动。
	ScrollTypeBoth
)

// PULL_RATIO 边缘回弹比例 - 下拉过顶或者上拉过底时允许超过的距离占显示区域的比例
const PULL_RATIO = 0.5

// TWEEN_TIME_GO 调用 SetPos(ani) 时使用的缓动时间
const TWEEN_TIME_GO = 0.5

// TWEEN_TIME_DEFAULT 惯性滚动的最小缓动时间
const TWEEN_TIME_DEFAULT = 0.3

// ScrollPaneLoopOwner 定义循环列表拥有者的接口，避免循环导入
type ScrollPaneLoopOwner interface {
	ColumnGap() int
	LineGap() int
}

// ScrollBarDisplayType mirrors FairyGUI 的滚动条显示模式。
type ScrollBarDisplayType int

const (
	ScrollBarDisplayDefault ScrollBarDisplayType = iota
	ScrollBarDisplayVisible
	ScrollBarDisplayAuto
	ScrollBarDisplayHidden
)

// ScrollInfo 描述当前滚动位置与可视比例。
type ScrollInfo struct {
	PercX        float64
	PercY        float64
	DisplayPercX float64
	DisplayPercY float64
}

// ScrollListener 响应滚动状态变化。
type ScrollListener func(ScrollInfo)

// ScrollPane 管理 GComponent 的滚动视图区域。
type ScrollPane struct {
	owner         *GComponent
	container     *laya.Sprite
	maskContainer *laya.Sprite

	scrollType        ScrollType
	scrollStep        float64
	mouseWheelStep    float64
	xPos              float64
	yPos              float64
	viewSize          laya.Point
	contentSize       laya.Point
	overlapSize       laya.Point
	wheelListener     laya.Listener
	wheelEnabled      bool
	scrollRectDirty   bool
	touchEffect       bool
	bouncebackEffect  bool
	pageMode          bool
	pageSize          laya.Point
	snapToItem        bool
	inertiaDisabled   bool
	beginTouch        laya.Point
	lastTouch         laya.Point
	containerOrigin   laya.Point
	velocity          laya.Point
	lastMoveTime      time.Time
	dragging          bool
	mouseDownListener laya.Listener
	stageMoveListener laya.Listener
	stageUpListener   laya.Listener
	scrollListeners   map[int]ScrollListener
	nextListenerID    int

	// 滚动条相关
	hzScrollBar      *GObject // 水平滚动条
	vtScrollBar      *GObject // 垂直滚动条
	hzScrollBarURL   string   // 水平滚动条资源 URL
	vtScrollBarURL   string   // 垂直滚动条资源 URL
	scrollBarDisplay int      // 滚动条显示模式
	displayOnLeft    bool     // 垂直滚动条是否在左侧显示
	floating         bool     // 浮动滚动条（不占用 viewSize）

	// 循环滚动支持
	// 0=无循环, 1=水平循环, 2=垂直循环, 3=both
	loop int

	// Tween 相关字段
	tweening         int        // 0=无动画, 1=setPos动画, 2=惯性/回弹动画
	tweenStart       laya.Point // 动画起始位置
	tweenChange      laya.Point // 动画变化量
	tweenDuration    laya.Point // 动画持续时间
	tweenTime        laya.Point // 当前动画时间
	tickerCleanup    func()     // ticker 取消注册函数
	decelerationRate float64    // 减速速率，默认 0.997
	velocityScale    float64    // 速度缩放因子，默认 1.0
}

func newScrollPane(owner *GComponent) *ScrollPane {
	if owner == nil {
		return nil
	}
	pane := &ScrollPane{
		owner:            owner,
		scrollType:       ScrollTypeBoth,
		scrollStep:       25,
		mouseWheelStep:   50,
		wheelEnabled:     true,
		touchEffect:      true,
		bouncebackEffect: true, // 默认启用边缘回弹效果
		pageSize:         laya.Point{X: 1, Y: 1},
		scrollListeners:  make(map[int]ScrollListener),
	}
	// 初始化 Tween 相关字段
	pane.tweening = 0
	pane.tweenStart = laya.Point{X: 0, Y: 0}
	pane.tweenChange = laya.Point{X: 0, Y: 0}
	pane.tweenDuration = laya.Point{X: 0, Y: 0}
	pane.tweenTime = laya.Point{X: 0, Y: 0}
	pane.decelerationRate = 0.997 // TypeScript 默认值
	pane.velocityScale = 1.0      // TypeScript 默认值
	container := owner.ensureContainer()
	display := owner.DisplayObject()
	if container != nil && display != nil {
		if container.Parent() == display {
			display.RemoveChild(container)
		}
		mask := laya.NewSprite()
		mask.SetOwner(owner)
		mask.SetMouseEnabled(true)
		display.AddChild(mask)
		mask.AddChild(container)
		pane.container = container
		pane.maskContainer = mask

		// 应用 margin 偏移到 container（对应 TypeScript ScrollPane.ts:128-130）
		margin := owner.Margin()
		container.SetPosition(float64(margin.Left), float64(margin.Top))
	}
	owner.scrollPane = pane
	pane.refreshViewSize()
	pane.contentSize = pane.viewSize
	pane.refreshOverlap()
	pane.registerEvents()
	return pane
}

// ScrollPane returns the scroll pane attached to this component.
func (c *GComponent) ScrollPane() *ScrollPane {
	if c == nil {
		return nil
	}
	return c.scrollPane
}

// EnsureScrollPane creates the scroll pane if absent and returns it.
func (c *GComponent) EnsureScrollPane(scrollType ScrollType) *ScrollPane {
	if c == nil {
		return nil
	}
	if c.scrollPane == nil {
		c.scrollPane = newScrollPane(c)
	}
	if c.scrollPane != nil {
		c.scrollPane.scrollType = scrollType
		c.scrollPane.refreshViewSize()
		c.scrollPane.updateScrollRect()
	}
	return c.scrollPane
}

// AddScrollListener 注册滚动状态监听，立即回调一次并返回监听 ID。
func (p *ScrollPane) AddScrollListener(fn ScrollListener) int {
	if p == nil || fn == nil {
		return 0
	}
	p.nextListenerID++
	id := p.nextListenerID
	if p.scrollListeners == nil {
		p.scrollListeners = make(map[int]ScrollListener)
	}
	p.scrollListeners[id] = fn
	fn(p.currentScrollInfo())
	return id
}

// RemoveScrollListener 移除指定监听。
func (p *ScrollPane) RemoveScrollListener(id int) {
	if p == nil || id == 0 {
		return
	}
	delete(p.scrollListeners, id)
}

// ScrollType 返回当前滚动类型。
func (p *ScrollPane) ScrollType() ScrollType {
	if p == nil {
		return ScrollTypeBoth
	}
	return p.scrollType
}

// SetScrollType updates the allowed scroll direction.
func (p *ScrollPane) SetScrollType(t ScrollType) {
	if p == nil {
		return
	}
	p.scrollType = t
}

// MouseWheelEnabled reports whether滚轮滚动生效。
func (p *ScrollPane) MouseWheelEnabled() bool {
	if p == nil {
		return false
	}
	return p.wheelEnabled
}

// SetMouseWheelEnabled toggles滚轮滚动。
func (p *ScrollPane) SetMouseWheelEnabled(enabled bool) {
	if p == nil {
		return
	}
	p.wheelEnabled = enabled
}

// ViewWidth returns the current viewport width。
func (p *ScrollPane) ViewWidth() float64 {
	if p == nil {
		return 0
	}
	return p.viewSize.X
}

// ViewHeight returns the current viewport height。
func (p *ScrollPane) ViewHeight() float64 {
	if p == nil {
		return 0
	}
	return p.viewSize.Y
}

// SetViewSize updates the viewport dimensions。
// 对应 TypeScript 版本 ScrollPane.ts:694-736 (setSize)
func (p *ScrollPane) SetViewSize(width, height float64) {
	if p == nil {
		return
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	// 先设置滚动条的位置和尺寸（对应 TypeScript ScrollPane.ts:697-719）
	if p.hzScrollBar != nil {
		// 水平滚动条放在底部
		hzY := height - p.hzScrollBar.Height()

		if p.vtScrollBar != nil {
			// 如果有垂直滚动条，水平滚动条宽度要减去垂直滚动条的宽度
			p.hzScrollBar.SetSize(width-p.vtScrollBar.Width(), p.hzScrollBar.Height())
			if p.displayOnLeft {
				p.hzScrollBar.SetPosition(p.vtScrollBar.Width(), hzY)
			} else {
				p.hzScrollBar.SetPosition(0, hzY)
			}
		} else {
			p.hzScrollBar.SetSize(width, p.hzScrollBar.Height())
			p.hzScrollBar.SetPosition(0, hzY)
		}
	}

	if p.vtScrollBar != nil {
		// 垂直滚动条放在右侧（或左侧，如果 displayOnLeft=true）
		var vtX float64
		if !p.displayOnLeft {
			vtX = width - p.vtScrollBar.Width()
		} else {
			vtX = 0
		}

		if p.hzScrollBar != nil {
			// 如果有水平滚动条，垂直滚动条高度要减去水平滚动条的高度
			p.vtScrollBar.SetSize(p.vtScrollBar.Width(), height-p.hzScrollBar.Height())
		} else {
			p.vtScrollBar.SetSize(p.vtScrollBar.Width(), height)
		}
		p.vtScrollBar.SetPosition(vtX, 0)
	}

	// 设置初始 viewSize（对应 TypeScript ScrollPane.ts:721-722）
	p.viewSize = laya.Point{X: width, Y: height}

	// 如果有滚动条且不是浮动模式，减去滚动条尺寸
	// 对应 TypeScript ScrollPane.ts:723-726
	if p.hzScrollBar != nil && !p.floating {
		p.viewSize.Y -= p.hzScrollBar.Height()
	}
	if p.vtScrollBar != nil && !p.floating {
		p.viewSize.X -= p.vtScrollBar.Width()
	}

	// 减去 margin（对应 TypeScript ScrollPane.ts:727-728）
	if p.owner != nil {
		margin := p.owner.Margin()
		p.viewSize.X -= float64(margin.Left + margin.Right)
		p.viewSize.Y -= float64(margin.Top + margin.Bottom)
	}

	// 确保最小为 1（对应 TypeScript ScrollPane.ts:730-731）
	if p.viewSize.X < 1 {
		p.viewSize.X = 1
	}
	if p.viewSize.Y < 1 {
		p.viewSize.Y = 1
	}

	// 更新 pageSize（对应 TypeScript ScrollPane.ts:732-733）
	p.pageSize.X = p.viewSize.X
	p.pageSize.Y = p.viewSize.Y

	p.updateScrollRect()
	p.refreshOverlap()
	p.clampPosition()
	p.notifyScrollListeners()
	p.updateScrollBars() // 更新滚动条显示百分比（对应 TypeScript ScrollPane.ts:735）
}

// SetLoop 设置循环滚动模式
// mode: 0=无循环, 1=水平循环, 2=垂直循环, 3=both
func (p *ScrollPane) SetLoop(mode int) {
	if p == nil {
		return
	}
	p.loop = mode
}

// LoopCheckingCurrent 检查当前位置并进行循环调整
// 返回是否发生了位置改变
func (p *ScrollPane) LoopCheckingCurrent() bool {
	changed := false

	// 水平循环
	if (p.loop == 1 || p.loop == 3) && p.overlapSize.X > 0 {
		if p.xPos < 0.001 {
			p.xPos += p.getLoopPartSize(2, "x")
			changed = true
		} else if p.xPos >= p.overlapSize.X {
			p.xPos -= p.getLoopPartSize(2, "x")
			changed = true
		}
	}

	// 垂直循环
	if (p.loop == 2 || p.loop == 3) && p.overlapSize.Y > 0 {
		if p.yPos < 0.001 {
			p.yPos += p.getLoopPartSize(2, "y")
			changed = true
		} else if p.yPos >= p.overlapSize.Y {
			p.yPos -= p.getLoopPartSize(2, "y")
			changed = true
		}
	}

	if changed {
		p.container.SetPosition(float64(-int(p.xPos)), float64(-int(p.yPos)))
	}

	return changed
}

// getLoopPartSize 获取循环部分的大小
func (p *ScrollPane) getLoopPartSize(division int, axis string) float64 {
	var gap float64

	// 检查 owner 是否有循环列表的数据
	if p.owner != nil {
		// 如果 owner 有 Data() 方法，尝试获取 ScrollPaneLoopOwner
		type dataGetter interface {
			Data() interface{}
		}
		if dataOwner, ok := interface{}(p.owner).(dataGetter); ok {
			if listData, ok := dataOwner.Data().(ScrollPaneLoopOwner); ok {
				if axis == "x" {
					gap = float64(listData.ColumnGap())
					return (p.contentSize.X + gap) / float64(division)
				} else {
					gap = float64(listData.LineGap())
					return (p.contentSize.Y + gap) / float64(division)
				}
			}
		}
	}

	// 默认返回内容尺寸的一部分
	if axis == "x" {
		return p.contentSize.X / float64(division)
	}
	return p.contentSize.Y / float64(division)
}

// SetContentSize updates the scrollable content size。
// ContentSize returns the content dimensions.
func (p *ScrollPane) ContentSize() laya.Point {
	if p == nil {
		return laya.Point{}
	}
	return p.contentSize
}

func (p *ScrollPane) SetContentSize(width, height float64) {
	if p == nil {
		return
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	p.contentSize = laya.Point{X: width, Y: height}
	p.refreshOverlap()
	p.clampPosition()
	p.notifyScrollListeners()
	p.updateScrollBars() // 更新滚动条显示百分比
}

// SetScrollStep sets the per-step scroll amount used for wheel滚动。
func (p *ScrollPane) SetScrollStep(step float64) {
	if p == nil {
		return
	}
	if step <= 0 {
		step = 25
	}
	p.scrollStep = step
}

// SetMouseWheelStep sets the per-step delta used for mouse wheel。
func (p *ScrollPane) SetMouseWheelStep(step float64) {
	if p == nil {
		return
	}
	if step <= 0 {
		step = p.scrollStep * 2
	}
	p.mouseWheelStep = step
}

// PosX returns the current horizontal scroll position。
func (p *ScrollPane) PosX() float64 {
	if p == nil {
		return 0
	}
	return p.xPos
}

// PosY returns the current vertical scroll position。
func (p *ScrollPane) PosY() float64 {
	if p == nil {
		return 0
	}
	return p.yPos
}

// SetPos updates the scroll offsets (ani 参数保留以兼容原接口，目前忽略)。
func (p *ScrollPane) SetPos(x, y float64, _ bool) {
	if p == nil {
		return
	}
	p.setPos(x, y)
}

// SetPercX updates horizontal scroll by百分比。
func (p *ScrollPane) SetPercX(value float64, _ bool) {
	if p == nil {
		return
	}
	if p.overlapSize.X <= 0 {
		p.setPos(0, p.yPos)
		return
	}
	p.setPos(value*p.overlapSize.X, p.yPos)
}

// SetPercY updates vertical scroll by百分比。
func (p *ScrollPane) SetPercY(value float64, _ bool) {
	if p == nil {
		return
	}
	if p.overlapSize.Y <= 0 {
		p.setPos(p.xPos, 0)
		return
	}
	p.setPos(p.xPos, value*p.overlapSize.Y)
}

// ScrollToRect scrolls the viewport so the specified rectangle becomes visible.
// The rectangle is expressed in the owner's local coordinates.
func (p *ScrollPane) ScrollToRect(x, y, width, height float64, _ bool) {
	if p == nil {
		return
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	targetX := p.xPos
	targetY := p.yPos
	viewW := p.viewSize.X
	viewH := p.viewSize.Y
	if viewW > 0 {
		visibleRight := p.xPos + viewW
		rectRight := x + width
		if width >= viewW {
			targetX = x
		} else {
			if x < p.xPos {
				targetX = x
			} else if rectRight > visibleRight {
				targetX = rectRight - viewW
			}
		}
	}
	if viewH > 0 {
		visibleBottom := p.yPos + viewH
		rectBottom := y + height
		if height >= viewH {
			targetY = y
		} else {
			if y < p.yPos {
				targetY = y
			} else if rectBottom > visibleBottom {
				targetY = rectBottom - viewH
			}
		}
	}
	p.SetPos(targetX, targetY, false)
}

// OnOwnerSizeChanged updates viewport when宿主尺寸发生变化。
func (p *ScrollPane) OnOwnerSizeChanged() {
	if p == nil || p.owner == nil {
		return
	}
	p.SetViewSize(p.owner.Width(), p.owner.Height())
}

// Dispose detaches事件监听。
func (p *ScrollPane) Dispose() {
	if p == nil {
		return
	}
	// 停止动画并取消 ticker 注册
	p.killTween()

	p.unregisterEvents()
	// 清理滚动条引用
	p.hzScrollBar = nil
	p.vtScrollBar = nil
}

// SetHzScrollBar 设置水平滚动条。
func (p *ScrollPane) SetHzScrollBar(bar *GObject) {
	if p == nil {
		return
	}
	p.hzScrollBar = bar
}

// SetVtScrollBar 设置垂直滚动条。
func (p *ScrollPane) SetVtScrollBar(bar *GObject) {
	if p == nil {
		return
	}
	p.vtScrollBar = bar
}

// HzScrollBarURL 返回水平滚动条资源 URL。
func (p *ScrollPane) HzScrollBarURL() string {
	if p == nil {
		return ""
	}
	return p.hzScrollBarURL
}

// VtScrollBarURL 返回垂直滚动条资源 URL。
func (p *ScrollPane) VtScrollBarURL() string {
	if p == nil {
		return ""
	}
	return p.vtScrollBarURL
}

// ScrollBarDisplay 返回滚动条显示模式。
func (p *ScrollPane) ScrollBarDisplay() int {
	if p == nil {
		return 0
	}
	return p.scrollBarDisplay
}

// DisplayOnLeft 返回垂直滚动条是否在左侧显示。
func (p *ScrollPane) DisplayOnLeft() bool {
	if p == nil {
		return false
	}
	return p.displayOnLeft
}

// Floating 返回滚动条是否浮动（不占用 viewSize）。
func (p *ScrollPane) Floating() bool {
	if p == nil {
		return false
	}
	return p.floating
}

func (p *ScrollPane) setPos(x, y float64) {
	if p == nil {
		return
	}
	maxX := p.overlapSize.X
	maxY := p.overlapSize.Y
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	if x < 0 {
		x = 0
	} else if x > maxX {
		x = maxX
	}
	if y < 0 {
		y = 0
	} else if y > maxY {
		y = maxY
	}
	changed := x != p.xPos || y != p.yPos
	p.xPos = x
	p.yPos = y
	if changed {
		p.applyPosition()
		p.notifyScrollListeners()
		p.updateScrollBars() // 更新滚动条位置
	}
}

func (p *ScrollPane) applyPosition() {
	if p == nil || p.container == nil || p.owner == nil {
		return
	}
	// container 的位置 = margin 偏移 - scroll 偏移
	// 这样 container 中的子对象（位置为 child.X, child.Y）最终会出现在正确的位置
	margin := p.owner.Margin()
	p.container.SetPosition(
		float64(margin.Left)-p.xPos,
		float64(margin.Top)-p.yPos,
	)
}

func (p *ScrollPane) refreshViewSize() {
	if p == nil || p.owner == nil {
		return
	}
	width := p.owner.Width()
	height := p.owner.Height()
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	p.viewSize = laya.Point{X: width, Y: height}

	// 减去 margin（对应 TypeScript ScrollPane.ts:727-728）
	margin := p.owner.Margin()
	p.viewSize.X -= float64(margin.Left + margin.Right)
	p.viewSize.Y -= float64(margin.Top + margin.Bottom)

	// 确保最小为 1
	if p.viewSize.X < 1 {
		p.viewSize.X = 1
	}
	if p.viewSize.Y < 1 {
		p.viewSize.Y = 1
	}

	p.updateScrollRect()
	p.refreshOverlap()
}

func (p *ScrollPane) refreshOverlap() {
	if p == nil {
		return
	}
	p.overlapSize = laya.Point{
		X: math.Max(p.contentSize.X-p.viewSize.X, 0),
		Y: math.Max(p.contentSize.Y-p.viewSize.Y, 0),
	}
}

func (p *ScrollPane) clampPosition() {
	if p == nil {
		return
	}
	p.setPos(p.xPos, p.yPos)
}

func (p *ScrollPane) updateScrollRect() {
	if p == nil || p.maskContainer == nil {
		return
	}

	// scrollRect 的位置应该考虑 container 的偏移（margin）
	margin := p.owner.Margin()
	rect := &laya.Rect{
		X: float64(margin.Left),
		Y: float64(margin.Top),
		W: p.viewSize.X,
		H: p.viewSize.Y,
	}
	p.maskContainer.SetScrollRect(rect)
}

func (p *ScrollPane) registerEvents() {
	if p == nil || p.owner == nil {
		return
	}
	display := p.owner.DisplayObject()
	if display == nil {
		return
	}
	dispatcher := display.Dispatcher()
	if p.wheelListener == nil {
		p.wheelListener = func(evt *laya.Event) {
			pe, ok := evt.Data.(laya.PointerEvent)
			if !ok {
				return
			}
			p.handleMouseWheel(pe.WheelX, pe.WheelY)
		}
		dispatcher.On(laya.EventMouseWheel, p.wheelListener)
	}
	if p.mouseDownListener == nil {
		p.mouseDownListener = func(evt *laya.Event) {
			p.onMouseDown(*evt)
		}
		dispatcher.On(laya.EventMouseDown, p.mouseDownListener)
	}
}

func (p *ScrollPane) unregisterEvents() {
	if p == nil || p.owner == nil {
		return
	}
	display := p.owner.DisplayObject()
	if display == nil {
		return
	}
	dispatcher := display.Dispatcher()
	if p.wheelListener != nil {
		dispatcher.Off(laya.EventMouseWheel, p.wheelListener)
		p.wheelListener = nil
	}
	if p.mouseDownListener != nil {
		dispatcher.Off(laya.EventMouseDown, p.mouseDownListener)
		p.mouseDownListener = nil
	}
	p.unregisterStageDragEvents()
}

func (p *ScrollPane) handleMouseWheel(deltaX, deltaY float64) {
	if p == nil {
		return
	}
	if !p.wheelEnabled {
		return
	}
	if deltaY != 0 && (p.scrollType == ScrollTypeVertical || p.scrollType == ScrollTypeBoth) {
		p.setPos(p.xPos, p.yPos-deltaY*p.mouseWheelStep)
	}
	if deltaX != 0 && (p.scrollType == ScrollTypeHorizontal || p.scrollType == ScrollTypeBoth) {
		p.setPos(p.xPos-deltaX*p.mouseWheelStep, p.yPos)
	}
}

// ScrollUp 按当前步长向上滚动。
func (p *ScrollPane) ScrollUp() {
	p.scrollBy(0, -p.scrollStep)
}

// ScrollDown 按当前步长向下滚动。
func (p *ScrollPane) ScrollDown() {
	p.scrollBy(0, p.scrollStep)
}

// ScrollLeft 按当前步长向左滚动。
func (p *ScrollPane) ScrollLeft() {
	p.scrollBy(-p.scrollStep, 0)
}

// ScrollRight 按当前步长向右滚动。
func (p *ScrollPane) ScrollRight() {
	p.scrollBy(p.scrollStep, 0)
}

// ScrollTop 滚动到顶部。
// 参数 ani 控制是否使用动画（当前未实现动画效果）。
func (p *ScrollPane) ScrollTop(ani bool) {
	p.SetPercY(0, ani)
}

// ScrollBottom 滚动到底部。
// 参数 ani 控制是否使用动画（当前未实现动画效果）。
func (p *ScrollPane) ScrollBottom(ani bool) {
	p.SetPercY(1, ani)
}

func (p *ScrollPane) scrollBy(dx, dy float64) {
	if p == nil {
		return
	}
	x := p.xPos
	y := p.yPos
	if dx != 0 && (p.scrollType == ScrollTypeHorizontal || p.scrollType == ScrollTypeBoth) {
		x += dx
	}
	if dy != 0 && (p.scrollType == ScrollTypeVertical || p.scrollType == ScrollTypeBoth) {
		y += dy
	}
	p.setPos(x, y)
}

func (p *ScrollPane) currentScrollInfo() ScrollInfo {
	info := ScrollInfo{}
	if p == nil {
		return info
	}
	if p.overlapSize.X > 0 {
		info.PercX = clamp01(p.xPos / p.overlapSize.X)
	}
	if p.overlapSize.Y > 0 {
		info.PercY = clamp01(p.yPos / p.overlapSize.Y)
	}
	if p.contentSize.X > 0 {
		info.DisplayPercX = clamp01(p.viewSize.X / p.contentSize.X)
	} else {
		info.DisplayPercX = 1
	}
	if p.contentSize.Y > 0 {
		info.DisplayPercY = clamp01(p.viewSize.Y / p.contentSize.Y)
	} else {
		info.DisplayPercY = 1
	}

	return info

}

func (p *ScrollPane) notifyScrollListeners() {
	if p == nil || len(p.scrollListeners) == 0 {
		return
	}
	info := p.currentScrollInfo()
	for _, fn := range p.scrollListeners {
		if fn != nil {
			fn(info)
		}
	}
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func (p *ScrollPane) onMouseDown(evt laya.Event) {
	if p == nil || !p.touchEffect || p.owner == nil {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok {
		return
	}
	display := p.owner.DisplayObject()
	if display == nil || p.container == nil {
		return
	}
	local := display.GlobalToLocal(pe.Position)
	p.beginTouch = local
	p.lastTouch = local
	p.containerOrigin = p.container.Position()
	p.velocity = laya.Point{}
	p.lastMoveTime = time.Now()
	p.dragging = true
	p.registerStageDragEvents()
}

func (p *ScrollPane) onStageMouseMove(evt laya.Event) {
	if p == nil || !p.dragging || p.owner == nil {
		return
	}

	// 如果正在 tween 动画，先停止（对应 TypeScript __mouseDown 第984行）
	if p.tweening != 0 {
		p.killTween()
	}

	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok {
		return
	}
	display := p.owner.DisplayObject()
	if display == nil {
		return
	}
	local := display.GlobalToLocal(pe.Position)
	now := time.Now()
	elapsed := now.Sub(p.lastMoveTime).Seconds()
	deltaX := local.X - p.lastTouch.X
	deltaY := local.Y - p.lastTouch.Y

	// 只有当移动距离足够大时才更新 velocity（避免最后一帧重置为0）
	if elapsed > 0 && (math.Abs(deltaX) > 0.5 || math.Abs(deltaY) > 0.5) {
		p.velocity = laya.Point{
			X: deltaX / elapsed,
			Y: deltaY / elapsed,
		}
		// 调试：记录 velocity 计算
		debugLog("[ScrollPane] velocity: (%.2f, %.2f), delta: (%.2f, %.2f), elapsed: %.4f",
			p.velocity.X, p.velocity.Y, deltaX, deltaY, elapsed)
	}
	p.lastMoveTime = now
	p.lastTouch = local

	targetX := p.containerOrigin.X + (local.X - p.beginTouch.X)
	targetY := p.containerOrigin.Y + (local.Y - p.beginTouch.Y)
	p.applyDragPosition(targetX, targetY)
}

func (p *ScrollPane) onStageMouseUp(evt laya.Event) {
	if p == nil || !p.dragging {
		return
	}
	p.dragging = false
	p.unregisterStageDragEvents()

	if !p.touchEffect {
		// 没有触摸效果，直接返回
		if p.container != nil {
			p.containerOrigin = p.container.Position()
		}
		return
	}

	if p.container == nil {
		return
	}

	// 保存当前 container 位置（对应 TypeScript __mouseUp 第1203行）
	pos := p.container.Position()
	p.tweenStart.X = pos.X
	p.tweenStart.Y = pos.Y

	// 初始化目标位置（对应 TypeScript sEndPos 第1205行）
	endX := p.tweenStart.X
	endY := p.tweenStart.Y

	flag := false

	// 检查是否需要回弹到边界（对应 TypeScript 1207-1222行）
	if p.tweenStart.X > 0 {
		endX = 0
		flag = true
	} else if p.tweenStart.X < -p.overlapSize.X {
		endX = -p.overlapSize.X
		flag = true
	}
	if p.tweenStart.Y > 0 {
		endY = 0
		flag = true
	} else if p.tweenStart.Y < -p.overlapSize.Y {
		endY = -p.overlapSize.Y
		flag = true
	}

	// 调试：记录启动参数
	debugLog("[ScrollPane] onStageMouseUp: start=(%.2f,%.2f), overlap=(%.2f,%.2f), velocity=(%.2f,%.2f)",
		p.tweenStart.X, p.tweenStart.Y, p.overlapSize.X, p.overlapSize.Y, p.velocity.X, p.velocity.Y)

	if flag {
		// 需要回弹到边界（对应 TypeScript 1223-1252行）
		p.tweenChange.X = endX - p.tweenStart.X
		p.tweenChange.Y = endY - p.tweenStart.Y

		// 设置回弹动画持续时间
		p.tweenDuration.X = TWEEN_TIME_DEFAULT
		p.tweenDuration.Y = TWEEN_TIME_DEFAULT

		debugLog("[ScrollPane] 启动边界回弹: change=(%.2f,%.2f), duration=%.2f",
			p.tweenChange.X, p.tweenChange.Y, TWEEN_TIME_DEFAULT)
	} else {
		// 不需要回弹，检查是否需要惯性滚动（对应 TypeScript 1254-1288行）
		if !p.inertiaDisabled {
			// 使用完整的惯性滚动算法计算目标位置
			endX = p.updateTargetAndDuration2(p.tweenStart.X, "x")
			endY = p.updateTargetAndDuration2(p.tweenStart.Y, "y")

			// 计算变化量
			p.tweenChange.X = endX - p.tweenStart.X
			p.tweenChange.Y = endY - p.tweenStart.Y

			debugLog("[ScrollPane] 启动惯性滚动: change=(%.2f,%.2f), duration=(%.2f,%.2f)",
				p.tweenChange.X, p.tweenChange.Y, p.tweenDuration.X, p.tweenDuration.Y)
		} else {
			// 没有惯性，直接回到当前位置（无动画）
			p.tweenChange.X = 0
			p.tweenChange.Y = 0
			p.tweenDuration.X = TWEEN_TIME_DEFAULT
			p.tweenDuration.Y = TWEEN_TIME_DEFAULT

			debugLog("[ScrollPane] 惯性已禁用，无动画")
		}
	}

	// 如果无需动画，直接返回
	if p.tweenChange.X == 0 && p.tweenChange.Y == 0 {
		p.containerOrigin = p.container.Position()
		p.updateScrollBarVisible()
		debugLog("[ScrollPane] 无需动画，直接返回")
		return
	}

	// 启动 Tween 动画（对应 TypeScript 1290行）
	debugLog("[ScrollPane] 启动 Tween 动画类型 2")
	p.startTween(2)

	// 调试：检查是否需要动画
	if p.tweenChange.X == 0 && p.tweenChange.Y == 0 {
		// 无需动画
		p.containerOrigin = p.container.Position()
		return
	}
}

func (p *ScrollPane) applyDragPosition(containerX, containerY float64) {
	if p == nil || p.container == nil {
		return
	}
	if p.scrollType == ScrollTypeVertical {
		containerX = p.container.Position().X
	}
	if p.scrollType == ScrollTypeHorizontal {
		containerY = p.container.Position().Y
	}
	if p.overlapSize.X <= 0 {
		containerX = 0
	} else {
		if containerX > 0 {
			containerX = 0
		}
		if containerX < -p.overlapSize.X {
			containerX = -p.overlapSize.X
		}
	}
	if p.overlapSize.Y <= 0 {
		containerY = 0
	} else {
		if containerY > 0 {
			containerY = 0
		}
		if containerY < -p.overlapSize.Y {
			containerY = -p.overlapSize.Y
		}
	}
	p.setPos(-containerX, -containerY)
}

func (p *ScrollPane) registerStageDragEvents() {
	root := Root()
	if root == nil {
		return
	}
	stage := root.Stage()
	if stage == nil {
		return
	}
	dispatcher := stage.Root().Dispatcher()
	if p.stageMoveListener == nil {
		p.stageMoveListener = func(evt *laya.Event) {
			p.onStageMouseMove(*evt)
		}
	}
	if p.stageUpListener == nil {
		p.stageUpListener = func(evt *laya.Event) {
			p.onStageMouseUp(*evt)
		}
	}
	dispatcher.On(laya.EventMouseMove, p.stageMoveListener)
	dispatcher.On(laya.EventStageMouseUp, p.stageUpListener)
}

func (p *ScrollPane) unregisterStageDragEvents() {
	root := Root()
	if root == nil {
		return
	}
	stage := root.Stage()
	if stage == nil {
		return
	}
	dispatcher := stage.Root().Dispatcher()
	if p.stageMoveListener != nil {
		dispatcher.Off(laya.EventMouseMove, p.stageMoveListener)
	}
	if p.stageUpListener != nil {
		dispatcher.Off(laya.EventStageMouseUp, p.stageUpListener)
	}
}

func (p *ScrollPane) snapToNearestPage() {
	if p == nil {
		return
	}
	if !p.pageMode {
		p.clampPosition()
		return
	}
	pageW := p.pageSize.X
	if pageW <= 0 {
		pageW = p.viewSize.X
	}
	pageH := p.pageSize.Y
	if pageH <= 0 {
		pageH = p.viewSize.Y
	}
	targetX := p.xPos
	targetY := p.yPos
	if p.scrollType != ScrollTypeVertical && pageW > 0 && p.overlapSize.X > 0 {
		targetX = math.Round(targetX/pageW) * pageW
	}
	if p.scrollType != ScrollTypeHorizontal && pageH > 0 && p.overlapSize.Y > 0 {
		targetY = math.Round(targetY/pageH) * pageH
	}
	p.setPos(targetX, targetY)
}

// updateScrollBars 更新滚动条的位置和显示百分比
// 对应 TypeScript 版本 ScrollPane.ts:1316-1324 (updateScrollBarPos)
// 以及 ScrollPane.ts:816-827 (handleSizeChanged中的displayPerc设置)
func (p *ScrollPane) updateScrollBars() {
	if p == nil {
		return
	}

	info := p.currentScrollInfo()

	// 更新垂直滚动条
	if p.vtScrollBar != nil {
		if scrollBar, ok := p.vtScrollBar.Data().(interface{ SyncFromPane(ScrollInfo) }); ok {
			scrollBar.SyncFromPane(info)
		}
	}

	// 更新水平滚动条
	if p.hzScrollBar != nil {
		if scrollBar, ok := p.hzScrollBar.Data().(interface{ SyncFromPane(ScrollInfo) }); ok {
			scrollBar.SyncFromPane(info)
		}
	}

	// 更新滚动条可见性（对应 TypeScript ScrollPane.ts:829）
	p.updateScrollBarVisible()
}

// updateScrollBarVisible 更新滚动条的可见性
// 对应 TypeScript 版本 ScrollPane.ts:1326-1340
func (p *ScrollPane) updateScrollBarVisible() {
	if p == nil {
		return
	}

	// 检查垂直滚动条可见性
	if p.vtScrollBar != nil {
		vScrollNone := p.contentSize.Y <= p.viewSize.Y
		if vScrollNone {
			p.vtScrollBar.SetVisible(false)
		} else {
			// 根据 scrollBarDisplay 模式决定可见性
			// scrollBarDisplayAuto (0) - 自动显示/隐藏
			// 其他模式 - 始终显示
			p.vtScrollBar.SetVisible(true)
		}
	}

	// 检查水平滚动条可见性
	if p.hzScrollBar != nil {
		hScrollNone := p.contentSize.X <= p.viewSize.X
		if hScrollNone {
			p.hzScrollBar.SetVisible(false)
		} else {
			p.hzScrollBar.SetVisible(true)
		}
	}
}

// easeFunc 缓动函数 - cubicOut
// t: 当前时间, d: 总时间
func easeFunc(t, d float64) float64 {
	if d <= 0 {
		return 1
	}
	t = t/d - 1
	return t*t*t + 1
}

// startTween 启动 Tween 动画
// type: 0=无, 1=setPos动画, 2=惯性/回弹动画
func (p *ScrollPane) startTween(tweenType int) {
	if p == nil {
		return
	}

	// 如果已经在动画中，先停止之前的
	if p.tweening != 0 {
		p.killTween()
	}

	// 重置时间
	p.tweenTime.X = 0
	p.tweenTime.Y = 0
	p.tweening = tweenType

	debugLog("[ScrollPane] startTween: type=%d, change=(%.2f,%.2f), duration=(%.2f,%.2f)",
		tweenType, p.tweenChange.X, p.tweenChange.Y, p.tweenDuration.X, p.tweenDuration.Y)

	// 注册到 ticker 更新
	// 注意：这里不应该检查 tickerCleanup 是否为 nil，因为 killTween 已经清除了它
	p.tickerCleanup = RegisterTicker(func(delta time.Duration) {
		// 每帧更新
		completed := !p.tweenUpdate(delta.Seconds())
		if completed && p.tickerCleanup != nil {
			// 动画完成，取消注册
			debugLog("[ScrollPane] Tween 动画完成")
			p.tickerCleanup()
			p.tickerCleanup = nil
		}
	})

	debugLog("[ScrollPane] Ticker 已注册: cleanup=%v", p.tickerCleanup != nil)

	p.updateScrollBarVisible()
}

// killTween 停止 Tween 动画
func (p *ScrollPane) killTween() {
	if p == nil || p.tweening == 0 {
		return
	}

	// 如果是类型1的 tween，需要立即设置到终点
	if p.tweening == 1 {
		if p.container != nil {
			endX := p.tweenStart.X + p.tweenChange.X
			endY := p.tweenStart.Y + p.tweenChange.Y
			p.container.SetPosition(endX, endY)
		}
		if p.owner != nil && p.owner.DisplayObject() != nil {
			// 通知滚动事件
		}
	}

	p.tweening = 0

	// 取消 ticker 注册
	if p.tickerCleanup != nil {
		p.tickerCleanup()
		p.tickerCleanup = nil
	}

	p.updateScrollBarVisible()

	// 通知滚动结束
	if p.owner != nil && p.owner.DisplayObject() != nil {
		// 触发 SCROLL_END 事件
	}
}

// tweenUpdate 每帧更新 Tween 动画
// 返回 false 表示动画完成，应该停止调用
func (p *ScrollPane) tweenUpdate(delta float64) bool {
	if p == nil || p.tweening == 0 {
		return false
	}

	changed := false

	// 调试：记录帧更新
	debugLog("[ScrollPane] tweenUpdate: delta=%.4f, type=%d, change=(%.2f,%.2f), time=(%.2f,%.2f)",
		delta, p.tweening, p.tweenChange.X, p.tweenChange.Y, p.tweenTime.X, p.tweenTime.Y)

	// 更新 X 轴动画
	if p.tweenChange.X != 0 {
		p.tweenTime.X += delta
		var newX float64

		if p.tweenTime.X >= p.tweenDuration.X {
			// 动画完成
			newX = p.tweenStart.X + p.tweenChange.X
			debugLog("[ScrollPane] X 动画完成: newX=%.2f", newX)
			p.tweenChange.X = 0
		} else {
			// 使用缓动函数计算插值
			ratio := easeFunc(p.tweenTime.X, p.tweenDuration.X)
			newX = p.tweenStart.X + p.tweenChange.X*ratio
			debugLog("[ScrollPane] X 动画更新: time=%.2f/%.2f, ratio=%.2f, newX=%.2f",
				p.tweenTime.X, p.tweenDuration.X, ratio, newX)
		}

		// 边界检查
		threshold1 := 0.0
		threshold2 := -p.overlapSize.X

		// 回弹效果检查
		if p.tweening == 2 && p.bouncebackEffect {
			if (newX > 20+threshold1 && p.tweenChange.X > 0) ||
				(newX > threshold1 && p.tweenChange.X == 0) {
				// 开始回弹
				debugLog("[ScrollPane] X 开始回弹: newX=%.2f > threshold=%.2f", newX, threshold1)
				p.tweenTime.X = 0
				p.tweenDuration.X = TWEEN_TIME_DEFAULT
				p.tweenChange.X = -newX + threshold1
				p.tweenStart.X = newX
			} else if (newX < threshold2-20 && p.tweenChange.X < 0) ||
				(newX < threshold2 && p.tweenChange.X == 0) {
				// 开始回弹
				debugLog("[ScrollPane] X 开始回弹: newX=%.2f < threshold=%.2f", newX, threshold2)
				p.tweenTime.X = 0
				p.tweenDuration.X = TWEEN_TIME_DEFAULT
				p.tweenChange.X = threshold2 - newX
				p.tweenStart.X = newX
			}
		} else {
			// 不使用回弹，直接限制边界
			if newX > threshold1 {
				newX = threshold1
				p.tweenChange.X = 0
			} else if newX < threshold2 {
				newX = threshold2
				p.tweenChange.X = 0
			}
		}

		if p.container != nil {
			if pos := p.container.Position(); pos.X != newX {
				p.container.SetPosition(newX, pos.Y)
				changed = true
			}
		}
	}

	// 更新 Y 轴动画（逻辑相同）
	if p.tweenChange.Y != 0 {
		p.tweenTime.Y += delta
		var newY float64

		if p.tweenTime.Y >= p.tweenDuration.Y {
			newY = p.tweenStart.Y + p.tweenChange.Y
			debugLog("[ScrollPane] Y 动画完成: newY=%.2f", newY)
			p.tweenChange.Y = 0
		} else {
			ratio := easeFunc(p.tweenTime.Y, p.tweenDuration.Y)
			newY = p.tweenStart.Y + p.tweenChange.Y*ratio
			debugLog("[ScrollPane] Y 动画更新: time=%.2f/%.2f, ratio=%.2f, newY=%.2f",
				p.tweenTime.Y, p.tweenDuration.Y, ratio, newY)
		}

		threshold1 := 0.0
		threshold2 := -p.overlapSize.Y

		if p.tweening == 2 && p.bouncebackEffect {
			if (newY > 20+threshold1 && p.tweenChange.Y > 0) ||
				(newY > threshold1 && p.tweenChange.Y == 0) {
				p.tweenTime.Y = 0
				p.tweenDuration.Y = TWEEN_TIME_DEFAULT
				p.tweenChange.Y = -newY + threshold1
				p.tweenStart.Y = newY
			} else if (newY < threshold2-20 && p.tweenChange.Y < 0) ||
				(newY < threshold2 && p.tweenChange.Y == 0) {
				p.tweenTime.Y = 0
				p.tweenDuration.Y = TWEEN_TIME_DEFAULT
				p.tweenChange.Y = threshold2 - newY
				p.tweenStart.Y = newY
			}
		} else {
			if newY > threshold1 {
				newY = threshold1
				p.tweenChange.Y = 0
			} else if newY < threshold2 {
				newY = threshold2
				p.tweenChange.Y = 0
			}
		}

		if p.container != nil {
			if pos := p.container.Position(); pos.Y != newY {
				p.container.SetPosition(pos.X, newY)
				changed = true
			}
		}
	}

	// 更新 posX/posY
	if p.tweening == 2 {
		if p.overlapSize.X > 0 && p.container != nil {
			pos := p.container.Position()
			p.xPos = clamp01(-pos.X / p.overlapSize.X)
		}
		if p.overlapSize.Y > 0 && p.container != nil {
			pos := p.container.Position()
			p.yPos = clamp01(-pos.Y / p.overlapSize.Y)
		}
		if p.pageMode {
			p.snapToNearestPage()
		}
	}

	// 检查动画是否完成
	if p.tweenChange.X == 0 && p.tweenChange.Y == 0 {
		p.tweening = 0
		// 注意：LoopCheckingCurrent 方法尚未实现，暂时跳过
		// p.LoopCheckingCurrent()
		p.updateScrollBars()
		p.updateScrollBarVisible()
		return false
	}

	if changed {
		p.updateScrollBars()
	}

	return true
}


// updateTargetAndDuration2 根据速度计算目标位置和动画时间（完整实现）
// 对应 TypeScript ScrollPane.ts:1540-1587
func (p *ScrollPane) updateTargetAndDuration2(pos float64, axis string) float64 {
	var velocity float64
	if axis == "x" {
		velocity = p.velocity.X
	} else {
		velocity = p.velocity.Y
	}

	debugLog("[ScrollPane] updateTargetAndDuration2: axis=%s, pos=%.2f, velocity=%.2f", axis, pos, velocity)

	var duration float64 = TWEEN_TIME_DEFAULT

	// 边界检查
	if pos > 0 {
		debugLog("[ScrollPane] 超出左边界，返回 0")
		return 0
	}

	maxPos := -p.overlapSize.X
	if axis == "y" {
		maxPos = -p.overlapSize.Y
	}

	if pos < maxPos {
		debugLog("[ScrollPane] 超出右边界，返回 %.2f", maxPos)
		return maxPos
	}

	// 以屏幕像素为基准
	v2 := math.Abs(velocity) * p.velocityScale
	debugLog("[ScrollPane] 速度计算: v2=%.2f (threshold=50)", v2)

	// 速度阈值判断 - 使用更平滑的削弱曲线
	if v2 > 50 { // 阈值 50
		// 使用线性削弱而不是平方削弱，避免接近阈值时过度削弱
		ratio := (v2 - 50) / 200 // 使用 200 作为分母，而不是 500
		if ratio > 1 {
			ratio = 1
		}

		// 只在高速时削弱，低速时保持原汁原味
		if v2 > 200 {
			v2 *= ratio
			velocity *= ratio
		}

		// 更新 velocity（根据 axis）
		if axis == "x" {
			p.velocity.X = velocity
		} else {
			p.velocity.Y = velocity
		}

		debugLog("[ScrollPane] 速度足够: v2=%.2f, velocity=%.2f, ratio=%.2f", v2, velocity, ratio)

		// 如果速度足够，计算持续时间
		if v2 > 10 {
			// 算法：v*（decelerationRate的n次幂）= 60，即在n帧后速度降为60（假设每秒60帧）。
			// 对数求解：n = log(60/v2) / log(decelerationRate) / 60
			// 其中 60 是速度阈值
			duration = math.Log(60/v2) / math.Log(p.decelerationRate) / 60

			// 计算距离使用经验公式
			// change = floor(v * duration * 0.4)
			change := math.Floor(velocity * duration * 0.4)
			debugLog("[ScrollPane] 计算惯性: duration=%.2f, change=%.2f, oldPos=%.2f", duration, change, pos)
			pos += change
			debugLog("[ScrollPane] 新位置: %.2f", pos)
		} else {
			debugLog("[ScrollPane] 速度不足 v2=%.2f (需要 >10)", v2)
		}
	} else {
		debugLog("[ScrollPane] 速度过低 v2=%.2f (需要 >50)", v2)
	}

	// 确保最小动画时间
	if duration < TWEEN_TIME_DEFAULT {
		duration = TWEEN_TIME_DEFAULT
	}
	debugLog("[ScrollPane] 最终 duration=%.2f", duration)

	// 设置动画持续时间
	if axis == "x" {
		p.tweenDuration.X = duration
	} else {
		p.tweenDuration.Y = duration
	}

	return pos
}
