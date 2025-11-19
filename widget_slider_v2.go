package fairygui

import (
	"math"
	"sync"
)

// ============================================================================
// Slider - 滑块控件 V2 (基于新架构)
// ============================================================================

type Slider struct {
	*ComponentImpl

	// 资源相关
	packageItem *PackageItemWrapper
	template    *ComponentImpl

	// 子对象
	titleObject DisplayObject
	barObjectH  DisplayObject
	barObjectV  DisplayObject
	gripObject  DisplayObject

	// 布局参数
	barMaxWidth       float64
	barMaxHeight      float64
	barMaxWidthDelta  float64
	barMaxHeightDelta float64
	barStartX         float64
	barStartY         float64

	// 值范围
	min   float64
	max   float64
	value float64

	// 标题类型
	titleType ProgressTitleType

	// 行为设置
	reverse       bool
	wholeNumbers  bool
	changeOnClick bool

	// 拖拽状态
	clickPos     Point
	clickPercent float64
	dragging     bool

	// 监听器
	stageMoveListener EventHandler
	stageUpListener   EventHandler
	mu                sync.Mutex
}

// Point 表示一个点
type Point struct {
	X float64
	Y float64
}

// NewSlider 创建新的滑块
func NewSlider() *Slider {
	s := &Slider{
		ComponentImpl: NewComponent(),
		max:           100,
		value:         0,
		titleType:     ProgressTitleTypePercent,
		changeOnClick: true,
		wholeNumbers:  false,
		reverse:       false,
	}

	return s
}

// ============================================================================
// 资源相关
// ============================================================================

// SetPackageItem 设置资源项
func (s *Slider) SetPackageItem(item PackageItem) {
	if item == nil {
		s.packageItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		s.packageItem = wrapper
	}
}

// PackageItem 返回资源项
func (s *Slider) PackageItem() PackageItem {
	return s.packageItem
}

// SetTemplateComponent 设置模板组件
func (s *Slider) SetTemplateComponent(comp *ComponentImpl) {
	if s.template != nil {
		s.RemoveChild(s.template)
	}
	s.template = comp
	if comp != nil {
		comp.SetPosition(0, 0)
		s.AddChild(comp)
	}
	s.resolveTemplate()
	s.refreshLayout()
	s.applyValue(false)
}

// TemplateComponent 返回模板组件
func (s *Slider) TemplateComponent() *ComponentImpl {
	return s.template
}

// ============================================================================
// 子对象设置
// ============================================================================

// SetTitleObject 设置标题对象
func (s *Slider) SetTitleObject(obj DisplayObject) {
	s.titleObject = obj
	s.applyTitle()
}

// TitleObject 返回标题对象
func (s *Slider) TitleObject() DisplayObject {
	return s.titleObject
}

// SetHorizontalBar 设置横向进度条
func (s *Slider) SetHorizontalBar(obj DisplayObject) {
	s.barObjectH = obj
	s.refreshLayout()
}

// HorizontalBar 返回横向进度条
func (s *Slider) HorizontalBar() DisplayObject {
	return s.barObjectH
}

// SetVerticalBar 设置纵向进度条
func (s *Slider) SetVerticalBar(obj DisplayObject) {
	s.barObjectV = obj
	s.refreshLayout()
}

// VerticalBar 返回纵向进度条
func (s *Slider) VerticalBar() DisplayObject {
	return s.barObjectV
}

// SetGripObject 设置抓手对象
func (s *Slider) SetGripObject(obj DisplayObject) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 移除旧事件监听
	if s.gripObject != nil {
		s.gripObject.Off("mousedown", s.onGripMouseDown)
	}

	s.gripObject = obj

	// 添加新事件监听
	if obj != nil {
		obj.On("mousedown", s.onGripMouseDown)
	}
}

// GripObject 返回抓手对象
func (s *Slider) GripObject() DisplayObject {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gripObject
}

// ============================================================================
// 值范围设置
// ============================================================================

// SetMin 设置最小值
func (s *Slider) SetMin(val float64) *Slider {
	if s.min == val {
		return s
	}
	s.min = val
	if s.max < s.min {
		s.max = s.min
	}
	if s.value < s.min {
		s.value = s.min
	}
	s.applyValue(false)
	return s
}

// Min 返回最小值
func (s *Slider) Min() float64 {
	return s.min
}

// SetMax 设置最大值
func (s *Slider) SetMax(val float64) *Slider {
	if s.max == val {
		return s
	}
	s.max = val
	if s.max < s.min {
		s.min = s.max
	}
	if s.value > s.max {
		s.value = s.max
	}
	s.applyValue(false)
	return s
}

// Max 返回最大值
func (s *Slider) Max() float64 {
	return s.max
}

// SetValue 设置当前值
func (s *Slider) SetValue(val float64) *Slider {
	if val < s.min {
		val = s.min
	} else if val > s.max {
		val = s.max
	}
	if s.wholeNumbers {
		val = math.Round(val)
	}
	if s.value == val {
		return s
	}
	s.value = val
	s.applyValue(false)
	return s
}

// Value 返回当前值
func (s *Slider) Value() float64 {
	return s.value
}

// ============================================================================
// 行为设置
// ============================================================================

// SetWholeNumbers 设置是否使用整数
func (s *Slider) SetWholeNumbers(enabled bool) *Slider {
	if s.wholeNumbers == enabled {
		return s
	}
	s.wholeNumbers = enabled
	s.applyValue(false)
	return s
}

// WholeNumbers 返回是否使用整数
func (s *Slider) WholeNumbers() bool {
	return s.wholeNumbers
}

// SetChangeOnClick 设置点击条是否改变值
func (s *Slider) SetChangeOnClick(enabled bool) *Slider {
	s.changeOnClick = enabled
	return s
}

// ChangeOnClick 返回点击条是否改变值
func (s *Slider) ChangeOnClick() bool {
	return s.changeOnClick
}

// SetReverse 设置反向填充
func (s *Slider) SetReverse(value bool) *Slider {
	if s.reverse == value {
		return s
	}
	s.reverse = value
	s.applyValue(false)
	return s
}

// Reverse 返回是否反向填充
func (s *Slider) Reverse() bool {
	return s.reverse
}

// SetTitleType 设置标题类型
func (s *Slider) SetTitleType(tp ProgressTitleType) *Slider {
	if s.titleType == tp {
		return s
	}
	s.titleType = tp
	s.applyTitle()
	return s
}

// TitleType 返回标题类型
func (s *Slider) TitleType() ProgressTitleType {
	return s.titleType
}

// ============================================================================
// 尺寸处理
// ============================================================================

// SetSize 设置尺寸（重写父类方法）
func (s *Slider) SetSize(width, height float64) {
	s.ComponentImpl.SetSize(width, height)
	s.refreshLayout()
	s.applyValue(false)
}

// ============================================================================
// 内部方法
// ============================================================================

// resolveTemplate 解析模板
func (s *Slider) resolveTemplate() {
	if s.template == nil {
		return
	}

	// 查找 title
	if child := s.template.GetChildByName("title"); child != nil {
		s.titleObject = child
	}

	// 查找 bar
	if child := s.template.GetChildByName("bar"); child != nil {
		s.barObjectH = child
	}

	// 查找 bar_v
	if child := s.template.GetChildByName("bar_v"); child != nil {
		s.barObjectV = child
	}

	// 查找 grip
	if child := s.template.GetChildByName("grip"); child != nil {
		s.SetGripObject(child)
	}

	// 绑定条点击事件
	if s.template != nil {
		s.template.OnMouseDown(func() {
			s.onBarMouseDown(nil)
		})
	}
}

// refreshLayout 刷新布局
func (s *Slider) refreshLayout() {
	sW, sH := s.Size()

	if s.barObjectH != nil {
		barW, _ := s.barObjectH.Size()
		s.barMaxWidth = barW
		s.barStartX, _ = s.barObjectH.Position()
		s.barMaxWidthDelta = sW - s.barMaxWidth
	} else {
		s.barMaxWidth = 0
		s.barMaxWidthDelta = 0
	}

	if s.barObjectV != nil {
		_, barH := s.barObjectV.Size()
		s.barMaxHeight = barH
		_, s.barStartY = s.barObjectV.Position()
		s.barMaxHeightDelta = sH - s.barMaxHeight
	} else {
		s.barMaxHeight = 0
		s.barMaxHeightDelta = 0
	}
}

// applyValue 应用值
func (s *Slider) applyValue(fromEvent bool) {
	if s.max <= s.min {
		return
	}

	percent := clamp01((s.value - s.min) / (s.max - s.min))
	if fromEvent && s.wholeNumbers {
		s.value = math.Round(s.value)
		percent = clamp01((s.value - s.min) / (s.max - s.min))
	}

	s.applyTitle()

	sW, sH := s.Size()
	fullWidth := sW - s.barMaxWidthDelta
	fullHeight := sH - s.barMaxHeightDelta

	// 更新横向进度条
	if s.barObjectH != nil {
		size := math.Round(fullWidth * percent)
		if size < 0 {
			size = 0
		}
		_, barH := s.barObjectH.Size()
		if !s.reverse {
			s.barObjectH.SetSize(size, barH)
			_, y := s.barObjectH.Position()
			s.barObjectH.SetPosition(s.barStartX, y)
		} else {
			s.barObjectH.SetSize(size, barH)
			_, y := s.barObjectH.Position()
			s.barObjectH.SetPosition(s.barStartX+(fullWidth-size), y)
		}
	}

	// 更新纵向进度条
	if s.barObjectV != nil {
		size := math.Round(fullHeight * percent)
		if size < 0 {
			size = 0
		}
		barW, _ := s.barObjectV.Size()
		if !s.reverse {
			s.barObjectV.SetSize(barW, size)
			x, _ := s.barObjectV.Position()
			s.barObjectV.SetPosition(x, s.barStartY)
		} else {
			x, _ := s.barObjectV.Position()
			s.barObjectV.SetSize(barW, size)
			s.barObjectV.SetPosition(x, s.barStartY+(fullHeight-size))
		}
	}

	// 更新抓手位置
	if s.gripObject != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		gripW, gripH := s.gripObject.Size()

		switch {
		case s.barObjectH != nil && fullWidth > 0:
			pos := s.barStartX + (fullWidth-gripW)*percent
			if s.reverse {
				pos = s.barStartX + (fullWidth-gripW)*(1-percent)
			}
			_, gy := s.gripObject.Position()
			s.gripObject.SetPosition(pos, gy)

		case s.barObjectV != nil && fullHeight > 0:
			pos := s.barStartY + (fullHeight-gripH)*percent
			if s.reverse {
				pos = s.barStartY + (fullHeight-gripH)*(1-percent)
			}
			gx, _ := s.gripObject.Position()
			s.gripObject.SetPosition(gx, pos)
		}
	}
}

// applyTitle 应用标题
func (s *Slider) applyTitle() {
	if s.titleObject == nil {
		return
	}

	span := s.max - s.min
	if span <= 0 {
		span = 1
	}

	text := ""
	switch s.titleType {
	case ProgressTitleTypePercent:
		percent := (s.value - s.min) / span * 100
		text = formatPercent(percent)
	case ProgressTitleTypeValue:
		text = formatNumber(s.value)
	case ProgressTitleTypeMax:
		text = formatNumber(s.max)
	case ProgressTitleTypeValueAndMax:
		text = formatNumber(s.value) + "/" + formatNumber(s.max)
	default:
		text = formatNumber(s.value)
	}

	// 尝试设置文本
	if tf, ok := s.titleObject.(*TextField); ok {
		tf.SetText(text)
	} else if comp, ok := s.titleObject.(*ComponentImpl); ok {
		comp.SetData(text)
	}
}

// onGripMouseDown 抓手按下
func (s *Slider) onGripMouseDown(event Event) {
	// 阻止事件冒泡
	if baseEvent, ok := event.(*BaseEvent); ok {
		baseEvent.StopPropagation()
	}

	if s.gripObject == nil {
		return
	}

	s.dragging = true

	// 获取点击位置（这里简化处理）
	// 实际应该在事件系统中获取鼠标/触摸坐标
	s.clickPercent = clamp01((s.value - s.min) / (s.max - s.min))

	s.registerStageDrag()
}

// onStageMouseMove 舞台鼠标移动
func (s *Slider) onStageMouseMove(event Event) {
	if !s.dragging {
		return
	}

	// 这里简化处理，实际应该从事件中获取位置
	// 假设鼠标/触摸位置改变
	s.updateFromInteraction(s.clickPercent)
}

// onStageMouseUp 舞台鼠标释放
func (s *Slider) onStageMouseUp(event Event) {
	if !s.dragging {
		return
	}
	s.dragging = false
	s.unregisterStageDrag()
}

// onBarMouseDown 条按下
func (s *Slider) onBarMouseDown(event Event) {
	if !s.changeOnClick {
		return
	}

	// 简化处理，实际应该计算点击位置在条上的百分比
	percent := 0.5
	s.updateFromInteraction(percent)
}

// updateFromInteraction 从交互更新值
func (s *Slider) updateFromInteraction(percent float64) {
	percent = clamp01(percent)
	newValue := s.min + (s.max-s.min)*percent
	if s.wholeNumbers {
		newValue = math.Round(newValue)
	}
	if newValue == s.value {
		s.applyValue(false)
		return
	}
	s.value = newValue
	s.applyValue(true)

	// 触发状态改变事件
	s.Emit(NewUIEvent("statechanged", s, s.value))
}

// registerStageDrag 注册舞台拖拽
func (s *Slider) registerStageDrag() {
	s.stageMoveListener = func(event Event) {
		s.onStageMouseMove(event)
	}
	s.stageUpListener = func(event Event) {
		s.onStageMouseUp(event)
	}

	// 这里应该注册到舞台事件系统
	// 简化实现：在组件级别监听
	s.On("mousemove", s.stageMoveListener)
	s.On("mouseup", s.stageUpListener)
}

// unregisterStageDrag 取消注册舞台拖拽
func (s *Slider) unregisterStageDrag() {
	if s.stageMoveListener != nil {
		s.Off("mousemove", s.stageMoveListener)
		s.stageMoveListener = nil
	}
	if s.stageUpListener != nil {
		s.Off("mouseup", s.stageUpListener)
		s.stageUpListener = nil
	}
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertSlider 类型断言
func AssertSlider(obj DisplayObject) (*Slider, bool) {
	slider, ok := obj.(*Slider)
	return slider, ok
}

// IsSlider 检查是否是 Slider
func IsSlider(obj DisplayObject) bool {
	_, ok := obj.(*Slider)
	return ok
}
