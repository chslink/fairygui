package widgets

import (
	"math"
	"strconv"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// GSlider 代表可拖动的滑杆控件。
type GSlider struct {
	*core.GComponent
	packageItem *assets.PackageItem
	template    *core.GComponent

	titleObject *core.GObject
	barObjectH  *core.GObject
	barObjectV  *core.GObject
	gripObject  *core.GObject

	barMaxWidth       float64
	barMaxHeight      float64
	barMaxWidthDelta  float64
	barMaxHeightDelta float64
	barStartX         float64
	barStartY         float64

	min           float64
	max           float64
	value         float64
	titleType     ProgressTitleType
	reverse       bool
	wholeNumbers  bool
	changeOnClick bool

	clickPos     laya.Point
	clickPercent float64
	dragging     bool

	stageMoveListener laya.Listener
	stageUpListener   laya.Listener
}

// NewSlider 构建默认 slider。
func NewSlider() *GSlider {
	comp := core.NewGComponent()
	s := &GSlider{
		GComponent:    comp,
		max:           100,
		titleType:     ProgressTitleTypePercent,
		changeOnClick: true,
	}
	comp.SetData(s)
	return s
}

// SetPackageItem 保存模板资源。
func (s *GSlider) SetPackageItem(item *assets.PackageItem) {
	s.packageItem = item
}

// PackageItem 返回模板资源。
func (s *GSlider) PackageItem() *assets.PackageItem {
	return s.packageItem
}

// SetTemplateComponent 安装模板。
func (s *GSlider) SetTemplateComponent(comp *core.GComponent) {
	if s.template != nil && s.GComponent != nil {
		s.GComponent.RemoveChild(s.template.GObject)
	}
	s.template = comp
	if comp != nil && s.GComponent != nil {
		s.GComponent.AddChild(comp.GObject)
	}
	s.resolveTemplate()
	s.refreshLayout()
	s.applyValue(false)
}

// TemplateComponent 返回模板。
func (s *GSlider) TemplateComponent() *core.GComponent {
	return s.template
}

// SetTitleObject 缓存标题对象。
func (s *GSlider) SetTitleObject(obj *core.GObject) {
	s.titleObject = obj
	s.applyTitle()
}

// SetHorizontalBar 缓存横向条。
func (s *GSlider) SetHorizontalBar(obj *core.GObject) {
	s.barObjectH = obj
	s.refreshLayout()
}

// SetVerticalBar 缓存纵向条。
func (s *GSlider) SetVerticalBar(obj *core.GObject) {
	s.barObjectV = obj
	s.refreshLayout()
}

// SetGripObject 缓存抓手。
func (s *GSlider) SetGripObject(obj *core.GObject) {
	if s.gripObject != nil && s.gripObject.DisplayObject() != nil {
		s.gripObject.DisplayObject().Dispatcher().Off(laya.EventMouseDown, s.onGripMouseDown)
	}
	s.gripObject = obj
	if obj != nil && obj.DisplayObject() != nil {
		obj.DisplayObject().Dispatcher().On(laya.EventMouseDown, s.onGripMouseDown)
	}
}

// SetMin 设置最小值。
func (s *GSlider) SetMin(val float64) {
	s.min = val
	if s.max < s.min {
		s.max = s.min
	}
	if s.value < s.min {
		s.value = s.min
	}
	s.applyValue(false)
}

// SetMax 设置最大值。
func (s *GSlider) SetMax(val float64) {
	s.max = val
	if s.max < s.min {
		s.min = s.max
	}
	if s.value > s.max {
		s.value = s.max
	}
	s.applyValue(false)
}

// SetWholeNumbers 启用整数模式。
func (s *GSlider) SetWholeNumbers(enabled bool) {
	s.wholeNumbers = enabled
	s.applyValue(false)
}

// SetChangeOnClick 控制点击条是否修改数值。
func (s *GSlider) SetChangeOnClick(enabled bool) {
	s.changeOnClick = enabled
}

// SetReverse 设置反向填充。
func (s *GSlider) SetReverse(value bool) {
	s.reverse = value
	s.applyValue(false)
}

// SetTitleType 更新标题模式。
func (s *GSlider) SetTitleType(tp ProgressTitleType) {
	s.titleType = tp
	s.applyTitle()
}

// SetValue 更新当前值。
func (s *GSlider) SetValue(val float64) {
	if val < s.min {
		val = s.min
	} else if val > s.max {
		val = s.max
	}
	if s.wholeNumbers {
		val = math.Round(val)
	}
	if s.value == val {
		return
	}
	s.value = val
	s.applyValue(false)
}

// Value 返回当前值。
func (s *GSlider) Value() float64 {
	return s.value
}

// Min 返回最小值。
func (s *GSlider) Min() float64 {
	return s.min
}

// Max 返回最大值。
func (s *GSlider) Max() float64 {
	return s.max
}

// SetSize 重写尺寸变化。
func (s *GSlider) SetSize(width, height float64) {
	if s.GComponent == nil {
		return
	}
	s.GComponent.SetSize(width, height)
	s.refreshLayout()
	s.applyValue(false)
}

func (s *GSlider) resolveTemplate() {
	if s.template == nil {
		return
	}
	if child := s.template.ChildByName("title"); child != nil {
		s.titleObject = child
	}
	if child := s.template.ChildByName("bar"); child != nil {
		s.barObjectH = child
	}
	if child := s.template.ChildByName("bar_v"); child != nil {
		s.barObjectV = child
	}
	if child := s.template.ChildByName("grip"); child != nil {
		s.SetGripObject(child)
	}
	if s.template.GObject != nil && s.template.GObject.DisplayObject() != nil {
		s.template.GObject.DisplayObject().Dispatcher().On(laya.EventMouseDown, s.onBarMouseDown)
	}
}

func (s *GSlider) refreshLayout() {
	if s.GComponent == nil {
		return
	}
	if s.barObjectH != nil {
		s.barMaxWidth = s.barObjectH.Width()
		s.barStartX = s.barObjectH.X()
		s.barMaxWidthDelta = s.Width() - s.barMaxWidth
	}
	if s.barObjectV != nil {
		s.barMaxHeight = s.barObjectV.Height()
		s.barStartY = s.barObjectV.Y()
		s.barMaxHeightDelta = s.Height() - s.barMaxHeight
	}
}

func (s *GSlider) applyValue(fromEvent bool) {
	if s.max <= s.min {
		return
	}
	percent := clamp01((s.value - s.min) / (s.max - s.min))
	if fromEvent && s.wholeNumbers {
		s.value = math.Round(s.value)
		percent = clamp01((s.value - s.min) / (s.max - s.min))
	}
	s.applyTitle()

	fullWidth := s.Width() - s.barMaxWidthDelta
	fullHeight := s.Height() - s.barMaxHeightDelta

	if s.barObjectH != nil {
		width := math.Round(fullWidth * percent)
		if width < 0 {
			width = 0
		}
		if !s.reverse {
			s.barObjectH.SetSize(width, s.barObjectH.Height())
			s.barObjectH.SetPosition(s.barStartX, s.barObjectH.Y())
		} else {
			s.barObjectH.SetSize(width, s.barObjectH.Height())
			s.barObjectH.SetPosition(s.barStartX+(fullWidth-width), s.barObjectH.Y())
		}
	}
	if s.barObjectV != nil {
		height := math.Round(fullHeight * percent)
		if height < 0 {
			height = 0
		}
		if !s.reverse {
			s.barObjectV.SetSize(s.barObjectV.Width(), height)
			s.barObjectV.SetPosition(s.barObjectV.X(), s.barStartY)
		} else {
			s.barObjectV.SetSize(s.barObjectV.Width(), height)
			s.barObjectV.SetPosition(s.barObjectV.X(), s.barStartY+(fullHeight-height))
		}
	}

	if s.gripObject != nil {
		switch {
		case s.barObjectH != nil && fullWidth > 0:
			pos := s.barStartX + (fullWidth-s.gripObject.Width())*percent
			if s.reverse {
				pos = s.barStartX + (fullWidth-s.gripObject.Width())*(1-percent)
			}
			s.gripObject.SetPosition(pos, s.gripObject.Y())
		case s.barObjectV != nil && fullHeight > 0:
			pos := s.barStartY + (fullHeight-s.gripObject.Height())*percent
			if s.reverse {
				pos = s.barStartY + (fullHeight-s.gripObject.Height())*(1-percent)
			}
			s.gripObject.SetPosition(s.gripObject.X(), pos)
		}
	}
}

func (s *GSlider) applyTitle() {
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
		text = strconv.Itoa(int(math.Round((s.value - s.min) * 100 / span)))
		text += "%"
	case ProgressTitleTypeValue:
		text = strconv.Itoa(int(math.Round(s.value)))
	case ProgressTitleTypeMax:
		text = strconv.Itoa(int(math.Round(s.max)))
	case ProgressTitleTypeValueAndMax:
		text = strconv.Itoa(int(math.Round(s.value))) + "/" + strconv.Itoa(int(math.Round(s.max)))
	default:
		text = strconv.Itoa(int(math.Round(s.value)))
	}
	applyTextToObject(s.titleObject, text)
}

func (s *GSlider) onGripMouseDown(evt laya.Event) {
	if s.gripObject == nil || s.GComponent == nil {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok {
		return
	}
	local := s.GComponent.DisplayObject().GlobalToLocal(pe.Position)
	s.clickPos = local
	s.clickPercent = clamp01((s.value - s.min) / (s.max - s.min))
	s.dragging = true
	s.registerStageDrag()
}

func (s *GSlider) onStageMouseMove(evt laya.Event) {
	if !s.dragging || s.GComponent == nil {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok {
		return
	}
	local := s.GComponent.DisplayObject().GlobalToLocal(pe.Position)
	deltaX := local.X - s.clickPos.X
	deltaY := local.Y - s.clickPos.Y
	if s.reverse {
		deltaX = -deltaX
		deltaY = -deltaY
	}
	percent := s.clickPercent
	if s.barObjectH != nil && s.barMaxWidth != 0 {
		percent += deltaX / s.barMaxWidth
	} else if s.barObjectV != nil && s.barMaxHeight != 0 {
		percent += deltaY / s.barMaxHeight
	}
	s.updateFromInteraction(percent)
}

func (s *GSlider) onStageMouseUp(evt laya.Event) {
	if !s.dragging {
		return
	}
	s.dragging = false
	s.unregisterStageDrag()
}

func (s *GSlider) onBarMouseDown(evt laya.Event) {
	if !s.changeOnClick || s.GComponent == nil {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok {
		return
	}
	if s.barObjectH != nil && s.barMaxWidth > 0 {
		local := s.GComponent.DisplayObject().GlobalToLocal(pe.Position)
		percent := clamp01((local.X - s.barStartX) / s.barMaxWidth)
		if s.reverse {
			percent = 1 - percent
		}
		s.updateFromInteraction(percent)
		return
	}
	if s.barObjectV != nil && s.barMaxHeight > 0 {
		local := s.GComponent.DisplayObject().GlobalToLocal(pe.Position)
		percent := clamp01((local.Y - s.barStartY) / s.barMaxHeight)
		if s.reverse {
			percent = 1 - percent
		}
		s.updateFromInteraction(percent)
	}
}

func (s *GSlider) updateFromInteraction(percent float64) {
	percent = clamp01(percent)
	newValue := s.min + (s.max-s.min)*percent
	if s.wholeNumbers {
		newValue = math.Round(newValue)
		percent = clamp01((newValue - s.min) / (s.max - s.min))
	}
	if newValue == s.value {
		s.applyValue(false)
		return
	}
	s.value = newValue
	s.applyValue(true)
	if s.GComponent != nil {
		s.GComponent.GObject.Emit(laya.EventStateChanged, s.value)
	}
}

func (s *GSlider) registerStageDrag() {
	root := core.Root()
	if root == nil {
		return
	}
	stage := root.Stage()
	if stage == nil {
		return
	}
	dispatcher := stage.Root().Dispatcher()
	if s.stageMoveListener == nil {
		s.stageMoveListener = func(evt laya.Event) {
			s.onStageMouseMove(evt)
		}
	}
	if s.stageUpListener == nil {
		s.stageUpListener = func(evt laya.Event) {
			s.onStageMouseUp(evt)
		}
	}
	dispatcher.On(laya.EventMouseMove, s.stageMoveListener)
	dispatcher.On(laya.EventStageMouseUp, s.stageUpListener)
}

func (s *GSlider) unregisterStageDrag() {
	root := core.Root()
	if root == nil {
		return
	}
	stage := root.Stage()
	if stage == nil {
		return
	}
	dispatcher := stage.Root().Dispatcher()
	if s.stageMoveListener != nil {
		dispatcher.Off(laya.EventMouseMove, s.stageMoveListener)
	}
	if s.stageUpListener != nil {
		dispatcher.Off(laya.EventStageMouseUp, s.stageUpListener)
	}
}
