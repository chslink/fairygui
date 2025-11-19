package fairygui

import (
	"fmt"
	"math"
)

// ============================================================================
// ProgressTitleType - 进度标题类型
// ============================================================================

type ProgressTitleType int

const (
	ProgressTitleTypePercent ProgressTitleType = iota
	ProgressTitleTypeValueAndMax
	ProgressTitleTypeValue
	ProgressTitleTypeMax
)

// ============================================================================
// ProgressBar - 进度条控件 V2 (基于新架构)
// ============================================================================

type ProgressBar struct {
	*ComponentImpl

	// 资源相关
	packageItem *PackageItemWrapper
	template    *ComponentImpl

	// 子对象
	titleObject DisplayObject
	aniObject   DisplayObject
	barObjectH  DisplayObject // 横向进度条
	barObjectV  DisplayObject // 纵向进度条

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
	title ProgressTitleType

	// 反向填充
	reverse bool
}

// NewProgressBar 创建新的进度条
func NewProgressBar() *ProgressBar {
	pb := &ProgressBar{
		ComponentImpl: NewComponent(),
		max:           100,
		value:         0,
		title:         ProgressTitleTypePercent,
		reverse:       false,
	}

	pb.titleObject = nil
	pb.aniObject = nil
	pb.barObjectH = nil
	pb.barObjectV = nil

	return pb
}

// ============================================================================
// 资源相关
// ============================================================================

// SetPackageItem 设置资源项
func (pb *ProgressBar) SetPackageItem(item PackageItem) {
	if item == nil {
		pb.packageItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		pb.packageItem = wrapper
	}
}

// PackageItem 返回资源项
func (pb *ProgressBar) PackageItem() PackageItem {
	return pb.packageItem
}

// SetTemplateComponent 设置模板组件
func (pb *ProgressBar) SetTemplateComponent(comp *ComponentImpl) {
	if pb.template != nil {
		pb.RemoveChild(pb.template)
	}
	pb.template = comp
	if comp != nil {
		comp.SetPosition(0, 0)
		pb.AddChild(comp)
		pb.inheritSizeFromTemplate()
	}
	pb.resolveTemplate()
	pb.refreshLayout()
	pb.applyValue()
}

// TemplateComponent 返回模板组件
func (pb *ProgressBar) TemplateComponent() *ComponentImpl {
	return pb.template
}

// ============================================================================
// 子对象设置
// ============================================================================

// SetTitleObject 设置标题对象
func (pb *ProgressBar) SetTitleObject(obj DisplayObject) {
	pb.titleObject = obj
	pb.applyTitle()
}

// TitleObject 返回标题对象
func (pb *ProgressBar) TitleObject() DisplayObject {
	return pb.titleObject
}

// SetAnimationObject 设置动画对象
func (pb *ProgressBar) SetAnimationObject(obj DisplayObject) {
	pb.aniObject = obj
}

// AnimationObject 返回动画对象
func (pb *ProgressBar) AnimationObject() DisplayObject {
	return pb.aniObject
}

// SetHorizontalBar 设置横向进度条对象
func (pb *ProgressBar) SetHorizontalBar(obj DisplayObject) {
	pb.barObjectH = obj
	pb.refreshLayout()
}

// HorizontalBar 返回横向进度条对象
func (pb *ProgressBar) HorizontalBar() DisplayObject {
	return pb.barObjectH
}

// SetVerticalBar 设置纵向进度条对象
func (pb *ProgressBar) SetVerticalBar(obj DisplayObject) {
	pb.barObjectV = obj
	pb.refreshLayout()
}

// VerticalBar 返回纵向进度条对象
func (pb *ProgressBar) VerticalBar() DisplayObject {
	return pb.barObjectV
}

// ============================================================================
// 值范围设置
// ============================================================================

// SetMin 设置最小值
func (pb *ProgressBar) SetMin(value float64) {
	if pb.min == value {
		return
	}
	pb.min = value
	if pb.max < pb.min {
		pb.max = pb.min
	}
	if pb.value < pb.min {
		pb.value = pb.min
	}
	pb.applyValue()
}

// Min 返回最小值
func (pb *ProgressBar) Min() float64 {
	return pb.min
}

// SetMax 设置最大值
func (pb *ProgressBar) SetMax(value float64) {
	if pb.max == value {
		return
	}
	pb.max = value
	if pb.max < pb.min {
		pb.min = pb.max
	}
	if pb.value > pb.max {
		pb.value = pb.max
	}
	pb.applyValue()
}

// Max 返回最大值
func (pb *ProgressBar) Max() float64 {
	return pb.max
}

// SetValue 设置当前值
func (pb *ProgressBar) SetValue(val float64) {
	if val < pb.min {
		val = pb.min
	} else if val > pb.max {
		val = pb.max
	}
	if pb.value == val {
		return
	}
	pb.value = val
	pb.applyValue()
}

// Value 返回当前值
func (pb *ProgressBar) Value() float64 {
	return pb.value
}

// SetTitleType 设置标题类型
func (pb *ProgressBar) SetTitleType(tp ProgressTitleType) {
	if pb.title == tp {
		return
	}
	pb.title = tp
	pb.applyTitle()
}

// TitleType 返回标题类型
func (pb *ProgressBar) TitleType() ProgressTitleType {
	return pb.title
}

// SetReverse 设置反向填充
func (pb *ProgressBar) SetReverse(value bool) {
	if pb.reverse == value {
		return
	}
	pb.reverse = value
	pb.applyValue()
}

// Reverse 返回是否反向填充
func (pb *ProgressBar) Reverse() bool {
	return pb.reverse
}

// ============================================================================
// 尺寸处理
// ============================================================================

// SetSize 设置尺寸（重写父类方法）
func (pb *ProgressBar) SetSize(width, height float64) {
	pb.ComponentImpl.SetSize(width, height)
	pb.refreshLayout()
	pb.applyValue()
}

// inheritSizeFromTemplate 从模板继承尺寸
func (pb *ProgressBar) inheritSizeFromTemplate() {
	if pb.template == nil {
		return
	}

	tmplW, tmplH := pb.template.Size()
	curW, curH := pb.Size()

	var newW, newH float64
	if curW == 0 && tmplW > 0 {
		newW = tmplW
	} else {
		newW = curW
	}

	if curH == 0 && tmplH > 0 {
		newH = tmplH
	} else {
		newH = curH
	}

	if newW > 0 || newH > 0 {
		pb.ComponentImpl.SetSize(newW, newH)
	}
}

// ============================================================================
// 内部方法
// ============================================================================

// resolveTemplate 解析模板
func (pb *ProgressBar) resolveTemplate() {
	if pb.template == nil {
		return
	}

	// 查找 title
	if child := pb.template.GetChildByName("title"); child != nil {
		pb.titleObject = child
	}

	// 查找 ani
	if child := pb.template.GetChildByName("ani"); child != nil {
		pb.aniObject = child
	}

	// 查找 bar（横向）
	if child := pb.template.GetChildByName("bar"); child != nil {
		pb.barObjectH = child
		pb.setupCircularProgress(child)
	}

	// 查找 bar_v（纵向）
	if child := pb.template.GetChildByName("bar_v"); child != nil {
		pb.barObjectV = child
		pb.setupCircularProgress(child)
	}
}

// setupCircularProgress 设置圆形进度条
func (pb *ProgressBar) setupCircularProgress(bar DisplayObject) {
	// 检查是否是 Image（支持 fillAmount）
	if img, ok := bar.(*Image); ok && img != nil {
		if item := img.PackageItem(); item != nil {
			// 如果是圆形进度条图片，设置径向填充
			if item.ID() == "gzpr80" || containsString(item.Name(), "circle") ||
				containsString(item.Name(), "radial") || item.Name() == "bar" {
				img.SetFill(5, 0, true, 100.0) // 5 = Radial360
			}
		}
	}
}

// refreshLayout 刷新布局
func (pb *ProgressBar) refreshLayout() {
	pbW, pbH := pb.Size()

	if pb.barObjectH != nil {
		barW, _ := pb.barObjectH.Size()
		pb.barMaxWidth = barW
		pb.barStartX, _ = pb.barObjectH.Position()
		pb.barMaxWidthDelta = pbW - pb.barMaxWidth
	} else {
		pb.barMaxWidth = 0
		pb.barMaxWidthDelta = 0
	}

	if pb.barObjectV != nil {
		_, barH := pb.barObjectV.Size()
		pb.barMaxHeight = barH
		_, pb.barStartY = pb.barObjectV.Position()
		pb.barMaxHeightDelta = pbH - pb.barMaxHeight
	} else {
		pb.barMaxHeight = 0
		pb.barMaxHeightDelta = 0
	}
}

// applyValue 应用值（核心方法）
func (pb *ProgressBar) applyValue() {
	if pb.max <= pb.min {
		return
	}

	// 计算百分比
	span := pb.max - pb.min
	if span <= 0 {
		span = 1
	}
	percent := clamp01((pb.value - pb.min) / span)

	// 更新标题
	pb.applyTitle()

	// 计算完整尺寸
	pbW, pbH := pb.Size()
	fullWidth := pbW - pb.barMaxWidthDelta
	fullHeight := pbH - pb.barMaxHeightDelta

	// 更新横向进度条
	if pb.barObjectH != nil {
		pb.updateBar(pb.barObjectH, percent, fullWidth, false)
	}

	// 更新纵向进度条
	if pb.barObjectV != nil {
		pb.updateBar(pb.barObjectV, percent, fullHeight, true)
	}

	// 更新动画对象
	if pb.aniObject != nil {
		pb.aniObject.SetData(percent)
	}
}

// updateBar 更新进度条
func (pb *ProgressBar) updateBar(bar DisplayObject, percent, fullSize float64, vertical bool) {
	if bar == nil {
		return
	}

	// 优先尝试使用 fillAmount
	if pb.setFillAmount(bar, percent) {
		return
	}

	// 回退方案：修改尺寸
	size := math.Round(fullSize * percent)
	if size < 0 {
		size = 0
	}

	if vertical {
		// 纵向进度条
		barW, _ := bar.Size()
		if !pb.reverse {
			// 从上到下
			bar.SetSize(barW, size)
			x, y := bar.Position()
			bar.SetPosition(x, y)
		} else {
			// 从下到上
			bar.SetSize(barW, size)
			x, _ := bar.Position()
			bar.SetPosition(x, pb.barStartY+(fullSize-size))
		}
	} else {
		// 横向进度条
		_, barH := bar.Size()
		if !pb.reverse {
			// 从左到右
			bar.SetSize(size, barH)
			_, y := bar.Position()
			bar.SetPosition(pb.barStartX, y)
		} else {
			// 从右到左
			bar.SetSize(size, barH)
			_, y := bar.Position()
			bar.SetPosition(pb.barStartX+(fullSize-size), y)
		}
	}
}

// setFillAmount 尝试设置 fillAmount
func (pb *ProgressBar) setFillAmount(bar DisplayObject, percent float64) bool {
	if bar == nil {
		return false
	}

	// 检查是否是 Image
	if img, ok := bar.(*Image); ok && img != nil {
		method, _, _, _ := img.Fill()
		if method > 0 {
			// 有 fillMethod，使用 fillAmount（百分比，0-100）
			img.SetFill(method, 0, true, percent*100.0)
			return true
		}
	}

	return false
}

// applyTitle 应用标题
func (pb *ProgressBar) applyTitle() {
	if pb.titleObject == nil {
		return
	}

	span := pb.max - pb.min
	if span <= 0 {
		span = 1
	}

	text := ""
	switch pb.title {
	case ProgressTitleTypePercent:
		text = formatPercent((pb.value - pb.min) / span * 100)
	case ProgressTitleTypeValue:
		text = formatNumber(pb.value)
	case ProgressTitleTypeMax:
		text = formatNumber(pb.max)
	case ProgressTitleTypeValueAndMax:
		text = fmt.Sprintf("%.0f/%.0f", pb.value, pb.max)
	default:
		text = formatNumber(pb.value)
	}

	// 尝试设置文本
	if tf, ok := pb.titleObject.(*TextField); ok {
		tf.SetText(text)
	} else if comp, ok := pb.titleObject.(*ComponentImpl); ok {
		comp.SetData(text)
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

// formatPercent 格式化百分比
func formatPercent(value float64) string {
	percent := math.Round(value)
	return fmt.Sprintf("%.0f%%", percent)
}

// formatNumber 格式化数字
func formatNumber(value float64) string {
	return fmt.Sprintf("%.0f", value)
}

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// clamp01 限制值在 0-1 之间
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertProgressBar 类型断言
func AssertProgressBar(obj DisplayObject) (*ProgressBar, bool) {
	pb, ok := obj.(*ProgressBar)
	return pb, ok
}

// IsProgressBar 检查是否是 ProgressBar
func IsProgressBar(obj DisplayObject) bool {
	_, ok := obj.(*ProgressBar)
	return ok
}
