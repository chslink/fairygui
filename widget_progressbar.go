package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// ProgressBar - 简化的进度条控件
// ============================================================================

// ProgressBar 是简化的进度条控件，包装了 pkg/fgui/widgets.GProgressBar。
type ProgressBar struct {
	bar *widgets.GProgressBar
}

// NewProgressBar 创建一个新的进度条控件。
//
// 示例：
//
//	bar := fairygui.NewProgressBar()
//	bar.SetMin(0)
//	bar.SetMax(100)
//	bar.SetValue(50)
func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		bar: widgets.NewProgressBar(),
	}
}

// Min 返回最小值。
func (p *ProgressBar) Min() float64 {
	return p.bar.Min()
}

// SetMin 设置最小值。
//
// 示例：
//
//	bar.SetMin(0)
func (p *ProgressBar) SetMin(min float64) {
	p.bar.SetMin(min)
}

// Max 返回最大值。
func (p *ProgressBar) Max() float64 {
	return p.bar.Max()
}

// SetMax 设置最大值。
//
// 示例：
//
//	bar.SetMax(100)
func (p *ProgressBar) SetMax(max float64) {
	p.bar.SetMax(max)
}

// Value 返回当前值。
func (p *ProgressBar) Value() float64 {
	return p.bar.Value()
}

// SetValue 设置当前值。
//
// 示例：
//
//	bar.SetValue(50)  // 设置为 50%
func (p *ProgressBar) SetValue(value float64) {
	p.bar.SetValue(value)
}

// TitleType 返回标题类型。
func (p *ProgressBar) TitleType() widgets.ProgressTitleType {
	return p.bar.TitleType()
}

// SetTitleType 设置标题类型。
//
// 示例：
//
//	bar.SetTitleType(widgets.ProgressTitleTypePercent)  // 显示百分比
func (p *ProgressBar) SetTitleType(titleType widgets.ProgressTitleType) {
	p.bar.SetTitleType(titleType)
}

// SetReverse 设置是否反向显示。
//
// 示例：
//
//	bar.SetReverse(true)  // 从右到左或从下到上
func (p *ProgressBar) SetReverse(reverse bool) {
	p.bar.SetReverse(reverse)
}

// Position 返回进度条位置。
func (p *ProgressBar) Position() (x, y float64) {
	return p.bar.X(), p.bar.Y()
}

// SetPosition 设置进度条位置。
func (p *ProgressBar) SetPosition(x, y float64) {
	p.bar.SetPosition(x, y)
}

// Size 返回进度条大小。
func (p *ProgressBar) Size() (width, height float64) {
	return p.bar.Width(), p.bar.Height()
}

// SetSize 设置进度条大小。
func (p *ProgressBar) SetSize(width, height float64) {
	p.bar.SetSize(width, height)
}

// Visible 返回进度条是否可见。
func (p *ProgressBar) Visible() bool {
	return p.bar.Visible()
}

// SetVisible 设置进度条可见性。
func (p *ProgressBar) SetVisible(visible bool) {
	p.bar.SetVisible(visible)
}

// Name 返回进度条名称。
func (p *ProgressBar) Name() string {
	return p.bar.Name()
}

// SetName 设置进度条名称。
func (p *ProgressBar) SetName(name string) {
	p.bar.SetName(name)
}

// Alpha 返回进度条透明度（0-1）。
func (p *ProgressBar) Alpha() float64 {
	return p.bar.Alpha()
}

// SetAlpha 设置进度条透明度（0-1）。
func (p *ProgressBar) SetAlpha(alpha float64) {
	p.bar.SetAlpha(alpha)
}

// RawProgressBar 返回底层的 widgets.GProgressBar 对象。
//
// 仅在需要访问底层 API 时使用。
func (p *ProgressBar) RawProgressBar() *widgets.GProgressBar {
	return p.bar
}
