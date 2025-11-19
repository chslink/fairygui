package fairygui

// ============================================================================
// ScrollPane V2 - 滚动面板桩实现
// ============================================================================

type ScrollPaneV2 struct {
	*ComponentImpl
	perX           float64
	perY           float64
	scrollPercX    float64
	scrollPercY    float64
	scrollStep     float64
	isNone         bool
	isDragged      bool
	listeners      []ScrollListener
	nextListenerID int
}

// NewScrollPaneV2 创建新的滚动面板
func NewScrollPaneV2() *ScrollPaneV2 {
	return &ScrollPaneV2{
		ComponentImpl: NewComponent(),
		scrollStep:    10,
		listeners:     make([]ScrollListener, 0),
	}
}

// AddScrollListener 添加滚动监听器
func (p *ScrollPaneV2) AddScrollListener(fn ScrollListener) int {
	p.listeners = append(p.listeners, fn)
	id := p.nextListenerID
	p.nextListenerID++
	return id
}

// RemoveScrollListener 移除滚动监听器
func (p *ScrollPaneV2) RemoveScrollListener(id int) {
	// 简化实现：这里不做实际移除，仅作示意
	_ = id
}

// SetPercX 设置水平滚动百分比
func (p *ScrollPaneV2) SetPercX(value float64, ani bool) {
	p.perX = value
	p.scrollPercX = value
	p.notifyListeners()
	_ = ani
}

// SetPercY 设置垂直滚动百分比
func (p *ScrollPaneV2) SetPercY(value float64, ani bool) {
	p.perY = value
	p.scrollPercY = value
	p.notifyListeners()
	_ = ani
}

// ScrollUp 向上滚动
func (p *ScrollPaneV2) ScrollUp() {
	p.perY -= p.scrollStep / 100
	if p.perY < 0 {
		p.perY = 0
	}
	p.scrollPercY = p.perY
	p.notifyListeners()
}

// ScrollDown 向下滚动
func (p *ScrollPaneV2) ScrollDown() {
	p.perY += p.scrollStep / 100
	if p.perY > 1 {
		p.perY = 1
	}
	p.scrollPercY = p.perY
	p.notifyListeners()
}

// ScrollLeft 向左滚动
func (p *ScrollPaneV2) ScrollLeft() {
	p.perX -= p.scrollStep / 100
	if p.perX < 0 {
		p.perX = 0
	}
	p.scrollPercX = p.perX
	p.notifyListeners()
}

// ScrollRight 向右滚动
func (p *ScrollPaneV2) ScrollRight() {
	p.perX += p.scrollStep / 100
	if p.perX > 1 {
		p.perX = 1
	}
	p.scrollPercX = p.perX
	p.notifyListeners()
}

// notifyListeners 通知监听器
func (p *ScrollPaneV2) notifyListeners() {
	info := ScrollInfo{
		PercX:        p.scrollPercX,
		PercY:        p.scrollPercY,
		DisplayPercX: 0.5, // 默认值
		DisplayPercY: 0.5,
	}

	for _, listener := range p.listeners {
		if listener != nil {
			listener(info)
		}
	}
}

// GetScrollPos 返回滚动位置
func (p *ScrollPaneV2) GetScrollPos() Point {
	return Point{X: p.perX * 100, Y: p.perY * 100}
}

// GetViewSize 返回视口大小
func (p *ScrollPaneV2) GetViewSize() Point {
	return Point{X: 100, Y: 100}
}

// ScrollToView 滚动到指定位置
func (p *ScrollPaneV2) ScrollToView(index int, animated bool) {
	// 桩实现
	_ = index
	_ = animated
}

// Clear 清空滚动面板
func (p *ScrollPaneV2) Clear() {
	// 桩实现
}

// SetBoundsChanged 设置边界改变（桩实现）
func (p *ScrollPaneV2) SetBoundsChanged() {
	// 桩实现
}

// EnsureSizeCorrect 确保尺寸正确（桩实现）
func (p *ScrollPaneV2) EnsureSizeCorrect() {
	// 桩实现
}

// GetFirstChildInView 获取第一个可见的子对象
func (p *ScrollPaneV2) GetFirstChildInView() int {
	return 0
}
