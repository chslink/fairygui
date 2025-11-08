package widgets

import (
	"math"
	"strconv"
	"strings"

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
		// 关键修复：ProgressBar的尺寸应该从模板组件继承
		// 参考TypeScript版本，模板组件的尺寸会自动影响ProgressBar的尺寸
		if comp.Width() > 0 || comp.Height() > 0 {
			if b.GComponent.Width() == 0 && comp.Width() > 0 {
				b.GComponent.SetSize(comp.Width(), b.GComponent.Height())
			}
			if b.GComponent.Height() == 0 && comp.Height() > 0 {
				b.GComponent.SetSize(b.GComponent.Width(), comp.Height())
			}
		}
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
		// 关键修复：手动设置圆形进度条的fillMethod
		// CircleProgress组件中的bar GImage使用radial360填充，但.fui中没有保存此属性
		// 因此我们需要手动设置
		if image, ok := child.Data().(*GImage); ok && image != nil {
			// 检查PackageItem的ID，如果是径向填充的图片，则设置fillMethod
			if pkgItem := image.PackageItem(); pkgItem != nil {
				// 某些特殊的进度条图片需要径向填充
				// 根据PackageItem的ID或Name判断
				if pkgItem.ID == "gzpr80" || strings.Contains(pkgItem.Name, "circle") ||
				   strings.Contains(pkgItem.Name, "radial") || pkgItem.Name == "bar" {
					image.SetFill(int(LoaderFillMethodRadial360), 0, true, 100.0)
				}
			}
		}
	}
	if child := b.template.ChildByName("bar_v"); child != nil {
		b.barObjectV = child
		// 同样处理纵向进度条
		if image, ok := child.Data().(*GImage); ok && image != nil {
			if pkgItem := image.PackageItem(); pkgItem != nil {
				if pkgItem.ID == "gzpr80" || strings.Contains(pkgItem.Name, "circle") ||
				   strings.Contains(pkgItem.Name, "radial") || pkgItem.Name == "bar" {
					image.SetFill(int(LoaderFillMethodRadial360), 0, true, 100.0)
				}
			}
		}
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
		// TypeScript版本优先使用setFillAmount，只有失败时才修改width/height
		if !b.setFillAmount(b.barObjectH, percent) {
			width := math.Round(fullWidth * percent)
			if width < 0 {
				width = 0
			}
			if !b.reverse {
				if width != b.barObjectH.Width() {
					// 调试：记录bar宽度变化
					// fmt.Printf("[ProgressBar] barH width: %.0f -> %.0f (percent=%.2f)\n", b.barObjectH.Width(), width, percent)
					b.barObjectH.SetSize(width, b.barObjectH.Height())
					b.barObjectH.SetPosition(b.barStartX, b.barObjectH.Y())
				}
			} else {
				if width != b.barObjectH.Width() {
					// fmt.Printf("[ProgressBar] barH width: %.0f -> %.0f (reverse, percent=%.2f)\n", b.barObjectH.Width(), width, percent)
					b.barObjectH.SetSize(width, b.barObjectH.Height())
					b.barObjectH.SetPosition(b.barStartX+(fullWidth-width), b.barObjectH.Y())
				}
			}
		} else {
		}
	} else {
	}

	if b.barObjectV != nil {
		// TypeScript版本优先使用setFillAmount，只有失败时才修改width/height
		if !b.setFillAmount(b.barObjectV, percent) {
			height := math.Round(fullHeight * percent)
			if height < 0 {
				height = 0
			}
			if !b.reverse {
				if height != b.barObjectV.Height() {
					// fmt.Printf("[ProgressBar] barV height: %.0f -> %.0f (percent=%.2f)\n", b.barObjectV.Height(), height, percent)
					b.barObjectV.SetSize(b.barObjectV.Width(), height)
					b.barObjectV.SetPosition(b.barObjectV.X(), b.barStartY)
				}
			} else {
				if height != b.barObjectV.Height() {
					// fmt.Printf("[ProgressBar] barV height: %.0f -> %.0f (reverse, percent=%.2f)\n", b.barObjectV.Height(), height, percent)
					b.barObjectV.SetSize(b.barObjectV.Width(), height)
					b.barObjectV.SetPosition(b.barObjectV.X(), b.barStartY+(fullHeight-height))
				}
			}
		}
	}

	if b.aniObject != nil {
		b.aniObject.SetData(percent)
	}
}

// setFillAmount 尝试设置bar的fillAmount（TypeScript版本的关键方法）
// 返回true表示成功使用fillAmount，false表示需要回退到修改width/height
func (b *GProgressBar) setFillAmount(bar *core.GObject, percent float64) bool {
	if bar == nil {
		return false
	}

	// 检查是否是GImage
	if image, ok := bar.Data().(*GImage); ok {
		method, _, _, _ := image.Fill()
		if method > 0 {
			// 有fillMethod，使用fillAmount（百分比，0-100）
			image.SetFill(method, image.fillOrigin, image.fillClockwise, percent*100.0)
			return true
		} else {
		}
	}

	// 检查是否是GLoader
	if loader, ok := bar.Data().(*GLoader); ok {
		if loader.FillMethod() > 0 {
			// 有fillMethod，使用fillAmount（百分比，0-100）
			loader.SetFillAmount(percent * 100.0)
			return true
		}
	}

	// 既不是GImage也不是GLoader，或者没有fillMethod
	return false
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

// ensureSizeFromTemplate 确保ProgressBar的尺寸从模板组件继承
// 这个修复解决了ProgressBar显示为0x0的问题
func (b *GProgressBar) ensureSizeFromTemplate() {
	if b == nil || b.GComponent == nil {
		return
	}
	// 如果模板组件存在且ProgressBar的尺寸为0，则从模板继承
	if b.template != nil {
		tmplW := b.template.Width()
		tmplH := b.template.Height()
		if tmplW > 0 || tmplH > 0 {
			curW := b.GComponent.Width()
			curH := b.GComponent.Height()
			// 如果ProgressBar尺寸仍然是0x0，则从模板继承
			if (curW == 0 && tmplW > 0) || (curH == 0 && tmplH > 0) {
				newW := curW
				newH := curH
				if newW == 0 && tmplW > 0 {
					newW = tmplW
				}
				if newH == 0 && tmplH > 0 {
					newH = tmplH
				}
				if newW > 0 || newH > 0 {
					b.GComponent.SetSize(newW, newH)
					// 尺寸改变后需要刷新布局和应用值
					b.refreshLayout()
					b.applyValue()
				}
			}
		}
	}
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
		// 修复：确保SetupAfterAdd后尺寸正确
		b.ensureSizeFromTemplate()
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
		// 修复：确保SetupAfterAdd后尺寸正确
		b.ensureSizeFromTemplate()
		return
	}
	if buf.Remaining() < 4 {
		b.applyValue()
		// 修复：确保SetupAfterAdd后尺寸正确
		b.ensureSizeFromTemplate()
		return
	}
	value := float64(buf.ReadInt32())
	if buf.Remaining() < 4 {
		b.applyValue()
		// 修复：确保SetupAfterAdd后尺寸正确
		b.ensureSizeFromTemplate()
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
	// 修复：确保SetupAfterAdd后尺寸正确
	b.ensureSizeFromTemplate()
}
