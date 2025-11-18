package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Slider - 简化的滑杆控件
// ============================================================================

// Slider 是简化的滑杆控件，包装了 pkg/fgui/widgets.GSlider。
type Slider struct {
	slider *widgets.GSlider
}

// NewSlider 创建一个新的滑杆控件。
//
// 示例：
//
//	slider := fairygui.NewSlider()
//	slider.SetMin(0)
//	slider.SetMax(100)
//	slider.SetValue(50)
func NewSlider() *Slider {
	return &Slider{
		slider: widgets.NewSlider(),
	}
}

// Min 返回最小值。
func (s *Slider) Min() float64 {
	return s.slider.Min()
}

// SetMin 设置最小值。
//
// 示例：
//
//	slider.SetMin(0)
func (s *Slider) SetMin(min float64) {
	s.slider.SetMin(min)
}

// Max 返回最大值。
func (s *Slider) Max() float64 {
	return s.slider.Max()
}

// SetMax 设置最大值。
//
// 示例：
//
//	slider.SetMax(100)
func (s *Slider) SetMax(max float64) {
	s.slider.SetMax(max)
}

// Value 返回当前值。
func (s *Slider) Value() float64 {
	return s.slider.Value()
}

// SetValue 设置当前值。
//
// 示例：
//
//	slider.SetValue(50)
func (s *Slider) SetValue(value float64) {
	s.slider.SetValue(value)
}

// SetWholeNumbers 设置是否为整数模式。
//
// 在整数模式下，值会被四舍五入为整数。
//
// 示例：
//
//	slider.SetWholeNumbers(true)
func (s *Slider) SetWholeNumbers(whole bool) {
	s.slider.SetWholeNumbers(whole)
}

// SetChangeOnClick 设置是否点击改变值。
//
// 启用后，点击滑杆会直接跳到点击位置。
//
// 示例：
//
//	slider.SetChangeOnClick(true)
func (s *Slider) SetChangeOnClick(change bool) {
	s.slider.SetChangeOnClick(change)
}

// SetReverse 设置是否反向显示。
//
// 示例：
//
//	slider.SetReverse(true)
func (s *Slider) SetReverse(reverse bool) {
	s.slider.SetReverse(reverse)
}

// SetTitleType 设置标题类型。
//
// 示例：
//
//	slider.SetTitleType(widgets.ProgressTitleTypePercent)
func (s *Slider) SetTitleType(titleType widgets.ProgressTitleType) {
	s.slider.SetTitleType(titleType)
}

// Position 返回滑杆位置。
func (s *Slider) Position() (x, y float64) {
	return s.slider.X(), s.slider.Y()
}

// SetPosition 设置滑杆位置。
func (s *Slider) SetPosition(x, y float64) {
	s.slider.SetPosition(x, y)
}

// Size 返回滑杆大小。
func (s *Slider) Size() (width, height float64) {
	return s.slider.Width(), s.slider.Height()
}

// SetSize 设置滑杆大小。
func (s *Slider) SetSize(width, height float64) {
	s.slider.SetSize(width, height)
}

// Visible 返回滑杆是否可见。
func (s *Slider) Visible() bool {
	return s.slider.Visible()
}

// SetVisible 设置滑杆可见性。
func (s *Slider) SetVisible(visible bool) {
	s.slider.SetVisible(visible)
}

// Name 返回滑杆名称。
func (s *Slider) Name() string {
	return s.slider.Name()
}

// SetName 设置滑杆名称。
func (s *Slider) SetName(name string) {
	s.slider.SetName(name)
}

// Alpha 返回滑杆透明度（0-1）。
func (s *Slider) Alpha() float64 {
	return s.slider.Alpha()
}

// SetAlpha 设置滑杆透明度（0-1）。
func (s *Slider) SetAlpha(alpha float64) {
	s.slider.SetAlpha(alpha)
}

// RawSlider 返回底层的 widgets.GSlider 对象。
//
// 仅在需要访问底层 API 时使用。
func (s *Slider) RawSlider() *widgets.GSlider {
	return s.slider
}
