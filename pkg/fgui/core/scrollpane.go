package core

import (
	"math"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
)

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
}

func newScrollPane(owner *GComponent) *ScrollPane {
	if owner == nil {
		return nil
	}
	pane := &ScrollPane{
		owner:           owner,
		scrollType:      ScrollTypeBoth,
		scrollStep:      25,
		mouseWheelStep:  50,
		wheelEnabled:    true,
		touchEffect:     true,
		pageSize:        laya.Point{X: 1, Y: 1},
		scrollListeners: make(map[int]ScrollListener),
	}
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
	p.viewSize = laya.Point{X: width, Y: height}
	if p.pageMode {
		if p.pageSize.X <= 0 {
			p.pageSize.X = width
		}
		if p.pageSize.Y <= 0 {
			p.pageSize.Y = height
		}
	}
	p.updateScrollRect()
	p.refreshOverlap()
	p.clampPosition()
	p.notifyScrollListeners()
}

// SetContentSize updates the scrollable content size。
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
	p.unregisterEvents()
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
	}
}

func (p *ScrollPane) applyPosition() {
	if p == nil || p.container == nil {
		return
	}
	p.container.SetPosition(-p.xPos, -p.yPos)
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
	rect := &laya.Rect{
		X: 0,
		Y: 0,
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
		p.wheelListener = func(evt laya.Event) {
			pe, ok := evt.Data.(laya.PointerEvent)
			if !ok {
				return
			}
			p.handleMouseWheel(pe.WheelX, pe.WheelY)
		}
		dispatcher.On(laya.EventMouseWheel, p.wheelListener)
	}
	if p.mouseDownListener == nil {
		p.mouseDownListener = func(evt laya.Event) {
			p.onMouseDown(evt)
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
	if elapsed > 0 {
		p.velocity = laya.Point{
			X: deltaX / elapsed,
			Y: deltaY / elapsed,
		}
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
	if p.container != nil {
		p.containerOrigin = p.container.Position()
	}
	if p.pageMode {
		p.snapToNearestPage()
	} else {
		p.clampPosition()
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
		p.stageMoveListener = func(evt laya.Event) {
			p.onStageMouseMove(evt)
		}
	}
	if p.stageUpListener == nil {
		p.stageUpListener = func(evt laya.Event) {
			p.onStageMouseUp(evt)
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
