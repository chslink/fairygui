package widgets

import (
	"math"
	"strconv"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// ProgressTitleType mirrors FairyGUI 进度文本显示方式。
type ProgressTitleType int

const (
	ProgressTitleTypePercent ProgressTitleType = iota
	ProgressTitleTypeValueAndMax
	ProgressTitleTypeValue
	ProgressTitleTypeMax
)

// GProgressBar 是百分比进度条控件。
type GProgressBar struct {
	*core.GComponent
	packageItem *assets.PackageItem
	template    *core.GComponent

	titleObject *core.GObject
	aniObject   *core.GObject
	barObjectH  *core.GObject
	barObjectV  *core.GObject

	barMaxWidth       float64
	barMaxHeight      float64
	barMaxWidthDelta  float64
	barMaxHeightDelta float64
	barStartX         float64
	barStartY         float64

	min                float64
	max                float64
	value              float64
	title              ProgressTitleType
	reverse            bool
	listenerRegistered bool
}

// ComponentRoot exposes the embedded component for helpers.
func (b *GProgressBar) ComponentRoot() *core.GComponent {
	if b == nil {
		return nil
	}
	return b.GComponent
}

// NewProgressBar 创建默认进度条。
func NewProgressBar() *GProgressBar {
	comp := core.NewGComponent()
	bar := &GProgressBar{
		GComponent: comp,
		max:        100,
		value:      50,
		title:      ProgressTitleTypePercent,
	}
	comp.SetData(bar)
	return bar
}

// SetPackageItem 记录模板来源。
func (b *GProgressBar) SetPackageItem(item *assets.PackageItem) {
	b.packageItem = item
}

// PackageItem 返回模板来源。
func (b *GProgressBar) PackageItem() *assets.PackageItem {
	return b.packageItem
}

// SetTemplateComponent 挂载模板组件。
func (b *GProgressBar) SetTemplateComponent(comp *core.GComponent) {
	if b.template != nil && b.GComponent != nil {
		b.GComponent.RemoveChild(b.template.GObject)
	}
	b.template = comp
	if comp != nil && b.GComponent != nil {
		b.GComponent.AddChild(comp.GObject)
	}
	b.resolveTemplate()
	b.refreshLayout()
	b.applyValue()
}

// TemplateComponent 返回模板组件。
func (b *GProgressBar) TemplateComponent() *core.GComponent {
	return b.template
}

// SetTitleObject 缓存标题对象。
func (b *GProgressBar) SetTitleObject(obj *core.GObject) {
	b.titleObject = obj
	b.applyTitle()
}

// TitleObject 返回标题对象。
func (b *GProgressBar) TitleObject() *core.GObject {
	return b.titleObject
}

// SetAnimationObject 设置动画对象。
func (b *GProgressBar) SetAnimationObject(obj *core.GObject) {
	b.aniObject = obj
}

// SetHorizontalBar 设置横向进度条对象。
func (b *GProgressBar) SetHorizontalBar(obj *core.GObject) {
	b.barObjectH = obj
	b.refreshLayout()
}

// SetVerticalBar 设置纵向进度条对象。
func (b *GProgressBar) SetVerticalBar(obj *core.GObject) {
	b.barObjectV = obj
	b.refreshLayout()
}

// SetMin 更新最小值。
func (b *GProgressBar) SetMin(value float64) {
	b.min = value
	if b.max < b.min {
		b.max = b.min
	}
	if b.value < b.min {
		b.value = b.min
	}
	b.applyValue()
}

// Min 返回最小值。
func (b *GProgressBar) Min() float64 {
	return b.min
}

// SetMax 更新最大值。
func (b *GProgressBar) SetMax(value float64) {
	b.max = value
	if b.max < b.min {
		b.min = b.max
	}
	if b.value > b.max {
		b.value = b.max
	}
	b.applyValue()
}

// Max 返回最大值。
func (b *GProgressBar) Max() float64 {
	return b.max
}

// SetValue 设置当前值。
func (b *GProgressBar) SetValue(val float64) {
	if val < b.min {
		val = b.min
	} else if val > b.max {
		val = b.max
	}
	if b.value == val && b.barObjectH != nil && b.barObjectV != nil {
		return
	}
	b.value = val
	b.applyValue()
}

// Value 返回当前值。
func (b *GProgressBar) Value() float64 {
	return b.value
}

// SetTitleType 配置标题显示模式。
func (b *GProgressBar) SetTitleType(tp ProgressTitleType) {
	b.title = tp
	b.applyTitle()
}

// TitleType 返回标题显示模式。
func (b *GProgressBar) TitleType() ProgressTitleType {
	return b.title
}

// SetReverse 设置反向填充。
func (b *GProgressBar) SetReverse(value bool) {
	b.reverse = value
	b.applyValue()
}

// Reverse 返回是否反向。
func (b *GProgressBar) Reverse() bool {
	return b.reverse
}

// SetSize 重写尺寸变化，确保内部布局刷新。
func (b *GProgressBar) SetSize(width, height float64) {
	if b.GComponent == nil {
		return
	}
	b.GComponent.SetSize(width, height)
	b.refreshLayout()
	b.applyValue()
}

func (b *GProgressBar) resolveTemplate() {
	if b.template == nil {
		return
	}
	if child := b.template.ChildByName("title"); child != nil {
		b.titleObject = child
	}
	if child := b.template.ChildByName("ani"); child != nil {
		b.aniObject = child
	}
	if child := b.template.ChildByName("bar"); child != nil {
		b.barObjectH = child
	}
	if child := b.template.ChildByName("bar_v"); child != nil {
		b.barObjectV = child
	}
}

func (b *GProgressBar) refreshLayout() {
	if b.GComponent == nil {
		return
	}
	if b.barObjectH != nil {
		b.barMaxWidth = b.barObjectH.Width()
		b.barStartX = b.barObjectH.X()
		b.barMaxWidthDelta = b.Width() - b.barMaxWidth
	}
	if b.barObjectV != nil {
		b.barMaxHeight = b.barObjectV.Height()
		b.barStartY = b.barObjectV.Y()
		b.barMaxHeightDelta = b.Height() - b.barMaxHeight
	}
}

func (b *GProgressBar) applyValue() {
	if b.max <= b.min {
		return
	}
	percent := clamp01((b.value - b.min) / (b.max - b.min))
	b.applyTitle()
	fullWidth := b.Width() - b.barMaxWidthDelta
	fullHeight := b.Height() - b.barMaxHeightDelta

	if b.barObjectH != nil {
		width := math.Round(fullWidth * percent)
		if width < 0 {
			width = 0
		}
		if !b.reverse {
			b.barObjectH.SetSize(width, b.barObjectH.Height())
			b.barObjectH.SetPosition(b.barStartX, b.barObjectH.Y())
		} else {
			b.barObjectH.SetSize(width, b.barObjectH.Height())
			b.barObjectH.SetPosition(b.barStartX+(fullWidth-width), b.barObjectH.Y())
		}
	}

	if b.barObjectV != nil {
		height := math.Round(fullHeight * percent)
		if height < 0 {
			height = 0
		}
		if !b.reverse {
			b.barObjectV.SetSize(b.barObjectV.Width(), height)
			b.barObjectV.SetPosition(b.barObjectV.X(), b.barStartY)
		} else {
			b.barObjectV.SetSize(b.barObjectV.Width(), height)
			b.barObjectV.SetPosition(b.barObjectV.X(), b.barStartY+(fullHeight-height))
		}
	}

	if b.aniObject != nil {
		b.aniObject.SetData(percent)
	}
}

func (b *GProgressBar) applyTitle() {
	if b.titleObject == nil {
		return
	}
	span := b.max - b.min
	if span <= 0 {
		span = 1
	}
	text := ""
	switch b.title {
	case ProgressTitleTypePercent:
		text = formatPercent((b.value - b.min) / span)
	case ProgressTitleTypeValue:
		text = formatNumber(b.value)
	case ProgressTitleTypeMax:
		text = formatNumber(b.max)
	case ProgressTitleTypeValueAndMax:
		text = formatNumber(b.value) + "/" + formatNumber(b.max)
	default:
		text = formatNumber(b.value)
	}
	applyTextToObject(b.titleObject, text)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func formatPercent(v float64) string {
	return formatNumber(math.Floor(v*100+0.5)) + "%"
}

func formatNumber(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func applyTextToObject(obj *core.GObject, text string) {
	if obj == nil {
		return
	}
	switch data := obj.Data().(type) {
	case *GTextField:
		data.SetText(text)
	case *GLabel:
		data.SetTitle(text)
	case *GButton:
		data.SetTitle(text)
	case string:
		if data != text {
			obj.SetData(text)
		}
	case nil:
		obj.SetData(text)
	default:
		obj.SetData(text)
	}
}

// SetupAfterAdd mirrors FairyGUI 的 runtime 初始化逻辑，解析当前值与区间。
func (b *GProgressBar) SetupAfterAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	if b == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(0, 6) || buf.Remaining() <= 0 {
		b.applyValue()
		return
	}
	expected := assets.ObjectTypeProgressBar
	if ctx != nil {
		switch {
		case ctx.ResolvedItem != nil && ctx.ResolvedItem.ObjectType != assets.ObjectTypeComponent:
			expected = ctx.ResolvedItem.ObjectType
		case ctx.Child != nil && ctx.Child.Type != 0:
			expected = ctx.Child.Type
		}
	}
	if assets.ObjectType(buf.ReadByte()) != expected {
		b.applyValue()
		return
	}
	if buf.Remaining() < 4 {
		b.applyValue()
		return
	}
	value := float64(buf.ReadInt32())
	if buf.Remaining() < 4 {
		b.applyValue()
		return
	}
	max := float64(buf.ReadInt32())
	min := b.Min()
	if buf.Version >= 2 && buf.Remaining() >= 4 {
		min = float64(buf.ReadInt32())
	}
	b.SetMin(min)
	b.SetMax(max)
	b.SetValue(value)
}
